import request from './request'
import type { LoginReq, LoginRes, UserInfo } from '@/types/api'

export const authApi = {
  login(data: LoginReq) { return request.post<LoginRes>('/auth/login', data) },
  refresh(refreshToken: string) { return request.post<LoginRes>('/auth/refresh', { refresh_token: refreshToken }) },
  logout() { return request.post('/auth/logout') },
  changePassword(data: { old_password: string; new_password: string }) { return request.put('/auth/password', data) },
  getProfile() { return request.get<UserInfo>('/profile') },
}