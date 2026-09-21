<template>
  <a-modal
    v-model:open="visible"
    :title="isEdit ? '编辑歌曲' : '添加歌曲'"
    :confirm-loading="saving"
    width="640px"
    @ok="handleSave"
    @cancel="handleCancel"
  >
    <a-form layout="vertical">
      <!-- B站链接输入 + 解析按钮（仅添加模式） -->
      <a-form-item v-if="!isEdit" label="B站视频链接">
        <a-input-group compact>
          <a-input
            v-model:value="biliUrl"
            placeholder="粘贴B站视频链接，如 https://www.bilibili.com/video/BV..."
            style="width: calc(100% - 90px)"
            @pressEnter="handleParse"
          />
          <a-button
            type="primary"
            :loading="parsing"
            style="width: 90px"
            @click="handleParse"
          >
            解析
          </a-button>
        </a-input-group>
      </a-form-item>

      <!-- 解析状态 -->
      <div v-if="parseStatus" class="parse-status">
        <a-spin v-if="parseStatus === 'processing'" size="small" />
        <span>{{ parseStatusText }}</span>
      </div>

      <!-- 编辑模式 或 单首结果：单首表单 -->
      <template v-if="isEdit || (parseResults.length <= 1 && formState.bvid)">
        <a-form-item label="歌曲名称" required>
          <a-input v-model:value="formState.title" placeholder="请输入歌曲名称" />
        </a-form-item>
        <a-form-item label="歌手" required>
          <a-input v-model:value="formState.artist" placeholder="请输入歌手名称" />
        </a-form-item>
        <a-form-item label="封面">
          <div class="cover-field">
            <div class="cover-preview" @click="showMediaPicker = true">
              <template v-if="formState.cover_url">
                <img :src="formState.cover_url.replace(/^http:\/\//, 'https://')" alt="封面预览" referrerpolicy="no-referrer" class="cover-preview-img" />
                <div class="cover-preview-overlay">
                  <PictureOutlined />
                </div>
              </template>
              <div v-else class="cover-placeholder">
                <PictureOutlined />
                <span>选择封面</span>
              </div>
            </div>
            <div class="cover-input">
              <a-input v-model:value="formState.cover_url" placeholder="封面图片URL，或从左侧媒体库选择" />
              <div class="cover-input-hint">点击缩略图可从媒体库选择，也可直接粘贴外部图片 URL（B站封面会自动保存到存储）。</div>
            </div>
          </div>
        </a-form-item>
        <a-form-item label="分类">
          <a-select
            v-model:value="formState.category_id"
            placeholder="选择分类"
            allow-clear
          >
            <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </template>

      <!-- 多首结果：批量选择列表 -->
      <template v-if="!isEdit && parseResults.length > 1">
        <div class="batch-header">
          <a-checkbox v-model:checked="selectAll">全选</a-checkbox>
          <span class="batch-summary">共 {{ parseResults.length }} 首，已选 {{ selectedCount }} 首</span>
        </div>

        <div class="batch-list">
          <div v-for="(item, index) in parseResults" :key="index" class="batch-item">
            <a-checkbox v-model:checked="item.selected" />
            <img :src="item.cover_url ? item.cover_url.replace(/^http:\/\//, 'https://') : '/favicon.svg'" :alt="item.title" referrerpolicy="no-referrer" class="batch-cover" />
            <div class="batch-fields">
              <a-input v-model:value="item.title" placeholder="歌曲名称" size="small" />
              <a-input v-model:value="item.artist" placeholder="歌手" size="small" />
            </div>
            <span class="batch-duration">{{ formatDuration(item.duration) }}</span>
          </div>
        </div>

        <a-form-item label="统一分类" style="margin-top: 16px">
          <a-select
            v-model:value="batchCategoryId"
            placeholder="为选中歌曲选择分类"
            allow-clear
          >
            <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </template>
    </a-form>

    <!-- 媒体库选择封面 -->
    <MediaPicker v-model:visible="showMediaPicker" @selected="handleMediaSelected" />
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import { PictureOutlined } from '@ant-design/icons-vue'
import { musicApi } from '@/api/music'
import type { Song, SongCreateReq, ParseResult, BatchCreateSongReq } from '@/types/music'
import type { Category } from '@/types/category'
import MediaPicker from '@/components/media/MediaPicker.vue'

