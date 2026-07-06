package req

// PageReq 分页请求参数。
type PageReq struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetPage 获取页码，默认返回 1。
func (p *PageReq) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetPageSize 获取每页条数，默认返回 20。
func (p *PageReq) GetPageSize() int {
	if p.PageSize <= 0 {
		return 20
	}
	return p.PageSize
}