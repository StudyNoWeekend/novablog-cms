package req

// UpdateThemeMarketConfigReq 更新主题模块运行配置请求（均为可选，仅更新提供的字段）。
type UpdateThemeMarketConfigReq struct {
	MarketBaseURL *string `json:"market_base_url"`
	PublicAPIBase *string `json:"public_api_base"`
}
