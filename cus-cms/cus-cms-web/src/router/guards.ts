import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { setupApi } from '@/api/setup'
import { storage } from '@/utils/storage'

export function setupGuards(router: Router) {
  router.beforeEach(async (to, _from, next) => {
    // When visiting /setup, always fetch from backend, ignore cache
    if (to.path === '/setup') {
      let isInit = false
      try {
        const res = await setupApi.getStatus()
        isInit = res.initialized
      } catch {
        isInit = false
      }
      storage.setInitialized(isInit)

      if (isInit) {
        return next('/auth/login')
      }
      return next()
    }

    // For other pages, use cached status or fetch if never cached
    let initialized = storage.getInitialized()
    if (initialized === null) {
      try {
        const res = await setupApi.getStatus()
        initialized = res.initialized
      } catch {
        initialized = false
      }
      storage.setInitialized(initialized)
    }

    // If not initialized, redirect to setup
    if (!initialized) {
      return next('/setup')
    }

    // Auth check logic (only for initialized system)
    const authStore = useAuthStore()
    if (to.path.startsWith('/auth')) {
      if (authStore.isLoggedIn) return next('/dashboard')
      return next()
    }
    if (!authStore.isLoggedIn) {
      return next(`/auth/login?redirect=${to.path}`)
    }
    next()
  })
}
