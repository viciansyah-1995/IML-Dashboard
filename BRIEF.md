# Brief UI Dashboard Brand dan Client Its My Live

Versi 1.1 — 7 Oktober 2026
Acuan: BRD Its My Live v0.2, screenshot LIVE zwitsal_id, dan lampiran Data Metric [Sirclo x Streamwise], sheet Performance Sirclo X Streamwise.

## 1. Tujuan pengerjaan

Buat prototype UI dashboard dari sisi **User Brand/Client** untuk melihat hasil livestream. Fokus pada visualisasi laporan, navigasi, filter, detail sesi, ekspor, dan panel tanya laporan. UI harus terasa seperti aplikasi operasional bisnis yang matang: rapi, mudah dibaca, dan tidak memakai estetika robot atau AI generik.

Deliverable tahap ini adalah **frontend prototype dengan data lokal**, tanpa backend, integrasi OpenClaw, autentikasi sungguhan, atau koneksi SapaGo. Semua interaksi utama harus dapat dicoba. Jangan membangun halaman Admin atau Superadmin.

Gunakan nama produk **Its My Live**. Tampilan utama berbahasa Indonesia. Pertahankan istilah metrik sumber dalam tooltip untuk memudahkan pencocokan dengan ledger.

## 2. Pengguna dan kebutuhan

Pengguna adalah owner brand, marketing, ecommerce manager, dan PIC client. Pertanyaan utama mereka:

1. Berapa hasil penjualan dan pesanan dari livestream pada periode terpilih?
2. Bagaimana jangkauan, klik produk, dan interaksi penonton?
3. Sesi mana yang menghasilkan performa baik?
4. Bagaimana hasil antar host, studio, dan channel jika datanya bisa dibandingkan?
5. Angka mana yang sudah tersedia, dan apa cakupan laporannya?
6. Bagaimana mengunduh laporan atau menanyakan insight?

User hanya melihat brand yang ditugaskan. Brand selector boleh tampil jika user memiliki beberapa brand. Pada prototype, gunakan satu brand aktif **Zwitsal / zwitsal_id**; jangan menciptakan client lain sebagai data aktual.

## 3. Prinsip data yang wajib diikuti

### 3.1 Pisahkan room LIVE dan slot host

Screenshot sumber menunjukkan hampir 12 jam LIVE, sekitar **06.00–18.00**. Ledger contoh mencantumkan host Linda dengan slot **16.00–18.00**. Angka screenshot belum dapat diatribusikan ke Linda.

Sediakan kontrol **Cakupan laporan**:

- **Sepanjang LIVE**: menampilkan metrik satu room/sumber LIVE.
- **Per slot host**: hanya menampilkan metrik yang terbukti sesuai rentang slot host.

Default dataset sumber memakai Sepanjang LIVE. Filter Host pada cakupan ini nonaktif dengan helper “Pilih Per slot host untuk melihat hasil per host”. Dalam detail, Linda tampil pada bagian **Slot host terkait**, bukan sebagai pemilik seluruh GMV.

### 3.2 Visual prototype dan status produksi

Prototype boleh memperlihatkan angka screenshot sebagai **preview sumber**, dengan label tetap “Pratinjau data contoh” di topbar. Ini bukan keputusan publikasi. Implementasi produksi hanya menampilkan data yang sudah dipublikasikan dan valid sesuai BRD.

Pada dataset contoh, tampilkan catatan singkat dekat filter: “Angka mengikuti laporan sepanjang LIVE. Hasil slot host belum tersedia.” Angka draft tidak boleh masuk total final secara diam-diam.

### 3.3 Jangan mengarang histori

Lampiran hanya berisi dua baris contoh terisi. Baris 1 September belum memiliki cakupan sumber yang tervalidasi; jangan menggabungkannya dengan contoh 26 September untuk menghasilkan tren aktual.

- Dataset sumber: satu titik data tanggal 26 September. Chart tren menggunakan satu marker/bar dan teks “Butuh minimal 2 tanggal untuk melihat tren”.
- Opsional, untuk menilai visual chart: sediakan switch **Contoh visual 14 hari** yang memakai data sintetis dan label global “Data simulasi”. Semua total dihitung dari dataset itu. Jangan mencampur data sintetis dengan angka sumber.
- Label simulasi harus ikut pada tooltip, detail, dan ekspor. Pilihan dataset tersimpan dalam state.
- Tidak ada persentase kenaikan, target, atau perbandingan periode palsu. Jika pembanding tidak tersedia, tampilkan “Belum ada pembanding”.

### 3.4 Arti metrik

