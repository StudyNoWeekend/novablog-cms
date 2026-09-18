import request from './request'
import type { CorsConfigRes, UpdateCorsConfigReq } from '@/types/api'

export const corsApi = {
  /** 获取跨域配置 */
  getConfig(): Promise<CorsConfigRes> {
    return request.get('/cors-config') as Promise<CorsConfigRes>
  },
  /** 更新跨域配置 */
  updateConfig(data: UpdateCorsConfigReq): Promise<void> {
    return request.put('/cors-config', data) as Promise<void>
  },
}
