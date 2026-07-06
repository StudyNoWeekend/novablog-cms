import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig, AxiosRequestConfig } from 'axios'

import { message } from 'ant-design-vue'
import { storage } from '@/utils/storage'
import type { ApiResponse } from '@/types/api'

interface ApiRequest {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

function addPendingRequest(cb: (token: string) => void) {
  pendingRequests.push(cb)
}

function resolvePendingRequests(token: string) {
  pendingRequests.forEach((cb) => cb(token))
  pendingRequests = []
}

function rejectPendingRequests() {
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

    if (error.response?.status === 401 && !config?._retry) {
      // 登录接口报错时不走 token 刷新逻辑，直接返回错误
      if (config?.url?.includes('/auth/login')) {
        return Promise.reject(error)
      }

      const refreshToken = storage.getRefreshToken()
      if (!refreshToken) {
        storage.clear()
        window.location.href = '/auth/login'
        return Promise.reject(error)
      }

      if (!isRefreshing) {
        isRefreshing = true

        try {
          const refreshResponse = await axios.post<ApiResponse<any>>(
            `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'}/auth/refresh`,
            { refresh_token: refreshToken },
          )

          const newData = refreshResponse.data
          if (newData.code === 0) {
            const { access_token, refresh_token } = newData.data
            storage.setToken(access_token)
            storage.setRefreshToken(refresh_token)
            resolvePendingRequests(access_token)
          } else {
            rejectPendingRequests()
            storage.clear()
            window.location.href = '/auth/login'
            return Promise.reject(error)
          }
        } catch {
          rejectPendingRequests()
          storage.clear()
          window.location.href = '/auth/login'
          return Promise.reject(error)
        } finally {
          isRefreshing = false
        }
      }

      config._retry = true
      return new Promise((resolve, reject) => {
        addPendingRequest((newToken: string) => {
          if (config.headers) {
            config.headers.Authorization = `Bearer ${newToken}`
          }
          resolve(request(config))
        })
      })
    }

    const errMsg = error.response?.data?.msg || error.message || '网络错误'
    message.error(errMsg)
    return Promise.reject(error)
  },
)

export default request as unknown as ApiRequest