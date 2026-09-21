import request from './request'
import { marketStorage } from '@/utils/storage'
import { MARKET_AUTH_EXPIRED_CODE } from '@/types/theme'
import type { ApiResponse } from '@/types/api'
import type { PaginatedData } from '@/types/api'
import type {
  HotTag,
  InstalledTheme,
  MarketLoginRes,
  ThemeDetail,
  ThemeInstallStatus,
  ThemeItem,
  ThemeMarketListParams,
  ThemeReleaseItem,
  ThemeStats,
} from '@/types/theme'

// 长任务超时：安装/更新需要下载制品（GitHub 直连可能较慢），
// 放宽到 10 分钟与后端下载客户端超时一致，避免默认 15s 超时中断任务。
const LONG_TASK_TIMEOUT = 10 * 60 * 1000

// MARKET_AUTH_EXPIRED_EVENT 官方登录失效事件名（由主题页面监听并弹回登录遮罩）
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

// PendingRetry 因官方登录失效而挂起的请求，等待重新登录后自动重放。
// 各请求返回类型不同，重放时按原请求类型回填，故用 any 承载。
interface PendingRetry {
  run: () => Promise<any>
  resolve: (value: any) => void
  reject: (reason?: unknown) => void
}

// pendingRetries 官方登录失效期间挂起的请求队列（模块级，页面内共享）。
let pendingRetries: PendingRetry[] = []

/**
 * resolveMarketRetries 重新登录成功后调用：用新 Token（buildMarketHeaders 实时读取）重放
 * 挂起请求，并把结果回填到原调用方的 Promise（安装/更新/下载等无需用户手动重试）。
 * 重放仍失败时直接拒绝（不再重新入队，避免死循环），错误照常由响应拦截器提示。
 */
export function resolveMarketRetries() {
  const queue = pendingRetries
  pendingRetries = []
  for (const item of queue) {
    item.run().then(item.resolve, item.reject)
  }
}

/** abortMarketRetries 登录弹窗被关闭或登出时调用：拒绝所有挂起请求，让调用方按钮恢复。 */
export function abortMarketRetries() {
  const queue = pendingRetries
  pendingRetries = []
  for (const item of queue) {
    item.reject(new Error('已取消，未重新登录官方账号'))
  }
}

// withMarketAuth 统一携带官方请求头执行请求；遇官方登录失效（401101）时：
// 清空凭据并广播事件（主题页弹登录遮罩），请求挂起进入等待队列，
// 重新登录成功后自动重放（resolveMarketRetries），弹窗关闭/登出时拒绝（abortMarketRetries）。
async function withMarketAuth<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    const code = (err as { response?: { data?: ApiResponse } })?.response?.data?.code
    if (code !== MARKET_AUTH_EXPIRED_CODE) {
      throw err
    }
    clearMarketAuth()
    return new Promise<T>((resolve, reject) => {
      pendingRetries.push({ run: fn, resolve, reject })
    })
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

  /** 登出官方账号。不走登录态守护：官方侧 Token 已失效时后端视为成功，本地凭据由调用方清理 */
  logout() {
    return request.post<null>('/themes/market/auth/logout', undefined, { headers: buildMarketHeaders() })
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

  /** 下载主题（后端请求官方代理计数并解析出 GitHub 真实直链，可直接打开下载） */
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

// themeApi 已安装主题管理接口封装（后台 JWT 鉴权）。
export const themeApi = {
  /** 已安装主题列表（含激活标记） */
  getList() {
    return request.get<InstalledTheme[]>('/themes') as Promise<InstalledTheme[]>
  },

  /** 从官方市场安装主题（长任务：下载/校验/解压/落库；官方登录失效时挂起，重登后自动重放） */
  install(data: { theme_id: number; version?: string; force?: boolean }) {
    return withMarketAuth(() =>
      request.post<InstalledTheme>('/themes/install', data, {
        headers: buildMarketHeaders(),
        timeout: LONG_TASK_TIMEOUT,
      }),
    ) as Promise<InstalledTheme>
  },

  /** 激活主题实例（秒级切换，访客侧立即生效） */
  activate(id: string) {
    return request.post<InstalledTheme>(`/themes/${id}/activate`) as Promise<InstalledTheme>
  },

  /** 卸载主题实例（使用中的主题拒绝卸载） */
  uninstall(id: string) {
    return request.delete<null>(`/themes/${id}`) as Promise<null>
  },

  /** 从官方市场更新已安装主题到最新版本（自动切换激活；官方登录失效时挂起，重登后自动重放） */
  updateTheme(id: string) {
    return withMarketAuth(() =>
      request.post<InstalledTheme>(`/themes/${id}/update`, undefined, {
        headers: buildMarketHeaders(),
        timeout: LONG_TASK_TIMEOUT,
      }),
    ) as Promise<InstalledTheme>
  },
}

// installThemeApi 首装主题初始化接口（公开，无需登录）。
export const installThemeApi = {
  /** 触发首装：拉取官方默认主题并激活 */
  start() {
    return request.post<ThemeInstallStatus>('/public/install/theme') as Promise<ThemeInstallStatus>
  },

  /** 轮询安装任务状态 */
  getStatus() {
    return request.get<ThemeInstallStatus>('/public/install/theme/status') as Promise<ThemeInstallStatus>
  },
}
