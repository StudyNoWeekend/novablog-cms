export const TravelStatus = {
  Draft: 1,
  Published: 2,
  Archived: 3,
} as const

export type TravelStatus = (typeof TravelStatus)[keyof typeof TravelStatus]

export type TravelContinent =
  | 'asia'
  | 'europe'
  | 'americas'
  | 'oceania'
  | 'africa'
  | 'polar'

export type TravelRegion =
  | 'china'
  | 'japan'
  | 'korea'
  | 'southeast-asia'
  | 'south-asia'
  | 'central-asia'
  | 'middle-east'
  | 'western-europe'
  | 'southern-europe'
  | 'northern-europe'
  | 'eastern-europe'
  | 'uk-ireland'
  | 'north-america'
  | 'central-caribbean'
  | 'south-america'
  | 'australia'
  | 'new-zealand'
  | 'pacific-islands'
  | 'north-africa'
  | 'east-africa'
  | 'southern-africa'
  | 'west-central-africa'
  | 'antarctica'
  | 'arctic'

export type TravelRegionFilter = TravelRegion | TravelContinent | 'all'

export interface RegionTreeNode {
  value: string
  label: string
  children?: RegionTreeNode[]
}

export const regionTree: RegionTreeNode[] = [
  {
    value: 'asia',
    label: '亚洲',
    children: [
      { value: 'china', label: '中国' },
      { value: 'japan', label: '日本' },
      { value: 'korea', label: '韩国' },
      { value: 'southeast-asia', label: '东南亚' },
      { value: 'south-asia', label: '南亚' },
      { value: 'central-asia', label: '中亚' },
      { value: 'middle-east', label: '中东' },
    ],
  },
  {
    value: 'europe',
    label: '欧洲',
    children: [
      { value: 'western-europe', label: '西欧' },
      { value: 'southern-europe', label: '南欧' },
      { value: 'northern-europe', label: '北欧' },
      { value: 'eastern-europe', label: '东欧' },
      { value: 'uk-ireland', label: '英国与爱尔兰' },
    ],
  },
  {
    value: 'americas',
    label: '美洲',
    children: [
      { value: 'north-america', label: '北美' },
      { value: 'central-caribbean', label: '中美与加勒比' },
      { value: 'south-america', label: '南美' },
    ],
  },
  {
    value: 'oceania',
    label: '大洋洲',
    children: [
      { value: 'australia', label: '澳大利亚' },
      { value: 'new-zealand', label: '新西兰' },
      { value: 'pacific-islands', label: '太平洋岛屿' },
    ],
  },
  {
    value: 'africa',
    label: '非洲',
    children: [
      { value: 'north-africa', label: '北非' },
      { value: 'east-africa', label: '东非' },
      { value: 'southern-africa', label: '南非' },
      { value: 'west-central-africa', label: '西非与中非' },
    ],
  },
  {
    value: 'polar',
    label: '极地',
    children: [
      { value: 'antarctica', label: '南极洲' },
      { value: 'arctic', label: '北极' },
    ],
  },
]

const regionToContinentMap: Record<TravelRegion, TravelContinent> = {
  china: 'asia',
  japan: 'asia',
  korea: 'asia',
  'southeast-asia': 'asia',
  'south-asia': 'asia',
  'central-asia': 'asia',
  'middle-east': 'asia',
  'western-europe': 'europe',
  'southern-europe': 'europe',
  'northern-europe': 'europe',
  'eastern-europe': 'europe',
  'uk-ireland': 'europe',
  'north-america': 'americas',
  'central-caribbean': 'americas',
  'south-america': 'americas',
  australia: 'oceania',
  'new-zealand': 'oceania',
  'pacific-islands': 'oceania',
  'north-africa': 'africa',
  'east-africa': 'africa',
  'southern-africa': 'africa',
  'west-central-africa': 'africa',
  antarctica: 'polar',
  arctic: 'polar',
}

