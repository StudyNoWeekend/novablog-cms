package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// StorageMigrationTask 存储迁移任务模型，对应 storage_migration_tasks 数据表。
type StorageMigrationTask struct {
	ID             string     `gorm:"type:uuid;primaryKey"`
	TaskType       string     `gorm:"column:task_type;type:varchar(20);not null"` // analyze/migrate
	TargetProvider string     `gorm:"column:target_provider;type:varchar(20);not null"`
	Status         string     `gorm:"type:varchar(20);not null;default:pending"` // pending/running/completed/failed/canceled
	Total          int        `gorm:"default:0"`
	Succeeded      int        `gorm:"default:0"`
	Failed         int        `gorm:"default:0"`
	StartedAt      *time.Time `gorm:"column:started_at;type:timestamptz"`
	FinishedAt     *time.Time `gorm:"column:finished_at;type:timestamptz"`
	Error          string     `gorm:"type:text"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (StorageMigrationTask) TableName() string {
	return "storage_migration_tasks"
}

// StorageMigrationItem 存储迁移条目模型，对应 storage_migration_items 数据表。
type StorageMigrationItem struct {
	ID         string    `gorm:"type:uuid;primaryKey"`
	TaskID     string    `gorm:"column:task_id;type:uuid;not null;index"`
	MediaID    string    `gorm:"column:media_id;type:uuid;not null"`
	SourceType string    `gorm:"column:source_type;type:varchar(20);not null;default:media"` // media/preset
	Status     string    `gorm:"type:varchar(20);not null;default:pending"`                  // pending/success/failed
	Error      string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"type:timestamptz;autoCreateTime"`
}

// TableName 指定数据表名称。
func (StorageMigrationItem) TableName() string {
	return "storage_migration_items"
}

// StorageMigrationModel 存储迁移模型操作结构体。
type StorageMigrationModel struct {
	db *gorm.DB
}

// NewStorageMigration 创建 StorageMigrationModel 实例。
func NewStorageMigration() *StorageMigrationModel {
	return &StorageMigrationModel{db: DB}
}

// CreateTask 创建迁移任务。
func (m *StorageMigrationModel) CreateTask(ctx context.Context, task *StorageMigrationTask) error {
	return m.db.WithContext(ctx).Create(task).Error
}

// UpdateTask 更新迁移任务。
func (m *StorageMigrationModel) UpdateTask(ctx context.Context, task *StorageMigrationTask) error {
	return m.db.WithContext(ctx).Save(task).Error
}

// GetTaskByID 根据 ID 查询迁移任务。
func (m *StorageMigrationModel) GetTaskByID(ctx context.Context, id string) (*StorageMigrationTask, error) {
	var task StorageMigrationTask
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetRunningTaskCount 返回 status=running 的任务数量，用于迁移锁。
func (m *StorageMigrationModel) GetRunningTaskCount(ctx context.Context) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&StorageMigrationTask{}).
		Where("status = ?", "running").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CreateItems 批量创建迁移条目。
func (m *StorageMigrationModel) CreateItems(ctx context.Context, items []StorageMigrationItem) error {
	if len(items) == 0 {
		return nil
	}
	return m.db.WithContext(ctx).Create(&items).Error
}

// UpdateItem 更新迁移条目。
func (m *StorageMigrationModel) UpdateItem(ctx context.Context, item *StorageMigrationItem) error {
	return m.db.WithContext(ctx).Save(item).Error
}

// GetItemsByTaskID 根据任务 ID 查询所有迁移条目。
func (m *StorageMigrationModel) GetItemsByTaskID(ctx context.Context, taskID string) ([]StorageMigrationItem, error) {
	var items []StorageMigrationItem
	err := m.db.WithContext(ctx).Where("task_id = ?", taskID).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
