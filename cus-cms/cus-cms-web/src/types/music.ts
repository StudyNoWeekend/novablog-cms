// 音乐模块的分类复用通用分类表，类型定义见 @/types/category 中的 Category

export interface Song {
  id: string
  title: string
  artist: string
  cover_url: string
  bvid: string
  cid: number
  source_url: string
  source_type: string
  category_id: string | null
  duration: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface SongCreateReq {
  title: string
  artist: string
  cover_url?: string
  bvid: string
  cid: number
  source_url: string
  source_type?: string
  category_id?: string | null
  duration?: number
  sort_order?: number
}

export interface SongUpdateReq {
  title?: string
  artist?: string
  cover_url?: string
  category_id?: string | null
  duration?: number
  sort_order?: number
}

export interface ParseResult {
  title: string
  artist: string
  cover_url: string
  duration: number
  bvid: string
  cid: number
}

export interface ParseTask {
  task_id: string
  status: 'pending' | 'processing' | 'success' | 'failed'
  results?: ParseResult[]
  error?: string
}

export interface BatchCreateSongReq {
  songs: SongCreateReq[]
}
