import request from './request'
import type { Category, CategoryCreateReq, CategoryUpdateReq } from '@/types/category'

export const categoryApi = {
  getList(type?: string) {
    return request.get<Category[]>('/categories', { params: type ? { type } : {} })
  },
  create(data: CategoryCreateReq) {
    return request.post<Category>('/categories', data)
  },
  update(id: string, data: CategoryUpdateReq) {
    return request.put<Category>(`/categories/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/categories/${id}`)
  },
}