import request from './request'
import type { SecurityConfig, UpdateSecurityConfigReq, BlacklistItem, SecurityStats } from '@/types/security'
import type { PaginatedData } from '@/types/api'

export const getSecurityConfigAPI = () => request.get<SecurityConfig>('/security/config')

export const updateSecurityConfigAPI = (data: UpdateSecurityConfigReq) => request.put('/security/config', data)

export const getBlacklistAPI = (params: { page: number; page_size: number }) =>
  request.get<PaginatedData<BlacklistItem>>('/security/blacklist', { params })

export const unbanIPAPI = (ip: string) => request.delete(`/security/blacklist/${ip}`)

export const getSecurityStatsAPI = () => request.get<SecurityStats>('/security/stats')
