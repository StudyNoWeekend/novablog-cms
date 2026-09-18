package req

// CreatePlaylistReq 创建第三方歌单请求参数。
type CreatePlaylistReq struct {
	Title       string `json:"title" binding:"required,min=1,max=500"`
	CoverURL    string `json:"cover_url"`
	Platform    string `json:"platform" binding:"required"`
	PlatformURL string `json:"platform_url" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Enabled     *bool  `json:"enabled"`
}

// UpdatePlaylistReq 更新第三方歌单请求参数。
type UpdatePlaylistReq struct {
	Title       *string `json:"title"`
	CoverURL    *string `json:"cover_url"`
	Platform    *string `json:"platform"`
	PlatformURL *string `json:"platform_url"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
	Enabled     *bool   `json:"enabled"`
}

// PlaylistListReq 第三方歌单列表请求参数。
type PlaylistListReq struct {
	PageReq
}
