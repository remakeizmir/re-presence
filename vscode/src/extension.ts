import * as vscode from "vscode";

/**
 * re:make presence — what you have open, on your hub profile.
 *
 * The editor knows more than a window title does: the language, the line, the
 * folder, whether a debug session is running. That is the whole reason this
 * exists alongside the agent — for VS Code and everything built on it (Cursor,
 * Windsurf, Antigravity), the card can be exact.
 *
 * What leaves the machine is the folder's name, the file's name, and the line
 * number. Never a path: on a shared screen, a path is a home address and the
 * name of a client who has not been told about us yet.
 */

/** The server forgets a report after two minutes; thirty seconds keeps the
 *  card up without the card ever flickering. */
const REPORT_EVERY_MS = 30_000;

/** Out of the window this long and the card comes down. Reading documentation
 *  in a browser is part of working; lunch is not. */
const IDLE_AFTER_MS = 5 * 60_000;

let timer: NodeJS.Timeout | undefined;
let lastActive = Date.now();
let sentIdle = false;
let status: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
  status = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 0);
  status.command = "remake.toggle";
  context.subscriptions.push(status);

  context.subscriptions.push(
    vscode.commands.registerCommand("remake.connect", connect),
    vscode.commands.registerCommand("remake.setToken", askForToken),
    vscode.commands.registerCommand("remake.toggle", toggle),
  );

  // Anything that means "still here": a keystroke, a different file, focus
  // coming back to the window.
  const touch = () => {
    lastActive = Date.now();
    sentIdle = false;
  };
  context.subscriptions.push(
    vscode.window.onDidChangeActiveTextEditor(touch),
    vscode.window.onDidChangeTextEditorSelection(touch),
    vscode.workspace.onDidChangeTextDocument(touch),
    vscode.window.onDidChangeWindowState((state) => {
      if (state.focused) touch();
    }),
  );

  timer = setInterval(tick, REPORT_EVERY_MS);
  context.subscriptions.push({ dispose: () => timer && clearInterval(timer) });

  void tick();
  refreshStatus();

  if (!config().get<string>("token")) {
    void vscode.window
      .showInformationMessage(
        "re:make presence kurulu. Bağlamak için bir tıklama yeter.",
        "Bağlan",
      )
      .then((choice) => {
        if (choice) void connect();
      });
  }
}

export async function deactivate() {
  if (timer) clearInterval(timer);
  // Closing the editor takes the card down rather than leaving it to expire.
  await send({ app: appName(), idle: true });
}

function config() {
  return vscode.workspace.getConfiguration("remake");
}

async function tick() {
  if (!config().get<boolean>("enabled", true)) return;
  if (!config().get<string>("token")) return;

  const idle = Date.now() - lastActive > IDLE_AFTER_MS;
  if (idle) {
    if (!sentIdle) {
      sentIdle = true;
      await send({ app: appName(), idle: true });
      refreshStatus();
    }
    return;
  }

  await send(currentReport());
  refreshStatus();
}

interface Report {
  app: string;
  project?: string;
  file?: string;
  line?: number;
  language?: string;
  debugging?: boolean;
  idle?: boolean;
}

function currentReport(): Report {
  const folder = vscode.workspace.workspaceFolders?.[0];
  const project = folder?.name ?? "";

  // A folder listed as private still shows that work is happening, without
  // saying what — the case for it is an NDA, and "hiding the whole card" is
  // the wrong answer to that.
  const privateProjects = config().get<string[]>("privateProjects", []);
  const isPrivate = privateProjects.some(
    (name) => name.toLowerCase() === project.toLowerCase(),
  );

  const editor = vscode.window.activeTextEditor;
  const report: Report = {
    app: appName(),
    project: isPrivate ? "gizli bir proje" : project,
    debugging: Boolean(vscode.debug.activeDebugSession),
  };

  if (editor && !isPrivate) {
    const path = editor.document.uri.path;
    report.file = path.slice(path.lastIndexOf("/") + 1);
    report.language = editor.document.languageId;
    report.line = editor.selection.active.line + 1;
  }
  return report;
}

/** Cursor, Windsurf and Antigravity are all VS Code underneath, and each one
 *  sets its own application name — so the card says which one it really is. */
function appName(): string {
  return vscode.env.appName || "VS Code";
}

async function send(report: Report) {
  const token = config().get<string>("token");
  if (!token) return;

  const base = apiBase();

  try {
    const res = await fetch(`${base}/presence/editor`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(report),
      signal: AbortSignal.timeout(8_000),
    });

    if (res.status === 401) {
      // A revoked key would otherwise fail silently every thirty seconds for
      // as long as the editor is open.
      await config().update("enabled", false, true);
      void vscode.window
        .showWarningMessage(
          "re:make: bağlantı iptal edilmiş. Yeniden bağlanmak ister misin?",
          "Bağlan",
        )
        .then((choice) => {
          if (choice) void connect();
        });
    }
  } catch {
    // The network comes and goes; the card expires on its own and the next
    // tick will try again. Nothing here is worth interrupting anyone for.
  }
}

