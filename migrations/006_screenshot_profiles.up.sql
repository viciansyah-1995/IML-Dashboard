-- New uploads use a separate contract; existing reports retain their mapping.
INSERT INTO metric_mapping_versions(id,channel_id,label,confirmed,is_demo) VALUES
('tiktok-screenshot-v1','tiktok','Screenshot TikTok · label asli & formula sumber',true,false),
('shopee-screenshot-v1','shopee','Screenshot Shopee · label asli & formula sumber',true,false);
INSERT INTO metric_definitions(mapping_version,key,label,unit,kind,aggregation,required,canonical_column,canonical_confirmed)
SELECT replace(mapping_version,'source-v1','screenshot-v1'),key,label,unit,kind,aggregation,required,NULL,false
FROM metric_definitions WHERE mapping_version IN ('tiktok-source-v1','shopee-source-v1');
INSERT INTO metric_definitions(mapping_version,key,label,unit,kind,aggregation,required,canonical_confirmed)
SELECT m.id,k.key,k.label,k.unit,k.kind,k.agg,false,false
FROM metric_mapping_versions m CROSS JOIN (VALUES
('source_duration_seconds','Durasi LIVE tertulis pada bukti','seconds','duration','source'),
('source_sales_per_hour','GMV per jam sumber','IDR','money','source'),
('source_comment_rate','Persentase komentar sumber','ratio','ratio','source'),
('source_ctor','Pesanan per klik / CTOR sumber','ratio','ratio','source'),
('source_aov','AOV / Nilai penjualan per pesanan','IDR','money','source')
) k(key,label,unit,kind,agg)
WHERE m.id IN ('tiktok-screenshot-v1','shopee-screenshot-v1') ON CONFLICT DO NOTHING;
UPDATE metric_definitions SET label=CASE
 WHEN key='sales' AND mapping_version='tiktok-screenshot-v1' THEN 'Perolehan GMV (Indirect internal; bukan LTV terverifikasi)'
 WHEN key='estimated_sales' THEN 'Perkiraan GMV (Direct internal; atribusi belum terverifikasi)'
 WHEN key='sales' THEN 'Penjualan · sesuai basis pesanan'
 WHEN key='views' AND mapping_version='shopee-screenshot-v1' THEN 'Dilihat · bukan impresi terverifikasi'
 WHEN key='views' THEN 'Tayangan · bukan unique viewers terverifikasi'
 WHEN key='viewers' THEN 'Penonton · bukan rata-rata concurrent'
 ELSE label END
WHERE mapping_version IN ('tiktok-screenshot-v1','shopee-screenshot-v1');

-- Older drafts can still be retried by the updated extractor.
INSERT INTO metric_definitions(mapping_version,key,label,unit,kind,aggregation,required,canonical_confirmed)
SELECT m.id,d.key,d.label,d.unit,d.kind,d.aggregation,false,false
FROM metric_mapping_versions m JOIN metric_definitions d
 ON d.mapping_version=m.channel_id||'-screenshot-v1'
WHERE m.id IN ('tiktok-source-v1','shopee-source-v1')
 AND d.key IN ('source_duration_seconds','source_sales_per_hour','source_comment_rate','source_ctor')
ON CONFLICT DO NOTHING;
