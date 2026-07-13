<template>
  <div class="travel-itinerary-builder">
    <div class="section-header">
      <div class="section-title-wrap">
        <span class="section-title-bar" aria-hidden="true" />
        <h2 class="section-title">行程安排</h2>
      </div>
    </div>

    <a-timeline v-if="guide.itinerary.length > 0" class="itinerary-timeline">
      <a-timeline-item
        v-for="(day, dayIndex) in guide.itinerary"
        :key="day.day"
      >
        <template #dot>
          <div class="day-badge">D{{ day.day }}</div>
        </template>

        <div
          class="day-card"
          :class="{ 'drag-over': dragOverDayIndex === dayIndex }"
          @dragover.prevent="handleDayDragOver(dayIndex)"
          @dragleave="handleDayDragLeave"
          @drop.prevent="handleDayDrop($event, dayIndex)"
        >
          <div class="day-field">
            <label :for="titleId(dayIndex)" class="day-label">
              第 {{ day.day }} 天标题
            </label>
            <a-input
              :id="titleId(dayIndex)"
              :value="day.title"
              placeholder="请输入当天标题"
              @update:value="updateDayTitle(dayIndex, $event)"
            />
          </div>

          <div class="day-field">
            <label :for="descId(dayIndex)" class="day-label">
              行程描述
            </label>
            <a-textarea
              :id="descId(dayIndex)"
              :value="day.description"
              :rows="2"
              placeholder="请输入当天行程描述"
              @update:value="updateDayDescription(dayIndex, $event)"
            />
          </div>

          <div class="day-attractions">
            <div class="attractions-toolbar">
              <span class="attractions-count">
                景点 {{ day.attractionIds.length }}
              </span>
              <a-button
                type="primary"
                size="small"
                @click="openAddAttraction(dayIndex)"
              >
                <PlusOutlined />
                添加景点
              </a-button>
            </div>

            <div
              v-if="day.attractionIds.length === 0"
              class="empty-state"
            >
              暂无景点，点击“添加景点”或将景点拖拽到此处
            </div>

            <div
              v-else
              class="attractions-list"
              role="list"
              :aria-label="`第 ${day.day} 天景点列表`"
            >
              <div
                v-for="(attractionId, index) in day.attractionIds"
                :key="attractionId"
                class="attraction-mini-card"
                :class="{
                  'drag-over':
                    dragOverAttraction?.dayIndex === dayIndex &&
                    dragOverAttraction?.index === index,
                }"
                role="listitem"
                draggable="true"
                @dragstart="handleAttractionDragStart($event, dayIndex, index, attractionId)"
                @dragover.prevent.stop="handleAttractionDragOver(dayIndex, index)"
                @dragleave="handleAttractionDragLeave"
                @drop.prevent.stop="handleAttractionDrop($event, dayIndex, index)"
                @dragend="handleAttractionDragEnd"
              >
                <img
                  v-if="attractionImage(attractionId)"
                  :src="attractionImage(attractionId)"
                  :alt="attractionName(attractionId)"
                  class="attraction-image"
                />
                <div
                  v-else
                  class="attraction-image attraction-image-placeholder"
                >
                  <PictureOutlined />
                </div>

                <div class="attraction-info">
                  <div class="attraction-name" :title="attractionName(attractionId)">
                    {{ attractionName(attractionId) }}
                  </div>
                  <div class="attraction-meta">
                    <a-tag
                      v-if="attractionLocation(attractionId)"
                      class="attraction-location-tag"
                      :title="attractionLocation(attractionId)!"
                    >
                      <EnvironmentOutlined />
                      {{ attractionLocation(attractionId) }}
                    </a-tag>
                    <span
                      v-if="attractionDuration(attractionId)"
                      class="attraction-duration"
                    >
                      <ClockCircleOutlined />
                      {{ attractionDuration(attractionId) }}
                    </span>
                  </div>
                </div>

                <div class="attraction-actions">
                  <a-button
                    type="text"
                    size="small"
                    :disabled="index === 0"
                    aria-label="上移"
                    @click="moveAttraction(dayIndex, index, index - 1)"
                  >
                    <UpOutlined />
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    :disabled="index === day.attractionIds.length - 1"
                    aria-label="下移"
                    @click="moveAttraction(dayIndex, index, index + 1)"
                  >
                    <DownOutlined />
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    danger
                    aria-label="移除景点"
                    @click="removeAttraction(dayIndex, index)"
                  >
                    <CloseOutlined />
                  </a-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </a-timeline-item>
    </a-timeline>

    <a-empty v-else description="暂无行程天数" />

    <a-modal
      v-model:open="modalOpen"
      title="添加景点"
      :footer="null"
      width="640px"
    >
      <a-empty
        v-if="modalAttractions.length === 0"
        description="没有可添加的景点"
      />
      <div v-else class="modal-attraction-list">
        <div
          v-for="attraction in modalAttractions"
          :key="attraction.id"
          class="modal-attraction-item"
          @click="selectAttraction(attraction.id)"
        >
          <img
            v-if="attraction.image"
            :src="attraction.image"
            :alt="attraction.name"
            class="modal-attraction-image"
          />
          <div
            v-else
            class="modal-attraction-image modal-attraction-image-placeholder"
          >
            <PictureOutlined />
          </div>
          <div class="modal-attraction-info">
            <div class="modal-attraction-name" :title="attraction.name">
              {{ attraction.name }}
            </div>
            <div
              v-if="attraction.location"
              class="modal-attraction-location"
              :title="attraction.location"
            >
              <EnvironmentOutlined />
              {{ attraction.location }}
            </div>
          </div>
          <a-button type="primary" size="small">
            <PlusOutlined />
          </a-button>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  PlusOutlined,
  CloseOutlined,
  UpOutlined,
  DownOutlined,
  ClockCircleOutlined,
  EnvironmentOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue'
