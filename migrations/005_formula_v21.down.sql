-- Preserve uploaded evidence on rollback. Remove only optional definitions that
-- have never been used, rather than deleting user-entered metric data.
DELETE FROM metric_definitions d WHERE key IN ('direct_gmv','indirect_gmv','unique_viewers','unique_buyers','avg_concurrent_viewers','avg_price_main_orders_k') AND NOT EXISTS(SELECT 1 FROM report_metric_values v WHERE v.metric_key=d.key AND v.version_id IN(SELECT id FROM report_versions WHERE definition_version=d.mapping_version));
