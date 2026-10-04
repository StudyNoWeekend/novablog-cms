import request from './request'
import type { FitnessRecord, FitnessListRes, CreateFitnessReq, UpdateFitnessReq } from '@/types/fitness'

export const fitnessApi = {
  getList(params?: { page?: number; page_size?: number; keyword?: string; type?: string; status?: number }) {
    return request.get<FitnessListRes>('/fitness', { params })
  },
  getById(id: string) {
    return request.get<FitnessRecord>(`/fitness/${id}`)
  },
  create(data: CreateFitnessReq) {
    return request.post<FitnessRecord>('/fitness', data)
  },
  update(id: string, data: UpdateFitnessReq) {
    return request.put<FitnessRecord>(`/fitness/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/fitness/${id}`)
  },
}
