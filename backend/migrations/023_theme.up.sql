-- 主题安装实例表（同主题多版本共存）
CREATE TABLE IF NOT EXISTS themes (
    id            UUID PRIMARY KEY,
    theme_id      VARCHAR(100) NOT NULL,
    name          VARCHAR(100) NOT NULL,
    version       VARCHAR(50)  NOT NULL,
    engine        VARCHAR(50)  NOT NULL,
    api_compat    VARCHAR(20)  NOT NULL,
    author        VARCHAR(100),
    description   TEXT,
    screenshots   JSONB,
    fallbacks     JSONB,
    source        VARCHAR(20)  NOT NULL,
    source_ref    VARCHAR(500),
    market_id     BIGINT,
    market_slug   VARCHAR(100),
    artifact_path VARCHAR(500) NOT NULL,
    checksum      VARCHAR(64),
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT uk_themes_id_version UNIQUE (theme_id, version)
);

CREATE INDEX IF NOT EXISTS idx_themes_market_id ON themes (market_id);

-- 激活指针：引用 themes.id（具体安装实例）
ALTER TABLE bloggers ADD COLUMN IF NOT EXISTS active_theme_id UUID;
