DELETE FROM metric_definitions WHERE mapping_version IN ('tiktok-screenshot-v1','shopee-screenshot-v1');
DELETE FROM metric_mapping_versions WHERE id IN ('tiktok-screenshot-v1','shopee-screenshot-v1');
DELETE FROM metric_definitions WHERE mapping_version IN ('tiktok-source-v1','shopee-source-v1') AND key IN ('source_duration_seconds','source_sales_per_hour','source_comment_rate');
DELETE FROM metric_definitions WHERE mapping_version='shopee-source-v1' AND key='source_ctor';
