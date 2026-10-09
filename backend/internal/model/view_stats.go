package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// viewTarget 浏览计数目标表配置。
type viewTarget struct {
	table       string
	status      int    // 发布状态值，0 表示表无 status 字段
	softDelete  bool   // 表是否有 deleted_at 字段
	titleColumn string // 热门排行 UNION 的标题列
	commentCol  string // 评论数列，空串表示无（按 0 计）
}

// viewTargets 浏览计数目标白名单：内容类型 → 表配置（新增模块时在此维护）。
var viewTargets = map[string]viewTarget{
	"article":     {"articles", 2, true, "title", "comment_count"},
	"travel":      {"travel_guides", 2, true, "title", "review_count"},
	"portfolio":   {"portfolios", 1, true, "name", ""},
	"video":       {"video_works", 1, true, "title", ""},
	"song":        {"songs", 0, false, "title", ""},
	"equipment":   {"photo_equipment", 0, true, "name", ""},
	"project":     {"projects", 1, true, "title", ""},
	"open_source": {"open_source_works", 1, true, "name", ""},
	"recipe":      {"recipes", 1, true, "title", ""},
	"book":        {"books", 1, true, "title", ""},
	"game":        {"games", 1, true, "title", ""},
	"fitness":     {"fitness_records", 1, true, "title", ""},
	"tech_stack":  {"tech_stack_items", 1, true, "name", ""},
}

// IsValidViewTarget 判断内容类型是否为合法的计数目标。
func IsValidViewTarget(contentType string) bool {
	_, ok := viewTargets[contentType]
	return ok
}

// viewTypeOrder 固定顺序的类型清单（稳定性与 SQL UNION 使用）。
var viewTypeOrder = []string{
	"article", "travel", "portfolio", "video", "song", "equipment", "project",
	"open_source", "recipe", "book", "game", "fitness", "tech_stack",
}

// ViewTypeOrder 返回类型顺序清单的副本。
func ViewTypeOrder() []string {
	return append([]string(nil), viewTypeOrder...)
}

// ContentViewLog 访问明细模型，对应 content_view_logs 表（保留 180 天）。
type ContentViewLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	ContentType string    `gorm:"column:content_type;size:32;index:idx_view_logs_type_id,priority:1"`
	ContentID   string    `gorm:"column:content_id;type:uuid;index:idx_view_logs_type_id,priority:2"`
	IP          string    `gorm:"column:ip;size:64"`
	CreatedAt   time.Time `gorm:"type:timestamptz;autoCreateTime;index"`
}

// TableName 指定数据表名称。
func (ContentViewLog) TableName() string {
	return "content_view_logs"
}

