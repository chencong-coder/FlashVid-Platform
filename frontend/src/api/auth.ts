import http from './http'

import type { ApiResponse } from '@/types/api'

// ===== 注册 =====
export interface RegisterPayload {
  username: string // 用户名，4-32 字符
  password: string // 密码，6-20 字符
  phone: string // 手机号，11 位
  code: string // 短信验证码，6 位
  birthday?: string // 生日，可选，YYYY-MM-DD
  email?: string // 邮箱，可选
}

export interface RegisterResult {
  userId: string
  username: string
}

export const register = (payload: RegisterPayload) =>
  http.post<ApiResponse<RegisterResult>>('/auth/register', payload)

// ===== 登录 =====
export interface LoginPayload {
  account: string // 用户名或手机号
  password: string
}

export interface LoginResult {
  userId: string
  username: string
  nickname: string
  avatar: string
  accessToken: string
  refreshToken: string
}

export const login = (payload: LoginPayload) =>
  http.post<ApiResponse<LoginResult>>('/auth/login', payload)

// ===== 刷新 Token =====
export interface RefreshResult {
  accessToken: string
  refreshToken: string
}

export const refreshToken = (token: string) =>
  http.post<ApiResponse<RefreshResult>>('/auth/refresh', { refreshToken: token })
