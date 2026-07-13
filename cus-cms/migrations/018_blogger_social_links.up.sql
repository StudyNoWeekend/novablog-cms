-- 为 bloggers 表新增社交平台链接字段
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS social_links JSON;
COMMENT ON COLUMN bloggers.social_links IS '社交平台链接 JSON 数组';
