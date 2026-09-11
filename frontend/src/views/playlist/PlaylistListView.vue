<template>
  <div class="playlist-page">
    <div class="playlist-content">
      <!-- 左侧分类侧边栏 -->
      <MusicCategorySidebar
        v-model="selectedCategory"
        :categories="categories"
        @refresh="loadCategories"
      />
      <!-- 右侧歌曲列表 -->
      <div class="song-list-area">
        <div class="page-header">
          <h1 class="page-title">音乐播放列表</h1>
          <a-button type="primary" @click="showAddModal = true">
            <PlusOutlined /> 添加歌曲
          </a-button>
        </div>

        <!-- 加载骨架屏 -->
        <div v-if="loading && songs.length === 0" class="song-list">
          <div v-for="i in 8" :key="i" class="skeleton-row">
            <a-skeleton active :paragraph="{ rows: 1 }" />
          </div>
        </div>

        <!-- 错误状态 -->
        <a-result
          v-else-if="error"
          status="error"
          title="加载失败"
          sub-title="获取歌曲列表时出错，请重试"
        >
          <template #extra>
            <a-button type="primary" @click="loadSongs">重试</a-button>
          </template>
        </a-result>

        <!-- 空状态 -->
        <a-empty
          v-else-if="songs.length === 0"
          description="还没有歌曲，点击添加吧"
        >
          <a-button type="primary" @click="showAddModal = true">
            <PlusOutlined /> 添加歌曲
          </a-button>
        </a-empty>

        <!-- 歌曲列表 -->
        <div v-else class="song-list">
          <SongCard
            v-for="song in songs"
            :key="song.id"
            :song="song"
            @play="playSong(song)"
          />
        </div>

        <!-- 分页 -->
        <div v-if="total > pageSize" class="pagination-wrapper">
          <a-pagination
            :current="page"
            :page-size="pageSize"
            :total="total"
            :show-size-changer="true"
            :show-total="(total: number) => `共 ${total} 首`"
            @change="onPageChange"
            @show-size-change="handlePageSizeChange"
          />
        </div>
      </div>
    </div>

    <!-- 底部播放器 -->
    <MusicPlayerBar
      :current-song="currentSong"
      :has-prev="hasPrev"
      :has-next="hasNext"
      @prev="playPrev"
      @next="playNext"
      @play="resumePlay"
      @pause="pausePlay"
    />

    <!-- 添加歌曲弹窗 -->
    <SongFormModal
      v-model:open="showAddModal"
      :categories="categories"
      @saved="loadSongs"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { musicApi } from '@/api/music'
import { categoryApi } from '@/api/category'
import { usePagination } from '@/composables/usePagination'
import type { Song } from '@/types/music'
import type { Category } from '@/types/category'
import MusicCategorySidebar from '@/components/music/MusicCategorySidebar.vue'
import SongCard from '@/components/music/SongCard.vue'
import SongFormModal from '@/components/music/SongFormModal.vue'
import MusicPlayerBar from '@/components/music/MusicPlayerBar.vue'

const categories = ref<Category[]>([])
const songs = ref<Song[]>([])
const selectedCategory = ref('')
const loading = ref(false)
const error = ref(false)
const showAddModal = ref(false)

const { page, pageSize, total, handlePageChange } = usePagination(20)

function onPageChange(newPage: number, newPageSize?: number) {
  handlePageChange(newPage, newPageSize)
  loadSongs()
}

function handlePageSizeChange(_p: number, size: number) {
  pageSize.value = size
  page.value = 1
  loadSongs()
}

const currentSong = ref<Song | null>(null)

const currentIndex = computed(() => {
  if (!currentSong.value) return -1
  return songs.value.findIndex((s) => s.id === currentSong.value!.id)
})

const hasPrev = computed(() => currentIndex.value > 0)
const hasNext = computed(() => currentIndex.value >= 0 && currentIndex.value < songs.value.length - 1)

onMounted(() => {
  loadCategories()
  loadSongs()
})

watch(selectedCategory, () => {
  page.value = 1
  loadSongs()
})

async function loadCategories() {
  try {
    categories.value = await categoryApi.getList('music')
  } catch {
    // 错误由拦截器处理
  }
}

async function loadSongs() {
  loading.value = true
  error.value = false
  try {
    const params: { category_id?: string; page: number; page_size: number } = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (selectedCategory.value) {
      params.category_id = selectedCategory.value
    }
    const res = await musicApi.getSongs(params)
    songs.value = res.list || []
    total.value = res.total || 0
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function playSong(song: Song) {
  currentSong.value = song
}

function playPrev() {
  if (hasPrev.value) {
    currentSong.value = songs.value[currentIndex.value - 1]
  }
}

function playNext() {
  if (hasNext.value) {
    currentSong.value = songs.value[currentIndex.value + 1]
  }
}

function resumePlay() {
  // 播放器内部处理
}

function pausePlay() {
  // 播放器内部处理
}
</script>

<style scoped>
.playlist-page {
  padding: 24px;
  padding-bottom: 96px;
  min-height: 100vh;
}

.playlist-content {
  display: flex;
  gap: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.song-list-area {
  flex: 1;
  min-width: 0;
}

.song-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.skeleton-row {
  background: var(--bg-card);
  border-radius: var(--border-radius);
  padding: 14px 16px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
