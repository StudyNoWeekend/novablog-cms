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
  folder_id: string | null
  folder_name: string
  created_at: string
}

export interface MediaListRes {
  list: MediaItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface MediaFolderNode {
  id: string
  name: string
  parent_id: string | null
  module_key: string | null
  media_count: number
  children: MediaFolderNode[]
  created_at: string
}

export interface MediaFolderTreeRes {
  list: MediaFolderNode[]
}

/** 上传归属选项：module=业务模块 key（归入对应模块文件夹），folder_id=显式目标文件夹（优先） */
export interface MediaUploadOptions {
  module?: string
  folder_id?: string
  onProgress?: (percent: number) => void
}

export interface MediaListParams {
  file_type?: number
  keyword?: string
  /** 'root'=根目录；文件夹 ID=指定文件夹；不传=全部（全局搜索时用） */
  folder_id?: string
  page?: number
  page_size?: number
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
  upload(file: File, opts: MediaUploadOptions = {}) {
    const formData = new FormData()
    formData.append('file', file)
    if (opts.module) {
      formData.append('module', opts.module)
    }
    if (opts.folder_id) {
      formData.append('folder_id', opts.folder_id)
    }
    return request.post<MediaItem>('/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && opts.onProgress) {
          opts.onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  uploadWithPreset(file: File, name?: string, opts: MediaUploadOptions = {}) {
    const formData = new FormData()
    formData.append('file', file)
    if (name) {
      formData.append('name', name)
    }
    if (opts.module) {
      formData.append('module', opts.module)
    }
    if (opts.folder_id) {
      formData.append('folder_id', opts.folder_id)
    }
    return request.post<UploadWithPresetRes>('/media/upload-with-preset', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && opts.onProgress) {
          opts.onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },
  getList(params?: MediaListParams) {
    return request.get<MediaListRes>('/media', { params })
  },
  getFolderTree() {
    return request.get<MediaFolderTreeRes>('/media/folders/tree')
  },
  createFolder(data: { name: string; parent_id?: string | null }) {
    return request.post<MediaFolderNode>('/media/folders', data)
  },
  updateFolder(id: string, data: { name?: string; /** 空字符串=移动到根目录；不传=不变更 */ parent_id?: string }) {
    return request.put<MediaFolderNode>(`/media/folders/${id}`, data)
  },
  deleteFolder(id: string) {
    return request.delete(`/media/folders/${id}`)
  },
  moveMedia(mediaIds: string[], folderId?: string | null) {
    return request.put<void>('/media/move', { media_ids: mediaIds, folder_id: folderId ?? null })
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
