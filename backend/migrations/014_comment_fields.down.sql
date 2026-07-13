-- 恢复 website 列为 email，重新添加 avatar 列
ALTER TABLE comments RENAME COLUMN website TO email;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS avatar VARCHAR(500);
