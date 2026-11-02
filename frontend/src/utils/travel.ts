import { TravelStatus, type TravelGuideFormData } from '@/types/travel'

// 创建空白游记表单数据（默认 3 天行程骨架）。
export function createEmptyGuide(): TravelGuideFormData {
  return {
    id: undefined,
    title: '',
    summary: '',
    coverImage: null,
    status: TravelStatus.Draft,
    destination: '',
    region: 'china',
    categoryId: undefined,
    days: 3,
    bestMonth: '',
    viewCount: 0,
    likeCount: 0,
    rating: 4.8,
    reviewCount: 0,
    createdAt: undefined,
    attractions: [],
    itinerary: Array.from({ length: 3 }, (_, index) => ({
      day: index + 1,
      title: '',
      description: '',
      attractions: [],
      attractionIds: [],
    })),
    reviews: [],
  }
}
