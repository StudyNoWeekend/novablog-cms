import request from './request'
import type {
  McpKeyItem,
  CreateMcpKeyReq,
  CreateMcpKeyRes,
  UpdateMcpKeyReq,
  McpKeyQuery,
} from '@/types/mcp'
import type { PaginatedData } from '@/types/api'

export const getMcpKeysAPI = (params: McpKeyQuery) =>
  request.get<PaginatedData<McpKeyItem>>('/mcp/keys', { params })

export const createMcpKeyAPI = (data: CreateMcpKeyReq) =>
  request.post<CreateMcpKeyRes>('/mcp/keys', data)

export const updateMcpKeyAPI = (id: string, data: UpdateMcpKeyReq) =>
  request.put<McpKeyItem>(`/mcp/keys/${id}`, data)

export const deleteMcpKeyAPI = (id: string) => request.delete(`/mcp/keys/${id}`)
