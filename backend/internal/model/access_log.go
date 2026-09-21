package model

import (
	"context"
	"strings"
	"time"
)

// AccessLog IP 访问日志模型，对应 access_logs 数据表。
type AccessLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	IP         string    `gorm:"index;size:64" json:"ip"`
	Method     string    `gorm:"size:16" json:"method"`
	Path       string    `gorm:"size:512" json:"path"`
	StatusCode int       `json:"status_code"`
	UserAgent  string    `gorm:"size:256" json:"user_agent"`
	Region     string    `gorm:"index;size:128" json:"region"` // IP 归属地区（"中国 广东省 深圳市"），写入时解析
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定数据表名称。
func (AccessLog) TableName() string {
	return "access_logs"
}

// AccessLogModel IP 访问日志模型操作结构体。
type AccessLogModel struct{}

// NewAccessLog 创建 AccessLogModel 实例。
func NewAccessLog() *AccessLogModel { return &AccessLogModel{} }

// Create 创建一条访问日志记录。
func (m *AccessLogModel) Create(ctx context.Context, log *AccessLog) error {
	return DB.WithContext(ctx).Create(log).Error
}

// CleanBefore 清理指定时间之前的访问日志。
func (m *AccessLogModel) CleanBefore(ctx context.Context, before time.Time) error {
	return DB.WithContext(ctx).Where("created_at < ?", before).Delete(&AccessLog{}).Error
}

// IPAccessStats 按 IP 聚合的访问统计结果。
type IPAccessStats struct {
	IP           string
	Region       string
	TotalCount   int64
	ErrorCount   int64
	LastAccessAt time.Time
}

// GetIPAccessStatistics 按 IP 聚合访问日志统计，支持分页、IP 前缀搜索与地区模糊搜索。
// 地区取该 IP 最近写入的非空归属地区（同一 IP 的所在地通常一致）。
func (m *AccessLogModel) GetIPAccessStatistics(ctx context.Context, ipKeyword, region string, page, pageSize int) ([]*IPAccessStats, int64, error) {
	var list []*IPAccessStats
	var total int64

	db := DB.WithContext(ctx).Model(&AccessLog{})
	if ipKeyword != "" {
		db = db.Where("ip ILIKE ?", ipKeyword+"%")
	}
	if region != "" {
		db = db.Where("region ILIKE ?", "%"+region+"%")
	}

	if err := db.Select("COUNT(DISTINCT ip)").Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	sql := `SELECT ip,
				   MAX(CASE WHEN region <> '' THEN region END) AS region,
				   COUNT(*) AS total_count,
				   SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS error_count,
				   MAX(created_at) AS last_access_at
			FROM access_logs`
	where, args := accessStatsWhere(ipKeyword, region)
	if where != "" {
		sql += where
	}
	sql += ` GROUP BY ip
		ORDER BY total_count DESC, last_access_at DESC
		LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	if err := db.Raw(sql, args...).Scan(&list).Error; err != nil {
		return nil, 0, err
	}

	if list == nil {
		list = []*IPAccessStats{}
	}
	return list, total, nil
}

// accessStatsWhere 组装访问统计筛选条件（IP 前缀 + 地区模糊），返回 SQL 片段与参数。
func accessStatsWhere(ipKeyword, region string) (string, []any) {
	var conds []string
	var args []any
	if ipKeyword != "" {
		conds = append(conds, "ip ILIKE ?")
		args = append(args, ipKeyword+"%")
	}
	if region != "" {
		conds = append(conds, "region ILIKE ?")
		args = append(args, "%"+region+"%")
	}
	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}
