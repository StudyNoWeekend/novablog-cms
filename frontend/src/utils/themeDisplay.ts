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
