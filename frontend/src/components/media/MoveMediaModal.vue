<template>
  <a-modal
    :open="visible"
    :title="$t('media.move.title')"
    :confirm-loading="submitting"
    @ok="handleOk"
    @cancel="emit('update:visible', false)"
  >
    <p class="move-hint">{{ hint }}</p>
    <a-spin :spinning="folderLoading">
      <a-tree
        v-if="treeData.length > 0"
        :tree-data="treeData"
        :selected-keys="selectedKeys"
        :expanded-keys="expandedKeys"
        block-node
        @select="handleSelect"
        @expand="handleExpand"
      >
        <template #title="{ title, count }">
          <span class="tree-folder">
            <FolderOutlined />
            <span class="tree-title">{{ title }}</span>
            <span v-if="count > 0" class="tree-count">{{ $t('media.itemCount', { count }) }}</span>
          </span>
        </template>
      </a-tree>
      <a-empty v-else :description="$t('media.folder.empty')" />
    </a-spin>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { FolderOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaFolderNode } from '@/api/media'
import { useMediaFolders, folderDisplayName } from '@/composables/useMediaFolders'
import type { TreeProps } from 'ant-design-vue'

const props = defineProps<{
  visible: boolean
  /** 移动模式：媒体 ID 列表（media 模式）或文件夹（folder 模式），二选一 */
  mediaIds?: string[]
  /** folder 模式：待移动的文件夹 */
  folder?: MediaFolderNode | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}>()

const { folderTree, folderLoading, fetchFolderTree, collectSubtreeIds } = useMediaFolders()

const { t } = useI18n()

const ROOT_KEY = '__root__'

const selectedKeys = ref<string[]>([])
const expandedKeys = ref<string[]>([])
const submitting = ref(false)

const hint = computed(() => {
  if (props.folder) return t('media.move.folderHint', { name: props.folder.name })
  const count = props.mediaIds?.length || 0
  return t('media.move.filesHint', { count })
})

interface TreeItem {
  title: string
  key: string
  count: number
  children?: TreeItem[]
}

/** 构建树数据：根节点（全部文件）+ 文件夹树，folder 模式下排除自身子树 */
const treeData = computed<TreeItem[]>(() => {
  const exclude = new Set<string>(props.folder ? collectSubtreeIds(props.folder) : [])

  const build = (nodes: MediaFolderNode[]): TreeItem[] =>
    nodes
      .filter((n) => !exclude.has(n.id))
      .map((n) => ({
        title: folderDisplayName(n.name, n.module_key),
        key: n.id,
        count: n.media_count,
        children: build(n.children || []),
      }))

  return [{ title: t('media.folder.rootName'), key: ROOT_KEY, count: 0, children: build(folderTree.value) }]
})

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      fetchFolderTree()
      selectedKeys.value = []
      expandedKeys.value = [ROOT_KEY]
    }
  }
)

function handleSelect(keys: (string | number)[]) {
  selectedKeys.value = keys.map(String)
}

const handleExpand: TreeProps['onExpand'] = (keys) => {
  expandedKeys.value = keys.map(String)
}

async function handleOk() {
  const targetId = selectedKeys.value[0] || ROOT_KEY
  const folderId = targetId === ROOT_KEY ? null : targetId
  submitting.value = true
  try {
    if (props.folder) {
      await mediaApi.updateFolder(props.folder.id, { parent_id: folderId ?? '' })
      message.success(t('media.folder.moved'))
    } else {
      await mediaApi.moveMedia(props.mediaIds || [], folderId)
      message.success(t('media.move.filesMoved'))
    }
    emit('update:visible', false)
    emit('success')
  } catch (err: any) {
    message.error(err?.message || t('media.move.failed'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.move-hint {
  color: var(--text-secondary, #667085);
  margin-bottom: 12px;
}

.tree-folder {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.tree-folder :deep(.anticon) {
  color: #B7791F;
}

.tree-count {
  color: var(--text-tertiary, #8A93A8);
  font-size: 12px;
}
</style>
