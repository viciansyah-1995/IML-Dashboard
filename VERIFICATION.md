# Pemeriksaan IML x Sapago

7 Oktober 2026.

## Hasil

- `npm run build`: TypeScript dan production build Vite berhasil.
- `npm test`: 5 tes lulus. Meliputi pemisahan room/slot Linda, tanggal/filter, total simulasi yang konsisten, ranking tie/zero/null/dedup/durasi, dan metadata CSV.
- Preview `http://127.0.0.1:5178/`: HTTP 200.
- Browser: halaman ringkasan sumber menampilkan GMV Rp15.484.967, 222 pesanan, 201 pembeli dan AOV sumber Rp69.750; chart satu titik.
- Browser: detail room dan detail Linda berhasil dibuka. Angka Linda tetap “—” dengan keterangan hasil khusus slot belum tersedia.
- Browser: CSV detail Linda berhasil diunduh sebagai `IML-x-Sapago_source_detail_2026-09-26_2026-09-26.csv`. Isi file dibaca kembali; ID Line dan Linda tersedia, GMV room 15484967 tidak ada. Riwayat unduhan dimulai dari aksi ini.
- Browser: simulasi menghasilkan 28 slot, 56 jam slot, GMV Rp74.798.000. Host A Rp32.885.000 (10 slot/20 jam); Host C Rp24.516.000 (9/18); Host B Rp17.397.000 (9/18). Kontribusi total sekitar 100% setelah pembulatan.
- Browser: chat simulasi menyebut periode 13–26 September, cakupan Per slot host, label Data simulasi, GMV Rp74.798.000 dan 1.276 pesanan.

## Batas pemeriksaan

Layout tablet terlihat melalui screenshot browser. Breakpoint mobile, modal filter, dan ukuran teks diselesaikan dalam CSS/React, tetapi pemeriksaan visual mobile terakhir belum terkonfirmasi: kontrol browser tidak lagi merespons/sinkron pada akhir pengujian. Build terakhir dan tes data tetap lulus. Screenshot bukti asli tidak tersedia; tab Bukti laporan memakai empty state. XLSX belum didukung. Tidak ada API/backend atau koneksi produksi untuk diuji.
