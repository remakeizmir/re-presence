#!/usr/bin/env bash
# The installer is served from remakeizmir.com/presence.sh, which is a file in
# the landing app's public folder. This copies it there; run it after changing
# install.sh, and commit both.
set -euo pipefail
kaynak="$(dirname "$0")/agent/install.sh"
hedef="$(dirname "$0")/../landing/public/presence.sh"
cp "$kaynak" "$hedef"
echo "kopyalandı → $hedef"

cp "$(dirname "$0")/agent/install.ps1" "$(dirname "$0")/../landing/public/presence.ps1"
echo "kopyalandı → landing/public/presence.ps1"

# The extension is downloaded the same way, for anyone whose editor is not in a
# marketplace we publish to — or before we publish at all.
vsix="$(ls "$(dirname "$0")"/vscode/*.vsix 2>/dev/null | head -1)"
if [ -n "$vsix" ]; then
  cp "$vsix" "$(dirname "$0")/../landing/public/re-presence.vsix"
  echo "kopyalandı → landing/public/re-presence.vsix ($(basename "$vsix"))"
else
  echo "uyarı: .vsix yok — önce vscode/ içinde 'npm run package'" >&2
fi
