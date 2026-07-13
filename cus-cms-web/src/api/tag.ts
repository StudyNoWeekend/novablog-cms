import request from './request'
import type { Tag, TagCreateReq, TagUpdateReq } from '@/types/tag'

export const tagApi = {
  getList() {
    return request.get<Tag[]>('/tags')
  },
  create(data: TagCreateReq) {
    return request.post<Tag>('/tags', data)
  },
  update(id: string, data: TagUpdateReq) {
    return request.put<Tag>(`/tags/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/tags/${id}`)
  },
}