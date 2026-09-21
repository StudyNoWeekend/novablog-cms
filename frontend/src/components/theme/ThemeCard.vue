<template>
  <div class="theme-card" @click="emit('detail')">
    <div class="theme-card__cover">
      <img
        v-if="coverURL && !coverFailed"
        :src="theme.preview"
        :alt="theme.title"
        loading="lazy"
        @error="coverFailed = true"
      />
      <div v-else class="theme-card__placeholder" :style="{ background: themeGradient(theme.type) }">
        <span class="theme-card__placeholder-text">{{ theme.title.slice(0, 1) }}</span>
      </div>
      <span class="theme-card__price" :class="{ 'theme-card__price--paid': isPaid }">
        {{ isPaid ? `¥${theme.price_amount}` : '免费' }}
      </span>
      <a-tag v-if="theme.status !== 2" color="warning" class="theme-card__status">
        {{ themeStatusText(theme.status) }}
      </a-tag>
    </div>

    <div class="theme-card__body">
      <h3 class="theme-card__title" :title="theme.title">{{ theme.title }}</h3>
      <div class="theme-card__meta">
        <span class="theme-card__author"><UserOutlined /> {{ theme.author }}</span>
        <span class="theme-card__type">{{ themeTypeLabel }}</span>
      </div>
      <div v-if="theme.styles?.length" class="theme-card__tags">
        <a-tag v-for="s in theme.styles.slice(0, 3)" :key="s">{{ s }}</a-tag>
        <a-tag v-if="theme.styles.length > 3">+{{ theme.styles.length - 3 }}</a-tag>
      </div>
      <div class="theme-card__stats">
        <span :title="`下载 ${theme.downloads} 次`"><DownloadOutlined /> {{ theme.downloads }}</span>
        <span :title="`点赞 ${theme.likes} 次`"><HeartOutlined /> {{ theme.likes }}</span>
        <span :title="`评分 ${theme.rating}`"><StarOutlined /> {{ theme.rating > 0 ? theme.rating.toFixed(1) : '暂无' }}</span>
      </div>
    </div>

    <div class="theme-card__actions" @click.stop>
      <a-button size="small" @click="emit('detail')">详情</a-button>
      <a-button size="small" :type="favorited ? 'primary' : 'default'" @click="emit('favorite')">
        <template #icon><StarFilled v-if="favorited" /><StarOutlined v-else /></template>
        {{ favorited ? '已收藏' : '收藏' }}
      </a-button>
      <a-button size="small" type="primary" :loading="installing" @click="emit('install')">
        <template #icon><CloudDownloadOutlined /></template>
        安装
      </a-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  CloudDownloadOutlined,
  DownloadOutlined,
  HeartOutlined,
  StarFilled,
  StarOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import type { ThemeItem } from '@/types/theme'
import { isPreviewURL, themeGradient, themeStatusText, THEME_TYPE_LABELS } from '@/utils/themeDisplay'

const props = defineProps<{
  theme: ThemeItem
  favorited?: boolean
  installing?: boolean
}>()

const emit = defineEmits<{
  (e: 'detail'): void
  (e: 'favorite'): void
  (e: 'install'): void
}>()

const coverFailed = ref(false)
const coverURL = computed(() => isPreviewURL(props.theme.preview))
const isPaid = computed(() => props.theme.price === 'paid')
const themeTypeLabel = computed(() => THEME_TYPE_LABELS[props.theme.type] || props.theme.type)

watch(
  () => props.theme.id,
  () => {
    coverFailed.value = false
  },
)
</script>

<style scoped>
.theme-card {
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.theme-card:hover {
  border-color: var(--primary-color, #1677ff);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
}

.theme-card__cover {
  position: relative;
  height: 140px;
  flex-shrink: 0;
}

.theme-card__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.theme-card__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.theme-card__placeholder-text {
  color: rgba(255, 255, 255, 0.92);
  font-size: 40px;
  font-weight: 600;
  user-select: none;
}

.theme-card__price {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 0 8px;
  height: 22px;
  line-height: 22px;
  border-radius: 4px;
  background: rgba(82, 196, 26, 0.92);
  color: #fff;
  font-size: 12px;
}

.theme-card__price--paid {
  background: rgba(250, 140, 22, 0.92);
}

.theme-card__status {
  position: absolute;
  top: 8px;
  left: 8px;
}

.theme-card__body {
  flex: 1;
  padding: 12px 14px;
}

.theme-card__title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.theme-card__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

.theme-card__author {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.theme-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 0;
  margin-bottom: 8px;
}

.theme-card__tags .ant-tag {
  margin-inline-end: 4px;
}

.theme-card__stats {
  display: flex;
  gap: 14px;
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
}

.theme-card__stats span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.theme-card__actions {
  display: flex;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}

.theme-card__actions .ant-btn {
  flex: 1;
}
</style>
