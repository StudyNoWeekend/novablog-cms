package logic

import "go.uber.org/zap"

// ThemeSettings 主题模块配置，由 bootstrap 阶段从 config.yaml 注入。
type ThemeSettings struct {
	DataDir       string // 主题制品解压根目录
	FrontendDir   string // 自备博客前端目录（含 theme.json 与 dist/）；非空时优先于已安装主题
	MarketBaseURL string // 官方主题市场地址（服务端直连；空=禁用首装拉取）
	GithubToken   string // 可选 GitHub PAT（提升 Release 查询限流额度）
	PublicAPIBase string // 跨域部署时注入 theme-config.js 的 apiBase；空=同域相对路径
	MaxArtifactMB int    // 制品下载上限（MB）
}

// themeSettings 主题模块运行时配置（默认值兜底，bootstrap 启动时覆盖）。
var themeSettings = &ThemeSettings{DataDir: "./data/themes", MaxArtifactMB: 100}

// SetThemeSettings 注入主题模块配置。
func SetThemeSettings(s *ThemeSettings) {
	if s != nil {
		themeSettings = s
	}
}

// SetThemeMarketBaseURL 运行时覆盖官方市场地址（DB 持久化值启动注入 / 修改后热更新）。
func SetThemeMarketBaseURL(base string) {
	if base != "" {
		themeSettings.MarketBaseURL = base
	}
}

// SetThemePublicAPIBase 运行时覆盖博客公开 API 地址（DB 持久化值启动注入；空值表示同域相对路径取数）。
func SetThemePublicAPIBase(base string) {
	if base != "" {
		themeSettings.PublicAPIBase = base
	}
}

// getThemeSettings 读取主题模块配置。
func getThemeSettings() *ThemeSettings {
	return themeSettings
}

// ThemeLogger 主题模块日志记录器（bootstrap 注入，未注入时静默）。
var ThemeLogger *zap.Logger

// themeLog 获取主题模块日志记录器。
func themeLog() *zap.Logger {
	if ThemeLogger != nil {
		return ThemeLogger
	}
	return zap.NewNop()
}
