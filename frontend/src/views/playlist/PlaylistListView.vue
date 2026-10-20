<template>
  <div class="playlist-page">
    <!-- 标签页切换：音频导入 / 第三方歌单 -->
    <a-tabs
      v-model:activeKey="activeTab"
      class="playlist-tabs"
      :tab-bar-gutter="32"
    >
      <a-tab-pane key="audio" tab="音频导入">
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
              <a-button type="primary" @click="openAddModal">
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
              <a-button type="primary" @click="openAddModal">
                <PlusOutlined /> 添加歌曲
              </a-button>
            </a-empty>

            <!-- 歌曲列表 -->
            <div v-else class="song-list">
              <SongCard
                v-for="(song, index) in songs"
                :key="song.id"
                :song="song"
                :categories="categories"
                :draggable="true"
                :class="{
                  'is-dragging': dragIndex === index,
                  'is-drag-over': dragOverIndex === index,
                }"
                @dragstart="handleDragStart(index, $event)"
                @dragover.prevent="handleDragOver(index)"
                @dragleave="dragOverIndex = null"
                @drop.prevent="handleDrop()"
                @dragend="handleDragEnd"
                @play="playSong(song)"
                @edit="openEditModal(song)"
                @delete="handleDeleteSong(song)"
                @top="handleTopSong(song)"
                @change-category="(categoryId: string) => handleChangeCategory(song, categoryId)"
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
          @close="closePlayer"
        />

        <!-- 添加/编辑歌曲弹窗 -->
        <SongFormModal
          v-model:open="showAddModal"
          :categories="categories"
          :song="editingSong"
          @saved="onSongSaved"
        />
      </a-tab-pane>

      <a-tab-pane key="third-party" tab="第三方歌单">
        <div class="third-party-area">
          <div class="page-header">
            <h1 class="page-title">第三方歌单</h1>
            <a-button type="primary" @click="showPlaylistModal = true">
              <PlusOutlined /> 添加歌单
            </a-button>
          </div>

          <!-- 加载骨架屏 -->
          <div v-if="loadingPlaylists" class="playlist-grid">
            <div v-for="i in 6" :key="i" class="skeleton-card">
              <a-skeleton active :paragraph="{ rows: 2 }" />
            </div>
          </div>

          <!-- 错误状态 -->
          <a-result
            v-else-if="playlistError"
            status="error"
            title="加载失败"
            sub-title="获取歌单列表时出错，请重试"
          >
            <template #extra>
              <a-button type="primary" @click="loadPlaylists">重试</a-button>
            </template>
          </a-result>

          <!-- 空状态 -->
          <a-empty
            v-else-if="playlists.length === 0"
            description="还没有第三方歌单，点击添加吧"
          >
            <a-button type="primary" @click="showPlaylistModal = true">
              <PlusOutlined /> 添加歌单
            </a-button>
          </a-empty>

          <!-- 歌单卡片网格 -->
          <div v-else class="playlist-grid">
            <ThirdPartyPlaylistCard
              v-for="pl in playlists"
              :key="pl.id"
              :playlist="pl"
              :editable="true"
              @edit="editPlaylist(pl)"
              @delete="handleDeletePlaylist(pl)"
            />
          </div>

          <!-- 分页 -->
          <div v-if="playlistTotal > playlistPageSize" class="pagination-wrapper">
            <a-pagination
              :current="playlistPage"
              :page-size="playlistPageSize"
              :total="playlistTotal"
              :show-size-changer="true"
              :show-total="(total: number) => `共 ${total} 个歌单`"
              @change="onPlaylistPageChange"
              @show-size-change="handlePlaylistPageSizeChange"
            />
          </div>
        </div>

        <!-- 添加/编辑歌单弹窗 -->
        <PlaylistFormModal
          v-model:open="showPlaylistModal"
          :playlist="editingPlaylist"
          @saved="onPlaylistSaved"
        />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { message, Modal } from 'ant-design-vue'
