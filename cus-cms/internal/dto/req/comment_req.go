package req

// CommentListReq 评论列表查询请求参数。
type CommentListReq struct {
	PageReq
	TargetType *string `form:"target_type" json:"target_type"` // 目标类型 article/travel_guide
	TargetID   *string `form:"target_id" json:"target_id"`     // 目标 ID
	Status     *int16  `form:"status" json:"status"`           // 评论状态 1=待审核 2=已通过 3=已拒绝
	Keyword    *string `form:"keyword" json:"keyword"`         // 模糊搜索 content
}

// ReplyCommentReq 博主回复评论请求参数。
type ReplyCommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=2000"` // 回复内容
}

// UpdateCommentStatusReq 更新评论状态请求参数。
type UpdateCommentStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=2 3"` // 状态 2=已通过 3=已拒绝
}
