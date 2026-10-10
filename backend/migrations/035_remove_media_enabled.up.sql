-- 移除 media_enabled：媒体库为基础设施（多个模块存放图片/视频的公共资源），
-- 恒可用、不参与模块开关管理，开放接口 /module-config 不再下发该开关字段。
ALTER TABLE module_configs DROP COLUMN IF EXISTS media_enabled;
