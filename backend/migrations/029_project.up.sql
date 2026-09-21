-- 项目经历表迁移

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT '',
    role VARCHAR(100) NOT NULL DEFAULT '',
    client VARCHAR(255) NOT NULL DEFAULT '',
    cover_url VARCHAR(1024),
    summary VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT,
    tech_stack VARCHAR(500) NOT NULL DEFAULT '',
    start_date DATE,
    end_date DATE,
    project_url VARCHAR(1024),
    repo_url VARCHAR(1024),
    status INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX IF NOT EXISTS idx_projects_sort_order ON projects(sort_order);
CREATE INDEX IF NOT EXISTS idx_projects_status_sort ON projects(status, sort_order);
CREATE INDEX IF NOT EXISTS idx_projects_category ON projects(category);

COMMENT ON TABLE projects IS '项目经历表，存储通用项目经历信息（摄影/视频剪辑/技术开发等）';
COMMENT ON COLUMN projects.id IS '项目唯一标识';
COMMENT ON COLUMN projects.title IS '项目名称';
COMMENT ON COLUMN projects.category IS '领域分类：摄影/视频剪辑/技术开发/设计等';
COMMENT ON COLUMN projects.role IS '担任角色';
COMMENT ON COLUMN projects.client IS '客户/所属组织';
COMMENT ON COLUMN projects.cover_url IS '封面图地址';
COMMENT ON COLUMN projects.summary IS '一句话简介';
COMMENT ON COLUMN projects.description IS '详细描述（背景/成果/亮点）';
COMMENT ON COLUMN projects.tech_stack IS '技能/工具标签，逗号分隔';
COMMENT ON COLUMN projects.start_date IS '开始时间（按月粒度，存当月1号）';
COMMENT ON COLUMN projects.end_date IS '结束时间，NULL 表示至今';
COMMENT ON COLUMN projects.project_url IS '项目/作品在线链接';
COMMENT ON COLUMN projects.repo_url IS '代码仓库链接';
COMMENT ON COLUMN projects.status IS '状态：0=草稿, 1=已发布';
COMMENT ON COLUMN projects.sort_order IS '排序权重';
COMMENT ON COLUMN projects.created_at IS '创建时间';
COMMENT ON COLUMN projects.updated_at IS '更新时间';
COMMENT ON COLUMN projects.deleted_at IS '软删除时间';
