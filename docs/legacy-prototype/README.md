# IML x Sapago

Frontend prototype lokal **Its My Live** untuk User Brand/Client, berdasarkan brief versi 1.1 (7 Oktober 2026). React + TypeScript + Vite. Tidak memakai backend, credential, autentikasi produksi, atau koneksi SapaGo.

## Jalankan

```sh
npm install
npm run dev
```

Buka http://127.0.0.1:5178. Port dibuat tetap agar alamat preview tidak berubah.

```sh
npm run build
npm test
npm run preview
```

## Yang tersedia

- Ringkasan, laporan livestream, detail room dan slot, performa host dan detail host, serta unduhan sesi.
- Filter tanggal, cakupan, channel, studio dan host; tombol Terapkan/Reset, pencarian, urutan, pagination, dan kolom tambahan.
- Data sumber 26 September 2026. GMV room Rp15.484.967 tetap terpisah dari slot Linda yang belum memiliki atribusi hasil.
- Switch dataset simulasi 14 hari, disimpan di localStorage. Seluruh KPI, grafik, leaderboard, chat dan ekspor mengikuti dataset/filter aktif.
- CSV nyata untuk ringkasan, detail dan leaderboard; unduh ulang melalui Laporan Unduhan. XLSX nonaktif karena belum didukung.
- Chat lokal dengan fixture jawaban, reset saat dataset/filter berubah; fallback untuk pertanyaan lain.
- Loading skeleton, empty state, data parsial, simulasi gangguan dan retry, profil dan keluar simulasi.
- UI desktop/tablet/mobile, focus ring, modal keyboard dengan Escape dan focus trap, chart dengan deskripsi dan nilai aksesibel.

## Struktur

- `src/data.ts`: fixture sumber/simulasi, filter, agregasi, ranking, formatter dan CSV.
- `src/components.tsx`: kartu, tooltip, chart SVG, bar chart, overlay dan empty state.
- `src/App.tsx`: navigasi, state dan alur dashboard.
- `src/main.tsx`: entry React; `src/context.ts`: konteks label dataset.
- `src/styles.css`: token visual dan layout responsif.
- `src/data.test.ts`: tes atribusi, filter, ranking dan ekspor.
- `BRIEF.md`: salinan brief sumber.

## Catatan data

Room sumber tidak memiliki studio terkonfirmasi. Memilih Studio 1 pada cakupan room tidak mengatribusikan GMV room ke Studio 1; hasil menjadi kosong. Studio 1 tersedia pada slot Linda. Pembeli lintas laporan, snapshot viewers, AOV sumber dan rasio sumber tidak dijumlahkan. Pada simulasi metrik tanpa basis yang sesuai dibiarkan kosong. Tidak ada pembanding periode.

Screenshot sumber belum disertakan; tab Bukti laporan menampilkan “Bukti gambar belum dimuat”. Tidak ada gambar bukti buatan atau timestamp publikasi palsu. Font Inter dimuat opsional dari Google Fonts dengan fallback system sans-serif; semua data dan interaksi tetap lokal.
