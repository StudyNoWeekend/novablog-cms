<template>
  <div class="travel-attraction-pool">
    <div class="section-header">
      <div class="section-title-wrap">
        <span class="section-title-bar" aria-hidden="true" />
        <h2 class="section-title">景点池</h2>
      </div>
      <p class="section-subtitle">先添加景点，再拖拽或选择到对应日期</p>
    </div>

    <a-button
      type="primary"
      class="add-button"
      aria-label="添加景点"
      @click="openAddModal"
    >
      <PlusOutlined />
      添加景点
    </a-button>

    <div
      v-if="attractions.length > 0"
      class="attractions-grid"
      role="list"
      aria-label="景点列表"
    >
      <div
        v-for="attraction in attractions"
        :key="attraction.id"
        class="attraction-card"
        role="listitem"
        draggable="true"
        tabindex="0"
        :aria-label="`景点：${attraction.name}`"
        @dragstart="handleDragStart($event, attraction)"
      >
        <div class="attraction-thumbnail">
          <img
            v-if="attraction.image"
            :src="attraction.image"
            :alt="`${attraction.name} 图片`"
            loading="lazy"
          />
          <div v-else class="thumbnail-placeholder">
            <PictureOutlined />
          </div>
        </div>

        <div class="attraction-content">
          <h3 class="attraction-name" :title="attraction.name">
            {{ attraction.name }}
          </h3>

          <div class="attraction-meta">
            <a-tag
              v-if="attraction.location"
              class="attraction-location"
              :title="locationTitle(attraction)"
            >
              <EnvironmentOutlined /> {{ attraction.location }}
            </a-tag>

            <span v-if="attraction.duration" class="attraction-duration">
              <ClockCircleOutlined /> {{ attraction.duration }}
            </span>
          </div>

          <div
            v-if="dayReferences.get(attraction.id)?.length"
            class="attraction-refs"
          >
            <a-tag
              v-for="(day, index) in dayReferences.get(attraction.id)"
              :key="day"
              :color="index % 2 === 0 ? 'success' : 'warning'"
              class="ref-tag"
            >
              第{{ day }}天
            </a-tag>
          </div>
        </div>

        <div class="attraction-actions">
          <a-button
            type="text"
            size="small"
            aria-label="编辑景点"
            @click.stop="openEdit(attraction)"
          >
            <EditOutlined />
          </a-button>
          <a-button
            type="text"
            size="small"
            danger
            aria-label="删除景点"
            @click.stop="handleDelete(attraction)"
          >
            <DeleteOutlined />
          </a-button>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <InboxOutlined class="empty-icon" />
      <p>还没有景点，点击上方按钮添加</p>
    </div>

    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :ok-text="isEditing ? '保存' : '添加'"
      cancel-text="取消"
      destroy-on-close
      @ok="handleModalOk"
      @cancel="closeModal"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item label="名称" required>
          <a-input
            v-model:value="form.name"
            placeholder="请输入景点名称"
            maxlength="100"
            show-count
          />
        </a-form-item>

        <a-form-item label="描述">
          <a-textarea
            v-model:value="form.description"
            placeholder="请输入景点描述"
            :rows="3"
            maxlength="500"
            show-count
          />
        </a-form-item>

        <a-form-item label="图片">
          <div class="image-field">
            <div
              class="image-card"
              role="button"
              tabindex="0"
              aria-label="选择图片"
              @click="imagePickerVisible = true"
              @keydown.enter="imagePickerVisible = true"
            >
              <template v-if="form.image">
                <img :src="form.image" alt="景点图片" />
                <div class="image-overlay"><span>更换图片</span></div>
                <a-button
                  type="text"
                  size="small"
                  class="image-clear-btn"
                  aria-label="移除图片"
                  @click.stop="form.image = ''"
                >
                  <CloseOutlined />
                </a-button>
              </template>
              <template v-else>
                <div class="image-placeholder">
                  <PictureOutlined class="image-placeholder-icon" />
                  <span class="image-placeholder-text">点击选择图片</span>
                </div>
              </template>
            </div>
          </div>
          <TravelCoverPicker
            v-model:visible="imagePickerVisible"
            @selected="onImageSelected"
          />
        </a-form-item>

        <a-form-item label="游玩天数" required>
          <a-input
            v-model:value="form.duration"
            placeholder="例如：1天、2天"
          />
        </a-form-item>

        <a-form-item label="选择定位" required>
          <div v-if="form.location" class="location-selected">
            <a-tag color="purple" class="location-tag">
              <EnvironmentOutlined /> {{ form.location }}
            </a-tag>
            <a-button type="link" size="small" @click="openLocationPicker">
              重新选择
            </a-button>
          </div>
          <a-button
            v-else
            type="dashed"
            block
            @click="openLocationPicker"
          >
            <EnvironmentOutlined /> 请选择定位
          </a-button>
        </a-form-item>

        <a-form-item label="备注">
          <a-textarea
            v-model:value="form.notes"
            placeholder="其他补充信息"
            :rows="2"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <LocationPickerModal
      v-model:visible="locationPickerVisible"
      :region="props.modelValue.region"
      @select="handleLocationSelect"
      @cancel="locationPickerVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EnvironmentOutlined,
  ClockCircleOutlined,
  PictureOutlined,
  InboxOutlined,
  CloseOutlined,
} from '@ant-design/icons-vue'
import type { TravelAttraction, TravelGuideFormData } from '@/types/travel'
import type { MapLocationResult } from '@/composables/mapTypes'
import LocationPickerModal from './LocationPickerModal.vue'
import TravelCoverPicker from './TravelCoverPicker.vue'

