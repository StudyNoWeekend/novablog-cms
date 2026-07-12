-- 回滚：删除 bloggers 表的 blog_icon 字段
ALTER TABLE bloggers DROP COLUMN IF EXISTS blog_icon;
