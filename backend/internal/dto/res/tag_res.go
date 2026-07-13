package res

import "time"

// TagRes 标签响应结构体。
type TagRes struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
