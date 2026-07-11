import AMapLoader from '@amap/amap-jsapi-loader'
import type { MapSearchResult } from './mapTypes'

let mapInstance: any = null
let AMapNamespace: any = null
let marker: any = null
let clickHandler: ((lng: number, lat: number) => void) | null = null
let placeSearch: any = null
let geocoder: any = null

export function useAmap() {
  const amapKey = import.meta.env.VITE_AMAP_KEY
  const amapSecurityCode = import.meta.env.VITE_AMAP_SECURITY_CODE

  async function initMap(container: HTMLElement, center?: [number, number]): Promise<void> {
    if (!amapKey) {
      console.warn('[useAmap] VITE_AMAP_KEY is not configured. Map will not load.')
      return
    }

    // Set security config before loading
    if (amapSecurityCode) {
      ;(window as any)._AMapSecurityConfig = {
        securityJsCode: amapSecurityCode,
      }
    }

    AMapNamespace = await AMapLoader.load({
      key: amapKey,
      version: '2.0',
      plugins: ['AMap.PlaceSearch', 'AMap.Geocoder', 'AMap.AutoComplete'],
    })

    const options: any = {
      zoom: 13,
      viewMode: '2D',
    }
    if (center) {
      options.center = center
    }

    mapInstance = new AMapNamespace.Map(container, options)

    // Initialize plugins
    placeSearch = new AMapNamespace.PlaceSearch({
      pageSize: 10,
      pageIndex: 1,
    })
    geocoder = new AMapNamespace.Geocoder()

    // Register click handler
    mapInstance.on('click', (e: any) => {
      const lng = e.lnglat.getLng()
      const lat = e.lnglat.getLat()
      if (clickHandler) {
        clickHandler(lng, lat)
      }
    })
  }

  function onMapClick(callback: (lng: number, lat: number) => void): void {
    clickHandler = callback
  }

  function searchPlaces(keyword: string): Promise<MapSearchResult[]> {
    return new Promise((resolve, reject) => {
      if (!placeSearch) {
        reject(new Error('Map not initialized'))
        return
      }
      placeSearch.search(keyword, (status: string, result: any) => {
        if (status === 'complete' && result.poiList) {
          const results: MapSearchResult[] = result.poiList.pois.map((poi: any) => ({
            name: poi.name,
            address: poi.address || poi.name,
            latitude: poi.location.getLat(),
            longitude: poi.location.getLng(),
          }))
          resolve(results)
        } else {
          resolve([])
        }
      })
    })
  }

  function reverseGeocode(lng: number, lat: number): Promise<string> {
    return new Promise((resolve, reject) => {
      if (!geocoder) {
        reject(new Error('Map not initialized'))
        return
      }
      const lnglat = [lng, lat]
      geocoder.getAddress(lnglat, (status: string, result: any) => {
        if (status === 'complete' && result.regeocode) {
          resolve(result.regeocode.formattedAddress)
        } else {
          resolve(`${lat.toFixed(6)}, ${lng.toFixed(6)}`)
        }
      })
    })
  }

  function placeMarker(lng: number, lat: number): void {
    if (!mapInstance || !AMapNamespace) return
    if (marker) {
      marker.setPosition([lng, lat])
    } else {
      marker = new AMapNamespace.Marker({
        position: [lng, lat],
        anchor: 'bottom-center',
      })
      mapInstance.add(marker)
    }
    mapInstance.setCenter([lng, lat])
  }

  function setCenter(lng: number, lat: number): void {
    if (!mapInstance) return
    mapInstance.setCenter([lng, lat])
  }

  function destroyMap(): void {
    if (mapInstance) {
      mapInstance.destroy()
      mapInstance = null
    }
    marker = null
    clickHandler = null
    placeSearch = null
    geocoder = null
  }

  return {
    initMap,
    onMapClick,
    searchPlaces,
    reverseGeocode,
    placeMarker,
    setCenter,
    destroyMap,
  }
}
