# Ketersediaan data upload untuk Formula V2.1

Audit ini mengacu pada contoh sumber dalam brief project yang sudah diimplementasikan (`docs/requirements-source.txt` bagian 11), bukan pada seluruh kemampuan platform TikTok/Shopee. Lampiran baru berisi formula dan dataset contoh; tidak membuktikan field itu ada pada screenshot upload pengguna.

| Field brief V2.1 | TikTok: contoh upload | Shopee: contoh upload | Dampak |
|---|---|---|---|
| Direct GMV ≤24h | Perolehan GMV ada, definisi direct/window belum dikonfirmasi | Penjualan ada, definisi direct/window belum dikonfirmasi | GPM dan skor menunggu nilai/definisi terverifikasi |
| Indirect GMV | Tidak terlihat; Perkiraan GMV bukan pengganti | Tidak terlihat | Konteks LTV N/A, bukan skor |
| Impressions | Ada (2,42M), label Show PV perlu verifikasi | Tidak terlihat | Shopee: GPM dan Hook N/A |
| Unique viewers (`Views` di brief) | Tayangan 61,52K ada; belum terbukti unique viewer | Views 980 / viewers 759 ada; definisi unique belum terverifikasi | Conversion, CTR, Follow, Hook serta bobot watch periode menunggu konfirmasi |
| Unique buyer | Pembeli 201 ada; arti unique perlu dikonfirmasi | Pembeli 1 ada; arti unique perlu dikonfirmasi | Conversion menunggu input eksplisit |
| Product clicks absolut | Ada (3,82K) | Tidak terlihat | Shopee: CTR/CTOR N/A; tidak dihitung dari click rate |
| New followers | Ada (59) | Tidak terlihat | Shopee: Follow Rate N/A |
| ATC | Tidak terlihat pada contoh | Ada (49) | TikTok: gate Cart→Order belum dapat dievaluasi |
| Avg viewing duration | Ada (15 detik) | Ada (26 detik) | Watch per sesi dapat dihitung; periode perlu bobot viewer yang valid |
| Orders | Ada (222) | Ada (1), basis Pesanan Dibuat | Gunakan basis filter yang sama |
| Comments | Ada (392) | Ada (49) | Comment Velocity memakai durasi slot yang terverifikasi |
| Avg concurrent viewers (`Viewers` di brief) | Current viewers 128 bukan rata-rata concurrent | Viewers 759 / aktif 61 / peak 8 bukan otomatis rata-rata concurrent | Konteks saja; belum terverifikasi |
| AvgPriceMainOrders_K | AOV 69,75K bukan otomatis Avg Price Main Orders | Belum ada nilai yang terverifikasi pada contoh | Konteks klien, tidak masuk bobot |
| ID, host, client, platform | Metadata internal tersedia | Metadata internal tersedia | Diisi pada upload, tidak ditebak dari grafik |
| Tanggal/start/end/durasi | Bukti contoh room 06–18 berbeda slot 16–18 | Periode pada contoh terpotong, perlu konfirmasi | Totals room tidak boleh dijadikan hasil satu host |
| Likes / shares | Ada (5.220 / 56) | Ada (228 / 2) | Konteks interaksi, tidak masuk skor |

Dampak praktis: sumber yang tersedia saat ini tidak cukup untuk menerbitkan skor penuh V2.1 secara jujur. Dashboard menampilkan N/A dan daftar field yang perlu dilengkapi, sementara metrik sumber lama tetap tampil sebagai konteks. Nilai nol eksplisit tetap nol; field tidak tersedia tetap null. Data tambahan diisi melalui review dengan bukti yang mendukung definisinya.

Data untuk roadmap yang juga belum tersedia: paid_traffic_share, retensi 15 detik asli, atribusi co-host per sesi, serta normalisasi prime-time/weekend/kategori. Tidak dibuat sebagai input wajib V2.1.
