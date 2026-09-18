-- 第三方歌单表
CREATE TABLE IF NOT EXISTS third_party_playlists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    cover_url TEXT,
    platform VARCHAR(50) NOT NULL,
    platform_url TEXT NOT NULL,
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE third_party_playlists IS '第三方歌单表';
COMMENT ON COLUMN third_party_playlists.id IS '歌单ID';
COMMENT ON COLUMN third_party_playlists.title IS '歌单名称';
COMMENT ON COLUMN third_party_playlists.cover_url IS '歌单封面URL';
COMMENT ON COLUMN third_party_playlists.platform IS '平台标识：qq_music, netease, bilibili, spotify, apple_music, other';
COMMENT ON COLUMN third_party_playlists.platform_url IS '歌单原始链接';
COMMENT ON COLUMN third_party_playlists.description IS '简短描述';
COMMENT ON COLUMN third_party_playlists.sort_order IS '排序';
COMMENT ON COLUMN third_party_playlists.enabled IS '是否在前台展示';
COMMENT ON COLUMN third_party_playlists.created_at IS '创建时间';
COMMENT ON COLUMN third_party_playlists.updated_at IS '更新时间';

CREATE INDEX idx_third_party_playlists_enabled ON third_party_playlists(enabled);
