package res

import "time"

// ModuleConfigRes 模块开关配置响应结构体。
type ModuleConfigRes struct {
	ArticleEnabled    bool      `json:"article_enabled"`
	MediaEnabled      bool      `json:"media_enabled"`
	MusicEnabled      bool      `json:"music_enabled"`
	VideoEnabled      bool      `json:"video_enabled"`
	TravelEnabled     bool      `json:"travel_enabled"`
	PortfolioEnabled  bool      `json:"portfolio_enabled"`
	EquipmentEnabled  bool      `json:"equipment_enabled"`
	ProjectEnabled    bool      `json:"project_enabled"`
	OpenSourceEnabled bool      `json:"open_source_enabled"`
	UpdatedAt         time.Time `json:"updated_at"`
}
