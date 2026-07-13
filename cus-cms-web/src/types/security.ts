export interface SecurityConfig {
  get_max_tokens: number
  get_window_seconds: number
  post_max_tokens: number
  post_window_seconds: number
  view_max_tokens: number
  view_window_seconds: number
  like_max_tokens: number
  like_window_seconds: number
  blacklist_threshold: number
  blacklist_ttl_minutes: number
}

export interface UpdateSecurityConfigReq {
  get_max_tokens?: number
  get_window_seconds?: number
  post_max_tokens?: number
  post_window_seconds?: number
  view_max_tokens?: number
  view_window_seconds?: number
  like_max_tokens?: number
  like_window_seconds?: number
  blacklist_threshold?: number
  blacklist_ttl_minutes?: number
}

export interface BlacklistItem {
  id: string
  ip_address: string
  reason: string
  banned_at: string
  expires_at: string
  is_active: boolean
}

export interface DailyCount {
  date: string
  count: number
}

export interface TopIP {
  ip_address: string
  count: number
}

export interface SecurityStats {
  blocked_ip_count: number
  today_rate_limit_count: number
  daily_trend: DailyCount[]
  top_violations: TopIP[]
}

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