const props = defineProps<{
  open: boolean
  categories: Category[]
  song?: Song
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const visible = computed({
  get: () => props.open,
  set: (val: boolean) => emit('update:open', val),
})

const isEdit = computed(() => !!props.song)

const saving = ref(false)
const biliUrl = ref('')
const parsing = ref(false)
const parseStatus = ref<'' | 'processing' | 'success' | 'failed'>('')
const showMediaPicker = ref(false)

const parseStatusText = computed(() => {
  switch (parseStatus.value) {
    case 'processing':
      return '正在解析，请稍候...'
    case 'success':
      return '解析成功，已自动填充信息'
    case 'failed':
      return '解析失败'
    default:
      return ''
  }
})

const formState = reactive<SongCreateReq>({
  title: '',
  artist: '',
  cover_url: '',
  bvid: '',
  cid: 0,
  source_url: '',
  source_type: 'bilibili',
  category_id: null,
  duration: 0,
  sort_order: 0,
})

// 解析结果列表（多首模式）
interface SelectableParseResult extends ParseResult {
  selected: boolean
  title: string
  artist: string
}
const parseResults = ref<SelectableParseResult[]>([])

// 批量分类
const batchCategoryId = ref<string | null>(null)

// 全选
const selectAll = computed({
  get: () => parseResults.value.length > 0 && parseResults.value.every(r => r.selected),
  set: (val: boolean) => {
    parseResults.value.forEach(r => (r.selected = val))
  },
})
const selectedCount = computed(() => parseResults.value.filter(r => r.selected).length)

let parseInterval: ReturnType<typeof setInterval> | null = null

function clearParseInterval() {
  if (parseInterval) {
    clearInterval(parseInterval)
    parseInterval = null
  }
}

onUnmounted(() => {
  clearParseInterval()
})

// 重置表单
watch(
  () => props.open,
  (val) => {
    if (val) {
      if (props.song) {
        formState.title = props.song.title
        formState.artist = props.song.artist
        formState.cover_url = props.song.cover_url
        formState.bvid = props.song.bvid
        formState.cid = props.song.cid
        formState.source_url = props.song.source_url
        formState.source_type = props.song.source_type
        formState.category_id = props.song.category_id
        formState.duration = props.song.duration
        formState.sort_order = props.song.sort_order
      } else {
        resetForm()
      }
      biliUrl.value = ''
      parseStatus.value = ''
      parseResults.value = []
      batchCategoryId.value = null
    }
  },
)

function resetForm() {
  formState.title = ''
  formState.artist = ''
  formState.cover_url = ''
  formState.bvid = ''
  formState.cid = 0
  formState.source_url = ''
  formState.source_type = 'bilibili'
  formState.category_id = null
  formState.duration = 0
  formState.sort_order = 0
}

async function handleParse() {
  if (!biliUrl.value.trim()) {
    message.warning('请输入B站视频链接')
    return
  }
  parsing.value = true
  parseStatus.value = 'processing'
  clearParseInterval()

  try {
    const { task_id } = await musicApi.parse(biliUrl.value.trim())

    parseInterval = setInterval(async () => {
      try {
        const task = await musicApi.getParseStatus(task_id)

        if (task.status === 'success' && task.results) {
          applyParseResults(task.results)
          parseStatus.value = 'success'
          parsing.value = false
          clearParseInterval()
          message.success('解析成功')
        } else if (task.status === 'failed') {
          parseStatus.value = 'failed'
          parsing.value = false
          clearParseInterval()
          message.error(task.error || '解析失败')
        }
      } catch {
        parseStatus.value = 'failed'
        parsing.value = false
        clearParseInterval()
      }
    }, 2000)
  } catch {
    parseStatus.value = 'failed'
    parsing.value = false
  }
}

function applyParseResults(results: ParseResult[]) {
  if (results.length === 1) {
    // 单首：填充表单
    const r = results[0]
    formState.title = r.title
    formState.artist = r.artist
    formState.cover_url = r.cover_url
    formState.bvid = r.bvid
    formState.cid = r.cid
    formState.source_url = biliUrl.value.trim()
    formState.duration = r.duration
  } else {
    // 多首：填充列表
    parseResults.value = results.map(r => ({
      ...r,
      selected: true,
      title: r.title,
      artist: r.artist,
    }))
  }
}

function handleMediaSelected(media: { id: string; url: string }) {
  formState.cover_url = media.url
}

async function handleSave() {
  if (isEdit.value && props.song) {
    // 编辑模式：单首更新
    if (!formState.title.trim()) {
      message.warning('请输入歌曲名称')
      return
    }
    if (!formState.artist.trim()) {
      message.warning('请输入歌手名称')
      return
    }
    saving.value = true
    try {
      await musicApi.updateSong(props.song.id, {
        title: formState.title,
        artist: formState.artist,
        cover_url: formState.cover_url,
        category_id: formState.category_id,
        duration: formState.duration,
        sort_order: formState.sort_order,
      })
      message.success('更新成功')
      emit('saved')
      emit('update:open', false)
    } catch {
      // 错误由拦截器处理
    } finally {
      saving.value = false
    }
    return
  }

  if (parseResults.value.length > 1) {
    // 多首批量创建
    const selected = parseResults.value.filter(r => r.selected)
    if (selected.length === 0) {
      message.warning('请至少选择一首歌曲')
      return
    }
    const songs: SongCreateReq[] = selected.map(r => ({
      title: r.title,
      artist: r.artist,
      cover_url: r.cover_url,
      bvid: r.bvid,
      cid: r.cid,
      source_url: biliUrl.value.trim(),
      source_type: 'bilibili',
      category_id: batchCategoryId.value,
      duration: r.duration,
    }))
    saving.value = true
    try {
      await musicApi.batchCreateSongs({ songs })
      message.success(`成功添加 ${songs.length} 首歌曲`)
      emit('saved')
      emit('update:open', false)
    } catch {
      // 错误由拦截器处理
    } finally {
      saving.value = false
    }
    return
  }

  // 单首创建
  if (!formState.title.trim()) {
    message.warning('请输入歌曲名称')
    return
  }
  if (!formState.artist.trim()) {
    message.warning('请输入歌手名称')
    return
  }
  if (!formState.bvid) {
    message.warning('请先解析B站视频链接')
    return
  }

  saving.value = true
  try {
    await musicApi.createSong({
      title: formState.title,
      artist: formState.artist,
      cover_url: formState.cover_url,
      bvid: formState.bvid,
      cid: formState.cid,
      source_url: formState.source_url,
      source_type: formState.source_type,
      category_id: formState.category_id,
      duration: formState.duration,
      sort_order: formState.sort_order,
    })
    message.success('添加成功')
    emit('saved')
    emit('update:open', false)
  } catch {
    // 错误由拦截器处理
  } finally {
    saving.value = false
  }
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

function handleCancel() {
  clearParseInterval()
  emit('update:open', false)
}
</script>

<style scoped>
.parse-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 16px;
  border-radius: var(--border-radius);
  background: var(--color-primary-light);
  font-size: 13px;
  color: var(--text-secondary);
}
.batch-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.batch-summary {
  font-size: 13px;
  color: var(--text-secondary);
}
.batch-list {
  max-height: 360px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  padding: 4px 0;
}
.batch-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
}
.batch-item:last-child {
  border-bottom: none;
}
.batch-cover {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
}
.batch-fields {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.batch-duration {
  font-size: 12px;
  color: var(--text-tertiary);
  flex-shrink: 0;
  white-space: nowrap;
}

/* 封面选择 */
.cover-field {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.cover-preview {
  position: relative;
  flex-shrink: 0;
  width: 88px;
  height: 66px;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  background: #f8fafc;
  border: 1px dashed var(--border-color, #cbd5e1);
  transition: all 0.2s;
}
.cover-preview:hover {
  border-color: var(--primary, #3b82f6);
}
.cover-preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.cover-preview-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 14px;
  opacity: 0;
  transition: opacity 0.2s;
}
.cover-preview:hover .cover-preview-overlay {
  opacity: 1;
}
.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
  gap: 2px;
}
.cover-placeholder:hover {
  color: var(--primary, #3b82f6);
}
.cover-input {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.cover-input-hint {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
}
</style>
