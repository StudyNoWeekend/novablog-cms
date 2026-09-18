// 第三方歌单
export interface ThirdPartyPlaylist {
  id: string
  title: string
  cover_url: string
  platform: string
  platform_url: string
  description: string
  sort_order: number
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreatePlaylistReq {
  title: string
  cover_url?: string
  platform: string
  platform_url: string
  description?: string
  sort_order?: number
  enabled?: boolean
}

export interface UpdatePlaylistReq {
  title?: string
  cover_url?: string
  platform?: string
  platform_url?: string
  description?: string
  sort_order?: number
  enabled?: boolean
}
