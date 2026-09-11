import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { authApi } from '@/api/auth'
import { storage, storageSession } from '@/utils/storage'
import type { LoginReq, UserInfo } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const user = ref<UserInfo | null>(null)
  const router = useRouter()

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.nickname || user.value?.username || '')

  // 当前使用的存储后端，login 时由 remember 决定
  let currentStorage = storage

  function init() {
    // 优先从 localStorage 读取，如果不存在则从 sessionStorage 读取
    token.value = storage.getToken() || storageSession.getToken()
    refreshToken.value = storage.getRefreshToken() || storageSession.getRefreshToken()
    user.value = storage.getUser() || storageSession.getUser()

    // 确定当前使用的存储后端
    if (storage.getToken()) {
      currentStorage = storage
    } else if (storageSession.getToken()) {
      currentStorage = storageSession
    }
  }

  async function login(data: LoginReq, remember = true) {
    try {
      currentStorage = remember ? storage : storageSession

      const res: any = await authApi.login(data)
      token.value = res.access_token
      refreshToken.value = res.refresh_token

      currentStorage.setToken(res.access_token)
      currentStorage.setRefreshToken(res.refresh_token)

      await fetchUserInfo()

      const redirect = router.currentRoute.value.query.redirect as string
      router.push(redirect || '/dashboard')
      message.success('登录成功')
    } catch (e) {
      throw e  // re-throw original error, interceptor already showed message
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // 即使后端接口失败也清除本地状态
    } finally {
      token.value = null
      refreshToken.value = null
      user.value = null
      storage.clear()
      storageSession.clear()
      router.push('/auth/login')
    }
  }

  async function refreshTokenAction() {
    if (!refreshToken.value) return
    try {
      const res: any = await authApi.refresh(refreshToken.value)
      token.value = res.access_token
      refreshToken.value = res.refresh_token
      currentStorage.setToken(res.access_token)
      currentStorage.setRefreshToken(res.refresh_token)
    } catch {
      token.value = null
      refreshToken.value = null
      user.value = null
      currentStorage.clear()
      router.push('/auth/login')
    }
  }

  async function fetchUserInfo() {
    try {
      const res: any = await authApi.getProfile()
      user.value = res
      currentStorage.setUser(res)
    } catch {
      // 获取用户信息失败不阻断流程
    }
  }

  init()

  return {
    token,
    refreshToken,
    user,
    isLoggedIn,
    username,
    init,
    login,
    logout,
    refreshTokenAction,
    fetchUserInfo,
  }
})