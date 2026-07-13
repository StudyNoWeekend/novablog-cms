import request from './request'
import type { VideoWork, VideoListRes, CreateVideoReq, UpdateVideoReq, ParseVideoReq, ParseVideoRes } from '@/types/video'

export const videoApi = {
  getList(params?: {
    page?: number
    page_size?: number
    keyword?: string
    status?: number
  }) {
    return request.get<VideoListRes>('/videos', { params })
  },
  getById(id: string) {
    return request.get<VideoWork>(`/videos/${id}`)
  },
  create(data: CreateVideoReq) {
    return request.post<VideoWork>('/videos', data)
  },
  update(id: string, data: UpdateVideoReq) {
    return request.put<VideoWork>(`/videos/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/videos/${id}`)
  },
  parse(data: ParseVideoReq) {
    return request.post<ParseVideoRes>('/videos/parse', data)
  },
}
