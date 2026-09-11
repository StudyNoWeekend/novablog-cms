-- 为安全配置表新增安全防护开关字段，默认开启
ALTER TABLE security_configs ADD COLUMN IF NOT EXISTS security_enabled BOOLEAN NOT NULL DEFAULT TRUE;