const regionLabelMap: Record<TravelRegionFilter, string> = {
  all: '全部',
  asia: '亚洲',
  europe: '欧洲',
  americas: '美洲',
  oceania: '大洋洲',
  africa: '非洲',
  polar: '极地',
  china: '中国',
  japan: '日本',
  korea: '韩国',
  'southeast-asia': '东南亚',
  'south-asia': '南亚',
  'central-asia': '中亚',
  'middle-east': '中东',
  'western-europe': '西欧',
  'southern-europe': '南欧',
  'northern-europe': '北欧',
  'eastern-europe': '东欧',
  'uk-ireland': '英国与爱尔兰',
  'north-america': '北美',
  'central-caribbean': '中美与加勒比',
  'south-america': '南美',
  australia: '澳大利亚',
  'new-zealand': '新西兰',
  'pacific-islands': '太平洋岛屿',
  'north-africa': '北非',
  'east-africa': '东非',
  'southern-africa': '南非',
  'west-central-africa': '西非与中非',
  antarctica: '南极洲',
  arctic: '北极',
}

export function getContinent(region: TravelRegion): TravelContinent
export function getContinent(region: TravelRegionFilter): TravelContinent | undefined
export function getContinent(region: TravelRegionFilter): TravelContinent | undefined {
  if (region in regionToContinentMap) {
    return regionToContinentMap[region as TravelRegion]
  }
  if ((regionTree as RegionTreeNode[]).some((continent) => continent.value === region)) {
    return region as TravelContinent
  }
  return undefined
}

export function getRegionPath(region: TravelRegionFilter): string {
  if (region === 'all') return '全部'

  const continent = getContinent(region)
  if (!continent) return regionLabelMap[region] ?? region

  const continentLabel = regionLabelMap[continent] ?? continent
  const regionLabel = regionLabelMap[region] ?? region

  if (region === continent) {
    return continentLabel
  }

  return `${continentLabel} / ${regionLabel}`
}

export function findRegionPath(region: TravelRegionFilter): string[] {
  if (region === 'all') return []

  for (const continent of regionTree) {
    if (continent.value === region) {
      return [continent.value]
    }
    const child = continent.children?.find((item) => item.value === region)
    if (child) {
      return [continent.value, child.value]
    }
  }

  return []
}

export interface TravelAttraction {
  id: string
  name: string
  description: string
  image: string
  duration: string
  location?: string
  latitude?: number
  longitude?: number
}

export interface TravelItineraryDay {
  day: number
  title: string
  description: string
  attractions: TravelAttraction[]
  attractionIds: string[]
}

export interface TravelReview {
  id: string
  username: string
  avatar: string
  rating: number
  content: string
  date: string
}

export interface TravelGuide {
  id: string
  title: string
  summary: string
  coverImage: string | null
  status: TravelStatus
  destination: string
  region: TravelRegion
  categoryId: string
  categoryName: string
  days: number
  bestMonth: string
  viewCount: number
  likeCount: number
  rating: number
  reviewCount: number
  createdAt: string
  itinerary: TravelItineraryDay[]
  attractions: TravelAttraction[]
  reviews: TravelReview[]
}

export interface TravelGuideFormData {
  id: string | undefined
  title: string
  summary: string
  coverImage: string | null
  status: TravelStatus
  destination: string
  region: TravelRegion
  categoryId: string | undefined
  days: number
  bestMonth: string
  viewCount?: number
  likeCount?: number
  rating?: number
  reviewCount?: number
  createdAt: string | undefined
  itinerary: TravelItineraryDay[]
  attractions: TravelAttraction[]
  reviews: TravelReview[]
}

export interface TravelFilters {
  keyword?: string
  region?: TravelRegionFilter
  days?: 'all' | '1-3' | '4-7' | '8-14' | '15+'
  sort?: 'latest' | 'views' | 'rating' | 'likes'
  categoryId?: string | undefined
}

export function toTravelGuide(formData: TravelGuideFormData): TravelGuide {
  return {
    ...formData,
    id: formData.id ?? '',
    categoryId: formData.categoryId ?? '',
    categoryName: '',
    viewCount: formData.viewCount ?? 0,
    likeCount: formData.likeCount ?? 0,
    rating: formData.rating ?? 4.8,
    reviewCount: formData.reviewCount ?? 0,
    createdAt: formData.createdAt ?? new Date().toISOString(),
  }
}

export function toFormData(guide: TravelGuide): TravelGuideFormData {
  return {
    ...guide,
  }
}
