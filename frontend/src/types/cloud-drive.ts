// 云盘导入模块类型（对应后端 cloud_drive_* 接口）

export type CloudDriveProvider = 'webdav' | 'baidu' | 'aliyun'

/** OAuth 应用凭据（对应后端 CloudDriveConfigRes，secret 脱敏） */
export interface CloudDriveConfig {
  id: string
  provider: 'baidu' | 'aliyun'
  client_id: string
  client_secret: string // 脱敏后 "******"
  callback_base_url: string
  created_at: string
  updated_at: string
}

/** 应用凭据保存表单（编辑时 client_secret 留空表示不修改） */
export interface CloudDriveConfigForm {
  provider: 'baidu' | 'aliyun'
  client_id: string
  client_secret: string
  callback_base_url: string
}

/** 发起授权响应 */
export interface CloudAuthorizeRes {
  auth_url: string
  state: string
}

/** 云盘账号 */
export interface CloudDriveAccount {
  id: string
  provider: CloudDriveProvider
  open_user_id: string
  display_name: string
  /** WebDAV 服务器地址 */
  base_url: string
  /** WebDAV 用户名 */
  username: string
  status: 'active' | 'expired' | 'revoked'
  expires_at: string | null
  quota_total: number
  quota_used: number
  created_at: string
  updated_at: string
}

/** WebDAV 绑定表单 */
export interface WebDAVAccountForm {
  base_url: string
  username: string
  password: string
}

/** 云盘文件 */
export interface CloudFile {
  file_id: string
  parent_id: string
  path: string
  name: string
  is_dir: boolean
  size: number
  ext: string
}

/** 云盘目录浏览响应（游标分页） */
export interface CloudFileListRes {
  list: CloudFile[]
  next_cursor: string
  parent: string
}

/** 导入任务 */
export interface CloudImportTask {
  id: string
  account_id: string
  provider: CloudDriveProvider
  target_folder_id: string | null
  status: 'pending' | 'running' | 'completed' | 'failed'
  total: number
  succeeded: number
  failed: number
  started_at: string | null
  finished_at: string | null
  error: string
  created_at: string
}

/** 导入条目（失败明细） */
export interface CloudImportItem {
  id: string
  source_file_id: string
  source_path: string
  filename: string
  file_size: number
  status: string
  media_id: string
  error: string
}

/** 导入进度响应 */
export interface CloudImportStatusRes {
  task: CloudImportTask
  failed: CloudImportItem[]
}

/** 发起导入请求 */
export interface CloudImportStartReq {
  account_id: string
  files: { file_id: string; path: string; name: string; size: number }[]
  folder_id?: string | null
}
