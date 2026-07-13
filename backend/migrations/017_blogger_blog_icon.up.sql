-- 为 bloggers 表新增博客 icon 图字段
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS blog_icon VARCHAR(500);

COMMENT ON COLUMN bloggers.blog_icon IS '博客 icon 图 URL';
