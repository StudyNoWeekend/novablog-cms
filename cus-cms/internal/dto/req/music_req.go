package req

// CreateSongReq 创建歌曲请求参数。
type CreateSongReq struct {
	Title      string  `json:"title" binding:"required,min=1,max=500"`
	Artist     string  `json:"artist" binding:"required,min=1,max=255"`
	CoverURL   string  `json:"cover_url"`
	BVID       string  `json:"bvid" binding:"required"`
	CID        int64   `json:"cid" binding:"required"`
	SourceURL  string  `json:"source_url" binding:"required"`
	SourceType string  `json:"source_type"`
	CategoryID *string `json:"category_id"`
	Duration   int     `json:"duration"`
	SortOrder  int     `json:"sort_order"`
}

// UpdateSongReq 更新歌曲请求参数。
type UpdateSongReq struct {
	Title      *string `json:"title"`
	Artist     *string `json:"artist"`
	CoverURL   *string `json:"cover_url"`
	CategoryID *string `json:"category_id"`
	Duration   *int    `json:"duration"`
	SortOrder  *int    `json:"sort_order"`
}

// SongListReq 歌曲列表请求参数。
type SongListReq struct {
	PageReq
	CategoryID string `form:"category_id" json:"category_id"`
}

// ParseMusicReq B站链接解析请求参数。
type ParseMusicReq struct {
	URL string `json:"url" binding:"required"`
}

// BatchCreateSongReq 批量创建歌曲请求参数。
type BatchCreateSongReq struct {
	Songs []CreateSongReq `json:"songs" binding:"required,min=1,dive"`
}
