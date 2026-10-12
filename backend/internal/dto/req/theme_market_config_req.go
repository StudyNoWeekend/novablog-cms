package req

// UpdateThemeMarketConfigReq 更新主题模块运行配置请求（可选，仅更新提供的字段）。
// 官方站点地址统一在社区推送配置页维护，本请求不再受理。
type UpdateThemeMarketConfigReq struct {
	PublicAPIBase *string `json:"public_api_base"`
}
