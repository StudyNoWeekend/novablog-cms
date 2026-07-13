-- 作品集封面模式迁移
-- cover_mode: 0=使用排序第一的作品作为封面, 1=独立设置封面（使用 cover_preset_id）

ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS cover_mode INT NOT NULL DEFAULT 0;

COMMENT ON COLUMN portfolios.cover_mode IS '封面模式：0=使用排序第一的作品作为封面, 1=独立设置封面';
