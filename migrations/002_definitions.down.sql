DELETE FROM metric_definitions WHERE mapping_version IN('tiktok-source-v1','shopee-source-v1','tiktok-demo-v1','shopee-demo-v1');
DELETE FROM metric_mapping_versions WHERE id IN('tiktok-source-v1','shopee-source-v1','tiktok-demo-v1','shopee-demo-v1');
