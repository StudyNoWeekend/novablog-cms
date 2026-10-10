-- 回滚：恢复 media_enabled 开关（默认开启，幂等）
ALTER TABLE module_configs ADD COLUMN IF NOT EXISTS media_enabled BOOLEAN NOT NULL DEFAULT TRUE;
