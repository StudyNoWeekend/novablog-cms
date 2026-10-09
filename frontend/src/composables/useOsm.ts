import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import type { MapSearchResult } from './mapTypes'

// OSM 免 Key 兜底选点：Leaflet + OpenStreetMap 瓦片 + Nominatim 检索（WGS-84 坐标系）。
// 用于高德/Google Key 未配置或加载失败时的降级，接口与 useAmap/useGoogleMap 保持一致。

let mapInstance: L.Map | null = null
let marker: L.Marker | null = null
let clickHandler: ((lat: number, lng: number) => void) | null = null

const NOMINATIM_BASE = 'https://nominatim.openstreetmap.org'

export function useOsm() {
  async function initMap(
    container: HTMLElement,
    center?: { lat: number; lng: number },
  ): Promise<void> {
    mapInstance = L.map(container, {
      center: center ? [center.lat, center.lng] : [35.8617, 104.1954],
      zoom: center ? 10 : 4,
      scrollWheelZoom: true,
    })
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 18,
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">OpenStreetMap</a> contributors',
    }).addTo(mapInstance)

    mapInstance.on('click', (e: L.LeafletMouseEvent) => {
      if (clickHandler) clickHandler(e.latlng.lat, e.latlng.lng)
    })
  }

  function onMapClick(callback: (lat: number, lng: number) => void): void {
    clickHandler = callback
  }

  async function searchPlaces(keyword: string): Promise<MapSearchResult[]> {
    const url =
      `${NOMINATIM_BASE}/search?format=jsonv2&limit=10&accept-language=zh-CN` +
      `&q=${encodeURIComponent(keyword)}`
    const res = await fetch(url, { headers: { Accept: 'application/json' } })
    if (!res.ok) throw new Error(`Nominatim search failed: ${res.status}`)
    const list = (await res.json()) as Array<{
      lat: string
      lon: string
      name?: string
      display_name: string
    }>
    return list.map((item) => ({
      name: item.name || item.display_name.split(',')[0],
      address: item.display_name,
      latitude: parseFloat(item.lat),
      longitude: parseFloat(item.lon),
    }))
  }

  async function reverseGeocode(lat: number, lng: number): Promise<string> {
    try {
      const url = `${NOMINATIM_BASE}/reverse?format=jsonv2&accept-language=zh-CN&lat=${lat}&lon=${lng}`
      const res = await fetch(url, { headers: { Accept: 'application/json' } })
      if (!res.ok) throw new Error(String(res.status))
      const data = (await res.json()) as { display_name?: string }
      return data.display_name || `${lat.toFixed(6)}, ${lng.toFixed(6)}`
    } catch {
      return `${lat.toFixed(6)}, ${lng.toFixed(6)}`
    }
  }

  function placeMarker(lat: number, lng: number): void {
    if (!mapInstance) return
    const icon = L.divIcon({
      className: 'osm-marker',
      html: '<span style="display:block;width:14px;height:14px;border-radius:50%;background:#526FE8;border:2px solid #fff;box-shadow:0 1px 4px rgba(0,0,0,.35)"></span>',
      iconSize: [14, 14],
      iconAnchor: [7, 7],
    })
    if (marker) {
      marker.setLatLng([lat, lng])
    } else {
      marker = L.marker([lat, lng], { icon }).addTo(mapInstance)
    }
    mapInstance.setView([lat, lng], Math.max(mapInstance.getZoom(), 12))
  }

  function setCenter(lat: number, lng: number): void {
    if (!mapInstance) return
    mapInstance.setView([lat, lng])
  }

  function destroyMap(): void {
    if (mapInstance) {
      mapInstance.remove()
      mapInstance = null
    }
    marker = null
    clickHandler = null
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
