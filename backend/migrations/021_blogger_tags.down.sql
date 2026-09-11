-- 回滚：删除 bloggers 表的 tags 字段
ALTER TABLE bloggers DROP COLUMN IF EXISTS tags;
