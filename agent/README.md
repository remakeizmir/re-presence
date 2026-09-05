# re:presence agent

For editors that cannot report for themselves: **Zed**, Xcode, Sublime,
JetBrains, terminal editors. It reads the title of the focused window every
thirty seconds and sends the project and file names to the hub.

It asks for no accessibility permission, reads no keystrokes and watches no
files. On macOS the first run may raise a one-time "allow Terminal to control
System Events" prompt — that is what reads the window title.

## Kurulum

Kullanıcıysan: [üstteki README](../README.md). Tek satır.

Geliştiriyorsan:

```sh
go build -o re-presence .
./re-presence -pair    # bir kez bağla
./re-presence          # çalıştır
./re-presence -once    # tek bildirim gönder, çık — hata ayıklamak için
```

`RE_PRESENCE_TOKEN` ve `RE_PRESENCE_API` ortam değişkenleri
`~/.config/re-presence/config.json` dosyasını ezer.

## Nasıl çalışıyor

Odaktaki pencerenin başlığını otuz saniyede bir okur. Editör açık ama odakta
değilse son bildirimi tekrarlar; editör kapanınca kartı indirir. Erişilebilirlik
izni istemez, tuş vuruşu okumaz, dosya izlemez.

macOS'ta AppleScript, Linux'ta `xdotool` (X11), Windows'ta Win32.

## Which editors it recognises

Zed, Antigravity, VS Code, Cursor, Windsurf, Xcode, Sublime Text, Neovide,
GoLand, WebStorm, IntelliJ, PyCharm, Android Studio, Rider, CLion, Emacs, Nova.
Anything else focused is treated as "not working" and nothing is sent. Adding
one is a line in `editors` in `main.go`.

VS Code and its forks are better served by the extension in `../vscode` — it
knows the line number, the language and whether a debug session is running,
none of which a title bar says.