const props = defineProps<{
  modelValue: TravelGuideFormData
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: TravelGuideFormData): void
  (e: 'dragStart', attraction: TravelAttraction): void
}>()

interface AttractionFormData extends TravelAttraction {
  notes?: string
}

const modalVisible = ref(false)
const editingId = ref<string | null>(null)
const locationPickerVisible = ref(false)
const imagePickerVisible = ref(false)

const emptyForm = (): AttractionFormData => ({
  id: '',
  name: '',
  description: '',
  image: '',
  duration: '',
  location: '',
  latitude: undefined,
  longitude: undefined,
  notes: '',
})

const form = reactive<AttractionFormData>(emptyForm())

const isEditing = computed(() => editingId.value !== null)
const modalTitle = computed(() => (isEditing.value ? '编辑景点' : '添加景点'))
const attractions = computed(() => props.modelValue.attractions)

const dayReferences = computed(() => {
  const map = new Map<string, number[]>()
  props.modelValue.itinerary.forEach((day) => {
    day.attractionIds.forEach((id) => {
      if (!map.has(id)) {
        map.set(id, [])
      }
      const days = map.get(id)!
      if (!days.includes(day.day)) {
        days.push(day.day)
      }
    })
  })
  return map
})

function generateId(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function resetForm() {
  Object.assign(form, emptyForm())
  editingId.value = null
}

function openAddModal() {
  resetForm()
  modalVisible.value = true
}

function openEdit(attraction: TravelAttraction) {
  editingId.value = attraction.id
  const data = attraction as AttractionFormData
  Object.assign(form, {
    ...data,
    notes: data.notes ?? '',
  })
  modalVisible.value = true
}

function closeModal() {
  modalVisible.value = false
  resetForm()
}

function handleModalOk() {
  const name = form.name.trim()
  const duration = form.duration.trim()
  const location = form.location?.trim() ?? ''

  if (!name || !duration || !location) {
    return
  }

  const payload: TravelAttraction = {
    id: editingId.value ?? generateId(),
    name,
    description: form.description?.trim() ?? '',
    image: form.image?.trim() ?? '',
    duration,
    location,
    latitude: form.latitude,
    longitude: form.longitude,
    notes: form.notes?.trim() || undefined,
  } as TravelAttraction

  if (editingId.value) {
    const nextAttractions = props.modelValue.attractions.map((item) =>
      item.id === editingId.value ? payload : item,
    )
    emit('update:modelValue', {
      ...props.modelValue,
      attractions: nextAttractions,
    })
  } else {
    emit('update:modelValue', {
      ...props.modelValue,
      attractions: [...props.modelValue.attractions, payload],
    })
  }

  closeModal()
}

function handleDelete(attraction: TravelAttraction) {
  const refs = dayReferences.value.get(attraction.id) ?? []

  const doDelete = () => {
    const nextAttractions = props.modelValue.attractions.filter(
      (item) => item.id !== attraction.id,
    )
    const nextItinerary = props.modelValue.itinerary.map((day) => {
      if (!day.attractionIds.includes(attraction.id)) {
        return day
      }
      return {
        ...day,
        attractionIds: day.attractionIds.filter((id) => id !== attraction.id),
        attractions: day.attractions.filter((item) => item.id !== attraction.id),
      }
    })

    emit('update:modelValue', {
      ...props.modelValue,
      attractions: nextAttractions,
      itinerary: nextItinerary,
    })
  }

  if (refs.length > 0) {
    Modal.confirm({
      title: '确认删除',
      content: `该景点已被第 ${refs.join('、')} 天引用，删除后会从所有日行程中移除。是否继续？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: doDelete,
    })
  } else {
    doDelete()
  }
}

function handleDragStart(event: DragEvent, attraction: TravelAttraction) {
  emit('dragStart', attraction)
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'copy'
    event.dataTransfer.setData('application/json', JSON.stringify(attraction))
    event.dataTransfer.setData('text/plain', attraction.id)
  }
}

function openLocationPicker() {
  locationPickerVisible.value = true
}

function handleLocationSelect(location: MapLocationResult) {
  form.location = location.address
  form.latitude = location.latitude
  form.longitude = location.longitude
  locationPickerVisible.value = false
}

function onImageSelected(payload: { url: string }) {
  form.image = payload.url
  imagePickerVisible.value = false
}

function locationTitle(attraction: TravelAttraction): string {
  if (attraction.latitude != null && attraction.longitude != null) {
    return `${attraction.location} (经度: ${attraction.longitude.toFixed(6)}, 纬度: ${attraction.latitude.toFixed(6)})`
  }
  return attraction.location ?? ''
}
</script>

<style scoped>
.travel-attraction-pool {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
}

.section-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.section-title-bar {
  width: 4px;
  height: 20px;
  border-radius: 2px;
  background: var(--color-primary, #4a6cf7);
}

.section-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.section-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary, #94a3b8);
}

.add-button {
  align-self: flex-start;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.attractions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.attraction-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
  cursor: grab;
  transition:
    transform var(--transition-base, 250ms cubic-bezier(0.4, 0, 0.2, 1)),
    box-shadow var(--transition-base, 250ms cubic-bezier(0.4, 0, 0.2, 1));
}

.attraction-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0, 0, 0, 0.08));
}

.attraction-card:focus-visible {
  outline: 2px solid var(--color-primary, #4a6cf7);
  outline-offset: 2px;
}

.attraction-card:active {
  cursor: grabbing;
}

.attraction-thumbnail {
  flex-shrink: 0;
  width: 60px;
  height: 60px;
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
  background: #f1f5f9;
}

.attraction-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumbnail-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #cbd5e1;
}

.attraction-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.attraction-name {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.attraction-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.attraction-location {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: var(--color-accent, #7c3aed);
  color: #ffffff;
  border-color: transparent;
}

.attraction-duration {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-secondary, #64748b);
}

.attraction-refs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.ref-tag {
  font-size: 12px;
}

.attraction-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 48px 16px;
  color: var(--text-tertiary, #94a3b8);
  font-size: 14px;
  background: var(--bg-card, #ffffff);
  border: 1px dashed var(--border-color, #e2e8f0);
  border-radius: var(--border-radius-lg, 12px);
}

.empty-icon {
  font-size: 48px;
  color: #cbd5e1;
}

.location-selected {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.location-tag {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-field {
  display: flex;
  align-items: flex-start;
}

.image-card {
  position: relative;
  width: 160px;
  height: 120px;
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
  border: 1px dashed var(--border-color, #e2e8f0);
  cursor: pointer;
  background: #f1f5f9;
  transition: border-color var(--transition-fast, 150ms);
}

.image-card:hover {
  border-color: var(--color-primary, #4a6cf7);
}

.image-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  color: #ffffff;
  font-size: 13px;
  opacity: 0;
  transition: opacity var(--transition-fast, 150ms);
}

.image-card:hover .image-overlay {
  opacity: 1;
}

.image-clear-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  color: #ffffff;
  background: rgba(0, 0, 0, 0.4);
  border-radius: 50%;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-tertiary, #94a3b8);
}

.image-placeholder-icon {
  font-size: 28px;
}

.image-placeholder-text {
  font-size: 13px;
}

@media (max-width: 1024px) {
  .attractions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .attractions-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (prefers-reduced-motion: reduce) {
  .attraction-card {
    transition: none;
  }

  .attraction-card:hover {
    transform: none;
  }
}
</style>
