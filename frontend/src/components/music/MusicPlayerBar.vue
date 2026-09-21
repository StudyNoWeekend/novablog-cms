<template>
  <div v-if="currentSong" class="player-bar">
    <div class="player-info">
      <img
        :src="coverUrl || '/favicon.svg'"
        :alt="currentSong.title"
        referrerpolicy="no-referrer"
        class="player-cover"
      />
      <div class="player-meta">
        <div class="player-title" :title="currentSong.title">{{ currentSong.title }}</div>
        <div class="player-artist" :title="currentSong.artist">{{ currentSong.artist }}</div>
      </div>
    </div>
    <div class="player-controls">
      <button
        class="control-btn"
        :disabled="!hasPrev"
        title="上一首"
        @click="emit('prev')"
      >
        <svg viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
          <path d="M6 6h2v12H6V6zm3.5 6l8.5 6V6l-8.5 6z" />
        </svg>
      </button>
      <button
        class="control-btn play-btn"
        :title="isPlaying ? '暂停' : '播放'"
        @click="togglePlay"
      >
        <svg v-if="isPlaying" viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
          <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
          <path d="M8 5v14l11-7z" />
        </svg>
      </button>
      <button
        class="control-btn"
        :disabled="!hasNext"
        title="下一首"
        @click="emit('next')"
      >
        <svg viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
          <path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z" />
        </svg>
      </button>
      <div class="progress-wrapper">
        <span class="time">{{ formatTime(currentTime) }}</span>
        <div ref="progressBar" class="progress-bar" @click="handleSeek">
          <div class="progress-filled" :style="{ width: progressPercent + '%' }"></div>
        </div>
        <span class="time">{{ formatTime(duration) }}</span>
      </div>
    </div>
    <!-- 隐藏的 audio 元素，referrerpolicy 避免 Referer 校验问题 -->
    <audio ref="audioRef" referrerpolicy="no-referrer" @timeupdate="onTimeUpdate" @loadedmetadata="onLoadedMetadata" @ended="onEnded" @error="onAudioError" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { musicApi } from '@/api/music'
import type { Song } from '@/types/music'
import { getThumbUrl } from '@/utils/image'

const props = defineProps<{
  currentSong: Song | null
  hasPrev: boolean
  hasNext: boolean
}>()

const emit = defineEmits<{
  (e: 'prev'): void
  (e: 'next'): void
  (e: 'play'): void
  (e: 'pause'): void
}>()

const audioRef = ref<HTMLAudioElement | null>(null)
const progressBar = ref<HTMLDivElement | null>(null)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)

const progressPercent = ref(0)
const currentSongId = ref('')
const urlFetchedAt = ref(0) // 记录 URL 获取时间戳
let isRetrying = false

// 将 B 站封面的 HTTP 协议转为 HTTPS，避免混合内容拦截
function fixCoverUrl(url: string): string {
  if (!url) return ''
  return url.replace(/^http:\/\//, 'https://')
}

const coverUrl = computed(() => getThumbUrl(fixCoverUrl(props.currentSong?.cover_url || ''), 96))

function formatTime(sec: number): string {
  if (!sec || isNaN(sec)) return '00:00'
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

async function loadAudio(songId: string) {
  try {
    const { url } = await musicApi.getAudioUrl(songId)
    currentSongId.value = songId
    urlFetchedAt.value = Date.now()
    if (audioRef.value) {
      audioRef.value.src = url
      await audioRef.value.play()
      isPlaying.value = true
    }
  } catch {
    isPlaying.value = false
  }
}

// 音频播放出错（URL 过期等），自动重新获取
async function onAudioError() {
  if (isRetrying || !currentSongId.value) return
  isRetrying = true
  // 等待 1 秒避免频繁重试
  await new Promise((r) => setTimeout(r, 1000))
  try {
    const { url } = await musicApi.getAudioUrl(currentSongId.value)
    urlFetchedAt.value = Date.now()
    if (audioRef.value) {
      const prevTime = currentTime.value
      audioRef.value.src = url
      audioRef.value.currentTime = prevTime
      await audioRef.value.play()
      isPlaying.value = true
    }
  } catch {
    isPlaying.value = false
  } finally {
    isRetrying = false
  }
}

function onTimeUpdate() {
  if (audioRef.value) {
    currentTime.value = audioRef.value.currentTime
    if (duration.value > 0) {
      progressPercent.value = (currentTime.value / duration.value) * 100
    }
  }
}

function onLoadedMetadata() {
  if (audioRef.value) {
    duration.value = audioRef.value.duration
  }
}

function onEnded() {
  isPlaying.value = false
  emit('next')
}

function togglePlay() {
  if (!audioRef.value) return
  if (isPlaying.value) {
    audioRef.value.pause()
    isPlaying.value = false
    emit('pause')
  } else {
    audioRef.value.play()
    isPlaying.value = true
    emit('play')
  }
}

function handleSeek(e: MouseEvent) {
  if (!progressBar.value || !audioRef.value || duration.value === 0) return
  const rect = progressBar.value.getBoundingClientRect()
  const percent = (e.clientX - rect.left) / rect.width
  audioRef.value.currentTime = percent * duration.value
}

// 监听歌曲切换
watch(
  () => props.currentSong,
  (newSong) => {
    if (newSong) {
      currentTime.value = 0
      duration.value = 0
      progressPercent.value = 0
      loadAudio(newSong.id)
    } else {
      isPlaying.value = false
      if (audioRef.value) {
        audioRef.value.pause()
        audioRef.value.src = ''
      }
    }
  },
)

defineExpose({ togglePlay })
</script>

<style scoped>
.player-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 72px;
  background: var(--bg-card);
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 24px;
  z-index: 100;
}

.player-info {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 240px;
  flex-shrink: 0;
}

.player-cover {
  width: 48px;
  height: 48px;
  border-radius: var(--border-radius);
  object-fit: cover;
  flex-shrink: 0;
}

.player-meta {
  overflow: hidden;
}

.player-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.player-artist {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.player-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: transparent;
  cursor: pointer;
  color: var(--text-primary);
  transition: background var(--transition-fast), color var(--transition-fast);
}

.control-btn:hover:not(:disabled) {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.control-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.play-btn {
  width: 40px;
  height: 40px;
  background: var(--color-primary);
  color: #fff;
}

.play-btn:hover:not(:disabled) {
  background: var(--color-primary-hover);
  color: #fff;
}

.ctrl-icon {
  width: 20px;
  height: 20px;
}

.progress-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  max-width: 500px;
}

.time {
  font-size: 12px;
  color: var(--text-tertiary);
  flex-shrink: 0;
  min-width: 40px;
  text-align: center;
}

.progress-bar {
  flex: 1;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.progress-filled {
  height: 100%;
  background: var(--color-primary);
  border-radius: 2px;
  transition: width 0.1s linear;
}
</style>