import { musicApi } from '@/api/music'
import { playlistApi } from '@/api/playlist'
import { categoryApi } from '@/api/category'
import { usePagination } from '@/composables/usePagination'
import type { Song } from '@/types/music'
import type { ThirdPartyPlaylist } from '@/types/playlist'
import type { Category } from '@/types/category'
import MusicCategorySidebar from '@/components/music/MusicCategorySidebar.vue'
import SongCard from '@/components/music/SongCard.vue'
import SongFormModal from '@/components/music/SongFormModal.vue'
import MusicPlayerBar from '@/components/music/MusicPlayerBar.vue'
import ThirdPartyPlaylistCard from '@/components/playlist/ThirdPartyPlaylistCard.vue'
import PlaylistFormModal from '@/components/playlist/PlaylistFormModal.vue'

// ---- 标签页 ----
const activeTab = ref('audio')

// ---- 音频播放相关 ----
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
  loadPlaylists()
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

function closePlayer() {
  currentSong.value = null
}

// ---- 编辑/删除歌曲 ----
const editingSong = ref<Song | undefined>(undefined)

function openAddModal() {
  editingSong.value = undefined
  showAddModal.value = true
}

function openEditModal(song: Song) {
  editingSong.value = song
  showAddModal.value = true
}

function onSongSaved() {
  editingSong.value = undefined
  // 保存时可能自动创建了新分类，同步刷新侧栏与卡片分类下拉
  loadCategories()
  loadSongs()
}

