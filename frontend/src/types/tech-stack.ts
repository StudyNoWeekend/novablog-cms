export interface TechStackItem {
  id: string
  name: string
  category: string
  icon: string
  level: number
  description: string
  status: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface TechStackListRes {
  list: TechStackItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateTechStackReq {
  name: string
  category?: string
  icon?: string
  level?: number
  description?: string
  status?: number
  sort_order?: number
}

export interface UpdateTechStackReq extends Partial<CreateTechStackReq> {}
