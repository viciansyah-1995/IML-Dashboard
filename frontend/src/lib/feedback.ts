export type Notice={id:string;kind:'info'|'success'|'error';message:string;pending?:boolean};
export function notify(message:string,kind:Notice['kind']='info',id:string=crypto.randomUUID(),pending=false){if(typeof window!=='undefined')window.dispatchEvent(new CustomEvent<Notice>('iml:notice',{detail:{id,kind,message,pending}}));return id}
export function actionCopy(path:string,body?:BodyInit|null){
 let action='';if(typeof body==='string'){try{action=JSON.parse(body).action||''}catch{/* FormData and invalid JSON are validated by the API. */}}
 const name=path.split('?')[0];
 if(name==='auth/login')return ['Memeriksa akun…','Berhasil masuk. Workspace Anda siap digunakan.'];
 if(name==='auth/logout')return ['Mengakhiri sesi…','Anda berhasil keluar.'];
 if(name==='auth/reset')return ['Menyimpan password baru…','Password berhasil diperbarui. Silakan masuk kembali.'];
 if(name==='uploads')return ['Menyimpan metadata dan lampiran…','Upload tersimpan. Review laporan sebelum mempublikasikan.'];
 if(name.endsWith('/review'))return ['Menyimpan review…','Review tersimpan. Periksa kelengkapan sebelum publikasi.'];
 if(name.endsWith('/publish'))return ['Memvalidasi dan mempublikasikan…','Laporan berhasil dipublikasikan. Ringkasan dan ranking diperbarui.'];
 if(name.endsWith('/revise'))return ['Membuat versi revisi…','Draft revisi dibuat. Versi final lama tetap aktif hingga penggantinya dipublikasikan.'];
 if(name.endsWith('/revoke'))return ['Mencabut publikasi…','Publikasi dicabut. Laporan tidak lagi masuk agregat final.'];
 if(name.startsWith('operations/jobs/'))return ['Mengantrekan percobaan baru…','Percobaan baru masuk antrean. Riwayat pekerjaan sebelumnya tetap disimpan.'];
 if(name==='insights')return ['Menyusun ringkasan dari laporan final…','Ringkasan siap. Periksa periode dan cakupan sebelum menggunakannya.'];
 if(name==='users'){const result:Record<string,string>={deactivate:'Pengguna dinonaktifkan dan akses sesi lama dicabut.',activate:'Pengguna diaktifkan kembali.',revoke:'Sesi pengguna berhasil dicabut.',reset:'Link reset lokal dibuat. Link hanya dapat digunakan sekali.',invite:'Link undangan lokal dibuat. Tidak ada email yang dikirim.'};return ['Memproses pengaturan pengguna…',result[action]||'Pengguna dan assignment berhasil disimpan.']}
 if(name.startsWith('masters/'))return ['Menyimpan master data…','Master data berhasil disimpan.'];
 if(name==='dashboard-config')return ['Menyimpan susunan widget…','Susunan dashboard brand berhasil disimpan.'];
 if(name==='definitions')return ['Memvalidasi konfigurasi definisi…','Konfigurasi definisi berhasil disimpan.'];
 if(name==='permissions')return ['Memperbarui hak akses…','Hak akses berhasil diperbarui dan berlaku pada request berikutnya.'];
 return ['Memproses perubahan…','Perubahan berhasil disimpan.'];
}
export function errorCopy(code:string,message:string){
 const help:Record<string,string>={NETWORK:'Koneksi ke server terputus. Periksa koneksi atau coba lagi.',FORBIDDEN:'Akun Anda belum memiliki hak akses untuk tindakan ini. Hubungi Admin atau Superadmin.',REVISION_CONFLICT:'Laporan telah diubah pengguna lain. Muat ulang, periksa versi terbaru, lalu simpan kembali.',STALE_LEASE:'Pekerjaan sudah kedaluwarsa atau diambil worker lain. Periksa antrean sebelum mencoba lagi.',RATE_LIMIT:'Terlalu banyak permintaan dalam waktu singkat. Tunggu sebentar sebelum mencoba lagi.',INCOMPATIBLE_DEFINITIONS:'Data memakai versi definisi berbeda. Pilih satu versi definisi pada filter.',LOGIN:'Email atau password belum cocok. Periksa kembali lalu coba masuk.',CSRF:'Sesi keamanan berubah. Muat ulang halaman sebelum melanjutkan.',DATABASE:'Server belum dapat menyimpan atau membaca data. Coba lagi setelah koneksi database pulih.'};
 return help[code]||((message&&message.length>8)?message:'Tindakan belum berhasil. Periksa data wajib dan syarat tindakan, lalu coba lagi.');
}
export function blockerCopy(code:string){const [flag,key]=code.split(':');const names:Record<string,string>={sales:'penjualan',views:'views',orders:'pesanan',buyers:'pembeli',viewers:'penonton',watch_duration:'durasi rata-rata menonton',likes:'likes',comments:'komentar',shares:'shares'};const metric=key?names[key]||key:'data';const messages:Record<string,string>={PERIOD_UNCONFIRMED:'Periode pada bukti belum terkonfirmasi. Gunakan bukti dengan waktu mulai dan akhir yang jelas.',SCOPE_MISMATCH:'Periode atau cakupan bukti berbeda dari slot host. Unggah bukti yang cocok; total sepanjang LIVE tidak boleh dipindahkan ke slot.',SNAPSHOT_NOT_FINAL:'Bukti ini masih berupa snapshot. Gunakan laporan seluruh periode sebelum mempublikasikan.',MISSING_METRIC:`Metrik ${metric} wajib dilengkapi dari bukti yang sah.`,MISSING_EVIDENCE:`Bukti untuk ${metric} belum lengkap. Pilih lampiran dan isi teks angka yang terbaca.`,CONFLICT:`Nilai ${metric} berbeda antar bukti. Tentukan sumber yang benar dan catat alasan koreksinya.`,LOW_READABILITY:`Bukti ${metric} belum terbaca jelas. Gunakan gambar lebih jelas atau konfirmasi nilai dari sumber yang dapat dibaca.`,DEFINITION_UNCONFIRMED:'Definisi metrik belum dikonfirmasi. Gunakan kontrak yang disetujui sebelum menggabungkan laporan.'};return messages[flag]||'Masih ada syarat publikasi yang belum terpenuhi. Periksa data dan bukti pada review.'}
