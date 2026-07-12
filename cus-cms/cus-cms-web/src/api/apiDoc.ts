import request from './request'
import type { APIDocItem } from '@/types/security'

export const getAPIDocsAPI = () => request.get<APIDocItem[]>('/api-docs')
