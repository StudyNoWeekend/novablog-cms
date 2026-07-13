package req

// PlatformLinkReq 平台链接请求参数。
type PlatformLinkReq struct {
	Platform string `json:"platform" binding:"required"`
	URL      string `json:"url" binding:"required"`
}

// CreateVideoReq 创建视频作品请求参数。
type CreateVideoReq struct {
	Title       string            `json:"title" binding:"required,min=1,max=255"`
	CoverURL    string            `json:"cover_url"`
	Description string            `json:"description"`
	Status      *int              `json:"status"`
	Platforms   []PlatformLinkReq `json:"platforms" binding:"required,min=1,dive"`
}

// UpdateVideoReq 更新视频作品请求参数。
type UpdateVideoReq struct {
	Title       *string           `json:"title" binding:"omitempty,min=1,max=255"`
	CoverURL    *string           `json:"cover_url"`
	Description *string           `json:"description"`
	Status      *int              `json:"status"`
	Platforms   []PlatformLinkReq `json:"platforms" binding:"omitempty,dive"`
}

// VideoListReq 视频作品列表查询请求参数。
type VideoListReq struct {
	PageReq
	Keyword *string `form:"keyword" json:"keyword"`
	Status  *int    `form:"status" json:"status"`
}

// ParseVideoReq 视频元信息解析请求参数。
type ParseVideoReq struct {
	Platform string `json:"platform" binding:"required"`
	URL      string `json:"url" binding:"required"`
}
