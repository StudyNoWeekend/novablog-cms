-- 全站访问量统计体系：内容表补浏览量字段 + 访问明细表 + 按日聚合表

-- 1. 其余 11 张内容表补 view_count（articles/travel_guides 已有）
ALTER TABLE portfolios        ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE video_works       ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE songs             ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE photo_equipment   ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE projects          ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE open_source_works ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE recipes           ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE books             ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE games             ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE fitness_records   ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
ALTER TABLE tech_stack_items  ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;

-- 2. 访问明细表（保留 180 天，由定时任务清理；按日聚合进 daily_view_stats 后历史不丢）
CREATE TABLE IF NOT EXISTS content_view_logs (
    id           BIGSERIAL PRIMARY KEY,
    content_type VARCHAR(32)  NOT NULL,              -- article/travel/portfolio/video/song/equipment/project/open_source/recipe/book/game/fitness/tech_stack
    content_id   UUID         NOT NULL,
    ip           VARCHAR(64)  NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_view_logs_type_id   ON content_view_logs (content_type, content_id);
CREATE INDEX IF NOT EXISTS idx_view_logs_created_at ON content_view_logs (created_at);

-- 3. 按日聚合表（每小时从明细表 rollup；uv 为当日去重 IP 数）
CREATE TABLE IF NOT EXISTS daily_view_stats (
    stat_date    DATE        NOT NULL,
    content_type VARCHAR(32) NOT NULL,
    pv           BIGINT      NOT NULL DEFAULT 0,
    uv           BIGINT      NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (stat_date, content_type)
);
