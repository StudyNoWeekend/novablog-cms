import type { ModuleKey } from './modules'

/** 模块配色：指标条色点 / 趋势图系列 / 分布图扇区共用，同一模块全工作台同色 */
export const MODULE_COLORS = {
  article: '#526FE8',
  travel: '#16805D',
  portfolio: '#B7791F',
  video: '#0E7490',
  song: '#EC4899',
  comment: '#8EA5FF',
  equipment: '#0D9488',
  project: '#7C3AED',
  open_source: '#475569',
  recipe: '#D64545',
  book: '#8B5E3C',
  game: '#A21CAF',
  fitness: '#DC2626',
  tech_stack: '#4F46E5',
} as const

/** 内容类型元数据：趋势图系列 / 模块数据表 / 热门排行共用，新增模块时在此维护 */
export interface ContentTypeDef {
  type: string
  label: string
  color: string
  moduleKey: ModuleKey
}

export const CONTENT_TYPE_DEFS: ContentTypeDef[] = [
  { type: 'article', label: '文章', color: MODULE_COLORS.article, moduleKey: 'article_enabled' },
  { type: 'travel', label: '旅行攻略', color: MODULE_COLORS.travel, moduleKey: 'travel_enabled' },
  { type: 'portfolio', label: '摄影作品', color: MODULE_COLORS.portfolio, moduleKey: 'portfolio_enabled' },
  { type: 'video', label: '视频作品', color: MODULE_COLORS.video, moduleKey: 'video_enabled' },
  { type: 'song', label: '音乐', color: MODULE_COLORS.song, moduleKey: 'music_enabled' },
  { type: 'equipment', label: '个人设备', color: MODULE_COLORS.equipment, moduleKey: 'equipment_enabled' },
  { type: 'project', label: '项目经历', color: MODULE_COLORS.project, moduleKey: 'project_enabled' },
  { type: 'open_source', label: '开源作品', color: MODULE_COLORS.open_source, moduleKey: 'open_source_enabled' },
  { type: 'recipe', label: '美食菜谱', color: MODULE_COLORS.recipe, moduleKey: 'recipe_enabled' },
  { type: 'book', label: '读书书架', color: MODULE_COLORS.book, moduleKey: 'book_enabled' },
  { type: 'game', label: '游戏库', color: MODULE_COLORS.game, moduleKey: 'game_enabled' },
  { type: 'fitness', label: '健身训练', color: MODULE_COLORS.fitness, moduleKey: 'fitness_enabled' },
  { type: 'tech_stack', label: '技术栈', color: MODULE_COLORS.tech_stack, moduleKey: 'tech_stack_enabled' },
]

/** 内容类型 → 元数据 快速查找表 */
export const CONTENT_TYPE_MAP: Record<string, ContentTypeDef> = Object.fromEntries(
  CONTENT_TYPE_DEFS.map((d) => [d.type, d])
)
