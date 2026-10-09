package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// TrendRow 内容产出趋势单日数据行，Counts 按内容类型（article/travel/...）统计当日新增数。
type TrendRow struct {
	Date   time.Time
	Counts map[string]int64
}

// trendSource 趋势统计的内容表配置，key 与前端内容类型标识一致。
type trendSource struct {
	key   string
	table string
	// where 附加过滤条件；songs 表无软删除字段，不传 where
	where string
}

// trendSources 参与内容产出趋势统计的模块表，新增模块时在此维护。
var trendSources = []trendSource{
	{key: "article", table: "articles", where: "deleted_at IS NULL"},
	{key: "travel", table: "travel_guides", where: "deleted_at IS NULL"},
	{key: "portfolio", table: "portfolios", where: "deleted_at IS NULL"},
	{key: "video", table: "video_works", where: "deleted_at IS NULL"},
	{key: "song", table: "songs"},
	{key: "equipment", table: "photo_equipment", where: "deleted_at IS NULL"},
	{key: "project", table: "projects", where: "deleted_at IS NULL"},
	{key: "open_source", table: "open_source_works", where: "deleted_at IS NULL"},
	{key: "recipe", table: "recipes", where: "deleted_at IS NULL"},
	{key: "book", table: "books", where: "deleted_at IS NULL"},
	{key: "game", table: "games", where: "deleted_at IS NULL"},
	{key: "fitness", table: "fitness_records", where: "deleted_at IS NULL"},
	{key: "tech_stack", table: "tech_stack_items", where: "deleted_at IS NULL"},
}

// TopArticleRow 热门文章排行数据行。
type TopArticleRow struct {
	ID           string     `gorm:"column:id"`
	Title        string     `gorm:"column:title"`
	Slug         string     `gorm:"column:slug"`
	ViewCount    int64      `gorm:"column:view_count"`
	CommentCount int64      `gorm:"column:comment_count"`
	PublishedAt  *time.Time `gorm:"column:published_at"`
}

// TopTravelRow 热门旅行攻略排行数据行。
type TopTravelRow struct {
	ID           string     `gorm:"column:id"`
	Title        string     `gorm:"column:title"`
	ViewCount    int64      `gorm:"column:view_count"`
	CommentCount int64      `gorm:"column:comment_count"`
	PublishedAt  *time.Time `gorm:"column:published_at"`
}

// RecentCommentRow 最近评论数据行。
type RecentCommentRow struct {
	ID          string    `gorm:"column:id"`
	Nickname    string    `gorm:"column:nickname"`
	Content     string    `gorm:"column:content"`
	TargetType  string    `gorm:"column:target_type"`
	TargetID    string    `gorm:"column:target_id"`
	TargetTitle string    `gorm:"column:target_title"`
	IsBlogger   bool      `gorm:"column:is_blogger"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// AnalyticsModel 工作台统计模型操作结构体。
type AnalyticsModel struct {
	db *gorm.DB
}

// NewAnalytics 创建 AnalyticsModel 实例。
func NewAnalytics() *AnalyticsModel {
	return &AnalyticsModel{db: DB}
}

// CountArticles 统计文章总数、已发布数(status=2)和草稿数(status=1)。
func (m *AnalyticsModel) CountArticles(ctx context.Context) (total, published, draft int64, err error) {
	type countRow struct {
		Total     int64 `gorm:"column:total"`
		Published int64 `gorm:"column:published"`
		Draft     int64 `gorm:"column:draft"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table("articles").
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 2) AS published, COUNT(*) FILTER (WHERE status = 1) AS draft").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, 0, 0, fmt.Errorf("统计文章数量失败: %w", err)
	}
	return row.Total, row.Published, row.Draft, nil
}

// CountPortfolios 统计作品集总数和已发布数(status=1)。
func (m *AnalyticsModel) CountPortfolios(ctx context.Context) (total, published int64, err error) {
	type countRow struct {
		Total     int64 `gorm:"column:total"`
		Published int64 `gorm:"column:published"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table("portfolios").
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 1) AS published").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, 0, fmt.Errorf("统计作品集数量失败: %w", err)
	}
	return row.Total, row.Published, nil
}

// CountVideos 统计视频总数和已发布数(status=1)。
func (m *AnalyticsModel) CountVideos(ctx context.Context) (total, published int64, err error) {
	type countRow struct {
		Total     int64 `gorm:"column:total"`
		Published int64 `gorm:"column:published"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table("video_works").
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 1) AS published").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, 0, fmt.Errorf("统计视频数量失败: %w", err)
	}
	return row.Total, row.Published, nil
}

