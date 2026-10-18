# Formula dashboard V2.1 — implementasi 18 Oktober 2026

Sumber: `DOKUMENTASI-FORMULA-V2.1.md`, bagian formula eksplisit 2–3, 7 dan 9; source upload project: `docs/requirements-source.txt` bagian 11. Angka pada dataset contoh tidak diimpor menjadi data produksi.

## Metrik dan perubahan

| Metrik | Formula | Bobot |
|---|---|---:|
| GPM | Direct GMV / Impressions × 1000 | 30% |
| Conversion | Unique buyer / Unique viewers × 100 | 20% |
| CTR | Product clicks / Unique viewers × 100 | 15% |
| Watch Time | Avg viewing duration (s) / 60 | 12% |
| CTOR | Orders / Product clicks × 100 | 8% |
| Follow Rate | New followers / Unique viewers × 100 | 6% |
| Hook proxy | Unique viewers / Impressions × 100 | 5% |
| Comment Velocity | Comments / Duration minutes | 4% |

Cart→Order = Orders / ATC ×100, gate tanpa bobot. Semua rasio periode menghitung total pembilang dan total penyebut dari subset sesi yang sama. Watch periode memakai rata-rata berbobot unique viewers, sehingga membutuhkan bobot viewers yang terverifikasi. Field lama sales/views/buyers tetap menjadi konteks sumber; tidak otomatis dianggap direct GMV/unique viewers/unique buyer.

Dashboard menampilkan metrik, bobot, unit, coverage, input belum lengkap, rincian benchmark dan skor. Leaderboard, profil, perbandingan host dan ekspor global mendukung skor V2.1. Ranking baru default skor V2.1; pilihan sorting URL yang sudah disimpan tetap dihormati. AOV, likes, shares, dan Indirect GMV tidak masuk bobot host.

## Benchmark, skor dan data parsial

- Benchmark per platform: trimmed mean 10% tiap sisi, jendela 90 tanggal kalender berakhir pada tanggal akhir filter (inklusif, `end - 89 hari`). Tidak memakai data masa depan; filter awal dashboard tidak mengubah benchmark.
- Pool benchmark agensi memakai host dengan `global_visible=true`, sesuai izin agregat global. Tidak mengirim client, source, slot, bukti, atau PII privat. Host tersembunyi tidak masuk pool global.
- Setiap metrik memakai input validnya sendiri. Jika sampel valid platform <5, gunakan pool kedua platform dan tampilkan warning. Jika pool <5, gunakan median sesuai helper referensi. Benchmark nol/tidak ada menghasilkan N/A, bukan angka contoh hardcoded.
- Skor = `clip(5 + 3.5 × log10(value / benchmark), 0, 10)`. Nilai nol nyata menghasilkan skor metrik 0. Denominator nol atau input hilang menghasilkan N/A; tidak diganti 1 atau nol.
- Skor periode dihitung ulang dari sesi yang lengkap untuk delapan metrik dan gate ATC, bukan rata-rata skor persentase atau bobot yang dinormalisasi ulang. Tampilkan jumlah sesi yang dikeluarkan.
- Gate formal: CTR score >7 dan conversion <1,5% → cap CTR 5,5; CTOR score >7 dan Orders/ATC <30% → cap CTOR 5; Hook score >8 dan Watch <1,5 menit → cap Hook 6.
- Shrinkage hanya n<15: `n/(n+10) × observed + 10/(n+10) ×5`. n adalah sesi historis valid host pada platform yang sama dalam window tersebut (termasuk sesi terpilih yang masuk window). Tidak menggandakan shrinkage per metrik. Pool publik dan histori privat yang berizin dideduplikasi per slot untuk n.
- Jika lebih dari satu host dipilih pada ringkasan brand, hanya tampilkan metrik agregat. Pilih satu host untuk skor host; tidak menganggap jumlah sesi gabungan semua host sebagai pengalaman satu host.
- Sesi room, snapshot, draft, versi tidak aktif, slot duplikat tidak menjadi skor host. Flag konflik/buram/definisi belum dikonfirmasi membuat input tidak layak untuk formula.

## Lintas platform

Hitung skor periode TikTok/Shopee secara terpisah. Bila keduanya tersedia: base = rata-rata dua skor; std populasi dari dua skor; consistency = clamp(1-std/mean,0,1); bonus=(consistency-0,5)×0,5. Skor akhir = clamp(base+bonus,0,10). Jika platform kedua tidak memiliki skor valid, bonus 0. Pilihan filter channel utama tetap menentukan kelayakan ranking: host tanpa skor utama tidak dinaikkan hanya karena ada platform lain.

Percentile = persentase skor pool yang <= skor host (ties setara; skor terbaik 100). Z = (score-mean)/std populasi, N/A jika variance=0. Pool adalah seluruh host dengan skor tersedia sebelum pencarian/pagination. Flag jika N<10. Angka ini merupakan indeks performa, bukan keputusan otomatis reward, akses, atau ketenagakerjaan.

## Keputusan atas konflik pada attachment

1. Rumus eksplisit mengalahkan contoh aritmetika: ratio 2 → 6,053605; 4 → 7,107210; 10 →8,5; 100 →10. Contoh 6,32/7,50/9,00/9,50 pada dokumen tidak sesuai `3.5 log10`.
2. Gate mengikuti bagian 3.3 dokumen, bukan threshold dinamis berbeda di Python.
3. Shrinkage mengikuti syarat n<15, bukan contoh ilustrasi n=15/30 yang masih menunjukkan shrinkage.
4. Benchmark memakai trimmed mean, bukan menyebut hasilnya median. Nilai 5 merupakan anchor benchmark; tidak selalu median distribusi.
5. Skor periode dihitung ulang sesuai bagian 7, bukan mean skor sesi seperti fungsi ranking Python.
6. Base lintas platform memakai rata-rata kedua platform agar simetris; contoh dokumen memakai skor salah satu platform tanpa aturan pemilihan yang jelas.

## Input dan kompatibilitas

Migrasi 005 menambah input opsional ke mapping, termasuk Shopee yang sebelumnya belum mempunyai impressions/product clicks/new followers. `direct_gmv`, `indirect_gmv`, `unique_viewers`, `unique_buyers`, `avg_concurrent_viewers`, `avg_price_main_orders_k` memiliki arti eksplisit dan bukti review. Ekstraktor diminta mengambil hanya data yang terlihat; metadata/label ambigu tetap null. Tidak ada backfill angka lama secara diam-diam dan aturan wajib publikasi sumber lama tidak berubah. Ekspor ledger A:AG dipertahankan; input baru tersimpan di detail sumber, sementara ekspor global menambah kolom formula agregat.

## Verifikasi

- Unit Go: nilai rasio/log, bobot, denominator nol, null vs nol, subset berpasangan, watch berbobot, tiga gate, window tanggal/fallback, boundary 14/15 sesi, multi-host dan room, trimmed mean, lintas platform dan pool kecil.
- Integrasi PostgreSQL `_test`: migrasi, input review baru, API skor dari fixture terverifikasi, null untuk data lama, ACL privat, allowlist respons global, serta regresi alur upload/review/publish/export.
- Frontend: 35 tes (termasuk bobot, unit dan source gap), TypeScript, ESLint, production build.
