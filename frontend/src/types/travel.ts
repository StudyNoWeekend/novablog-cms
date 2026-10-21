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
  /** 地区名称的 i18n key（文案见 travel.regions.*），由消费方用 t() 渲染 */
  labelKey: string
  children?: RegionTreeNode[]
}

export const regionTree: RegionTreeNode[] = [
  {
    value: 'asia',
    labelKey: 'travel.regions.asia',
    children: [
      { value: 'china', labelKey: 'travel.regions.china' },
      { value: 'japan', labelKey: 'travel.regions.japan' },
      { value: 'korea', labelKey: 'travel.regions.korea' },
      { value: 'southeast-asia', labelKey: 'travel.regions.southeast-asia' },
      { value: 'south-asia', labelKey: 'travel.regions.south-asia' },
      { value: 'central-asia', labelKey: 'travel.regions.central-asia' },
      { value: 'middle-east', labelKey: 'travel.regions.middle-east' },
    ],
  },
  {
    value: 'europe',
    labelKey: 'travel.regions.europe',
    children: [
      { value: 'western-europe', labelKey: 'travel.regions.western-europe' },
      { value: 'southern-europe', labelKey: 'travel.regions.southern-europe' },
      { value: 'northern-europe', labelKey: 'travel.regions.northern-europe' },
      { value: 'eastern-europe', labelKey: 'travel.regions.eastern-europe' },
      { value: 'uk-ireland', labelKey: 'travel.regions.uk-ireland' },
    ],
  },
  {
    value: 'americas',
    labelKey: 'travel.regions.americas',
    children: [
      { value: 'north-america', labelKey: 'travel.regions.north-america' },
      { value: 'central-caribbean', labelKey: 'travel.regions.central-caribbean' },
      { value: 'south-america', labelKey: 'travel.regions.south-america' },
    ],
  },
  {
    value: 'oceania',
    labelKey: 'travel.regions.oceania',
    children: [
      { value: 'australia', labelKey: 'travel.regions.australia' },
      { value: 'new-zealand', labelKey: 'travel.regions.new-zealand' },
      { value: 'pacific-islands', labelKey: 'travel.regions.pacific-islands' },
    ],
  },
  {
    value: 'africa',
    labelKey: 'travel.regions.africa',
    children: [
      { value: 'north-africa', labelKey: 'travel.regions.north-africa' },
      { value: 'east-africa', labelKey: 'travel.regions.east-africa' },
      { value: 'southern-africa', labelKey: 'travel.regions.southern-africa' },
      { value: 'west-central-africa', labelKey: 'travel.regions.west-central-africa' },
    ],
  },
  {
    value: 'polar',
    labelKey: 'travel.regions.polar',
    children: [
      { value: 'antarctica', labelKey: 'travel.regions.antarctica' },
      { value: 'arctic', labelKey: 'travel.regions.arctic' },
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

// 地区枚举 → i18n key 映射（只存 key 字符串，禁止在此调用 t()；文案见 travel.regions.*）
const regionLabelMap: Record<TravelRegionFilter, string> = {
  all: 'travel.regions.all',
  asia: 'travel.regions.asia',
  europe: 'travel.regions.europe',
  americas: 'travel.regions.americas',
  oceania: 'travel.regions.oceania',
  africa: 'travel.regions.africa',
  polar: 'travel.regions.polar',
  china: 'travel.regions.china',
  japan: 'travel.regions.japan',
  korea: 'travel.regions.korea',
  'southeast-asia': 'travel.regions.southeast-asia',
  'south-asia': 'travel.regions.south-asia',
  'central-asia': 'travel.regions.central-asia',
  'middle-east': 'travel.regions.middle-east',
  'western-europe': 'travel.regions.western-europe',
  'southern-europe': 'travel.regions.southern-europe',
  'northern-europe': 'travel.regions.northern-europe',
  'eastern-europe': 'travel.regions.eastern-europe',
  'uk-ireland': 'travel.regions.uk-ireland',
  'north-america': 'travel.regions.north-america',
  'central-caribbean': 'travel.regions.central-caribbean',
  'south-america': 'travel.regions.south-america',
  australia: 'travel.regions.australia',
  'new-zealand': 'travel.regions.new-zealand',
  'pacific-islands': 'travel.regions.pacific-islands',
  'north-africa': 'travel.regions.north-africa',
  'east-africa': 'travel.regions.east-africa',
  'southern-africa': 'travel.regions.southern-africa',
  'west-central-africa': 'travel.regions.west-central-africa',
  antarctica: 'travel.regions.antarctica',
  arctic: 'travel.regions.arctic',
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

// 返回地区路径对应的 i18n key 列表（大洲在前，如 ['travel.regions.asia', 'travel.regions.china']），
// 消费方用 t() 逐项渲染后以「 / 」连接展示（跟随当前语言）
export function getRegionLabelKeys(region: TravelRegionFilter): string[] {
  if (region === 'all') return [regionLabelMap.all]

  const continent = getContinent(region)
  if (!continent) return [regionLabelMap[region] ?? region]

  if (region === continent) {
    return [regionLabelMap[continent]]
  }

  return [regionLabelMap[continent], regionLabelMap[region]]
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
  /** 选点来源坐标系：高德为 gcj02，Google 为 wgs84；旧数据可能缺失 */
  coordType?: 'gcj02' | 'wgs84'
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