- Gunakan **Perolehan GMV** untuk angka screenshot; tooltip menjelaskan kolom sumber ledger P bernama Direct GMV, definisinya masih perlu dikonfirmasi.
- **Perkiraan GMV** ditampilkan terpisah, bukan dijumlahkan dengan Perolehan GMV dan bukan diberi label Indirect GMV tanpa kontrak definisi.
- **Penonton saat screenshot** adalah snapshot 128 orang; bukan total/unique viewers sesi.
- **AOV sumber** ditampilkan sebagai nilai dari screenshot. Jangan menyebutnya rata-rata harga produk sebelum definisi Avg. Price disepakati.
- Impresi dan Tayangan mengikuti label screenshot. Tooltip menyebut label PV pada ledger masih perlu dikonfirmasi.
- Angka K/M adalah nilai tampilan dibulatkan. Detail menunjukkan teks asli, misalnya 2.42M, agar presisinya jelas.
- Data kosong tampil “—” dan keterangan “Belum tersedia”; nol tetap tampil 0.

## 4. Arah visual

### Karakter

Dashboard bisnis dengan latar abu-abu hangat, kartu putih, garis pembatas tipis, tipografi kuat, serta grafik sederhana. Utamakan angka dan keterbacaan. Wordmark Its My Live cukup berupa teks; jangan membuat logo brand Zwitsal jika aset resmi belum tersedia.

### Token desain yang disarankan

| Elemen | Spesifikasi |
|---|---|
| Font | Inter; fallback system sans-serif |
| Latar aplikasi | #F6F7F9 |
| Permukaan kartu | #FFFFFF |
| Teks utama | #17212B |
| Teks pendukung | #667085 |
| Border | #E4E7EC |
| Aksen utama | Teal #0F766E |
| Chart sekunder | Biru #5271A3 |
| Peringatan | Amber #B7791F dengan latar #FFF8E8 |
| Positif | #16804A |
| Negatif | #C54444 |
| Radius | Kartu 12 px; tombol/input 8 px |
| Bayangan | Sangat tipis; gunakan border sebagai pemisah utama |
| Spacing | Kelipatan 4/8 px; antar bagian 24 px |
| Ukuran teks | Judul 24–28 px; isi 14 px; caption 12 px; KPI 28–32 px |

Ikon outline sederhana dan konsisten. Hindari gradient neon, glassmorphism, dekorasi robot, emoji sebagai navigasi, tombol sparkle besar, serta kartu yang semuanya memakai warna berbeda. Label status cukup teks dengan titik/ikon kecil, tidak memenuhi layar.

## 5. Struktur halaman dan navigasi

Sidebar desktop lebar sekitar 224 px, dengan menu:

1. Ringkasan
2. Laporan Livestream
3. Performa Host
4. Laporan Unduhan

Bagian bawah sidebar: akun user dan Keluar. Nama client dalam prototype boleh “Pengguna Brand” agar tidak mengarang identitas.

Topbar memuat brand selector, nama halaman, penanda pratinjau/simulasi, dan avatar. Notifikasi tidak perlu ditambahkan jika belum memiliki fungsi.

**Performa channel dan studio berada di Ringkasan**, sehingga navigasi tetap sederhana. Panel “Tanya Laporan” dibuka melalui launcher widget dan CTA pada insight, bukan halaman pengaturan AI.

## 6. Halaman Ringkasan

### Urutan layout desktop

1. Header: “Ringkasan Livestream”; subteks akun aktif dan periode.
2. Toolbar filter dan tombol “Unduh laporan”.
3. Catatan cakupan sumber yang ringkas, hanya saat relevan.
4. Empat kartu KPI utama.
5. Grafik tren penjualan di kiri (8 kolom) dan ringkasan audiens di kanan (4 kolom).
6. Empat kartu metrik engagement.
7. Grafik interaksi dan ringkasan efisiensi.
8. Leaderboard host berdasarkan omzet slot terverifikasi, lalu perbandingan channel/studio.
9. Ringkasan laporan dan tabel sesi terbaru.

Gunakan grid 12 kolom dan konten maksimum sekitar 1440 px. Bagian sekunder boleh menggunakan accordion “Metrik tambahan” agar layar pertama tidak terlalu padat. Semua metrik tetap dapat ditemukan melalui halaman detail.

### Filter

- Rentang tanggal: Hari ini, 7 hari terakhir, Bulan ini, dan Custom.
- Default dataset sumber: **26 September 2026**, bukan tanggal sekarang yang tidak memiliki data.
- Cakupan laporan: Sepanjang LIVE / Per slot host.
- Channel dan Studio.
- Host hanya aktif pada cakupan Per slot host.
- Tombol Terapkan dan Reset. Tampilkan chip filter aktif setelah diterapkan.
- Perbandingan periode opsional, default off; dinonaktifkan jika pembanding tidak tersedia.

