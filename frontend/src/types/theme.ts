// 官方主题市场相关类型，字段与 NovaBlog 官方接口文档 ThemeItem 对齐

// MARKET_AUTH_EXPIRED_CODE 官方登录已失效的业务码
// （后端把官方 401 映射为该码并以 HTTP 400 返回，避免与后台管理员自身的 401 刷新流程冲突）
export const MARKET_AUTH_EXPIRED_CODE = 401101

export interface ThemeItem {
  id: number
  title: string
  slug: string
  description: string
  author_id: string
  author: string
  price: string
  price_amount: number
  type: string
  styles: string[]
  features: string[]
  preview: string
  version: string
  downloads: number
  likes: number
  rating: number
  status: number
  created_at: string
}

export interface ThemeDetail extends ThemeItem {
  download_url: string
  liked: boolean
  user_rating: number
}

/** 主题版本历史条目（官方自 GitHub Release 同步的更新日志） */
export interface ThemeReleaseItem {
  version: string
  tag: string
  notes: string
  published_at: string
}

export interface ThemeStats {
  total: number
  authors: number
  downloads: number
}

export interface HotTag {
  name: string
  count: number
}

export interface ThemeMarketFilters {
  type?: string
  styles?: string[]
  price?: string
  search?: string
  sort?: string
}

export interface ThemeMarketListParams extends ThemeMarketFilters {
  page: number
  page_size: number
}

export interface MarketUser {
  id: string
  username: string
  email: string
  avatar: string
  role: string
}

export interface MarketLoginRes {
  access_token: string
  expires_in: number
  user: MarketUser
}

export interface MarketTokenPair {
  access_token: string
  expires_in: number
}

/** 已安装主题实例（后台 themes 表） */
export interface InstalledTheme {
  id: string
  theme_id: string
  name: string
  version: string
  engine: string
  api_compat: string
  author: string
  description: string
  source: 'official' | 'builtin'
  screenshots: string[]
  fallbacks: Record<string, string>
  market_id: number
  market_slug: string
  artifact_path: string
  checksum: string
  active: boolean
  created_at: string
}

/** 首装主题安装任务状态 */
export interface ThemeInstallStatus {
  status: 'not_started' | 'running' | 'success' | 'failed'
  stage: 'fetching' | 'installing' | 'activating' | 'done' | ''
  message: string
  theme_id: string
  theme_name: string
  theme_version: string
  started_at?: string | null
  finished_at?: string | null
}

/** 主题预览地址（新窗口打开） */
export function themePreviewURL(themeId: string): string {
  return `/preview/${themeId}/`
}
