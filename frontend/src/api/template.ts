import request from './request'
import { marketStorage } from '@/utils/storage'
import { MARKET_AUTH_EXPIRED_CODE } from '@/types/template'
import type { ApiResponse } from '@/types/api'
import type { PaginatedData } from '@/types/api'
import type {
  HotTag,
  MarketLoginRes,
  ThemeDetail,
  ThemeItem,
  ThemeMarketListParams,
  ThemeReleaseItem,
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

// clearMarketAuth 清空官方凭据并广播登录失效事件，由页面弹回登录遮罩。
function clearMarketAuth() {
  marketStorage.clear()
  window.dispatchEvent(new CustomEvent(MARKET_AUTH_EXPIRED_EVENT))
}

// withMarketAuth 统一携带官方请求头执行请求；遇官方登录失效时广播事件并 reject。
async function withMarketAuth<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    const code = (err as { response?: { data?: ApiResponse } })?.response?.data?.code
    if (code === MARKET_AUTH_EXPIRED_CODE) {
      clearMarketAuth()
      throw err
    }
    throw err
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

  /** 主题版本历史与更新日志（按发布时间倒序，无记录返回空数组） */
  getReleases(id: number) {
    return withMarketAuth(() =>
      request.get<ThemeReleaseItem[]>(`/themes/market/${id}/releases`, { headers: buildMarketHeaders() }),
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
