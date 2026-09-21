export interface OpenSourceWork {
  id: string
  /** 仓库名称 */
  name: string
  /** 仓库链接 */
  repo_url: string
  /** 一句话介绍 */
  summary: string
  /** README 原文（Markdown），列表接口不返回 */
  readme?: string
  /** 主语言 */
  language: string
  /** 主题标签，逗号分隔 */
  topics: string
  /** Star 数（同步时快照） */
  stars: number
  /** 主页/演示地址 */
  homepage: string
  /** 0=草稿, 1=已发布 */
  status: number
  sort_order: number
  /** README 最近拉取时间，null 表示尚未拉取 */
  readme_updated_at?: string | null
  created_at: string
  updated_at: string
}

export interface OpenSourceListRes {
  list: OpenSourceWork[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateOpenSourceReq {
  name: string
  repo_url: string
  summary?: string
  language?: string
  topics?: string
  homepage?: string
  status?: number
  sort_order?: number
  /** 是否拉取远端仓库元数据与 README（默认 true） */
  fetch_remote?: boolean
}

export interface UpdateOpenSourceReq {
  name?: string
  repo_url?: string
  summary?: string
  language?: string
  topics?: string
  homepage?: string
  status?: number
  sort_order?: number
}
