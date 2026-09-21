import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig, AxiosRequestConfig } from 'axios'

import { message } from 'ant-design-vue'
import { storage } from '@/utils/storage'
import type { ApiResponse } from '@/types/api'
import { MARKET_AUTH_EXPIRED_CODE } from '@/types/theme'

interface ApiRequest {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

let isRefreshing = false
let pendingRequests: Array<{ resolve: (token: string) => void; reject: (err: any) => void }> = []

function addPendingRequest(resolve: (token: string) => void, reject: (err: any) => void) {
  pendingRequests.push({ resolve, reject })
}

function resolvePendingRequests(token: string) {
  pendingRequests.forEach((req) => req.resolve(token))
  pendingRequests = []
}

function rejectPendingRequests(error: any) {
  pendingRequests.forEach((req) => req.reject(error))
  pendingRequests = []
}

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = storage.getToken()
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

request.interceptors.response.use(
  (response) => {
    const res = response.data as ApiResponse
    if (res.code === 0) {
      return res.data
    }
    message.error(res.msg || '请求失败')
    return Promise.reject(new Error(res.msg || '请求失败'))
  },
  async (error: AxiosError<ApiResponse>) => {
    const config = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
    const errMsg = error.response?.data?.msg || error.message || '网络错误'

    // 官方主题市场登录失效（业务码 401101）由市场模块自行刷新重试或弹回登录遮罩，不做全局提示
    if (error.response?.data?.code === MARKET_AUTH_EXPIRED_CODE) {
      return Promise.reject(error)
    }

    if (error.response?.status === 401 && !config?._retry) {
      // 登录接口只提示错误，不触发 token 刷新
      if (config?.url?.includes('/auth/login')) {
        message.error(errMsg)
        return Promise.reject(error)
      }

      const refreshToken = storage.getRefreshToken()
      if (!refreshToken) {
        message.error('登录已过期，请重新登录')
        storage.clear()
        window.location.href = '/auth/login'
        return Promise.reject(error)
      }

      if (!isRefreshing) {
        isRefreshing = true

        try {
          const refreshResponse = await axios.post<ApiResponse<any>>(
            `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/auth/refresh`,
            { refresh_token: refreshToken },
          )

          const newData = refreshResponse.data
          if (newData.code === 0) {
            const { access_token, refresh_token } = newData.data
            storage.setToken(access_token)
            storage.setRefreshToken(refresh_token)
            resolvePendingRequests(access_token)
          } else {
            message.error(newData.msg || '登录已过期，请重新登录')
            rejectPendingRequests(new Error('Token refresh failed'))
            storage.clear()
            window.location.href = '/auth/login'
            return Promise.reject(error)
          }
        } catch (refreshError) {
          const refreshMsg = (refreshError as any).response?.data?.msg || (refreshError as any).message || '登录已过期，请重新登录'
          message.error(refreshMsg)
          rejectPendingRequests(new Error('Token refresh failed'))
          storage.clear()
          window.location.href = '/auth/login'
          return Promise.reject(error)
        } finally {
          isRefreshing = false
        }
      }

      config._retry = true
      return new Promise((resolve, reject) => {
        addPendingRequest(
          (newToken: string) => {
            if (config.headers) {
              config.headers.Authorization = `Bearer ${newToken}`
            }
            resolve(request(config))
          },
          (err: any) => reject(err)
        )
      })
    }

    message.error(errMsg)
    return Promise.reject(error)
  },
)

export default request as unknown as ApiRequest