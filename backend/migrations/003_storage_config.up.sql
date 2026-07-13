-- 对象存储配置及迁移任务相关表迁移

-- 存储配置表
CREATE TABLE IF NOT EXISTS storage_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(20) UNIQUE NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    region VARCHAR(50),
    bucket VARCHAR(255) NOT NULL,
    access_key VARCHAR(255) NOT NULL,
    access_secret TEXT NOT NULL,
    path_prefix VARCHAR(255),
    custom_domain VARCHAR(255),
    extra TEXT,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 部分唯一索引：保证同一时间只有一个 active 配置
CREATE UNIQUE INDEX IF NOT EXISTS storage_configs_one_active
    ON storage_configs (is_active)
    WHERE is_active = true;

-- 存储迁移任务表
CREATE TABLE IF NOT EXISTS storage_migration_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_type VARCHAR(20) NOT NULL,
    target_provider VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total INT NOT NULL DEFAULT 0,
    succeeded INT NOT NULL DEFAULT 0,
    failed INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 存储迁移条目表
CREATE TABLE IF NOT EXISTS storage_migration_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES storage_migration_tasks(id) ON DELETE CASCADE,
    media_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_storage_migration_items_task_id ON storage_migration_items(task_id);

-- media 表新增 storage_type 列
ALTER TABLE media ADD COLUMN IF NOT EXISTS storage_type VARCHAR(20) DEFAULT 'local';

-- 回填 storage_type
UPDATE media SET storage_type = 'local' WHERE storage_type IS NULL;

-- 表注释
COMMENT ON TABLE storage_configs IS '对象存储配置表';
COMMENT ON TABLE storage_migration_tasks IS '存储迁移任务表';
COMMENT ON TABLE storage_migration_items IS '存储迁移条目表';

-- storage_configs 字段注释
COMMENT ON COLUMN storage_configs.id IS '配置唯一标识';
COMMENT ON COLUMN storage_configs.provider IS '存储提供商：aliyun/tencent/minio';
COMMENT ON COLUMN storage_configs.endpoint IS '访问端点';
COMMENT ON COLUMN storage_configs.region IS '区域';
COMMENT ON COLUMN storage_configs.bucket IS '存储桶名称';
COMMENT ON COLUMN storage_configs.access_key IS '访问密钥ID';
COMMENT ON COLUMN storage_configs.access_secret IS '访问密钥（AES加密后的密文）';
COMMENT ON COLUMN storage_configs.path_prefix IS '对象key前缀，如 backend/';
COMMENT ON COLUMN storage_configs.custom_domain IS '自定义访问域名';
COMMENT ON COLUMN storage_configs.extra IS '扩展参数（JSON，存放平台特有参数，如 MinIO 的 use_ssl）';
COMMENT ON COLUMN storage_configs.is_active IS '是否为当前激活的配置';
COMMENT ON COLUMN storage_configs.created_at IS '创建时间';
COMMENT ON COLUMN storage_configs.updated_at IS '更新时间';

-- storage_migration_tasks 字段注释
COMMENT ON COLUMN storage_migration_tasks.id IS '任务唯一标识';
COMMENT ON COLUMN storage_migration_tasks.task_type IS '任务类型：analyze/migrate';
COMMENT ON COLUMN storage_migration_tasks.target_provider IS '目标存储提供商：aliyun/tencent/minio';
COMMENT ON COLUMN storage_migration_tasks.status IS '任务状态：pending/running/completed/failed/canceled';
COMMENT ON COLUMN storage_migration_tasks.total IS '待迁移总数';
COMMENT ON COLUMN storage_migration_tasks.succeeded IS '成功数量';
COMMENT ON COLUMN storage_migration_tasks.failed IS '失败数量';
COMMENT ON COLUMN storage_migration_tasks.started_at IS '开始时间';
COMMENT ON COLUMN storage_migration_tasks.finished_at IS '完成时间';
COMMENT ON COLUMN storage_migration_tasks.error IS '错误信息';
COMMENT ON COLUMN storage_migration_tasks.created_at IS '创建时间';
COMMENT ON COLUMN storage_migration_tasks.updated_at IS '更新时间';

-- storage_migration_items 字段注释
COMMENT ON COLUMN storage_migration_items.id IS '条目唯一标识';
COMMENT ON COLUMN storage_migration_items.task_id IS '所属任务ID';
COMMENT ON COLUMN storage_migration_items.media_id IS '媒体文件ID';
COMMENT ON COLUMN storage_migration_items.status IS '条目状态：pending/success/failed';
COMMENT ON COLUMN storage_migration_items.error IS '错误信息';
COMMENT ON COLUMN storage_migration_items.created_at IS '创建时间';

-- media 表字段注释
COMMENT ON COLUMN media.storage_type IS '存储类型：local/aliyun/tencent/minio';
