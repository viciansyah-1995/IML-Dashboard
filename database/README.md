# Instalasi database di komputer lokal lain

Paket ini membangun database dari source yang sama dengan aplikasi terbaru: migrasi 001–005, profil host/pengguna, serta input Formula V2.1. Tidak ada database pengguna, session login, atau password bersama di repository.

## Instalasi otomatis (disarankan)

Prasyarat: Python >=3.9, Go sesuai `go.mod`, Node.js >=22.18 dan npm. Database memerlukan PostgreSQL 17 native **atau** Docker Desktop/Engine dengan Compose v2. Pada Windows, jalankan melalui WSL2.

Setelah clone:

```sh
git clone https://github.com/viciansyah-1995/IML-Dashboard.git
cd IML-Dashboard
python3 scripts/install-local.py --database native
```

Mode native memerlukan `initdb`, `pg_ctl`, `psql`, dan `createdb` pada PATH. macOS dengan Homebrew: tambahkan direktori `postgresql@17/bin` ke PATH sesuai instalasinya. Linux harus menjalankan installer sebagai pengguna biasa, bukan root.

Alternatif menggunakan PostgreSQL Docker (Go dan Node tetap berjalan native):

```sh
python3 scripts/install-local.py --database docker
```

Perintah ekuivalen jika `make` tersedia:

```sh
make install-local
make install-local ARGS="--database docker"
```

Installer akan:

1. Membuat `.env` dengan password PostgreSQL dan worker token acak, mode file privat.
2. Menyiapkan PostgreSQL di `127.0.0.1:55432`, database `iml` dan `iml_test`.
3. Menjalankan seluruh migrasi terbaru.
4. Membuat data demo sintetis (brand, host, sesi, dan akun Superadmin/Admin/Client) jika `DEMO_MODE=true`.
5. Menginstal dependency frontend dengan `npm ci`.

Jalankan aplikasi pada terminal terpisah:

```sh
make run
make frontend-dev
# Opsional untuk pekerjaan ekstraksi berantrean:
make worker
```

Buka [login lokal](http://127.0.0.1:5173/login). Email/password instalasi baru tersimpan pada `.local/demo-credentials.json`. Password setiap komputer berbeda, sehingga kredensial komputer sebelumnya tidak berlaku.

## Instalasi ulang dan port

Installer dapat dijalankan ulang. `.env`, data database, volume Docker, password, dan akun yang sudah ada tidak direset. Migrasi yang telah diterapkan dilewati; seed demo juga tidak menggandakan data. Mengubah environment di shell tidak mengganti koneksi yang dipilih installer: langkah Go memakai `.env` project tersebut.

Jika port default terpakai, tentukan port lain **pada instalasi baru**:

```sh
python3 scripts/install-local.py --database native --port 55433
```

Untuk `.env` yang sudah ada, port dari `DATABASE_URL` tetap dipakai. Jangan mengganti `POSTGRES_PASSWORD` setelah volume Docker berisi data: ubah password di PostgreSQL secara terencana dan selaraskan konfigurasi. Installer tidak menghapus volume atau mengganti password otomatis.

Opsi tambahan:

```sh
python3 scripts/install-local.py --database native --skip-frontend
python3 scripts/install-local.py --database native --database-only
python3 scripts/install-local.py --database native --go-bin /path/to/go
```

`--database-only` hanya menyiapkan cluster/database. Jalankan migrasi dan seed secara terpisah bila memakai opsi tersebut.

## File SQL untuk instalasi manual

`database/schema.sql` berisi schema lengkap dan data referensi (role/permission/mapping metrik), termasuk penanda migrasi. Gunakan **hanya pada database kosong**. Untuk menyiapkan PostgreSQL tanpa Go/Node, jalankan opsi `--database-only`, lalu impor SQL:

```sh
python3 scripts/install-local.py --database native --database-only
psql -h 127.0.0.1 -p 55432 -U iml -d iml -v ON_ERROR_STOP=1 -f database/schema.sql
```

Password dimasukkan lewat prompt PostgreSQL atau konfigurasi lokal, bukan ditulis pada perintah yang dibagikan. SQL memakai transaksi dan berhenti pada error. Tabel yang sudah ada tidak dihapus; import ke database yang telah berisi schema akan gagal dan rollback.

File SQL tidak berisi akun login aplikasi. Untuk data demo, setelah `.env` dikonfigurasi dengan `DEMO_MODE=true`, jalankan `make seed-demo`. Untuk update schema pada instalasi yang sudah ada, jalankan `make migrate`, bukan mengimpor ulang SQL.

Setelah menambah migrasi, regenerasi paket SQL:

```sh
make db-schema
```

## Memindahkan data aktual

Installer default memakai data demo, bukan salinan data pengguna. Untuk menyalin isi database aktual dan bukti upload, gunakan `make backup` pada komputer asal dan pindahkan folder backup melalui media privat. Panduan pemeriksaan restore tersedia pada `scripts/restore-check.py`. Backup, `.env`, dan `.local` tetap dikecualikan dari Git; jangan mengunggahnya ke repository publik.
