package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 社区同步状态。
const (
	CommunitySyncStatusSynced  = "synced"  // 已发布且与社区一致
	CommunitySyncStatusHidden  = "hidden"  // 社区侧已下架（数据保留）
	CommunitySyncStatusFailed  = "failed"  // 最近一次同步失败（重试 worker 处理）
	CommunitySyncStatusDeleted = "deleted" // 已从社区删除
)

// CommunitySyncRecord CMS 内容与 NovaBlog 社区的同步状态记录，对应 community_sync_records 表。
// TargetID 即社区推送契约的 external_id（CMS 各内容表 UUID 主键）。
// 记录存在 = 该内容已开启社区推送（opt-in），后续更新/下架/删除自动跟随同步。
type CommunitySyncRecord struct {
	ID                 string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TargetType         string     `gorm:"type:varchar(30);not null;uniqueIndex:idx_community_target,priority:1"` // article/travel/recipe/portfolio/open_source
	TargetID           string     `gorm:"type:uuid;not null;uniqueIndex:idx_community_target,priority:2"`
	ContentTitle       string     `gorm:"type:varchar(255);not null;default:''"` // 冗余标题，便于记录列表展示
	CommunityArticleID uint       `gorm:"type:int;default:0"`                    // 社区侧 article_id
	SyncStatus         string     `gorm:"type:varchar(20);not null;default:synced;index"`
	LastError          string     `gorm:"type:varchar(1000);not null;default:''"`
	ContentHash        string     `gorm:"type:varchar(64);not null;default:''"` // 上次成功推送的 payload 摘要，用于跳过无变化推送
	RetryCount         int        `gorm:"type:int;default:0"`
	PushedAt           *time.Time `gorm:"type:timestamptz"`
	CreatedAt          time.Time  `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (CommunitySyncRecord) TableName() string {
	return "community_sync_records"
}

// CommunitySyncRecordModel 同步状态记录模型操作结构体。
type CommunitySyncRecordModel struct {
	db *gorm.DB
}

// NewCommunitySyncRecord 创建 CommunitySyncRecordModel 实例。
func NewCommunitySyncRecord() *CommunitySyncRecordModel {
	return &CommunitySyncRecordModel{db: DB}
}

// GetByTarget 按目标读取同步记录；不存在返回 nil（未开启推送）。
func (m *CommunitySyncRecordModel) GetByTarget(ctx context.Context, targetType, targetID string) (*CommunitySyncRecord, error) {
	var record CommunitySyncRecord
	err := m.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetListByTargets 批量读取同步记录。
func (m *CommunitySyncRecordModel) GetListByTargets(ctx context.Context, targetType string, targetIDs []string) ([]CommunitySyncRecord, error) {
	if len(targetIDs) == 0 {
		return nil, nil
	}
	var records []CommunitySyncRecord
	err := m.db.WithContext(ctx).
		Where("target_type = ? AND target_id IN ?", targetType, targetIDs).Find(&records).Error
	return records, err
}

// Save 保存记录（upsert by ID）。
func (m *CommunitySyncRecordModel) Save(ctx context.Context, record *CommunitySyncRecord) error {
	return m.db.WithContext(ctx).Save(record).Error
}

// UpsertByTarget 按目标 upsert 同步记录（并发下依赖唯一索引兜底）。
func (m *CommunitySyncRecordModel) UpsertByTarget(ctx context.Context, record *CommunitySyncRecord) error {
	existing, err := m.GetByTarget(ctx, record.TargetType, record.TargetID)
	if err != nil {
		return err
	}
	if existing == nil {
		record.ID = uuid.New().String()
		return m.db.WithContext(ctx).Create(record).Error
	}
	record.ID = existing.ID
	record.CreatedAt = existing.CreatedAt
	return m.db.WithContext(ctx).Save(record).Error
}

// GetFailed 重试 worker：读取失败且未超重试上限的记录。
func (m *CommunitySyncRecordModel) GetFailed(ctx context.Context, maxRetry int, limit int) ([]CommunitySyncRecord, error) {
	var records []CommunitySyncRecord
	err := m.db.WithContext(ctx).
		Where("sync_status = ? AND retry_count < ?", CommunitySyncStatusFailed, maxRetry).
		Order("updated_at ASC").Limit(limit).Find(&records).Error
	return records, err
}

// GetList 分页同步记录（供管理页展示）。
func (m *CommunitySyncRecordModel) GetList(ctx context.Context, page, size int, targetType, syncStatus string) ([]CommunitySyncRecord, int64, error) {
	query := m.db.WithContext(ctx).Model(&CommunitySyncRecord{})
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if syncStatus != "" {
		query = query.Where("sync_status = ?", syncStatus)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []CommunitySyncRecord
	if err := query.Order("updated_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
