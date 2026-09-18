<template>
  <div class="playlist-card">
    <!-- 封面 -->
    <div class="playlist-card__cover">
      <img
        v-if="playlist.cover_url"
        :src="playlist.cover_url"
        :alt="playlist.title"
        loading="lazy"
        @error="coverFailed = true"
      />
      <div v-else class="playlist-card__placeholder" :style="{ background: platform.color }">
        <span class="playlist-card__placeholder-text">{{ platform.icon }}</span>
      </div>
      <!-- 平台徽标 -->
      <div class="playlist-card__platform-badge" :style="{ background: platform.color }">
        {{ platform.icon }}
      </div>
    </div>

    <!-- 内容 -->
    <div class="playlist-card__body">
      <h3 class="playlist-card__title" :title="playlist.title">{{ playlist.title }}</h3>
      <p v-if="playlist.description" class="playlist-card__description" :title="playlist.description">
        {{ playlist.description }}
      </p>
      <div class="playlist-card__meta">
        <span class="playlist-card__platform-label">{{ platform.label }}</span>
        <span v-if="editable" class="playlist-card__status">
          <a-tag :color="playlist.enabled ? 'green' : 'default'">
            {{ playlist.enabled ? '已展示' : '已隐藏' }}
          </a-tag>
        </span>
      </div>
    </div>

    <!-- 操作按钮（仅管理端） -->
    <div v-if="editable" class="playlist-card__actions" @click.stop>
      <a-button size="small" @click="emit('edit')">编辑</a-button>
      <a-button size="small" danger @click="emit('delete')">删除</a-button>
    </div>

    <!-- 访客查看时整张卡片可点击跳转 -->
    <a
      v-if="!editable"
      :href="playlist.platform_url"
      target="_blank"
      rel="noopener noreferrer"
      class="playlist-card__overlay-link"
      :title="`在 ${platform.label} 中查看歌单`"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ThirdPartyPlaylist } from '@/types/playlist'
import { getPlatformConfig } from '@/utils/platformConfig'

const props = defineProps<{
  playlist: ThirdPartyPlaylist
  editable?: boolean
}>()

const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const coverFailed = ref(false)
const platform = getPlatformConfig(props.playlist.platform)
</script>

<style scoped>
.playlist-card {
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
  position: relative;
}

.playlist-card:hover {
  border-color: var(--primary-color, #1677ff);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
}

/* 封面区域 */
.playlist-card__cover {
  position: relative;
  height: 140px;
  flex-shrink: 0;
  overflow: hidden;
}

.playlist-card__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.playlist-card__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.playlist-card__placeholder-text {
  color: rgba(255, 255, 255, 0.92);
  font-size: 40px;
  font-weight: 600;
  user-select: none;
}

/* 平台徽标 */
.playlist-card__platform-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

/* 内容 */
.playlist-card__body {
  flex: 1;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.playlist-card__title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.playlist-card__description {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.playlist-card__meta {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  margin-top: auto;
  padding-top: 4px;
}

.playlist-card__platform-label {
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
  background: var(--bg-elevated, #f5f5f5);
  padding: 1px 8px;
  border-radius: 4px;
}

.playlist-card__status {
  margin-left: auto;
}

/* 操作按钮 */
.playlist-card__actions {
  display: flex;
  gap: 8px;
  padding: 8px 14px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}

.playlist-card__actions .ant-btn {
  flex: 1;
}

/* 访客点击链接遮罩 */
.playlist-card__overlay-link {
  position: absolute;
  inset: 0;
  z-index: 1;
  cursor: pointer;
}
</style>
