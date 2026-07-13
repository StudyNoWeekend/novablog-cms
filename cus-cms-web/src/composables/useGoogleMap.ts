/// <reference types="google.maps" />
import { setOptions, importLibrary } from '@googlemaps/js-api-loader'
import type { MapSearchResult } from './mapTypes'

let mapInstance: google.maps.Map | null = null
let marker: google.maps.Marker | null = null
let clickHandler: ((lat: number, lng: number) => void) | null = null
let geocoder: google.maps.Geocoder | null = null
let placesService: google.maps.places.PlacesService | null = null
let autocompleteService: google.maps.places.AutocompleteService | null = null

export function useGoogleMap() {
  const apiKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY

  async function initMap(
    container: HTMLElement,
    center?: { lat: number; lng: number },
  ): Promise<void> {
    if (!apiKey) {
      console.warn('[useGoogleMap] VITE_GOOGLE_MAPS_API_KEY is not configured. Map will not load.')
      return
    }

    setOptions({
      key: apiKey,
      v: 'weekly',
    })

    // Load required libraries
    const mapsLib = await importLibrary('maps')
    await importLibrary('geometry')
    const placesLib = await importLibrary('places')
    const geocodingLib = await importLibrary('geocoding')
    await importLibrary('marker')

    const options: google.maps.MapOptions = {
      zoom: 13,
      center: center ?? { lat: 0, lng: 0 },
    }

    mapInstance = new mapsLib.Map(container, options)
    geocoder = new geocodingLib.Geocoder()
    placesService = new placesLib.PlacesService(mapInstance)
    autocompleteService = new placesLib.AutocompleteService()

    // Register click handler
    mapInstance.addListener('click', (e: google.maps.MapMouseEvent) => {
      if (e.latLng && clickHandler) {
        clickHandler(e.latLng.lat(), e.latLng.lng())
      }
    })
  }

  function onMapClick(callback: (lat: number, lng: number) => void): void {
    clickHandler = callback
  }

  function searchPlaces(keyword: string): Promise<MapSearchResult[]> {
    return new Promise((resolve, reject) => {
      if (!autocompleteService || !placesService) {
        reject(new Error('Map not initialized'))
        return
      }

      autocompleteService.getPlacePredictions(
        {
          input: keyword,
          types: ['geocode', 'establishment'],
        },
        (predictions, status) => {
          if (status !== google.maps.places.PlacesServiceStatus.OK || !predictions) {
            // Fallback to text search
            if (placesService) {
              placesService.textSearch(
                { query: keyword },
                (results, textStatus) => {
                  if (textStatus === google.maps.places.PlacesServiceStatus.OK && results) {
                    const mapped: MapSearchResult[] = results.slice(0, 10).map((place) => ({
                      name: place.name ?? '',
                      address: place.formatted_address ?? place.name ?? '',
                      latitude: place.geometry?.location?.lat() ?? 0,
                      longitude: place.geometry?.location?.lng() ?? 0,
                    }))
                    resolve(mapped)
                  } else {
                    resolve([])
                  }
                },
              )
            } else {
              resolve([])
            }
            return
          }

          // Get details for each prediction
          const placeIds = predictions.slice(0, 10).map((p) => p.place_id)
          const results: MapSearchResult[] = []
          let completed = 0

          if (placeIds.length === 0) {
            resolve([])
            return
          }

          placeIds.forEach((placeId) => {
            placesService!.getDetails(
              {
                placeId,
                fields: ['name', 'formatted_address', 'geometry'],
              },
              (place, detailStatus) => {
                completed++
                if (
                  detailStatus === google.maps.places.PlacesServiceStatus.OK &&
                  place &&
                  place.geometry?.location
                ) {
                  results.push({
                    name: place.name ?? '',
                    address: place.formatted_address ?? place.name ?? '',
                    latitude: place.geometry.location.lat(),
                    longitude: place.geometry.location.lng(),
                  })
                }
                if (completed === placeIds.length) {
                  resolve(results)
                }
              },
            )
          })
        },
      )
    })
  }

  function reverseGeocode(lat: number, lng: number): Promise<string> {
    return new Promise((resolve, reject) => {
      if (!geocoder) {
        reject(new Error('Map not initialized'))
        return
      }
      geocoder.geocode(
        { location: { lat, lng } },
        (results, status) => {
          if (status === 'OK' && results && results.length > 0) {
            resolve(results[0].formatted_address)
          } else {
            resolve(`${lat.toFixed(6)}, ${lng.toFixed(6)}`)
          }
        },
      )
    })
  }

  function placeMarker(lat: number, lng: number): void {
    if (!mapInstance) return
    if (marker) {
      marker.setPosition({ lat, lng })
    } else {
      marker = new google.maps.Marker({
        position: { lat, lng },
        map: mapInstance,
      })
    }
    mapInstance.setCenter({ lat, lng })
  }

  function setCenter(lat: number, lng: number): void {
    if (!mapInstance) return
    mapInstance.setCenter({ lat, lng })
  }

  function destroyMap(): void {
    if (marker) {
      marker.setMap(null)
      marker = null
    }
    mapInstance = null
    clickHandler = null
    geocoder = null
    placesService = null
    autocompleteService = null
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
