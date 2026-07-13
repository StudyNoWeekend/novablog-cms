-- 摄影作品集表迁移

-- 作品集主表
CREATE TABLE IF NOT EXISTS portfolios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cover_preset_id UUID,
    status INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_portfolios_deleted_at ON portfolios(deleted_at);
CREATE INDEX IF NOT EXISTS idx_portfolios_status ON portfolios(status);
CREATE INDEX IF NOT EXISTS idx_portfolios_sort_order ON portfolios(sort_order);

COMMENT ON TABLE portfolios IS '摄影作品集表，存储作品集元数据';
COMMENT ON COLUMN portfolios.id IS '作品集唯一标识';
COMMENT ON COLUMN portfolios.name IS '作品集名称';
COMMENT ON COLUMN portfolios.description IS '作品集介绍';
COMMENT ON COLUMN portfolios.cover_preset_id IS '封面预设ID，引用 media_presets.id';
COMMENT ON COLUMN portfolios.status IS '状态：0=草稿, 1=已发布';
COMMENT ON COLUMN portfolios.sort_order IS '排序权重';
COMMENT ON COLUMN portfolios.created_at IS '创建时间';
COMMENT ON COLUMN portfolios.updated_at IS '更新时间';
COMMENT ON COLUMN portfolios.deleted_at IS '软删除时间';

-- 作品项表
CREATE TABLE IF NOT EXISTS portfolio_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    preset_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_portfolio_items_portfolio_id ON portfolio_items(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_items_preset_id ON portfolio_items(preset_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_items_deleted_at ON portfolio_items(deleted_at);
CREATE INDEX IF NOT EXISTS idx_portfolio_items_sort_order ON portfolio_items(sort_order);

COMMENT ON TABLE portfolio_items IS '作品项表，存储作品集下的每一项作品';
COMMENT ON COLUMN portfolio_items.id IS '作品项唯一标识';
COMMENT ON COLUMN portfolio_items.portfolio_id IS '所属作品集ID';
COMMENT ON COLUMN portfolio_items.preset_id IS '关联的媒体预设ID，引用 media_presets.id';
COMMENT ON COLUMN portfolio_items.title IS '作品名称';
COMMENT ON COLUMN portfolio_items.description IS '作品介绍';
COMMENT ON COLUMN portfolio_items.sort_order IS '排序序号，按添加顺序递增';
COMMENT ON COLUMN portfolio_items.created_at IS '创建时间';
COMMENT ON COLUMN portfolio_items.updated_at IS '更新时间';
COMMENT ON COLUMN portfolio_items.deleted_at IS '软删除时间';
