<template>
  <transition name="player-slide">
    <div v-if="currentSong" class="player-bar">
      <div class="player-side">
        <div class="player-info">
          <SongCoverImg
            :url="currentSong.cover_url"
            :alt="currentSong.title"
            :width="96"
            class="player-cover"
          />
          <div class="player-meta">
            <div class="player-title" :title="currentSong.title">{{ currentSong.title }}</div>
            <div class="player-artist" :title="currentSong.artist">{{ currentSong.artist }}</div>
            <div class="player-source">{{ $t('music.player.source') }}</div>
          </div>
        </div>
        <div class="player-controls">
          <button
            class="control-btn"
            :disabled="!hasPrev"
            :title="$t('music.player.prev')"
            @click="emit('prev')"
          >
            <svg viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
              <path d="M6 6h2v12H6V6zm3.5 6l8.5 6V6l-8.5 6z" />
            </svg>
          </button>
          <button
            class="control-btn"
            :disabled="!hasNext"
            :title="$t('music.player.next')"
            @click="emit('next')"
          >
            <svg viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
              <path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z" />
            </svg>
          </button>
          <button class="control-btn" :title="$t('music.player.close')" @click="emit('close')">
            <svg viewBox="0 0 24 24" fill="currentColor" class="ctrl-icon">
              <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
            </svg>
          </button>
        </div>
      </div>
      <!-- B 站官方外链播放器：iframe 加载即播，播放控制由 B 站 UI 提供 -->
      <div class="player-frame">
        <div v-if="isLoading" class="player-loading">
          <span class="loading-spinner" :aria-label="$t('common.state.loading')"></span>
          <span class="loading-text">{{ $t('music.player.loading') }}</span>
        </div>
        <iframe
          v-if="playerSrc"
          :src="playerSrc"
          class="player-iframe"
          :title="$t('music.player.frameTitle')"
          allow="autoplay; fullscreen; encrypted-media"
          allowfullscreen
          scrolling="no"
          frameborder="0"
          @load="onFrameLoad"
        ></iframe>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { musicApi } from '@/api/music'
import type { Song } from '@/types/music'
import { message } from 'ant-design-vue'
import SongCoverImg from '@/components/music/SongCoverImg.vue'

const props = defineProps<{
  currentSong: Song | null
  hasPrev: boolean
  hasNext: boolean
}>()

const emit = defineEmits<{
  (e: 'prev'): void
  (e: 'next'): void
  (e: 'close'): void
}>()

const { t } = useI18n()

const playerSrc = ref('')
const isLoading = ref(false)
const currentSongId = ref('')

function onFrameLoad() {
  isLoading.value = false
}

async function loadPlayer(songId: string) {
  isLoading.value = true
  try {
    const { url } = await musicApi.getAudioUrl(songId)
    // 快速切歌时丢弃过期响应
    if (currentSongId.value !== songId) return
    playerSrc.value = url
  } catch {
    isLoading.value = false
    if (currentSongId.value === songId) {
      message.error(t('music.player.loadFailed'))
    }
  }
}

// 监听歌曲切换：换曲即重新加载 iframe（src 替换后旧播放自动停止）
watch(
  () => props.currentSong,
  (newSong) => {
    if (newSong) {
      currentSongId.value = newSong.id
      playerSrc.value = ''
      loadPlayer(newSong.id)
    } else {
      currentSongId.value = ''
      playerSrc.value = ''
      isLoading.value = false
    }
  },
)
</script>

<style scoped>
.player-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: var(--bg-card);
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 16px 24px;
}

.player-slide-enter-active,
.player-slide-leave-active {
  transition: transform 0.25s ease, opacity 0.25s ease;
}

.player-slide-enter-from,
.player-slide-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.player-side {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-self: stretch;
  width: 240px;
  flex-shrink: 0;
  gap: 12px;
}

.player-info {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
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

.player-source {
  font-size: 11px;
  color: var(--text-tertiary);
}

.player-controls {
  display: flex;
  align-items: center;
  gap: 8px;
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

.ctrl-icon {
  width: 20px;
  height: 20px;
}

.player-frame {
  position: relative;
  flex: 1;
  max-width: 520px;
  aspect-ratio: 16 / 9;
  align-self: center;
  border-radius: var(--border-radius);
  overflow: hidden;
  background: #000;
}

.player-iframe {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  border: none;
  display: block;
}

.player-loading {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  background: var(--bg-card);
}

.loading-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-color);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-text {
  font-size: 12px;
  color: var(--text-tertiary);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .player-bar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
    padding: 12px 16px;
  }

  .player-side {
    width: 100%;
    flex-direction: row;
    align-items: center;
  }

  .player-controls {
    margin-left: auto;
  }

  .player-frame {
    max-width: 100%;
    width: 100%;
  }
}
</style>
