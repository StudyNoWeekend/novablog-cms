package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// CloudImportTask 云盘导入任务模型，对应 cloud_import_tasks 数据表。
// 结构与 storage_migration_tasks 同构：创建即 running，进度靠 DB 计数器 + 前端轮询。
type CloudImportTask struct {
	ID             string     `gorm:"type:uuid;primaryKey"`
	AccountID      string     `gorm:"column:account_id;type:uuid;not null;index"`
	Provider       string     `gorm:"type:varchar(20);not null"`
	TargetFolderID *string    `gorm:"type:uuid;column:target_folder_id"` // NULL=媒体库根目录
	Status         string     `gorm:"type:varchar(20);not null;default:pending"` // pending/running/completed/failed
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
func (CloudImportTask) TableName() string {
	return "cloud_import_tasks"
}

// CloudImportItem 云盘导入条目模型，对应 cloud_import_items 数据表。
type CloudImportItem struct {
	ID               string    `gorm:"type:uuid;primaryKey"`
	TaskID           string    `gorm:"column:task_id;type:uuid;not null;index"`
	SourceFileID     string    `gorm:"column:source_file_id;type:varchar(255);not null"` // 盘方文件 ID（WebDAV 为路径）
	SourcePath       string    `gorm:"column:source_path;type:varchar(500)"`             // 盘内完整路径
	Filename         string    `gorm:"type:varchar(255)"`
	FileSize         int64     `gorm:"type:bigint"`
	TransferredBytes int64     `gorm:"column:transferred_bytes;type:bigint;default:0"` // 已下载字节（进度轮询用，下载中节流回写）
	Status           string    `gorm:"type:varchar(20);not null;default:pending"`      // pending/success/failed
	MediaID          *string   `gorm:"column:media_id;type:uuid"`                      // 入库成功后回填
	Error            string    `gorm:"type:text"`
	CreatedAt        time.Time `gorm:"type:timestamptz;autoCreateTime"`
}

// TableName 指定数据表名称。
func (CloudImportItem) TableName() string {
	return "cloud_import_items"
}

// CloudImportModel 云盘导入模型操作结构体。
type CloudImportModel struct {
	db *gorm.DB
}

// NewCloudImport 创建 CloudImportModel 实例。
func NewCloudImport() *CloudImportModel {
	return &CloudImportModel{db: DB}
}

// CreateTask 创建导入任务。
func (m *CloudImportModel) CreateTask(ctx context.Context, task *CloudImportTask) error {
	return m.db.WithContext(ctx).Create(task).Error
}

// UpdateTask 更新导入任务。
func (m *CloudImportModel) UpdateTask(ctx context.Context, task *CloudImportTask) error {
	return m.db.WithContext(ctx).Save(task).Error
}

// GetTaskByID 根据 ID 查询导入任务。
func (m *CloudImportModel) GetTaskByID(ctx context.Context, id string) (*CloudImportTask, error) {
	var task CloudImportTask
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetRunningTaskCountByAccount 统计指定账号 running 状态的任务数（同账号串行导入）。
func (m *CloudImportModel) GetRunningTaskCountByAccount(ctx context.Context, accountID string) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&CloudImportTask{}).
		Where("account_id = ? AND status = ?", accountID, "running").Count(&count).Error
	return count, err
}

// RecoverRunningTasks 将所有 running 任务置为 failed（服务重启中断自愈），返回受影响行数。
func (m *CloudImportModel) RecoverRunningTasks(ctx context.Context) (int64, error) {
	res := m.db.WithContext(ctx).Model(&CloudImportTask{}).
		Where("status = ?", "running").
		Updates(map[string]interface{}{
			"status": "failed",
			"error":  "服务重启，导入任务中断",
		})
	return res.RowsAffected, res.Error
}

// GetTasks 分页查询导入任务，按创建时间倒序。
func (m *CloudImportModel) GetTasks(ctx context.Context, page, pageSize int) ([]CloudImportTask, int64, error) {
	var total int64
	if err := m.db.WithContext(ctx).Model(&CloudImportTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tasks []CloudImportTask
	offset := (page - 1) * pageSize
	err := m.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

// CreateItems 批量创建导入条目。
func (m *CloudImportModel) CreateItems(ctx context.Context, items []CloudImportItem) error {
	if len(items) == 0 {
		return nil
	}
	return m.db.WithContext(ctx).Create(&items).Error
}

// UpdateItem 更新导入条目。
func (m *CloudImportModel) UpdateItem(ctx context.Context, item *CloudImportItem) error {
	return m.db.WithContext(ctx).Save(item).Error
}

// UpdateItemProgress 按列更新条目已下载字节（下载中高频进度回写，避免 Save 全量）。
func (m *CloudImportModel) UpdateItemProgress(ctx context.Context, itemID string, transferred int64) error {
	return m.db.WithContext(ctx).Model(&CloudImportItem{}).
		Where("id = ?", itemID).
		Updates(map[string]interface{}{"transferred_bytes": transferred}).Error
}

// GetItemsByTaskID 根据任务 ID 查询所有导入条目（按创建顺序）。
func (m *CloudImportModel) GetItemsByTaskID(ctx context.Context, taskID string) ([]CloudImportItem, error) {
	var items []CloudImportItem
	err := m.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at").Find(&items).Error
	return items, err
}
