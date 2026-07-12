package req

// ContentTrendReq 内容产出趋势请求参数。
type ContentTrendReq struct {
	Range string `form:"range" json:"range" binding:"omitempty,oneof=7d 30d 90d"`
}

// GetRange 获取时间范围，默认返回 30d。
func (r *ContentTrendReq) GetRange() string {
	if r.Range == "" {
		return "30d"
	}
	return r.Range
}

// TopContentReq 热门内容排行请求参数。
type TopContentReq struct {
	Type  string `form:"type" json:"type" binding:"omitempty,oneof=article travel"`
	Sort  string `form:"sort" json:"sort" binding:"omitempty,oneof=views comments"`
	Limit int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=20"`
}

// GetType 获取类型，默认返回 article。
func (r *TopContentReq) GetType() string {
	if r.Type == "" {
		return "article"
	}
	return r.Type
}

// GetSort 获取排序方式，默认返回 views。
func (r *TopContentReq) GetSort() string {
	if r.Sort == "" {
		return "views"
	}
	return r.Sort
}

// GetLimit 获取数量限制，默认返回 5。
func (r *TopContentReq) GetLimit() int {
	if r.Limit <= 0 {
		return 5
	}
	if r.Limit > 20 {
		return 20
	}
	return r.Limit
}

// RecentCommentsReq 最近评论请求参数。
type RecentCommentsReq struct {
	Limit int `form:"limit" json:"limit" binding:"omitempty,min=1,max=20"`
}

// GetLimit 获取数量限制，默认返回 5。
func (r *RecentCommentsReq) GetLimit() int {
	if r.Limit <= 0 {
		return 5
	}
	if r.Limit > 20 {
		return 20
	}
	return r.Limit
}
