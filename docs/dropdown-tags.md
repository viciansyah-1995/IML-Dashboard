# Dropdown dan tags

Semua dropdown yang dirender melalui `Field` menggunakan komponen `Select`: pencarian, pilihan aktif, kondisi kosong, opsi nonaktif, serta navigasi Arrow/Enter/Escape. Popup ditempatkan di dalam dialog bila field berada dalam dialog supaya tetap tersedia bagi pembaca layar. Native select tersembunyi hanya menjadi proxy FormData dan validasi; pengguna berinteraksi dengan combobox custom.

`TagInput` dipakai pada bahasa dan spesialisasi profil host. Mendukung saran, nilai baru, Enter/koma untuk menambahkan, Backspace untuk menghapus pilihan terakhir saat input kosong, tombol hapus per tag, dan commit saat blur. Duplikasi diabaikan tanpa membedakan kapital. Data lama dengan pemisah koma, titik koma, atau baris baru terbaca. Nilai diserialisasi sebagai string dipisahkan koma, sesuai kontrak API saat ini. Maksimal 20 tags, dengan batas 300 karakter untuk bahasa dan 500 untuk spesialisasi.

Validasi: 33 tes frontend, typecheck, ESLint, dan build. Browser: perubahan filter master, dropdown dalam dialog, penambahan beberapa tags, serta viewport mobile 390 px tanpa overflow horizontal. Screenshot tersedia di `screenshots/custom-dropdown.png` dan `screenshots/host-tags.png`.
