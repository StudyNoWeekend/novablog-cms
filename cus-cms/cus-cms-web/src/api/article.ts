import request from './request'
import type { Article, ArticleCreateReq, ArticleUpdateReq, ArticleFilters } from '@/types/article'
import type { PaginatedData } from '@/types/api'

export const articleApi = {
  getList(params: { page?: number; page_size?: number } & ArticleFilters) {
    return request.get<PaginatedData<Article>>('/articles', { params })
  },
  getDetail(id: string) {
    return request.get<Article>(`/articles/${id}`)
  },
  create(data: ArticleCreateReq) {
    return request.post<Article>('/articles', data)
  },
  update(id: string, data: ArticleUpdateReq) {
    return request.put<Article>(`/articles/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/articles/${id}`)
  },
  updateStatus(id: string, status: number) {
    return request.put(`/articles/${id}/status`, { status })
  },
}