// DailyViewStat 按日聚合模型，对应 daily_view_stats 表（历史永久保留）。
type DailyViewStat struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	StatDate    time.Time `gorm:"column:stat_date;type:date;uniqueIndex:uk_daily_view_stats,priority:1"`
	ContentType string    `gorm:"column:content_type;size:32;uniqueIndex:uk_daily_view_stats,priority:2"`
	PV          int64     `gorm:"column:pv"`
	UV          int64     `gorm:"column:uv"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (DailyViewStat) TableName() string {
	return "daily_view_stats"
}

// TypePvRow 按内容类型聚合的浏览量数据行。
type TypePvRow struct {
	ContentType string `gorm:"column:content_type"`
	PV          int64  `gorm:"column:pv"`
	UV          int64  `gorm:"column:uv"`
}

// DailyTypePvRow 按天 + 内容类型聚合的浏览量数据行（rollup 与趋势共用）。
type DailyTypePvRow struct {
	StatDate    time.Time `gorm:"column:stat_date"`
	ContentType string    `gorm:"column:content_type"`
	PV          int64     `gorm:"column:pv"`
	UV          int64     `gorm:"column:uv"`
}

// ViewStatsModel 访问统计模型操作结构体。
type ViewStatsModel struct {
	db *gorm.DB
}

// NewViewStats 创建 ViewStatsModel 实例。
func NewViewStats() *ViewStatsModel {
	return &ViewStatsModel{db: DB}
}

// InsertViewLog 写入访问明细。
func (m *ViewStatsModel) InsertViewLog(ctx context.Context, contentType, contentID, ip string) error {
	log := ContentViewLog{ContentType: contentType, ContentID: contentID, IP: ip}
	if err := m.db.WithContext(ctx).Create(&log).Error; err != nil {
		return fmt.Errorf("写入访问明细失败: %w", err)
	}
	return nil
}

// ContentViewExists 校验目标内容存在且已发布。
func (m *ViewStatsModel) ContentViewExists(ctx context.Context, contentType, contentID string) (bool, error) {
	target, ok := viewTargets[contentType]
	if !ok {
		return false, fmt.Errorf("未知的内容类型: %s", contentType)
	}
	query := m.db.WithContext(ctx).Table(target.table).Where("id = ?", contentID)
	if target.status > 0 {
		query = query.Where("status = ?", target.status)
	}
	if target.softDelete {
		query = query.Where("deleted_at IS NULL")
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("校验%s存在性失败: %w", contentType, err)
	}
	return count > 0, nil
}

// IncrementContentView 内容表浏览量自增。
func (m *ViewStatsModel) IncrementContentView(ctx context.Context, contentType, contentID string) error {
	target, ok := viewTargets[contentType]
	if !ok {
		return fmt.Errorf("未知的内容类型: %s", contentType)
	}
	if err := m.db.WithContext(ctx).
		Table(target.table).
		Where("id = ?", contentID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return fmt.Errorf("自增%s浏览量失败: %w", contentType, err)
	}
	return nil
}

// SumContentViews 汇总全部内容表浏览量总和（含历史累计值）。
func (m *ViewStatsModel) SumContentViews(ctx context.Context) (int64, error) {
	arms := make([]string, 0, len(viewTargets))
	for _, target := range viewTargets {
		where := ""
		if target.softDelete {
			where = " WHERE deleted_at IS NULL"
		}
		arms = append(arms, "SELECT COALESCE(SUM(view_count), 0) AS c FROM "+target.table+where)
	}
	var row struct {
		Total int64 `gorm:"column:total"`
	}
	sql := "SELECT COALESCE(SUM(c), 0) AS total FROM (" + strings.Join(arms, " UNION ALL ") + ") t"
	if err := m.db.WithContext(ctx).Raw(sql).Scan(&row).Error; err != nil {
		return 0, fmt.Errorf("汇总内容浏览量失败: %w", err)
	}
	return row.Total, nil
}

// SumContentViewsByType 按内容类型汇总内容表浏览量累计值。
func (m *ViewStatsModel) SumContentViewsByType(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64, len(viewTargets))
	for contentType, target := range viewTargets {
		where := ""
		if target.softDelete {
			where = " WHERE deleted_at IS NULL"
		}
		var row struct {
			Total int64 `gorm:"column:total"`
		}
		sql := "SELECT COALESCE(SUM(view_count), 0) AS total FROM " + target.table + where
		if err := m.db.WithContext(ctx).Raw(sql).Scan(&row).Error; err != nil {
			return nil, fmt.Errorf("汇总%s浏览量失败: %w", contentType, err)
		}
		result[contentType] = row.Total
	}
	return result, nil
}

// LogsPvByType 统计起始时间（含）以来明细表按内容类型的 PV/UV。
func (m *ViewStatsModel) LogsPvByType(ctx context.Context, start time.Time) ([]TypePvRow, error) {
	var rows []TypePvRow
	err := m.db.WithContext(ctx).
		Table("content_view_logs").
		Select("content_type, COUNT(*) AS pv, COUNT(DISTINCT ip) AS uv").
		Where("created_at >= ?", start).
		Group("content_type").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("统计访问明细失败: %w", err)
	}
	return rows, nil
}

// LogsGlobalPvUv 统计起始时间（含）以来的全局 PV/UV（IP 跨内容类型去重）。
func (m *ViewStatsModel) LogsGlobalPvUv(ctx context.Context, start time.Time) (pv, uv int64, err error) {
	type row struct {
		PV int64 `gorm:"column:pv"`
		UV int64 `gorm:"column:uv"`
	}
	var r row
	err = m.db.WithContext(ctx).
		Table("content_view_logs").
		Select("COUNT(*) AS pv, COUNT(DISTINCT ip) AS uv").
		Where("created_at >= ?", start).
		Scan(&r).Error
	if err != nil {
		return 0, 0, fmt.Errorf("统计全局访问量失败: %w", err)
	}
	return r.PV, r.UV, nil
}

// RollupDailyViews 将起始时间（含）以来的明细按 (日, 类型) 聚合 UPSERT 进按日聚合表。
// 幂等：重复执行以明细表为准覆盖，可修正部分小时的数据。
func (m *ViewStatsModel) RollupDailyViews(ctx context.Context, start time.Time) (int64, error) {
	sql := `
