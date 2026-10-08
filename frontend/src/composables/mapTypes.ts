export interface MapLocationResult {
  address: string
  latitude: number
  longitude: number
  coordType: 'gcj02' | 'wgs84'
}

export interface MapSearchResult {
  name: string
  address: string
  latitude: number
  longitude: number
}
