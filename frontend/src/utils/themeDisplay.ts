import type { ThemeItem } from '@/types/theme'

// THEME_TYPE_LABELS 主题类型取值与展示文案（与官方接口文档对齐）
export const THEME_TYPE_LABELS: Record<string, string> = {
  personal: '个人博客',
  tech: '技术博客',
  photo: '摄影博客',
  travel: '旅行博客',
  music: '音乐博客',
  team: '企业/团队',
}

// THEME_TYPE_GRADIENTS 各主题类型封面占位渐变色
const THEME_TYPE_GRADIENTS: Record<string, string> = {
  personal: 'linear-gradient(135deg, #6a11cb 0%, #2575fc 100%)',
  tech: 'linear-gradient(135deg, #0f2027 0%, #2c5364 100%)',
  photo: 'linear-gradient(135deg, #355c7d 0%, #c06c84 100%)',
  travel: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)',
  music: 'linear-gradient(135deg, #fc466b 0%, #3f5efb 100%)',
  team: 'linear-gradient(135deg, #f7971e 0%, #ffd200 100%)',
}

// THEME_DISPLAY_NAMES 主题 id → 规范展示名（与 novablog-web 各主题包 theme.json 的 name 对齐）。
// 官方市场上架数据曾把 title 误填为主题 id（如 melody-notes），展示层据此映射回中文名；未知 id 回退原 title。
const THEME_DISPLAY_NAMES: Record<string, string> = {
  'melody-notes': '旋律笔记',
  'game-diary': '游戏日记',
  'food-diary': '美食日记',
  'wanderlust-template': '漫游世界',
  'tech-geek': '极客风',
  'photo-creator': '光影集',
  'media-creator': '帧记影像',
  'lens-life-template': '镜头生活',
}

// themeDisplayName 解析市场主题的展示名：优先按 slug（主题包 id）映射，其次按 title 匹配，均未命中时回退 title。
export function themeDisplayName(title: string, slug?: string): string {
  const key = slug || ''
  return THEME_DISPLAY_NAMES[key] || THEME_DISPLAY_NAMES[title] || title
}

// hasNewVersion 判断官方市场版本与本地已安装版本是否不同（不同即视为有新版本可更新）。
export function hasNewVersion(localVersion: string, marketVersion?: string): boolean {
  return !!marketVersion && marketVersion !== localVersion
}

// themeGradient 获取主题类型对应的封面渐变背景，未知类型走默认色。
export function themeGradient(type: string): string {
  return THEME_TYPE_GRADIENTS[type] || 'linear-gradient(135deg, #4b6cb7 0%, #182848 100%)'
}

// isPreviewURL 判断主题 preview 字段是否为可展示的图片地址。
export function isPreviewURL(preview: string): boolean {
  if (!preview) return false
  return /^https?:\/\//.test(preview) || preview.startsWith('/')
}

// themeStatusText 主题状态文案（1 审核中 / 2 已上架 / 3 已下架）。
export function themeStatusText(status: number): string {
  const map: Record<number, string> = { 1: '审核中', 2: '已上架', 3: '已下架' }
  return map[status] || '未知'
}
