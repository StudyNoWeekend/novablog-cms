export interface TechStackItem {
  id: string
  name: string
  /** 兼容镜像：通常等于分类 slug（language/framework/tool/database 等） */
  category: string
  category_id: string
  category_name: string
  icon: string
  /** 图标对应的媒体库文件 ID（非空时后端优先按 ID 解析） */
  icon_media_id?: string
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
  category_id?: string
  /** 兼容保留，建议用 category_id */
  category?: string
  icon?: string
  /** 图标对应的媒体库文件 ID（非空时后端优先按 ID 解析，为空时用 icon 外链） */
  icon_media_id?: string
  level?: number
  description?: string
  status?: number
  sort_order?: number
}

export interface UpdateTechStackReq extends Partial<CreateTechStackReq> {}
