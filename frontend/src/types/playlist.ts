// 第三方歌单
export interface ThirdPartyPlaylist {
  id: string
  title: string
  cover_url: string
  /** 封面对应的媒体库文件 ID（非空时后端优先按 ID 解析） */
  cover_media_id?: string
  /** 封面对应的预设成品图 ID */
  cover_preset_id?: string
  platform: string
  platform_url: string
  /** 平台品牌 Logo（后端内置下发），无封面时用作占位图 */
  platform_logo?: string
  description: string
  sort_order: number
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreatePlaylistReq {
  title: string
  cover_url?: string
  cover_media_id?: string
  cover_preset_id?: string
  platform: string
  platform_url: string
  description?: string
  sort_order?: number
  enabled?: boolean
}

export interface UpdatePlaylistReq {
  title?: string
  cover_url?: string
  cover_media_id?: string
  cover_preset_id?: string
  platform?: string
  platform_url?: string
  description?: string
  sort_order?: number
  enabled?: boolean
}
