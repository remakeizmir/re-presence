# re:make presence — VS Code

Editöründe ne üzerinde çalıştığın re:make hub profilinde görünür: proje, dosya,
satır ve ne zamandır çalıştığın.

Kurduktan sonra çıkan bildirimde **Bağlan**'a bas. Tarayıcı açılır, tek düğmeye
basarsın, biter. Kopyalanacak anahtar yok.

## Ne gönderilir

Klasörün **adı**, dosyanın **adı**, satır numarası, dil ve hata ayıklama
oturumu olup olmadığı. Yol asla gönderilmez —
`/Users/can/is/gizli-musteri/src/main.go` yalnızca `main.go` olarak gider.

Hiçbir şey kaydedilmez: sunucu son bildirimi iki dakika tutar, sonra unutur.

## Ayarlar

| | |
|---|---|
| `remake.enabled` | Kapatınca hiçbir şey gönderilmez. Durum çubuğundaki `re:make`'e tıklayarak da açılıp kapanır. |
| `remake.privateProjects` | Bu klasör adlarında dosya adı gönderilmez; kartta yalnızca "gizli bir proje" yazar. |
| `remake.apiUrl` | Yerel geliştirme için `http://localhost:4000/api/v1`. |

## Geliştirme

```sh
npm install
npm run compile
npm run package   # re-presence-0.1.0.vsix üretir
```

Cursor, Windsurf ve Antigravity de VS Code türevi — aynı .vsix hepsine kurulur.
