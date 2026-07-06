import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface BreadcrumbItem {
  title: string
  path?: string
}

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const breadcrumbs = ref<BreadcrumbItem[]>([])

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setBreadcrumbs(items: BreadcrumbItem[]) {
    breadcrumbs.value = items
  }

  return {
    sidebarCollapsed,
    breadcrumbs,
    toggleSidebar,
    setBreadcrumbs,
  }
})