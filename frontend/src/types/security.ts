export interface SecurityConfig {
  security_enabled: boolean
  blacklist_ttl_minutes: number
  log_retention_days: number
}

export interface UpdateSecurityConfigReq {
  security_enabled?: boolean
  blacklist_ttl_minutes?: number
  log_retention_days?: number
}

export interface BlacklistItem {
  id: string
  ip: string
  reason: string
  created_at: string
}

export interface CreateBlacklistReq {
  ip: string
  reason: string
}

export interface UpdateBlacklistReq {
  ip?: string
  reason?: string
}

export interface BlacklistQuery {
  page?: number
  page_size?: number
  keyword?: string
}

export interface IPAccessStats {
  ip: string
  region: string
  total_count: number
  error_count: number
  last_access_at: string
}

export interface AccessStatsQuery {
  page?: number
  page_size?: number
  ip?: string
  region?: string
}

// API 文档相关类型
export interface APIDocParam {
  name: string
  type: string
  required: boolean
  desc: string
}

export interface APIDocField {
  name: string
  type: string
  desc: string
}

export interface APIDocItem {
  module: string
  method: string
  path: string
  description: string
  params: APIDocParam[]
  response: APIDocField[]
}
