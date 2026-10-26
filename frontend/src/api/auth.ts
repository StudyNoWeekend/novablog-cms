import request from './request'
import type { LoginReq, LoginRes, UserInfo, UpdateProfileReq, UpdateRolesReq, ProfileMeta } from '@/types/api'

/** 资料图片上传响应：url 用于展示，media_id 为媒体库文件 ID（保存资料时随表单提交） */
export interface ProfileUploadRes {
  url: string
  media_id: string
}

export const authApi = {
  login(data: LoginReq) { return request.post<LoginRes>('/auth/login', data) },
  refresh(refreshToken: string) { return request.post<LoginRes>('/auth/refresh', { refresh_token: refreshToken }) },
  logout() { return request.post('/auth/logout') },
  changePassword(data: { old_password: string; new_password: string }) { return request.put('/auth/password', data) },
  getProfile() { return request.get<UserInfo>('/profile') },
  updateProfile(data: UpdateProfileReq) { return request.put<UserInfo>('/profile', data) },
  /** 个人资料选项元数据（星座 + 性格全量选项，公开接口） */
  getProfileMeta() { return request.get<ProfileMeta>('/public/profile-meta') },
  /** 补选创作方向（旧版本用户未选爱好时的兼容入口），返回更新后的用户资料 */
  updateRoles(data: UpdateRolesReq) { return request.put<UserInfo>('/profile/roles', data) },
  uploadIcon(file: File, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<ProfileUploadRes>('/profile/upload-icon', formData, {
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
    return request.post<ProfileUploadRes>('/profile/upload-background', formData, {
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
    return request.post<ProfileUploadRes>('/profile/upload-avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
}