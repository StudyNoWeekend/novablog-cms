-- 回滚媒体预设表迁移

DROP INDEX IF EXISTS idx_media_presets_deleted_at;
DROP INDEX IF EXISTS idx_media_presets_media_id;
DROP TABLE IF EXISTS media_presets;
