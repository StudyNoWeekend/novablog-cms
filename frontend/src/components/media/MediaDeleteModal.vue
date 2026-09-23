<template>
  <a-modal
    :open="visible"
    title="删除媒体文件"
    :ok-text="used ? '仍要删除' : '删除'"
    :ok-button-props="{ danger: true }"
    :confirm-loading="deleting"
    @ok="handleDelete"
    @cancel="handleCancel"
  >
    <a-spin :spinning="loading">
      <div v-if="media" class="delete-target">
        <img
          v-if="media.file_type === 1"
          :src="media.thumb_url || media.url"
          :alt="media.filename"
          class="delete-thumb"
        />
        <div v-else class="delete-thumb video-thumb">
          <PlayCircleOutlined />
        </div>
        <span class="delete-filename" :title="media.filename">{{ media.filename }}</span>
      </div>

      <template v-if="!loading">
        <a-alert
          v-if="used"
          type="error"
          show-icon
          :message="`该文件被 ${total} 处内容引用，删除后对应位置将无法显示图片`"
          style="margin-top: 12px"
        />
        <a-alert
          v-else
          type="warning"
          show-icon
          message="删除后文件将从存储中永久移除，不可恢复"
          style="margin-top: 12px"
        />

        <div v-if="used && groups.length > 0" class="usage-groups">
          <div v-for="group in groups" :key="group.module" class="usage-group">
            <div class="usage-module">
              <a-tag color="processing">{{ group.module }}</a-tag>
              <span class="usage-count">{{ group.items.length }} 处</span>
            </div>
            <ul class="usage-list">
              <li v-for="item in group.items" :key="`${item.id}-${item.field}`">
                <span class="usage-title" :title="item.title">{{ item.title }}</span>
                <a-tag>{{ item.field }}</a-tag>
              </li>
            </ul>
          </div>
        </div>
      </template>
    </a-spin>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlayCircleOutlined } from '@ant-design/icons-vue'
import { mediaApi, type MediaItem, type MediaUsageGroup } from '@/api/media'

const props = defineProps<{
  visible: boolean
  media: MediaItem | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}>()

const loading = ref(false)
const deleting = ref(false)
const used = ref(false)
const total = ref(0)
const groups = ref<MediaUsageGroup[]>([])

watch(
  () => props.visible,
  (val) => {
    if (val && props.media) {
      fetchUsages()
    }
  }
)

async function fetchUsages() {
  if (!props.media) return
  loading.value = true
  used.value = false
  total.value = 0
  groups.value = []
  try {
    const res = await mediaApi.getUsages(props.media.id)
    used.value = res.used
    total.value = res.total
    groups.value = res.groups || []
  } catch {
    // 查询失败时按未引用处理，仅提示普通删除风险
  } finally {
    loading.value = false
  }
}

async function handleDelete() {
  if (!props.media) return
  deleting.value = true
  try {
    await mediaApi.remove(props.media.id, used.value)
    message.success('删除成功')
    emit('success')
    emit('update:visible', false)
  } catch (err: any) {
    message.error(err?.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

function handleCancel() {
  emit('update:visible', false)
}
</script>

<style scoped>
.delete-target {
  display: flex;
  align-items: center;
  gap: 12px;
}

.delete-thumb {
  width: 48px;
  height: 48px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--border-color, #f0f0f0);
  flex-shrink: 0;
}

.video-thumb {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: var(--text-secondary, #999);
  background: var(--bg-secondary, #f5f5f5);
}

.delete-filename {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.usage-groups {
  margin-top: 12px;
  max-height: 260px;
  overflow-y: auto;
}

.usage-group {
  margin-bottom: 12px;
}

.usage-group:last-child {
  margin-bottom: 0;
}

.usage-module {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.usage-count {
  font-size: 12px;
  color: var(--text-secondary, #999);
}

.usage-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.usage-list li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 3px 0 3px 8px;
  font-size: 13px;
}

.usage-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
