-- 技术栈与健身训练迁入统一分类表：
-- 两表加 category_id 外键，并把旧自由文本/固定枚举值回填为 type=tech_stack / type=fitness 的结构化分类。
-- 旧字段（tech_stack_items.category、fitness_records.type）保留作兼容镜像，由逻辑层同步为分类 slug。

-- 1. 两表添加 category_id 字段
ALTER TABLE tech_stack_items ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE tech_stack_items ADD CONSTRAINT fk_tech_stack_items_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN tech_stack_items.category_id IS '分类ID，关联 categories 表（type=tech_stack）';
CREATE INDEX IF NOT EXISTS idx_tech_stack_items_category_id ON tech_stack_items(category_id);

ALTER TABLE fitness_records ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE fitness_records ADD CONSTRAINT fk_fitness_records_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
COMMENT ON COLUMN fitness_records.category_id IS '分类ID，关联 categories 表（type=fitness）';
CREATE INDEX IF NOT EXISTS idx_fitness_records_category_id ON fitness_records(category_id);

-- 2. 技术栈：已知 5 个约定值按固定中文名播种，其余自由值 name=slug=原值
INSERT INTO categories (id, name, slug, type)
SELECT uuid_generate_v4(), m.name, m.slug, 'tech_stack'
FROM (VALUES
    ('language', '语言'),
    ('framework', '框架'),
    ('tool', '工具'),
    ('database', '数据库'),
    ('other', '其他')
) AS m(slug, name)
WHERE EXISTS (SELECT 1 FROM tech_stack_items t WHERE t.category = m.slug)
  AND NOT EXISTS (
    SELECT 1 FROM categories c
    WHERE c.type = 'tech_stack' AND (c.slug = m.slug OR c.name = m.name)
  );

INSERT INTO categories (id, name, slug, type)
SELECT uuid_generate_v4(), d.category, d.category, 'tech_stack'
FROM (
    SELECT DISTINCT category
    FROM tech_stack_items
    WHERE category IS NOT NULL AND category <> ''
) d
WHERE NOT EXISTS (
    SELECT 1 FROM categories c
    WHERE c.type = 'tech_stack' AND (c.slug = d.category OR c.name = d.category)
);

UPDATE tech_stack_items t
SET category_id = c.id
FROM categories c
WHERE c.type = 'tech_stack'
  AND (c.slug = t.category OR c.name = t.category)
  AND t.category IS NOT NULL AND t.category <> ''
  AND t.category_id IS NULL;

-- 3. 健身训练：strength/cardio/stretch 播种为中文名分类
INSERT INTO categories (id, name, slug, type)
SELECT uuid_generate_v4(), m.name, m.slug, 'fitness'
FROM (VALUES
    ('strength', '力量'),
    ('cardio', '有氧'),
    ('stretch', '拉伸')
) AS m(slug, name)
WHERE EXISTS (SELECT 1 FROM fitness_records f WHERE f.type = m.slug)
  AND NOT EXISTS (
    SELECT 1 FROM categories c
    WHERE c.type = 'fitness' AND (c.slug = m.slug OR c.name = m.name)
  );

UPDATE fitness_records f
SET category_id = c.id
FROM categories c
WHERE c.type = 'fitness'
  AND (c.slug = f.type OR c.name = f.type)
  AND f.type IS NOT NULL AND f.type <> ''
  AND f.category_id IS NULL;
