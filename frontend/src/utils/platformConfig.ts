// 第三方平台配置 - 品牌色和图标
export interface PlatformConfig {
  id: string
  label: string
  color: string
  icon: string // SVG path or text
}

export const PLATFORM_CONFIGS: Record<string, PlatformConfig> = {
  qq_music: {
    id: 'qq_music',
    label: 'QQ音乐',
    color: '#00D4FF',
    icon: 'QQ',
  },
  netease: {
    id: 'netease',
    label: '网易云音乐',
    color: '#D43C33',
    icon: 'N',
  },
  bilibili: {
    id: 'bilibili',
    label: 'B站',
    color: '#FB7299',
    icon: 'B',
  },
  spotify: {
    id: 'spotify',
    label: 'Spotify',
    color: '#1DB954',
    icon: 'S',
  },
  apple_music: {
    id: 'apple_music',
    label: 'Apple Music',
    color: '#FA243C',
    icon: 'A',
  },
  other: {
    id: 'other',
    label: '其他平台',
    color: '#999999',
    icon: '♪',
  },
}

export function getPlatformConfig(platform: string): PlatformConfig {
  return PLATFORM_CONFIGS[platform] || PLATFORM_CONFIGS.other
}
