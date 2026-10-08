import type { RoleKey } from '@/constants/setupRoles'

export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
  trace_id: string
}

/** 应用版本信息（后端构建期 -ldflags 注入） */
export interface AppVersionInfo {
  version: string
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface LoginReq {
  username: string
  password: string
}

export interface LoginRes {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface SocialLink {
  platform: string
  url: string
  sort_order: number
}

/** 星座选项元数据（image 为 24×24 viewBox 内的自包含 SVG 片段，配合 v-html 渲染） */
export interface ZodiacMeta {
  key: string
  name: string
  image: string
  date_range: string
  element: string
}

/** 性格选项元数据（MBTI 16 型，image 为四字母徽章 SVG 片段） */
export interface PersonalityMeta {
  key: string
  name: string
  image: string
  description: string
}

/** 个人资料选项元数据（星座 + 性格全量选项） */
export interface ProfileMeta {
  zodiac: ZodiacMeta[]
  personality: PersonalityMeta[]
}

export interface UserInfo {
  id: string
  username: string
  nickname: string
  avatar: string
  bio: string
  page_background: string
  blog_icon: string
  blog_title: string
  blog_description: string
  email: string
  city: string
  /** 性格（MBTI key，如 INTJ）；空串表示未设置 */
  personality?: string
  /** 星座 key（如 aries）；空串表示未设置 */
  zodiac?: string
  /** 邮箱是否对外展示 */
  show_email?: boolean
  /** 城市是否对外展示 */
  show_city?: boolean
  /** 星座是否对外展示 */
  show_zodiac?: boolean
  /** 性格是否对外展示 */
  show_personality?: boolean
  /** 创作方向角色 key，多选逗号分隔；空串表示未选过（旧版本用户/首装跳过），需弹窗补选 */
  role?: string
  social_links?: SocialLink[]
  tags?: string[]
}

export interface UpdateProfileReq {
  nickname?: string
  avatar?: string
  bio?: string
  email?: string
  city?: string
  personality?: string
  zodiac?: string
  show_email?: boolean
  show_city?: boolean
  show_zodiac?: boolean
  show_personality?: boolean
  page_background?: string
  blog_icon?: string
  blog_title?: string
  blog_description?: string
  social_links?: SocialLink[]
  tags?: string[]
}

/** 补选创作方向请求（旧版本用户未选爱好的兼容入口） */
export interface UpdateRolesReq {
  /** 角色 key 数组，与后端 rolePresetModules 白名单一致 */
  roles: RoleKey[]
  /** 最终模块开关全量 map（computeModulePreset 产物）；不传时后端按角色预设并集兜底 */
  modules?: Record<string, boolean>
}

// ---- 跨域配置 ----

export interface CorsConfigRes {
  allowed_origins: string
  updated_at: string
}

export interface UpdateCorsConfigReq {
  allowed_origins?: string
}

// ---- 官方主题市场配置 ----

/** 主题模块运行配置（管理端：官方市场地址 + 博客公开 API 地址） */
export interface ThemeMarketConfigRes {
  market_base_url: string
  public_api_base: string
  updated_at: string
}

export interface UpdateThemeMarketConfigReq {
  market_base_url: string
  /** 博客公开 API 地址；空串/缺省 = 清除，主题回退同域相对路径取数 */
  public_api_base?: string
}

/** 公共配置下发（免鉴权，默认值由后端控制） */
export interface PublicConfigRes {
  market_base_url: string
}

// ---- 地图服务配置 ----

/** 地图服务配置（管理端；安全密钥返回解密后明文，供选点器注入 _AMapSecurityConfig） */
export interface MapConfigRes {
  amap_key: string
  amap_security_code: string
  google_key: string
  updated_at: string
}

export interface UpdateMapConfigReq {
  amap_key: string
  /** 留空 = 保持后端原值（后端加密存储） */
  amap_security_code: string
  google_key: string
}

export interface FrameConfig {
  template: 'gallery' | 'movie' | 'floating'
  fontScale: number
  borderScale: number
  borderColor: string
  textColor: 'auto' | 'black' | 'white'
  fontFamily: string
  logoMode: 'text' | 'image'
  showExif: boolean
}

export interface DisplayParams {
  make: string
  model: string
  lens: string
  focalLength: string
  aperture: string
  shutter: string
  iso: string
  date: string
}

export interface ExifInfo {
  filename: string
  width: number
  height: number
  size: number
  make: string
  model: string
  lens: string
  software: string
  focalLength: string
  aperture: string
  shutter: string
  iso: string
}

export interface MediaPreset {
  id: string
  media_id: string
  name: string
  frame_config: string
  display_params: string
  output_url: string
  output_storage_path: string
  output_size: number
  mime_type: string
  created_at: string
}

export interface MediaPresetListRes {
  list: MediaPreset[]
}