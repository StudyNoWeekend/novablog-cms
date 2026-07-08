export interface Portfolio {
  id: string
  name: string
  description: string
  cover_mode: number
  cover_preset_id: string
  cover_url: string
  status: number
  sort_order: number
  item_count: number
  created_at: string
  updated_at: string
}

export interface PortfolioItem {
  id: string
  portfolio_id: string
  preset_id: string
  title: string
  description: string
  sort_order: number
  output_url: string
  mime_type: string
  output_size: number
  created_at: string
  updated_at: string
}

export interface PortfolioDetail extends Portfolio {
  items: PortfolioItem[]
}

export interface PortfolioListRes {
  list: Portfolio[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreatePortfolioReq {
  name: string
  description?: string
  cover_mode?: number
  cover_preset_id?: string
}

export interface UpdatePortfolioReq extends Partial<CreatePortfolioReq> {
  status?: number
}

export interface CreatePortfolioItemReq {
  preset_id: string
  title: string
  description?: string
}

export interface UpdatePortfolioItemReq extends Partial<CreatePortfolioItemReq> {}

export interface SortPortfolioItemsReq {
  items: { id: string; sort_order: number }[]
}
