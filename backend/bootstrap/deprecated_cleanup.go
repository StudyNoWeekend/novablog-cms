package bootstrap

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// deprecatedCleanups 存量部署的孤儿库对象清理清单（幂等，全部 IF EXISTS）。
// 背景：AutoMigrate 只增不删，历史上模型删字段/删模型后列与表会残留在库中；
// 此处统一在启动时清理，失败仅记日志不阻塞启动。
var deprecatedCleanups = []struct {
	name string // 日志用名称
	sql  string
}{
	// security_configs：015 迁移建表时的旧限流列，旧模型（git 9dee501）删除字段后残留，全仓零读写
	{"security_configs.get_max_tokens", `ALTER TABLE security_configs DROP COLUMN IF EXISTS get_max_tokens`},
	{"security_configs.get_window_seconds", `ALTER TABLE security_configs DROP COLUMN IF EXISTS get_window_seconds`},
	{"security_configs.post_max_tokens", `ALTER TABLE security_configs DROP COLUMN IF EXISTS post_max_tokens`},
	{"security_configs.post_window_seconds", `ALTER TABLE security_configs DROP COLUMN IF EXISTS post_window_seconds`},
	{"security_configs.view_max_tokens", `ALTER TABLE security_configs DROP COLUMN IF EXISTS view_max_tokens`},
	{"security_configs.view_window_seconds", `ALTER TABLE security_configs DROP COLUMN IF EXISTS view_window_seconds`},
	{"security_configs.like_max_tokens", `ALTER TABLE security_configs DROP COLUMN IF EXISTS like_max_tokens`},
	{"security_configs.like_window_seconds", `ALTER TABLE security_configs DROP COLUMN IF EXISTS like_window_seconds`},
	{"security_configs.blacklist_threshold", `ALTER TABLE security_configs DROP COLUMN IF EXISTS blacklist_threshold`},
	// theme_market_configs：官方站点地址已统一托管到 community_push_configs，列已废弃（模型字段已删）
	{"theme_market_configs.market_base_url", `ALTER TABLE theme_market_configs DROP COLUMN IF EXISTS market_base_url`},
	// ip_blacklist：旧版黑名单表（单数），现运行时黑名单走 ip_blacklists 表 + Redis，模型已删除
	{"ip_blacklist (table)", `DROP TABLE IF EXISTS ip_blacklist`},
}

// cleanupDeprecatedColumns 启动时幂等清理废弃的列与表。
func cleanupDeprecatedColumns(db *gorm.DB, logger *zap.Logger) {
	for _, c := range deprecatedCleanups {
		if err := db.Exec(c.sql).Error; err != nil {
			logger.Warn("清理废弃库对象失败（不影响启动）", zap.String("target", c.name), zap.Error(err))
			continue
		}
		logger.Info("废弃库对象清理完成（若不存在则为幂等空操作）", zap.String("target", c.name))
	}
}
