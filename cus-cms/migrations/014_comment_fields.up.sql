-- 重命名 email 列为 website，删除 avatar 列
ALTER TABLE comments RENAME COLUMN email TO website;
ALTER TABLE comments DROP COLUMN IF EXISTS avatar;
