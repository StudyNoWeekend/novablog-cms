import request from './request'

// CMS 内容推送到 NovaBlog 社区（系统设置-社区推送 + 各内容列表/编辑页按钮）

export interface CommunityPushConfig {
  base_url: string
  site_url: string
  push_key_prefix: string
  has_push_key: boolean
  verify_code: string
}

// 保存配置的返回：附验证码提交结果（skipped=未填写 verified=已通过 failed=被拒绝）
export interface UpdateCommunityPushConfigResult extends CommunityPushConfig {
  verify_status?: 'verified' | 'failed' | 'skipped'
  verify_message?: string
}

export interface CommunityPingInfo {
  username: string
  blog_url: string
  credential_name: string
}

export type CommunitySyncStatus = 'none' | 'synced' | 'hidden' | 'failed' | 'deleted'

export interface CommunitySyncState {
  target_type: string
  target_id: string
  sync_status: CommunitySyncStatus
  community_article_id: number
  last_error?: string
  pushed_at?: string
}

export interface CommunitySyncRecord extends CommunitySyncState {
  content_title: string
  retry_count: number
  updated_at: string
}

export type CommunityPushAction = 'push' | 'hide' | 'remove'

export interface CommunityPushSyncResult {
  result: string
  community_id: number
  sync_status: CommunitySyncStatus
  content_title?: string
}

export const communityPushApi = {
  getConfig: () => request.get<CommunityPushConfig>('/community-push/config'),

  updateConfig: (data: { base_url: string; site_url: string; push_key?: string; verify_code?: string }) =>
    request.put<UpdateCommunityPushConfigResult>('/community-push/config', data),

  testConnection: () => request.post<CommunityPingInfo>('/community-push/test'),

  getState: (targetType: string, targetId: string) =>
    request.get<CommunitySyncState>('/community-push/state', {
      params: { target_type: targetType, target_id: targetId },
    }),

  getStates: (targetType: string, ids: string[]) =>
    request.get<Record<string, CommunitySyncState>>('/community-push/states', {
      params: { target_type: targetType, ids: ids.join(',') },
    }),

  sync: (targetType: string, targetId: string, action: CommunityPushAction) =>
    request.post<CommunityPushSyncResult>('/community-push/sync', {
      target_type: targetType,
      target_id: targetId,
      action,
    }),

  getRecords: (params: { page?: number; size?: number; target_type?: string; sync_status?: string }) =>
    request.get<{ total: number; page: number; size: number; items: CommunitySyncRecord[] }>(
      '/community-push/records',
      { params },
    ),

  getTypes: () =>
    request.get<{ id: number; slug: string; name: string; sort_order: number }[]>('/community-push/types'),
}
