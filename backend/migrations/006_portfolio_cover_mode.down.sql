-- 回滚作品集封面模式迁移

ALTER TABLE portfolios DROP COLUMN IF EXISTS cover_mode;
