<template>
  <a-modal
    v-model:open="visible"
    :title="isEdit ? '编辑歌单' : '添加第三方歌单'"
    :confirm-loading="saving"
    width="600px"
    @ok="handleSave"
    @cancel="handleCancel"
  >
    <a-form layout="vertical">
      <!-- 平台选择 -->
      <a-form-item label="平台" required>
        <a-select v-model:value="formState.platform" placeholder="选择平台">
          <a-select-option v-for="p in platformOptions" :key="p.id" :value="p.id">
            <div class="platform-option">
              <span class="platform-dot" :style="{ background: p.color }">{{ p.icon }}</span>
              {{ p.label }}
            </div>
          </a-select-option>
        </a-select>
      </a-form-item>

      <!-- 歌单链接 -->
      <a-form-item label="歌单链接" required>
        <a-input
          v-model:value="formState.platform_url"
          placeholder="输入第三方平台歌单链接，如 https://music.163.com/playlist/..."
        />
      </a-form-item>

      <!-- 歌单名称 -->
      <a-form-item label="歌单名称" required>
        <a-input v-model:value="formState.title" placeholder="输入歌单名称" />
      </a-form-item>

      <!-- 封面URL -->
      <a-form-item label="封面图 URL">
        <a-input v-model:value="formState.cover_url" placeholder="歌单封面图片URL（可选）" />
      </a-form-item>

      <!-- 描述 -->
      <a-form-item label="描述">
        <a-textarea
          v-model:value="formState.description"
          placeholder="简短描述这个歌单（可选）"
          :rows="2"
          :maxlength="500"
          show-count
        />
      </a-form-item>

      <!-- 排序 -->
      <a-form-item label="排序">
        <a-input-number v-model:value="formState.sort_order" :min="0" :max="9999" style="width: 120px" />
      </a-form-item>

      <!-- 是否展示 -->
      <a-form-item label="前台展示">
        <a-switch v-model:checked="formState.enabled" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { playlistApi } from '@/api/playlist'
import { PLATFORM_CONFIGS } from '@/utils/platformConfig'
import type { ThirdPartyPlaylist, CreatePlaylistReq, UpdatePlaylistReq } from '@/types/playlist'

const props = defineProps<{
  open: boolean
  playlist?: ThirdPartyPlaylist
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const visible = computed({
  get: () => props.open,
  set: (val: boolean) => emit('update:open', val),
})

const isEdit = computed(() => !!props.playlist)

const saving = ref(false)

const platformOptions = Object.values(PLATFORM_CONFIGS)

const formState = reactive<CreatePlaylistReq>({
  title: '',
  platform: '',
  platform_url: '',
  cover_url: '',
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
        formState.description = props.playlist.description
        formState.sort_order = props.playlist.sort_order
        formState.enabled = props.playlist.enabled
      } else {
        formState.title = ''
        formState.platform = ''
        formState.platform_url = ''
        formState.cover_url = ''
        formState.description = ''
        formState.sort_order = 0
        formState.enabled = true
      }
    }
  },
)

async function handleSave() {
  if (!formState.title.trim()) {
    message.warning('请输入歌单名称')
    return
  }
  if (!formState.platform) {
    message.warning('请选择平台')
    return
  }
  if (!formState.platform_url.trim()) {
    message.warning('请输入歌单链接')
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
        description: formState.description || undefined,
        sort_order: formState.sort_order,
        enabled: formState.enabled,
      }
      await playlistApi.update(props.playlist.id, data)
      message.success('更新成功')
    } else {
      const data: CreatePlaylistReq = {
        title: formState.title,
        platform: formState.platform,
        platform_url: formState.platform_url,
        cover_url: formState.cover_url,
        description: formState.description,
        sort_order: formState.sort_order,
        enabled: formState.enabled,
      }
      await playlistApi.create(data)
      message.success('添加成功')
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
