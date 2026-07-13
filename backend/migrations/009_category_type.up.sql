-- 给 categories 表添加 type 字段，用于区分不同模块的分类

ALTER TABLE categories ADD COLUMN IF NOT EXISTS type VARCHAR(50) DEFAULT 'article';
COMMENT ON COLUMN categories.type IS '分类类型(article/music/travel/portfolio)';
CREATE INDEX IF NOT EXISTS idx_categories_type ON categories(type);
