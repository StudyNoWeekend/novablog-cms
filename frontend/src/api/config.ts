import request from './request'
import type { PublicConfigRes, ThemeMarketConfigRes, UpdateThemeMarketConfigReq } from '@/types/api'

export const configApi = {
  /** 获取公共配置（免鉴权下发，含官方市场地址默认值） */
  getPublicConfig(): Promise<PublicConfigRes> {
    return request.get('/public/config') as Promise<PublicConfigRes>
  },
  /** 获取官方市场配置 */
  getThemeMarketConfig(): Promise<ThemeMarketConfigRes> {
    return request.get('/theme-market-config') as Promise<ThemeMarketConfigRes>
  },
  /** 更新官方市场地址（持久化到后端并热更新全局默认） */
  updateThemeMarketConfig(data: UpdateThemeMarketConfigReq): Promise<ThemeMarketConfigRes> {
    return request.put('/theme-market-config', data) as Promise<ThemeMarketConfigRes>
  },
}
