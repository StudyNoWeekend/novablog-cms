import type { ModuleKey } from '@/constants/modules'

/** 创作方向角色 key，与后端 setup_logic.go 的 rolePresetModules 白名单保持一致 */
export type RoleKey =
  | 'tech'
  | 'digital'
  | 'photo'
  | 'video'
  | 'music'
  | 'travelvlog'
  | 'travel'
  | 'food'
  | 'fitness'
  | 'fashion'
  | 'pet'
  | 'reading'
  | 'gaming'
  | 'designer'
  | 'craft'
  | 'lifestyle'
  | 'all'

/**
 * 角色卡片元数据（icon 为 Ant Design 图标组件名，向导渲染时解析）；
 * 文案一律走 i18n：labelKey/descKey 指向 setup.roles.*（尾段 = RoleKey 原值，描述加 Desc 后缀）
 */
export interface RoleDef {
  key: RoleKey
  labelKey: string
  descKey: string
  /** 该角色在通用模块之外额外开启的模块开关 key */
  modules: ModuleKey[]
}

/** 17 张创作方向卡片，按展示分组 */
export interface RoleGroup {
  titleKey: string
  roles: RoleDef[]
}

export const ROLE_GROUPS: RoleGroup[] = [
  {
    titleKey: 'setup.roles.group.tech',
    roles: [
      { key: 'tech', labelKey: 'setup.roles.tech', descKey: 'setup.roles.techDesc', modules: ['project_enabled', 'open_source_enabled', 'tech_stack_enabled'] },
      { key: 'digital', labelKey: 'setup.roles.digital', descKey: 'setup.roles.digitalDesc', modules: ['equipment_enabled', 'video_enabled'] },
    ],
  },
  {
    titleKey: 'setup.roles.group.media',
    roles: [
      { key: 'photo', labelKey: 'setup.roles.photo', descKey: 'setup.roles.photoDesc', modules: ['portfolio_enabled', 'equipment_enabled'] },
      { key: 'video', labelKey: 'setup.roles.video', descKey: 'setup.roles.videoDesc', modules: ['video_enabled'] },
      { key: 'music', labelKey: 'setup.roles.music', descKey: 'setup.roles.musicDesc', modules: ['music_enabled'] },
      { key: 'travelvlog', labelKey: 'setup.roles.travelvlog', descKey: 'setup.roles.travelvlogDesc', modules: ['travel_enabled', 'video_enabled', 'equipment_enabled'] },
    ],
  },
  {
    titleKey: 'setup.roles.group.life',
    roles: [
      { key: 'travel', labelKey: 'setup.roles.travel', descKey: 'setup.roles.travelDesc', modules: ['travel_enabled', 'equipment_enabled'] },
      { key: 'food', labelKey: 'setup.roles.food', descKey: 'setup.roles.foodDesc', modules: ['recipe_enabled'] },
      { key: 'fitness', labelKey: 'setup.roles.fitness', descKey: 'setup.roles.fitnessDesc', modules: ['fitness_enabled', 'equipment_enabled'] },
      { key: 'fashion', labelKey: 'setup.roles.fashion', descKey: 'setup.roles.fashionDesc', modules: ['portfolio_enabled'] },
      { key: 'pet', labelKey: 'setup.roles.pet', descKey: 'setup.roles.petDesc', modules: ['portfolio_enabled'] },
    ],
  },
  {
    titleKey: 'setup.roles.group.hobby',
    roles: [
      { key: 'reading', labelKey: 'setup.roles.reading', descKey: 'setup.roles.readingDesc', modules: ['book_enabled'] },
      { key: 'gaming', labelKey: 'setup.roles.gaming', descKey: 'setup.roles.gamingDesc', modules: ['game_enabled', 'video_enabled'] },
      { key: 'designer', labelKey: 'setup.roles.designer', descKey: 'setup.roles.designerDesc', modules: ['portfolio_enabled', 'project_enabled'] },
      { key: 'craft', labelKey: 'setup.roles.craft', descKey: 'setup.roles.craftDesc', modules: ['portfolio_enabled'] },
    ],
  },
]

/** 不参与分组展示的两个特殊角色 */
export const SPECIAL_ROLES: RoleDef[] = [
  { key: 'lifestyle', labelKey: 'setup.roles.lifestyle', descKey: 'setup.roles.lifestyleDesc', modules: [] },
  { key: 'all', labelKey: 'setup.roles.all', descKey: 'setup.roles.allDesc', modules: ['music_enabled', 'video_enabled', 'travel_enabled', 'portfolio_enabled', 'equipment_enabled', 'project_enabled', 'open_source_enabled', 'recipe_enabled', 'book_enabled', 'game_enabled', 'fitness_enabled', 'tech_stack_enabled'] },
]

/** 全部角色（含特殊角色）扁平列表 */
export const ALL_ROLES: RoleDef[] = [...ROLE_GROUPS.flatMap((g) => g.roles), ...SPECIAL_ROLES]

/**
 * 根据所选角色集合计算最终模块开关（通用模块恒开）：
 * 所选角色预置模块取并集，未选任何角色时全部开启（与现状一致）。
 * 媒体库为基础设施、恒可用，不在模块开关中。
 */
export function computeModulePreset(selectedRoles: RoleKey[]): Record<ModuleKey, boolean> {
  const preset = {
    article_enabled: true,
    music_enabled: false,
    video_enabled: false,
    travel_enabled: false,
    portfolio_enabled: false,
    equipment_enabled: false,
    project_enabled: false,
    open_source_enabled: false,
    recipe_enabled: false,
    book_enabled: false,
    game_enabled: false,
    fitness_enabled: false,
    tech_stack_enabled: false,
  } as Record<ModuleKey, boolean>

  if (selectedRoles.length === 0) {
    (Object.keys(preset) as ModuleKey[]).forEach((key) => {
      preset[key] = true
    })
    return preset
  }

  for (const roleKey of selectedRoles) {
    const role = ALL_ROLES.find((r) => r.key === roleKey)
    if (!role) continue
    for (const moduleKey of role.modules) {
      preset[moduleKey] = true
    }
  }
  return preset
}
