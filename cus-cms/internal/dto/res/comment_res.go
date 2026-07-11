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
	Nickname    string    `json:"nickname"`     // 评论者昵称
	Email       string    `json:"email"`        // 评论者邮箱
	Avatar      string    `json:"avatar"`       // 评论者头像
	Content     string    `json:"content"`      // 评论内容
	IsBlogger   bool      `json:"is_blogger"`   // 是否为博主回复
	Status      int16     `json:"status"`       // 评论状态
	IPAddress   string    `json:"ip_address"`   // IP 地址
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}
