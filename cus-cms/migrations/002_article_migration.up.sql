-- 文章相关表迁移

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    slug VARCHAR(50) UNIQUE,
    description VARCHAR(255),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 文章表
CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE,
    summary VARCHAR(500),
    content TEXT NOT NULL,
    cover_image VARCHAR(500),
    category_id UUID REFERENCES categories(id),
    status SMALLINT DEFAULT 1,
    type SMALLINT DEFAULT 1,
    extra JSONB,
    view_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    is_top BOOLEAN DEFAULT false,
    is_comment BOOLEAN DEFAULT true,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- 文章标签关联表
CREATE TABLE IF NOT EXISTS article_tags (
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);

-- 媒体资源表
CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filename VARCHAR(255) NOT NULL,
    file_type SMALLINT NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT,
    url VARCHAR(500) NOT NULL,
    thumb_url VARCHAR(500),
    width INT,
    height INT,
    storage_path VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category_id);
CREATE INDEX IF NOT EXISTS idx_articles_published ON articles(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_media_type ON media(file_type);

-- 表注释
COMMENT ON TABLE categories IS '文章分类表';
COMMENT ON TABLE tags IS '文章标签表';
COMMENT ON TABLE articles IS '文章表';
COMMENT ON TABLE article_tags IS '文章标签关联表';
COMMENT ON TABLE media IS '媒体资源表';

-- categories 字段注释
COMMENT ON COLUMN categories.id IS '分类唯一标识';
COMMENT ON COLUMN categories.name IS '分类名称';
COMMENT ON COLUMN categories.slug IS '分类别名（URL友好标识）';
COMMENT ON COLUMN categories.description IS '分类描述';
COMMENT ON COLUMN categories.sort_order IS '排序序号';
COMMENT ON COLUMN categories.created_at IS '创建时间';

-- tags 字段注释
COMMENT ON COLUMN tags.id IS '标签唯一标识';
COMMENT ON COLUMN tags.name IS '标签名称';
COMMENT ON COLUMN tags.created_at IS '创建时间';

-- articles 字段注释
COMMENT ON COLUMN articles.id IS '文章唯一标识';
COMMENT ON COLUMN articles.title IS '文章标题';
COMMENT ON COLUMN articles.slug IS '文章别名（URL友好标识）';
COMMENT ON COLUMN articles.summary IS '文章摘要';
COMMENT ON COLUMN articles.content IS '文章内容';
COMMENT ON COLUMN articles.cover_image IS '封面图URL';
COMMENT ON COLUMN articles.category_id IS '所属分类ID';
COMMENT ON COLUMN articles.status IS '文章状态：1草稿 2已发布 3已下架';
COMMENT ON COLUMN articles.type IS '编辑器类型：1Markdown 2富文本';
COMMENT ON COLUMN articles.extra IS '扩展字段（JSON格式）';
COMMENT ON COLUMN articles.view_count IS '浏览次数';
COMMENT ON COLUMN articles.comment_count IS '评论次数';
COMMENT ON COLUMN articles.is_top IS '是否置顶';
COMMENT ON COLUMN articles.is_comment IS '是否允许评论';
COMMENT ON COLUMN articles.published_at IS '发布时间';
COMMENT ON COLUMN articles.created_at IS '创建时间';
COMMENT ON COLUMN articles.updated_at IS '更新时间';
COMMENT ON COLUMN articles.deleted_at IS '软删除时间';

-- article_tags 字段注释
COMMENT ON COLUMN article_tags.article_id IS '文章ID';
COMMENT ON COLUMN article_tags.tag_id IS '标签ID';

-- media 字段注释
COMMENT ON COLUMN media.id IS '媒体唯一标识';
COMMENT ON COLUMN media.filename IS '文件名';
COMMENT ON COLUMN media.file_type IS '文件类型：1图片 2视频 3音频';
COMMENT ON COLUMN media.mime_type IS 'MIME类型';
COMMENT ON COLUMN media.size IS '文件大小（字节）';
COMMENT ON COLUMN media.url IS '文件访问URL';
COMMENT ON COLUMN media.thumb_url IS '缩略图URL';
COMMENT ON COLUMN media.width IS '宽度（像素）';
COMMENT ON COLUMN media.height IS '高度（像素）';
COMMENT ON COLUMN media.storage_path IS '存储路径';
COMMENT ON COLUMN media.created_at IS '创建时间';
COMMENT ON COLUMN media.deleted_at IS '软删除时间';