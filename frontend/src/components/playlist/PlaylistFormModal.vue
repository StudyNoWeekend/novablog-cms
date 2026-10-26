<template>
  <a-modal
    v-model:open="visible"
    :title="isEdit ? $t('playlist.form.editTitle') : $t('playlist.form.addTitle')"
    :confirm-loading="saving"
    width="600px"
    @ok="handleSave"
    @cancel="handleCancel"
  >
    <a-form layout="vertical">
      <!-- 平台选择 -->
      <a-form-item :label="$t('playlist.form.platform')" required>
        <a-select v-model:value="formState.platform" :placeholder="$t('playlist.form.platformPlaceholder')">
          <a-select-option v-for="p in platformOptions" :key="p.id" :value="p.id">
            <div class="platform-option">
              <span class="platform-dot" :style="{ background: p.color }">{{ p.icon }}</span>
              {{ p.label }}
            </div>
          </a-select-option>
        </a-select>
      </a-form-item>

      <!-- 歌单链接 -->
      <a-form-item :label="$t('playlist.form.url')" required>
        <a-input
          v-model:value="formState.platform_url"
          :placeholder="$t('playlist.form.urlPlaceholder')"
        />
      </a-form-item>

      <!-- 歌单名称 -->
      <a-form-item :label="$t('playlist.form.name')" required>
        <a-input v-model:value="formState.title" :placeholder="$t('playlist.form.namePlaceholder')" />
      </a-form-item>

      <!-- 封面URL -->
      <a-form-item :label="$t('playlist.form.coverUrl')">
        <a-input
          v-model:value="formState.cover_url"
          :placeholder="$t('playlist.form.coverUrlPlaceholder')"
          @change="handleCoverUrlInput"
        >
          <template #addonAfter>
            <a :aria-label="$t('playlist.form.selectCoverFromMedia')" @click.prevent="mediaPickerVisible = true">
              <PictureOutlined /> {{ $t('playlist.form.mediaLibrary') }}
            </a>
          </template>
        </a-input>
      </a-form-item>

      <!-- 描述 -->
      <a-form-item :label="$t('common.field.description')">
        <a-textarea
          v-model:value="formState.description"
          :placeholder="$t('playlist.form.descriptionPlaceholder')"
          :rows="2"
          :maxlength="500"
          show-count
        />
      </a-form-item>

      <!-- 排序 -->
      <a-form-item :label="$t('common.field.sort')">
        <a-input-number v-model:value="formState.sort_order" :min="0" :max="9999" style="width: 120px" />
      </a-form-item>

      <!-- 是否展示 -->
      <a-form-item :label="$t('playlist.form.showOnFront')">
        <a-switch v-model:checked="formState.enabled" />
      </a-form-item>
    </a-form>

    <!-- 媒体库选择封面 -->
    <MediaPicker v-model:visible="mediaPickerVisible" module="music" @selected="handleMediaSelected" />
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { PictureOutlined } from '@ant-design/icons-vue'
import { playlistApi } from '@/api/playlist'
import { PLATFORM_CONFIGS } from '@/utils/platformConfig'
import type { ThirdPartyPlaylist, CreatePlaylistReq, UpdatePlaylistReq } from '@/types/playlist'
import MediaPicker from '@/components/media/MediaPicker.vue'

const props = defineProps<{
  open: boolean
  playlist?: ThirdPartyPlaylist
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const { t } = useI18n()

const visible = computed({
  get: () => props.open,
  set: (val: boolean) => emit('update:open', val),
})

const isEdit = computed(() => !!props.playlist)

const saving = ref(false)
const mediaPickerVisible = ref(false)

const platformOptions = Object.values(PLATFORM_CONFIGS)

// 表单媒体引用字段（coverMediaId 非空时后端优先按 ID 解析，手输 URL 则为外链模式）
const formState = reactive<CreatePlaylistReq & { coverMediaId: string; coverPresetId: string }>({
  title: '',
  platform: '',
  platform_url: '',
  cover_url: '',
  coverMediaId: '',
  coverPresetId: '',
  description: '',
  sort_order: 0,
  enabled: true,
})

// 重置表单
watch(
  () => props.open,
  (val) => {
    if (val) {
      if (props.playlist) {
        formState.title = props.playlist.title
        formState.platform = props.playlist.platform
        formState.platform_url = props.playlist.platform_url
        formState.cover_url = props.playlist.cover_url
        formState.coverMediaId = props.playlist.cover_media_id || ''
        formState.coverPresetId = props.playlist.cover_preset_id || ''
        formState.description = props.playlist.description
        formState.sort_order = props.playlist.sort_order
        formState.enabled = props.playlist.enabled
      } else {
        formState.title = ''
        formState.platform = ''
        formState.platform_url = ''
        formState.cover_url = ''
        formState.coverMediaId = ''
        formState.coverPresetId = ''
        formState.description = ''
        formState.sort_order = 0
        formState.enabled = true
      }
    }
  },
)

// 媒体库选图：记录媒体引用并回填 URL 用于展示
function handleMediaSelected(media: { id: string; url: string; presetId?: string }) {
  formState.cover_url = media.url
  formState.coverMediaId = media.id
  formState.coverPresetId = media.presetId || ''
}

// 手输 URL = 外链模式，清空媒体引用
function handleCoverUrlInput() {
  formState.coverMediaId = ''
  formState.coverPresetId = ''
}

async function handleSave() {
  if (!formState.title.trim()) {
    message.warning(t('playlist.form.nameRequired'))
    return
  }
  if (!formState.platform) {
    message.warning(t('playlist.form.platformRequired'))
    return
  }
  if (!formState.platform_url.trim()) {
    message.warning(t('playlist.form.urlRequired'))
    return
  }

  saving.value = true
  try {
    if (isEdit.value && props.playlist) {
      const data: UpdatePlaylistReq = {
        title: formState.title,
        platform: formState.platform,
        platform_url: formState.platform_url,
        cover_url: formState.cover_url || undefined,
        cover_media_id: formState.coverMediaId || undefined,
        cover_preset_id: formState.coverPresetId || undefined,
        description: formState.description || undefined,
        sort_order: formState.sort_order,
        enabled: formState.enabled,
      }
      await playlistApi.update(props.playlist.id, data)
      message.success(t('common.message.updateSuccess'))
    } else {
      const data: CreatePlaylistReq = {
        title: formState.title,
        platform: formState.platform,
        platform_url: formState.platform_url,
        cover_url: formState.cover_url,
        cover_media_id: formState.coverMediaId || undefined,
        cover_preset_id: formState.coverPresetId || undefined,
        description: formState.description,
        sort_order: formState.sort_order,
        enabled: formState.enabled,
      }
      await playlistApi.create(data)
      message.success(t('playlist.form.addSuccess'))
    }
    emit('saved')
    emit('update:open', false)
  } catch {
    // 错误由拦截器处理
  } finally {
    saving.value = false
  }
}

function handleCancel() {
  emit('update:open', false)
}
</script>

<style scoped>
.platform-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.platform-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}
</style>
