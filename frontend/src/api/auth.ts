import request from './request'
import type { LoginReq, LoginRes, UserInfo, UpdateProfileReq, UpdateRolesReq } from '@/types/api'

export const authApi = {
  login(data: LoginReq) { return request.post<LoginRes>('/auth/login', data) },
  refresh(refreshToken: string) { return request.post<LoginRes>('/auth/refresh', { refresh_token: refreshToken }) },
  logout() { return request.post('/auth/logout') },
  changePassword(data: { old_password: string; new_password: string }) { return request.put('/auth/password', data) },
  getProfile() { return request.get<UserInfo>('/profile') },
  updateProfile(data: UpdateProfileReq) { return request.put<UserInfo>('/profile', data) },
  /** 补选创作方向（旧版本用户未选爱好时的兼容入口），返回更新后的用户资料 */
  updateRoles(data: UpdateRolesReq) { return request.put<UserInfo>('/profile/roles', data) },
  uploadIcon(file: File, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<{ url: string }>('/profile/upload-icon', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  uploadBackground(file: File, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<{ url: string }>('/profile/upload-background', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  uploadAvatar(file: File, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<{ url: string }>('/profile/upload-avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
}