import axios, { type AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

import router from '@/router'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import type { ApiResponse } from '@/types/api'

const AUTH_EXPIRED_CODES = new Set([10013, 10014, 10015])

const expireSession = (): void => {
  const userStore = useUserStore()
  if (!userStore.token) return
  const currentRoute = router.currentRoute.value
  const redirect = currentRoute.fullPath
  userStore.logout()
  useAuthModalStore().open('prompt', redirect)
  if (currentRoute.meta.requiresAuth) void router.replace({ name: 'recommend' })
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
  (response: AxiosResponse<ApiResponse<unknown>>) => {
    if (response.data.code !== 0 && response.data.code !== 200) {
      if (AUTH_EXPIRED_CODES.has(response.data.code)) expireSession()
      return Promise.reject(new Error(response.data.message || '请求失败'))
    }
    return response
  },
  (error: AxiosError<ApiResponse<unknown>>) => {
    if (error.response?.status === 401) expireSession()
    return Promise.reject(error)
  },
)

export default http
