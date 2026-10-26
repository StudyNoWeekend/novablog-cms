package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SystemConfig 通用系统配置键值模型，对应 system_configs 数据表。
// 收纳分散在 config.yaml 的运行时配置项（站点对外地址、HTTP 代理、GitHub Token、
// 各模块限额等）；数据库为运行时唯一主源，管理后台「系统配置」即时热更。
type SystemConfig struct {
	Key         string    `gorm:"type:varchar(100);primaryKey" json:"key"`
	Value       string    `gorm:"type:text" json:"value"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName 指定数据表名称。
func (SystemConfig) TableName() string {
	return "system_configs"
}

// 系统配置键常量（消费方经 SystemConfigLogic 的类型化访问器读取，勿直接引用字面量）。
const (
	SystemConfigKeySiteURL               = "site_url"                 // 站点对外地址（媒体 URL 基址；权威值，非空时覆盖 yaml upload.base_url）
	SystemConfigKeyProxyURL              = "proxy_url"                // 出网 HTTP 代理（主题/开源作品/图片搜索/云盘/geoip 共用，空=直连）
	SystemConfigKeyGithubToken           = "github_token"             // GitHub 访问令牌（主题 Release 与开源作品查询限流额度）
	SystemConfigKeyMaxArtifactMB         = "max_artifact_mb"          // 主题制品下载上限（MB）
	SystemConfigKeyCloudMaxFileSizeMB    = "cloud_max_file_size_mb"   // 云盘导入单文件上限（MB）
	SystemConfigKeyCloudImportBatchLimit = "cloud_import_batch_limit" // 云盘导入单批文件数上限
	SystemConfigKeyGeoipDownloadURL      = "geoip_download_url"       // geoip 离线库下载地址（空=内置默认源）
)

// SystemConfigModel 系统配置键值模型操作结构体。
type SystemConfigModel struct {
	db *gorm.DB
}

// NewSystemConfig 创建 SystemConfigModel 实例。
func NewSystemConfig() *SystemConfigModel {
	return &SystemConfigModel{db: DB}
}

// GetAll 查询全部系统配置键值。
func (m *SystemConfigModel) GetAll(ctx context.Context) ([]SystemConfig, error) {
	var list []SystemConfig
	err := m.db.WithContext(ctx).Order("key").Find(&list).Error
	return list, err
}

// GetByKey 按键查询单条（无记录返回 gorm.ErrRecordNotFound）。
func (m *SystemConfigModel) GetByKey(ctx context.Context, key string) (*SystemConfig, error) {
	var cfg SystemConfig
	err := m.db.WithContext(ctx).Where("\"key\" = ?", key).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Upsert 写入或更新配置项（description 仅在新建时落库）。
func (m *SystemConfigModel) Upsert(ctx context.Context, key, value, description string) error {
	cfg := SystemConfig{Key: key, Value: value, Description: description}
	return m.db.WithContext(ctx).
		Where("\"key\" = ?", key).
		Assign(SystemConfig{Value: value}).
		FirstOrCreate(&cfg).Error
}

// Delete 删除配置项（恢复内置默认值时使用）。
func (m *SystemConfigModel) Delete(ctx context.Context, key string) error {
	return m.db.WithContext(ctx).Where("\"key\" = ?", key).Delete(&SystemConfig{}).Error
}

// UpsertBatch 事务内批量写入（启动种子导入用；已存在的键不覆盖）。
func (m *SystemConfigModel) UpsertBatch(ctx context.Context, items []SystemConfig, skipExisting bool) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if skipExisting {
				var count int64
				if err := tx.Model(&SystemConfig{}).Where("\"key\" = ?", item.Key).Count(&count).Error; err != nil {
					return err
				}
				if count > 0 {
					continue
				}
			}
			if err := tx.Where("\"key\" = ?", item.Key).Assign(SystemConfig{Value: item.Value}).FirstOrCreate(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
