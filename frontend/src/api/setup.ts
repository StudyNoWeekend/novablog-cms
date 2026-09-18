import request from './request'
import type { ThemeInstallStatus } from '@/types/template'

export interface StatusRes {
  initialized: boolean
}

export interface InitReq {
  username: string
  password: string
  nickname?: string
}

export interface InitRes {
  success: boolean
  message: string
}

/** 首装向导存储配置请求（与后端 SetupStorageReq 对应） */
export interface SetupStorageReq {
  provider?: string
  endpoint?: string
  region?: string
  bucket?: string
  access_key?: string
  access_secret?: string
  path_prefix?: string
  custom_domain?: string
  extra?: string
}

export const setupApi = {
  getStatus(): Promise<StatusRes> {
    return request.get('/public/install/status') as Promise<StatusRes>
  },
  init(data: InitReq): Promise<InitRes> {
    return request.post('/public/install/init', data) as Promise<InitRes>
  },
  /** 首装第 3 步：拉取官方默认主题并激活；marketBaseURL 覆盖配置文件 */
  initTheme(marketBaseURL?: string): Promise<ThemeInstallStatus> {
    return request.post('/public/install/theme', { market_base_url: marketBaseURL || '' }) as Promise<ThemeInstallStatus>
  },
  /** 轮询首装主题安装任务状态 */
  getThemeStatus(): Promise<ThemeInstallStatus> {
    return request.get('/public/install/theme/status') as Promise<ThemeInstallStatus>
  },
  /** 首装第 2 步：配置对象存储（可选，跳过则使用本地存储） */
  setupStorage(data: SetupStorageReq): Promise<InitRes> {
    return request.post('/public/install/storage', data) as Promise<InitRes>
  },
}