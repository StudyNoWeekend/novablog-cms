export interface Book {
  id: string
  title: string
  author: string
  cover: string
  cover_media_id?: string
  cover_preset_id?: string
  rating: number
  reading_status: 'want' | 'reading' | 'done'
  review: string
  started_at: string | null
  finished_at: string | null
  status: number
  sort_order: number
  category_id: string
  category_name: string
  created_at: string
  updated_at: string
}

export interface BookListRes {
  list: Book[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateBookReq {
  title: string
  author?: string
  cover?: string
  cover_media_id?: string
  cover_preset_id?: string
  rating?: number
  reading_status?: 'want' | 'reading' | 'done'
  review?: string
  started_at?: string | null
  finished_at?: string | null
  category_id?: string
  status?: number
  sort_order?: number
}

export interface UpdateBookReq extends Partial<CreateBookReq> {}
