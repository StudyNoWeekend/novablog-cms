-- 为视频作品、读书书架、美食菜谱、游戏库、项目经历、开源作品补齐分类能力
-- 并将 categories.slug 的全局唯一约束改为 (type, slug) 联合唯一，允许不同模块使用同名分类

-- 1. categories.name/slug 放宽到 varchar(100)，与 GORM 模型一致，避免项目经历回填超长文本失败
ALTER TABLE categories ALTER COLUMN name TYPE VARCHAR(100);
ALTER TABLE categories ALTER COLUMN slug TYPE VARCHAR(100);

-- 2. 各内容表添加 category_id 字段
ALTER TABLE video_works ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE video_works ADD CONSTRAINT fk_video_works_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN video_works.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_video_works_category_id ON video_works(category_id);

ALTER TABLE books ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE books ADD CONSTRAINT fk_books_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN books.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_books_category_id ON books(category_id);

ALTER TABLE recipes ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE recipes ADD CONSTRAINT fk_recipes_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN recipes.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_recipes_category_id ON recipes(category_id);

ALTER TABLE games ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE games ADD CONSTRAINT fk_games_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN games.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_games_category_id ON games(category_id);

ALTER TABLE projects ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE projects ADD CONSTRAINT fk_projects_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN projects.category_id IS '分类ID，关联 categories 表（结构化分类，替代原自由文本 category）';
CREATE INDEX IF NOT EXISTS idx_projects_category_id ON projects(category_id);

ALTER TABLE open_source_works ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE open_source_works ADD CONSTRAINT fk_open_source_works_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN open_source_works.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_open_source_works_category_id ON open_source_works(category_id);

-- 3. 项目经历：将已有自由文本 category 回填为 type=project 的结构化分类
INSERT INTO categories (id, name, slug, type)
SELECT uuid_generate_v4(), c.category, c.category, 'project'
FROM (
    SELECT DISTINCT category
    FROM projects
    WHERE category IS NOT NULL AND category <> ''
) c
WHERE NOT EXISTS (
    SELECT 1 FROM categories cat WHERE cat.type = 'project' AND cat.name = c.category
);

UPDATE projects p
SET category_id = cat.id
FROM categories cat
WHERE cat.type = 'project'
  AND cat.name = p.category
  AND p.category IS NOT NULL AND p.category <> ''
  AND p.category_id IS NULL;

-- 4. slug 唯一性从全局改为 (type, slug)，跨模块同名分类不再冲突
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_slug_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_type_slug ON categories(type, slug);