Filter harus memperbarui kartu, grafik, tabel, dan payload ekspor secara konsisten. Tampilkan “1 laporan sumber” terpisah dari “1 slot host tercatat”. Jangan menyebut keduanya sebagai dua sesi menghasilkan penjualan.

### KPI utama

| Kartu | Nilai preview sumber | Visual dan konteks |
|---|---:|---|
| Perolehan GMV | Rp15.484.967 | Angka besar; label Sepanjang LIVE dan tooltip sumber |
| Pesanan teratribusi | 222 | Angka besar; label periode sumber |
| Pembeli | 201 | Angka besar; jangan klaim unique buyers lintas periode |
| AOV sumber | Rp69.750 | Angka besar; catatan nilai tampilan dibulatkan |

Di bawahnya tampilkan ringkasan kecil: 1 laporan LIVE, durasi sumber hampir 12 jam, 1 slot host tercatat dengan durasi 2 jam. Durasi sumber dan durasi host tidak boleh dipakai bergantian.

### Metrik audiens dan engagement

| Metrik | Preview sumber | Bentuk visual yang direkomendasikan |
|---|---:|---|
| Impresi | 2.420.000 | KPI kecil dan tren garis jika histori kompatibel tersedia |
| Tayangan | 61.520 | KPI kecil; tooltip label sumber |
| Klik produk | 3.820 | KPI kecil; tren batang/garis bila data cukup |
| Durasi tonton rata-rata | 15 detik | KPI kecil; bukan donut atau gauge |
| Penonton saat screenshot | 128 | Snapshot dengan waktu pengambilan; jangan agregasikan |
| Pengikut baru | 59 | Kartu engagement |
| Suka | 5.220 | Kartu engagement |
| Komentar | 392 | Kartu engagement |
| Dibagikan | 56 | Kartu engagement |

## 7. Pilihan diagram dan aturan penggunaannya

| Komponen | Diagram | Detail desain dan batas data |
|---|---|---|
| Tren perolehan GMV | Garis dengan marker | X tanggal, Y rupiah. Satu titik tetap satu titik; jangan menggambar kurva histori palsu. |
| Tren pesanan | Batang, melalui tab metrik pada chart tren | Hindari dua sumbu untuk angka uang dan jumlah. Tab GMV/Pesanan/Klik menjaga chart bersih. |
| Interaksi | Batang horizontal | Suka, Komentar, Dibagikan; tampilkan nilai. Karena suka dominan, jangan menghilangkan label angka kategori kecil. Pengikut baru ditampilkan terpisah. |
| Perbandingan channel | Batang horizontal GMV | Hanya metrik dengan definisi dan mata uang kompatibel. Jika satu channel, tampilkan satu bar atau kartu ringkas. |
| Perbandingan studio | Batang horizontal GMV + tabel ringkas | Sertakan jumlah sesi dan jam live pada tabel agar konteks terlihat. |
| Leaderboard omzet host | Batang horizontal total GMV host + tabel ranking | Urutkan omzet slot terverifikasi terbesar; GMV/jam sebagai metrik pendamping. Default contoh Linda belum memiliki hasil teratribusi, sehingga belum masuk ranking. |
| Efisiensi | Kartu angka | AOV sumber, rata-rata tonton, GMV/jam jika scope valid. Tidak memerlukan speedometer. |
| Produk terjual | KPI tambahan | 270 dari screenshot; bukan Orders. |
| Rasio sumber | Kartu kecil | CTR LIVE 6,21% dan CTOR 5,81%; tampilkan sebagai rasio sumber, bukan hitungan sendiri. |
| Perkiraan GMV | Kartu terpisah di metrik tambahan | Rp12.240.000; jangan dibuat stacked bar dengan GMV aktual. |
| Biaya dan ROI sumber | Dua kartu opsional | Biaya iklan Rp3.340.000 dan ROI GMV Max 5,03. Jangan mengganti nama menjadi ROAS hitungan aplikasi. |

**Jangan membuat funnel Impresi → Tayangan → Klik → Pesanan** sebelum definisi populasi dan basis periodenya diketahui. Hindari pie chart untuk GMV vs Perkiraan GMV. Donut hanya dipakai nanti jika ada komposisi saling eksklusif yang benar; dataset awal tidak memerlukannya.

Jangan mengekstrak deret waktu dari bentuk garis pada foto atau menebak persentase traffic/demografi. Traffic sources, demografi, dan product leaderboard menjadi placeholder opsional bertuliskan “Belum ada data terstruktur”, bukan visual terisi dari dugaan.

## 8. Halaman Laporan Livestream

Header memuat judul, jumlah laporan sesuai filter, kolom pencarian, dan tombol ekspor. Gunakan filter yang konsisten dengan Ringkasan.

