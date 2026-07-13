-- 音乐播放器模块迁移

-- 歌曲表
CREATE TABLE IF NOT EXISTS songs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    cover_url TEXT,
    bvid VARCHAR(50) NOT NULL,
    cid BIGINT NOT NULL,
    source_url TEXT NOT NULL,
    source_type VARCHAR(50) DEFAULT 'bilibili',
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    duration INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE songs IS '歌曲表';
COMMENT ON COLUMN songs.id IS '歌曲ID';
COMMENT ON COLUMN songs.title IS '歌曲标题';
COMMENT ON COLUMN songs.artist IS '艺术家';
COMMENT ON COLUMN songs.cover_url IS '封面URL';
COMMENT ON COLUMN songs.bvid IS 'B站视频BV号';
COMMENT ON COLUMN songs.cid IS 'B站视频CID';
COMMENT ON COLUMN songs.source_url IS '音频源URL';
COMMENT ON COLUMN songs.source_type IS '源类型，默认bilibili';
COMMENT ON COLUMN songs.category_id IS '分类ID，关联categories';
COMMENT ON COLUMN songs.duration IS '时长（秒）';
COMMENT ON COLUMN songs.sort_order IS '排序';
COMMENT ON COLUMN songs.created_at IS '创建时间';
COMMENT ON COLUMN songs.updated_at IS '更新时间';

CREATE INDEX idx_songs_category_id ON songs(category_id);
