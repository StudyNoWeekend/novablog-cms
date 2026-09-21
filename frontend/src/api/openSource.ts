import request from './request'
import type {
  OpenSourceWork,
  OpenSourceListRes,
  CreateOpenSourceReq,
  UpdateOpenSourceReq,
} from '@/types/openSource'

export const openSourceApi = {
  getList(params?: { page?: number; page_size?: number; keyword?: string; status?: number }) {
    return request.get<OpenSourceListRes>('/open-sources', { params })
  },
  getById(id: string) {
    return request.get<OpenSourceWork>(`/open-sources/${id}`)
  },
  create(data: CreateOpenSourceReq) {
    return request.post<OpenSourceWork>('/open-sources', data)
  },
  update(id: string, data: UpdateOpenSourceReq) {
    return request.put<OpenSourceWork>(`/open-sources/${id}`, data)
  },
  /** 重新拉取远端仓库元数据与 README */
  refresh(id: string) {
    return request.post<OpenSourceWork>(`/open-sources/${id}/refresh`)
  },
  remove(id: string) {
    return request.delete(`/open-sources/${id}`)
  },
}