Tabel default **Sepanjang LIVE**:

- Tanggal
- Brand/akun
- Channel
- Studio terkait jika diketahui
- Rentang laporan sumber
- Perolehan GMV
- Pesanan
- Klik produk
- Cakupan/status data
- Aksi “Lihat detail”

Tabel **Per slot host**:

- Tanggal
- ID Line
- Host
- Studio
- Channel
- Jam mulai–selesai
- Durasi slot
- GMV teratribusi jika tersedia
- Pesanan teratribusi jika tersedia
- Status hasil
- Aksi detail

Kolom lain dapat dibuka lewat “Atur kolom”. Jangan menampilkan 33 kolom sekaligus. Mendukung sort, pencarian ID Line/akun/host, pagination, dan empty state. Seluruh baris read-only; tidak ada edit, upload, atau publish.

Contoh slot Linda: tanggal 26 September 2026, ID 260926-00002, Studio 1, TikTok, 16.00–18.00, 2 jam. Kolom GMV/pesanan slot “—”; status “Hasil slot belum tersedia”. Jangan menyalin GMV room ke baris ini.

## 9. Detail laporan

Buka sebagai halaman detail dengan tombol Kembali dan breadcrumb. Pada desktop, informasi sumber berada di panel kanan; pada mobile ditampilkan setelah metrik.

Urutan konten:

1. Identitas laporan: akun, tanggal, channel, cakupan sumber, waktu mulai/akhir dan waktu screenshot jika diketahui.
2. KPI utama.
3. Tab **Penjualan**, **Audiens**, **Interaksi**, dan **Bukti laporan**.
4. Panel konteks: studio, slot host terkait, ID Line, dan catatan periode.
5. Tautan “Lihat screenshot” membuka viewer dengan zoom.
6. Daftar definisi metrik dan teks sumber asli.

Untuk contoh room, tampilkan “Sepanjang LIVE sekitar 06.00–18.00” dan slot Linda sebagai relasi. Waktu pengambilan screenshot belum ditentukan secara pasti dari laporan: jangan membuat timestamp publikasi atau update palsu. Gunakan “Waktu pembaruan belum tersedia” bila diperlukan.

Pada detail slot Linda, informasi jadwal tersedia tetapi angka performa belum teratribusi. Beri penjelasan: “Bukti yang tersedia mencakup seluruh LIVE. Hasil khusus slot ini belum tersedia.”

NO SKP, PIC DATA H+1, PIC DATA, Notes, Flag Data, dan Remarks Data termasuk metadata pendukung; tampilkan di bagian “Informasi laporan” hanya sesuai izin client. Flag teknis diterjemahkan menjadi bahasa pengguna. Kelengkapan metrik bukan label otomatis “Data valid”.

## 10. Halaman Performa Host

Sediakan filter periode, studio, dan channel. Susunan: ringkasan jumlah host/slot, leaderboard omzet host, chart batang horizontal total omzet per host, lalu tabel ranking lengkap. Ranking utama berdasarkan total omzet dari slot valid. GMV per jam tetap tersedia sebagai metrik pendamping dan pilihan sort, dengan label urutan yang berubah sesuai metrik pilihan.

Dataset sumber awal: tampilkan slot Linda dan durasinya, tetapi panel performa menunjukkan empty state. Copy: “Belum ada laporan per slot host yang dapat dibandingkan.” CTA: “Lihat slot livestream”.

Pada mode simulasi boleh menampilkan perbandingan beberapa host fiktif dengan penanda simulasi. Perbandingan harus menjelaskan jumlah slot dan durasi; jangan menampilkan skor prestasi atau menyatakan host menyebabkan kenaikan penjualan.

## 11. Ringkasan dan widget Tanya Laporan

Ringkasan menggunakan panel putih biasa dengan judul **Ringkasan laporan**. Tidak memakai icon robot, gradient, atau badge AI mencolok.

Contoh isi berbasis sumber:

- Perolehan GMV sepanjang LIVE pada sumber tercatat Rp15.484.967 dengan 222 pesanan teratribusi.
- Klik produk tampil 3.82K dan pengikut baru 59.
- Hasil khusus slot Linda belum tersedia karena cakupan screenshot lebih panjang dari slot host.

CTA “Tanya tentang laporan” membuka drawer chat kanan lebar sekitar 380 px. Launcher kecil di kanan bawah, tidak menutupi tombol atau footer.

Prompt contoh:

1. “Ringkas hasil laporan ini.”
2. “Berapa GMV dan pesanan pada periode ini?”
3. “Apakah performa Linda sudah tersedia?”

