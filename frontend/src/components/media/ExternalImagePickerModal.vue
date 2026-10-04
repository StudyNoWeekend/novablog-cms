<template>
  <a-modal
    v-model:open="open"
    :title="modalTitle"
    width="680px"
    :footer="null"
    @cancel="handleClose"
  >
    <div class="image-search-picker">
      <div class="picker-toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="searchPlaceholder"
          allow-clear
          enter-button
          aria-label="搜索关键词"
          @search="handleSearch"
        />
        <span class="toolbar-hint">{{ sourceHint }}</span>
      </div>

      <a-spin :spinning="loading || saving">
        <!-- 品牌图标网格 -->
        <div v-if="type === 'tech' && results.length > 0" class="icon-grid">
          <button
            v-for="item in results"
            :key="item.url"
            type="button"
            class="icon-item"
            :aria-label="`选择图标 ${item.name}`"
            @click="handlePick(item)"
          >
            <img :src="thumbSrc(item.url)" :alt="item.name" loading="lazy" />
            <span class="icon-title">{{ item.name }}</span>
          </button>
        </div>

        <!-- 封面/图片网格 -->
        <div v-else-if="results.length > 0" class="cover-grid">
          <button
            v-for="(item, index) in results"
            :key="`${item.url}-${index}`"
            type="button"
            class="cover-item"
            :aria-label="`选择图片 ${item.name || index + 1}`"
            @click="handlePick(item)"
          >
            <img :src="thumbSrc(item.url)" :alt="item.name || '候选图片'" loading="lazy" referrerpolicy="no-referrer" />
            <span v-if="item.name" class="cover-name">{{ item.name }}</span>
          </button>
        </div>

        <a-empty
          v-else-if="!loading"
          :description="emptyText"
          style="margin-top: 48px"
        />
      </a-spin>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { imageSearchApi } from '@/api/image-search'
import type { ImageResult, ImageSearchType } from '@/api/image-search'

/** 外部图片统一走后端代理取图，避免浏览器直连图床触发反爬（豆瓣 418）/防盗链/直连不可达 */
const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1'
function thumbSrc(url: string) {
  return `${API_BASE}/image-search/thumbnail?url=${encodeURIComponent(url)}`
}

const props = defineProps<{
  /** 受控显隐（v-model:visible） */
  visible: boolean
  /** 图源语义类型：tech=品牌图标, game=游戏封面, book=书籍封面, recipe=美食图片 */
  type: 'tech' | 'game' | 'book' | 'recipe'
  /** 打开弹窗时预填的搜索关键词（通常传表单里的名称字段） */
  initialKeyword?: string
  /** 转存归属的媒体库模块文件夹 key（如 tech_stack/game/book/recipe） */
  module?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'selected', url: string): void
}>()

/** 语义类型 → 后端图源类型 */
const API_TYPE: Record<string, ImageSearchType> = {
  tech: 'icons',
  game: 'games',
  book: 'books',
  recipe: 'food',
}

/** 语义类型 → 默认媒体库模块文件夹 key（未显式传 module 时兜底） */
const DEFAULT_MODULE: Record<string, string> = {
  tech: 'tech_stack',
  game: 'game',
  book: 'book',
  recipe: 'recipe',
}

const saveModule = computed(() => props.module || DEFAULT_MODULE[props.type] || '')

const META: Record<string, { title: string; placeholder: string; hint: string }> = {
  tech: { title: '搜索品牌图标', placeholder: '输入技术名称，如 Vue / Docker / PostgreSQL', hint: '来自 Simple Icons 品牌图标库' },
  game: { title: '搜索游戏封面', placeholder: '输入游戏名称，如 塞尔达传说 / 黑神话', hint: '来自 Bangumi 条目库' },
  book: { title: '搜索书籍封面', placeholder: '输入书名，如 三体 / 深入理解计算机系统', hint: '来自豆瓣读书' },
  recipe: { title: '搜索美食图片', placeholder: '输入菜名，如 红烧肉 / 宫保鸡丁', hint: '来自 Bing 图片搜索' },
}

const open = ref(false)
const keyword = ref('')
const results = ref<ImageResult[]>([])
const loading = ref(false)
const saving = ref(false)
const failed = ref(false)

const modalTitle = computed(() => META[props.type].title)
const searchPlaceholder = computed(() => META[props.type].placeholder)
const sourceHint = computed(() => META[props.type].hint)
const emptyText = computed(() =>
  failed.value ? '搜索失败，请稍后重试，或改用本地上传' : '没有找到结果，换个关键词试试'
)

let debounceTimer: ReturnType<typeof setTimeout> | null = null
let skipNextAutoSearch = false

watch(
  () => props.visible,
  (v) => {
    open.value = v
    if (v) {
      results.value = []
      failed.value = false
      skipNextAutoSearch = true
      keyword.value = props.initialKeyword || ''
      if (keyword.value.trim()) {
        handleSearch()
      }
    }
  }
)

watch(open, (v) => {
  if (!v) {
    emit('update:visible', false)
  }
})

watch(keyword, () => {
  if (skipNextAutoSearch) {
    skipNextAutoSearch = false
    return
  }
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    handleSearch()
  }, 300)
})

async function handleSearch() {
  const q = keyword.value.trim()
  if (!q) {
    results.value = []
    return
  }
  loading.value = true
  failed.value = false
  try {
    results.value = await imageSearchApi.search(API_TYPE[props.type], q)
  } catch {
    // 错误提示已由 axios 拦截器统一弹出
    failed.value = true
    results.value = []
  } finally {
    loading.value = false
  }
}

async function handlePick(item: ImageResult) {
  if (!item.url) return
  // 除品牌图标外统一转存为本站资源：
  // 豆瓣图床有 Cookie 反爬、Bing 图片源自第三方站点有防盗链、lain.bgm.tv 部分网络直连不可达，
  // 直接外链在博客前台可能失效。品牌图标是 SVG 且 Simple Icons CDN 稳定，直链直接使用。
  if (props.type !== 'tech') {
    saving.value = true
    try {
      const media = await imageSearchApi.save(item.url, saveModule.value)
      emit('selected', media.url)
      message.success('已选择并转存图片')
      open.value = false
    } catch {
      // 错误提示已由 axios 拦截器统一弹出
    } finally {
      saving.value = false
    }
    return
  }
  emit('selected', item.url)
  message.success('已选择图片')
  open.value = false
}

function handleClose() {
  open.value = false
}
</script>

<style scoped>
.image-search-picker {
  display: flex;
  flex-direction: column;
}

.picker-toolbar {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.toolbar-hint {
  color: var(--text-secondary);
  font-size: 12px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
  padding: 4px;
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px 8px;
  cursor: pointer;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  background: var(--bg-card);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.icon-item:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.icon-item img {
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.icon-title {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: 12px;
}

.cover-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 420px;
  overflow-y: auto;
  padding: 4px;
}

.cover-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px;
  cursor: pointer;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  background: var(--bg-card);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.cover-item:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.cover-item img {
  width: 100%;
  aspect-ratio: 3 / 4;
  object-fit: cover;
  border-radius: 4px;
  background: var(--bg-card);
}

.cover-name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: 12px;
  text-align: center;
}

@media (max-width: 640px) {
  .icon-grid {
    grid-template-columns: repeat(4, 1fr);
  }

  .cover-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
