import request from './request'
import type {
  StorageConfig,
  StorageConfigForm,
  StorageTestForm,
  MigrationTask,
  MigrationItem,
  MigrationStartReq,
  AnalyzeResult,
} from '@/types/storage'

// API 响应中的 PaginatedData<MigrationItem>
interface PaginatedItems {
  list: MigrationItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const storageApi = {
  // ===== 存储配置 =====
  // 获取所有配置列表
  getConfigs() {
    return request.get<StorageConfig[]>('/storage/config')
  },
  // 新增/更新配置
  upsertConfig(data: StorageConfigForm) {
    return request.post<StorageConfig>('/storage/config', data)
  },
  // 删除配置
  deleteConfig(provider: string) {
    return request.delete(`/storage/config/${provider}`)
  },
  // 激活配置
  activateConfig(provider: string) {
    return request.put(`/storage/config/${provider}/activate`)
  },
  // 测试连通性（不落库）
  testConfig(data: StorageTestForm) {
    return request.post('/storage/config/test', data)
  },

  // ===== 素材迁移 =====
  // 分析素材
  analyze(targetProvider: string) {
    return request.post<MigrationTask>('/storage/migration/analyze', {
      target_provider: targetProvider,
    } as MigrationStartReq)
  },
  // 获取分析结果
  getAnalyzeResult(taskId: string) {
    return request.get<AnalyzeResult>(`/storage/migration/analyze/${taskId}`)
  },
  // 启动迁移
  startMigration(data: MigrationStartReq) {
    return request.post<MigrationTask>('/storage/migration/start', data)
  },
  // 获取迁移状态/进度
  getMigrationStatus(taskId: string) {
    return request.get<{ task: MigrationTask; items: MigrationItem[] }>(
      `/storage/migration/status/${taskId}`,
    )
  },
}
