-- 通用系统配置键值表：收纳原 config.yaml 中的运行时配置项
-- （站点对外地址、出网代理、GitHub Token、主题/云盘限额、geoip 下载源）。
-- 数据库为运行时配置唯一主源；存量值由启动种子逻辑从 config.yaml 幂等导入
-- （bootstrap/config_seed.go），本脚本为手工部署等价存档。
CREATE TABLE IF NOT EXISTS system_configs (
    "key" VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    description VARCHAR(255) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE system_configs IS '通用系统配置键值表（运行时配置唯一主源）';
