import axios from 'axios'
import type { AxiosError } from 'axios'
import { message } from 'ant-design-vue'

import request from './request'
import { marketStorage } from '@/utils/storage'
import { MARKET_AUTH_EXPIRED_CODE } from '@/types/template'
import type { ApiResponse, PaginatedData } from '@/types/api'
import type {
  HotTag,
  MarketLoginRes,
  MarketTokenPair,
  ThemeDetail,
  ThemeItem,
  ThemeMarketListParams,
  ThemeStats,
} from '@/types/template'

// MARKET_AUTH_EXPIRED_EVENT 官方登录失效事件名（由模板页面监听并弹回登录遮罩）
export const MARKET_AUTH_EXPIRED_EVENT = 'novablog:market-auth-expired'

// buildMarketHeaders 构造官方转发所需请求头：官方地址必带，官方 Token 有则带。
function buildMarketHeaders(baseURL?: string): Record<string, string> {
  const headers: Record<string, string> = {}
  const url = baseURL || marketStorage.getBaseURL()
  if (url) headers['X-Market-Base-URL'] = url
  const token = marketStorage.getToken()
  if (token) headers['X-Market-Token'] = token
  return headers
}

// isMarketAuthExpired 判断错误是否为官方登录失效（业务码 401101）。
function isMarketAuthExpired(err: unknown): boolean {
  return (err as AxiosError<ApiResponse>)?.response?.data?.code === MARKET_AUTH_EXPIRED_CODE
}

let refreshing = false
let pendingQueue: Array<{ resolve: (ok: boolean) => void; reject: (err: unknown) => void }> = []

// clearMarketAuth 清空官方凭据并广播登录失效事件，由页面弹回登录遮罩。
function clearMarketAuth() {
  marketStorage.clear()
  const queue = pendingQueue
  pendingQueue = []
  queue.forEach((p) => p.resolve(false))
  window.dispatchEvent(new CustomEvent(MARKET_AUTH_EXPIRED_EVENT))
}

// refreshOfficialToken 用官方 refresh_token 换新 Token（官方轮换式，需保存新双 Token）。
// 失败时清空官方凭据并触发登录遮罩。并发调用会挂起等待同一次刷新结果。
async function refreshOfficialToken(): Promise<boolean> {
  if (refreshing) {
    return new Promise((resolve, reject) => pendingQueue.push({ resolve, reject }))
  }
  refreshing = true
  try {
    const baseURL = marketStorage.getBaseURL()
    const refreshToken = marketStorage.getRefreshToken()
    if (!baseURL || !refreshToken) {
      clearMarketAuth()
      return false
    }

    const resp = await axios.post<ApiResponse<MarketTokenPair>>(
      `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/themes/market/auth/refresh`,
      { refresh_token: refreshToken },
      { headers: { 'X-Market-Base-URL': baseURL } },
    )
    const body = resp.data
    if (body.code === 0 && body.data?.access_token) {
      marketStorage.setToken(body.data.access_token)
      marketStorage.setRefreshToken(body.data.refresh_token)
      const queue = pendingQueue
      pendingQueue = []
      queue.forEach((p) => p.resolve(true))
      return true
    }

    clearMarketAuth()
    message.error('官方登录已失效，请重新登录')
    return false
  } catch {
    clearMarketAuth()
    message.error('官方登录已失效，请重新登录')
    return false
  } finally {
    refreshing = false
  }
}

// withMarketAuth 统一携带官方请求头执行请求；官方登录失效时自动刷新并重试一次。
async function withMarketAuth<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    if (!isMarketAuthExpired(err)) throw err
    const ok = await refreshOfficialToken()
    if (!ok) throw err
    return fn()
  }
}

// themeMarketApi 官方主题市场接口封装（经后台 /themes/market 无状态代理转发）。
export const themeMarketApi = {
  /** 登录官方主题市场（baseURL 来自登录弹窗输入） */
  login(baseURL: string, data: { email: string; password: string }) {
    return request.post<MarketLoginRes>('/themes/market/auth/login', data, {
      headers: buildMarketHeaders(baseURL),
    })
  },

  /** 登出官方账号（官方 Token 拉黑；失败不阻塞本地退出） */
  logout() {
    return withMarketAuth(() =>
      request.post<null>('/themes/market/auth/logout', undefined, { headers: buildMarketHeaders() }),
    )
  },

  /** 官方主题市场列表（仅已上架，支持筛选/搜索/排序/分页） */
  getList(params: ThemeMarketListParams) {
    return withMarketAuth(() =>
      request.get<PaginatedData<ThemeItem>>('/themes/market', {
        params,
        paramsSerializer: { indexes: null },
        headers: buildMarketHeaders(),
      }),
    )
  },

  /** 主题详情（携带官方 Token 时返回个人点赞/评分状态） */
  getDetail(id: number) {
    return withMarketAuth(() =>
      request.get<ThemeDetail>(`/themes/market/detail/${id}`, { headers: buildMarketHeaders() }),
    )
  },

  /** 市场统计（主题总数/作者数/总安装次数） */
  getStats() {
    return withMarketAuth(() =>
      request.get<ThemeStats>('/themes/market/stats', { headers: buildMarketHeaders() }),
    )
  },

  /** 热门风格标签（前 15 个） */
  getHotTags() {
    return withMarketAuth(() =>
      request.get<HotTag[]>('/themes/market/hot-tags', { headers: buildMarketHeaders() }),
    )
  },

  /** 点赞/取消点赞（toggle） */
  toggleLike(id: number) {
    return withMarketAuth(() =>
      request.post<{ liked: boolean }>(`/themes/market/${id}/like`, undefined, { headers: buildMarketHeaders() }),
    )
  },

  /** 收藏/取消收藏（toggle） */
  toggleFavorite(id: number) {
    return withMarketAuth(() =>
      request.post<{ favorited: boolean }>(`/themes/market/${id}/favorite`, undefined, { headers: buildMarketHeaders() }),
    )
  },

  /** 评分（1-5 分，同一用户覆盖式） */
  rate(id: number, score: number) {
    return withMarketAuth(() =>
      request.post<{ rating: number }>(`/themes/market/${id}/rating`, { score }, { headers: buildMarketHeaders() }),
    )
  },

  /** 下载/安装（计数 +1），返回主题包地址 */
  download(id: number) {
    return withMarketAuth(() =>
      request.post<{ download_url: string }>(`/themes/market/${id}/download`, undefined, { headers: buildMarketHeaders() }),
    )
  },

  /** 我的收藏列表（按收藏时间倒序） */
  getFavorites(params: { page: number; page_size: number }) {
    return withMarketAuth(() =>
      request.get<PaginatedData<ThemeItem>>('/themes/market/favorites', {
        params,
        headers: buildMarketHeaders(),
      }),
    )
  },

  /** 收藏主题 ID 集合（用于批量点亮收藏状态） */
  getFavoriteIds() {
    return withMarketAuth(() =>
      request.get<number[]>('/themes/market/favorites/ids', { headers: buildMarketHeaders() }),
    )
  },
}
