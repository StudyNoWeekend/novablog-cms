<template>
  <div class="song-row" @click="emit('play')">
    <!-- 来源图标（可点击跳转B站） -->
    <a
      :href="song.source_url"
      target="_blank"
      rel="noopener noreferrer"
      class="source-link"
      title="在B站查看"
      @click.stop
    >
      <svg viewBox="0 0 24 24" fill="currentColor" class="source-icon">
        <path d="M17.813 4.653h.854c1.51.054 2.769.578 3.773 1.574 1.004.995 1.524 2.249 1.56 3.76v7.36c-.036 1.51-.556 2.769-1.56 3.773s-2.262 1.524-3.773 1.56H5.333c-1.51-.036-2.769-.556-3.773-1.56S.036 18.858 0 17.347v-7.36c.036-1.511.556-2.765 1.56-3.76 1.004-.996 2.262-1.52 3.773-1.574h.774l-1.174-1.12a1.234 1.234 0 0 1-.373-.906c0-.356.124-.658.373-.907l.027-.027c.267-.249.573-.373.92-.373.347 0 .653.124.92.373L9.653 4.44c.071.071.134.142.187.213h4.267a.836.836 0 0 1 .16-.213l2.853-2.747c.267-.249.573-.373.92-.373.347 0 .662.151.929.4.267.249.391.551.391.907 0 .355-.124.657-.373.906L17.813 4.653zM5.333 7.24c-.746.018-1.373.276-1.88.773-.506.498-.769 1.12-.786 1.867v7.52c.017.746.28 1.373.786 1.88.507.506 1.134.769 1.88.786h13.334c.746-.017 1.373-.28 1.88-.786.506-.507.769-1.134.786-1.88v-7.52c-.017-.747-.28-1.369-.786-1.867-.507-.497-1.134-.755-1.88-.773H5.333zm2.347 3.293c-.356 0-.658.124-.907.373-.249.25-.373.551-.373.907 0 .355.124.657.373.906.249.25.551.374.907.374.355 0 .657-.124.906-.374.25-.249.374-.551.374-.906 0-.356-.124-.658-.374-.907-.249-.249-.551-.373-.906-.373zm8.64 0c-.356 0-.658.124-.907.373-.249.249-.373.551-.373.907 0 .355.124.657.373.906.249.25.551.374.907.374.355 0 .657-.124.906-.374.25-.249.374-.551.374-.906 0-.356-.124-.658-.374-.907-.249-.249-.551-.373-.906-.373z" />
      </svg>
    </a>

    <!-- 圆角封面 -->
    <img
      :src="coverUrl"
      :alt="song.title"
      referrerpolicy="no-referrer"
      class="cover-img"
    />

    <!-- 标题 -->
    <div class="song-title" :title="song.title">{{ song.title }}</div>

    <!-- 作者 -->
    <div class="song-artist" :title="song.artist">{{ song.artist }}</div>

    <!-- 时长 -->
    <div class="song-duration">{{ formatDuration(song.duration) }}</div>

    <!-- 播放按钮（hover 显示） -->
    <div class="play-icon-wrapper">
      <svg viewBox="0 0 24 24" fill="currentColor" class="play-icon">
        <path d="M8 5v14l11-7z" />
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Song } from '@/types/music'
import { getThumbUrl } from '@/utils/image'

const props = defineProps<{
  song: Song
}>()

const emit = defineEmits<{
  (e: 'play'): void
}>()

// B站封面 http -> https，避免混合内容拦截；COS 封面生成缩略图节省流量
const coverUrl = computed(() => {
  const url = props.song.cover_url
  if (!url) return '/favicon.svg'
  return getThumbUrl(url.replace(/^http:\/\//, 'https://'), 160)
})

function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
</script>

<style scoped>
.song-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: var(--bg-card);
  border-radius: var(--border-radius);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.song-row:hover {
  background: var(--color-primary-light);
}

/* 来源图标 */
.source-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  cursor: pointer;
  color: #fb7299;
  transition: background var(--transition-fast);
}

.source-link:hover {
  background: rgba(251, 114, 153, 0.1);
}

.source-icon {
  width: 22px;
  height: 22px;
}

/* 圆角封面 */
.cover-img {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
  background: #f1f5f9;
}

/* 标题 */
.song-title {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 作者 */
.song-artist {
  flex-shrink: 0;
  font-size: 13px;
  color: var(--text-secondary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 时长 */
.song-duration {
  flex-shrink: 0;
  font-size: 13px;
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
  min-width: 40px;
  text-align: right;
}

/* 播放按钮 */
.play-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  color: var(--color-primary);
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.song-row:hover .play-icon-wrapper {
  opacity: 1;
}

.play-icon {
  width: 20px;
  height: 20px;
}
</style>
