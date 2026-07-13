<template>
  <a-drawer
    :open="open"
    :width="drawerWidth"
    placement="right"
    :title="drawerTitle"
    destroy-on-close
    @update:open="handleUpdateOpen"
    @close="handleClose"
  >
    <div v-if="loading" class="travel-detail-loading">
      <a-skeleton active :paragraph="{ rows: 8 }" />
    </div>

    <div v-else-if="guide" class="travel-detail" role="region" :aria-label="guide.title">
      <header class="travel-detail-header">
        <div class="travel-detail-cover">
          <img
            v-if="guide.coverImage"
            :src="guide.coverImage"
            :alt="`${guide.title} 封面`"
          />
          <div v-else class="travel-detail-cover-placeholder">
            <PictureOutlined />
          </div>
        </div>
        <div class="travel-detail-header-body">
          <h2 class="travel-detail-title">{{ guide.title }}</h2>
          <div class="travel-detail-meta">
            <a-tag color="blue">{{ guide.destination }}</a-tag>
            <a-tag>{{ getRegionPath(guide.region) }}</a-tag>
            <a-tag>{{ guide.days }} 天</a-tag>
            <span class="travel-detail-rating">
              <StarFilled class="travel-detail-rating-icon" />
              <span>{{ guide.rating }}</span>
            </span>
            <span class="travel-detail-stats">
              <EyeOutlined /> {{ guide.viewCount }}
            </span>
            <span class="travel-detail-stats">
              <LikeOutlined /> {{ guide.likeCount }}
            </span>
          </div>
        </div>
      </header>

      <section class="travel-detail-section" aria-labelledby="itinerary-title">
        <h3 id="itinerary-title" class="travel-detail-section-title">行程安排</h3>
        <div class="travel-detail-map">
          <EnvironmentOutlined class="travel-detail-map-icon" />
          <span>地图轨迹（占位）</span>
        </div>
        <a-timeline>
          <a-timeline-item v-for="day in guide.itinerary" :key="day.day">
            <div class="travel-detail-day">
              <div class="travel-detail-day-label">第 {{ day.day }} 天</div>
              <div class="travel-detail-day-title">{{ day.title }}</div>
              <div class="travel-detail-day-desc">{{ day.description }}</div>
              <div v-if="getDayAttractions(day).length > 0" class="travel-detail-day-attractions">
                <div
                  v-for="attraction in getDayAttractions(day)"
                  :key="attraction.id"
                  class="travel-detail-day-attraction-card"
                >
                  <div class="travel-detail-day-attraction-cover">
                    <img :src="attraction.image" :alt="attraction.name" />
                  </div>
                  <div class="travel-detail-day-attraction-body">
                    <h4 class="travel-detail-day-attraction-name">{{ attraction.name }}</h4>
                    <div class="travel-detail-day-attraction-meta">
                      <a-tag v-if="attraction.location" size="small">{{ attraction.location }}</a-tag>
                      <span class="travel-detail-day-attraction-duration">
                        <ClockCircleOutlined /> {{ attraction.duration }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="travel-detail-day-empty">当天暂无景点安排</div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </section>

      <section class="travel-detail-section" aria-labelledby="attractions-title">
        <h3 id="attractions-title" class="travel-detail-section-title">景点推荐</h3>
        <div class="travel-detail-attractions">
          <div
            v-for="attraction in guide.attractions"
            :key="attraction.id"
            class="travel-detail-attraction-card"
          >
            <div class="travel-detail-attraction-cover">
              <img :src="attraction.image" :alt="attraction.name" />
            </div>
            <div class="travel-detail-attraction-body">
              <h4 class="travel-detail-attraction-name">{{ attraction.name }}</h4>
              <p class="travel-detail-attraction-desc">{{ attraction.description }}</p>
              <span class="travel-detail-attraction-duration">
                <ClockCircleOutlined /> {{ attraction.duration }}
              </span>
            </div>
          </div>
        </div>
      </section>

      <section class="travel-detail-section" aria-labelledby="reviews-title">
        <h3 id="reviews-title" class="travel-detail-section-title">用户评价</h3>
        <div class="travel-detail-reviews">
          <div v-for="review in guide.reviews" :key="review.id" class="travel-detail-review">
            <img :src="review.avatar" :alt="`${review.username} 头像`" class="travel-detail-review-avatar" />
            <div class="travel-detail-review-body">
              <div class="travel-detail-review-header">
                <span class="travel-detail-review-user">{{ review.username }}</span>
                <span class="travel-detail-review-rating">
                  <StarFilled v-for="n in review.rating" :key="`filled-${review.id}-${n}`" />
                  <StarOutlined v-for="n in 5 - review.rating" :key="`empty-${review.id}-${n}`" />
                </span>
              </div>
              <p class="travel-detail-review-content">{{ review.content }}</p>
              <time class="travel-detail-review-date" :datetime="review.date">
                {{ formatDate(review.date) }}
              </time>
            </div>
          </div>
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
  StarOutlined,
} from '@ant-design/icons-vue'
import { getRegionPath } from '@/types/travel'
import type { TravelGuide, TravelItineraryDay, TravelAttraction } from '@/types/travel'

const props = defineProps<{
  open: boolean
  guide: TravelGuide | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'close'): void
}>()

