export interface ModuleConfig {
  article_enabled: boolean
  music_enabled: boolean
  video_enabled: boolean
  travel_enabled: boolean
  portfolio_enabled: boolean
  equipment_enabled: boolean
  project_enabled: boolean
  open_source_enabled: boolean
  recipe_enabled: boolean
  book_enabled: boolean
  game_enabled: boolean
  fitness_enabled: boolean
  tech_stack_enabled: boolean
  updated_at: string
}

export interface UpdateModuleConfigReq {
  article_enabled?: boolean
  music_enabled?: boolean
  video_enabled?: boolean
  travel_enabled?: boolean
  portfolio_enabled?: boolean
  equipment_enabled?: boolean
  project_enabled?: boolean
  open_source_enabled?: boolean
  recipe_enabled?: boolean
  book_enabled?: boolean
  game_enabled?: boolean
  fitness_enabled?: boolean
  tech_stack_enabled?: boolean
}
