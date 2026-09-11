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
  refresh_token: string
  expires_in: number
  user: MarketUser
}

export interface MarketTokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}
