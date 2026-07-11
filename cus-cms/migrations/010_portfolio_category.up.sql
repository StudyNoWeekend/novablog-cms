-- 给 portfolios 表添加 category_id 字段，关联 categories 表
ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE portfolios ADD CONSTRAINT fk_portfolios_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN portfolios.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_portfolios_category_id ON portfolios(category_id);
