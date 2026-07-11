export interface Category {
  id: string
  name: string
  slug: string
  description: string
  sort_order: number
  type: string
  created_at: string
}

export interface CategoryCreateReq {
  name: string
  slug?: string
  description?: string
  sort_order?: number
  type?: string
}

export interface CategoryUpdateReq extends Partial<CategoryCreateReq> {}