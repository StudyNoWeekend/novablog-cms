-- 媒体库文件夹化：媒体文件夹表 + 媒体表归属列

CREATE TABLE IF NOT EXISTS media_folders (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    parent_id UUID,
    module_key VARCHAR(50),
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_media_folders_parent_id ON media_folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_media_folders_module_key ON media_folders(module_key);

ALTER TABLE media ADD COLUMN IF NOT EXISTS folder_id UUID;

CREATE INDEX IF NOT EXISTS idx_media_folder_id ON media(folder_id);

COMMENT ON TABLE media_folders IS '媒体文件夹（多级树，模块顶级文件夹由启动播种生成）';
COMMENT ON COLUMN media_folders.id IS '文件夹唯一标识';
COMMENT ON COLUMN media_folders.name IS '文件夹名称（同级唯一，禁止路径分隔符等字符）';
COMMENT ON COLUMN media_folders.parent_id IS '父文件夹 ID，NULL=顶级';
COMMENT ON COLUMN media_folders.module_key IS '业务模块 key（如 recipe），标记模块专属文件夹';
COMMENT ON COLUMN media_folders.sort IS '排序权重';
COMMENT ON COLUMN media_folders.created_at IS '创建时间';
COMMENT ON COLUMN media_folders.updated_at IS '更新时间';
COMMENT ON COLUMN media.folder_id IS '所属媒体文件夹 ID，NULL=根目录';
