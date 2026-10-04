import request from './request'
import type { MediaItem } from './media'

/** 图源类型：icons=品牌图标(Simple Icons)、games=游戏封面(Bangumi)、books=书籍封面(豆瓣)、food=美食图片(Bing) */
export type ImageSearchType = 'icons' | 'games' | 'books' | 'food'

// ImageResult 统一图片搜索结果。
export interface ImageResult {
  /** 名称（品牌图标为图标名、条目为条目名；无名称元数据时为空串） */
  name: string
  /** 图片直链 */
  url: string
}

export const imageSearchApi = {
  search(type: ImageSearchType, q: string, limit?: number) {
    return request.get<ImageResult[]>('/image-search', { params: { type, q, limit } })
  },
  /** 转存外部图片为本站资源（豆瓣图床有 Cookie 反爬、第三方站点图片有防盗链，直链可能失效）；module 为归属媒体库模块文件夹 key */
  save(url: string, module?: string) {
    return request.post<MediaItem>('/image-search/save', { url, module: module || undefined })
  },
}
