export interface Category {
  id: string
  name: string
  slug: string
  description: string
  sort_order: number
  created_at: string
}

export interface CategoryCreateReq {
  name: string
  slug?: string
  description?: string
  sort_order?: number
}

export interface CategoryUpdateReq extends Partial<CategoryCreateReq> {}