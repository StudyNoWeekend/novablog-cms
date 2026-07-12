-- 视频作品表迁移

-- 视频作品主表
CREATE TABLE IF NOT EXISTS video_works (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    cover_url VARCHAR(1024),
    description TEXT,
    status INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_video_works_deleted_at ON video_works(deleted_at);
CREATE INDEX IF NOT EXISTS idx_video_works_status ON video_works(status);
CREATE INDEX IF NOT EXISTS idx_video_works_sort_order ON video_works(sort_order);

COMMENT ON TABLE video_works IS '视频作品表，存储视频元数据';
COMMENT ON COLUMN video_works.id IS '视频作品唯一标识';
COMMENT ON COLUMN video_works.title IS '视频标题';
COMMENT ON COLUMN video_works.cover_url IS '封面图地址';
COMMENT ON COLUMN video_works.description IS '视频介绍';
COMMENT ON COLUMN video_works.status IS '状态：0=草稿, 1=已发布';
COMMENT ON COLUMN video_works.sort_order IS '排序权重';
COMMENT ON COLUMN video_works.created_at IS '创建时间';
COMMENT ON COLUMN video_works.updated_at IS '更新时间';
COMMENT ON COLUMN video_works.deleted_at IS '软删除时间';

-- 视频平台链接表
CREATE TABLE IF NOT EXISTS video_platform_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES video_works(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    url VARCHAR(1024) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_video_platform_links_video_id ON video_platform_links(video_id);
CREATE INDEX IF NOT EXISTS idx_video_platform_links_platform ON video_platform_links(platform);
CREATE INDEX IF NOT EXISTS idx_video_platform_links_deleted_at ON video_platform_links(deleted_at);

COMMENT ON TABLE video_platform_links IS '视频平台链接表，存储视频在各平台的发布链接';
COMMENT ON COLUMN video_platform_links.id IS '平台链接唯一标识';
COMMENT ON COLUMN video_platform_links.video_id IS '所属视频作品ID';
COMMENT ON COLUMN video_platform_links.platform IS '平台标识，如 bilibili、youtube 等';
COMMENT ON COLUMN video_platform_links.url IS '平台播放链接';
COMMENT ON COLUMN video_platform_links.created_at IS '创建时间';
COMMENT ON COLUMN video_platform_links.updated_at IS '更新时间';
COMMENT ON COLUMN video_platform_links.deleted_at IS '软删除时间';
