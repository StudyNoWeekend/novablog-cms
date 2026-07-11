<template>
  <a-modal
    :open="visible"
    title="选择定位"
    :width="800"
    :footer="null"
    destroy-on-close
    @cancel="handleCancel"
  >
    <div class="location-picker">
      <!-- 地图服务标签 -->
      <div class="map-service-tag">
        <a-tag :color="isAmap ? 'green' : 'blue'">
          {{ isAmap ? '高德地图' : 'Google Maps' }}
        </a-tag>
        <span class="map-service-hint">
          {{ isAmap ? '当前地区使用高德地图' : '当前地区使用 Google Maps' }}
        </span>
      </div>

      <!-- 搜索框 -->
      <div class="search-bar">
        <a-input-search
          v-model:value="searchKeyword"
          placeholder="搜索地点名称或地址"
          enter-button="搜索"
          :loading="searching"
          @search="handleSearch"
        />
      </div>

      <!-- 搜索结果列表 -->
      <div v-if="searchResults.length > 0" class="search-results">
        <div
          v-for="(item, index) in searchResults"
          :key="index"
          class="search-result-item"
          @click="handleSelectSearchResult(item)"
        >
          <EnvironmentOutlined class="result-icon" />
          <div class="result-info">
            <div class="result-name">{{ item.name }}</div>
            <div class="result-address">{{ item.address }}</div>
          </div>
        </div>
      </div>

      <!-- 地图容器 -->
      <div ref="mapContainer" class="map-container" />

      <!-- 选中位置信息 -->
      <div v-if="selectedAddress" class="selected-info">
        <div class="selected-address">
          <EnvironmentOutlined />
          <span>{{ selectedAddress }}</span>
        </div>
        <div class="selected-coords">
          经度: {{ selectedLongitude?.toFixed(6) }}, 纬度: {{ selectedLatitude?.toFixed(6) }}
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="actions">
        <a-button @click="handleCancel">取消</a-button>
        <a-button type="primary" :disabled="!selectedAddress" @click="handleConfirm">
          确认
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { EnvironmentOutlined } from '@ant-design/icons-vue'
import { useAmap } from '@/composables/useAmap'
import { useGoogleMap } from '@/composables/useGoogleMap'
import type { MapSearchResult, MapLocationResult } from '@/composables/mapTypes'
import type { TravelRegion } from '@/types/travel'

const props = defineProps<{
  visible: boolean
  region: TravelRegion
}>()

const emit = defineEmits<{
  (e: 'select', location: MapLocationResult): void
  (e: 'cancel'): void
}>()

const mapContainer = ref<HTMLElement | null>(null)
const searchKeyword = ref('')
const searching = ref(false)
const searchResults = ref<MapSearchResult[]>([])
const selectedAddress = ref('')
const selectedLatitude = ref<number | null>(null)
const selectedLongitude = ref<number | null>(null)

const isAmap = computed(() => props.region === 'china')

// 区域中心坐标映射 (高德用 [lng, lat], 谷歌用 {lat, lng})
const regionCenters: Record<string, { lat: number; lng: number }> = {
  china: { lat: 35.8617, lng: 104.1954 },
  japan: { lat: 36.2048, lng: 138.2529 },
  korea: { lat: 35.9078, lng: 127.7669 },
  'southeast-asia': { lat: 1.3521, lng: 103.8198 },
  'south-asia': { lat: 22.5937, lng: 78.9629 },
  'central-asia': { lat: 48.0196, lng: 66.9237 },
  'middle-east': { lat: 29.9765, lng: 45.8366 },
  'western-europe': { lat: 46.6034, lng: 1.8883 },
  'southern-europe': { lat: 40.4379, lng: 14.2898 },
  'northern-europe': { lat: 60.1282, lng: 18.6435 },
  'eastern-europe': { lat: 50.4501, lng: 30.5234 },
  'uk-ireland': { lat: 54.3781, lng: -3.4360 },
  'north-america': { lat: 40.7128, lng: -100.0060 },
  'central-caribbean': { lat: 18.1096, lng: -77.2975 },
  'south-america': { lat: -14.2350, lng: -51.9253 },
  australia: { lat: -25.2744, lng: 133.7751 },
  'new-zealand': { lat: -40.9006, lng: 174.8860 },
  'pacific-islands': { lat: -16.5783, lng: -179.5469 },
  'north-africa': { lat: 26.3351, lng: 17.2283 },
  'east-africa': { lat: -1.9403, lng: 37.2882 },
  'southern-africa': { lat: -28.8166, lng: 24.6094 },
  'west-central-africa': { lat: 6.6111, lng: 16.3939 },
  antarctica: { lat: -75.2509, lng: 0.0714 },
  arctic: { lat: 78.5551, lng: 15.6496 },
}

