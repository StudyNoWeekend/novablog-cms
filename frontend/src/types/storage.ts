// 存储配置（对应后端 StorageConfigRes）
export interface StorageConfig {
  id: string
  provider: 'aliyun' | 'tencent' | 'minio' | 'local'
  endpoint: string
  region: string
  bucket: string
  access_key: string
  access_secret: string  // 脱敏后 "******"
  path_prefix: string
  custom_domain: string
  extra: string  // JSON 字符串
  is_active: boolean
  created_at: string
  updated_at: string
}

// 存储配置表单（用于新增/编辑）
export interface StorageConfigForm {
  provider: 'aliyun' | 'tencent' | 'minio' | 'local' | ''
  old_provider: string  // 编辑时传入原 provider
  endpoint: string
  region: string
  bucket: string
  access_key: string
  access_secret: string  // 新建必填，编辑时空表示不修改
  path_prefix: string
  custom_domain: string
  extra: string
  use_ssl?: boolean  // 仅对 minio 有效，前端用于生成 extra JSON
}

// 存储配置测试请求
export interface StorageTestForm {
  provider: 'aliyun' | 'tencent' | 'minio' | 'local'
  endpoint: string
  region: string
  bucket: string
  access_key: string
  access_secret: string
  path_prefix: string
  custom_domain: string
  extra: string
}

// 迁移任务
export interface MigrationTask {
  id: string
  task_type: 'analyze' | 'migrate'
  target_provider: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'canceled'
  total: number
  succeeded: number
  failed: number
  started_at: string
  finished_at: string
  error: string
  created_at: string
}

// 迁移明细
export interface MigrationItem {
  id: string
  media_id: string
  status: 'pending' | 'success' | 'failed'
  error: string
  // 分析场景额外字段（由后端 analyze result 返回）
  filename?: string
  storage_type?: string
  source_type?: string  // media/preset
  url?: string          // 当前 URL
}

// 分析请求
export interface MigrationAnalyzeReq {
  target_provider: string
}

// 迁移启动请求
export interface MigrationStartReq {
  target_provider: string
  media_ids?: string[]
  all?: boolean
}

// 分析结果
export interface AnalyzeResult {
  task: MigrationTask
  missing: MigrationItem[]
  existing: MigrationItem[]
}
