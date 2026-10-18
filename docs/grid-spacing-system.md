# Grid & spacing Its My Live

Sumber aturan: `frontend/src/styles/layout.css`, dimuat setelah warna dan tipografi. Semua perubahan layout memakai token; perilaku API, permission, dan data tidak berubah.

## Skala

4, 8, 12, 16, 20, 24, 32, 40, 48 px. Gunakan 8 px untuk label ke input, 16–20 px untuk jarak field, 24 px untuk gutter desktop dan pemisah section, 32 px untuk ruang awal halaman. Padding kartu 24 px desktop / 16 px mobile. Kontrol utama minimal 44 px.

## Grid

Lebar konten maksimum 1440 px. Halaman memakai pembagian setara grid 12 kolom: KPI 3 kolom, panel berpasangan 6 kolom, section penuh 12 kolom. Track selalu `minmax(0, 1fr)` agar konten panjang tidak memperlebar halaman.

- Desktop >=1280 px: empat KPI; dua panel grafik; form dua kolom; gutter 24 px; inset halaman 32 px.
- 1024–1279 px: dua KPI; inset halaman 24 px; review bukti satu kolom.
- 768–1023 px: panel grafik bertumpuk; form tetap dua kolom.
- <=767 px: gutter/inset/padding kartu 16 px; form satu kolom; filter dalam drawer; dialog memenuhi viewport dengan header tetap dan body menggulir.
- <=379 px: KPI satu kolom agar angka mata uang panjang tidak terpotong.

Grid mengatur jarak antar field: jangan tambahkan margin bawah pada field yang sudah menjadi anak `form-grid`, `profile-form-grid`, atau `ds-fields`. Kartu di dalam grid tidak memiliki margin tambahan. Tabel menggulir di wrapper; kelompok aksi dapat membungkus di kolomnya.

Contoh visual tersedia di halaman `/design-system` setelah login. Screenshot hasil ada di `screenshots/grid-spacing-desktop.png` dan `screenshots/grid-spacing-mobile.png`.

## Verifikasi

33 tes frontend, TypeScript, ESLint, dan production build berhasil. Browser diperiksa pada dashboard, form pengguna, tabel, serta dropdown filter dalam dialog. Viewport 1440, 1024, 768, 390, dan 320 px diperiksa terhadap overflow halaman; form mobile menggunakan satu kolom.
