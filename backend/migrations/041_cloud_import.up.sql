-- 云盘导入：导入任务表 + 导入条目表

CREATE TABLE IF NOT EXISTS cloud_import_tasks (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    provider VARCHAR(20) NOT NULL,
    target_folder_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total INT DEFAULT 0,
    succeeded INT DEFAULT 0,
    failed INT DEFAULT 0,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cloud_import_tasks_account_id ON cloud_import_tasks(account_id);
CREATE INDEX IF NOT EXISTS idx_cloud_import_tasks_status ON cloud_import_tasks(status);

CREATE TABLE IF NOT EXISTS cloud_import_items (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL,
    source_file_id VARCHAR(255) NOT NULL,
    source_path VARCHAR(500),
    filename VARCHAR(255),
    file_size BIGINT DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    media_id UUID,
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cloud_import_items_task_id ON cloud_import_items(task_id);

COMMENT ON TABLE cloud_import_tasks IS '云盘导入任务（云盘文件转存媒体库，创建即 running，进度走 DB 计数器）';
COMMENT ON COLUMN cloud_import_tasks.target_folder_id IS '目标媒体文件夹，NULL=媒体库根目录';
COMMENT ON TABLE cloud_import_items IS '云盘导入条目（单个云盘文件的转存明细）';
COMMENT ON COLUMN cloud_import_items.source_file_id IS '盘方文件 ID（WebDAV 为文件路径）';
