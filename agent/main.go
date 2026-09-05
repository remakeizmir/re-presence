// re:presence agent — what you have open, for editors that cannot say so
// themselves.
//
// Zed's extensions are WebAssembly with no network and no timers. Xcode has no
// extension API for this. Sublime, JetBrains and the terminal editors each
// have their own plugin story and none of them is worth writing four times. So
// this reads the title of the focused window instead, which every editor
// already fills with the file and the project, and reports that.
//
// It is deliberately dumb: no accessibility permissions, no keystroke reading,
// no file watching. A window title, every thirty seconds.
//
//	go build -o re-presence . && ./re-presence
//
// Configuration lives in ~/.config/re-presence/config.json:
//
//	{
//	  "token": "rmk_dev_…",
//	  "api": "https://api.remakeizmir.com/api/v1"
//	}
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	reportEvery = 30 * time.Second
	// idleAfter is how long the editor may be out of focus before the card
	// comes down. Long enough to read a page of documentation in the browser,
	// short enough that "working" means it.
	idleAfter = 5 * time.Minute
)

// editors maps the application name the window manager reports to the name
// shown on the card. Anything not in here is not an editor as far as this is
// concerned, and nothing is reported for it.
var editors = map[string]string{
	"zed":            "Zed",
	"antigravity":    "Antigravity",
	"code":           "VS Code",
	"cursor":         "Cursor",
	"windsurf":       "Windsurf",
	"xcode":          "Xcode",
	"sublime text":   "Sublime Text",
	"neovide":        "Neovim",
	"goland":         "GoLand",
	"webstorm":       "WebStorm",
	"intellij idea":  "IntelliJ",
	"pycharm":        "PyCharm",
	"android studio": "Android Studio",
	"rider":          "Rider",
	"clion":          "CLion",
	"emacs":          "Emacs",
	"nova":           "Nova",
	"zeditor":        "Zed",
}

type config struct {
	Token string `json:"token"`
	API   string `json:"api"`
}

type report struct {
	App       string `json:"app"`
	Project   string `json:"project,omitempty"`
	File      string `json:"file,omitempty"`
	Language  string `json:"language,omitempty"`
	Debugging bool   `json:"debugging,omitempty"`
	Idle      bool   `json:"idle,omitempty"`
}

func main() {
	once := flag.Bool("once", false, "send a single report and exit — for testing")
	pairOnly := flag.Bool("pair", false, "connect this machine and exit — used by the installer")
	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("yapılandırma okunamadı: %v", err)
	}

	// No key yet: ask for a code, show it, and wait for someone to type it
	// into the hub. Nobody edits a file to set this up.
	if cfg.Token == "" || *pairOnly {
		cfg, err = pair(cfg)
		if err != nil {
			log.Fatalf("bağlanamadı: %v", err)
		}
	}
	if *pairOnly {
		return
	}

	if *once {
		window, ok := focusedEditor()
		if !ok {
			fmt.Println("odakta bir editör yok")
			return
		}
		fmt.Printf("gönderiliyor: %+v\n", window)
		if err := send(cfg, window); err != nil {
			log.Fatalf("gönderilemedi: %v", err)
		}
		fmt.Println("tamam")
		return
	}

	log.Printf("[re:presence] başladı — %s", cfg.API)

	// A clean exit takes the card down with it, rather than leaving someone
	// shown as working for two minutes after they shut the laptop.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(reportEvery)
	defer ticker.Stop()

	var lastSeen time.Time
	var lastSentIdle bool

	tick := func() {
		window, ok := focusedEditor()
		if ok {
			lastSeen = time.Now()
			lastSentIdle = false
			if err := send(cfg, window); err != nil {
				log.Printf("[re:presence] gönderilemedi: %v", err)
			}
			return
		}

		// Not in the editor. Give it a while before taking the card down —
		// looking something up is part of working.
		if !lastSentIdle && !lastSeen.IsZero() && time.Since(lastSeen) > idleAfter {
			lastSentIdle = true
			if err := send(cfg, report{App: "editör", Idle: true}); err != nil {
				log.Printf("[re:presence] boşta bildirimi gönderilemedi: %v", err)
			}
		}
	}

	tick()
	for {
		select {
		case <-ticker.C:
			tick()
		case <-stop:
			log.Println("[re:presence] kapanıyor")
			_ = send(cfg, report{App: "editör", Idle: true})
			return
		}
	}
}

