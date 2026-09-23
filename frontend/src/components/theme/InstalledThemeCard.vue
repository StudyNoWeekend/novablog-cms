<template>
  <div class="installed-card">
    <div class="installed-card__cover" :style="{ background: gradient }">
      <img
        v-if="marketPreview && !coverFailed"
        :src="market!.preview"
        :alt="displayName"
        loading="lazy"
        @error="coverFailed = true"
      />
      <span v-else class="installed-card__cover-text">{{ displayName.slice(0, 1) }}</span>
      <a-tag v-if="theme.active" color="success" class="installed-card__badge">
        <CheckCircleOutlined /> 使用中
      </a-tag>
    </div>

    <div class="installed-card__body">
      <h3 class="installed-card__title" :title="displayName">{{ displayName }}</h3>
      <div class="installed-card__meta">
        <span class="installed-card__version">v{{ theme.version }}</span>
        <span class="installed-card__engine">{{ theme.engine }}</span>
        <span class="installed-card__time">{{ installedTime }}</span>
      </div>
      <p v-if="displayDesc" class="installed-card__desc">{{ displayDesc }}</p>
    </div>

    <div class="installed-card__actions" @click.stop>
      <a-button
        size="small"
        type="primary"
        :disabled="theme.active"
        :loading="activating"
        @click="emit('activate')"
      >
        {{ theme.active ? '使用中' : '启用' }}
      </a-button>
      <a-button size="small" @click="emit('preview')">
        <template #icon><EyeOutlined /></template>
        预览
      </a-button>
      <a-button
        v-if="theme.source === 'official'"
        size="small"
        type="dashed"
        :loading="updating"
        @click="emit('update')"
      >
        <template #icon><ReloadOutlined /></template>
        更新
      </a-button>
      <a-popconfirm
        title="卸载后主题目录与记录将被删除，确定卸载？"
        ok-text="卸载"
        cancel-text="取消"
        :ok-button-props="{ danger: true }"
        @confirm="emit('uninstall')"
      >
        <a-button size="small" danger :disabled="theme.active">卸载</a-button>
      </a-popconfirm>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CheckCircleOutlined, EyeOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import type { InstalledTheme, ThemeItem } from '@/types/theme'
import { isPreviewURL, themeGradient } from '@/utils/themeDisplay'

const props = defineProps<{
  theme: InstalledTheme
  /** 市场元数据（按 market_id 关联），用于对齐市场展示；缺省时回退制品内信息 */
  market?: ThemeItem | null
  activating?: boolean
  updating?: boolean
}>()

const emit = defineEmits<{
  (e: 'activate'): void
  (e: 'preview'): void
  (e: 'uninstall'): void
  (e: 'update'): void
}>()

const coverFailed = ref(false)

const marketPreview = computed(() => (props.market ? isPreviewURL(props.market.preview) : false))
const displayName = computed(() => props.market?.title || props.theme.name)
const displayDesc = computed(() => props.market?.description || props.theme.description)
// 渐变按市场主题类型取色（themeGradient 以类型为键），无市场数据时走默认色
const gradient = computed(() => themeGradient(props.market?.type || ''))

watch(
  () => props.theme.id,
  () => {
    coverFailed.value = false
  },
)

const installedTime = computed(() => {
  const d = new Date(props.theme.created_at)
  return Number.isNaN(d.getTime()) ? '' : d.toLocaleDateString('zh-CN')
})
</script>

<style scoped>
.installed-card {
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.installed-card:hover {
  border-color: var(--primary-color, #1677ff);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
}

.installed-card__cover {
  position: relative;
  height: 110px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.installed-card__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.installed-card__cover-text {
  color: rgba(255, 255, 255, 0.92);
  font-size: 36px;
  font-weight: 600;
  user-select: none;
}

.installed-card__badge {
  position: absolute;
  top: 8px;
  left: 8px;
}

.installed-card__body {
  flex: 1;
  padding: 12px 14px;
}

.installed-card__title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.installed-card__meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

.installed-card__version {
  font-family: ui-monospace, Menlo, monospace;
}

.installed-card__desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.installed-card__actions {
  display: flex;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}

.installed-card__actions .ant-btn {
  flex: 1;
}
</style>
