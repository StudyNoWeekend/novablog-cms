-- 恢复 CORS 原有默认白名单；仅回滚从未被修改过的播种行（updated_at = created_at）
ALTER TABLE cors_configs
    ALTER COLUMN allowed_origins SET DEFAULT 'http://localhost:5173,http://localhost:5174';

UPDATE cors_configs
SET allowed_origins = 'http://localhost:5173,http://localhost:5174'
WHERE updated_at = created_at;
