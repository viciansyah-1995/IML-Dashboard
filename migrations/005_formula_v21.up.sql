-- Explicit semantics: never backfill legacy source sales/views/buyers into
-- V2.1 direct GMV or unique viewers/buyers without reviewing the evidence.
INSERT INTO metric_definitions(mapping_version,key,label,unit,kind,aggregation,required,canonical_confirmed)
SELECT m.id,k.key,k.label,k.unit,k.kind,k.agg,false,false
FROM metric_mapping_versions m CROSS JOIN (VALUES
 ('direct_gmv','Direct GMV terverifikasi · atribusi ≤24 jam','IDR','money','sum'),
 ('indirect_gmv','Indirect GMV · pasca-24 jam (konteks LTV)','IDR','money','sum'),
 ('unique_viewers','Unique viewers per sesi · definisi terverifikasi','count','count','per_session'),
 ('unique_buyers','Unique buyer per sesi · definisi terverifikasi','count','count','per_session'),
 ('avg_concurrent_viewers','Rata-rata concurrent viewers','count','count','weighted_only'),
 ('avg_price_main_orders_k','Avg Price Main Orders · Rp ribu (konteks klien)','IDR_thousands','money','source'),
 ('impressions','Impresi','count','count','sum'),
 ('product_clicks','Klik produk absolut','count','count','sum'),
 ('new_followers','Pengikut baru','count','count','sum'),
 ('add_to_cart','Tambah ke keranjang','count','count','sum')
) k(key,label,unit,kind,agg)
ON CONFLICT(mapping_version,key) DO NOTHING;
