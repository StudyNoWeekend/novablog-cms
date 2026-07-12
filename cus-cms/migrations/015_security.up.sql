-- 安全配置表：存储限流与黑名单相关配置
CREATE TABLE IF NOT EXISTS security_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    get_max_tokens INT NOT NULL DEFAULT 20,
    get_window_seconds INT NOT NULL DEFAULT 60,
    post_max_tokens INT NOT NULL DEFAULT 5,
    post_window_seconds INT NOT NULL DEFAULT 60,
    view_max_tokens INT NOT NULL DEFAULT 10,
    view_window_seconds INT NOT NULL DEFAULT 60,
    like_max_tokens INT NOT NULL DEFAULT 10,
    like_window_seconds INT NOT NULL DEFAULT 60,
    blacklist_threshold INT NOT NULL DEFAULT 5,
    blacklist_ttl_minutes INT NOT NULL DEFAULT 60,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 插入默认配置行
INSERT INTO security_configs (id) VALUES (gen_random_uuid())
ON CONFLICT DO NOTHING;

-- IP 黑名单表：记录被限制访问的 IP
CREATE TABLE IF NOT EXISTS ip_blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address VARCHAR(50) NOT NULL,
    reason VARCHAR(200),
    banned_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ip_blacklist_ip_address ON ip_blacklist(ip_address);
CREATE INDEX IF NOT EXISTS idx_ip_blacklist_is_active ON ip_blacklist(is_active);