import type {
  TravelGuideFormData,
  TravelItineraryDay,
  TravelAttraction,
} from '@/types/travel'

const props = defineProps<{
  modelValue: TravelGuideFormData
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: TravelGuideFormData): void
}>()

const guide = computed(() => props.modelValue)

function emitUpdate(next: TravelGuideFormData) {
  emit('update:modelValue', next)
}

function updateItinerary(nextItinerary: TravelItineraryDay[]) {
  emitUpdate({ ...guide.value, itinerary: nextItinerary })
}

function updateDay(dayIndex: number, patch: Partial<TravelItineraryDay>) {
  const next = [...guide.value.itinerary]
  next[dayIndex] = { ...next[dayIndex], ...patch }
  updateItinerary(next)
}

function updateDayTitle(dayIndex: number, title: string) {
  updateDay(dayIndex, { title })
}

function updateDayDescription(dayIndex: number, description: string) {
  updateDay(dayIndex, { description })
}

function attractionById(id: string): TravelAttraction | undefined {
  return guide.value.attractions.find((item) => item.id === id)
}

function attractionsByIds(ids: string[]): TravelAttraction[] {
  return ids
    .map((id) => attractionById(id))
    .filter((item): item is TravelAttraction => item !== undefined)
}

function setDayAttractionIds(dayIndex: number, ids: string[]) {
  updateDay(dayIndex, {
    attractionIds: ids,
    attractions: attractionsByIds(ids),
  })
}

// 根据 days 自动扩展/缩减 itinerary
watch(
  () => guide.value.days,
  (days) => {
    const current = guide.value.itinerary
    if (days > current.length) {
      const added: TravelItineraryDay[] = []
      for (let d = current.length + 1; d <= days; d++) {
        added.push({
          day: d,
          title: '',
          description: '',
          attractions: [],
          attractionIds: [],
        })
      }
      updateItinerary([...current, ...added])
    } else if (days < current.length) {
      updateItinerary(current.slice(0, days))
    }
  },
  { immediate: true },
)

// 与 attractions 池保持同步：移除已不存在的景点引用
watch(
  () => guide.value.attractions,
  (attractions) => {
    const validIds = new Set(attractions.map((item) => item.id))
    let changed = false
    const next = guide.value.itinerary.map((day) => {
      const filtered = day.attractionIds.filter((id) => validIds.has(id))
      if (filtered.length !== day.attractionIds.length) {
        changed = true
      }
      return {
        ...day,
        attractionIds: filtered,
        attractions: attractionsByIds(filtered),
      }
    })
    if (changed) {
      updateItinerary(next)
    }
  },
  { deep: true, immediate: true },
)

// 添加景点弹窗
const modalOpen = ref(false)
const selectedDayIndex = ref<number>(-1)

const modalAttractions = computed(() => {
  if (selectedDayIndex.value < 0) return []
  const day = guide.value.itinerary[selectedDayIndex.value]
  return guide.value.attractions.filter(
    (item) => !day.attractionIds.includes(item.id),
  )
})

function openAddAttraction(dayIndex: number) {
  selectedDayIndex.value = dayIndex
  modalOpen.value = true
}

function selectAttraction(attractionId: string) {
  if (selectedDayIndex.value < 0) return
  const day = guide.value.itinerary[selectedDayIndex.value]
  if (day.attractionIds.includes(attractionId)) return
  setDayAttractionIds(selectedDayIndex.value, [
    ...day.attractionIds,
    attractionId,
  ])
  modalOpen.value = false
}

function removeAttraction(dayIndex: number, index: number) {
  const day = guide.value.itinerary[dayIndex]
  const nextIds = [...day.attractionIds]
  nextIds.splice(index, 1)
  setDayAttractionIds(dayIndex, nextIds)
}

function moveAttraction(
  dayIndex: number,
  fromIndex: number,
  toIndex: number,
) {
  const day = guide.value.itinerary[dayIndex]
  if (toIndex < 0 || toIndex >= day.attractionIds.length) return
  const nextIds = [...day.attractionIds]
  const [moved] = nextIds.splice(fromIndex, 1)
  nextIds.splice(toIndex, 0, moved)
  setDayAttractionIds(dayIndex, nextIds)
}

// 拖拽：将景点拖入某天
const dragOverDayIndex = ref<number | null>(null)

function handleDayDragOver(dayIndex: number) {
  dragOverDayIndex.value = dayIndex
}

