package res

import "time"

// ArticleRes 文章列表响应结构体。
type ArticleRes struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Summary      string     `json:"summary"`
	CoverImage   string     `json:"cover_image"`
	CategoryID   string     `json:"category_id"`
	CategoryName string     `json:"category_name"`
	TagIDs       []string   `json:"tag_ids"`
	TagNames     []string   `json:"tag_names"`
	Status       int16      `json:"status"`
	Type         int16      `json:"type"`
	ViewCount    int        `json:"view_count"`
	CommentCount int        `json:"comment_count"`
	IsTop        bool       `json:"is_top"`
	IsComment    bool       `json:"is_comment"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ArticleDetailRes 文章详情响应结构体。
type ArticleDetailRes struct {
	ArticleRes
	Content string         `json:"content"`
	Extra   map[string]any `json:"extra"`
}
