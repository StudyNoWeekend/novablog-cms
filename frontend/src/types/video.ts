export interface VideoPlatformLink {
  id: string
  video_id: string
  platform: string
  url: string
  created_at: string
  updated_at: string
}

export interface VideoWork {
  id: string
  title: string
  cover_url: string
  /** 封面对应的媒体库文件 ID（非空时后端优先按 ID 解析） */
  cover_media_id?: string
  /** 封面对应的预设成品图 ID */
  cover_preset_id?: string
  description: string
  status: number
  sort_order: number
  category_id: string
  category_name: string
  platforms: VideoPlatformLink[]
  created_at: string
  updated_at: string
}

export interface VideoListRes {
  list: VideoWork[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface PlatformLinkReq {
  platform: string
  url: string
}

export interface CreateVideoReq {
  title: string
  cover_url?: string
  cover_media_id?: string
  cover_preset_id?: string
  description?: string
  category_id?: string
  status?: number
  platforms: PlatformLinkReq[]
}

export interface UpdateVideoReq {
  title?: string
  cover_url?: string
  cover_media_id?: string
  cover_preset_id?: string
  description?: string
  category_id?: string
  status?: number
  platforms?: PlatformLinkReq[]
}

export interface ParseVideoReq {
  platform: string
  url: string
}

export interface ParseVideoRes {
  title: string
  cover_url: string
  description: string
}
