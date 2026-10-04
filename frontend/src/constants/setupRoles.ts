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

/** 角色卡片元数据（icon 为 Ant Design 图标组件名，向导渲染时解析） */
export interface RoleDef {
  key: RoleKey
  label: string
  description: string
  /** 该角色在通用模块之外额外开启的模块开关 key */
  modules: ModuleKey[]
}

/** 17 张创作方向卡片，按展示分组 */
export interface RoleGroup {
  title: string
  roles: RoleDef[]
}

export const ROLE_GROUPS: RoleGroup[] = [
  {
    title: '技术 · 数码',
    roles: [
      { key: 'tech', label: '技术开发者', description: '写技术文章，展示项目与开源作品', modules: ['project_enabled', 'open_source_enabled', 'tech_stack_enabled'] },
      { key: 'digital', label: '数码测评', description: '装备测评与上手视频', modules: ['equipment_enabled', 'video_enabled'] },
    ],
  },
  {
    title: '影像 · 声音',
    roles: [
      { key: 'photo', label: '摄影爱好者', description: '摄影作品集与器材清单', modules: ['portfolio_enabled', 'equipment_enabled'] },
      { key: 'video', label: '视频创作者', description: '视频作品与平台链接汇总', modules: ['video_enabled'] },
      { key: 'music', label: '音乐人', description: '歌单与音乐作品分享', modules: ['music_enabled'] },
      { key: 'travelvlog', label: '旅行 Vlogger', description: '旅行 vlog、攻略与装备', modules: ['travel_enabled', 'video_enabled', 'equipment_enabled'] },
    ],
  },
  {
    title: '旅行 · 生活',
    roles: [
      { key: 'travel', label: '旅行博主', description: '旅行攻略与出行装备', modules: ['travel_enabled', 'equipment_enabled'] },
      { key: 'food', label: '美食爱好者', description: '菜谱、食材与烹饪步骤', modules: ['recipe_enabled'] },
      { key: 'fitness', label: '健身运动', description: '训练记录与健身装备', modules: ['fitness_enabled', 'equipment_enabled'] },
      { key: 'fashion', label: '时尚穿搭', description: '穿搭 lookbook 作品集', modules: ['portfolio_enabled'] },
      { key: 'pet', label: '宠物日常', description: '毛孩子相册与作品集', modules: ['portfolio_enabled'] },
    ],
  },
  {
    title: '兴趣 · 收藏',
    roles: [
      { key: 'reading', label: '读书', description: '读书书架、评分与书评', modules: ['book_enabled'] },
      { key: 'gaming', label: '游戏玩家', description: '游戏库、时长与短评', modules: ['game_enabled', 'video_enabled'] },
      { key: 'designer', label: '设计师 / 插画师', description: '作品集与项目经历', modules: ['portfolio_enabled', 'project_enabled'] },
      { key: 'craft', label: '手工 DIY', description: '手工作品图集', modules: ['portfolio_enabled'] },
    ],
  },
]

/** 不参与分组展示的两个特殊角色 */
export const SPECIAL_ROLES: RoleDef[] = [
  { key: 'lifestyle', label: '生活方式', description: '以文章与媒体库记录生活，暂不开启专属模块', modules: [] },
  { key: 'all', label: '全能创作者', description: '开启全部内容模块，之后随时可调', modules: ['music_enabled', 'video_enabled', 'travel_enabled', 'portfolio_enabled', 'equipment_enabled', 'project_enabled', 'open_source_enabled', 'recipe_enabled', 'book_enabled', 'game_enabled', 'fitness_enabled', 'tech_stack_enabled'] },
]

/** 全部角色（含特殊角色）扁平列表 */
export const ALL_ROLES: RoleDef[] = [...ROLE_GROUPS.flatMap((g) => g.roles), ...SPECIAL_ROLES]

/**
 * 根据所选角色集合计算最终模块开关（通用模块恒开）：
 * 所选角色预置模块取并集，未选任何角色时全部开启（与现状一致）。
 */
export function computeModulePreset(selectedRoles: RoleKey[]): Record<ModuleKey, boolean> {
  const preset = {
    article_enabled: true,
    media_enabled: true,
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
