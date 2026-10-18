# IML Dashboard · Powered by SapaGo AI

Dashboard Its My Live untuk laporan TikTok Shop dan Shopee Live, pengelolaan host, pengguna/akses, review upload, serta leaderboard Formula V2.1.

Aplikasi aktif memakai **React 19 + TypeScript + Vite** (`frontend/`), **Go** (`cmd/`, `internal/`), dan **PostgreSQL 17**. Folder `src/` root dan `docs/legacy-prototype/` menyimpan prototype awal sebagai referensi; perintah npm di root mengarah ke frontend terbaru.

## Fitur

- Dashboard, laporan sesi, filter, pagination, ekspor CSV/XLSX, dan leaderboard agregat global.
- Formula V2.1: GPM, Conversion, CTR, Watch Time, CTOR, Follow Rate, Hook proxy, Comment Velocity; benchmark rolling 90 hari, gates, shrinkage, dan analisis lintas platform.
- Coverage dan audit input sumber: null berbeda dari nol; data lama tidak otomatis dianggap Direct GMV atau unique viewers.
- Upload bukti privat, review manual, publikasi/revisi, antrean dan audit. Mode simulasi eksplisit untuk demo/test.
- Profil host/pengguna, foto opsional dengan fallback inisial, assignment beberapa brand, role Superadmin/Admin/Client.
- Design tokens warna, konfigurasi Diatype dengan fallback, searchable dropdown, tags, grid/spacing responsif dan micro-interactions.

## Menjalankan lokal

Prasyarat: Node.js >=22.18, Go sesuai `go.mod`, PostgreSQL 17 dan `make`. Untuk jalur PostgreSQL native, sediakan juga `initdb`, `pg_ctl`, `psql`, `createdb`, `openssl`, dan `rg` pada PATH.

```sh
npm run setup
make db-native
make migrate
make seed-demo
```

`make db-native` menyiapkan cluster khusus project di `127.0.0.1:55432` dan membuat `.env` privat jika belum ada. Perintah ini tidak menimpa `.env` yang sudah ada. Kredensial demo dibuat oleh seed dan tersimpan lokal di `.local/demo-credentials.json`; file tersebut tidak disertakan di Git.

Jalankan pada terminal terpisah:

```sh
make run          # API: http://127.0.0.1:8080
make frontend-dev # UI: http://127.0.0.1:5173
make worker       # diperlukan untuk pekerjaan ekstraksi berantrean
```

`npm run dev` di root juga menjalankan UI terbaru. Login di [halaman login lokal](http://127.0.0.1:5173/login). API menyediakan `/healthz` dan `/readyz`.

Alternatif Docker: salin `.env.example` menjadi `.env`, isi `DATABASE_URL`, `POSTGRES_PASSWORD`, serta `WORKER_TOKEN` dengan nilai lokal yang kuat, lalu jalankan `make db-up` dan langkah migrasi/seed di atas. Jangan memakai placeholder `CHANGE_ME` sebagai kredensial nyata. Aktifkan `DEMO_MODE=true` jika ingin memakai seed demo.

## Verifikasi

```sh
make test
make vet
make frontend-check
make frontend-test
make frontend-build
make test-integration
```

Integration tests hanya boleh memakai database dengan nama berakhiran `_test`, karena schema database test akan dibuat ulang. Script memakai `IML_TEST_DATABASE_URL` bila disediakan, atau database `iml_test` dari koneksi lokal.

## Dokumen

- [Implementasi Formula V2.1](docs/formula-v21-implementation.md)
- [Kekurangan data pada contoh upload](docs/source-upload-gaps-v21.md)
- [Verifikasi Formula V2.1](docs/formula-v21-verification.md)
- [Grid & spacing](docs/grid-spacing-system.md)
- [Dropdown & tags](docs/dropdown-tags.md)
- [Profil host dan akses pengguna](docs/enhancement-host-users.md)
- [Design system](docs/design-system-implementation.md)
- [Keputusan arsitektur](docs/decisions.md)

## Batas integrasi dan aset

Mode manual berjalan tanpa OpenClaw/SapaGo. Gateway ekstraksi OpenClaw memerlukan konfigurasi dan pengujian terpisah; integrasi model/chat SapaGo belum terhubung. Tidak ada fallback simulasi otomatis ketika layanan nyata gagal.

Font Diatype berlisensi tidak disertakan. Gunakan file font yang memiliki izin pemakaian sesuai [panduan aset font](frontend/src/assets/fonts/README.md); aplikasi memakai system fallback jika font belum tersedia.

Repository berisi source, migrasi, fixture sintetis, dan dokumentasi. `.env`, database, backup, upload privat, kredensial, dependensi terpasang, serta build artifact dikecualikan dari Git. Screenshot dokumentasi menggunakan data demo.
