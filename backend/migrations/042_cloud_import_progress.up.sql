-- 云盘导入进度：导入条目增加已下载字节数列

ALTER TABLE cloud_import_items ADD COLUMN IF NOT EXISTS transferred_bytes BIGINT DEFAULT 0;

COMMENT ON COLUMN cloud_import_items.transferred_bytes IS '已下载字节数（下载中由导入协程节流回写，供前端逐文件进度展示）';
