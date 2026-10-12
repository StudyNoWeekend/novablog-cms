package logic

import "go.uber.org/zap"

// CommunityPushLogger 推送模块日志记录器（bootstrap 注入，未注入时静默）。
var CommunityPushLogger *zap.Logger

// communityPushLog 获取推送模块日志记录器。
func communityPushLog() *zap.Logger {
	if CommunityPushLogger != nil {
		return CommunityPushLogger
	}
	return zap.NewNop()
}

// CommunityPushSettings CMS 推送到 NovaBlog 社区的模块配置。
// SiteURL 出厂值来自 config.yaml `community:` 段，DB 单行表（community_push_configs）启动注入并可后台热更。
// 官方站点地址统一由 official_site_settings 托管（与主题市场共用），本结构不再持有。
type CommunityPushSettings struct {
	SiteURL string // 博客对外地址（用于拼接推送 source_url）
}

// communityPushSettings 推送模块运行时配置（空值兜底，bootstrap 启动时覆盖）。
var communityPushSettings = &CommunityPushSettings{}

// SetCommunityPushSettings 注入推送模块配置。
func SetCommunityPushSettings(s *CommunityPushSettings) {
	if s != nil {
		communityPushSettings = s
	}
}

// SetCommunitySiteURL DB 持久化值启动注入 / 修改后热更新。
func SetCommunitySiteURL(site string) {
	if site != "" {
		communityPushSettings.SiteURL = site
	}
}

// getCommunityPushSettings 读取推送模块配置。
func getCommunityPushSettings() *CommunityPushSettings {
	return communityPushSettings
}

// getCommunityPushKey 读取推送密钥（DB 单行表持久化值）。
func getCommunityPushKey() string {
	return communityPushKey
}

// communityPushKey 推送密钥运行时值（bootstrap 从 DB 启动注入，修改后热更新）。
var communityPushKey string

// SetCommunityPushKey 设置推送密钥运行时值。
func SetCommunityPushKey(key string) {
	communityPushKey = key
}