INSERT INTO daily_view_stats (stat_date, content_type, pv, uv, updated_at)
SELECT date_trunc('day', created_at)::date AS stat_date, content_type, COUNT(*) AS pv, COUNT(DISTINCT ip) AS uv, NOW()
FROM content_view_logs
WHERE created_at >= ?
GROUP BY 1, 2
ON CONFLICT (stat_date, content_type)
DO UPDATE SET pv = EXCLUDED.pv, uv = EXCLUDED.uv, updated_at = NOW()`
	res := m.db.WithContext(ctx).Exec(sql, start)
	if res.Error != nil {
		return 0, fmt.Errorf("聚合访问数据失败: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// CleanLogsBefore 删除指定时间之前的访问明细。
func (m *ViewStatsModel) CleanLogsBefore(ctx context.Context, before time.Time) (int64, error) {
	res := m.db.WithContext(ctx).Where("created_at < ?", before).Delete(&ContentViewLog{})
	return res.RowsAffected, res.Error
}

// TypeCountRow 内容总数与已发布数数据行。
type TypeCountRow struct {
	Total     int64
	Published int64
}

// CountContentAll 统计各内容类型的内容总数与已发布数（无状态字段的表 published=total）。
func (m *ViewStatsModel) CountContentAll(ctx context.Context) (map[string]TypeCountRow, error) {
	result := make(map[string]TypeCountRow, len(viewTypeOrder))
	for _, contentType := range viewTypeOrder {
		target := viewTargets[contentType]
		selectExpr := "COUNT(*) AS total, COUNT(*) AS published"
		where := ""
		if target.status > 0 {
			selectExpr = fmt.Sprintf("COUNT(*) AS total, COUNT(*) FILTER (WHERE status = %d) AS published", target.status)
		}
		if target.softDelete {
			where = " WHERE deleted_at IS NULL"
		}
		var row TypeCountRow
		sql := "SELECT " + selectExpr + " FROM " + target.table + where
		if err := m.db.WithContext(ctx).Raw(sql).Scan(&row).Error; err != nil {
			return nil, fmt.Errorf("统计%s内容数失败: %w", contentType, err)
		}
		result[contentType] = row
	}
	return result, nil
}

// TopAnyRow 全模块热门内容排行数据行（UNION 结果）。
type TopAnyRow struct {
	ID           string    `gorm:"column:id"`
	Type         string    `gorm:"column:type"`
	Title        string    `gorm:"column:title"`
	ViewCount    int64     `gorm:"column:view_count"`
	CommentCount int64     `gorm:"column:comment_count"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// GetTopAllContent 跨全部内容表按浏览量排行（仅已发布内容）。
func (m *ViewStatsModel) GetTopAllContent(ctx context.Context, limit int) ([]TopAnyRow, error) {
	arms := make([]string, 0, len(viewTargets))
	// 固定顺序遍历，保证 UNION 结果稳定
	for _, contentType := range viewTypeOrder {
		target := viewTargets[contentType]
		commentExpr := "0 AS comment_count"
		if target.commentCol != "" {
			commentExpr = target.commentCol + " AS comment_count"
		}
		where := fmt.Sprintf(" WHERE status = %d", target.status)
		if target.status == 0 {
			where = ""
		}
		if target.softDelete {
			if where == "" {
				where = " WHERE deleted_at IS NULL"
			} else {
				where += " AND deleted_at IS NULL"
			}
		}
		arms = append(arms, fmt.Sprintf(
			"SELECT id::text AS id, '%s' AS type, %s AS title, view_count, %s, created_at FROM %s%s",
			contentType, target.titleColumn, commentExpr, target.table, where))
	}
	sql := "SELECT * FROM (" + strings.Join(arms, " UNION ALL ") + ") t ORDER BY view_count DESC, created_at DESC LIMIT ?"
	var rows []TopAnyRow
	if err := m.db.WithContext(ctx).Raw(sql, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询全模块热门内容失败: %w", err)
	}
	return rows, nil
}

// StatsDailySince 读取聚合表中起始日期（含）以来的按天按类型数据（历史永久保留）。
func (m *ViewStatsModel) StatsDailySince(ctx context.Context, start time.Time) ([]DailyTypePvRow, error) {
	var rows []DailyTypePvRow
	err := m.db.WithContext(ctx).
		Table("daily_view_stats").
		Select("stat_date, content_type, pv, uv").
		Where("stat_date >= ?", start.Format("2006-01-02")).
		Order("stat_date").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("读取访问聚合数据失败: %w", err)
	}
	return rows, nil
}
