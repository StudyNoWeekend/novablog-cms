import request from './request'
import type { MapConfigRes, UpdateMapConfigReq } from '@/types/api'

// 模块级缓存：地图选点器等组件运行时读取 Key，避免每次打开弹窗都请求
let cached: MapConfigRes | null = null
let cachePromise: Promise<MapConfigRes | null> | null = null

function invalidateCache(config: MapConfigRes) {
  cached = config
}

export const mapApi = {
  /** 获取地图服务配置（管理端鉴权接口） */
  async getConfig(): Promise<MapConfigRes> {
    const config = (await request.get('/map-config')) as MapConfigRes
    invalidateCache(config)
    return config
  },
  /** 带缓存的读取：选点器运行时取 Key 使用；接口失败时返回 null（选点器回退编译期 env / OSM 模式） */
  async getConfigCached(): Promise<MapConfigRes | null> {
    if (cached) return cached
    if (!cachePromise) {
      cachePromise = this.getConfig().catch(() => null)
    }
    const result = await cachePromise
    cachePromise = null
    return result
  },
  /** 更新地图服务配置（安全密钥留空 = 保持原值，由后端处理） */
  async updateConfig(data: UpdateMapConfigReq): Promise<MapConfigRes> {
    const config = (await request.put('/map-config', data)) as MapConfigRes
    invalidateCache(config)
    return config
  },
}
