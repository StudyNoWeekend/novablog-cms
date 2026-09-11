-- 模块开关配置表
CREATE TABLE IF NOT EXISTS module_configs (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    media_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    music_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    video_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    travel_enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    portfolio_enabled  BOOLEAN NOT NULL DEFAULT TRUE,
    equipment_enabled  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认配置记录（确保单行模式）
INSERT INTO module_configs (id) VALUES (gen_random_uuid()) ON CONFLICT DO NOTHING;