Prototype memakai balasan lokal yang sudah disiapkan; beri label **Simulasi percakapan**. Jawaban selalu menyebut cakupan dan periode. Prompt di luar fixture dijawab “Prototype ini belum terhubung ke layanan insight.” Jangan memasang token, endpoint palsu, atau menyebut layanan sudah tersambung.

Footer widget boleh memuat teks kecil **Didukung oleh SapaGo AI**. Saat ganti brand/dataset/filter, konteks chat direset atau ditandai tidak berlaku agar jawaban periode lama tidak terlihat sebagai jawaban periode baru.

## 12. Laporan Unduhan

Halaman sederhana berisi daftar unduhan dalam sesi prototype: tanggal dibuat, periode, cakupan, filter, format, dan status. Tidak mengarang riwayat unduhan sebelum user melakukan aksi.

Tombol “Unduh laporan” membuka modal berisi:

- Periode dan filter aktif.
- Cakupan Sepanjang LIVE atau Per slot host.
- Format CSV; XLSX hanya jika implementasi lokal mendukung.
- Pilihan Ringkasan atau Detail.
- Penanda data contoh/simulasi yang ikut pada file.

CSV harus benar-benar dapat diunduh dari data lokal yang sedang difilter. Jika XLSX belum dibuat, tampilkan opsi nonaktif dengan alasan singkat; jangan memberi toast berhasil untuk file yang tidak ada. Unduhan tidak mencampur nilai room dan slot host.

## 13. Dataset contoh lokal

Simpan fixture terpisah untuk laporan sumber dan slot host. Nilai null tetap null, bukan 0. Identitas teknis di bawah hanya ID lokal prototype.

```json
{
  "dataset": "source_preview",
  "isDemo": true,
  "brand": {"id": "demo-zwitsal", "name": "Zwitsal", "account": "zwitsal_id"},
  "roomReport": {
    "id": "demo-room-report-20260926",
    "date": "2026-09-26",
    "channel": "TikTok",
    "scope": "room",
    "sourceTimeLabel": "Sekitar 06.00–18.00 WIB",
    "reportedDurationSeconds": 43192,
    "capturedAt": null,
    "publishedAt": null,
    "productionPublicationStatus": "unverified",
    "metrics": {
      "sourceGmv": 15484967,
      "estimatedGmv": 12240000,
      "impressions": 2420000,
      "views": 61520,
      "attributedOrders": 222,
      "buyers": 201,
      "averageViewingSeconds": 15,
      "currentViewers": 128,
      "sourceAovIdr": 69750,
      "productClicks": 3820,
      "newFollowers": 59,
      "likes": 5220,
      "comments": 392,
      "shares": 56,
      "productsSold": 270,
      "sourceLiveCtr": 0.0621,
      "sourceCtor": 0.0581,
      "sourceGmvMaxRoi": 5.03,
      "adSpend": 3340000
    },
    "mappingIssues": ["direct_gmv_definition", "estimated_vs_indirect_gmv", "pv_definition", "viewer_definition", "aov_vs_average_price"]
  },
  "hostSlot": {
    "id": "demo-slot-260926-00002",
    "date": "2026-09-26",
    "idLine": "260926-00002",
    "studio": "Studio 1",
    "host": "Linda",
    "channel": "TikTok",
    "startTime": "16:00",
    "endTime": "18:00",
    "durationMinutes": 120,
    "relatedSourceReportId": "demo-room-report-20260926",
    "attributedMetrics": null,
    "resultStatus": "scope_mismatch"
  }
}
```

Raw labels yang dibulatkan: Perkiraan GMV 12.24M; Impresi 2.42M; Tayangan 61.52K; AOV 69.75K; Klik produk 3.82K; Suka 5.22K; Biaya iklan 3.34M. Label format Indonesia pada UI boleh ditampilkan sebagai 12,24 jt, 2,42 jt, 61,52 rb, dan seterusnya; angka lengkap tetap tersedia dalam tooltip/detail.

Nilai ROI 5,03 adalah nilai sumber. Jangan menghitungnya ulang dari sourceGmv/adSpend karena basisnya belum diketahui. Perkiraan GMV tidak digunakan untuk menghitung total revenue.

## 14. Cakupan semua field ledger

| Kelompok kolom sumber | Penempatan UI |
|---|---|
| A nomor urut; B status upload | Nomor tabel dan metadata sumber; status upload tidak berarti sudah terbit |
| C tanggal; E studio; F channel; G akun brand | Filter, identitas laporan, dan tabel |
| H ID Line; I host; J–K waktu; L durasi | Tabel slot host dan detail jadwal |
| D NO SKP; M PIC DATA H+1; N PIC DATA | Detail informasi laporan, bersifat opsional dan permission-based |
| O Kelengkapan Metrics | Informasi kelengkapan sumber; terpisah dari validitas/cakupan |
| P–Q GMV | Perolehan GMV dan Perkiraan GMV sumber; definisi final masih perlu dikunci |
| R–W audiens dan penjualan | Kartu, tab detail, dan tooltip definisi |
| X AOV/Avg. Price; Y klik | Kartu nilai dan tren jika datanya cukup |
| Z–AC followers dan interaksi | Kartu serta bar chart interaksi |
| AD screenshot | Viewer bukti dalam detail |
| AE notes; AF flag; AG remarks | Catatan dan penjelasan kualitas data dengan bahasa pengguna |

