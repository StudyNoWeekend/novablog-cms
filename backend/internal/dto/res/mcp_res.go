package res

import "time"

// MCPKeyRes MCP 密钥响应结构体（不含明文密钥）。
type MCPKeyRes struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"` // 脱敏展示用前缀，如 nbt_mcp_a1b2c3d4…
	Status     int16      `json:"status"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CallCount  int64      `json:"call_count"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// MCPKeyCreatedRes 创建密钥响应：完整明文密钥仅此一次返回，之后无法再查看。
type MCPKeyCreatedRes struct {
	MCPKeyRes
	Key string `json:"key"` // 完整明文密钥，形如 nbt_mcp_xxx
}
