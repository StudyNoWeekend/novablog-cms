package req

import "time"

// CreateMCPKeyReq 创建 MCP 密钥请求参数。
type CreateMCPKeyReq struct {
	Name      string     `json:"name" binding:"required,min=1,max=100"`
	ExpiresAt *time.Time `json:"expires_at"` // 可选过期时间，为空表示永久有效
}

// UpdateMCPKeyReq 更新 MCP 密钥请求参数（改名 / 启停）。
type UpdateMCPKeyReq struct {
	Name   *string `json:"name" binding:"omitempty,min=1,max=100"`
	Status *int16  `json:"status" binding:"omitempty,oneof=1 2"`
}

// MCPKeyListReq MCP 密钥列表查询参数。
type MCPKeyListReq struct {
	PageReq
	Keyword *string `form:"keyword" json:"keyword"`
}
