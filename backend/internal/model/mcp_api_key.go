package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// MCP API Key 状态常量。
const (
	MCPKeyStatusEnabled  int16 = 1 // 启用
	MCPKeyStatusDisabled int16 = 2 // 禁用
)

// MCPAPIKey MCP 密钥模型，对应 mcp_api_keys 数据表。
// 仅存储密钥的 SHA-256 哈希，明文只在创建时返回一次。
type MCPAPIKey struct {
	ID         string     `gorm:"type:uuid;primaryKey"`
	Name       string     `gorm:"type:varchar(100);not null"`
	KeyHash    string     `gorm:"type:varchar(64);uniqueIndex;column:key_hash;not null"`
	KeyPrefix  string     `gorm:"type:varchar(24);column:key_prefix;not null"` // 展示用前缀，如 nbt_mcp_a1b2c3d4
	Status     int16      `gorm:"column:status;type:smallint;not null;default:1"`
	LastUsedAt *time.Time `gorm:"column:last_used_at;type:timestamptz"`
	CallCount  int64      `gorm:"column:call_count;type:bigint;default:0"`
	ExpiresAt  *time.Time `gorm:"column:expires_at;type:timestamptz"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (MCPAPIKey) TableName() string {
	return "mcp_api_keys"
}

// MCPAPIKeyModel MCP 密钥模型操作结构体。
type MCPAPIKeyModel struct {
	db *gorm.DB
}

// NewMCPAPIKey 创建 MCPAPIKeyModel 实例。
func NewMCPAPIKey() *MCPAPIKeyModel {
	return &MCPAPIKeyModel{db: DB}
}

// Create 创建密钥记录。
func (m *MCPAPIKeyModel) Create(ctx context.Context, key *MCPAPIKey) error {
	return m.db.WithContext(ctx).Create(key).Error
}

// GetByID 根据 ID 查询密钥。
func (m *MCPAPIKeyModel) GetByID(ctx context.Context, id string) (*MCPAPIKey, error) {
	var key MCPAPIKey
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// GetByHash 根据密钥哈希查询密钥（鉴权用）。
func (m *MCPAPIKeyModel) GetByHash(ctx context.Context, keyHash string) (*MCPAPIKey, error) {
	var key MCPAPIKey
	err := m.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// GetList 分页查询密钥列表，keyword 模糊搜索名称。
func (m *MCPAPIKeyModel) GetList(ctx context.Context, page, pageSize int, keyword *string) (keys []MCPAPIKey, total int64, err error) {
	query := m.db.WithContext(ctx).Model(&MCPAPIKey{})
	if keyword != nil && *keyword != "" {
		query = query.Where("name LIKE ?", "%"+*keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys).Error
	if err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

// Update 更新密钥记录。
func (m *MCPAPIKeyModel) Update(ctx context.Context, key *MCPAPIKey) error {
	return m.db.WithContext(ctx).Save(key).Error
}

// Delete 硬删除密钥记录（key_hash 有唯一索引，软删会导致唯一冲突）。
func (m *MCPAPIKeyModel) Delete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&MCPAPIKey{}).Error
}

// RecordUsage 更新使用统计：递增调用次数并刷新最近使用时间。
func (m *MCPAPIKeyModel) RecordUsage(ctx context.Context, id string) error {
	now := time.Now()
	return m.db.WithContext(ctx).
		Model(&MCPAPIKey{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"call_count":   gorm.Expr("call_count + 1"),
			"last_used_at": &now,
		}).Error
}
