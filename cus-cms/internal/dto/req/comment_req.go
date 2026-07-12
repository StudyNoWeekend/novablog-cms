package req

// CommentListReq 评论列表查询请求参数。
type CommentListReq struct {
	PageReq
	TargetType *string `form:"target_type" json:"target_type"` // 目标类型 article/travel_guide
	TargetID   *string `form:"target_id" json:"target_id"`     // 目标 ID
	Keyword    *string `form:"keyword" json:"keyword"`         // 模糊搜索 content
}

// ReplyCommentReq 博主回复评论请求参数。
type ReplyCommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=2000"` // 回复内容
}

// CreatePublicCommentReq 公开评论创建请求参数。
type CreatePublicCommentReq struct {
	TargetType string  `json:"target_type" binding:"required,oneof=article travel_guide"` // 目标类型
	TargetID   string  `json:"target_id" binding:"required"`                              // 目标 ID
	ParentID   *string `json:"parent_id"`                                                 // 父评论 ID
	Nickname   string  `json:"nickname" binding:"required,min=1,max=50"`                  // 评论者名称
	Website    string  `json:"website"`                                                   // 评论者博客地址
	Content    string  `json:"content" binding:"required,min=1,max=2000"`                 // 评论内容
}
