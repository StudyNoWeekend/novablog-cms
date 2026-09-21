package req

// CreateProjectReq 创建项目经历请求参数。
type CreateProjectReq struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Category    string `json:"category" binding:"omitempty,max=100"`
	Role        string `json:"role" binding:"omitempty,max=100"`
	Client      string `json:"client" binding:"omitempty,max=255"`
	CoverURL    string `json:"cover_url" binding:"omitempty,max=1024"`
	Summary     string `json:"summary" binding:"omitempty,max=500"`
	Description string `json:"description"`
	TechStack   string `json:"tech_stack" binding:"omitempty,max=500"` // 技能/工具标签，逗号分隔
	StartDate   string `json:"start_date" binding:"omitempty,len=7"`  // 按月粒度：YYYY-MM
	EndDate     string `json:"end_date" binding:"omitempty,len=7"`    // 按月粒度：YYYY-MM，空表示至今
	ProjectURL  string `json:"project_url" binding:"omitempty,max=1024"`
	RepoURL     string `json:"repo_url" binding:"omitempty,max=1024"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"` // 0=草稿, 1=已发布
	SortOrder   *int   `json:"sort_order"`
}

// UpdateProjectReq 更新项目经历请求参数。
type UpdateProjectReq struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=255"`
	Category    *string `json:"category" binding:"omitempty,max=100"`
	Role        *string `json:"role" binding:"omitempty,max=100"`
	Client      *string `json:"client" binding:"omitempty,max=255"`
	CoverURL    *string `json:"cover_url" binding:"omitempty,max=1024"`
	Summary     *string `json:"summary" binding:"omitempty,max=500"`
	Description *string `json:"description"`
	TechStack   *string `json:"tech_stack" binding:"omitempty,max=500"`
	// 日期字段不绑定 len 校验：空字符串（清空）合法，格式由 logic 层 parseMonthPtr 校验
	StartDate  *string `json:"start_date"` // 空字符串表示清空，格式 YYYY-MM
	EndDate    *string `json:"end_date"`   // 空字符串表示清空（至今），格式 YYYY-MM
	ProjectURL *string `json:"project_url" binding:"omitempty,max=1024"`
	RepoURL    *string `json:"repo_url" binding:"omitempty,max=1024"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int    `json:"sort_order"`
}

// ProjectListReq 项目经历列表查询请求参数。
type ProjectListReq struct {
	PageReq
	Keyword  *string `form:"keyword" json:"keyword"`
	Category *string `form:"category" json:"category"`
	Status   *int    `form:"status" json:"status"`
}
