import request from './request'
import type { MediaPreset, MediaPresetListRes } from '@/types/api'

export interface MediaItem {
  id: string
  filename: string
  file_type: number
  mime_type: string
  size: number
  url: string
  thumb_url: string
  width: number
  height: number
  created_at: string
}

export interface MediaListRes {
  list: MediaItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface UploadWithPresetRes {
  media: MediaItem
  preset: MediaPreset
}

export interface MediaUsageItem {
  id: string
  title: string
  field: string
}

export interface MediaUsageGroup {
  module: string
  items: MediaUsageItem[]
}

export interface MediaUsageRes {
  media_id: string
  used: boolean
  total: number
  groups: MediaUsageGroup[]
}

export const mediaApi = {
  upload(file: File, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<MediaItem>('/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  uploadWithPreset(file: File, name?: string, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    if (name) {
      formData.append('name', name)
    }
    return request.post<UploadWithPresetRes>('/media/upload-with-preset', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  getList(params?: { file_type?: number; keyword?: string; page?: number; page_size?: number }) {
    return request.get<MediaListRes>('/media', { params })
  },
  getUsages(id: string) {
    return request.get<MediaUsageRes>(`/media/${id}/usages`)
  },
  remove(id: string, force = false) {
    return request.delete(`/media/${id}`, { params: force ? { force: 'true' } : undefined })
  },
  getPresets(mediaId: string) {
    return request.get<MediaPresetListRes>(`/media/${mediaId}/presets`)
  },
  createPreset(payload: {
    media_id: string
    name: string
    frame_config: string
    display_params: string
    file: Blob
  }, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('media_id', payload.media_id)
    formData.append('name', payload.name)
    formData.append('frame_config', payload.frame_config)
    formData.append('display_params', payload.display_params)
    formData.append('file', payload.file, `${payload.name}.jpg`)
    return request.post<MediaPreset>('/media/preset', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  deletePreset(id: string) {
    return request.delete(`/media/preset/${id}`)
  },
}