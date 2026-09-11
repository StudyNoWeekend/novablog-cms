import request from './request'
import type {
  SecurityConfig,
  UpdateSecurityConfigReq,
  BlacklistItem,
  BlacklistQuery,
  CreateBlacklistReq,
  UpdateBlacklistReq,
  IPAccessStats,
  AccessStatsQuery,
} from '@/types/security'
import type { PaginatedData } from '@/types/api'

export const getSecurityConfigAPI = () => request.get<SecurityConfig>('/security/config')

export const updateSecurityConfigAPI = (data: UpdateSecurityConfigReq) =>
  request.put('/security/config', data)

export const getAccessStatsAPI = (params: AccessStatsQuery) =>
  request.get<PaginatedData<IPAccessStats>>('/security/access-stats', { params })

export const getBlacklistsAPI = (params: BlacklistQuery) =>
  request.get<PaginatedData<BlacklistItem>>('/security/blacklists', { params })

export const getBlacklistAPI = (id: string | number) =>
  request.get<BlacklistItem>(`/security/blacklist/${id}`)

export const createBlacklistAPI = (data: CreateBlacklistReq) =>
  request.post('/security/blacklist', data)

export const updateBlacklistAPI = (id: string | number, data: UpdateBlacklistReq) =>
  request.put(`/security/blacklist/${id}`, data)

export const deleteBlacklistAPI = (id: string | number) =>
  request.delete(`/security/blacklist/${id}`)
