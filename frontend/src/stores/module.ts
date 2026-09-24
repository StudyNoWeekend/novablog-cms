import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { moduleApi } from '@/api/module'
import type { ModuleConfig } from '@/types/module'

export type ModuleKey = Exclude<keyof ModuleConfig, 'updated_at'>

export const useModuleStore = defineStore('module', () => {
  const config = ref<ModuleConfig | null>(null)

  // 在途拉取 Promise，避免多个组件同时挂载时重复请求
  let fetchPromise: Promise<void> | null = null

  const loaded = computed(() => config.value !== null)

  async function fetchConfig(force = false) {
    if (loaded.value && !force) return
    if (fetchPromise) return fetchPromise
    fetchPromise = moduleApi
      .getConfig()
      .then((res) => {
        config.value = res
      })
      .finally(() => {
        fetchPromise = null
      })
    await fetchPromise
  }

  // 未加载完成时默认视为开启，避免菜单闪藏
  function isEnabled(key: ModuleKey): boolean {
    if (!config.value) return true
    return config.value[key]
  }

  return {
    config,
    loaded,
    fetchConfig,
    isEnabled,
  }
})
