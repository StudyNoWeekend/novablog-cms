-- 官方主题市场配置表（单行模式）
CREATE TABLE IF NOT EXISTS theme_market_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_base_url VARCHAR(500) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认配置行（单行模式；market_base_url 为空表示未自定义，使用 config.yaml 出厂值）
INSERT INTO theme_market_configs (id) VALUES (gen_random_uuid()) ON CONFLICT DO NOTHING;
