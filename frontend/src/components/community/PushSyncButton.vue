<template>
  <a-tooltip :title="tooltip">
    <a-button
      size="small"
      :type="buttonType"
      :danger="isDanger"
      :loading="loading"
      :disabled="isDisabled"
      @click="handleClick"
    >
      <template #icon>
        <component :is="icon" />
      </template>
      {{ label }}
    </a-button>
  </a-tooltip>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  CloudUploadOutlined,
  CloudDownloadOutlined,
  ReloadOutlined,
  WarningOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'

import { communityPushApi, type CommunitySyncState } from '@/api/communityPush'

// PushSyncButton 内容推送到社区的统一按钮：
// 按推送状态渲染「推送到社区 / 从社区下架 / 推送失败重试」，状态变化后向父组件发 updated 事件。
const props = withDefaults(
  defineProps<{
    targetType: string
    targetId: string
    /** 内容是否处于已发布状态（决定能否推送；不传视为可推送） */
    published?: boolean
    /** 外部已批量拉取的状态（传入则不再自行请求） */
    state?: CommunitySyncState | null
  }>(),
  {
    published: true,
    state: undefined,
  },
)

const emit = defineEmits<{ (e: 'updated', state: CommunitySyncState): void }>()

const selfState = ref<CommunitySyncState | null>(null)
const loading = ref(false)
const syncing = ref(false)

const state = computed<CommunitySyncState | null>(() => props.state ?? selfState.value)
const status = computed(() => state.value?.sync_status ?? 'none')

const canPush = computed(() => props.published && status.value !== 'deleted')

const label = computed(() => {
  switch (status.value) {
    case 'synced':
      return '从社区下架'
    case 'hidden':
      return '推送到社区'
    case 'failed':
      return '推送失败，重试'
    case 'deleted':
      return '已从社区删除'
    default:
      return '推送到社区'
  }
})

const icon = computed(() => {
  switch (status.value) {
    case 'synced':
      return CloudDownloadOutlined
    case 'failed':
      return WarningOutlined
    case 'deleted':
      return DeleteOutlined
    default:
      return CloudUploadOutlined
  }
})

const buttonType = computed(() => {
  if (status.value === 'failed') return 'default'
  if (status.value === 'synced') return 'default'
  return 'primary'
})

const isDanger = computed(() => status.value === 'failed')

const isDisabled = computed(() => {
  if (loading.value || syncing.value) return true
  if (status.value === 'deleted') return true
  if (!props.published && (status.value === 'none' || status.value === 'hidden')) return true
  return false
})

const tooltip = computed(() => {
  if (status.value === 'failed' && state.value?.last_error) return `最近一次推送失败：${state.value.last_error}`
  if (!props.published && status.value !== 'deleted') return '仅已发布的内容可推送到社区'
  if (status.value === 'synced') return '内容已上架社区；在 CMS 的后续修改会自动同步'
  if (status.value === 'hidden') return '已从社区下架；点击重新推送上架'
  return '将该内容推送到 NovaBlog 社区；开启后后续修改自动同步'
})

async function loadState() {
  if (!props.targetId) return
  try {
    selfState.value = await communityPushApi.getState(props.targetType, props.targetId)
  } catch {
    // 状态获取失败保持 none，由全局拦截器提示错误
  }
}

onMounted(loadState)
watch(() => props.targetId, loadState)

async function runSync(action: 'push' | 'hide') {
  syncing.value = true
  try {
    const result = await communityPushApi.sync(props.targetType, props.targetId, action)
    if (result.result === 'skipped') {
      message.success('内容无变化，社区已是最新')
    } else if (action === 'push') {
      message.success('已推送到 NovaBlog 社区')
    } else {
      message.success('已从社区下架')
    }
    await loadState()
    if (selfState.value) emit('updated', selfState.value)
  } finally {
    syncing.value = false
  }
}

function handleClick() {
  if (status.value === 'synced') {
    Modal.confirm({
      title: '从社区下架',
      content: '下架后社区读者不可见，但社区数据保留；重新推送即可恢复。确认下架？',
      okText: '下架',
      onOk: () => runSync('hide'),
    })
    return
  }
  runSync('push')
}
</script>
