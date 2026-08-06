import axios, { type AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

import router from '@/router'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import type { ApiResponse } from '@/types/api'

const CODE_TOKEN_EXPIRED = new Set([10013, 10014, 10015])

// 正在刷新时，把后续请求挂起，刷新完统一 resolve
let refreshing = false
let pendingQueue: Array<(token: string) => void> = []

const expireSession = (): void => {
  const userStore = useUserStore()
  const currentRoute = router.currentRoute.value
  const redirect = currentRoute.fullPath
  if (userStore.token) userStore.logout()
  useAuthModalStore().open('prompt', redirect)
  if (currentRoute.meta.requiresAuth) void router.replace({ name: 'profile' })
}

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const userStore = useUserStore()
  if (userStore.token) config.headers.Authorization = `Bearer ${userStore.token}`
  return config
})

http.interceptors.response.use(
  async (response: AxiosResponse<ApiResponse<unknown>>) => {
    if (response.data.code === 0) return response

    // token 过期，尝试用 refreshToken 无感续期
    if (CODE_TOKEN_EXPIRED.has(response.data.code)) {
      const userStore = useUserStore()
      const storedRefresh = userStore.refreshToken

      // 没有 refreshToken，直接弹登录框
      if (!storedRefresh) {
        expireSession()
        return Promise.reject(new Error(response.data.message || '登录已过期'))
      }

      // 已经有刷新在进行，排队等待
      if (refreshing) {
        return new Promise<AxiosResponse>((resolve, reject) => {
          pendingQueue.push((newToken: string) => {
            response.config.headers.Authorization = `Bearer ${newToken}`
            http(response.config).then(resolve).catch(reject)
          })
        })
      }

      refreshing = true
      try {
        // 直接用 axios 绕过拦截器，避免循环
        const { data } = await axios.post<ApiResponse<{ accessToken: string; refreshToken: string }>>(
          `${import.meta.env.VITE_API_BASE_URL}/auth/refresh`,
          { refreshToken: storedRefresh },
          { headers: { 'Content-Type': 'application/json' } },
        )

        if (data.code !== 0) throw new Error('refresh failed')

        const { accessToken, refreshToken: newRefresh } = data.data
        userStore.setTokens(accessToken, newRefresh)

        // 唤醒排队请求
        pendingQueue.forEach((cb) => cb(accessToken))
        pendingQueue = []

        // 重试原请求
        response.config.headers.Authorization = `Bearer ${accessToken}`
        return http(response.config)
      } catch {
        pendingQueue = []
        expireSession()
        return Promise.reject(new Error('登录已过期，请重新登录'))
      } finally {
        refreshing = false
      }
    }

    return Promise.reject(new Error(response.data.message || '请求失败'))
  },
  (error: AxiosError<ApiResponse<unknown>>) => {
    if (error.response?.status === 401) expireSession()
    return Promise.reject(error)
  },
)

export default http
