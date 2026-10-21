-- 恢复 categories.slug 全局唯一约束（若已出现跨模块同名 slug 会失败，需先手工处理数据）
DROP INDEX IF EXISTS uq_categories_type_slug;
ALTER TABLE categories ADD CONSTRAINT categories_slug_key UNIQUE (slug);

-- 移除各内容表的 category_id 字段（projects.category 自由文本列保留，无需回补）
ALTER TABLE video_works DROP CONSTRAINT IF EXISTS fk_video_works_category;
ALTER TABLE video_works DROP COLUMN IF EXISTS category_id;

ALTER TABLE books DROP CONSTRAINT IF EXISTS fk_books_category;
ALTER TABLE books DROP COLUMN IF EXISTS category_id;

ALTER TABLE recipes DROP CONSTRAINT IF EXISTS fk_recipes_category;
ALTER TABLE recipes DROP COLUMN IF EXISTS category_id;

ALTER TABLE games DROP CONSTRAINT IF EXISTS fk_games_category;
ALTER TABLE games DROP COLUMN IF EXISTS category_id;

ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_projects_category;
ALTER TABLE projects DROP COLUMN IF EXISTS category_id;

ALTER TABLE open_source_works DROP CONSTRAINT IF EXISTS fk_open_source_works_category;
ALTER TABLE open_source_works DROP COLUMN IF EXISTS category_id;
