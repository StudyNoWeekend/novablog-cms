-- CORS 跨域配置表（单行模式）
CREATE TABLE IF NOT EXISTS cors_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    allowed_origins TEXT NOT NULL DEFAULT 'http://localhost:5173,http://localhost:5174',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认配置行（单行模式）
INSERT INTO cors_configs (id) VALUES (gen_random_uuid()) ON CONFLICT DO NOTHING;
