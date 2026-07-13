export interface Article {
  id: string
  title: string
  slug: string
  summary: string
  content: string
  cover_image: string
  category_id: string
  category_name?: string
  tag_ids?: string[]
  tag_names?: string[]
  status: number       // 1草稿 2已发布 3已下架
  type: number         // 1Markdown 2富文本
  extra: Record<string, any> | null
  view_count: number
  comment_count: number
  is_top: boolean
  is_comment: boolean
  published_at: string | null
  created_at: string
  updated_at: string
}

export interface ArticleFilters {
  status?: number
  category_id?: string
  keyword?: string
}

export interface ArticleCreateReq {
  title: string
  content: string
  summary?: string
  cover_image?: string
  category_id?: string
  tag_ids?: string[]
  status?: number          // 默认1草稿，可选2直接发布
  type?: number            // 默认1
  is_top?: boolean
  is_comment?: boolean
  extra?: Record<string, any>
}

export interface ArticleUpdateReq extends Partial<ArticleCreateReq> {}