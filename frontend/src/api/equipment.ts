import request from './request'
import type { Equipment, EquipmentListRes, CreateEquipmentReq, UpdateEquipmentReq } from '@/types/equipment'

export const equipmentApi = {
  getList(params?: {
    page?: number
    page_size?: number
    keyword?: string
    brand?: string
  }) {
    return request.get<EquipmentListRes>('/equipments', { params })
  },
  getById(id: string) {
    return request.get<Equipment>(`/equipments/${id}`)
  },
  create(data: CreateEquipmentReq) {
    return request.post<Equipment>('/equipments', data)
  },
  update(id: string, data: UpdateEquipmentReq) {
    return request.put<Equipment>(`/equipments/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/equipments/${id}`)
  },
}