Screenshot file dapat disalin menjadi aset lokal prototype bila tersedia. Jangan memasukkan path absolut workstation sebagai URL browser. Jika gambar belum tersedia, pakai state “Bukti gambar belum dimuat”, bukan gambar buatan yang menyerupai bukti asli.

## 15. Responsive dan aksesibilitas

- Desktop ≥1280 px: sidebar tetap, KPI empat kolom, chart 8:4.
- Tablet 768–1279 px: sidebar ringkas/drawer, KPI dua kolom, grafik ditumpuk bila sempit.
- Mobile <768 px: header compact, menu drawer, filter modal/bottom sheet, KPI dua kolom dengan teks mudah dibaca; grafik full-width.
- Tabel boleh scroll horizontal dalam container, bukan seluruh halaman. Kolom identitas tetap terlihat atau gunakan kartu ringkas di mobile.
- Modal berubah full-screen pada mobile. Chat tidak menutupi navigasi utama.
- Tombol minimal target sentuh 44 px; focus ring terlihat; semua filter punya label dan dapat dioperasikan keyboard.
- Status dan chart tidak bergantung pada warna saja. Tooltip chart menampilkan nama metrik, tanggal, satuan, serta penanda contoh/simulasi.

## 16. State dan interaksi wajib

1. Loading skeleton untuk kartu, chart, dan tabel.
2. Empty state ketika tanggal/filter tidak memiliki data; tombol Reset filter.
3. Data parsial: tampilkan hanya nilai yang ada, jelaskan cakupan secara ringkas.
4. Grafik satu titik dan tanpa pembanding.
5. Hasil host belum tersedia.
6. Modal ekspor dan unduhan CSV nyata.
7. Detail room dan detail slot host.
8. Viewer screenshot dengan zoom jika aset tersedia.
9. Chat drawer dengan balasan simulasi sesuai konteks.
10. Simulasi gagal memuat data dengan tombol Coba lagi, melalui kontrol demo yang kecil dan terpisah.

Tidak perlu login page untuk tahap ini; masuk langsung sebagai User Brand. Avatar menu cukup Profil ringkas dan Keluar simulasi. Menu manajemen user, upload data, retry ekstraksi, atau edit dashboard tidak ditampilkan.

## 17. Arahan implementasi

Ikuti stack frontend project yang sudah tersedia. Jika project baru, boleh memakai React + TypeScript dengan CSS utility atau styling modular, ikon outline, dan library chart yang sesuai. Pemilihan library tidak menjadi syarat visual.

Pisahkan dataset, formatter rupiah/tanggal, fungsi filter, selector agregasi, komponen kartu, chart, tabel, dan drawer. Semua angka pada mode simulasi berasal dari satu dataset lokal yang sama. Jangan hardcode KPI yang tidak berubah saat filter diterapkan.

Tidak memerlukan database, API, model AI, atau credential. Jangan menampilkan kode, nama orchestrator, proses ekstraksi, maupun istilah teknis validasi sebagai alur utama pengguna brand.

## 18. Leaderboard host berdasarkan omzet

### Tujuan dan definisi

Tampilkan kontribusi penjualan setiap host berdasarkan **omzet yang dihasilkan pada slot livestream miliknya** dalam periode dan filter aktif. Leaderboard tersedia sebagai ringkasan Top 5 pada Ringkasan dan tabel lengkap pada Performa Host.

Dalam prototype, omzet menggunakan **GMV yang teratribusi ke slot host**, bukan omzet bersih, settlement, atau laba. Label visual: **Omzet live host**. Tooltip wajib: “GMV dari slot live terverifikasi. Belum dikurangi pembatalan, retur, atau biaya jika sumber tidak menyediakannya.” Jika basis sumber berbeda, label dan definisi harus mengikuti kontrak metrik yang disepakati pada BRD.

### Metrik leaderboard

