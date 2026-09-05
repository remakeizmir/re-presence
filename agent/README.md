# re:presence agent

For editors that cannot report for themselves: **Zed**, Xcode, Sublime,
JetBrains, terminal editors. It reads the title of the focused window every
thirty seconds and sends the project and file names to the hub.

It asks for no accessibility permission, reads no keystrokes and watches no
files. On macOS the first run may raise a one-time "allow Terminal to control
System Events" prompt — that is what reads the window title.

## Setup

```sh
go build -o re-presence .
mkdir -p ~/.config/re-presence
cat > ~/.config/re-presence/config.json <<'JSON'
{
  "token": "rmk_dev_…",
  "api": "https://api.remakeizmir.com/api/v1"
}
JSON

./re-presence -once   # tek seferlik dene
./re-presence         # sürekli çalıştır
```

`RE_PRESENCE_TOKEN` and `RE_PRESENCE_API` override the file, which is what the
launch agent below uses.

## Start it with the machine (macOS)

`~/Library/LaunchAgents/com.remakeizmir.presence.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.remakeizmir.presence</string>
  <key>ProgramArguments</key>
  <array><string>/Users/KULLANICI/bin/re-presence</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict>
</plist>
```

```sh
launchctl load ~/Library/LaunchAgents/com.remakeizmir.presence.plist
```

## Linux

Needs `xdotool` (X11). Wayland is not supported: no compositor there hands out
the focused window's title to an unprivileged process, which is the whole
mechanism.

## Which editors it recognises

Zed, Antigravity, VS Code, Cursor, Windsurf, Xcode, Sublime Text, Neovide,
GoLand, WebStorm, IntelliJ, PyCharm, Android Studio, Rider, CLion, Emacs, Nova.
Anything else focused is treated as "not working" and nothing is sent. Adding
one is a line in `editors` in `main.go`.

VS Code and its forks are better served by the extension in `../vscode` — it
knows the line number, the language and whether a debug session is running,
none of which a title bar says.