/**
 * Connects this editor without anyone copying a secret.
 *
 * The extension asks the hub for a six-character code, shows it, and waits.
 * The person types those six characters into the hub page they are already
 * signed in to. Nothing long is copied, nothing is pasted into settings, and
 * the key itself never appears on screen.
 *
 * The alternative — "create a token, copy it, paste it here" — is four steps
 * and a chance to paste a credential into the wrong window.
 */
async function connect() {
  const base = apiBase();

  let code: string;
  let secret: string;
  let pollEvery = 3;
  let verificationUrl = "";
  try {
    const res = await fetch(`${base}/presence/pair`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ label: vscode.env.appName }),
      signal: AbortSignal.timeout(10_000),
    });
    const payload = (await res.json()) as {
      data?: {
        code: string;
        secret: string;
        poll_every: number;
        verification_url: string;
      };
    };
    if (!payload.data?.code) throw new Error("kod alınamadı");
    code = payload.data.code;
    secret = payload.data.secret;
    pollEvery = payload.data.poll_every || 3;
    verificationUrl = payload.data.verification_url ?? "";
  } catch {
    void vscode.window.showErrorMessage("re:make: sunucuya ulaşılamadı.");
    return;
  }

  // The browser opens on a page that already has the code in the box, and the
  // clipboard carries it too — reading six characters off one screen and
  // typing them into another is the step people get wrong.
  await vscode.env.clipboard.writeText(code);
  if (verificationUrl) {
    await vscode.env.openExternal(vscode.Uri.parse(verificationUrl));
  }

  const deadline = Date.now() + 10 * 60_000;
  const connected = await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: `re:make bağlanıyor — kod: ${code}`,
      cancellable: true,
    },
    async (progress, cancel) => {
      progress.report({
        message: verificationUrl
          ? "Tarayıcıda açılan sayfada Bağla'ya bas"
          : "hub → Ayarlar → Bağlantılar → Kod editörü",
      });

      while (Date.now() < deadline && !cancel.isCancellationRequested) {
        await new Promise((resolve) => setTimeout(resolve, pollEvery * 1000));

        try {
          const res = await fetch(`${base}/presence/pair/poll`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ code, secret }),
            signal: AbortSignal.timeout(10_000),
          });
          if (res.status === 202) continue; // nobody has said yes yet
          if (!res.ok) return false;

          const payload = (await res.json()) as { data?: { key: string } };
          if (payload.data?.key) {
            await config().update("token", payload.data.key, true);
            await config().update("enabled", true, true);
            return true;
          }
        } catch {
          // The network comes and goes; keep waiting until the code expires.
        }
      }
      return false;
    },
  );

  if (connected) {
    lastActive = Date.now();
    await tick();
    refreshStatus();
    void vscode.window.showInformationMessage("re:make: bağlandı.");
  } else {
    void vscode.window.showWarningMessage(
      "re:make: bağlanılamadı — kodun süresi dolmuş olabilir.",
    );
  }
}

function apiBase(): string {
  return config()
    .get<string>("apiUrl", "https://api.remakeizmir.com/api/v1")
    .replace(/\/+$/, "");
}

/** The manual way in, kept for anyone who already has a key. */
async function askForToken() {
  const key = await vscode.window.showInputBox({
    title: "re:make cihaz anahtarı",
    prompt: "hub → Ayarlar → Bağlantılar → Kod editörü",
    placeHolder: "rmk_dev_…",
    password: true,
    ignoreFocusOut: true,
  });
  if (!key) return;

  await config().update("token", key.trim(), true);
  await config().update("enabled", true, true);
  lastActive = Date.now();
  await tick();
  void vscode.window.showInformationMessage("re:make: bağlandı.");
}

async function toggle() {
  const enabled = config().get<boolean>("enabled", true);
  await config().update("enabled", !enabled, true);

  if (enabled) {
    await send({ app: appName(), idle: true });
  } else {
    lastActive = Date.now();
    await tick();
  }
  refreshStatus();
}

function refreshStatus() {
  const enabled = config().get<boolean>("enabled", true);
  const hasToken = Boolean(config().get<string>("token"));

  if (!hasToken) {
    status.text = "$(plug) re:make";
    status.tooltip = "Bağlanmak için tıkla";
    status.command = "remake.connect";
  } else {
    status.text = enabled ? "$(broadcast) re:make" : "$(circle-slash) re:make";
    status.tooltip = enabled
      ? "Aktivite paylaşılıyor — kapatmak için tıkla"
      : "Aktivite kapalı — açmak için tıkla";
    status.command = "remake.toggle";
  }
  status.show();
}
