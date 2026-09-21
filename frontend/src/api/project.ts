import request from './request'
import type { Project, ProjectListRes, CreateProjectReq, UpdateProjectReq } from '@/types/project'

export const projectApi = {
  getList(params?: {
    page?: number
    page_size?: number
    keyword?: string
    category?: string
    status?: number
  }) {
    return request.get<ProjectListRes>('/projects', { params })
  },
  getById(id: string) {
    return request.get<Project>(`/projects/${id}`)
  },
  create(data: CreateProjectReq) {
    return request.post<Project>('/projects', data)
  },
  update(id: string, data: UpdateProjectReq) {
    return request.put<Project>(`/projects/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/projects/${id}`)
  },
}
