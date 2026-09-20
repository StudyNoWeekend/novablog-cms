import request from './request'
import { marketStorage } from '@/utils/storage'
import type { InstalledTheme, ThemeInstallStatus } from '@/types/template'

// buildMarketHeaders 构造官方转发所需请求头。
function buildThemeMarketHeaders(): Record<string, string> {
  const headers: Record<string, string> = {}
  const url = marketStorage.getBaseURL()
  if (url) headers['X-Market-Base-URL'] = url
  return headers
}

// 长任务超时：安装/更新需要下载制品（GitHub 直连可能较慢），
// 放宽到 10 分钟与后端下载客户端超时一致，避免默认 15s 超时中断任务。
const LONG_TASK_TIMEOUT = 10 * 60 * 1000

// themeApi 已安装主题管理接口封装（后台 JWT 鉴权）。
export const themeApi = {
  /** 已安装主题列表（含激活标记） */
  getList() {
    return request.get<InstalledTheme[]>('/themes') as Promise<InstalledTheme[]>
  },

  /** 从官方市场安装主题（长任务：下载/校验/解压/落库） */
  install(data: { theme_id: number; version?: string; force?: boolean }) {
    return request.post<InstalledTheme>('/themes/install', data, {
      headers: buildThemeMarketHeaders(),
      timeout: LONG_TASK_TIMEOUT,
    }) as Promise<InstalledTheme>
  },

  /** 激活主题实例（秒级切换，访客侧立即生效） */
  activate(id: string) {
    return request.post<InstalledTheme>(`/themes/${id}/activate`) as Promise<InstalledTheme>
  },

  /** 卸载主题实例（使用中的主题拒绝卸载） */
  uninstall(id: string) {
    return request.delete<null>(`/themes/${id}`) as Promise<null>
  },

  /** 从官方市场更新已安装主题到最新版本（自动切换激活） */
  updateTheme(id: string) {
    return request.post<InstalledTheme>(`/themes/${id}/update`, undefined, {
      headers: buildThemeMarketHeaders(),
      timeout: LONG_TASK_TIMEOUT,
    }) as Promise<InstalledTheme>
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
