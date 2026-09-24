import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useModuleStore, type ModuleKey } from '@/stores/module'
import { setupApi } from '@/api/setup'
import { storage } from '@/utils/storage'

// 受模块开关控制的路由根段（含详情/编辑等子路由，按第一段匹配）
const pathModuleMap: Record<string, ModuleKey> = {
  articles: 'article_enabled',
  media: 'media_enabled',
  playlists: 'music_enabled',
  videos: 'video_enabled',
  travels: 'travel_enabled',
  portfolios: 'portfolio_enabled',
  equipments: 'equipment_enabled',
  projects: 'project_enabled',
  'open-sources': 'open_source_enabled',
}

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
    // 缓存为空、或目标是登录页时实时查询后端：
    // 防止「浏览器缓存 initialized=true 但后端数据库已重置」的错位场景
    // （如容器重新部署/清库后，老浏览器仍被放行到登录页而非首装向导）
    if (initialized === null || to.path.startsWith('/auth')) {
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

    // 模块开关拦截：被关闭模块的页面直接重定向到工作台
    const moduleKey = pathModuleMap[to.path.split('/')[1]]
    if (moduleKey) {
      const moduleStore = useModuleStore()
      try {
        await moduleStore.fetchConfig()
      } catch {
        // 拉取失败按开启处理，避免把管理员锁在页面外
      }
      if (!moduleStore.isEnabled(moduleKey)) {
        return next('/dashboard')
      }
    }
    next()
  })
}
