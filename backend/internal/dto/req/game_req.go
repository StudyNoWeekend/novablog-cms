package req

// CreateGameReq 创建游戏请求参数。
type CreateGameReq struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Cover       string `json:"cover" binding:"omitempty,max=500"`
	Platform    string `json:"platform" binding:"omitempty,max=100"`
	Genre       string `json:"genre" binding:"omitempty,max=100"`
	PlayStatus  string `json:"play_status" binding:"omitempty,oneof=want playing played"`
	PlayHours   *int   `json:"play_hours" binding:"omitempty,min=0"`
	Rating      *int   `json:"rating" binding:"omitempty,min=0,max=10"`
	ShortReview string `json:"short_review" binding:"omitempty,max=1000"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int   `json:"sort_order"`
}

// UpdateGameReq 更新游戏请求参数（部分更新语义）。
type UpdateGameReq struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=255"`
	Cover       *string `json:"cover" binding:"omitempty,max=500"`
	Platform    *string `json:"platform" binding:"omitempty,max=100"`
	Genre       *string `json:"genre" binding:"omitempty,max=100"`
	PlayStatus  *string `json:"play_status" binding:"omitempty,oneof=want playing played"`
	PlayHours   *int    `json:"play_hours" binding:"omitempty,min=0"`
	Rating      *int    `json:"rating" binding:"omitempty,min=0,max=10"`
	ShortReview *string `json:"short_review" binding:"omitempty,max=1000"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int    `json:"sort_order"`
}

// GameListReq 游戏列表查询请求参数。
type GameListReq struct {
	PageReq
	Keyword    *string `form:"keyword" json:"keyword"`
	Platform   *string `form:"platform" json:"platform"`
	Genre      *string `form:"genre" json:"genre"`
	PlayStatus *string `form:"play_status" json:"play_status"`
	Status     *int    `form:"status" json:"status"`
}
