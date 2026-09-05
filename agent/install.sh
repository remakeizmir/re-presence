#!/usr/bin/env bash
#
# re:presence — tek satırlık kurulum.
#
#   curl -fsSL https://remakeizmir.com/presence.sh | bash
#
# İndirir, bağlar, açılışta çalışacak şekilde ayarlar. Terminali bir daha
# açmana gerek kalmaz.
#
# Kaldırmak için:
#   ~/.local/share/re-presence/uninstall.sh

set -euo pipefail

REPO="remakeizmir/re-presence"
API="${RE_PRESENCE_API:-https://api.remakeizmir.com/api/v1}"
PREFIX="$HOME/.local/share/re-presence"
BIN="$PREFIX/re-presence"

renk() { printf "\033[%sm%s\033[0m\n" "$1" "$2"; }
bilgi() { renk "0;36" "  $1"; }
hata()  { renk "0;31" "  $1" >&2; }

echo
renk "1" "  re:presence"
echo "  Editöründe ne üzerinde çalıştığın hub profilinde görünür."
echo

# ── Hangi makine ────────────────────────────────────────────────────────────
isletim="$(uname -s)"
mimari="$(uname -m)"

case "$isletim" in
  Darwin) platform="darwin" ;;
  Linux)  platform="linux" ;;
  *) hata "Bu sistem ($isletim) desteklenmiyor."; exit 1 ;;
esac

case "$mimari" in
  arm64|aarch64) mimari="arm64" ;;
  x86_64|amd64)  mimari="amd64" ;;
  *) hata "Bu işlemci ($mimari) desteklenmiyor."; exit 1 ;;
esac

if [ "$platform" = "linux" ] && ! command -v xdotool >/dev/null 2>&1; then
  hata "xdotool gerekiyor: sudo apt install xdotool"
  hata "(Wayland kullanıyorsan pencere başlığı okunamıyor, çalışmaz.)"
  exit 1
fi

# ── İndir ───────────────────────────────────────────────────────────────────
mkdir -p "$PREFIX"
indirme="https://github.com/$REPO/releases/latest/download/re-presence-$platform-$mimari"

bilgi "İndiriliyor…"
if ! curl -fsSL "$indirme" -o "$BIN.indiriliyor"; then
  # Yayınlanmış sürüme ulaşılamadı: Go varsa kaynaktan derle.
  # Piped through bash, $0 is "bash" — the source fallback only applies when
  # this file was cloned and run from inside the repository.
  betik_dizini="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd || echo .)"
  if command -v go >/dev/null 2>&1 && [ -f "$betik_dizini/main.go" ]; then
    bilgi "Hazır sürüm alınamadı, kaynaktan derleniyor…"
    (cd "$betik_dizini" && go build -o "$BIN" .)
  else
    hata "İndirilemedi: $indirme"
    hata "İnternet bağlantını kontrol et; sürmezse hub'dan bize yaz."
    exit 1
  fi
else
  mv "$BIN.indiriliyor" "$BIN"
fi
chmod +x "$BIN"

# ── Bağla ───────────────────────────────────────────────────────────────────
# İlk çalıştırma altı haneli kodu gösterir ve onaylanmasını bekler.
RE_PRESENCE_API="$API" "$BIN" -pair

# ── Açılışta çalıştır ───────────────────────────────────────────────────────
if [ "$platform" = "darwin" ]; then
  plist="$HOME/Library/LaunchAgents/com.remakeizmir.presence.plist"
  mkdir -p "$(dirname "$plist")"
  cat > "$plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.remakeizmir.presence</string>
  <key>ProgramArguments</key>
  <array><string>$BIN</string></array>
  <key>EnvironmentVariables</key>
  <dict><key>RE_PRESENCE_API</key><string>$API</string></dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardErrorPath</key><string>$PREFIX/log</string>
</dict>
</plist>
PLIST
  launchctl unload "$plist" >/dev/null 2>&1 || true
  launchctl load "$plist"

  cat > "$PREFIX/uninstall.sh" <<UNINSTALL
#!/usr/bin/env bash
launchctl unload "$plist" 2>/dev/null || true
rm -f "$plist"
rm -rf "$PREFIX" "\$HOME/.config/re-presence"
echo "re:presence kaldırıldı."
UNINSTALL

else
  unit="$HOME/.config/systemd/user/re-presence.service"
  mkdir -p "$(dirname "$unit")"
  cat > "$unit" <<UNIT
[Unit]
Description=re:presence

[Service]
ExecStart=$BIN
Environment=RE_PRESENCE_API=$API
Restart=always

[Install]
WantedBy=default.target
UNIT
  systemctl --user daemon-reload
  systemctl --user enable --now re-presence.service

  cat > "$PREFIX/uninstall.sh" <<UNINSTALL
#!/usr/bin/env bash
systemctl --user disable --now re-presence.service 2>/dev/null || true
rm -f "$unit"
rm -rf "$PREFIX" "\$HOME/.config/re-presence"
echo "re:presence kaldırıldı."
UNINSTALL
fi

chmod +x "$PREFIX/uninstall.sh"

echo
renk "1;32" "  Hazır."
echo "  Editörünü aç, birkaç saniye sonra profilinde görünür."
echo "  Kaldırmak istersen: $PREFIX/uninstall.sh"
echo
