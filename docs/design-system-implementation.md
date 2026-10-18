# Its My Live — penerapan Color Design System v1.0

Acuan: Color Design System ItsMyLive.md, 11 Oktober 2026. Tema terang.

## Struktur

- `frontend/src/styles/tokens.css`: empat warna inti persis sesuai brief, turunan brand, netral, feedback, grafik, dan alias untuk komponen lama.
- `frontend/src/styles/main.css`: layout dan micro-interaction yang ada; seluruh nilai warna menggunakan token.
- `frontend/src/styles/semantic.css`: pemetaan warna berdasarkan fungsi dan state komponen.
- Import sekali pada `frontend/src/app/main.tsx`. Chart.js membaca nilai CSS yang telah di-resolve, bukan mengirim `var()` ke canvas.
- `/design-system`: halaman contoh komponen terautentikasi, tidak mengubah data bisnis; tombol primary/secondary/accent/destructive, loading, disabled, required, invalid, lima tone feedback, workflow badge, zero/N/A, dan skeleton.

## Pemetaan fungsi

Kanvas cream, permukaan kartu/form/tabel/drawer putih, judul/angka slate gelap, navigasi aktif biru lembut dengan indikator biru gelap. Aksi utama dan tautan memakai blue.700; hover blue.800 dan pressed blue.900. Oranye inti dibatasi pada detail merek dan contoh CTA accent. Tidak ada teks putih di atas core blue.500/orange.500.

Success hijau untuk publikasi atau hasil aksi berhasil; error merah untuk kegagalan; warning amber untuk review/data parsial/stale insight; info biru untuk processing/ekstraksi/ready; neutral slate untuk draft/queue/archive/N/A/sampel terbatas. Badge memiliki teks dan ikon, tanpa live-region pada daftar workflow. Ekstraksi selesai tidak disamakan dengan validitas atau publikasi final. Blocker periode/definisi saja memakai warning; blocker scope/konflik/kelengkapan lain memakai error. Syarat backend tetap berlaku.

Grafik: Views biru gelap/circle, Interaksi ungu/triangle, Aktivitas teal/rect, Penjualan oranye gelap/diamond. Fill ringan, stroke/marker gelap, tooltip putih dengan outline jelas, label axis gelap, null tetap gap dan zero tetap angka. Tabel angka masih tersedia. Keempat KPI tetap putih dengan ukuran dan bobot setara.

Input putih dengan border.control, placeholder muted, focus ring biru 2px dengan offset putih 2px; error border merah gelap + teks/ikon, label tetap netral. Disabled menggunakan background/teks/border eksplisit tanpa opacity. Loading tetap warna tombol aktif dengan spinner dan aria-busy. Avatar memakai selected blue, status toggle memakai primary blue, bukan hijau success.

## Penyesuaian kontras berdasarkan konteks

Pemeriksaan menemukan text.muted (#64748B) di atas cream hanya 4.481:1 dan border.control (#82939C) terhadap cream 2.995:1. Keduanya tidak dibulatkan menjadi lulus. Teks di kanvas (footer, eyebrow, data-context) memakai `text-on-canvas` → text.secondary. Outline kontrol yang berada langsung di kanvas memakai `border-on-canvas` → border.strong. Teks kecil pada row selected memakai text.secondary. Token asli tetap tersedia untuk pasangan putih yang sesuai.

`python3 scripts/check-design-contrast.py` memeriksa 22 pasangan aktual dengan luminance sRGB. Hasil disimpan pada `docs/design-system-contrast.json`. Disabled dikecualikan dari klaim minimum teks aktif; warnanya tetap sesuai spesifikasi. Ini bukan klaim seluruh aplikasi telah memenuhi WCAG.

## Verifikasi

TypeScript, ESLint, production build, 28 tes frontend. Tes tambahan menjaga pemisahan badge published/succeeded/ready/failed/sample dan error field dari server. Browser memeriksa warna aktual kanvas/primary/loading/disabled/error/status, semua warna kategori grafik, keempat KPI setara, lebar halaman 360/768/1440 tanpa overflow, focus keyboard 2px, dan Escape pada dialog pengguna. Tabel panjang mempertahankan horizontal scroll lokal. Data, permission, rumus laporan, dan ranking tidak berubah.

Screenshot: `docs/screenshots/design-system-dashboard.png`, `design-system-components.png`, `design-system-mobile.png`.