const loading = ref(false)
let loadingTimer: ReturnType<typeof setTimeout> | null = null

const windowWidth = ref(window.innerWidth)

const drawerWidth = computed(() => (windowWidth.value >= 768 ? 720 : '100%'))

const drawerTitle = computed(() => props.guide?.title || '')

function handleUpdateOpen(value: boolean) {
  emit('update:open', value)
}

function handleClose() {
  emit('close')
}

function onResize() {
  windowWidth.value = window.innerWidth
}

function startLoading() {
  loading.value = true
  if (loadingTimer) clearTimeout(loadingTimer)
  loadingTimer = setTimeout(() => {
    loading.value = false
  }, 300)
}

function formatDate(iso: string) {
  const d = new Date(iso)
  return d.toLocaleDateString('zh-CN')
}

function getDayAttractions(day: TravelItineraryDay): TravelAttraction[] {
  const attractions = props.guide?.attractions ?? []
  return day.attractionIds
    .map((id) => attractions.find((a) => a.id === id))
    .filter((a): a is TravelAttraction => Boolean(a))
}

watch(
  () => props.open,
  (value) => {
    if (value) {
      startLoading()
    } else if (loadingTimer) {
      clearTimeout(loadingTimer)
      loadingTimer = null
      loading.value = false
    }
  },
)

onMounted(() => {
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  if (loadingTimer) clearTimeout(loadingTimer)
})
</script>

<style scoped>
.travel-detail-loading {
  padding: 8px 0;
}

.travel-detail {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.travel-detail-header {
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.travel-detail-cover {
  width: 100%;
  height: 240px;
  background: #f1f5f9;
}

.travel-detail-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.travel-detail-cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-detail-header-body {
  padding: 16px;
}

.travel-detail-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 12px 0;
  line-height: 1.4;
}

.travel-detail-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: var(--text-secondary, #64748b);
}

.travel-detail-rating {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #f59e0b;
  font-weight: 600;
}

.travel-detail-rating-icon {
  color: #f59e0b;
}

.travel-detail-stats {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-detail-section {
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  padding: 16px;
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
}

.travel-detail-section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 16px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
}

.travel-detail-map {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 160px;
  margin-bottom: 16px;
  background: #f1f5f9;
  border-radius: var(--border-radius, 8px);
  color: var(--text-tertiary, #94a3b8);
  font-size: 14px;
}

.travel-detail-map-icon {
  font-size: 28px;
}

.travel-detail-day {
  padding-bottom: 8px;
}

.travel-detail-day-label {
  display: inline-block;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary, #4a6cf7);
  background: var(--color-primary-light, rgba(74, 108, 247, 0.08));
  padding: 2px 8px;
  border-radius: 999px;
  margin-bottom: 6px;
}

.travel-detail-day-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin-bottom: 4px;
}

.travel-detail-day-desc {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
  line-height: 1.6;
  margin-bottom: 12px;
}

.travel-detail-day-attractions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.travel-detail-day-attraction-card {
  display: flex;
  gap: 12px;
  padding: 10px;
  background: var(--bg-secondary, #f8fafc);
  border-radius: var(--border-radius, 8px);
  border: 1px solid var(--border-color, #e2e8f0);
}

.travel-detail-day-attraction-cover {
  width: 72px;
  height: 72px;
  flex-shrink: 0;
  border-radius: var(--border-radius-sm, 6px);
  overflow: hidden;
  background: #f1f5f9;
}

.travel-detail-day-attraction-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.travel-detail-day-attraction-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}

.travel-detail-day-attraction-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0;
}

.travel-detail-day-attraction-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.travel-detail-day-attraction-duration {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.travel-detail-day-empty {
  font-size: 13px;
  color: var(--text-tertiary, #94a3b8);
  padding: 10px;
  background: var(--bg-secondary, #f8fafc);
  border-radius: var(--border-radius, 8px);
  text-align: center;
}

.travel-detail-attractions {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
}

.travel-detail-attraction-card {
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
  transition: box-shadow var(--transition-base, 250ms);
}

.travel-detail-attraction-card:hover {
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0, 0, 0, 0.08));
}

.travel-detail-attraction-cover {
  width: 100%;
  height: 100px;
  background: #f1f5f9;
}

.travel-detail-attraction-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.travel-detail-attraction-body {
  padding: 10px;
}

.travel-detail-attraction-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 6px 0;
}

.travel-detail-attraction-desc {
  font-size: 12px;
  color: var(--text-secondary, #64748b);
  margin: 0 0 8px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
}

.travel-detail-attraction-duration {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.travel-detail-reviews {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.travel-detail-review {
  display: flex;
  gap: 12px;
}

.travel-detail-review-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.travel-detail-review-body {
  flex: 1;
  min-width: 0;
}

.travel-detail-review-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.travel-detail-review-user {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.travel-detail-review-rating {
  display: inline-flex;
  gap: 2px;
  color: #f59e0b;
  font-size: 12px;
}

.travel-detail-review-content {
  font-size: 13px;
  color: var(--text-secondary, #64748b);
  margin: 0 0 6px 0;
  line-height: 1.6;
}

.travel-detail-review-date {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
}
</style>
