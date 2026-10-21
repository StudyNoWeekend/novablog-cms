<template>
  <div class="travel-attraction-pool">
    <div class="section-header">
      <div class="section-title-wrap">
        <span class="section-title-bar" aria-hidden="true" />
        <h2 class="section-title">{{ $t('travel.pool.title') }}</h2>
      </div>
      <p class="section-subtitle">{{ $t('travel.pool.subtitle') }}</p>
    </div>

    <a-button
      type="primary"
      class="add-button"
      :aria-label="$t('travel.actions.addAttraction')"
      @click="openAddModal"
    >
      <PlusOutlined />
      {{ $t('travel.actions.addAttraction') }}
    </a-button>

    <div
      v-if="attractions.length > 0"
      class="attractions-grid"
      role="list"
      :aria-label="$t('travel.pool.listAria')"
    >
      <div
        v-for="attraction in attractions"
        :key="attraction.id"
        class="attraction-card"
        role="listitem"
        draggable="true"
        tabindex="0"
        :aria-label="$t('travel.pool.attractionAria', { name: attraction.name })"
        @dragstart="handleDragStart($event, attraction)"
      >
        <div class="attraction-thumbnail">
          <img
            v-if="attraction.image"
            :src="attraction.image"
            :alt="$t('travel.pool.imageAlt', { name: attraction.name })"
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
              {{ $t('travel.pool.dayRef', { day }) }}
            </a-tag>
          </div>
        </div>

        <div class="attraction-actions">
          <a-button
            type="text"
            size="small"
            :aria-label="$t('travel.actions.editAttraction')"
            @click.stop="openEdit(attraction)"
          >
            <EditOutlined />
          </a-button>
          <a-button
            type="text"
            size="small"
            danger
            :aria-label="$t('travel.actions.deleteAttraction')"
            @click.stop="handleDelete(attraction)"
          >
            <DeleteOutlined />
          </a-button>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <InboxOutlined class="empty-icon" />
      <p>{{ $t('travel.pool.empty') }}</p>
    </div>

    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :ok-text="isEditing ? $t('common.actions.save') : $t('common.actions.add')"
      :cancel-text="$t('common.actions.cancel')"
      destroy-on-close
      @ok="handleModalOk"
      @cancel="closeModal"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item :label="$t('common.field.name')" required>
          <a-input
            v-model:value="form.name"
            :placeholder="$t('travel.pool.namePlaceholder')"
            maxlength="100"
            show-count
          />
        </a-form-item>

        <a-form-item :label="$t('common.field.description')">
          <a-textarea
            v-model:value="form.description"
            :placeholder="$t('travel.pool.descriptionPlaceholder')"
            :rows="3"
            maxlength="500"
            show-count
          />
        </a-form-item>

        <a-form-item :label="$t('common.field.image')">
          <div class="image-field">
            <div
              class="image-card"
              role="button"
              tabindex="0"
              :aria-label="$t('travel.pool.selectImageAria')"
              @click="imagePickerVisible = true"
              @keydown.enter="imagePickerVisible = true"
            >
              <template v-if="form.image">
                <img :src="form.image" :alt="$t('travel.pool.attractionImageAlt')" />
                <div class="image-overlay"><span>{{ $t('travel.pool.changeImage') }}</span></div>
                <a-button
                  type="text"
                  size="small"
                  class="image-clear-btn"
                  :aria-label="$t('travel.pool.removeImageAria')"
                  @click.stop="form.image = ''"
                >
                  <CloseOutlined />
                </a-button>
              </template>
              <template v-else>
                <div class="image-placeholder">
                  <PictureOutlined class="image-placeholder-icon" />
                  <span class="image-placeholder-text">{{ $t('travel.pool.selectImageText') }}</span>
                </div>
              </template>
            </div>
          </div>
          <TravelCoverPicker
            v-model:visible="imagePickerVisible"
            @selected="onImageSelected"
          />
        </a-form-item>

        <a-form-item :label="$t('travel.pool.durationLabel')" required>
          <a-input
            v-model:value="form.duration"
            :placeholder="$t('travel.pool.durationPlaceholder')"
          />
        </a-form-item>

        <a-form-item :label="$t('travel.pool.locationLabel')" required>
          <div v-if="form.location" class="location-selected">
            <a-tag color="purple" class="location-tag">
              <EnvironmentOutlined /> {{ form.location }}
            </a-tag>
            <a-button type="link" size="small" @click="openLocationPicker">
              {{ $t('travel.pool.reselect') }}
            </a-button>
          </div>
          <a-button
            v-else
            type="dashed"
            block
            @click="openLocationPicker"
          >
            <EnvironmentOutlined /> {{ $t('travel.pool.selectLocationRequired') }}
          </a-button>
        </a-form-item>

        <a-form-item :label="$t('common.field.remark')">
          <a-textarea
            v-model:value="form.notes"
            :placeholder="$t('travel.pool.notesPlaceholder')"
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
import { useI18n } from 'vue-i18n'
import { message, Modal } from 'ant-design-vue'
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

const { t } = useI18n()

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
  coordType: undefined,
  notes: '',
})

const form = reactive<AttractionFormData>(emptyForm())

const isEditing = computed(() => editingId.value !== null)
// 弹窗标题（跟随当前语言）
const modalTitle = computed(() => (isEditing.value ? t('travel.actions.editAttraction') : t('travel.actions.addAttraction')))
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
    // 给出可见反馈，避免静默失败让用户误以为已添加
    message.warning(t('travel.pool.incompleteWarning'))
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
    coordType: form.latitude != null ? form.coordType : undefined,
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
      title: t('travel.delete.confirmTitle'),
      content: t('travel.delete.attractionInUse', { days: refs.join(t('travel.pool.daySeparator')) }),
      okText: t('common.actions.delete'),
      okType: 'danger',
      cancelText: t('common.actions.cancel'),
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
  form.coordType = location.coordType
  locationPickerVisible.value = false
}

function onImageSelected(payload: { url: string }) {
  form.image = payload.url
  imagePickerVisible.value = false
}

function locationTitle(attraction: TravelAttraction): string {
  if (attraction.latitude != null && attraction.longitude != null) {
    return `${attraction.location} (${t('travel.pool.coordsTitle', {
      lng: attraction.longitude.toFixed(6),
      lat: attraction.latitude.toFixed(6),
    })})`
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
  background: var(--color-primary, #526FE8);
}

.section-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #29365C);
}

.section-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary, #8A93A8);
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
  border: 1px solid var(--border-color, #E5E9F2);
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
  outline: 2px solid var(--color-primary, #526FE8);
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
  background: #F5F7FC;
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
  color: #8A93A8;
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
  color: var(--text-primary, #29365C);
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
  background: var(--color-primary, #526FE8);
  color: #ffffff;
  border-color: transparent;
}

.attraction-duration {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-secondary, #667085);
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
  color: var(--text-tertiary, #8A93A8);
  font-size: 14px;
  background: var(--bg-card, #ffffff);
  border: 1px dashed var(--border-color, #E5E9F2);
  border-radius: var(--border-radius-lg, 12px);
}

.empty-icon {
  font-size: 48px;
  color: #8A93A8;
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
  border: 1px dashed var(--border-color, #E5E9F2);
  cursor: pointer;
  background: #F5F7FC;
  transition: border-color var(--transition-fast, 150ms);
}

.image-card:hover {
  border-color: var(--color-primary, #526FE8);
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
  color: var(--text-tertiary, #8A93A8);
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
