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
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()

const selfState = ref<CommunitySyncState | null>(null)
const loading = ref(false)
const syncing = ref(false)

const state = computed<CommunitySyncState | null>(() => props.state ?? selfState.value)
const status = computed(() => state.value?.sync_status ?? 'none')

const canPush = computed(() => props.published && status.value !== 'deleted')

// labelMap 各状态下按钮文案，未知状态回退「推送到社区」
const labelMap = computed<Record<string, string>>(() => ({
  synced: t('system.syncButton.unpublish'),
  hidden: t('system.syncButton.push'),
  failed: t('system.syncButton.retryFailed'),
  deleted: t('system.syncButton.deleted'),
}))

const label = computed(() => labelMap.value[status.value] || t('system.syncButton.push'))

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
  if (status.value === 'failed' && state.value?.last_error)
    return t('system.syncTooltip.failed', { error: state.value.last_error })
  if (!props.published && status.value !== 'deleted') return t('system.syncTooltip.unpublished')
  if (status.value === 'synced') return t('system.syncTooltip.synced')
  if (status.value === 'hidden') return t('system.syncTooltip.hidden')
  return t('system.syncTooltip.default')
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
      message.success(t('system.message.noChanges'))
    } else if (action === 'push') {
      message.success(t('system.message.pushed'))
    } else {
      message.success(t('system.message.unpublished'))
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
      title: t('system.modal.unpublishTitle'),
      content: t('system.modal.unpublishContent'),
      okText: t('system.modal.unpublishOk'),
      onOk: () => runSync('hide'),
    })
    return
  }
  runSync('push')
}
</script>
