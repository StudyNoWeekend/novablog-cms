-- 媒体库文件夹化回滚

DROP INDEX IF EXISTS idx_media_folder_id;
ALTER TABLE media DROP COLUMN IF EXISTS folder_id;

DROP TABLE IF EXISTS media_folders;
