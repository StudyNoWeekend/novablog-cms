export interface ModuleConfig {
  article_enabled: boolean
  media_enabled: boolean
  music_enabled: boolean
  video_enabled: boolean
  travel_enabled: boolean
  portfolio_enabled: boolean
  equipment_enabled: boolean
  project_enabled: boolean
  updated_at: string
}

export interface UpdateModuleConfigReq {
  article_enabled?: boolean
  media_enabled?: boolean
  music_enabled?: boolean
  video_enabled?: boolean
  travel_enabled?: boolean
  portfolio_enabled?: boolean
  equipment_enabled?: boolean
  project_enabled?: boolean
}
