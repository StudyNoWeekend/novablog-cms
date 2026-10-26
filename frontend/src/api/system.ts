import axios from 'axios'
import request from './request'
import type { AppVersionInfo } from '@/types/api'
import { storage } from '@/utils/storage'
import { message } from 'ant-design-vue'
import i18n from '@/locales'

export const systemApi = {
  /** 获取应用版本号（后端构建期注入） */
  getVersion(): Promise<AppVersionInfo> {
    return request.get('/public/version') as Promise<AppVersionInfo>
  },
}

export interface SystemConfigItem {
  key: string
  value: string
  description: string
}

export interface SystemConfigListRes {
  items: SystemConfigItem[]
}

export interface SystemConfigUpdateRes {
  site_url_changed: boolean
}

/** 配置导出文件结构（同时作为导入请求体） */
export interface SystemConfigExportRes {
  version: number | string
  exported_at: string
  secrets_included: boolean
  groups: Record<string, Record<string, string>>
}

export interface SystemConfigImportRes {
  applied: Record<string, number>
}

export const getSystemConfigAPI = () => request.get<SystemConfigListRes>('/system-config')

/** 仅提交要修改的键；值为字符串（清空传空串），null 表示删除该键 */
export const updateSystemConfigAPI = (updates: Record<string, string | null>) =>
  request.put<SystemConfigUpdateRes>('/system-config', { updates })

export const importSystemConfigAPI = (payload: SystemConfigExportRes) =>
  request.post<SystemConfigImportRes>('/system-config/import', payload)

// resolveExportFilename 从 Content-Disposition 解析导出文件名（兼容 filename*=UTF-8'' 编码写法），失败回退默认名
function resolveExportFilename(disposition?: string): string {
  const utf8Match = disposition?.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1])
    } catch {
      // 编码文件名解析失败时继续尝试普通写法
    }
  }
  const plainMatch = disposition?.match(/filename="?([^";]+)"?/i)
  if (plainMatch?.[1]) return plainMatch[1]
  return `system-config-${new Date().toISOString().slice(0, 10)}.json`
}

// saveBlob 通过临时 <a> 标签触发浏览器下载
function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

/**
 * 导出配置为 JSON 文件下载。
 * 响应为二进制流，走共享 request 实例会被响应拦截器按 JSON 剥壳逻辑拒绝（Blob 无 code 字段），
 * 因此这里用原生 axios 携带 Authorization 请求后自行保存文件。
 */
export async function exportSystemConfigAPI(includeSecrets: boolean): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  const token = storage.getToken()
  try {
    const res = await axios.get<Blob>(`${baseURL}/system-config/export`, {
      params: includeSecrets ? { include_secrets: '1' } : undefined,
      responseType: 'blob',
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    })
    saveBlob(res.data, resolveExportFilename(res.headers?.['content-disposition']))
  } catch (err) {
    // 失败响应体可能是 JSON 错误（Blob 形式），尽量解析出后端 msg 提示
    const blob = (err as { response?: { data?: unknown } }).response?.data
    if (blob instanceof Blob && blob.type.includes('application/json')) {
      try {
        const parsed = JSON.parse(await blob.text()) as { msg?: string }
        if (parsed?.msg) {
          message.error(parsed.msg)
          return
        }
      } catch {
        // 解析失败则使用通用错误提示
      }
    }
    message.error(i18n.global.t('system.export.failed'))
  }
}
