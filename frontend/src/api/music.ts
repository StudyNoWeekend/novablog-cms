import request from './request'
import type { PaginatedData } from '@/types/api'
import type {
  Song,
  SongCreateReq,
  SongUpdateReq,
  ParseTask,
  BatchCreateSongReq,
} from '@/types/music'

export const musicApi = {
  // 歌曲
  getSongs(params?: { category_id?: string; page?: number; page_size?: number }) {
    return request.get<PaginatedData<Song>>('/music/songs', { params })
  },
  createSong(data: SongCreateReq) {
    return request.post<Song>('/music/songs', data)
  },
  batchCreateSongs(data: BatchCreateSongReq) {
    return request.post<Song[]>('/music/songs/batch', data)
  },
  updateSong(id: string, data: SongUpdateReq) {
    return request.put<Song>(`/music/songs/${id}`, data)
  },
  deleteSong(id: string) {
    return request.delete(`/music/songs/${id}`)
  },

  // 解析
  parse(url: string) {
    return request.post<{ task_id: string }>('/music/parse', { url })
  },
  getParseStatus(taskId: string) {
    return request.get<ParseTask>(`/music/parse/${taskId}`)
  },

  // 获取播放地址（B站官方外链播放器地址，供 iframe 内嵌播放）
  getAudioUrl(songId: string) {
    return request.get<{ url: string }>(`/music/audio-url/${songId}`)
  },
}
