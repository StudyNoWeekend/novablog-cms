-- 地图服务配置表（单行模式）
CREATE TABLE IF NOT EXISTS map_configs (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amap_key           VARCHAR(200) NOT NULL DEFAULT '',
    amap_security_code TEXT NOT NULL DEFAULT '',
    google_key         VARCHAR(200) NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 插入默认配置行（单行模式；Key 为空表示未配置，选点自动降级 OSM 免 Key 模式）
INSERT INTO map_configs (id) VALUES (gen_random_uuid()) ON CONFLICT DO NOTHING;
