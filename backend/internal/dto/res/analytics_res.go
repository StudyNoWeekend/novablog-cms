package res

import "time"

// OverviewRes 工作台概览统计响应结构体。
type OverviewRes struct {
	// 内容统计
	ArticleTotal       int64 `json:"article_total"`
	ArticlePublished   int64 `json:"article_published"`
	ArticleDraft       int64 `json:"article_draft"`
	PortfolioTotal     int64 `json:"portfolio_total"`
	PortfolioPublished int64 `json:"portfolio_published"`
	VideoTotal         int64 `json:"video_total"`
	VideoPublished     int64 `json:"video_published"`
	TravelTotal        int64 `json:"travel_total"`
	TravelPublished    int64 `json:"travel_published"`
	SongTotal          int64 `json:"song_total"`

	// 评论统计
	CommentTotal int64 `json:"comment_total"`

	// 攻略浏览量
	TravelViews int64 `json:"travel_views"`

	// 发布节奏
	LastPublishAt        *time.Time `json:"last_publish_at"`
	DaysSinceLastPublish int        `json:"days_since_last_publish"`
}

// ContentTrendItem 内容产出趋势单日数据。
type ContentTrendItem struct {
	Date         string `json:"date"`
	ArticleCount int64  `json:"article_count"`
	TravelCount  int64  `json:"travel_count"`
}

// ContentTrendRes 内容产出趋势响应结构体。
type ContentTrendRes struct {
	Range string             `json:"range"`
	Items []ContentTrendItem `json:"items"`
}

// TopContentItem 热门内容排行项。
type TopContentItem struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	ViewCount    int64      `json:"view_count"`
	CommentCount int64      `json:"comment_count"`
	Slug         string     `json:"slug,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
}

// TopContentRes 热门内容排行响应结构体。
type TopContentRes struct {
	Type  string           `json:"type"`
	Sort  string           `json:"sort"`
	Items []TopContentItem `json:"items"`
}

// DistributionItem 内容类型分布项。
type DistributionItem struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// DistributionRes 内容类型分布响应结构体。
type DistributionRes struct {
	Items []DistributionItem `json:"items"`
	Total int64              `json:"total"`
}

// RecentCommentItem 最近评论项。
type RecentCommentItem struct {
	ID          string    `json:"id"`
	Nickname    string    `json:"nickname"`
	Content     string    `json:"content"`
	TargetType  string    `json:"target_type"`
	TargetID    string    `json:"target_id"`
	TargetTitle string    `json:"target_title"`
	IsBlogger   bool      `json:"is_blogger"`
	CreatedAt   time.Time `json:"created_at"`
}

// RecentCommentsRes 最近评论响应结构体。
type RecentCommentsRes struct {
	Items []RecentCommentItem `json:"items"`
}
