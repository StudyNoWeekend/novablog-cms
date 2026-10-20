import type { UpdateModuleConfigReq } from '@/types/module'

/** 模块开关 key 类型（ModuleConfig 除 updated_at 外的全部键） */
export type ModuleKey = Exclude<keyof UpdateModuleConfigReq, never>

/**
 * 模块元数据：模块管理页与首装向导共用，新增模块时在此维护；
 * 文案走 i18n：labelKey/descKey 指向 module.items.*（尾段 = ModuleKey 去掉 _enabled 后缀的 camelCase）
 */
export interface ModuleDef {
  key: ModuleKey
  labelKey: string
  descKey: string
}

export const MODULE_DEFS: ModuleDef[] = [
  { key: 'article_enabled', labelKey: 'module.items.article.label', descKey: 'module.items.article.description' },
  { key: 'music_enabled', labelKey: 'module.items.music.label', descKey: 'module.items.music.description' },
  { key: 'video_enabled', labelKey: 'module.items.video.label', descKey: 'module.items.video.description' },
  { key: 'travel_enabled', labelKey: 'module.items.travel.label', descKey: 'module.items.travel.description' },
  { key: 'portfolio_enabled', labelKey: 'module.items.portfolio.label', descKey: 'module.items.portfolio.description' },
  { key: 'equipment_enabled', labelKey: 'module.items.equipment.label', descKey: 'module.items.equipment.description' },
  { key: 'project_enabled', labelKey: 'module.items.project.label', descKey: 'module.items.project.description' },
  { key: 'open_source_enabled', labelKey: 'module.items.openSource.label', descKey: 'module.items.openSource.description' },
  { key: 'recipe_enabled', labelKey: 'module.items.recipe.label', descKey: 'module.items.recipe.description' },
  { key: 'book_enabled', labelKey: 'module.items.book.label', descKey: 'module.items.book.description' },
  { key: 'game_enabled', labelKey: 'module.items.game.label', descKey: 'module.items.game.description' },
  { key: 'fitness_enabled', labelKey: 'module.items.fitness.label', descKey: 'module.items.fitness.description' },
  { key: 'tech_stack_enabled', labelKey: 'module.items.techStack.label', descKey: 'module.items.techStack.description' },
]

/** 通用模块：所有角色恒开（媒体库为基础设施，不参与角色与模块管理） */
export const COMMON_MODULE_KEYS: ModuleKey[] = ['article_enabled']
