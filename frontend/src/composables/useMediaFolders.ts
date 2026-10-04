import { ref } from 'vue'
import { mediaApi } from '@/api/media'
import type { MediaFolderNode } from '@/api/media'

/** 媒体文件夹树状态与查找工具（媒体库页与选图弹窗共用） */
export function useMediaFolders() {
  const folderTree = ref<MediaFolderNode[]>([])
  const folderLoading = ref(false)

  async function fetchFolderTree() {
    folderLoading.value = true
    try {
      const res = await mediaApi.getFolderTree()
      folderTree.value = res.list || []
    } finally {
      folderLoading.value = false
    }
  }

  /** 在树中查找指定文件夹 */
  function findFolder(tree: MediaFolderNode[], id: string): MediaFolderNode | null {
    for (const node of tree) {
      if (node.id === id) return node
      const found = findFolder(node.children || [], id)
      if (found) return found
    }
    return null
  }

  /** 查找从顶级到指定文件夹的路径（不含根目录）；找不到返回 [] */
  function findFolderPath(tree: MediaFolderNode[], id: string): MediaFolderNode[] {
    for (const node of tree) {
      if (node.id === id) return [node]
      const sub = findFolderPath(node.children || [], id)
      if (sub.length > 0) return [node, ...sub]
    }
    return []
  }

  /** 按业务模块 key 查找模块文件夹 */
  function findModuleFolder(tree: MediaFolderNode[], moduleKey: string): MediaFolderNode | null {
    for (const node of tree) {
      if (node.module_key === moduleKey) return node
      const found = findModuleFolder(node.children || [], moduleKey)
      if (found) return found
    }
    return null
  }

  /** 收集文件夹及其全部子孙 ID（移动文件夹时用于排除自身子树） */
  function collectSubtreeIds(folder: MediaFolderNode): string[] {
    const ids = [folder.id]
    for (const child of folder.children || []) {
      ids.push(...collectSubtreeIds(child))
    }
    return ids
  }

  return {
    folderTree,
    folderLoading,
    fetchFolderTree,
    findFolder,
    findFolderPath,
    findModuleFolder,
    collectSubtreeIds,
  }
}
