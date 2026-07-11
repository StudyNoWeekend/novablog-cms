package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Comment 评论模型，对应 comments 数据表。
type Comment struct {
	ID         string         `gorm:"type:uuid;primaryKey"`                         // UUID 主键
	TargetType string         `gorm:"type:varchar(20);not null;column:target_type"` // 目标类型 article/travel_guide
	TargetID   string         `gorm:"type:uuid;not null;column:target_id"`          // 目标 ID
	ParentID   *string        `gorm:"type:uuid;column:parent_id"`                   // 父评论 ID，顶级评论为 nil
	BloggerID  *string        `gorm:"type:uuid;column:blogger_id"`                  // 博主 ID（博主回复时填充）
	Nickname   string         `gorm:"type:varchar(50);not null"`                    // 评论者昵称
	Email      string         `gorm:"type:varchar(100)"`                            // 评论者邮箱
	Avatar     string         `gorm:"type:varchar(500)"`                            // 评论者头像 URL
	Content    string         `gorm:"type:text;not null"`                           // 评论内容
	IsBlogger  bool           `gorm:"type:boolean;default:false;column:is_blogger"` // 是否为博主回复
	Status     int16          `gorm:"type:smallint;default:1"`                      // 1=待审核 2=已通过 3=已拒绝
	IPAddress  string         `gorm:"type:varchar(50);column:ip_address"`           // 评论者 IP 地址
	CreatedAt  time.Time      `gorm:"type:timestamptz;autoCreateTime"`              // 创建时间
	UpdatedAt  time.Time      `gorm:"type:timestamptz;autoUpdateTime"`              // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index"`                                        // 软删除时间
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

// GetList 分页查询评论列表，支持 target_type、target_id、status 筛选和 keyword 模糊搜索 content。
func (m *CommentModel) GetList(ctx context.Context, page, pageSize int, targetType *string, targetID *string, status *int16, keyword *string) ([]Comment, int64, error) {
	query := m.db.WithContext(ctx).Model(&Comment{})

	if targetType != nil && *targetType != "" {
		query = query.Where("target_type = ?", *targetType)
	}
	if targetID != nil && *targetID != "" {
		query = query.Where("target_id = ?", *targetID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
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

// UpdateStatus 更新评论状态。
func (m *CommentModel) UpdateStatus(ctx context.Context, id string, status int16) error {
	return m.db.WithContext(ctx).Model(&Comment{}).Where("id = ?", id).Update("status", status).Error
}

// SoftDelete 软删除评论。
func (m *CommentModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Comment{}).Error
}

// SoftDeleteByParentID 根据父评论 ID 级联软删除所有子回复。
func (m *CommentModel) SoftDeleteByParentID(ctx context.Context, parentID string) error {
	return m.db.WithContext(ctx).Where("parent_id = ?", parentID).Delete(&Comment{}).Error
}
