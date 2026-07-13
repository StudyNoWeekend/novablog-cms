package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"cus-cms/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// activeBundle 当前活跃的 Provider 与脱敏后的配置快照。
type activeBundle struct {
	provider StorageProvider
	config   *model.StorageConfig // access_secret 已清空的脱敏快照
}

// Manager 存储管理器，支持热重载与连通性测试。
type Manager struct {
	active      atomic.Pointer[activeBundle] // 当前活跃 Provider + 配置快照（无锁读）
	mu          sync.Mutex                   // 序列化重建
	configModel *model.StorageConfigModel
	cryptoKey   string
	logger      *zap.Logger
}

// NewManager 创建存储管理器实例。
func NewManager(configModel *model.StorageConfigModel, cryptoKey string, logger *zap.Logger) *Manager {
	return &Manager{
		configModel: configModel,
		cryptoKey:   cryptoKey,
		logger:      logger,
	}
}

// GetProvider 返回当前活跃的 Provider，无锁读。
// 返回 nil 表示尚未配置活跃存储。
func (m *Manager) GetProvider() StorageProvider {
	bundle := m.active.Load()
	if bundle == nil {
		return nil
	}
	return bundle.provider
}

// GetActiveConfig 返回当前活跃配置快照（已脱敏，access_secret 为空）。
// 返回 nil 表示尚未配置活跃存储。
func (m *Manager) GetActiveConfig() *model.StorageConfig {
	bundle := m.active.Load()
	if bundle == nil {
		return nil
	}
	return bundle.config
}

// Reload 从 DB 读取活跃配置，解密密钥并构建 Provider，原子替换。
// 若无活跃配置（ErrRecordNotFound），则清空 active。
func (m *Manager) Reload(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, err := m.configModel.GetActive(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 无活跃配置，清空 active
			m.active.Store(nil)
			m.logger.Info("未找到活跃存储配置，已清空当前 Provider")
			return nil
		}
		return fmt.Errorf("查询活跃存储配置失败: %w", err)
	}

	provider, err := NewProvider(config, m.cryptoKey)
	if err != nil {
		return fmt.Errorf("构建存储 Provider 失败: %w", err)
	}

	// 构建脱敏配置快照
	snapshot := sanitizeConfig(config)

	m.active.Store(&activeBundle{
		provider: provider,
		config:   snapshot,
	})

	m.logger.Info("存储 Provider 已重载",
		zap.String("provider", config.Provider),
		zap.String("bucket", config.Bucket),
	)
	return nil
}

// TestProvider 临时构建 Provider 并校验连通性，不落库不替换 active。
// config.AccessSecret 为明文（用户刚输入未加密）。
func (m *Manager) TestProvider(ctx context.Context, config *model.StorageConfig) error {
	// config.AccessSecret 是明文，直接构建（跳过解密步骤）
	provider, err := buildProvider(config, config.AccessSecret)
	if err != nil {
		return fmt.Errorf("构建存储 Provider 失败: %w", err)
	}

	// 校验连通性：检查一个不存在的 key，返回 false 且无错误表示连通正常
	exist, err := provider.Exists(ctx, "__connection_test__")
	if err != nil {
		return fmt.Errorf("存储连通性校验失败: %w", err)
	}
	if exist {
		// 不应存在但仍返回 true，不影响连通性判断
		m.logger.Warn("连通性测试 key 意外存在，但连接正常",
			zap.String("provider", config.Provider),
		)
	}
	return nil
}

// sanitizeConfig 返回脱敏后的配置快照（access_secret 已清空）。
func sanitizeConfig(config *model.StorageConfig) *model.StorageConfig {
	snapshot := *config
	snapshot.AccessSecret = ""
	return &snapshot
}
