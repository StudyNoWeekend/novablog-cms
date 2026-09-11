package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StorageConfig 对象存储配置模型，对应 storage_configs 数据表。
type StorageConfig struct {
	ID           string    `gorm:"type:uuid;primaryKey"`
	Provider     string    `gorm:"type:varchar(20);uniqueIndex;not null"` // aliyun/tencent/minio
	Endpoint     string    `gorm:"type:varchar(255);not null"`
	Region       string    `gorm:"type:varchar(50)"`
	Bucket       string    `gorm:"type:varchar(255);not null"`
	AccessKey    string    `gorm:"column:access_key;type:varchar(255);not null"`
	AccessSecret string    `gorm:"column:access_secret;type:text;not null"` // AES加密
	PathPrefix   string    `gorm:"column:path_prefix;type:varchar(255)"`
	CustomDomain string    `gorm:"column:custom_domain;type:varchar(255)"`
	Extra        string    `gorm:"type:text"` // JSON
	IsActive     bool      `gorm:"column:is_active;not null;default:false"`
	CreatedAt    time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (StorageConfig) TableName() string {
	return "storage_configs"
}

// StorageConfigModel 存储配置模型操作结构体。
type StorageConfigModel struct {
	db *gorm.DB
}

// NewStorageConfig 创建 StorageConfigModel 实例。
func NewStorageConfig() *StorageConfigModel {
	return &StorageConfigModel{db: DB}
}

// GetAll 返回所有存储配置，按 created_at 排序。
func (m *StorageConfigModel) GetAll(ctx context.Context) ([]StorageConfig, error) {
	var configs []StorageConfig
	err := m.db.WithContext(ctx).Order("created_at").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetByProvider 根据存储提供商查询配置。
func (m *StorageConfigModel) GetByProvider(ctx context.Context, provider string) (*StorageConfig, error) {
	var config StorageConfig
	err := m.db.WithContext(ctx).Where("provider = ?", provider).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetActive 返回当前激活的存储配置（is_active=true）。
func (m *StorageConfigModel) GetActive(ctx context.Context) (*StorageConfig, error) {
	var config StorageConfig
	err := m.db.WithContext(ctx).Where("is_active = ?", true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Upsert 根据 provider 执行 upsert：存在则更新非空字段，不存在则创建。
func (m *StorageConfigModel) Upsert(ctx context.Context, config *StorageConfig) error {
	var existing StorageConfig
	err := m.db.WithContext(ctx).Where("provider = ?", config.Provider).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if config.ID == "" {
				config.ID = uuid.New().String()
			}
			return m.db.WithContext(ctx).Create(config).Error
		}
		return err
	}

	// 存在则更新非空字段
	updates := map[string]interface{}{}
	if config.Endpoint != "" {
		updates["endpoint"] = config.Endpoint
	}
	if config.Region != "" {
		updates["region"] = config.Region
	}
	if config.Bucket != "" {
		updates["bucket"] = config.Bucket
	}
	if config.AccessKey != "" {
		updates["access_key"] = config.AccessKey
	}
	if config.AccessSecret != "" {
		updates["access_secret"] = config.AccessSecret
	}
	if config.PathPrefix != "" {
		updates["path_prefix"] = config.PathPrefix
	}
	if config.CustomDomain != "" {
		updates["custom_domain"] = config.CustomDomain
	}
	if config.Extra != "" {
		updates["extra"] = config.Extra
	}

	if len(updates) == 0 {
		return nil
	}
	return m.db.WithContext(ctx).Model(&StorageConfig{}).
		Where("provider = ?", config.Provider).
		Updates(updates).Error
}

// DeactivateAll 将所有存储配置的 is_active 置为 false。
func (m *StorageConfigModel) DeactivateAll(ctx context.Context) error {
	return m.db.WithContext(ctx).Model(&StorageConfig{}).Where("is_active = ?", true).
		Update("is_active", false).Error
}

// Create 创建新的存储配置。
func (m *StorageConfigModel) Create(ctx context.Context, config *StorageConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	return m.db.WithContext(ctx).Create(config).Error
}

// Delete 删除指定 provider 的存储配置，不能删除激活中的配置。
func (m *StorageConfigModel) Delete(ctx context.Context, provider string) error {
	var config StorageConfig
	err := m.db.WithContext(ctx).Where("provider = ?", provider).First(&config).Error
	if err != nil {
		return err
	}
	if config.IsActive {
		return gorm.ErrInvalidData
	}
	return m.db.WithContext(ctx).Where("provider = ?", provider).Delete(&StorageConfig{}).Error
}

// SetActive 事务：先将所有 is_active 置为 false，再将指定 provider 置为 true。
func (m *StorageConfigModel) SetActive(ctx context.Context, provider string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先将所有配置的 is_active 置为 false
		if err := tx.Model(&StorageConfig{}).Where("is_active = ?", true).
			Update("is_active", false).Error; err != nil {
			return err
		}
		// 再将指定 provider 的配置置为 true
		if err := tx.Model(&StorageConfig{}).Where("provider = ?", provider).
			Update("is_active", true).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteNonActive deletes all non-active storage configs.
func (m *StorageConfigModel) DeleteNonActive(ctx context.Context) error {
	return m.db.WithContext(ctx).Where("is_active = ?", false).Delete(&StorageConfig{}).Error
}
