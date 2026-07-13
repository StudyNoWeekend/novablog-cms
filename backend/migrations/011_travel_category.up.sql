-- 给 travel_guides 表添加 category_id 字段，关联 categories 表
ALTER TABLE travel_guides ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE travel_guides ADD CONSTRAINT fk_travel_guides_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN travel_guides.category_id IS '分类ID，关联 categories 表';
CREATE INDEX IF NOT EXISTS idx_travel_guides_category_id ON travel_guides(category_id);
