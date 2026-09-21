-- 开源作品表迁移

CREATE TABLE IF NOT EXISTS open_source_works (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    repo_url VARCHAR(1024) NOT NULL,
    summary VARCHAR(500) NOT NULL DEFAULT '',
    readme TEXT,
    language VARCHAR(100) NOT NULL DEFAULT '',
    topics VARCHAR(500) NOT NULL DEFAULT '',
    stars INT NOT NULL DEFAULT 0,
    homepage VARCHAR(1024),
    status INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    readme_updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_open_source_works_deleted_at ON open_source_works(deleted_at);
CREATE INDEX IF NOT EXISTS idx_open_source_works_sort_order ON open_source_works(sort_order);
CREATE INDEX IF NOT EXISTS idx_open_source_works_status_sort ON open_source_works(status, sort_order);

COMMENT ON TABLE open_source_works IS '开源作品表，展示技术人员维护的开源仓库（名称/链接/README 自动拉取）';
COMMENT ON COLUMN open_source_works.id IS '作品唯一标识';
COMMENT ON COLUMN open_source_works.name IS '仓库名称';
COMMENT ON COLUMN open_source_works.repo_url IS '仓库链接（GitHub）';
COMMENT ON COLUMN open_source_works.summary IS '一句话介绍（为空时自动回填仓库描述）';
COMMENT ON COLUMN open_source_works.readme IS 'README 原文（Markdown，自动拉取）';
COMMENT ON COLUMN open_source_works.language IS '主语言';
COMMENT ON COLUMN open_source_works.topics IS '主题标签，逗号分隔';
COMMENT ON COLUMN open_source_works.stars IS 'Star 数（刷新时快照）';
COMMENT ON COLUMN open_source_works.homepage IS '主页/演示地址';
COMMENT ON COLUMN open_source_works.status IS '状态：0=草稿, 1=已发布';
COMMENT ON COLUMN open_source_works.sort_order IS '排序权重';
COMMENT ON COLUMN open_source_works.readme_updated_at IS 'README 最近拉取时间';
COMMENT ON COLUMN open_source_works.created_at IS '创建时间';
COMMENT ON COLUMN open_source_works.updated_at IS '更新时间';
COMMENT ON COLUMN open_source_works.deleted_at IS '软删除时间';
