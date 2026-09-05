# re:presence

What you have open in your editor, on your hub profile card — the way a song
is. A project, a file, and how long it has been going.

There are two ways in, because editors differ in what they will let a plugin do:

| | What it is | Works with |
|---|---|---|
| **`vscode/`** | A VS Code extension | VS Code, Cursor, Windsurf, Antigravity — anything built on VS Code |
| **`agent/`** | A small background program that reads the focused window's title | **Zed**, Xcode, Sublime, JetBrains, Visual Studio — anything else |

Systems the agent runs on: **macOS**, **Windows**, and **Linux under X11**.
Wayland is the exception — no compositor there hands the focused window's
title to an unprivileged process, which is the whole mechanism.

The agent exists because most editors have no way for a plugin to run in the
background and speak HTTP. Zed's extensions are WebAssembly with no network or
timers; Xcode has no extension API for this at all. Reading the window title is
what is left, and it is enough for the project and the file.

Both send the same request, so a third way is easy to add: see `API.md`.

## Kurulum

**VS Code, Cursor, Windsurf, Antigravity:** eklentiyi kur, çıkan bildirimde
*Bağlan*'a bas. Tarayıcı kodu hazır gelmiş sayfada açılır, tek düğmeye basarsın.

Nereden kurulur:

| Editör | Yer |
|---|---|
| VS Code | [Marketplace](https://marketplace.visualstudio.com/items?itemName=remake.re-presence) |
| Cursor, Windsurf, Antigravity, VSCodium | [Open VSX](https://open-vsx.org/extension/remake/re-presence) — bu editörler Microsoft'un mağazasını kullanamıyor |
| Hepsi (mağazasız) | [remakeizmir.com/re-presence.vsix](https://remakeizmir.com/re-presence.vsix) → Eklentiler → ⋯ → VSIX'ten Yükle |

> **Yayınlamadan önce:** iki mağaza da hesap istiyor ve bunları bir insanın
> açması gerekiyor.
>
> - **VS Code Marketplace:** Azure DevOps hesabı → publisher `remake` →
>   Personal Access Token (Marketplace: Manage yetkisi) → GitHub'da
>   `VSCE_PAT` secret'ı.
> - **Open VSX:** open-vsx.org'a GitHub ile giriş → Eclipse Publisher Agreement
>   → access token → `OVSX_PAT` secret'ı.
>
> İkisi de ücretsiz. Secret'lar konduktan sonra `ext-v0.1.0` gibi bir etiket
> atmak yetiyor: `.github/workflows/extension.yml` paketleyip ikisine de
> yolluyor, ayrıca .vsix'i release'e ekliyor. Secret yoksa yalnız .vsix üretilir
> — indirme bağlantısı yine çalışır.

**Zed, Xcode, JetBrains, diğerleri:** tek satır —

```sh
# macOS ve Linux
curl -fsSL https://remakeizmir.com/presence.sh | bash
```

```powershell
# Windows (PowerShell, yönetici gerekmez)
irm https://remakeizmir.com/presence.ps1 | iex
```

İndirir, altı haneli kodu gösterir, açılışta çalışacak şekilde kurar. Kaldırmak
için `~/.local/share/re-presence/uninstall.sh`.

Kimse uzun bir anahtarı kopyalamıyor: anahtar sunucudan doğrudan editöre
gidiyor, ekranda yalnız altı karakter görünüyor. Kod on dakika geçerli ve tek
kullanımlık. Elle anahtar oluşturma seçeneği duruyor — kendi aracını yazanlar
için.

> `install.sh` hem burada hem `landing/public/presence.sh` içinde duruyor;
> değiştirince `./sync-landing.sh` çalıştır.

## What is sent, and what is kept

Sent: the editor's name, the **folder name** of the project, the **file name**,
the line number, the language, and whether you are debugging. Never the path —
`/Users/can/is/gizli-musteri/src/main.go` is sent as `main.go` in
`gizli-musteri`.

Kept: nothing. The hub holds the last report in memory for two minutes and then
forgets it. There is no history, and no way to ask what someone was doing
yesterday.

Turn it off any time from the same settings panel — the switch stops the hub
showing it, without uninstalling anything.
