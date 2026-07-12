-- 博主表新增页面背景图字段
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS page_background VARCHAR(500);
COMMENT ON COLUMN bloggers.page_background IS '页面背景图URL';
