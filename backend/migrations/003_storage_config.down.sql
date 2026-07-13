-- 回滚对象存储配置及迁移任务相关表迁移

-- media 表删除 storage_type 列
ALTER TABLE media DROP COLUMN IF EXISTS storage_type;

-- 删除索引
DROP INDEX IF EXISTS idx_storage_migration_items_task_id;

-- 删除表
DROP TABLE IF EXISTS storage_migration_items;
DROP TABLE IF EXISTS storage_migration_tasks;

-- 删除部分唯一索引
DROP INDEX IF EXISTS storage_configs_one_active;

-- 删除存储配置表
DROP TABLE IF EXISTS storage_configs;
