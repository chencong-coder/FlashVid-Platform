import http from './http'

import type { ApiResponse } from '@/types/api'
import type { LoginPayload, LoginResult } from '@/types/user'

export const login = (payload: LoginPayload) =>
  http.post<ApiResponse<LoginResult>>('/auth/login', payload)
