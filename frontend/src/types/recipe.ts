export interface RecipeIngredient {
  name: string
  amount?: string
}

export interface RecipeStep {
  text: string
  image?: string
}

export interface Recipe {
  id: string
  title: string
  cover: string
  summary: string
  ingredients: RecipeIngredient[] | null
  steps: RecipeStep[] | null
  difficulty: number
  minutes: number
  servings: number
  tags: string
  status: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface RecipeListRes {
  list: Recipe[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateRecipeReq {
  title: string
  cover?: string
  summary?: string
  ingredients?: RecipeIngredient[]
  steps?: RecipeStep[]
  difficulty?: number
  minutes?: number
  servings?: number
  tags?: string
  status?: number
  sort_order?: number
}

export interface UpdateRecipeReq extends Partial<CreateRecipeReq> {}
