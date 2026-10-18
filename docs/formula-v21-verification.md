# Verification Report: Formula V2.1

Story: pengelola mengisi input terverifikasi saat review upload, mempublikasikan sesi, kemudian dashboard dan leaderboard menampilkan formula, coverage, dan skor tanpa mengubah arti data sumber lama.

| Boundary | Status | Evidence |
|---|---|---|
| UI renders | Lulus | Delapan kartu formula terlihat pada dashboard; profil host memuat formula dan konteks input |
| Client → API | Lulus | Filter TikTok/Shopee, ranking skor V2.1 dan skor lintas platform merender hasil; tidak ada error console browser |
| API → Data | Lulus | Migrasi 005 diterapkan; field opsional diterima melalui endpoint review pada integration DB `_test` |
| Data → Response | Lulus | Test upload/review/publish menyimpan Direct GMV 1000000 dan menghasilkan skor pada API ringkasan host |
| Response → UI | Lulus | Nilai parsial tampil N/A + 0/n coverage; Comment Velocity yang tersedia tampil 0,17/menit |
| Privacy | Lulus | ACL privat dan respons global bebas brand/source/attachment tetap lulus |
| Responsive | Lulus | 1440 px: delapan kartu, tanpa page overflow; 390 px: grid dua kolom, page width 390 px |

Pemeriksaan yang berhasil: `make test`, `make test-integration`, `make vet`, `make frontend-check`, `make frontend-test` (35 tests), `make frontend-build`.

Issue yang ditemukan dan diperbaiki: parameter SQL fixture sempat mempunyai inferensi numeric/text yang ambigu; diberi cast eksplisit. Assertion numerik memakai decimal equality agar presisi PostgreSQL tidak mengakibatkan perbandingan string yang salah. Ini perubahan pengujian, bukan konversi data produksi.

Batas verifikasi: sumber lokal lama belum memiliki seluruh input V2.1 yang terverifikasi, sehingga tidak diklaim menghasilkan skor riil penuh. Skor penuh diuji dengan fixture terkontrol di database test; data CSV paket formula tidak diimpor ke database lokal pengguna.