| Metrik | Rumus atau aturan | Penempatan |
|---|---|---|
| Peringkat omzet | Urutan total GMV slot valid terbesar | Nomor ranking pada grafik dan tabel |
| Total omzet live | Jumlah GMV slot unik yang valid milik host pada periode aktif | Angka utama dan panjang batang |
| Kontribusi omzet host | Omzet host ÷ total omzet seluruh host yang memenuhi syarat pada filter aktif × 100% | Kolom persentase; penyebut nol menjadi — |
| Jumlah slot valid | Hitung ID slot unik yang memiliki GMV valid | Kolom konteks |
| Jam live dengan data omzet | Total durasi slot yang dipakai dalam total omzet | Kolom konteks |
| Omzet per jam | Total omzet ÷ jam live dari kumpulan slot yang sama | Metrik efisiensi pendamping |
| Rata-rata omzet per slot | Total omzet ÷ jumlah slot valid | Detail host atau kolom opsional |
| Pesanan teratribusi | Jumlah pesanan dari slot yang memiliki metrik pesanan valid | Kolom opsional dengan jumlah slot tercakup |
| AOV terhitung | GMV ÷ pesanan dari subset slot yang memiliki kedua metrik dengan basis sama | Detail host; jangan mencampur subset berbeda |
| Cakupan data omzet | Slot dengan GMV valid ÷ seluruh slot host yang tercatat pada filter aktif | Kolom kualitas data |

Persentase kontribusi hanya dibandingkan terhadap omzet yang sudah teratribusi pada host, **bukan total room LIVE**. Tampilkan caption “Kontribusi terhadap omzet host terverifikasi”. Hasil dapat parsial jika ada slot belum memiliki data.

### Aturan ranking

1. Hanya slot yang sudah dipublikasikan, periode metriknya sesuai, dan atribusi hostnya valid yang boleh masuk ranking produksi. Mode prototype simulasi memakai fixture lokal dengan penanda simulasi.
2. Default sort adalah Total omzet live, descending. Caption menyebut periode dan filter yang berlaku.
3. Nilai omzet sama menghasilkan peringkat sama dengan pola 1, 2, 2, 4. Nama host menjadi urutan tampilan di antara nilai sama, tanpa menyatakan salah satu lebih unggul.
4. Host dengan GMV kosong tidak diberi angka nol atau peringkat. Tampilkan pada daftar terpisah “Host belum memiliki data omzet”. GMV nol yang valid tetap boleh masuk ranking.
5. Koreksi laporan memakai versi aktif terbaru dan menghitung ulang ranking. ID slot yang sama tidak boleh dihitung dua kali.
6. Total room LIVE tidak boleh disalin ke beberapa host. Slot Linda 16.00–18.00 belum dapat memakai GMV Rp15.484.967 dari bukti sepanjang LIVE.
7. MVP satu host per slot. Untuk slot multi host di masa depan, jangan menggandakan omzet penuh pada semua host; perlukan aturan atribusi yang disepakati terlebih dahulu.
8. Jangan gabungkan mata uang atau basis gross/net yang berbeda. Jangan menjumlahkan Perkiraan GMV dengan omzet aktual sumber.
9. Filter Host pada tabel lengkap boleh mempersempit hasil; ranking dan kontribusi kemudian dihitung untuk kumpulan host yang difilter, dengan caption “Ranking dalam filter aktif”. Ringkasan default menampilkan semua host dalam cakupan brand dan periode.
10. Ranking omzet menunjukkan kontribusi penjualan teramati. Jangan menampilkan skor kualitas host atau menyimpulkan pengaruh kausal karena jadwal, produk, promo, dan durasi dapat berbeda.

### Desain komponen

**Ringkasan Top 5**

- Header “Leaderboard omzet host” dengan caption periode serta jumlah slot terverifikasi.
- Gunakan batang horizontal, urut omzet tertinggi di atas, maksimum lima host.
- Nama host di kiri; rupiah di ujung batang; jam live dan jumlah slot sebagai subteks.
- Semua batang memakai warna teal yang konsisten. Batang pertama boleh lebih tegas, tanpa podium, medali, piala, atau dekorasi kompetisi berlebihan.
- Tombol “Lihat semua host” menuju halaman Performa Host dan mempertahankan filter periode, studio, dan channel.
- Jika hanya satu host valid, tampilkan satu batang dengan konteks; jangan menciptakan pesaing.
- Jika belum ada atribusi valid, tampilkan empty state “Leaderboard tersedia setelah laporan omzet per slot host terverifikasi”. Jadwal Linda tetap dapat dibuka melalui “Lihat slot host”.

**Tabel lengkap**

Kolom default: Peringkat, Host, Total omzet live, Kontribusi omzet, Slot valid, Jam live, Omzet/jam, dan Cakupan data. Pesanan serta rata-rata omzet per slot tersedia melalui Atur kolom.

Klik host membuka detail dengan ringkasan omzet, chart omzet per slot, daftar slot yang masuk perhitungan, serta daftar slot yang belum memiliki data. Semua angka dapat ditelusuri ke ID Line, tanggal, dan rentang waktu slot. Pada mobile, gunakan kartu per host dengan ranking, omzet, dan tiga informasi pendamping; tabel lengkap tetap dapat dibuka.

