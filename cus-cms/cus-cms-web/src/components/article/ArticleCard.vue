<template>
  <div class="article-card" @click="handleEdit">
    <div class="article-card-cover">
      <img v-if="article.cover_image" :src="article.cover_image" :alt="article.title" />
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
          <a-button size="small" type="text" @click="handleEdit">
            <EditOutlined />
          </a-button>
          <a-dropdown :trigger="['click']">
            <a-button size="small" type="text"><MoreOutlined /></a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item v-if="article.status === 1" @click="handlePublish">
                  <SendOutlined /> 发布
                </a-menu-item>
                <a-menu-item v-if="article.status === 2" @click="handleUnpublish">
                  <StopOutlined /> 下架
                </a-menu-item>
                <a-menu-item v-if="article.status !== 1" @click="handleDraft">
                  <FileTextOutlined /> 转为草稿
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item danger @click="handleDelete">
                  <DeleteOutlined /> 删除
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
import {
  FileTextOutlined, TagsOutlined, ClockCircleOutlined,
  EyeOutlined, MessageOutlined, EditOutlined, MoreOutlined,
  SendOutlined, StopOutlined, DeleteOutlined,
} from '@ant-design/icons-vue'
import type { Article } from '@/types/article'

const props = defineProps<{
  article: Article
}>()

const emit = defineEmits<{
  (e: 'edit', id: string): void
  (e: 'delete', id: string): void
  (e: 'statusChange', id: string, status: number): void
}>()

const statusMap: Record<number, { text: string; color: string }> = {
  1: { text: '草稿', color: 'default' },
  2: { text: '已发布', color: 'green' },
  3: { text: '已下架', color: 'orange' },
}

const statusText = computed(() => statusMap[props.article.status]?.text || '未知')
const statusColor = computed(() => statusMap[props.article.status]?.color || 'default')

const editorTypeMap: Record<number, { text: string; color: string }> = {
  1: { text: 'MD', color: 'blue' },
  2: { text: 'HTML', color: 'purple' },
}

const editorTypeText = computed(() => editorTypeMap[props.article.type]?.text || 'MD')
const editorTypeColor = computed(() => editorTypeMap[props.article.type]?.color || 'blue')

function formatTime(time: string) {
  if (!time) return ''
  const d = new Date(time)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
  return d.toLocaleDateString('zh-CN')
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
  background: #f1f5f9;
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
  color: #cbd5e1;
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
  color: var(--text-secondary, #64748b);
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
  color: var(--text-tertiary, #94a3b8);
  margin-bottom: 12px;
}
.article-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid var(--border-color, #f1f5f9);
}
.article-card-stats {
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
  display: flex;
  gap: 12px;
}
.article-card-actions {
  display: flex;
  gap: 4px;
}
</style>