-- Tüm tabloları ve ilişkili nesneleri sil (rollback için)

DROP TRIGGER IF EXISTS update_content_stats_updated_at ON content_stats;
DROP TRIGGER IF EXISTS update_contents_updated_at ON contents;
DROP TRIGGER IF EXISTS update_providers_updated_at ON providers;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_sync_logs_started;
DROP INDEX IF EXISTS idx_sync_logs_provider;
DROP INDEX IF EXISTS idx_content_tags_tag;
DROP INDEX IF EXISTS idx_content_tags_content;
DROP INDEX IF EXISTS idx_tags_name;
DROP INDEX IF EXISTS idx_scores_content_id;
DROP INDEX IF EXISTS idx_scores_final;
DROP INDEX IF EXISTS idx_stats_content_id;
DROP INDEX IF EXISTS idx_stats_views;
DROP INDEX IF EXISTS idx_contents_provider;
DROP INDEX IF EXISTS idx_contents_published;
DROP INDEX IF EXISTS idx_contents_type;
DROP INDEX IF EXISTS idx_contents_title_pattern;
DROP INDEX IF EXISTS idx_contents_title;

DROP TABLE IF EXISTS scoring_rules;
DROP TABLE IF EXISTS provider_sync_logs;
DROP TABLE IF EXISTS content_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS content_scores;
DROP TABLE IF EXISTS content_stats;
DROP TABLE IF EXISTS contents;
DROP TABLE IF EXISTS providers;
