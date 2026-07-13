// 工作台概览统计
export interface OverviewData {
  article_total: number
  article_published: number
  article_draft: number
  portfolio_total: number
  portfolio_published: number
  video_total: number
  video_published: number
  travel_total: number
  travel_published: number
  song_total: number
  comment_total: number
  travel_views: number
  last_publish_at: string | null
  days_since_last_publish: number
}

// 内容产出趋势单日数据
export interface ContentTrendItem {
  date: string
  article_count: number
  travel_count: number
}

// 内容产出趋势响应
export interface ContentTrendData {
  range: string
  items: ContentTrendItem[]
}

// 热门内容排行项
export interface TopContentItem {
  id: string
  title: string
  view_count: number
  comment_count: number
  slug?: string
  published_at?: string | null
}

// 热门内容排行响应
export interface TopContentData {
  type: string
  sort: string
  items: TopContentItem[]
}

// 内容类型分布项
export interface DistributionItem {
  type: string
  name: string
  count: number
  percentage: number
}

// 内容类型分布响应
export interface DistributionData {
  items: DistributionItem[]
  total: number
}

// 最近评论项
export interface RecentComment {
  id: string
  nickname: string
  content: string
  target_type: string
  target_id: string
  target_title: string
  is_blogger: boolean
  created_at: string
}

// 最近评论响应
export interface RecentCommentsData {
  items: RecentComment[]
}

// 热门内容查询参数
export interface TopContentParams {
  type?: 'article' | 'travel'
  sort?: 'views' | 'comments'
  limit?: number
}

// 内容趋势查询参数
export interface ContentTrendParams {
  range?: '7d' | '30d' | '90d'
}

// 最近评论查询参数
export interface RecentCommentsParams {
  limit?: number
}
