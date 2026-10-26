# Kontrak screenshot operator — 22 Oktober 2026

Upload baru memakai `tiktok-screenshot-v1` atau `shopee-screenshot-v1`.
Laporan lama mempertahankan mapping aslinya. Migrasi 006 tersedia juga dalam
`database/schema.sql` untuk instalasi lokal baru.

## Data admin

Form menerima teks Tanggal, Studio, Channel, Brand, ID-Line, Host, Jam Start-End.
Tanggal Indonesia, `21okt 2026`, `7.00 10.00`, `15:00-18:00` dan pergantian tanggal
didukung. Tanggal kalender, jam dan prefix tanggal ID-Line diperiksa. Master harus
aktif, cocok secara unik dan dapat diakses. Data belum dikenal tidak dibuat atau
dipilih secara perkiraan. Jadwal host tidak otomatis mengisi periode sumber.

## Mapping dan formula

- TikTok: Perolehan GMV=`sales` (Indirect internal), Perkiraan GMV=`estimated_sales`
  (Direct internal); Tayangan=`views`, Impresi=`impressions`, Penonton saat ini=
  `current_viewers`. Tidak menganggap Direct/Indirect internal sebagai atribusi
  ≤24 jam / LTV terverifikasi V2.1.
- Shopee: Dilihat=`views`, Penonton=`viewers`, aktif/tertinggi disimpan terpisah;
  Penjualan=`sales`, Tambah ke Keranjang=`add_to_cart`, Pesanan per Klik=
  `source_ctor`, Persentase Klik=`source_click_rate`.
- Ekstraktor mempertahankan bukti, label asli, angka mentah dan penanda K/M
  dibulatkan. Field hilang null, bukan nol. Klik tidak ditebak dari persentase
  atau tabel produk yang hanya terlihat sebagian. ATC tidak diganti klik.
- API detail dan ringkasan dashboard brand berizin mengirim `screenshot`: conversion sumber,
  CTR/CTOR hitung ulang, follow/tayangan, tayangan/impresi, Watch Time,
  Comment Velocity, Cart→Order, AOV, serta GMV/1.000 impresi TikTok dan
  penjualan/1.000 Dilihat Shopee. Penyebut diberi nama sesuai sumber.
- Rasio periode menggunakan subset pembilang/penyebut yang sama dengan bobot
  denominator. Watch periode memakai bobot Tayangan/Dilihat sumber dan diberi
  keterangan, bukan klaim unique viewers. Tidak menormalisasi ulang bobot skor
  V2.1 untuk menutupi metrik yang tidak tersedia.
- Durasi memakai periode sumber, tidak memakai durasi host untuk total room.
  Jika durasi tertulis berbeda dari rentang timestamp, Comment Velocity N/A.

## AI dan publikasi

Worker menerima metadata operator dari database sebagai konteks, bukan bukti
atribusi. Observasi periode/basis/durasi AI tersimpan di extraction run dan tampil
pada detail laporan. Server memeriksa timestamp, basis, durasi dan perbedaan dengan
slot. Perbedaan periode host tetap memblokir publikasi host_slot; tidak membagi
total room secara proporsional terhadap jam host. Reviewer memeriksa bukti dan
publikasi tetap terpisah.

Contoh TikTok Nurul: periode sumber 06:42:38–18:03:33 bukan slot 15–18; durasi
tertulis 40.846 detik berbeda 9 detik dari rentang timestamp. Contoh Shopee Seli:
durasi 10.816 detik, Cart→Order=8/77=10,38961%, penjualan/1.000 Dilihat=
2.025.300/768×1.000=2.637.109,375. Keduanya fixture perhitungan; bukan data yang
diimpor ke database operasional.

OpenClaw lokal harus dikonfigurasi secara terpisah. Uji kontrak HTTP dan fixture
tidak membuktikan akurasi vision pada provider nyata.

## Verifikasi lokal

- Migrasi 006 diterapkan pada database lokal; instalasi SQL portabel diperbarui.
- Unit Go dan integrasi PostgreSQL lulus, termasuk metadata worker, profil baru,
  review periode berselisih dan respons perhitungan screenshot.
- 43 tes frontend, TypeScript, ESLint, build produksi dan Go vet lulus.
- Browser: form mengenali master demo, mengisi akun/studio/host/ID-Line/jam,
  mempertahankan periode sumber kosong, dan memilih profil screenshot TikTok.
- Server lokal pada pengujian masih `EXTRACTION_MODE=manual`; koneksi/akurasi
  vision OpenClaw nyata belum diuji dan tidak diganti respons mock.

## Upload sederhana (migrasi 007)

Upload UI mengirim `auto_metadata=true`. Backend memilih profil screenshot sesuai
channel; pilihan mapping, basis, periode dan konfirmasi dari client diabaikan.
Periode sumber tidak disalin dari jadwal operator. TikTok memakai Pesanan
Teratribusi; Shopee menunggu status pesanan pada bukti (bukan default Dibuat).
Sumber baru sementara berkategori room. Hanya observasi host_slot dengan periode
persis sama yang dapat menjadi host_slot; snapshot tetap sementara.

Worker menyimpan observasi ke metadata sumber dan versi draft hanya jika revision
input belum berubah. Observasi berkonflik/tidak terbaca tetap memerlukan review.
Periode terkonfirmasi otomatis jika terbaca, memiliki bukti, dan tidak ada konflik.
Mapping tidak perlu dipilih uploader. Detail daftar metrik berada dalam disclosure
opsional pada preview. Reviewer dapat mengoreksi basis/cakupan dan periode sebelum
publikasi pertama; metadata sumber yang pernah dipublikasikan tetap dilindungi.
Laporan dengan basis Shopee belum terbaca tidak dapat dipublikasikan.

Verifikasi: 47 tes frontend, Go unit/integrasi PostgreSQL, lint, typecheck, build
produksi dan vet. Integrasi menguji upload Shopee otomatis, observasi worker, dan
koreksi reviewer. Pengujian worker memakai fixture; gateway OpenClaw nyata belum
terkonfigurasi pada server lokal (masih manual), sehingga ekstraksi vision nyata
memerlukan konfigurasi gateway dan worker sebelum upload otomatis dapat disimpan.
