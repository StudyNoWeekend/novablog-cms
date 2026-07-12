<template>
  <div
    class="travel-card"
    :style="cardStyle"
    role="article"
    tabindex="0"
    @click="handleView"
    @keydown.enter="handleView"
  >
    <div class="travel-card-cover">
      <img
        v-if="guide.coverImage"
        :src="getThumbUrl(guide.coverImage, 400)"
        :alt="`${guide.title} 封面图`"
        loading="lazy"
      />
      <div v-else class="travel-card-cover-placeholder">
        <CompassOutlined />
      </div>
      <a-tag :color="statusColor" class="travel-card-status">{{ statusText }}</a-tag>
      <div class="travel-card-cover-overlay" aria-hidden="true">
        <span>查看详情</span>
      </div>
    </div>

    <div class="travel-card-body">
      <h3 class="travel-card-title">{{ guide.title }}</h3>
      <p class="travel-card-summary">{{ guide.summary }}</p>
      <div class="travel-card-meta">
        <span class="travel-card-meta-item" :title="getRegionPath(guide.region)">
          <CompassOutlined /> {{ getRegionPath(guide.region) }}
        </span>
        <span v-if="guide.categoryName" class="travel-card-meta-item">
          <TagsOutlined /> {{ guide.categoryName }}
        </span>
        <span class="travel-card-meta-item" :title="guide.destination">
          <EnvironmentOutlined /> {{ guide.destination }}
        </span>
        <span class="travel-card-meta-item">
          <CalendarOutlined /> {{ guide.days }}天
        </span>
        <span class="travel-card-meta-item" :title="guide.bestMonth">
          <ClockCircleOutlined /> {{ guide.bestMonth }}
        </span>
        <span class="travel-card-meta-item">
          <EyeOutlined /> {{ guide.viewCount }}
        </span>
        <span class="travel-card-meta-item">
          <LikeOutlined /> {{ guide.likeCount }}
        </span>
        <span class="travel-card-meta-item">
          <StarFilled class="travel-card-rating-icon" /> {{ guide.rating }}
        </span>
      </div>
    </div>

    <div class="travel-card-footer">
      <div class="travel-card-rating">
        <span class="travel-card-rating-score">{{ guide.rating }}</span>
        <StarFilled class="travel-card-rating-icon" />
        <span class="travel-card-review-count">({{ guide.reviewCount }}条评价)</span>
      </div>
      <div class="travel-card-actions" @click.stop>
        <a-button
          size="small"
          type="text"
          aria-label="编辑"
          @click="handleEdit"
        >
          <EditOutlined />
        </a-button>
        <a-dropdown :trigger="['click']">
          <a-button size="small" type="text" aria-label="更多操作">
            <MoreOutlined />
          </a-button>
          <template #overlay>
            <a-menu>
              <a-menu-item
                v-if="guide.status === TravelStatus.Draft"
                @click="handleStatusChange(TravelStatus.Published)"
              >
                <SendOutlined /> 发布
              </a-menu-item>
              <a-menu-item
                v-if="guide.status === TravelStatus.Published"
                @click="handleStatusChange(TravelStatus.Archived)"
              >
                <StopOutlined /> 下架
              </a-menu-item>
              <a-menu-item
                v-if="guide.status !== TravelStatus.Draft"
                @click="handleStatusChange(TravelStatus.Draft)"
              >
                <FileTextOutlined /> 转为草稿
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item danger @click="handleDelete">
                <DeleteOutlined /> 删除
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  CompassOutlined,
  EnvironmentOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  EyeOutlined,
  LikeOutlined,
  StarFilled,
  EditOutlined,
  MoreOutlined,
  SendOutlined,
  StopOutlined,
  FileTextOutlined,
  DeleteOutlined,
  TagsOutlined,
} from '@ant-design/icons-vue'
import { TravelStatus, getRegionPath, type TravelGuide, type TravelStatus as TTravelStatus } from '@/types/travel'
import { getThumbUrl } from '@/utils/image'

const props = withDefaults(
  defineProps<{
    guide: TravelGuide
    index?: number
  }>(),
  {
    index: 0,
  },
)

const emit = defineEmits<{
  (e: 'view', id: string): void
  (e: 'edit', id: string): void
  (e: 'delete', id: string): void
  (e: 'statusChange', id: string, status: TTravelStatus): void
}>()

const statusMap: Record<TTravelStatus, { text: string; color: string }> = {
  [TravelStatus.Draft]: { text: '草稿', color: 'default' },
  [TravelStatus.Published]: { text: '已发布', color: 'green' },
  [TravelStatus.Archived]: { text: '已下架', color: 'orange' },
}

const statusText = computed(() => statusMap[props.guide.status]?.text ?? '未知')
const statusColor = computed(() => statusMap[props.guide.status]?.color ?? 'default')

const cardStyle = computed(() => ({
  '--delay': `${props.index * 50}ms`,
}))

function handleView() {
  emit('view', props.guide.id)
}

function handleEdit() {
  emit('edit', props.guide.id)
}

function handleDelete() {
  emit('delete', props.guide.id)
}

function handleStatusChange(status: TTravelStatus) {
  emit('statusChange', props.guide.id, status)
}
</script>

<style scoped>
.travel-card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #ffffff);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  cursor: pointer;
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
  transition:
    transform var(--transition-base, 250ms cubic-bezier(0.4, 0, 0.2, 1)),
    box-shadow var(--transition-base, 250ms cubic-bezier(0.4, 0, 0.2, 1));
  animation: travel-card-fade-in-up 400ms cubic-bezier(0.4, 0, 0.2, 1) forwards;
  animation-delay: var(--delay, 0ms);
  opacity: 0;
}

.travel-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0, 0, 0, 0.08));
}

.travel-card:focus-visible {
  outline: 2px solid var(--color-primary, #4a6cf7);
  outline-offset: 2px;
}

.travel-card-cover {
  position: relative;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  background: #f1f5f9;
}

.travel-card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.travel-card:hover .travel-card-cover img,
.travel-card:hover .travel-card-cover-placeholder {
  transform: scale(1.05);
}

.travel-card-cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  color: #cbd5e1;
  transition: transform 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.travel-card-status {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
}

.travel-card-cover-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  color: #ffffff;
  font-size: 14px;
  font-weight: 500;
  opacity: 0;
  transition: opacity 300ms cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 1;
  pointer-events: none;
}

.travel-card:hover .travel-card-cover-overlay {
  opacity: 1;
}

.travel-card-body {
  flex: 1;
  padding: 16px;
}

.travel-card-title {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--text-primary, #1e293b);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.travel-card-summary {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary, #64748b);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.travel-card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-card-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.travel-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-top: 1px solid var(--border-color, #e2e8f0);
}

.travel-card-rating {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--text-secondary, #64748b);
}

.travel-card-rating-score {
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.travel-card-rating-icon {
  color: #f59e0b;
  font-size: 12px;
}

.travel-card-review-count {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
}

.travel-card-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

@keyframes travel-card-fade-in-up {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .travel-card {
    animation-name: travel-card-fade-in;
    transform: none !important;
    transition: opacity 250ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .travel-card:hover {
    transform: none !important;
  }

  .travel-card-cover img,
  .travel-card-cover-placeholder {
    transition: none;
    transform: none !important;
  }

  .travel-card-cover-overlay {
    transition: opacity 250ms cubic-bezier(0.4, 0, 0.2, 1);
  }
}

@keyframes travel-card-fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
