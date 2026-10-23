import request from './request'
import type {
  CloudDriveConfig,
  CloudDriveConfigForm,
  CloudAuthorizeRes,
  CloudDriveAccount,
  WebDAVAccountForm,
  CloudFileListRes,
  CloudImportStartReq,
  CloudImportStatusRes,
  CloudImportTask,
} from '@/types/cloud-drive'

export const cloudDriveApi = {
  // ===== OAuth 应用凭据（百度/阿里开放平台）=====
  getConfigs() {
    return request.get<CloudDriveConfig[]>('/clouddrive/config')
  },
  upsertConfig(data: CloudDriveConfigForm) {
    return request.post<void>('/clouddrive/config', data)
  },
  deleteConfig(provider: string) {
    return request.delete<void>(`/clouddrive/config/${provider}`)
  },

  // ===== 授权 =====
  /** 获取官方授权页地址（前端 window.open 打开，回调由后端处理） */
  getAuthorizeURL(provider: string) {
    return request.post<CloudAuthorizeRes>(`/clouddrive/authorize/${provider}`)
  },

  // ===== 账号 =====
  getAccounts() {
    return request.get<CloudDriveAccount[]>('/clouddrive/accounts')
  },
  /** 绑定 WebDAV 账号（后端先测试连通再落库） */
  addWebDAVAccount(data: WebDAVAccountForm) {
    return request.post<CloudDriveAccount>('/clouddrive/accounts/webdav', data)
  },
  /** 测试 WebDAV 连通性（不落库） */
  testWebDAV(data: WebDAVAccountForm) {
    return request.post<void>('/clouddrive/accounts/webdav/test', data)
  },
  removeAccount(id: string) {
    return request.delete<void>(`/clouddrive/accounts/${id}`)
  },
  /** 浏览云盘目录（游标分页：webdav/百度按路径或页码，阿里按 file_id/marker） */
  listFiles(accountId: string, params: { parent?: string; cursor?: string; limit?: number }) {
    return request.get<CloudFileListRes>(`/clouddrive/accounts/${accountId}/files`, { params })
  },

  // ===== 导入 =====
  startImport(data: CloudImportStartReq) {
    return request.post<{ task_id: string }>('/clouddrive/import/start', data)
  },
  getImportStatus(taskId: string) {
    return request.get<CloudImportStatusRes>(`/clouddrive/import/status/${taskId}`)
  },
  getImportTasks(params?: { page?: number; page_size?: number }) {
    return request.get<{ list: CloudImportTask[]; total: number; page: number; page_size: number; total_pages: number }>(
      '/clouddrive/import/tasks',
      { params },
    )
  },
}
