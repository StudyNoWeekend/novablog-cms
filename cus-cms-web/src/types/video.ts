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
  description: string
  status: number
  sort_order: number
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
  description?: string
  status?: number
  platforms: PlatformLinkReq[]
}

export interface UpdateVideoReq {
  title?: string
  cover_url?: string
  description?: string
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
