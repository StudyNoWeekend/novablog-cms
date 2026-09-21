package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Comment 评论模型，对应 comments 数据表。
type Comment struct {
	ID         string         `gorm:"type:uuid;primaryKey"`                                                              // UUID 主键
	TargetType string         `gorm:"column:target_type;type:varchar(20);not null;index:idx_comments_target,priority:1"` // 目标类型 article/travel_guide
	TargetID   string         `gorm:"column:target_id;type:uuid;not null;index:idx_comments_target,priority:2"`          // 目标 ID
	ParentID   *string        `gorm:"column:parent_id;type:uuid;index:idx_comments_parent_id"`                           // 父评论 ID，顶级评论为 nil
	BloggerID  *string        `gorm:"type:uuid;column:blogger_id"`                                                       // 博主 ID（博主回复时填充）
	Nickname   string         `gorm:"type:varchar(50);not null"`                                                         // 评论者名称
	Website    string         `gorm:"type:varchar(500)"`                                                                 // 评论者博客地址
	Content    string         `gorm:"type:text;not null"`                                                                // 评论内容
	IsBlogger  bool           `gorm:"type:boolean;default:false;column:is_blogger"`                                      // 是否为博主回复
	Status     int16          `gorm:"type:smallint;default:2"`                                                           // 默认已通过，无需审核
	IPAddress  string         `gorm:"type:varchar(50);column:ip_address"`                                                // 评论者 IP 地址
	CreatedAt  time.Time      `gorm:"type:timestamptz;autoCreateTime"`                                                   // 创建时间
	UpdatedAt  time.Time      `gorm:"type:timestamptz;autoUpdateTime"`                                                   // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index"`                                                                             // 软删除时间
}

// TableName 指定数据表名称。
func (Comment) TableName() string {
	return "comments"
}

// CommentModel 评论模型操作结构体。
type CommentModel struct {
	db *gorm.DB
}

// NewComment 创建 CommentModel 实例。
func NewComment() *CommentModel {
	return &CommentModel{db: DB}
}

// Create 创建评论。
func (m *CommentModel) Create(ctx context.Context, comment *Comment) error {
	return m.db.WithContext(ctx).Create(comment).Error
}

// GetByID 根据 ID 查询评论。
func (m *CommentModel) GetByID(ctx context.Context, id string) (*Comment, error) {
	var comment Comment
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetList 分页查询评论列表，支持 target_type、target_id 筛选和 keyword 模糊搜索 content。
func (m *CommentModel) GetList(ctx context.Context, page, pageSize int, targetType *string, targetID *string, keyword *string) ([]Comment, int64, error) {
	query := m.db.WithContext(ctx).Model(&Comment{})

	if targetType != nil && *targetType != "" {
		query = query.Where("target_type = ?", *targetType)
	}
	if targetID != nil && *targetID != "" {
		query = query.Where("target_id = ?", *targetID)
	}
	if keyword != nil && *keyword != "" {
		query = query.Where("content LIKE ?", "%"+*keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var comments []Comment
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

// SoftDelete 软删除评论。
func (m *CommentModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Comment{}).Error
}

// SoftDeleteByParentID 根据父评论 ID 级联软删除所有子回复。
func (m *CommentModel) SoftDeleteByParentID(ctx context.Context, parentID string) error {
	return m.db.WithContext(ctx).Where("parent_id = ?", parentID).Delete(&Comment{}).Error
}

// CountByParentID 统计指定父评论下未删除的子回复数量。
func (m *CommentModel) CountByParentID(ctx context.Context, parentID string) (int64, error) {
	var total int64
	err := m.db.WithContext(ctx).
		Model(&Comment{}).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetPublicList 分页查询已通过审核的评论列表，按 target_type 和 target_id 筛选。
func (m *CommentModel) GetPublicList(ctx context.Context, targetType, targetID string, page, pageSize int) ([]Comment, int64, error) {
	query := m.db.WithContext(ctx).Model(&Comment{}).
		Where("status = ? AND target_type = ? AND target_id = ?", 2, targetType, targetID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var comments []Comment
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}