// pair walks someone through connecting this machine.
//
// The editor cannot show a login form and should not be handed a password, so
// it shows six characters instead and waits. This is the same flow a
// television uses, and for the same reason.
func pair(cfg config) (config, error) {
	body, _ := json.Marshal(map[string]string{"label": machineLabel()})
	res, err := http.Post(apiURL(cfg, "/presence/pair"), "application/json", bytes.NewReader(body))
	if err != nil {
		return cfg, err
	}
	defer res.Body.Close()

	var start struct {
		Data struct {
			Code            string `json:"code"`
			Secret          string `json:"secret"`
			PollEvery       int    `json:"poll_every"`
			VerificationURL string `json:"verification_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&start); err != nil {
		return cfg, err
	}
	if start.Data.Code == "" {
		return cfg, fmt.Errorf("sunucu kod vermedi")
	}

	// The code goes on the clipboard and the browser opens on the page that
	// asks "is this you?", with the code already in the box. Reading six
	// characters off a terminal and typing them into a browser is the step
	// people get wrong; this removes it without removing the confirmation.
	copyToClipboard(start.Data.Code)

	fmt.Printf(`
  ┌────────────────────────────────────────────────┐
  │                                                │
  │   Kodun:  %s                               │
  │                                                │
  └────────────────────────────────────────────────┘

  Tarayıcı açılıyor — sayfada "Bağla" düğmesine bas.
  Açılmazsa şu adrese git:

  %s

  Bekleniyor…`, start.Data.Code, start.Data.VerificationURL)

	openBrowser(start.Data.VerificationURL)

	every := time.Duration(start.Data.PollEvery) * time.Second
	if every <= 0 {
		every = 3 * time.Second
	}
	deadline := time.Now().Add(10 * time.Minute)

	for time.Now().Before(deadline) {
		time.Sleep(every)

		key, done, err := pollPair(cfg, start.Data.Code, start.Data.Secret)
		if err != nil {
			fmt.Println()
			return cfg, err
		}
		if !done {
			fmt.Print(".")
			continue
		}

		cfg.Token = key
		if err := saveConfig(cfg); err != nil {
			return cfg, err
		}
		fmt.Printf("\n\n  Bağlandı. Bundan sonra kendi kendine çalışır.\n\n")
		return cfg, nil
	}
	fmt.Println()
	return cfg, fmt.Errorf("kodun süresi doldu, tekrar dene")
}

func pollPair(cfg config, code, secret string) (key string, done bool, err error) {
	body, _ := json.Marshal(map[string]string{"code": code, "secret": secret})
	res, err := http.Post(apiURL(cfg, "/presence/pair/poll"), "application/json", bytes.NewReader(body))
	if err != nil {
		return "", false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusAccepted:
		return "", false, nil // nobody has said yes yet
	case http.StatusOK:
		var payload struct {
			Data struct {
				Key string `json:"key"`
			} `json:"data"`
		}
		if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
			return "", false, err
		}
		return payload.Data.Key, true, nil
	default:
		return "", false, fmt.Errorf("kod geçersiz ya da süresi doldu")
	}
}

// openBrowser is best-effort: a machine with no desktop still has the address
// printed above it.
func openBrowser(address string) {
	if address == "" {
		return
	}
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", address).Start()
	case "linux":
		_ = exec.Command("xdg-open", address).Start()
	case "windows":
		// Through the shell, because the URL carries a query string that
		// `start` would otherwise read as its own arguments.
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", address).Start()
	}
}

// copyToClipboard saves the six characters, for the case where the browser
// opens somewhere the address did not carry the code.
func copyToClipboard(text string) {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
	case "linux":
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
	case "windows":
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
	}
}

// machineLabel is what the confirmation screen calls this computer, so someone
// approving a code can see which machine is asking.
func machineLabel() string {
	if name, err := os.Hostname(); err == nil && name != "" {
		return strings.TrimSuffix(name, ".local")
	}
	return "Bilgisayar"
}

func apiURL(cfg config, path string) string {
	return strings.TrimRight(cfg.API, "/") + path
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "re-presence", "config.json"), nil
}

func saveConfig(cfg config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// The key is a credential; nobody else on the machine needs to read it.
	return os.WriteFile(path, data, 0o600)
}

func loadConfig() (config, error) {
	path, err := configPath()
	if err != nil {
		return config{}, err
	}

	cfg := config{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return config{}, err
		}
	case !os.IsNotExist(err):
		return config{}, err
		// A missing file is the first run, not a problem: pairing writes it.
	}
	if cfg.API == "" {
		cfg.API = "https://api.remakeizmir.com/api/v1"
	}
	// An environment variable wins, which is what a launchd plist or a systemd
	// unit will use.
	if token := os.Getenv("RE_PRESENCE_TOKEN"); token != "" {
		cfg.Token = token
	}
	if api := os.Getenv("RE_PRESENCE_API"); api != "" {
		cfg.API = api
	}
	return cfg, nil
}

func send(cfg config, r report) error {
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost,
		strings.TrimRight(cfg.API, "/")+"/presence/editor", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	res, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("anahtar geçersiz ya da iptal edilmiş")
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("sunucu %d döndü", res.StatusCode)
	}
	return nil
}

// ── Reading the window ───────────────────────

// focusedEditor returns what the focused window says it is showing, or false
// when the focused window is not an editor.
func focusedEditor() (report, bool) {
	app, title, err := focusedWindow()
	if err != nil || app == "" {
		return report{}, false
	}

	name, known := editorFromProcess(app)
	if !known {
		return report{}, false
	}

	file, project := splitTitle(title)
	return report{
		App:      name,
		Project:  project,
		File:     file,
		Language: languageOf(file),
	}, true
}

// splitTitle pulls the file and the project out of a window title.
//
// Editors have converged on "file — project" with an em dash (Zed, VS Code
// with a dash, Sublime with a hyphen), sometimes with a leading bullet for
// unsaved changes. Anything that does not fit is treated as a project name on
// its own, which is still worth showing.
func splitTitle(title string) (file, project string) {
	title = strings.TrimSpace(title)
	title = strings.TrimPrefix(title, "●")
	title = strings.TrimPrefix(title, "•")
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ""
	}

	for _, separator := range []string{" — ", " – ", " - ", " | "} {
		if parts := strings.Split(title, separator); len(parts) >= 2 {
			file = strings.TrimSpace(parts[0])
			project = strings.TrimSpace(parts[len(parts)-1])

			// VS Code ends with the editor's own name; the project is the part
			// before it.
			if editorNamed(project) && len(parts) >= 3 {
				project = strings.TrimSpace(parts[len(parts)-2])
			} else if editorNamed(project) {
				project = ""
			}

			// A title that is only "project — Zed" has no file in it.
			if project == "" {
				project, file = file, ""
			}
			return baseOrEmpty(file), baseOrEmpty(project)
		}
	}
	return "", baseOrEmpty(title)
}

// baseOrEmpty is filepath.Base without its habit of answering "." for nothing.
func baseOrEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return filepath.Base(value)
}

// editorNamed recognises the editor's own name at the end of a title. The
// window class and the words in the title bar are not the same string —
// macOS reports the class "Code" while the title ends "Visual Studio Code" —
// so both are checked, plus the names shown on the card.
func editorNamed(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if _, ok := editors[value]; ok {
		return true
	}
	for _, shown := range editors {
		if strings.EqualFold(shown, value) {
			return true
		}
	}
	switch value {
	case "visual studio code", "code - oss", "vscodium", "zed preview":
		return true
	}
	return false
}

// languageOf is a guess from the extension, only used for the little label on
// the card.
func languageOf(file string) string {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx", ".mjs":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".hpp":
		return "c++"
	case ".cs":
		return "c#"
	case ".css", ".scss":
		return "css"
	case ".html":
		return "html"
	case ".sql":
		return "sql"
	case ".md":
		return "markdown"
	case ".json", ".yaml", ".yml", ".toml":
		return "config"
	case ".sh", ".fish", ".zsh":
		return "shell"
	default:
		return ""
	}
}

// editorFromProcess turns what the operating system calls the running program
// into what the card should say. Windows reports executables ("Code.exe",
// "idea64.exe"), macOS and Linux report the application's own name — so both
// tables are consulted, and anything unknown is not an editor as far as this
// is concerned.
func editorFromProcess(name string) (string, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, ".exe")

	if shown, ok := windowsEditors[name]; ok {
		return shown, true
	}
	if shown, ok := editors[name]; ok {
		return shown, true
	}
	return "", false
}

// windowsEditors maps executable names, which rarely match the name on the
// window: JetBrains ships idea64.exe, Android Studio ships studio64.exe, and
// Visual Studio — a different program from VS Code — is devenv.exe.
var windowsEditors = map[string]string{
	"zed":             "Zed",
	"code":            "VS Code",
	"code - insiders": "VS Code",
	"cursor":          "Cursor",
	"windsurf":        "Windsurf",
	"antigravity":     "Antigravity",
	"devenv":          "Visual Studio",
	"rider64":         "Rider",
	"idea64":          "IntelliJ",
	"pycharm64":       "PyCharm",
	"goland64":        "GoLand",
	"webstorm64":      "WebStorm",
	"clion64":         "CLion",
	"studio64":        "Android Studio",
	"sublime_text":    "Sublime Text",
	"neovide":         "Neovim",
	"nvim-qt":         "Neovim",
	"notepad++":       "Notepad++",
	"rustrover64":     "RustRover",
	"phpstorm64":      "PhpStorm",
}