function handleDeleteSong(song: Song) {
  Modal.confirm({
    title: '确认删除',
    content: `删除歌曲「${song.title}」后无法恢复，确定要删除吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await musicApi.deleteSong(song.id)
        message.success('删除成功')
        loadSongs()
      } catch {
        // 错误由拦截器处理
      }
    },
  })
}

// ---- 分类快改 ----
async function handleChangeCategory(song: Song, categoryId: string) {
  const prev = song.category_id
  const next = categoryId || null
  // 乐观更新本地状态
  song.category_id = next
  // 当前按分类筛选且歌曲被移走时，从列表中移除
  if (selectedCategory.value && next !== selectedCategory.value) {
    songs.value = songs.value.filter((s) => s.id !== song.id)
    total.value = Math.max(0, total.value - 1)
  }
  try {
    await musicApi.updateSong(song.id, { category_id: categoryId })
    message.success('分类已更新')
  } catch {
    // 更新失败回滚并重拉列表
    song.category_id = prev
    loadSongs()
  }
}

// ---- 排序（拖拽 + 置顶）----
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

function handleDragStart(index: number, e: DragEvent) {
  dragIndex.value = index
  // Firefox 需要先 setData 才会触发拖拽
  e.dataTransfer?.setData('text/plain', String(index))
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
  }
}

function handleDragOver(index: number) {
  if (dragIndex.value === null || dragIndex.value === index) return
  dragOverIndex.value = index
}

function handleDragEnd() {
  dragIndex.value = null
  dragOverIndex.value = null
}

function handleDrop() {
  const from = dragIndex.value
  const to = dragOverIndex.value
  handleDragEnd()
  if (from === null || to === null || from === to) return
  persistSortWithMove(toGlobalIndex(from), toGlobalIndex(to))
}

function toGlobalIndex(localIndex: number): number {
  return (page.value - 1) * pageSize.value + localIndex
}

// 拉取当前分类筛选下的全量歌曲（后端 page_size 上限 100，需分页循环）
async function fetchAllSongsInScope(): Promise<Song[]> {
  const all: Song[] = []
  let p = 1
  for (;;) {
    const params: { category_id?: string; page: number; page_size: number } = {
      page: p,
      page_size: 100,
    }
    if (selectedCategory.value) params.category_id = selectedCategory.value
    const res = await musicApi.getSongs(params)
    const list = res.list || []
    all.push(...list)
    if (list.length === 0 || all.length >= (res.total || 0)) break
    p++
  }
  return all
}

// 全量归一化保存：把 [from → to] 的移动合入全量顺序后统一编号提交，失败时回滚重拉
async function persistSortWithMove(from: number, to: number) {
  try {
    const all = await fetchAllSongsInScope()
    if (from < 0 || from >= all.length) return
    const moved = all.splice(from, 1)[0]
    all.splice(Math.max(0, Math.min(to, all.length)), 0, moved)
    await musicApi.sortSongs({
      items: all.map((s, idx) => ({ id: s.id, sort_order: idx + 1 })),
    })
    message.success('排序已更新')
  } catch {
    // 错误由拦截器提示，finally 中回滚到服务端顺序
  } finally {
    loadSongs()
  }
}

// 置顶：移动到全量顺序第 0 位（置顶后仍可被拖拽换位）
function handleTopSong(song: Song) {
  const localIndex = songs.value.findIndex((s) => s.id === song.id)
  if (localIndex < 0) return
  const globalIndex = toGlobalIndex(localIndex)
  if (globalIndex === 0) {
    message.info('该歌曲已在最前')
    return
  }
  persistSortWithMove(globalIndex, 0)
}

// ---- 第三方歌单相关 ----
const playlists = ref<ThirdPartyPlaylist[]>([])
const loadingPlaylists = ref(false)
const playlistError = ref(false)
const showPlaylistModal = ref(false)
const editingPlaylist = ref<ThirdPartyPlaylist | undefined>(undefined)

const {
  page: playlistPage,
  pageSize: playlistPageSize,
  total: playlistTotal,
  handlePageChange: handlePlaylistPageChangeBase,
} = usePagination(20)

function onPlaylistPageChange(newPage: number, newPageSize?: number) {
  handlePlaylistPageChangeBase(newPage, newPageSize)
  loadPlaylists()
}

function handlePlaylistPageSizeChange(_p: number, size: number) {
  playlistPageSize.value = size
  playlistPage.value = 1
  loadPlaylists()
}

async function loadPlaylists() {
  loadingPlaylists.value = true
  playlistError.value = false
  try {
    const res = await playlistApi.getList({
      page: playlistPage.value,
      page_size: playlistPageSize.value,
    })
    playlists.value = res.list || []
    playlistTotal.value = res.total || 0
  } catch {
    playlistError.value = true
  } finally {
    loadingPlaylists.value = false
  }
}

function editPlaylist(pl: ThirdPartyPlaylist) {
  editingPlaylist.value = pl
  showPlaylistModal.value = true
}

function handleDeletePlaylist(pl: ThirdPartyPlaylist) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除歌单「${pl.title}」吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await playlistApi.delete(pl.id)
        message.success('删除成功')
        loadPlaylists()
      } catch {
        // 错误由拦截器处理
      }
    },
  })
}

function onPlaylistSaved() {
  editingPlaylist.value = undefined
  loadPlaylists()
}
</script>

<style scoped>
.playlist-page {
  padding: 24px;
  padding-bottom: 24px;
  min-height: 100vh;
}

.playlist-tabs {
  max-width: 1400px;
  margin: 0 auto;
}

.playlist-content {
  display: flex;
  gap: 24px;
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

/* 拖拽排序状态（作用于 SongCard 根元素） */
.is-dragging {
  opacity: 0.5;
}

.is-drag-over {
  outline: 2px dashed var(--color-primary);
  outline-offset: -2px;
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

/* 第三方歌单区域 */
.third-party-area {
  min-height: 300px;
}

.playlist-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 20px;
}

.skeleton-card {
  background: var(--bg-card);
  border-radius: var(--border-radius-lg, 12px);
  padding: 16px;
  height: 200px;
}

/* 页面头部 */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.page-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}
</style>
