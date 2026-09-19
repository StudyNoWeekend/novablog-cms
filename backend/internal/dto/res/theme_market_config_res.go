package res

import "time"

// ThemeMarketConfigRes 官方主题市场配置响应。
type ThemeMarketConfigRes struct {
	MarketBaseURL string    `json:"market_base_url"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PublicConfigRes 公共配置响应（免鉴权下发，默认值由后端控制）。
type PublicConfigRes struct {
	MarketBaseURL string `json:"market_base_url"`
}
