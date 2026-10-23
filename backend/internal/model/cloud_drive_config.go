package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CloudDriveConfig 云盘应用凭据配置模型，对应 cloud_drive_configs 数据表。
// 自部署场景下 OAuth 凭据由站长在开放平台申请后自行填写（百度/阿里各一行），
// 凭据留空时预留厂商内置默认值兜底（二期中转回调模式）。
type CloudDriveConfig struct {
	ID              string    `gorm:"type:uuid;primaryKey"`
	Provider        string    `gorm:"type:varchar(20);uniqueIndex;not null"` // baidu/aliyun
	ClientID        string    `gorm:"type:varchar(255)"`
	ClientSecret    string    `gorm:"type:text"` // AES加密
	CallbackBaseURL string    `gorm:"column:callback_base_url;type:varchar(255)"` // 本实例外部可达地址，用于拼 OAuth redirect_uri
	Extra           string    `gorm:"type:text"` // JSON 预留
	CreatedAt       time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (CloudDriveConfig) TableName() string {
	return "cloud_drive_configs"
}

// CloudDriveConfigModel 云盘应用凭据配置模型操作结构体。
type CloudDriveConfigModel struct {
	db *gorm.DB
}

// NewCloudDriveConfig 创建 CloudDriveConfigModel 实例。
func NewCloudDriveConfig() *CloudDriveConfigModel {
	return &CloudDriveConfigModel{db: DB}
}

// GetAll 返回所有云盘应用配置。
func (m *CloudDriveConfigModel) GetAll(ctx context.Context) ([]CloudDriveConfig, error) {
	var configs []CloudDriveConfig
	err := m.db.WithContext(ctx).Order("created_at").Find(&configs).Error
	return configs, err
}

// GetByProvider 根据云盘 provider 查询应用配置。
func (m *CloudDriveConfigModel) GetByProvider(ctx context.Context, provider string) (*CloudDriveConfig, error) {
	var config CloudDriveConfig
	err := m.db.WithContext(ctx).Where("provider = ?", provider).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Upsert 按 provider 执行 upsert：不存在则创建，存在则更新非空字段（secret 留空表示不修改）。
func (m *CloudDriveConfigModel) Upsert(ctx context.Context, config *CloudDriveConfig) error {
	var existing CloudDriveConfig
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

	updates := map[string]interface{}{}
	if config.ClientID != "" {
		updates["client_id"] = config.ClientID
	}
	if config.ClientSecret != "" {
		updates["client_secret"] = config.ClientSecret
	}
	if config.CallbackBaseURL != "" {
		updates["callback_base_url"] = config.CallbackBaseURL
	}
	if config.Extra != "" {
		updates["extra"] = config.Extra
	}
	if len(updates) == 0 {
		return nil
	}
	return m.db.WithContext(ctx).Model(&CloudDriveConfig{}).
		Where("provider = ?", config.Provider).Updates(updates).Error
}

// Delete 删除指定 provider 的云盘应用配置。
func (m *CloudDriveConfigModel) Delete(ctx context.Context, provider string) error {
	return m.db.WithContext(ctx).Where("provider = ?", provider).Delete(&CloudDriveConfig{}).Error
}
