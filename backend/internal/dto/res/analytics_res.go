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

	// 扩展模块统计
	EquipmentTotal      int64 `json:"equipment_total"`
	ProjectTotal        int64 `json:"project_total"`
	ProjectPublished    int64 `json:"project_published"`
	OpenSourceTotal     int64 `json:"open_source_total"`
	OpenSourcePublished int64 `json:"open_source_published"`
	RecipeTotal         int64 `json:"recipe_total"`
	RecipePublished     int64 `json:"recipe_published"`
	BookTotal           int64 `json:"book_total"`
	BookPublished       int64 `json:"book_published"`
	GameTotal           int64 `json:"game_total"`
	GamePublished       int64 `json:"game_published"`
	FitnessTotal        int64 `json:"fitness_total"`
	FitnessPublished    int64 `json:"fitness_published"`
	TechStackTotal      int64 `json:"tech_stack_total"`
	TechStackPublished  int64 `json:"tech_stack_published"`

	// 评论统计
	CommentTotal int64 `json:"comment_total"`

	// 攻略浏览量
	TravelViews int64 `json:"travel_views"`

	// 发布节奏
	LastPublishAt        *time.Time `json:"last_publish_at"`
	DaysSinceLastPublish int        `json:"days_since_last_publish"`
}

// ContentTrendItem 内容产出趋势单日数据，Counts 按内容类型（article/travel/portfolio/...）统计当日新增数。
type ContentTrendItem struct {
	Date   string           `json:"date"`
	Counts map[string]int64 `json:"counts"`
}

// ContentTrendRes 内容产出趋势响应结构体。
type ContentTrendRes struct {
	Range string             `json:"range"`
	Items []ContentTrendItem `json:"items"`
}

// TopContentItem 热门内容排行项。
type TopContentItem struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"` // 内容类型：article/travel
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

// TrafficSummaryRes 访问概览响应结构体。
type TrafficSummaryRes struct {
	// 当日访问（实时来自明细表）
	TodayPV int64 `json:"today_pv"`
	TodayUV int64 `json:"today_uv"`
	// 昨日访问（来自按日聚合表）
	YesterdayPV int64 `json:"yesterday_pv"`
	// 全部内容表浏览量累计值（含接入统计体系前的历史数据）
	TotalPV int64 `json:"total_pv"`
}

// TrafficTrendItem 访问趋势单日数据，Counts 按内容类型统计当日 PV。
type TrafficTrendItem struct {
	Date    string           `json:"date"`
	Counts  map[string]int64 `json:"counts"`
	TotalPV int64            `json:"total_pv"`
}

// TrafficTrendRes 访问趋势响应结构体。
type TrafficTrendRes struct {
	Range string             `json:"range"`
	Items []TrafficTrendItem `json:"items"`
}

// ModuleStatRow 模块数据表行：发布与访问量汇总。
type ModuleStatRow struct {
	Type       string `json:"type"`
	Total      int64  `json:"total"`
	Published  int64  `json:"published"`
	TotalViews int64  `json:"total_views"`
	TodayViews int64  `json:"today_views"`
	WeekViews  int64  `json:"week_views"`
}

// ModuleStatsRes 模块数据表响应结构体。
type ModuleStatsRes struct {
	Items []ModuleStatRow `json:"items"`
}
