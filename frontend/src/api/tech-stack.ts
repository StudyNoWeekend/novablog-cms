import request from './request'
import type { TechStackItem, TechStackListRes, CreateTechStackReq, UpdateTechStackReq } from '@/types/tech-stack'

export const techStackApi = {
  getList(params?: { page?: number; page_size?: number; keyword?: string; category?: string; level?: number; status?: number }) {
    return request.get<TechStackListRes>('/tech-stacks', { params })
  },
  getById(id: string) {
    return request.get<TechStackItem>(`/tech-stacks/${id}`)
  },
  create(data: CreateTechStackReq) {
    return request.post<TechStackItem>('/tech-stacks', data)
  },
  update(id: string, data: UpdateTechStackReq) {
    return request.put<TechStackItem>(`/tech-stacks/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/tech-stacks/${id}`)
  },
}
