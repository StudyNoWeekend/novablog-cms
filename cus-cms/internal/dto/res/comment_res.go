package res

import "time"

// CommentRes 评论响应结构体。
type CommentRes struct {
	ID          string    `json:"id"`           // 评论 ID
	TargetType  string    `json:"target_type"`  // 目标类型
	TargetID    string    `json:"target_id"`    // 目标 ID
	TargetTitle string    `json:"target_title"` // 关联目标标题
	ParentID    *string   `json:"parent_id"`    // 父评论 ID
	BloggerID   *string   `json:"blogger_id"`   // 博主 ID
	Nickname    string    `json:"nickname"`     // 评论者名称
	Website     string    `json:"website"`      // 评论者博客地址
	Content     string    `json:"content"`      // 评论内容
	IsBlogger   bool      `json:"is_blogger"`   // 是否为博主回复
	IPAddress   string    `json:"ip_address"`   // IP 地址
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}

// CommentPublicRes 公开评论响应结构体（排除 IP 地址和博主 ID 等敏感字段）。
type CommentPublicRes struct {
	ID         string    `json:"id"`          // 评论 ID
	TargetType string    `json:"target_type"` // 目标类型
	TargetID   string    `json:"target_id"`   // 目标 ID
	ParentID   *string   `json:"parent_id"`   // 父评论 ID
	Nickname   string    `json:"nickname"`    // 评论者名称
	Website    string    `json:"website"`     // 评论者博客地址
	Content    string    `json:"content"`     // 评论内容
	IsBlogger  bool      `json:"is_blogger"`  // 是否为博主回复
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
}
