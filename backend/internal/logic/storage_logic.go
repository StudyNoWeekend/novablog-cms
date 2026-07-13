package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"
	"novablog/utils/crypto"
)

// StorageLogic 存储配置业务逻辑。
type StorageLogic struct {
	model          *model.StorageConfigModel
	manager        *storage.Manager
	cryptoKey      string
	migrationModel *model.StorageMigrationModel // 用于检查迁移锁
}

// NewStorageLogic 创建 StorageLogic 实例。
func NewStorageLogic(manager *storage.Manager, cryptoKey string, migrationModel *model.StorageMigrationModel) *StorageLogic {
	return &StorageLogic{
		model:          model.NewStorageConfig(),
		manager:        manager,
		cryptoKey:      cryptoKey,
		migrationModel: migrationModel,
	}
}

// GetAllConfigs 查询所有存储配置，access_secret 返回 "******" 脱敏。
func (l *StorageLogic) GetAllConfigs(ctx context.Context) ([]res.StorageConfigRes, error) {
	configs, err := l.model.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询存储配置列表失败: %w", err)
	}

	result := make([]res.StorageConfigRes, 0, len(configs))
	for _, c := range configs {
		result = append(result, res.StorageConfigRes{
			ID:           c.ID,
			Provider:     c.Provider,
			Endpoint:     c.Endpoint,
			Region:       c.Region,
			Bucket:       c.Bucket,
			AccessKey:    c.AccessKey,
			AccessSecret: "******",
			PathPrefix:   c.PathPrefix,
			CustomDomain: c.CustomDomain,
			Extra:        c.Extra,
			IsActive:     c.IsActive,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		})
	}
	return result, nil
}

// UpsertConfig 加密 access_secret 后创建或更新存储配置。
func (l *StorageLogic) UpsertConfig(ctx context.Context, r *req.StorageConfigReq) error {
	encryptedSecret, err := crypto.Encrypt(r.AccessSecret, l.cryptoKey)
	if err != nil {
		return fmt.Errorf("加密 access_secret 失败: %w", err)
	}

	config := &model.StorageConfig{
		Provider:     r.Provider,
		Endpoint:     r.Endpoint,
		Region:       r.Region,
		Bucket:       r.Bucket,
		AccessKey:    r.AccessKey,
		AccessSecret: encryptedSecret,
		PathPrefix:   r.PathPrefix,
		CustomDomain: r.CustomDomain,
		Extra:        r.Extra,
	}

	if err := l.model.Upsert(ctx, config); err != nil {
		return fmt.Errorf("保存存储配置失败: %w", err)
	}
	return nil
}

// DeleteConfig 删除指定 provider 的存储配置（不能删除激活中的配置）。
func (l *StorageLogic) DeleteConfig(ctx context.Context, provider string) error {
	if err := l.model.Delete(ctx, provider); err != nil {
		return fmt.Errorf("删除存储配置失败: %w", err)
	}
	return nil
}

// ActivateConfig 激活指定 provider 的存储配置。
// 迁移进行中禁止切换，切换后热重载存储管理器。
func (l *StorageLogic) ActivateConfig(ctx context.Context, provider string) error {
	// 检查迁移锁：迁移进行中禁止切换存储
	runningCount, err := l.migrationModel.GetRunningTaskCount(ctx)
	if err != nil {
		return fmt.Errorf("查询迁移任务状态失败: %w", err)
	}
	if runningCount > 0 {
		return fmt.Errorf("存储迁移进行中，禁止切换存储配置")
	}

	if err := l.model.SetActive(ctx, provider); err != nil {
		return fmt.Errorf("激活存储配置失败: %w", err)
	}

	// 热重载存储管理器
	if err := l.manager.Reload(ctx); err != nil {
		return fmt.Errorf("存储配置已激活，但热重载失败: %w", err)
	}
	return nil
}

// TestConfig 测试存储连通性（不落库），AccessSecret 为明文。
func (l *StorageLogic) TestConfig(ctx context.Context, r *req.StorageTestReq) error {
	config := &model.StorageConfig{
		Provider:     r.Provider,
		Endpoint:     r.Endpoint,
		Region:       r.Region,
		Bucket:       r.Bucket,
		AccessKey:    r.AccessKey,
		AccessSecret: r.AccessSecret, // 明文，不落库
		PathPrefix:   r.PathPrefix,
		CustomDomain: r.CustomDomain,
		Extra:        r.Extra,
	}

	if err := l.manager.TestProvider(ctx, config); err != nil {
		return fmt.Errorf("存储连通性测试失败: %w", err)
	}
	return nil
}
