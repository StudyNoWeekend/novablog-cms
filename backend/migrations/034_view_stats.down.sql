-- 回滚访问量统计体系
DROP TABLE IF EXISTS daily_view_stats;
DROP TABLE IF EXISTS content_view_logs;
ALTER TABLE portfolios        DROP COLUMN IF EXISTS view_count;
ALTER TABLE video_works       DROP COLUMN IF EXISTS view_count;
ALTER TABLE songs             DROP COLUMN IF EXISTS view_count;
ALTER TABLE photo_equipment   DROP COLUMN IF EXISTS view_count;
ALTER TABLE projects          DROP COLUMN IF EXISTS view_count;
ALTER TABLE open_source_works DROP COLUMN IF EXISTS view_count;
ALTER TABLE recipes           DROP COLUMN IF EXISTS view_count;
ALTER TABLE books             DROP COLUMN IF EXISTS view_count;
ALTER TABLE games             DROP COLUMN IF EXISTS view_count;
ALTER TABLE fitness_records   DROP COLUMN IF EXISTS view_count;
ALTER TABLE tech_stack_items  DROP COLUMN IF EXISTS view_count;