let mapReady = false

const amap = useAmap()
const googleMap = useGoogleMap()

async function initMap() {
  if (!mapContainer.value) return

  const center = regionCenters[props.region] ?? { lat: 35.8617, lng: 104.1954 }

  if (isAmap.value) {
    await amap.initMap(mapContainer.value, [center.lng, center.lat])
    amap.onMapClick(async (lng: number, lat: number) => {
      await handleMapClick(lng, lat)
    })
  } else {
    await googleMap.initMap(mapContainer.value, { lat: center.lat, lng: center.lng })
    googleMap.onMapClick(async (lat: number, lng: number) => {
      await handleMapClick(lng, lat)
    })
  }

  mapReady = true
}

function destroyMap() {
  if (isAmap.value) {
    amap.destroyMap()
  } else {
    googleMap.destroyMap()
  }
  mapReady = false
}

async function handleMapClick(lng: number, lat: number) {
  selectedLatitude.value = lat
  selectedLongitude.value = lng

  // 放置标记
  if (isAmap.value) {
    amap.placeMarker(lng, lat)
    const address = await amap.reverseGeocode(lng, lat)
    selectedAddress.value = address
  } else {
    googleMap.placeMarker(lat, lng)
    const address = await googleMap.reverseGeocode(lat, lng)
    selectedAddress.value = address
  }
}

async function handleSearch() {
  const keyword = searchKeyword.value.trim()
  if (!keyword || !mapReady) return

  searching.value = true
  try {
    if (isAmap.value) {
      searchResults.value = await amap.searchPlaces(keyword)
    } else {
      searchResults.value = await googleMap.searchPlaces(keyword)
    }
  } catch {
    searchResults.value = []
  } finally {
    searching.value = false
  }
}

async function handleSelectSearchResult(item: MapSearchResult) {
  selectedAddress.value = item.address || item.name
  selectedLatitude.value = item.latitude
  selectedLongitude.value = item.longitude

  if (isAmap.value) {
    amap.placeMarker(item.longitude, item.latitude)
  } else {
    googleMap.placeMarker(item.latitude, item.longitude)
  }

  searchResults.value = []
  searchKeyword.value = ''
}

function handleConfirm() {
  if (!selectedAddress.value || selectedLatitude.value === null || selectedLongitude.value === null) {
    return
  }

  emit('select', {
    address: selectedAddress.value,
    latitude: selectedLatitude.value,
    longitude: selectedLongitude.value,
  })
}

function handleCancel() {
  emit('cancel')
}

function resetState() {
  searchKeyword.value = ''
  searchResults.value = []
  selectedAddress.value = ''
  selectedLatitude.value = null
  selectedLongitude.value = null
}

watch(
  () => props.visible,
  async (val) => {
    if (val) {
      resetState()
      await nextTick()
      await initMap()
    } else {
      destroyMap()
    }
  },
)
</script>

<style scoped>
.location-picker {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.map-service-tag {
  display: flex;
  align-items: center;
  gap: 8px;
}

.map-service-hint {
  font-size: 13px;
  color: var(--text-tertiary, #94a3b8);
}

.search-bar {
  position: relative;
}

.search-results {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  background: var(--bg-card, #ffffff);
}

.search-result-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.2s;
}

.search-result-item:hover {
  background: var(--bg-hover, #f1f5f9);
}

.search-result-item + .search-result-item {
  border-top: 1px solid var(--border-color, #e2e8f0);
}

.result-icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: var(--color-primary, #4a6cf7);
}

.result-info {
  flex: 1;
  min-width: 0;
}

.result-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #1e293b);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-address {
  font-size: 12px;
  color: var(--text-secondary, #64748b);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.map-container {
  width: 100%;
  height: 350px;
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
  border: 1px solid var(--border-color, #e2e8f0);
  background: #e8eaed;
}

.selected-info {
  padding: 10px 12px;
  background: var(--bg-hover, #f8fafc);
  border-radius: var(--border-radius, 8px);
  border: 1px solid var(--border-color, #e2e8f0);
}

.selected-address {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--text-primary, #1e293b);
  font-weight: 500;
}

.selected-address span {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selected-coords {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
