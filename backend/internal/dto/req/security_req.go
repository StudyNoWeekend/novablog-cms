package req

// UpdateSecurityConfigReq 更新安全配置请求参数。
type UpdateSecurityConfigReq struct {
	GetMaxTokens        *int `json:"get_max_tokens" binding:"omitempty,min=1"`
	GetWindowSeconds    *int `json:"get_window_seconds" binding:"omitempty,min=1"`
	PostMaxTokens       *int `json:"post_max_tokens" binding:"omitempty,min=1"`
	PostWindowSeconds   *int `json:"post_window_seconds" binding:"omitempty,min=1"`
	ViewMaxTokens       *int `json:"view_max_tokens" binding:"omitempty,min=1"`
	ViewWindowSeconds   *int `json:"view_window_seconds" binding:"omitempty,min=1"`
	LikeMaxTokens       *int `json:"like_max_tokens" binding:"omitempty,min=1"`
	LikeWindowSeconds   *int `json:"like_window_seconds" binding:"omitempty,min=1"`
	BlacklistThreshold  *int `json:"blacklist_threshold" binding:"omitempty,min=1"`
	BlacklistTTLMinutes *int `json:"blacklist_ttl_minutes" binding:"omitempty,min=1"`
}

// BlacklistQueryReq 黑名单列表查询请求参数。
type BlacklistQueryReq struct {
	PageReq
}
