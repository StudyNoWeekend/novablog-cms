import request from './request'
import type { ModuleConfig, UpdateModuleConfigReq } from '@/types/module'

export const moduleApi = {
  getConfig() {
    return request.get<ModuleConfig>('/module-config')
  },
  updateConfig(data: UpdateModuleConfigReq) {
    return request.put('/module-config', data)
  },
}
