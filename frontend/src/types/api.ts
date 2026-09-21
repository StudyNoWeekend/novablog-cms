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
  social_links?: SocialLink[]
  tags?: string[]
}

export interface UpdateProfileReq {
  nickname?: string
  avatar?: string
  bio?: string
  email?: string
  city?: string
  page_background?: string
  blog_icon?: string
  blog_title?: string
  blog_description?: string
  social_links?: SocialLink[]
  tags?: string[]
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