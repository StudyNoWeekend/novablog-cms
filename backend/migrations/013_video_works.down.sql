-- 回滚视频作品表迁移

DROP INDEX IF EXISTS idx_video_platform_links_deleted_at;
DROP INDEX IF EXISTS idx_video_platform_links_platform;
DROP INDEX IF EXISTS idx_video_platform_links_video_id;
DROP TABLE IF EXISTS video_platform_links;

DROP INDEX IF EXISTS idx_video_works_sort_order;
DROP INDEX IF EXISTS idx_video_works_status;
DROP INDEX IF EXISTS idx_video_works_deleted_at;
DROP TABLE IF EXISTS video_works;
