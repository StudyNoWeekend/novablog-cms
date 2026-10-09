import request from './request'
import type { TravelGuide, TravelGuideFormData, TravelAttraction, TravelItineraryDay } from '@/types/travel'
import type { PaginatedData } from '@/types/api'

export interface TravelGuideListParams {
  page?: number
  page_size?: number
  keyword?: string
  region?: string
  status?: number
  days_range?: string
  sort?: string
  category_id?: string
}

function toCreateReq(data: TravelGuideFormData) {
  return {
    title: data.title,
    summary: data.summary,
    cover_image: data.coverImage,
    status: data.status,
    destination: data.destination,
    region: data.region,
    // 空分类传 null（Go 侧指针为 nil 视为未变更），传 "" 会触发 Postgres uuid 解析错误
    category_id: data.categoryId || null,
    days: data.days,
    best_month: data.bestMonth,
    attractions: data.attractions,
    itinerary: data.itinerary,
    reviews: data.reviews,
  }
}

function toUpdateReq(data: TravelGuideFormData) {
  return toCreateReq(data)
}

// 景点内容指纹：用于给缺 id 的老数据生成稳定 id（内容不变则 id 跨刷新不变）
function attractionKey(a: any): string {
  return [a?.name ?? '', a?.location ?? '', a?.latitude ?? '', a?.longitude ?? ''].join('|')
}

// 保证每个景点有唯一且稳定的 id：老数据（无 id）按内容哈希生成 nb-xxx 形式 id
function ensureAttractionId(a: any): string {
  if (a && typeof a.id === 'string' && a.id) return a.id
  const key = attractionKey(a)
  let h = 2166136261
  for (let i = 0; i < key.length; i++) {
    h ^= key.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  return 'nb-' + (h >>> 0).toString(36)
}

// 对齐行程：老数据行程内嵌景点无 id、attractionIds 为空，按内容映射到景点池 id；
// 仅保留池中存在的景点，并同步重建内嵌 attractions
function normalizeItinerary(rawItinerary: any[], pool: any[]): any[] {
  const byKey = new Map<string, any>()
  for (const item of pool) byKey.set(attractionKey(item), item)

  return (rawItinerary || [])
    .map((day) => {
      const embedded: any[] = Array.isArray(day?.attractions) ? day.attractions : []
      let ids: string[] = Array.isArray(day?.attractionIds)
        ? day.attractionIds.filter((id: unknown): id is string => typeof id === 'string' && id !== '')
        : []
      if (ids.length === 0 && embedded.length > 0) {
        ids = embedded
          .map((e) => byKey.get(attractionKey(e))?.id)
          .filter((id: string | undefined): id is string => !!id)
      }
      const resolved = ids
        .map((id) => pool.find((p) => p.id === id))
        .filter((item) => item !== undefined)
      return {
        ...(day ?? {}),
        attractions: resolved,
        attractionIds: resolved.map((p) => p.id),
      }
    })
    .flat()
}

function toTravelGuide(raw: any): TravelGuide {
  const attractions: TravelAttraction[] = (raw.attractions || []).map((a: any) => ({
    ...a,
    id: ensureAttractionId(a),
  }))
  return {
    id: raw.id,
    title: raw.title,
    summary: raw.summary,
    coverImage: raw.cover_image,
    status: raw.status,
    destination: raw.destination,
    region: raw.region,
    categoryId: raw.category_id || '',
    categoryName: raw.category_name || '',
    days: raw.days,
    bestMonth: raw.best_month,
    viewCount: raw.view_count,
    likeCount: raw.like_count,
    rating: raw.rating,
    reviewCount: raw.review_count,
    createdAt: raw.created_at,
    attractions,
    itinerary: normalizeItinerary(raw.itinerary, attractions) as TravelItineraryDay[],
    reviews: raw.reviews || [],
  }
}

export const travelApi = {
  async getList(params: TravelGuideListParams) {
    const raw = await request.get<PaginatedData<any>>('/travels', { params })
    return {
      ...raw,
      list: (raw.list || []).map(toTravelGuide),
    } as PaginatedData<TravelGuide>
  },
  async getDetail(id: string) {
    const raw = await request.get<any>(`/travels/${id}`)
    return toTravelGuide(raw)
  },
  async create(data: TravelGuideFormData) {
    const raw = await request.post<any>('/travels', toCreateReq(data))
    return toTravelGuide(raw)
  },
  async update(id: string, data: TravelGuideFormData) {
    const raw = await request.put<any>(`/travels/${id}`, toUpdateReq(data))
    return toTravelGuide(raw)
  },
  remove(id: string) {
    return request.delete(`/travels/${id}`)
  },
  updateStatus(id: string, status: number) {
    return request.put(`/travels/${id}/status`, { status })
  },
}
