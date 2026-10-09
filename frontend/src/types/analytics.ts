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
  equipment_total: number
  project_total: number
  project_published: number
  open_source_total: number
  open_source_published: number
  recipe_total: number
  recipe_published: number
  book_total: number
  book_published: number
  game_total: number
  game_published: number
  fitness_total: number
  fitness_published: number
  tech_stack_total: number
  tech_stack_published: number
  comment_total: number
  travel_views: number
  last_publish_at: string | null
  days_since_last_publish: number
}

// 内容产出趋势单日数据，counts 按内容类型（article/travel/portfolio/...）统计当日新增数
export interface ContentTrendItem {
  date: string
  counts: Record<string, number>
}

// 内容产出趋势响应
export interface ContentTrendData {
  range: string
  items: ContentTrendItem[]
}

// 热门内容排行项
export interface TopContentItem {
  id: string
  type: 'article' | 'travel'
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
  type?: 'article' | 'travel' | 'all'
  sort?: 'views' | 'comments'
  limit?: number
}

// 访问概览（当日实时 + 昨日聚合 + 累计）
export interface TrafficSummaryData {
  today_pv: number
  today_uv: number
  yesterday_pv: number
  total_pv: number
}

// 访问趋势单日数据，counts 按内容类型统计当日 PV
export interface TrafficTrendItem {
  date: string
  counts: Record<string, number>
  total_pv: number
}

// 访问趋势响应
export interface TrafficTrendData {
  range: string
  items: TrafficTrendItem[]
}

// 模块数据表行：发布数与访问量汇总
export interface ModuleStatRow {
  type: string
  total: number
  published: number
  total_views: number
  today_views: number
  week_views: number
}

// 模块数据表响应
export interface ModuleStatsData {
  items: ModuleStatRow[]
}

// 内容趋势查询参数
export interface ContentTrendParams {
  range?: '7d' | '30d' | '90d'
}

// 最近评论查询参数
export interface RecentCommentsParams {
  limit?: number
}
