import type { UpdateModuleConfigReq } from '@/types/module'

/** 模块开关 key 类型（ModuleConfig 除 updated_at 外的全部键） */
export type ModuleKey = Exclude<keyof UpdateModuleConfigReq, never>

/** 模块元数据：模块管理页与首装向导共用，新增模块时在此维护 */
export interface ModuleDef {
  key: ModuleKey
  label: string
  description: string
}

export const MODULE_DEFS: ModuleDef[] = [
  { key: 'article_enabled', label: '文章管理', description: '博客前台的文章列表、详情展示' },
  { key: 'music_enabled', label: '音乐管理', description: '博客前台的音乐播放器展示' },
  { key: 'video_enabled', label: '视频管理', description: '博客前台的视频作品展示' },
  { key: 'travel_enabled', label: '旅行管理', description: '博客前台的旅行攻略展示' },
  { key: 'portfolio_enabled', label: '作品集管理', description: '博客前台的摄影作品集展示' },
  { key: 'equipment_enabled', label: '个人设备', description: '博客前台的个人设备展示' },
  { key: 'project_enabled', label: '项目经历', description: '博客前台的项目经历展示' },
  { key: 'open_source_enabled', label: '开源作品', description: '博客前台的 GitHub 开源作品展示' },
  { key: 'recipe_enabled', label: '美食菜谱', description: '博客前台的美食菜谱展示' },
  { key: 'book_enabled', label: '读书书架', description: '博客前台的读书书架与书评展示' },
  { key: 'game_enabled', label: '游戏库', description: '博客前台的游戏库与短评展示' },
  { key: 'fitness_enabled', label: '健身训练', description: '博客前台的健身训练记录展示' },
  { key: 'tech_stack_enabled', label: '技术栈', description: '博客前台的技术栈展示' },
]

/** 通用模块：所有角色恒开（媒体库为基础设施，不参与角色与模块管理） */
export const COMMON_MODULE_KEYS: ModuleKey[] = ['article_enabled']
