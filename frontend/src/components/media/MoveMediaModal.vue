<template>
  <a-modal
    :open="visible"
    title="移动到文件夹"
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
            <span v-if="count > 0" class="tree-count">{{ count }} 项</span>
          </span>
        </template>
      </a-tree>
      <a-empty v-else description="暂无文件夹" />
    </a-spin>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { FolderOutlined } from '@ant-design/icons-vue'
import { mediaApi } from '@/api/media'
import type { MediaFolderNode } from '@/api/media'
import { useMediaFolders } from '@/composables/useMediaFolders'
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

const ROOT_KEY = '__root__'

const selectedKeys = ref<string[]>([])
const expandedKeys = ref<string[]>([])
const submitting = ref(false)

const hint = computed(() => {
  if (props.folder) return `将文件夹「${props.folder.name}」移动到所选位置（默认移动到根目录）`
  const count = props.mediaIds?.length || 0
  return `将选中的 ${count} 个文件移动到所选位置（默认移动到根目录）`
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
        title: n.name,
        key: n.id,
        count: n.media_count,
        children: build(n.children || []),
      }))

  return [{ title: '全部文件', key: ROOT_KEY, count: 0, children: build(folderTree.value) }]
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
      message.success('文件夹已移动')
    } else {
      await mediaApi.moveMedia(props.mediaIds || [], folderId)
      message.success('文件已移动')
    }
    emit('update:visible', false)
    emit('success')
  } catch (err: any) {
    message.error(err?.message || '移动失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.move-hint {
  color: var(--text-secondary, #595959);
  margin-bottom: 12px;
}

.tree-folder {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.tree-folder :deep(.anticon) {
  color: #faad14;
}

.tree-count {
  color: var(--text-tertiary, #bfbfbf);
  font-size: 12px;
}
</style>
