<template>
  <div class="article-card" @click="handleEdit">
    <div class="article-card-cover">
      <img v-if="article.cover_image" :src="getThumbUrl(article.cover_image, 400)" :alt="article.title" />
      <div v-else class="article-card-cover-placeholder">
        <FileTextOutlined />
      </div>
      <a-tag :color="statusColor" class="article-card-status">{{ statusText }}</a-tag>
      <a-tag :color="editorTypeColor" class="article-card-type">{{ editorTypeText }}</a-tag>
    </div>
    <div class="article-card-body">
      <h3 class="article-card-title">{{ article.title }}</h3>
      <p v-if="article.summary" class="article-card-summary">{{ article.summary }}</p>
      <div class="article-card-meta">
        <span v-if="article.category_name" class="article-card-category">
          <TagsOutlined /> {{ article.category_name }}
        </span>
        <span class="article-card-time">
          <ClockCircleOutlined /> {{ formatTime(article.published_at || article.created_at) }}
        </span>
      </div>
      <div class="article-card-footer">
        <span class="article-card-stats">
          <EyeOutlined /> {{ article.view_count }}
          <MessageOutlined /> {{ article.comment_count }}
        </span>
        <div class="article-card-actions" @click.stop>
          <PushSyncButton
            target-type="article"
            :target-id="article.id"
            :published="article.status === 2"
            @updated="emit('pushUpdated')"
          />
          <a-button size="small" type="text" @click="handleEdit">
            <EditOutlined />
          </a-button>
          <a-dropdown :trigger="['click']">
            <a-button size="small" type="text"><MoreOutlined /></a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item v-if="article.status === 1" @click="handlePublish">
                  <SendOutlined /> {{ $t('common.actions.publish') }}
                </a-menu-item>
                <a-menu-item v-if="article.status === 2" @click="handleUnpublish">
                  <StopOutlined /> {{ $t('common.actions.offline') }}
                </a-menu-item>
                <a-menu-item v-if="article.status !== 1" @click="handleDraft">
                  <FileTextOutlined /> {{ $t('article.card.moveToDraft') }}
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item danger @click="handleDelete">
                  <DeleteOutlined /> {{ $t('common.actions.delete') }}
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PushSyncButton from '@/components/community/PushSyncButton.vue'
import {
  FileTextOutlined, TagsOutlined, ClockCircleOutlined,
  EyeOutlined, MessageOutlined, EditOutlined, MoreOutlined,
  SendOutlined, StopOutlined, DeleteOutlined,
} from '@ant-design/icons-vue'
import type { Article } from '@/types/article'
import { getThumbUrl } from '@/utils/image'
import { formatDate } from '@/utils/datetime'

const { t } = useI18n()

const props = defineProps<{
  article: Article
}>()

const emit = defineEmits<{
  (e: 'edit', id: string): void
  (e: 'delete', id: string): void
  (e: 'statusChange', id: string, status: number): void
  (e: 'pushUpdated'): void
}>()

// 状态/编辑器类型枚举 → i18n key（文案见 article.status.* / article.editorType.*）
const statusKeyMap: Record<number, string> = {
  1: 'draft',
  2: 'published',
  3: 'unpublished',
}

const statusText = computed(() => {
  const key = statusKeyMap[props.article.status]
  return key ? t(`article.status.${key}`) : t('common.status.unknown')
})

const statusColorMap: Record<number, string> = {
  1: 'default',
  2: 'green',
  3: 'orange',
}

const statusColor = computed(() => statusColorMap[props.article.status] || 'default')

const editorTypeKeyMap: Record<number, string> = {
  1: 'markdown',
  2: 'html',
}

const editorTypeText = computed(() => t(`article.editorType.${editorTypeKeyMap[props.article.type] || 'markdown'}`))

const editorTypeColorMap: Record<number, string> = {
  1: 'blue',
  2: 'purple',
}

const editorTypeColor = computed(() => editorTypeColorMap[props.article.type] || 'blue')

// 相对时间 + 超过一周回退为本地化日期（跟随当前语言）
function formatTime(time: string) {
  if (!time) return ''
  const d = new Date(time)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 3600000) return t('article.card.minutesAgo', { count: Math.floor(diff / 60000) })
  if (diff < 86400000) return t('article.card.hoursAgo', { count: Math.floor(diff / 3600000) })
  if (diff < 604800000) return t('article.card.daysAgo', { count: Math.floor(diff / 86400000) })
  return formatDate(d)
}

function handleEdit() { emit('edit', props.article.id) }
function handlePublish() { emit('statusChange', props.article.id, 2) }
function handleUnpublish() { emit('statusChange', props.article.id, 3) }
function handleDraft() { emit('statusChange', props.article.id, 1) }
function handleDelete() { emit('delete', props.article.id) }
</script>

<style scoped>
.article-card {
  background: var(--bg-card, #fff);
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
  cursor: pointer;
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0,0,0,.06));
  transition: transform var(--transition-base, .25s), box-shadow var(--transition-base, .25s);
}
.article-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-dropdown, 0 4px 16px rgba(0,0,0,.08));
}
.article-card-cover {
  position: relative;
  height: 160px;
  overflow: hidden;
  background: #F5F7FC;
}
.article-card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.article-card-cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  color: #8A93A8;
}
.article-card-status {
  position: absolute;
  top: 8px;
  right: 8px;
}
.article-card-type {
  position: absolute;
  top: 8px;
  left: 8px;
  font-size: 11px;
}
.article-card-body {
  padding: 16px;
}
.article-card-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.article-card-summary {
  font-size: 13px;
  color: var(--text-secondary, #667085);
  margin-bottom: 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.article-card-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--text-tertiary, #8A93A8);
  margin-bottom: 12px;
}
.article-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid var(--border-color, #F5F7FC);
}
.article-card-stats {
  font-size: 12px;
  color: var(--text-tertiary, #8A93A8);
  display: flex;
  gap: 12px;
}
.article-card-actions {
  display: flex;
  gap: 4px;
}
</style>