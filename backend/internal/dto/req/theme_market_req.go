package req

// MarketLoginReq 登录官方主题市场请求参数（官方地址通过 X-Market-Base-URL 请求头传递）。
type MarketLoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ThemeMarketListReq 官方主题市场列表查询参数。
type ThemeMarketListReq struct {
	PageReq
	Type     string   `form:"type" json:"type"`         // 主题类型，空为全部
	Styles   []string `form:"styles" json:"styles"`     // 风格筛选，多值取交集
	Features []string `form:"features" json:"features"` // 功能筛选，多值取交集
	Price    string   `form:"price" json:"price"`       // free / paid，空为全部
	Search   string   `form:"search" json:"search"`     // 关键词，模糊匹配标题和描述
	Sort     string   `form:"sort" json:"sort"`         // latest / downloads / rating
}

// ThemeMarketPageReq 我的收藏列表分页参数。
type ThemeMarketPageReq struct {
	PageReq
}

// ThemeMarketRatingReq 主题评分请求参数。
type ThemeMarketRatingReq struct {
	Score int `json:"score" binding:"required,min=1,max=5"`
}
