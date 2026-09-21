package req

// CreateOpenSourceReq 创建开源作品请求参数。
type CreateOpenSourceReq struct {
	Name      string `json:"name" binding:"required,min=1,max=255"`
	RepoURL   string `json:"repo_url" binding:"required,max=1024"`
	Summary   string `json:"summary" binding:"omitempty,max=500"`
	Language  string `json:"language" binding:"omitempty,max=100"`
	Topics    string `json:"topics" binding:"omitempty,max=500"` // 主题标签，逗号分隔
	Homepage  string `json:"homepage" binding:"omitempty,max=1024"`
	Status    *int   `json:"status" binding:"omitempty,oneof=0 1"` // 0=草稿, 1=已发布
	SortOrder *int   `json:"sort_order"`
	// FetchRemote 是否拉取远端仓库元数据与 README（默认 true；导入测试数据时可关闭）
	FetchRemote *bool `json:"fetch_remote"`
}

// UpdateOpenSourceReq 更新开源作品请求参数（部分更新语义）。
type UpdateOpenSourceReq struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=255"`
	RepoURL   *string `json:"repo_url" binding:"omitempty,max=1024"`
	Summary   *string `json:"summary" binding:"omitempty,max=500"`
	Language  *string `json:"language" binding:"omitempty,max=100"`
	Topics    *string `json:"topics" binding:"omitempty,max=500"`
	Homepage  *string `json:"homepage" binding:"omitempty,max=1024"`
	Status    *int    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder *int    `json:"sort_order"`
}

// OpenSourceListReq 开源作品列表查询请求参数。
type OpenSourceListReq struct {
	PageReq
	Keyword *string `form:"keyword" json:"keyword"`
	Status  *int    `form:"status" json:"status"`
}
