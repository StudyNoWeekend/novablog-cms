import request from './request'
import type { AppVersionInfo } from '@/types/api'

export const systemApi = {
  /** 获取应用版本号（后端构建期注入） */
  getVersion(): Promise<AppVersionInfo> {
    return request.get('/public/version') as Promise<AppVersionInfo>
  },
}
