# Profil host dan pengguna

Host: nama tampil, nama lengkap, email opsional, tanggal lahir opsional, jenis kelamin opsional, foto PNG/JPEG opsional (2 MB), tanggal bergabung opsional, domisili/kota, bahasa, spesialisasi konten/produk, assignment beberapa brand, profil publik dan status aktif. Inisial nama digunakan otomatis jika foto kosong atau gagal dimuat. Foto dapat diganti atau dihapus.

Domisili, bahasa, spesialisasi dan tanggal bergabung direkomendasikan untuk analisis operasional berikutnya. Semua parameter disimpan dalam master; perubahan ini tidak mengganti rumus peringkat atau menambahkan filter demografis leaderboard. Nama lengkap, email, tanggal lahir dan jenis kelamin tidak dikirim pada API leaderboard global. Client tidak menerima data pribadi pada master. Foto publik host mengikuti izin profil global; foto pengguna mengikuti hak akses pengguna dan assignment.

Superadmin dapat membuat dan mengelola Client, Admin dan Superadmin dengan seluruh permission bawaan Superadmin. Admin hanya dapat membuat Client baru pada menu Pengguna & akses; edit, toggle status, reset/undangan ulang dan pencabutan sesi pengguna lama dibatasi ke Superadmin. Hak dashboard/laporan Client yang ada tetap sama. Pembatasan diberlakukan pada endpoint, bukan hanya tampilan.

Pengguna: foto opsional, nama, email, role, tanggal bergabung wajib, toggle aktif/nonaktif, dan assignment beberapa brand dengan Pilih semua/Batalkan semua. Assignment lama dimuat saat edit. Superadmin otomatis mengakses semua brand. Akun baru mendapatkan link undangan lokal sekali pakai dengan kedaluwarsa satu jam; tidak ada email dikirim. Perubahan akses mencabut sesi lama. Akun sendiri tidak dapat menonaktifkan/mencabut sesi atau mengubah role sendiri melalui menu ini.

Migrasi 004_profiles diterapkan pada database lokal. Tanggal bergabung pengguna lama diisi dari tanggal pembuatan akun dalam zona Asia/Jakarta; nama lengkap host lama diisi dari nama yang sudah ada. Tidak ada tanggal lahir atau jenis kelamin yang ditebak.

Foto disimpan atomik bersama profil pada PostgreSQL. API memeriksa ukuran, signature PNG/JPEG, validitas decoding dan batas dimensi; menolak PDF, SVG, file palsu/rusak. Foto disajikan melalui endpoint terautentikasi dengan cache private/no-store. Personal fields tidak dicatat dalam detail audit.

Validasi: TypeScript, ESLint, build, 26 tes frontend; suite integrasi PostgreSQL mencakup mandatory date, create-only Admin, pencegahan eskalasi role, assignment lintas portofolio, multi-brand, semua role, foto roundtrip/preserve/remove, penolakan foto palsu/oversize, tanggal tidak valid, privasi master/global, dan aktivasi pengguna. Browser: form host/Admin/Superadmin, inisial langsung mengikuti nama, role options, Pilih semua, assignment lama, tanggal wajib, dan switch status.
