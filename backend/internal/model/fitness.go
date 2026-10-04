package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// FitnessRecord 健身训练记录模型，对应 fitness_records 数据表。
type FitnessRecord struct {
	ID          string           `gorm:"type:uuid;primaryKey" json:"id"`
	Date        time.Time        `gorm:"type:timestamptz;not null" json:"date"` // 训练日期
	Title       string           `gorm:"type:varchar(255);not null" json:"title"`
	Type        string           `gorm:"type:varchar(20);default:'strength'" json:"type"`            // strength=力量, cardio=有氧, stretch=拉伸
	DurationMin int              `gorm:"column:duration_min;type:int;default:0" json:"duration_min"` // 时长（分钟）
	Calories    int              `gorm:"type:int;default:0" json:"calories"`                         // 消耗热量（千卡）
	Content     []map[string]any `gorm:"type:jsonb;serializer:json" json:"content"`                  // 动作清单 [{name,sets,reps,note}]
	Notes       string           `gorm:"type:text" json:"notes"`
	Status      int              `gorm:"type:int;default:0" json:"status"` // 0=草稿, 1=已发布
	SortOrder   int              `gorm:"column:sort_order;type:int;default:0" json:"sort_order"`
	CreatedAt   time.Time        `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定数据表名称。
func (FitnessRecord) TableName() string {
	return "fitness_records"
}

// FitnessRecordModel 健身训练记录模型操作结构体。
type FitnessRecordModel struct {
	db *gorm.DB
}

// NewFitnessRecord 创建 FitnessRecordModel 实例。
func NewFitnessRecord() *FitnessRecordModel {
	return &FitnessRecordModel{db: DB}
}

// Create 创建训练记录。
func (m *FitnessRecordModel) Create(ctx context.Context, r *FitnessRecord) error {
	return m.db.WithContext(ctx).Create(r).Error
}

// GetByID 根据 ID 查询训练记录。
func (m *FitnessRecordModel) GetByID(ctx context.Context, id string) (*FitnessRecord, error) {
	var r FitnessRecord
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GetList 分页查询训练记录列表，支持 keyword 模糊搜索标题、type 筛选、status 过滤，按训练日期倒序。
func (m *FitnessRecordModel) GetList(ctx context.Context, keyword *string, recordType *string, status *int, page, pageSize int) ([]FitnessRecord, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&FitnessRecord{})

	if keyword != nil && *keyword != "" {
		query = query.Where("title ILIKE ?", "%"+*keyword+"%")
	}
	if recordType != nil && *recordType != "" && *recordType != "all" {
		query = query.Where("type = ?", *recordType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []FitnessRecord
	offset := (page - 1) * pageSize
	err := query.
		Order("date DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新训练记录。
func (m *FitnessRecordModel) Update(ctx context.Context, r *FitnessRecord) error {
	return m.db.WithContext(ctx).Save(r).Error
}

// SoftDelete 软删除训练记录。
func (m *FitnessRecordModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&FitnessRecord{}).Error
}
