-- 云盘导入：应用凭据配置表 + 授权账号表

CREATE TABLE IF NOT EXISTS cloud_drive_configs (
    id UUID PRIMARY KEY,
    provider VARCHAR(20) NOT NULL,
    client_id VARCHAR(255),
    client_secret TEXT,
    callback_base_url VARCHAR(255),
    extra TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cloud_drive_configs_provider ON cloud_drive_configs(provider);

CREATE TABLE IF NOT EXISTS cloud_drive_accounts (
    id UUID PRIMARY KEY,
    provider VARCHAR(20) NOT NULL,
    open_user_id VARCHAR(128) NOT NULL,
    display_name VARCHAR(128),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ,
    base_url VARCHAR(255),
    username VARCHAR(128),
    drive_id VARCHAR(64),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    quota_total BIGINT DEFAULT 0,
    quota_used BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_open_user
    ON cloud_drive_accounts(provider, open_user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cloud_drive_accounts_deleted_at ON cloud_drive_accounts(deleted_at);

COMMENT ON TABLE cloud_drive_configs IS '云盘应用凭据配置（自部署站长在开放平台申请后填写，百度/阿里各一行）';
COMMENT ON COLUMN cloud_drive_configs.client_secret IS 'AES 加密后的 client_secret';
COMMENT ON COLUMN cloud_drive_configs.callback_base_url IS '本实例外部可达地址，用于拼接 OAuth redirect_uri';
COMMENT ON TABLE cloud_drive_accounts IS '云盘授权账号（OAuth 盘存 token；WebDAV 复用 access_token 存加密密码）';
COMMENT ON COLUMN cloud_drive_accounts.open_user_id IS 'OAuth=盘方用户 ID；WebDAV=base_url|username';
COMMENT ON COLUMN cloud_drive_accounts.drive_id IS '阿里云盘 default_drive_id';
