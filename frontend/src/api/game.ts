import request from './request'
import type { Game, GameListRes, CreateGameReq, UpdateGameReq } from '@/types/game'

export const gameApi = {
  getList(params?: { page?: number; page_size?: number; keyword?: string; platform?: string; genre?: string; play_status?: string; status?: number }) {
    return request.get<GameListRes>('/games', { params })
  },
  getById(id: string) {
    return request.get<Game>(`/games/${id}`)
  },
  create(data: CreateGameReq) {
    return request.post<Game>('/games', data)
  },
  update(id: string, data: UpdateGameReq) {
    return request.put<Game>(`/games/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/games/${id}`)
  },
}
