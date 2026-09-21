package req

// UpdateModuleConfigReq 更新模块开关配置请求参数。
type UpdateModuleConfigReq struct {
	ArticleEnabled   *bool `json:"article_enabled"`
	MediaEnabled     *bool `json:"media_enabled"`
	MusicEnabled     *bool `json:"music_enabled"`
	VideoEnabled     *bool `json:"video_enabled"`
	TravelEnabled    *bool `json:"travel_enabled"`
	PortfolioEnabled *bool `json:"portfolio_enabled"`
	EquipmentEnabled *bool `json:"equipment_enabled"`
	ProjectEnabled   *bool `json:"project_enabled"`
}
