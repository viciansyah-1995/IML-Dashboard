# Upload screenshot-first — 26 Oktober 2026

## Flow

1. Pilih TikTok/Shopee dan 1–10 screenshot PNG/JPEG (maksimal 10 MB per file).
   Screenshot disimpan sebagai draft privat sebelum metadata operator diketahui.
2. Worker OpenClaw membaca screenshot. Hasil tampil sebagai tabel label sumber,
   nilai, status, dan bukti. Field koreksi baru dibuka ketika user menekan Koreksi.
   Hasil asli AI tetap disimpan; nilai koreksi, teks bukti, dan alasan disimpan terpisah.
3. User melengkapi brand/akun, studio, host dan jadwal. Isi cepat dari teks operator
   tetap tersedia. Master dipilih dari data aktif dan assignment yang berizin.
4. Ringkasan hanya baca → Ajukan ke admin. Hanya tahap ini yang membuat host slot,
   source report dan report version berstatus review, tanpa publikasi otomatis dan
   tanpa menjalankan ekstraksi kedua kali.

## Penyimpanan & pemulihan

Migrasi `008_upload_intakes` menambah `upload_intakes` dan `intake_attachments`.
Draft menyimpan channel, profil mapping, hasil asli AI, koreksi, metadata, tahap,
revision dan status proses. URL `admin/upload?intake=ID` dapat dibuka kembali.
Daftar upload milik sendiri memiliki pagination. Draft yang belum lengkap tidak
masuk query dashboard, export atau leaderboard.

Status draft: uploaded → queued → running → extracted → submitted. Failure
sementara memiliki retry terbatas. Lease kadaluwarsa dapat diklaim ulang;
setelah batas percobaan draft menjadi failed dan dapat dicoba manual kembali.
Channel dan gambar hasil ekstraksi tidak diganti dalam draft yang sama; gunakan
upload baru supaya bukti dan hasil tetap konsisten.

`POST /api/upload-intakes` menyimpan gambar dan menerima Idempotency-Key.
GET daftar/detail/lampiran hanya untuk pemilik draft yang memiliki upload_report.
Aksi POST: extract, check, corrections, metadata, submit. Revision diperiksa pada
perubahan hasil/metadata; pengajuan bersifat transaksional dan idempotent.

Worker lama tetap memakai `/internal/jobs`; worker baru juga memeriksa
`/internal/intakes/claim`. Claim intake memberi `resource=intakes`; fetch lampiran,
heartbeat, result dan failure memakai prefix `/internal/intakes/{id}`.
Token machine dan lease yang sama tetap diwajibkan. Draft baru selalu memakai
OpenClaw, tanpa fallback mock. Data hasil di-normalisasi dan referensi bukti serta
metric key diverifikasi server sebelum diterima.

## Metadata, scope, dan review

Jadwal operator bukan bukti periode sumber. Observasi AI divalidasi server;
periode berbeda dipertahankan sebagai room dan tidak masuk skor pribadi host.
Periode tidak terbaca, basis Shopee tidak diketahui, konflik, koreksi uploader dan
angka yang kurang jelas tetap perlu review admin. Publikasi menggunakan validator
lama. Nilai kosong adalah tidak tersedia, bukan nol. Koreksi uploader tidak boleh
menghapus flag konflik/keterbacaan dari hasil AI.

## Role Host

Superadmin dapat membuat role Host dengan `host_id` yang menunjuk master host.
Assignment brand harus tercakup pada assignment host; studio juga diassign.
Role ini hanya memperoleh upload_report. UI Host diarahkan ke wizard, tidak
mendapat kontrol review/publikasi atau dashboard brand. Saat pengajuan, server
menolak host_id selain profil sendiri dan menolak alur upload lama untuk Host.
Akses Client dan kebijakan Admin hanya membuat Client tetap dipertahankan.
Tidak ada akun atau kredensial Host baru yang dibuat otomatis di database lokal.

## Lingkungan lokal

Draft tetap dapat menyimpan gambar ketika OpenClaw belum tersedia. Menjalankan
extract memberi pesan konfigurasi dan tidak mengisi hasil palsu. Backend harus
memakai EXTRACTION_MODE=openclaw dengan Gateway/token/agent, dan worker harus
berjalan sebelum vision nyata dapat diuji. Script backup digunakan sebelum migrasi;
SQL installer portabel diperbarui menggunakan scripts/export-schema.py.
