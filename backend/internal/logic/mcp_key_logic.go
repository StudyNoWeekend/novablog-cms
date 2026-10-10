package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// MCPKeyPrefix MCP 密钥前缀。
const MCPKeyPrefix = "nbt_mcp_"

// MCPKeyMaxCount 单个站点允许创建的最大密钥数量。
const MCPKeyMaxCount = 20

// MCPKeyLogic MCP 密钥业务逻辑结构体。
type MCPKeyLogic struct {
	model *model.MCPAPIKeyModel
}

// NewMCPKeyLogic 创建 MCPKeyLogic 实例。
func NewMCPKeyLogic() *MCPKeyLogic {
	return &MCPKeyLogic{model: model.NewMCPAPIKey()}
}

// GenerateKey 生成完整明文密钥与展示用前缀。
func GenerateKey() (key, prefix string, err error) {
	buf := make([]byte, 24)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("生成随机密钥失败: %w", err)
	}
	key = MCPKeyPrefix + hex.EncodeToString(buf)
	prefix = key[:len(MCPKeyPrefix)+8]
	return key, prefix, nil
}

// HashKey 计算密钥的 SHA-256 十六进制哈希。
func HashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// GetList 分页查询密钥列表。
func (l *MCPKeyLogic) GetList(ctx context.Context, r *req.MCPKeyListReq) (*res.PageRes[res.MCPKeyRes], error) {
	keys, total, err := l.model.GetList(ctx, r.GetPage(), r.GetPageSize(), r.Keyword)
	if err != nil {
		return nil, fmt.Errorf("查询 MCP 密钥列表失败: %w", err)
	}
	items := make([]res.MCPKeyRes, 0, len(keys))
	for i := range keys {
		items = append(items, toMCPKeyRes(&keys[i]))
	}
	return res.NewPageRes(items, total, r.GetPage(), r.GetPageSize()), nil
}

// Create 创建密钥：明文仅在本次响应中返回，之后无法查看。
func (l *MCPKeyLogic) Create(ctx context.Context, r *req.CreateMCPKeyReq) (*res.MCPKeyCreatedRes, error) {
	if r.ExpiresAt != nil && r.ExpiresAt.Before(time.Now()) {
		return nil, enum.ErrInvalidParam
	}

	// 名称查重
	keys, total, err := l.model.GetList(ctx, 1, MCPKeyMaxCount, &r.Name)
	if err != nil {
		return nil, fmt.Errorf("检查密钥名称失败: %w", err)
	}
	if len(keys) > 0 {
		return nil, enum.ErrMCPKeyNameExists
	}
	// 数量上限
	if total >= MCPKeyMaxCount {
		return nil, enum.ErrMCPKeyLimit
	}

	key, prefix, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	record := &model.MCPAPIKey{
		ID:        uuid.New().String(),
		Name:      r.Name,
		KeyHash:   HashKey(key),
		KeyPrefix: prefix,
		Status:    model.MCPKeyStatusEnabled,
		ExpiresAt: r.ExpiresAt,
	}
	if err := l.model.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("保存 MCP 密钥失败: %w", err)
	}

	return &res.MCPKeyCreatedRes{
		MCPKeyRes: toMCPKeyRes(record),
		Key:       key,
	}, nil
}

// Update 更新密钥（改名 / 启停）。
func (l *MCPKeyLogic) Update(ctx context.Context, id string, r *req.UpdateMCPKeyReq) (*res.MCPKeyRes, error) {
	record, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, enum.ErrMCPKeyNotFound
	}
	if r.Name != nil && *r.Name != record.Name {
		keys, _, err := l.model.GetList(ctx, 1, MCPKeyMaxCount, r.Name)
		if err != nil {
			return nil, fmt.Errorf("检查密钥名称失败: %w", err)
		}
		if len(keys) > 0 && keys[0].ID != id {
			return nil, enum.ErrMCPKeyNameExists
		}
		record.Name = *r.Name
	}
	if r.Status != nil {
		record.Status = *r.Status
	}
	if err := l.model.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("更新 MCP 密钥失败: %w", err)
	}
	result := toMCPKeyRes(record)
	return &result, nil
}

// Delete 删除密钥（硬删除）。
func (l *MCPKeyLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.model.GetByID(ctx, id); err != nil {
		return enum.ErrMCPKeyNotFound
	}
	return l.model.Delete(ctx, id)
}

// VerifyKey 校验明文密钥，返回密钥记录；失败返回业务错误（供 MCP 鉴权中间件使用）。
func (l *MCPKeyLogic) VerifyKey(ctx context.Context, key string) (*model.MCPAPIKey, error) {
	record, err := l.model.GetByHash(ctx, HashKey(key))
	if err != nil {
		return nil, enum.ErrMCPKeyInvalid
	}
	if record.Status != model.MCPKeyStatusEnabled {
		return nil, enum.ErrMCPKeyDisabled
	}
	if record.ExpiresAt != nil && record.ExpiresAt.Before(time.Now()) {
		return nil, enum.ErrMCPKeyExpired
	}
	return record, nil
}

// RecordUsage 记录密钥使用统计（调用次数 + 最近使用时间）。
func (l *MCPKeyLogic) RecordUsage(ctx context.Context, id string) {
	// 统计失败不影响主流程
	_ = l.model.RecordUsage(ctx, id)
}

// toMCPKeyRes 转换为密钥响应结构体。
func toMCPKeyRes(k *model.MCPAPIKey) res.MCPKeyRes {
	return res.MCPKeyRes{
		ID:         k.ID,
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix + "…",
		Status:     k.Status,
		LastUsedAt: k.LastUsedAt,
		CallCount:  k.CallCount,
		ExpiresAt:  k.ExpiresAt,
		CreatedAt:  k.CreatedAt,
		UpdatedAt:  k.UpdatedAt,
	}
}