### Data contoh untuk visual

Default data sumber tidak memiliki omzet Linda yang terverifikasi sehingga leaderboard aktual tetap kosong. Jika mode simulasi diaktifkan, gunakan slot fiktif beratribusi valid, terpisah dari laporan sumber. Semua kartu, grafik, ranking, kontribusi, dan ekspor dihitung dari fixture simulasi yang sama; jangan hardcode jumlah leaderboard.

Contoh bentuk record slot untuk perhitungan, dengan nilai sintetis dan penanda yang jelas:

```json
{
  "dataset": "synthetic_preview",
  "isDemo": true,
  "slotId": "sim-slot-001",
  "hostId": "sim-host-a",
  "hostName": "Host A",
  "date": "2026-09-26",
  "channel": "TikTok",
  "studio": "Studio 1",
  "durationMinutes": 120,
  "attributedGmvIdr": 2000000,
  "attributedOrders": 40,
  "gmvValid": true,
  "ordersValid": true,
  "scope": "host_slot",
  "currency": "IDR",
  "metricBasis": "demo_gross_gmv",
  "status": "demo_valid"
}
```

Record di atas hanya contoh schema. Jika dipakai dalam preview, tampilkan label simulasi dan hitung omzet/jam Rp1.000.000 serta AOV Rp50.000 dari record yang sama. Dataset simulasi dapat ditambah untuk membentuk leaderboard beberapa host.

### Acceptance criteria leaderboard

1. Ranking berubah sesuai periode, studio, channel, dan filter host aktif.
2. Total omzet host cocok dengan jumlah slot unik yang masuk perhitungan; total room tidak dihitung ulang sebagai omzet host.
3. Kontribusi seluruh host dalam kumpulan terpilih berjumlah sekitar 100% dengan toleransi pembulatan, jika penyebut positif. Top 5 saja boleh kurang dari 100% dan harus memakai penyebut semua host terpilih.
4. Omzet/jam memakai durasi dari slot yang sama dengan omzet yang dijumlahkan.
5. Nilai sama, nol, kosong, dan slot belum valid menghasilkan perilaku sesuai aturan ranking.
6. Linda tidak diberi ranking omzet dari screenshot sepanjang LIVE yang belum teratribusi.
7. Mode simulasi diberi label pada leaderboard, detail, tooltip, dan ekspor.
8. Ekspor leaderboard memuat periode, filter, basis omzet, ranking, total, jam live, serta cakupan data.

## 19. Checklist hasil

1. Ringkasan langsung menampilkan konteks akun, periode, dan cakupan.
2. Seluruh 14 metrik ledger tercakup melalui ringkasan/detail dengan label yang jujur.
3. Metrik tambahan screenshot dipisahkan dari schema ledger dan tidak menghasilkan rumus atribusi baru.
4. Sepanjang LIVE dan slot host tidak tercampur pada angka, filter, grafik, maupun unduhan.
5. GMV Linda tidak tampil Rp15.484.967 tanpa bukti slot.
6. Tidak ada tren, kenaikan, target, atau ranking aktual yang dikarang.
7. Data simulasi diberi label konsisten dan tidak tercampur data sumber.
8. UI rapi pada desktop, tablet, dan mobile; tidak ada overflow halaman atau teks terpotong.
9. Filter, navigasi, detail, chat simulasi, serta ekspor CSV dapat dicoba.
10. Empty state tetap terasa seperti produk selesai dan memberi langkah berikutnya yang jelas.
11. Visual terasa seperti aplikasi bisnis, dengan aksen warna terbatas dan hierarki angka yang kuat.
12. Leaderboard omzet host tersedia pada Ringkasan dan Performa Host; hanya slot beratribusi valid yang dihitung.

## 20. Prompt eksekusi singkat

Bangun frontend prototype **Its My Live untuk User Brand/Client** sesuai brief ini. Kerjakan Ringkasan terlebih dahulu, lalu Laporan Livestream, Detail Laporan, Performa Host dengan leaderboard omzet, Laporan Unduhan, dan widget Tanya Laporan. Gunakan data sumber 26 September 2026 sebagai default dan label pratinjau yang jelas. Pastikan angka sepanjang LIVE tidak diatribusikan ke slot Linda. Prioritaskan UI bisnis minimal, chart yang tepat, filter konsisten, dan responsive layout. Implementasikan interaksi menggunakan data lokal; layanan produksi belum dihubungkan. Jika menambahkan dataset 14 hari untuk menunjukkan chart, tandai seluruh tampilannya sebagai simulasi.
