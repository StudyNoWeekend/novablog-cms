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
  total_count: number
  error_count: number
  last_access_at: string
}

export interface AccessStatsQuery {
  page?: number
  page_size?: number
  ip?: string
}
