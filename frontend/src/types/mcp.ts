/** MCP 密钥项（不含明文密钥，key_prefix 为脱敏展示前缀） */
export interface McpKeyItem {
  id: string
  name: string
  key_prefix: string
  /** 1 启用 / 2 禁用 */
  status: number
  last_used_at: string | null
  call_count: number
  expires_at: string | null
  created_at: string
  updated_at: string
}

/** 创建密钥请求（expires_at 为空表示永久有效） */
export interface CreateMcpKeyReq {
  name: string
  expires_at?: string | null
}

/** 创建密钥响应：完整明文密钥仅此一次返回 */
export interface CreateMcpKeyRes extends McpKeyItem {
  key: string
}

export interface UpdateMcpKeyReq {
  name?: string
  status?: number
}

export interface McpKeyQuery {
  page?: number
  page_size?: number
  keyword?: string
}

/** 密钥状态常量 */
export const MCP_KEY_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
} as const