function handleDayDragLeave() {
  dragOverDayIndex.value = null
}

function handleDayDrop(e: DragEvent, dayIndex: number) {
  dragOverDayIndex.value = null
  const id = e.dataTransfer?.getData('text/plain')?.trim()
  if (!id) return
  const day = guide.value.itinerary[dayIndex]
  if (day.attractionIds.includes(id)) return
  if (!guide.value.attractions.some((item) => item.id === id)) return
  setDayAttractionIds(dayIndex, [...day.attractionIds, id])
}

// 拖拽：天内排序
const dragSource = ref<{ dayIndex: number; index: number } | null>(null)
const dragOverAttraction = ref<{ dayIndex: number; index: number } | null>(null)

function handleAttractionDragStart(
  e: DragEvent,
  dayIndex: number,
  index: number,
  attractionId: string,
) {
  dragSource.value = { dayIndex, index }
  if (e.dataTransfer) {
    e.dataTransfer.setData('text/plain', attractionId)
    e.dataTransfer.effectAllowed = 'move'
  }
}

function handleAttractionDragOver(dayIndex: number, index: number) {
  dragOverAttraction.value = { dayIndex, index }
}

function handleAttractionDragLeave() {
  dragOverAttraction.value = null
}

function handleAttractionDrop(
  _e: DragEvent,
  dayIndex: number,
  index: number,
) {
  dragOverAttraction.value = null
  if (!dragSource.value) return
  const { dayIndex: fromDay, index: fromIndex } = dragSource.value
  if (fromDay !== dayIndex) return
  moveAttraction(dayIndex, fromIndex, index)
}

function handleAttractionDragEnd() {
  dragSource.value = null
  dragOverAttraction.value = null
}

function titleId(dayIndex: number) {
  return `day-${guide.value.itinerary[dayIndex]?.day}-title`
}

function descId(dayIndex: number) {
  return `day-${guide.value.itinerary[dayIndex]?.day}-desc`
}

function attractionName(id: string) {
  return attractionById(id)?.name ?? '未知景点'
}

function attractionImage(id: string) {
  return attractionById(id)?.image ?? ''
}

function attractionLocation(id: string) {
  return attractionById(id)?.location
}

function attractionDuration(id: string) {
  return attractionById(id)?.duration ?? ''
}
</script>

<style scoped>
.travel-itinerary-builder {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  align-items: center;
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

.itinerary-timeline {
  padding-top: 8px;
}

.day-badge {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--color-primary, #4a6cf7);
  color: #ffffff;
  font-size: 12px;
  font-weight: 700;
}

.day-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
  transition:
    border-color var(--transition-fast, 150ms cubic-bezier(0.4, 0, 0.2, 1)),
    box-shadow var(--transition-fast, 150ms cubic-bezier(0.4, 0, 0.2, 1));
}

.day-card.drag-over {
  box-shadow: 0 0 0 2px var(--color-primary, #4a6cf7),
    var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.day-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.day-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary, #64748b);
}

.day-attractions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.attractions-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.attractions-count {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
}

.empty-state {
  padding: 24px 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary, #94a3b8);
  background: var(--bg-page, #f8f9fb);
  border: 1px dashed var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
}

.attractions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.attraction-mini-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: var(--bg-page, #f8f9fb);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  cursor: grab;
  transition:
    border-color var(--transition-fast, 150ms cubic-bezier(0.4, 0, 0.2, 1)),
    box-shadow var(--transition-fast, 150ms cubic-bezier(0.4, 0, 0.2, 1));
}

.attraction-mini-card:hover {
  border-color: var(--color-primary, #4a6cf7);
}

.attraction-mini-card.drag-over {
  border-color: var(--color-primary, #4a6cf7);
  border-style: dashed;
  box-shadow: 0 0 0 2px var(--color-primary-light, rgba(74, 108, 247, 0.08));
}

.attraction-mini-card:active {
  cursor: grabbing;
}

.attraction-image {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: 6px;
  object-fit: cover;
  background: #f1f5f9;
}

.attraction-image-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #cbd5e1;
}

.attraction-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.attraction-name {
  font-size: 14px;
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

.attraction-location-tag {
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

.attraction-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.modal-attraction-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 480px;
  overflow-y: auto;
}

.modal-attraction-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: var(--bg-page, #f8f9fb);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  cursor: pointer;
  transition: border-color var(--transition-fast, 150ms cubic-bezier(0.4, 0, 0.2, 1));
}

.modal-attraction-item:hover {
  border-color: var(--color-primary, #4a6cf7);
}

.modal-attraction-image {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: cover;
  background: #f1f5f9;
}

.modal-attraction-image-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #cbd5e1;
}

.modal-attraction-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.modal-attraction-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.modal-attraction-location {
  font-size: 12px;
  color: var(--text-secondary, #64748b);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (max-width: 640px) {
  .attraction-mini-card {
    flex-wrap: wrap;
  }

  .attraction-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (prefers-reduced-motion: reduce) {
  .day-card,
  .attraction-mini-card {
    transition: none;
  }
}
</style>
