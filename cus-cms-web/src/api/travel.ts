import request from './request'
import type { TravelGuide, TravelGuideFormData } from '@/types/travel'
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
    category_id: data.categoryId,
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

function toTravelGuide(raw: any): TravelGuide {
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
    attractions: raw.attractions || [],
    itinerary: raw.itinerary || [],
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
