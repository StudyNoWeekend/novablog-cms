export interface Project {
  id: string
  title: string
  category: string
  role: string
  client: string
  cover_url: string
  summary: string
  description: string
  /** 技能/工具标签，逗号分隔 */
  tech_stack: string
  /** 开始时间 YYYY-MM，空表示未设置 */
  start_date: string
  /** 结束时间 YYYY-MM，空表示至今或未设置 */
  end_date: string
  project_url: string
  repo_url: string
  /** 0=草稿, 1=已发布 */
  status: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ProjectListRes {
  list: Project[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateProjectReq {
  title: string
  category?: string
  role?: string
  client?: string
  cover_url?: string
  summary?: string
  description?: string
  tech_stack?: string
  start_date?: string
  end_date?: string
  project_url?: string
  repo_url?: string
  status?: number
  sort_order?: number
}

export interface UpdateProjectReq {
  title?: string
  category?: string
  role?: string
  client?: string
  cover_url?: string
  summary?: string
  description?: string
  tech_stack?: string
  start_date?: string
  end_date?: string
  project_url?: string
  repo_url?: string
  status?: number
  sort_order?: number
}
