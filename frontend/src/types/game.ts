export interface Game {
  id: string
  title: string
  cover: string
  platform: string
  genre: string
  play_status: 'want' | 'playing' | 'played'
  play_hours: number
  rating: number
  short_review: string
  status: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface GameListRes {
  list: Game[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateGameReq {
  title: string
  cover?: string
  platform?: string
  genre?: string
  play_status?: 'want' | 'playing' | 'played'
  play_hours?: number
  rating?: number
  short_review?: string
  status?: number
  sort_order?: number
}

export interface UpdateGameReq extends Partial<CreateGameReq> {}
