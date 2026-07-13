-- 回滚 categories 表的 type 字段
DROP INDEX IF EXISTS idx_categories_type;
ALTER TABLE categories DROP COLUMN IF EXISTS type;
