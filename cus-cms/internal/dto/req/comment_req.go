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
