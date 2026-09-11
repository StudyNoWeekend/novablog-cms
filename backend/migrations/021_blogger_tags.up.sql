-- 为 bloggers 表新增 tags 字段
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS tags JSONB;
COMMENT ON COLUMN bloggers.tags IS '标签 JSON 字符串数组';
