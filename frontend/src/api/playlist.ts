import request from './request'
import type { PaginatedData } from '@/types/api'
import type { ThirdPartyPlaylist, CreatePlaylistReq, UpdatePlaylistReq } from '@/types/playlist'

export const playlistApi = {
  getList(params?: { page?: number; page_size?: number }) {
    return request.get<PaginatedData<ThirdPartyPlaylist>>('/playlists', { params })
  },
  getByID(id: string) {
    return request.get<ThirdPartyPlaylist>(`/playlists/${id}`)
  },
  create(data: CreatePlaylistReq) {
    return request.post<ThirdPartyPlaylist>('/playlists', data)
  },
  update(id: string, data: UpdatePlaylistReq) {
    return request.put<ThirdPartyPlaylist>(`/playlists/${id}`, data)
  },
  delete(id: string) {
    return request.delete(`/playlists/${id}`)
  },
}