// CountTravelGuides 统计旅行攻略总数和已发布数(status=2)。
func (m *AnalyticsModel) CountTravelGuides(ctx context.Context) (total, published int64, err error) {
	type countRow struct {
		Total     int64 `gorm:"column:total"`
		Published int64 `gorm:"column:published"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table("travel_guides").
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 2) AS published").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, 0, fmt.Errorf("统计旅行攻略数量失败: %w", err)
	}
	return row.Total, row.Published, nil
}

// CountSongs 统计歌曲总数（songs 表无软删除字段）。
func (m *AnalyticsModel) CountSongs(ctx context.Context) (total int64, err error) {
	type countRow struct {
		Total int64 `gorm:"column:total"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table("songs").
		Select("COUNT(*) AS total").
		Scan(&row).Error
	if err != nil {
		return 0, fmt.Errorf("统计歌曲数量失败: %w", err)
	}
	return row.Total, nil
}

// countWithStatus 通用计数 helper：统计指定表的总数和已发布数（status=1），表需有 status 和 deleted_at 字段。
func (m *AnalyticsModel) countWithStatus(ctx context.Context, table, label string) (total, published int64, err error) {
	type countRow struct {
		Total     int64 `gorm:"column:total"`
		Published int64 `gorm:"column:published"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table(table).
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 1) AS published").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, 0, fmt.Errorf("统计%s数量失败: %w", label, err)
	}
	return row.Total, row.Published, nil
}

// CountProjects 统计项目总数和已发布数(status=1)。
func (m *AnalyticsModel) CountProjects(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "projects", "项目")
}

// CountOpenSourceWorks 统计开源作品总数和已发布数(status=1)。
func (m *AnalyticsModel) CountOpenSourceWorks(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "open_source_works", "开源作品")
}

// CountRecipes 统计菜谱总数和已发布数(status=1)。
func (m *AnalyticsModel) CountRecipes(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "recipes", "菜谱")
}

// CountBooks 统计读书总数和已发布数(status=1)。
func (m *AnalyticsModel) CountBooks(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "books", "读书")
}

// CountGames 统计游戏总数和已发布数(status=1)。
func (m *AnalyticsModel) CountGames(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "games", "游戏")
}

// CountFitnessRecords 统计健身训练记录总数和已发布数(status=1)。
func (m *AnalyticsModel) CountFitnessRecords(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "fitness_records", "健身训练")
}

// CountTechStackItems 统计技术栈条目总数和已发布数(status=1)。
func (m *AnalyticsModel) CountTechStackItems(ctx context.Context) (total, published int64, err error) {
	return m.countWithStatus(ctx, "tech_stack_items", "技术栈")
}

// CountEquipment 统计个人设备总数（photo_equipment 表无状态字段）。
func (m *AnalyticsModel) CountEquipment(ctx context.Context) (total int64, err error) {
	type countRow struct {
		Total int64 `gorm:"column:total"`
	}
	var row countRow
	err = m.db.WithContext(ctx).
		Table("photo_equipment").
		Select("COUNT(*) AS total").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, fmt.Errorf("统计个人设备数量失败: %w", err)
	}
	return row.Total, nil
}

// CountComments 统计评论总数。
func (m *AnalyticsModel) CountComments(ctx context.Context) (total int64, err error) {
	err = m.db.WithContext(ctx).
		Table("comments").
		Where("deleted_at IS NULL").
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("统计评论数量失败: %w", err)
	}
	return total, nil
}

// SumArticleViews 统计文章浏览量总和。
func (m *AnalyticsModel) SumArticleViews(ctx context.Context) (int64, error) {
	type sumRow struct {
		Total int64 `gorm:"column:total"`
	}
	var row sumRow
	err := m.db.WithContext(ctx).
		Table("articles").
		Select("COALESCE(SUM(view_count), 0) AS total").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, fmt.Errorf("统计文章浏览量失败: %w", err)
	}
	return row.Total, nil
}

// SumTravelViews 统计旅行攻略浏览量总和。
func (m *AnalyticsModel) SumTravelViews(ctx context.Context) (int64, error) {
	type sumRow struct {
		Total int64 `gorm:"column:total"`
	}
	var row sumRow
	err := m.db.WithContext(ctx).
		Table("travel_guides").
		Select("COALESCE(SUM(view_count), 0) AS total").
		Where("deleted_at IS NULL").
		Scan(&row).Error
	if err != nil {
		return 0, fmt.Errorf("统计旅行攻略浏览量失败: %w", err)
	}
	return row.Total, nil
}

// GetLastPublishDate 查询最近已发布文章的 published_at。
func (m *AnalyticsModel) GetLastPublishDate(ctx context.Context) (*time.Time, error) {
	type dateRow struct {
		PublishedAt *time.Time `gorm:"column:published_at"`
	}
	var row dateRow
	err := m.db.WithContext(ctx).
		Table("articles").
		Select("published_at").
		Where("status = ? AND deleted_at IS NULL AND published_at IS NOT NULL", 2).
		Order("published_at DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return nil, fmt.Errorf("查询最后发布时间失败: %w", err)
	}
	return row.PublishedAt, nil
}

// GetContentTrend 按日统计指定天数内各内容模块表的新增数量，补全无数据的日期为 0。
func (m *AnalyticsModel) GetContentTrend(ctx context.Context, days int) ([]TrendRow, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))

	type dateCountRow struct {
		Date  time.Time `gorm:"column:date"`
		Count int64     `gorm:"column:count"`
	}

	// 并行查询各内容表的日增量，countsMap: 内容类型 → 日期 → 数量
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)
	countsMap := make(map[string]map[string]int64, len(trendSources))

	for _, src := range trendSources {
		wg.Add(1)
		go func(src trendSource) {
			defer wg.Done()
			query := m.db.WithContext(ctx).
				Table(src.table).
				Select("date_trunc('day', created_at) as date, COUNT(*) as count").
				Where("created_at >= ?", startDate)
			if src.where != "" {
				query = query.Where(src.where)
			}
			var rows []dateCountRow
			if err := query.Group("date_trunc('day', created_at)").Order("date").Scan(&rows).Error; err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("查询%s趋势失败: %w", src.key, err)
				}
				mu.Unlock()
				return
			}
			dateCounts := make(map[string]int64, len(rows))
			for _, r := range rows {
				dateCounts[r.Date.Format("2006-01-02")] = r.Count
			}
			mu.Lock()
			countsMap[src.key] = dateCounts
			mu.Unlock()
		}(src)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// 合并到完整日期序列，无数据的类型与日期补 0
	rows := make([]TrendRow, 0, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		counts := make(map[string]int64, len(trendSources))
		for _, src := range trendSources {
			counts[src.key] = countsMap[src.key][dateStr]
		}
		rows = append(rows, TrendRow{Date: date, Counts: counts})
	}
	return rows, nil
}

// GetTopArticles 查询已发布文章(status=2)热门排行，按 view_count 或 comment_count DESC 排序。
func (m *AnalyticsModel) GetTopArticles(ctx context.Context, sort string, limit int) ([]TopArticleRow, error) {
	orderField := "view_count"
	if sort == "comments" {
		orderField = "comment_count"
	}
	var rows []TopArticleRow
	err := m.db.WithContext(ctx).
		Table("articles").
		Select("id, title, slug, view_count, comment_count, published_at").
		Where("status = ? AND deleted_at IS NULL", 2).
		Order(orderField + " DESC, published_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询热门文章失败: %w", err)
	}
	return rows, nil
}

// GetTopTravelGuides 查询已发布旅行攻略(status=2)热门排行，按 view_count 或 like_count DESC 排序。
// comment_count 使用 review_count 代替，published_at 使用 created_at 代替。
func (m *AnalyticsModel) GetTopTravelGuides(ctx context.Context, sort string, limit int) ([]TopTravelRow, error) {
	orderField := "view_count"
	if sort == "comments" {
		orderField = "review_count"
	}
	var rows []TopTravelRow
	err := m.db.WithContext(ctx).
		Table("travel_guides").
		Select("id, title, view_count, review_count AS comment_count, created_at AS published_at").
		Where("status = ? AND deleted_at IS NULL", 2).
		Order(orderField + " DESC, created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询热门旅行攻略失败: %w", err)
	}
	return rows, nil
}

// GetRecentComments 查询最新 N 条访客评论（排除博主回复），LEFT JOIN articles 和 travel_guides 获取目标标题。
func (m *AnalyticsModel) GetRecentComments(ctx context.Context, limit int) ([]RecentCommentRow, error) {
	var rows []RecentCommentRow
	err := m.db.WithContext(ctx).
		Table("comments c").
		Select("c.id, c.nickname, c.content, c.target_type, c.target_id, c.is_blogger, c.created_at, COALESCE(a.title, t.title, '') as target_title").
		Joins("LEFT JOIN articles a ON c.target_type = 'article' AND c.target_id = a.id AND a.deleted_at IS NULL").
		Joins("LEFT JOIN travel_guides t ON c.target_type = 'travel_guide' AND c.target_id = t.id AND t.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.is_blogger = false").
		Order("c.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询最近评论失败: %w", err)
	}
	return rows, nil
}
