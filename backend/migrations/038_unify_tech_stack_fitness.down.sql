-- 回滚：移除两表的 category_id 字段；对应播种的分类一并清理（此时已无外键引用）
ALTER TABLE tech_stack_items DROP CONSTRAINT IF EXISTS fk_tech_stack_items_category;
ALTER TABLE tech_stack_items DROP COLUMN IF EXISTS category_id;

ALTER TABLE fitness_records DROP CONSTRAINT IF EXISTS fk_fitness_records_category;
ALTER TABLE fitness_records DROP COLUMN IF EXISTS category_id;

DELETE FROM categories WHERE type = 'tech_stack';
DELETE FROM categories WHERE type = 'fitness';
