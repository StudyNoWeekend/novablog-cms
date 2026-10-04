import request from './request'
import type { Recipe, RecipeListRes, CreateRecipeReq, UpdateRecipeReq } from '@/types/recipe'

export const recipeApi = {
  getList(params?: { page?: number; page_size?: number; keyword?: string; difficulty?: number; status?: number }) {
    return request.get<RecipeListRes>('/recipes', { params })
  },
  getById(id: string) {
    return request.get<Recipe>(`/recipes/${id}`)
  },
  create(data: CreateRecipeReq) {
    return request.post<Recipe>('/recipes', data)
  },
  update(id: string, data: UpdateRecipeReq) {
    return request.put<Recipe>(`/recipes/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/recipes/${id}`)
  },
}
