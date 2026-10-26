package req

// CreateArticleReq 创建文章请求参数。
type CreateArticleReq struct {
	Title      string         `json:"title" binding:"required,min=1,max=200"`
	Content    string         `json:"content" binding:"required"`
	Summary    string         `json:"summary"`
	CoverImage string         `json:"cover_image"`
	// 封面媒体引用：与 cover_image 互斥，非空时封面按媒体库文件 ID 索引，读取时动态解析链接
	CoverMediaID  string         `json:"cover_media_id"`
	CoverPresetID string         `json:"cover_preset_id"`
	CategoryID    *string        `json:"category_id"`
	TagIDs        []string       `json:"tag_ids"`
	Status        int16          `json:"status"`
	Type          int16          `json:"type"`
	IsTop         *bool          `json:"is_top"`
	IsComment     *bool          `json:"is_comment"`
	Extra         map[string]any `json:"extra"`
}

// UpdateArticleReq 更新文章请求参数。
type UpdateArticleReq struct {
	Title      *string        `json:"title"`
	Content    *string        `json:"content"`
	Summary    *string        `json:"summary"`
	CoverImage *string        `json:"cover_image"`
	// 封面媒体引用：非 nil 时与 cover_image 互斥处理
	CoverMediaID  *string        `json:"cover_media_id"`
	CoverPresetID *string        `json:"cover_preset_id"`
	CategoryID    *string        `json:"category_id"`
	TagIDs        []string       `json:"tag_ids"`
	Status        *int16         `json:"status"`
	Type          *int16         `json:"type"`
	IsTop         *bool          `json:"is_top"`
	IsComment     *bool          `json:"is_comment"`
	Extra         map[string]any `json:"extra"`
}

// ArticleListReq 文章列表查询请求参数。
type ArticleListReq struct {
	PageReq
	Status     *int16  `form:"status" json:"status"`
	CategoryID *string `form:"category_id" json:"category_id"`
	Keyword    *string `form:"keyword" json:"keyword"`
}

// UpdateStatusReq 更新文章状态请求参数。
type UpdateStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=1 2 3"`
}

// HotArticleReq 热门文章请求参数。
type HotArticleReq struct {
	Count int `form:"count" json:"count"`
}

// RandomArticleReq 随机文章请求参数。
type RandomArticleReq struct {
	Count int `form:"count" json:"count"`
}
