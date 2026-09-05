# re:presence

Editöründe ne üzerinde çalıştığın, re:make profilinde görünür. Hangi proje,
hangi dosya, kaçıncı satır ve ne zamandır.

<!-- Bir topluluk aracının README'si iki soruya cevap vermeli: bu ne, ve nasıl
     kurarım. Üçüncü bir soru varsa — "kim yapıyor bunu" — cevabı üstte olmalı,
     altta değil. -->

## re:make

İzmir'de başlayan bağımsız bir geliştirici topluluğu. Beraber tasarlıyor,
geliştiriyor ve yeniden yapıyoruz. İzmir bizim için bir pilot bölge; işleyeni
başka şehirlere de taşımak niyetindeyiz.

Topluluk kendi platformunda buluşuyor: [remakeizmir.com](https://remakeizmir.com).
Orada sohbet odaları, projeler, etkinlikler ve herkesin bir profili var.

## Bu ne işe yarıyor

Bir odada kimin ne yaptığını görmek, o odayı yaşayan bir yer yapıyor. Birinin
kartında "server üzerinde çalışıyor, 2 saattir" yazması, ona ne zaman soru
sorulacağını — ya da ne zaman rahat bırakılacağını — söylüyor.

Gönderilen şey: editörün adı, klasörün **adı**, dosyanın **adı**, satır numarası
ve hata ayıklama oturumu olup olmadığı.

Gönderilmeyen şey: **yol**. `/Users/can/is/gizli-musteri/src/main.go` yalnızca
`main.go` olarak gider, proje adı da `gizli-musteri` olur. Sunucu ne gönderirsen
gönder son parçaya indirir.

Saklanan şey: **hiçbir şey**. Son bildirim iki dakika bellekte durur, sonra
unutulur. Dün ne yaptığını soracak bir yer yok.

Kart, editörün **açık olduğu sürece** durur — Discord'daki gibi, o pencereye
bakıyor olman gerekmiyor. Editörü kapatınca iner.

## Kurulum

Önce bir kez: hub → **Ayarlar** → **Bağlantılar** → **Kod editörü**. Aşağıdaki
adımlar sana altı haneli bir kod gösterecek, onu orada onaylayacaksın. Kopyalanıp
yapıştırılan uzun bir anahtar yok.

### VS Code · Cursor · Windsurf · Antigravity

Eklentiyi kur, çıkan bildirimde **Bağlan**'a bas. Tarayıcı kodu hazır gelmiş
sayfada açılır, tek düğmeye basarsın.

| Editör | Nereden |
|---|---|
| VS Code | [Marketplace](https://marketplace.visualstudio.com/items?itemName=remake.re-presence) |
| Cursor, Windsurf, Antigravity, VSCodium | [Open VSX](https://open-vsx.org/extension/remake/re-presence) |
| Hepsi | [.vsix indir](https://remakeizmir.com/re-presence.vsix) → Eklentiler → ⋯ → VSIX'ten Yükle |

Eklenti daha fazlasını bilir: satır numarası, dosyanın dili, hata ayıklama
oturumu. Ayrıca `remake.privateProjects` ayarına yazdığın klasörlerde dosya adı
göndermez, kartta yalnızca "gizli bir proje" yazar.

### Zed · Xcode · JetBrains · Sublime · diğerleri

Bu editörler bir eklentinin arka planda ağ kullanmasına izin vermiyor — Zed'in
eklentileri WebAssembly, Xcode'da böyle bir API hiç yok. Onların yerine küçük
bir program çalışıyor: odaktaki pencerenin başlığını okur, o kadar. Klavye
dinlemez, dosya izlemez.

```sh
# macOS · Linux
curl -fsSL https://remakeizmir.com/presence.sh | bash
```

```powershell
# Windows — PowerShell, yönetici olarak açmana gerek yok
irm https://remakeizmir.com/presence.ps1 | iex
```

İndirir, kodu gösterir, bilgisayar her açıldığında kendi başlar.

**Zed'e özel bir not:** Zed pencerelerini kendi çiziyor ve işletim sistemi
onların başlığını okuyamıyor. O yüzden Zed için proje ve dosya adı, Zed'in
kendi yerel veritabanından okunuyor — salt okunur, yalnız "hangi çalışma alanı
açık" ve "son hangi dosyadaydı" bilgisi. `sqlite3` komutu gerekiyor; macOS'ta
kurulu geliyor.

Tanıdığı editörler: Zed, Antigravity, VS Code, Cursor, Windsurf, Xcode, Sublime
Text, Neovide, GoLand, WebStorm, IntelliJ, PyCharm, Android Studio, Rider, CLion,
RustRover, PhpStorm, Visual Studio, Emacs, Nova. Odakta başka bir şey varken
hiçbir şey göndermez.

Linux'ta X11 ve `xdotool` gerekiyor. Wayland'da pencere başlığı okunamıyor —
orada VS Code eklentisi tek yol.

## Kaldırmak

Tamamen silmeden de durdurabilirsin: hub → Ayarlar → Bağlantılar → **Gizle**.
Ya da o cihazın anahtarını sil, o makine anında susar.

```sh
# macOS · Linux
~/.local/share/re-presence/uninstall.sh
```

```powershell
# Windows
& "$env:LOCALAPPDATA\re-presence\uninstall.ps1"
```

Eklenti için: durum çubuğundaki **re:make**'e tıkla, kapanır. Kaldırmak
istersen editörün eklenti listesinden.

## Kendi editörün

Tek bir POST isteği — [API.md](API.md). Hub'daki ayarlar sayfasından elle bir
anahtar oluşturup yirmi satırla kendi entegrasyonunu yazabilirsin.

## Depo

```
agent/    Zed ve diğerleri için küçük program (Go, bağımlılıksız)
vscode/   VS Code ailesi için eklenti (TypeScript)
API.md    Kendi aracını yazmak isteyenler için
```

MIT.
