export interface Equipment {
  id: string
  name: string
  image_url: string
  brand: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface EquipmentListRes {
  list: Equipment[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateEquipmentReq {
  name: string
  image_url?: string
  brand?: string
  description?: string
}

export interface UpdateEquipmentReq {
  name?: string
  image_url?: string
  brand?: string
  description?: string
}
