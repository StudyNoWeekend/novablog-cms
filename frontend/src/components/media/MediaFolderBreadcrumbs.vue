<template>
  <a-breadcrumb class="folder-breadcrumbs">
    <a-breadcrumb-item>
      <a @click.prevent="emit('navigate', null)">
        <HddOutlined /> 全部文件
      </a>
    </a-breadcrumb-item>
    <a-breadcrumb-item v-for="folder in path" :key="folder.id">
      <a @click.prevent="emit('navigate', folder.id)" :class="{ current: folder.id === currentId }">
        {{ folder.name }}
      </a>
    </a-breadcrumb-item>
  </a-breadcrumb>
</template>

<script setup lang="ts">
import { HddOutlined } from '@ant-design/icons-vue'
import type { MediaFolderNode } from '@/api/media'

defineProps<{
  /** 从顶级到当前文件夹的路径，根目录时为空数组 */
  path: MediaFolderNode[]
  currentId: string
}>()

const emit = defineEmits<{
  (e: 'navigate', folderId: string | null): void
}>()
</script>

<style scoped>
.folder-breadcrumbs {
  font-size: 14px;
}

.folder-breadcrumbs :deep(.ant-breadcrumb-link a) {
  color: var(--text-secondary, #667085);
  cursor: pointer;
}

.folder-breadcrumbs :deep(.ant-breadcrumb-link a:hover) {
  color: var(--primary, #526FE8);
}

.folder-breadcrumbs :deep(a.current) {
  color: var(--primary, #526FE8);
  font-weight: 500;
}
</style>
