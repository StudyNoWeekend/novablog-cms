package req

// CreateTravelGuideReq 创建旅行攻略请求参数。
type CreateTravelGuideReq struct {
	Title       string           `json:"title" binding:"required,min=1,max=200"`
	Summary     string           `json:"summary"`
	CoverImage  string           `json:"cover_image"`
	Status      int16            `json:"status"`
	Destination string           `json:"destination" binding:"required,min=1,max=200"`
	Region      string           `json:"region" binding:"required"`
	CategoryID  *string          `json:"category_id"`
	Days        int              `json:"days"`
	BestMonth   string           `json:"best_month"`
	Attractions []map[string]any `json:"attractions"`
	Itinerary   []map[string]any `json:"itinerary"`
	Reviews     []map[string]any `json:"reviews"`
}

// UpdateTravelGuideReq 更新旅行攻略请求参数（所有字段为指针类型）。
type UpdateTravelGuideReq struct {
	Title       *string          `json:"title"`
	Summary     *string          `json:"summary"`
	CoverImage  *string          `json:"cover_image"`
	Status      *int16           `json:"status"`
	Destination *string          `json:"destination"`
	Region      *string          `json:"region"`
	CategoryID  *string          `json:"category_id"`
	Days        *int             `json:"days"`
	BestMonth   *string          `json:"best_month"`
	Attractions []map[string]any `json:"attractions"`
	Itinerary   []map[string]any `json:"itinerary"`
	Reviews     []map[string]any `json:"reviews"`
}

// TravelGuideListReq 旅行攻略列表查询请求参数。
type TravelGuideListReq struct {
	PageReq
	Keyword    *string `form:"keyword" json:"keyword"`
	Region     *string `form:"region" json:"region"`
	CategoryID *string `form:"category_id" json:"category_id"`
	Status     *int16  `form:"status" json:"status"`
	DaysRange  *string `form:"days_range" json:"days_range"`
	Sort       *string `form:"sort" json:"sort"`
}

// UpdateTravelGuideStatusReq 更新旅行攻略状态请求参数。
type UpdateTravelGuideStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=1 2 3"`
}
