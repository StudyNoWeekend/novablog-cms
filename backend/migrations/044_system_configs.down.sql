-- 回滚：删除通用系统配置表（站点地址/代理等将回退到 config.yaml 首装默认值）
DROP TABLE IF EXISTS system_configs;
