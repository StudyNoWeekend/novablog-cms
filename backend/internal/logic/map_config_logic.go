package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/utils/crypto"
)

// MapConfigLogic 地图服务配置业务逻辑结构体。
type MapConfigLogic struct {
	model     *model.MapConfigModel
	cryptoKey string
}

// NewMapConfigLogic 创建 MapConfigLogic 实例。
func NewMapConfigLogic(cryptoKey string) *MapConfigLogic {
	return &MapConfigLogic{
		model:     model.NewMapConfig(),
		cryptoKey: cryptoKey,
	}
}

// GetConfig 获取地图服务配置（安全密钥解密后返回明文）。
func (l *MapConfigLogic) GetConfig(ctx context.Context) (*res.MapConfigRes, error) {
	config, err := l.model.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取地图配置失败: %w", err)
	}

	securityCode := config.AmapSecurityCode
	if securityCode != "" {
		plain, decErr := crypto.Decrypt(securityCode, l.cryptoKey)
		if decErr == nil {
			securityCode = plain
		}
		// 解密失败时保留密文原样返回，避免阻塞配置页
	}

	return &res.MapConfigRes{
		AmapKey:          config.AmapKey,
		AmapSecurityCode: securityCode,
		GoogleKey:        config.GoogleKey,
		UpdatedAt:        config.UpdatedAt,
	}, nil
}

// UpdateConfig 更新地图服务配置（安全密钥非空时加密落库），返回更新后的配置。
func (l *MapConfigLogic) UpdateConfig(ctx context.Context, r *req.UpdateMapConfigReq) (*res.MapConfigRes, error) {
	config, err := l.model.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取地图配置失败: %w", err)
	}

	config.AmapKey = r.AmapKey
	config.GoogleKey = r.GoogleKey
	if r.AmapSecurityCode != "" {
		encrypted, encErr := crypto.Encrypt(r.AmapSecurityCode, l.cryptoKey)
		if encErr != nil {
			return nil, fmt.Errorf("加密安全密钥失败: %w", encErr)
		}
		config.AmapSecurityCode = encrypted
	}

	if err := l.model.UpdateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("更新地图配置失败: %w", err)
	}
	return l.GetConfig(ctx)
}
