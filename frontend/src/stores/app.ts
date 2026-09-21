import { defineStore } from 'pinia'
import { ref } from 'vue'
import { systemApi } from '@/api/system'

export interface BreadcrumbItem {
  title: string
  path?: string
}

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const breadcrumbs = ref<BreadcrumbItem[]>([])
  const appVersion = ref('')

  // 在途拉取 Promise，避免多个组件挂载时重复请求
  let versionPromise: Promise<void> | null = null

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setBreadcrumbs(items: BreadcrumbItem[]) {
    breadcrumbs.value = items
  }

  async function fetchVersion() {
    if (appVersion.value || versionPromise) return
    versionPromise = systemApi
      .getVersion()
      .then((res) => {
        appVersion.value = res.version
      })
      .finally(() => {
        versionPromise = null
      })
    await versionPromise
  }

  return {
    sidebarCollapsed,
    breadcrumbs,
    appVersion,
    toggleSidebar,
    setBreadcrumbs,
    fetchVersion,
  }
})