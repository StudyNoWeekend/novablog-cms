package res

import "time"

// ThemeMarketConfigRes 主题模块运行配置响应。
// 官方站点地址统一由社区推送配置托管（GET /public/config 下发），此处不再返回。
type ThemeMarketConfigRes struct {
	PublicAPIBase string    `json:"public_api_base"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PublicConfigRes 公共配置响应（免鉴权下发，默认值由后端控制）。
// MarketBaseURL 字段名沿用历史契约，语义为统一托管的 NovaBlog 官方站点地址。
type PublicConfigRes struct {
	MarketBaseURL string `json:"market_base_url"`
}
