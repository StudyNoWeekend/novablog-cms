package res

import "time"

// ThemeMarketConfigRes 主题模块运行配置响应。
type ThemeMarketConfigRes struct {
	MarketBaseURL string    `json:"market_base_url"`
	PublicAPIBase string    `json:"public_api_base"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PublicConfigRes 公共配置响应（免鉴权下发，默认值由后端控制）。
type PublicConfigRes struct {
	MarketBaseURL string `json:"market_base_url"`
}
