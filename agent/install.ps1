# re:presence — Windows kurulumu.
#
#   irm https://remakeizmir.com/presence.ps1 | iex
#
# İndirir, bağlar, oturum açılışında çalışacak şekilde ayarlar. Yönetici
# yetkisi istemez: her şey kullanıcının kendi klasörüne kurulur.
#
# Kaldırmak için:
#   & "$env:LOCALAPPDATA\re-presence\uninstall.ps1"

$ErrorActionPreference = 'Stop'

$repo   = 'remakeizmir/re-presence'
$api    = if ($env:RE_PRESENCE_API) { $env:RE_PRESENCE_API } else { 'https://api.remakeizmir.com/api/v1' }
$prefix = Join-Path $env:LOCALAPPDATA 're-presence'
$binary = Join-Path $prefix 're-presence.exe'

function Bilgi($metin) { Write-Host "  $metin" -ForegroundColor Cyan }
function Hata($metin)  { Write-Host "  $metin" -ForegroundColor Red }

Write-Host ''
Write-Host '  re:presence' -ForegroundColor White
Write-Host '  Editöründe ne üzerinde çalıştığın hub profilinde görünür.'
Write-Host ''

# ── Hangi makine ────────────────────────────────────────────────────────────
$mimari = switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { 'amd64' }
  'ARM64' { 'arm64' }
  default { $null }
}
if (-not $mimari) {
  Hata "Bu işlemci ($env:PROCESSOR_ARCHITECTURE) desteklenmiyor."
  return
}

# ── İndir ───────────────────────────────────────────────────────────────────
New-Item -ItemType Directory -Force -Path $prefix | Out-Null
$indirme = "https://github.com/$repo/releases/latest/download/re-presence-windows-$mimari.exe"

Bilgi 'İndiriliyor…'
try {
  # Bir önceki sürüm çalışıyorsa dosya kilitli olur; önce onu durdur.
  Get-Process -Name 're-presence' -ErrorAction SilentlyContinue | Stop-Process -Force
  Invoke-WebRequest -Uri $indirme -OutFile $binary -UseBasicParsing
} catch {
  Hata "İndirilemedi: $indirme"
  Hata $_.Exception.Message
  return
}

# ── Bağla ───────────────────────────────────────────────────────────────────
# İlk çalıştırma altı haneli kodu gösterir, tarayıcıyı açar ve onay bekler.
$env:RE_PRESENCE_API = $api
& $binary -pair
if ($LASTEXITCODE -ne 0) {
  Hata 'Bağlanamadı — tekrar denemek için komutu yeniden çalıştır.'
  return
}

# ── Açılışta çalıştır ───────────────────────────────────────────────────────
# Görev Zamanlayıcı yerine Başlangıç klasörü: yönetici yetkisi istemiyor ve
# kaldırması bir dosya silmek kadar kolay.
$baslangic = [Environment]::GetFolderPath('Startup')
$kisayol   = Join-Path $baslangic 're-presence.lnk'

$launcher = Join-Path $prefix 'start.vbs'
@"
' Pencere açmadan başlatır — arka planda çalışan bir şey için konsol
' penceresi her açılışta göze batar.
Set kabuk = CreateObject("WScript.Shell")
kabuk.Environment("PROCESS").Item("RE_PRESENCE_API") = "$api"
kabuk.Run """$binary""", 0, False
"@ | Set-Content -Path $launcher -Encoding ASCII

$shell = New-Object -ComObject WScript.Shell
$link = $shell.CreateShortcut($kisayol)
$link.TargetPath = 'wscript.exe'
$link.Arguments  = """$launcher"""
$link.WorkingDirectory = $prefix
$link.Description = 're:presence'
$link.Save()

@"
Get-Process -Name 're-presence' -ErrorAction SilentlyContinue | Stop-Process -Force
Remove-Item -Force -ErrorAction SilentlyContinue '$kisayol'
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '$prefix'
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue "`$env:USERPROFILE\.config\re-presence"
Write-Host 're:presence kaldırıldı.'
"@ | Set-Content -Path (Join-Path $prefix 'uninstall.ps1') -Encoding UTF8

# Şimdi de başlat, yeniden oturum açmayı beklemeden.
Start-Process -FilePath 'wscript.exe' -ArgumentList """$launcher""" -WindowStyle Hidden

Write-Host ''
Write-Host '  Hazır.' -ForegroundColor Green
Write-Host '  Editörünü aç, birkaç saniye sonra profilinde görünür.'
Write-Host "  Kaldırmak istersen: & `"$prefix\uninstall.ps1`""
Write-Host ''
