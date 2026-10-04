export interface Book {
  id: string
  title: string
  author: string
  cover: string
  rating: number
  reading_status: 'want' | 'reading' | 'done'
  review: string
  started_at: string | null
  finished_at: string | null
  status: number
  sort_order: number
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
  rating?: number
  reading_status?: 'want' | 'reading' | 'done'
  review?: string
  started_at?: string | null
  finished_at?: string | null
  status?: number
  sort_order?: number
}

export interface UpdateBookReq extends Partial<CreateBookReq> {}
