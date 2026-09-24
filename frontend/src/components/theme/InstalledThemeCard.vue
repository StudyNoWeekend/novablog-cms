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
      <a-tag v-if="updateAvailable" color="warning" class="installed-card__badge--update">
        <ArrowUpOutlined /> 有新版本
      </a-tag>
    </div>

    <div class="installed-card__body">
      <h3 class="installed-card__title" :title="displayName">{{ displayName }}</h3>
      <div class="installed-card__meta">
        <span class="installed-card__version" :class="{ 'installed-card__version--update': updateAvailable }">
          <template v-if="updateAvailable">v{{ theme.version }} → v{{ market!.version }}</template>
          <template v-else>v{{ theme.version }}</template>
        </span>
        <span class="installed-card__engine">{{ theme.engine }}</span>
        <span v-if="typeLabel" class="installed-card__type">{{ typeLabel }}</span>
        <span class="installed-card__time">{{ installedTime }}</span>
      </div>
      <p v-if="displayDesc" class="installed-card__desc" :title="displayDesc">{{ displayDesc }}</p>
      <div v-if="marketStyles.length" class="installed-card__tags">
        <a-tag v-for="s in marketStyles.slice(0, 3)" :key="s">{{ s }}</a-tag>
      </div>
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
      <a-tooltip v-if="theme.source === 'official'" :title="updateTooltip">
        <a-button
          size="small"
          :type="updateAvailable ? 'primary' : 'default'"
          :disabled="latest"
          :loading="updating"
          @click="emit('update')"
        >
          <template #icon><ReloadOutlined /></template>
          {{ updateButtonText }}
        </a-button>
      </a-tooltip>
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
import { ArrowUpOutlined, CheckCircleOutlined, EyeOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import type { InstalledTheme, ThemeItem } from '@/types/theme'
import {
  hasNewVersion,
  isPreviewURL,
  themeDisplayName,
  themeGradient,
  THEME_TYPE_LABELS,
} from '@/utils/themeDisplay'

const props = defineProps<{
  theme: InstalledTheme
  /** 市场元数据（按 market_id 关联），用于对齐市场展示与版本比对；缺省时回退制品内信息 */
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
// 市场标题曾误填主题 id，经映射还原为规范展示名；无市场数据时回退制品内 name（清单规范名）
const displayName = computed(() =>
  props.market ? themeDisplayName(props.market.title, props.market.slug) : props.theme.name,
)
const displayDesc = computed(() => props.market?.description || props.theme.description)
// 渐变按市场主题类型取色（themeGradient 以类型为键），无市场数据时走默认色
const gradient = computed(() => themeGradient(props.market?.type || ''))
const typeLabel = computed(() => {
  const type = props.market?.type
  return type ? THEME_TYPE_LABELS[type] || type : ''
})
const marketStyles = computed(() => props.market?.styles?.filter(Boolean) ?? [])

// ===== 版本比对：本地已安装版本 vs 官方市场最新版本 =====
// 官方版本号不同即提示"有新版本可更新"；市场元数据未拉到时保持原"更新"入口
const updateAvailable = computed(() => hasNewVersion(props.theme.version, props.market?.version))
// 本地版本与官方一致且市场元数据可用时，无需再更新
const latest = computed(() => !!props.market && !updateAvailable.value)
const updateButtonText = computed(() => {
  if (updateAvailable.value) return '有新版本可更新'
  if (latest.value) return '已是最新'
  return '更新'
})
const updateTooltip = computed(() => {
  if (updateAvailable.value) return `本地 v${props.theme.version}，官方最新 v${props.market?.version}`
  if (latest.value) return `本地 v${props.theme.version} 已与官方最新版本一致`
  return '从官方市场拉取最新版本'
})

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
  height: 140px;
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

.installed-card__badge--update {
  position: absolute;
  bottom: 8px;
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
  flex-shrink: 0;
}

.installed-card__version--update {
  color: #fa8c16;
  font-weight: 600;
}

.installed-card__type {
  flex-shrink: 0;
}

.installed-card__time {
  margin-left: auto;
  flex-shrink: 0;
}

.installed-card__desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.installed-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 0;
  margin-top: 8px;
}

.installed-card__tags .ant-tag {
  margin-inline-end: 4px;
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
