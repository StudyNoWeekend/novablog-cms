<template>
  <!-- 桌面端：固定右侧预览面板 -->
  <aside
    v-show="isDesktop"
    class="travel-editor-preview travel-editor-preview--desktop"
    aria-label="实时预览"
  >
    <div class="travel-editor-preview-header">
      <span class="travel-editor-preview-title">实时预览</span>
      <span
        v-if="debouncing"
        class="travel-editor-preview-updating"
        aria-live="polite"
      >
        <SyncOutlined spin />
        预览更新中
      </span>
    </div>
    <div class="travel-editor-preview-body">
      <header class="travel-editor-preview-cover-card">
        <div class="travel-editor-preview-cover">
          <img
            v-if="previewGuide.coverImage"
            :src="previewGuide.coverImage"
            :alt="`${previewGuide.title} 封面`"
          />
          <div v-else class="travel-editor-preview-cover-placeholder">
            <PictureOutlined />
          </div>
        </div>
        <div class="travel-editor-preview-cover-body">
          <h2 class="travel-editor-preview-title-text">{{ previewGuide.title }}</h2>
          <div class="travel-editor-preview-meta">
            <a-tag color="blue">{{ previewGuide.destination }}</a-tag>
            <a-tag>{{ getRegionPath(previewGuide.region) }}</a-tag>
            <a-tag>{{ previewGuide.days }} 天</a-tag>
            <span class="travel-editor-preview-rating">
              <StarFilled class="travel-editor-preview-rating-icon" />
              <span>{{ previewGuide.rating }}</span>
            </span>
            <span class="travel-editor-preview-stats">
              <EyeOutlined /> {{ previewGuide.viewCount }}
            </span>
            <span class="travel-editor-preview-stats">
              <LikeOutlined /> {{ previewGuide.likeCount }}
            </span>
          </div>
        </div>
      </header>

      <section
        class="travel-editor-preview-section"
        aria-labelledby="preview-itinerary-title"
      >
        <h3 id="preview-itinerary-title" class="travel-editor-preview-section-title">
          行程安排
        </h3>
        <a-timeline>
          <a-timeline-item v-for="day in previewGuide.itinerary" :key="day.day">
            <div class="travel-editor-preview-day">
              <div class="travel-editor-preview-day-label">第 {{ day.day }} 天</div>
              <div class="travel-editor-preview-day-title">{{ day.title }}</div>
              <div class="travel-editor-preview-day-desc">{{ day.description }}</div>
              <div
                v-if="getDayAttractions(day).length"
                class="travel-editor-preview-day-attractions"
              >
                <div
                  v-for="attraction in getDayAttractions(day)"
                  :key="attraction.id"
                  class="travel-editor-preview-day-attraction"
                >
                  <div class="travel-editor-preview-day-attraction-cover">
                    <img :src="attraction.image" :alt="attraction.name" />
                  </div>
                  <div class="travel-editor-preview-day-attraction-body">
                    <div class="travel-editor-preview-day-attraction-name">
                      {{ attraction.name }}
                    </div>
                    <div
                      v-if="attraction.location"
                      class="travel-editor-preview-day-attraction-location"
                    >
                      <EnvironmentOutlined />
                      {{ attraction.location }}
                    </div>
                    <div class="travel-editor-preview-day-attraction-duration">
                      <ClockCircleOutlined />
                      {{ attraction.duration }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </section>

      <section
        class="travel-editor-preview-section"
        aria-labelledby="preview-attractions-title"
      >
        <h3 id="preview-attractions-title" class="travel-editor-preview-section-title">
          景点推荐
        </h3>
        <div class="travel-editor-preview-attractions">
          <div
            v-for="attraction in previewGuide.attractions"
            :key="attraction.id"
            class="travel-editor-preview-attraction-card"
          >
            <div class="travel-editor-preview-attraction-cover">
              <img :src="attraction.image" :alt="attraction.name" />
            </div>
            <div class="travel-editor-preview-attraction-body">
              <h4 class="travel-editor-preview-attraction-name">{{ attraction.name }}</h4>
              <p class="travel-editor-preview-attraction-desc">
                {{ attraction.description }}
              </p>
              <div class="travel-editor-preview-attraction-meta">
                <span
                  v-if="attraction.location"
                  class="travel-editor-preview-attraction-location"
                >
                  <EnvironmentOutlined />
                  {{ attraction.location }}
                </span>
                <span class="travel-editor-preview-attraction-duration">
                  <ClockCircleOutlined />
                  {{ attraction.duration }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        class="travel-editor-preview-section"
        aria-labelledby="preview-reviews-title"
      >
        <h3 id="preview-reviews-title" class="travel-editor-preview-section-title">
          用户评价
        </h3>
        <div class="travel-editor-preview-reviews-placeholder">
          评价将在发布后展示
        </div>
      </section>
    </div>
  </aside>

  <!-- 移动端：底部抽屉 -->
  <a-drawer
    :open="visible && !isDesktop"
    placement="bottom"
    title="实时预览"
    height="85%"
    destroy-on-close
    @update:open="handleUpdateVisible"
  >
    <template #title>
      <span>实时预览</span>
      <span
        v-if="debouncing"
        class="travel-editor-preview-updating"
        aria-live="polite"
      >
        <SyncOutlined spin />
        预览更新中
      </span>
    </template>
    <div class="travel-editor-preview-body">
      <header class="travel-editor-preview-cover-card">
        <div class="travel-editor-preview-cover">
          <img
            v-if="previewGuide.coverImage"
            :src="previewGuide.coverImage"
            :alt="`${previewGuide.title} 封面`"
          />
          <div v-else class="travel-editor-preview-cover-placeholder">
            <PictureOutlined />
          </div>
        </div>
        <div class="travel-editor-preview-cover-body">
          <h2 class="travel-editor-preview-title-text">{{ previewGuide.title }}</h2>
          <div class="travel-editor-preview-meta">
            <a-tag color="blue">{{ previewGuide.destination }}</a-tag>
            <a-tag>{{ getRegionPath(previewGuide.region) }}</a-tag>
            <a-tag>{{ previewGuide.days }} 天</a-tag>
            <span class="travel-editor-preview-rating">
              <StarFilled class="travel-editor-preview-rating-icon" />
              <span>{{ previewGuide.rating }}</span>
            </span>
            <span class="travel-editor-preview-stats">
              <EyeOutlined /> {{ previewGuide.viewCount }}
            </span>
            <span class="travel-editor-preview-stats">
              <LikeOutlined /> {{ previewGuide.likeCount }}
            </span>
          </div>
        </div>
      </header>

      <section
        class="travel-editor-preview-section"
        aria-labelledby="preview-itinerary-title-mobile"
      >
        <h3
          id="preview-itinerary-title-mobile"
          class="travel-editor-preview-section-title"
        >
          行程安排
        </h3>
        <a-timeline>
          <a-timeline-item v-for="day in previewGuide.itinerary" :key="day.day">
            <div class="travel-editor-preview-day">
              <div class="travel-editor-preview-day-label">第 {{ day.day }} 天</div>
              <div class="travel-editor-preview-day-title">{{ day.title }}</div>
              <div class="travel-editor-preview-day-desc">{{ day.description }}</div>
              <div
                v-if="getDayAttractions(day).length"
                class="travel-editor-preview-day-attractions"
              >
                <div
                  v-for="attraction in getDayAttractions(day)"
                  :key="attraction.id"
                  class="travel-editor-preview-day-attraction"
                >
                  <div class="travel-editor-preview-day-attraction-cover">
                    <img :src="attraction.image" :alt="attraction.name" />
                  </div>
                  <div class="travel-editor-preview-day-attraction-body">
                    <div class="travel-editor-preview-day-attraction-name">
                      {{ attraction.name }}
                    </div>
                    <div
                      v-if="attraction.location"
                      class="travel-editor-preview-day-attraction-location"
                    >
                      <EnvironmentOutlined />
                      {{ attraction.location }}
                    </div>
                    <div class="travel-editor-preview-day-attraction-duration">
                      <ClockCircleOutlined />
                      {{ attraction.duration }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </section>

      <section
        class="travel-editor-preview-section"
        aria-labelledby="preview-attractions-title-mobile"
      >
        <h3
          id="preview-attractions-title-mobile"
          class="travel-editor-preview-section-title"
        >
          景点推荐
        </h3>
        <div class="travel-editor-preview-attractions">
          <div
            v-for="attraction in previewGuide.attractions"
            :key="attraction.id"
            class="travel-editor-preview-attraction-card"
          >
            <div class="travel-editor-preview-attraction-cover">
              <img :src="attraction.image" :alt="attraction.name" />
            </div>
            <div class="travel-editor-preview-attraction-body">
              <h4 class="travel-editor-preview-attraction-name">{{ attraction.name }}</h4>
              <p class="travel-editor-preview-attraction-desc">
                {{ attraction.description }}
              </p>
              <div class="travel-editor-preview-attraction-meta">
                <span
                  v-if="attraction.location"
                  class="travel-editor-preview-attraction-location"
                >
                  <EnvironmentOutlined />
                  {{ attraction.location }}
                </span>
                <span class="travel-editor-preview-attraction-duration">
                  <ClockCircleOutlined />
                  {{ attraction.duration }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        class="travel-editor-preview-section"
        aria-labelledby="preview-reviews-title-mobile"
      >
        <h3
          id="preview-reviews-title-mobile"
          class="travel-editor-preview-section-title"
        >
          用户评价
        </h3>
        <div class="travel-editor-preview-reviews-placeholder">
          评价将在发布后展示
        </div>
      </section>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import {
  ClockCircleOutlined,
  EnvironmentOutlined,
  EyeOutlined,
  LikeOutlined,
  PictureOutlined,
  StarFilled,
  SyncOutlined,
} from '@ant-design/icons-vue'
import { getRegionPath, toTravelGuide } from '@/types/travel'
import type {
  TravelAttraction,
  TravelGuide,
  TravelGuideFormData,
  TravelItineraryDay,
} from '@/types/travel'

const props = defineProps<{
  guide: TravelGuideFormData
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const previewGuide = ref<TravelGuide>(toTravelGuide(props.guide))
const debouncing = ref(false)
const windowWidth = ref(window.innerWidth)

let updateTimer: ReturnType<typeof setTimeout> | null = null

const isDesktop = computed(() => windowWidth.value >= 768)

function handleUpdateVisible(value: boolean) {
  emit('update:visible', value)
}

function onResize() {
  windowWidth.value = window.innerWidth
}

function getDayAttractions(day: TravelItineraryDay): TravelAttraction[] {
  if (!day.attractionIds?.length) return []
  return previewGuide.value.attractions.filter((attraction) =>
    day.attractionIds.includes(attraction.id),
  )
}

function scheduleUpdate() {
  debouncing.value = true
  if (updateTimer) {
    clearTimeout(updateTimer)
  }
  updateTimer = setTimeout(() => {
    previewGuide.value = toTravelGuide(props.guide)
    debouncing.value = false
  }, 200)
}

watch(() => props.guide, scheduleUpdate, { deep: true })

onMounted(() => {
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  if (updateTimer) {
    clearTimeout(updateTimer)
  }
})
</script>

<style scoped>
.travel-editor-preview--desktop {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 420px;
  max-width: 40vw;
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px) 0 0 var(--border-radius-lg, 12px);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
  overflow: hidden;
  z-index: 50;
}

.travel-editor-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
  background: var(--bg-card, #ffffff);
}

.travel-editor-preview-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.travel-editor-preview-updating {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--color-primary, #4a6cf7);
}

.travel-editor-preview-body {
  flex: 1;
  min-height: 0;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.travel-editor-preview-cover-card {
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.travel-editor-preview-cover {
  width: 100%;
  height: 180px;
  background: #f1f5f9;
}

.travel-editor-preview-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.travel-editor-preview-cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-editor-preview-cover-body {
  padding: 14px;
}

.travel-editor-preview-title-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 10px 0;
  line-height: 1.4;
}

.travel-editor-preview-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary, #64748b);
}

.travel-editor-preview-rating {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #f59e0b;
  font-weight: 600;
}

.travel-editor-preview-rating-icon {
  color: #f59e0b;
}

.travel-editor-preview-stats {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-editor-preview-section {
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  padding: 14px;
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.travel-editor-preview-section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 14px 0;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
}

.travel-editor-preview-day {
  padding-bottom: 6px;
}

.travel-editor-preview-day-label {
  display: inline-block;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary, #4a6cf7);
  background: var(--color-primary-light, rgba(74, 108, 247, 0.08));
  padding: 2px 8px;
  border-radius: 999px;
  margin-bottom: 6px;
}

.travel-editor-preview-day-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin-bottom: 4px;
}

.travel-editor-preview-day-desc {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
  line-height: 1.6;
  margin-bottom: 10px;
}

.travel-editor-preview-day-attractions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.travel-editor-preview-day-attraction {
  display: flex;
  gap: 10px;
  padding: 8px;
  background: var(--bg-secondary, #f8fafc);
  border-radius: var(--border-radius, 8px);
}

.travel-editor-preview-day-attraction-cover {
  width: 56px;
  height: 56px;
  flex-shrink: 0;
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
  background: #f1f5f9;
}

.travel-editor-preview-day-attraction-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.travel-editor-preview-day-attraction-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
}

.travel-editor-preview-day-attraction-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.travel-editor-preview-day-attraction-location,
.travel-editor-preview-day-attraction-duration {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.travel-editor-preview-attractions {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.travel-editor-preview-attraction-card {
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
  transition: box-shadow var(--transition-base, 250ms);
}

.travel-editor-preview-attraction-card:hover {
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0, 0, 0, 0.08));
}

.travel-editor-preview-attraction-cover {
  width: 100%;
  height: 90px;
  background: #f1f5f9;
}

.travel-editor-preview-attraction-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.travel-editor-preview-attraction-body {
  padding: 10px;
}

.travel-editor-preview-attraction-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 6px 0;
}

.travel-editor-preview-attraction-desc {
  font-size: 12px;
  color: var(--text-secondary, #64748b);
  margin: 0 0 8px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
}

.travel-editor-preview-attraction-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-editor-preview-attraction-location,
.travel-editor-preview-attraction-duration {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.travel-editor-preview-reviews-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100px;
  font-size: 13px;
  color: var(--text-tertiary, #94a3b8);
  background: var(--bg-secondary, #f8fafc);
  border-radius: var(--border-radius, 8px);
}

@media (max-width: 767px) {
  .travel-editor-preview--desktop {
    display: none;
  }
}
</style>
