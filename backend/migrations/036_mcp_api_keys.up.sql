-- MCP 密钥表：用户自助开通 API Key，通过 MCP Streamable HTTP 端点远程发布内容。
-- 密钥明文不落库，仅存储 SHA-256 哈希；key_hash 有唯一索引，因此不做软删除。
CREATE TABLE IF NOT EXISTS mcp_api_keys (
    id            uuid PRIMARY KEY,
    name          varchar(100) NOT NULL,
    key_hash      varchar(64)  NOT NULL,
    key_prefix    varchar(24)  NOT NULL,
    status        smallint     NOT NULL DEFAULT 1,
    last_used_at  timestamptz,
    call_count    bigint       NOT NULL DEFAULT 0,
    expires_at    timestamptz,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_mcp_api_keys_key_hash ON mcp_api_keys (key_hash);
CREATE INDEX IF NOT EXISTS idx_mcp_api_keys_created_at ON mcp_api_keys (created_at DESC);

COMMENT ON TABLE mcp_api_keys IS 'MCP API 密钥（status: 1启用 2禁用）';
