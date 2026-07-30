import { defineStore } from 'pinia'

import type { LoginResult } from '@/api/auth'
import type { UserProfile } from '@/types/user'
import { storage } from '@/utils/storage'

interface UserState {
  token: string
  refreshToken: string
  profile: UserProfile | null
}

const TOKEN_KEY = 'flashvid_token'
const REFRESH_TOKEN_KEY = 'flashvid_refresh_token'
const PROFILE_KEY = 'flashvid_profile'
const MAX_AGE = 30 * 24 * 60 * 60 * 1000

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: storage.get<string>(TOKEN_KEY) ?? '',
    refreshToken: storage.get<string>(REFRESH_TOKEN_KEY) ?? '',
    profile: storage.get<UserProfile>(PROFILE_KEY),
  }),
  getters: {
    isLoggedIn: (state): boolean => Boolean(state.token),
  },
  actions: {
    setSession(result: LoginResult): void {
      // 登录响应只含基础字段，其余（bio / 关注数等）由 getUserInfo 拉取补全
      const profile: UserProfile = {
        id: String(result.userId),
        nickname: result.nickname || result.username,
        avatar: result.avatar,
        bio: '',
        following: 0,
        followers: 0,
        likes: 0,
      }
      this.token = result.accessToken
      this.refreshToken = result.refreshToken
      this.profile = profile
      storage.set(TOKEN_KEY, result.accessToken, MAX_AGE)
      storage.set(REFRESH_TOKEN_KEY, result.refreshToken, MAX_AGE)
      storage.set(PROFILE_KEY, profile, MAX_AGE)
    },
    // 刷新 Token 后仅更新令牌
    setTokens(accessToken: string, refreshToken: string): void {
      this.token = accessToken
      this.refreshToken = refreshToken
      storage.set(TOKEN_KEY, accessToken, MAX_AGE)
      storage.set(REFRESH_TOKEN_KEY, refreshToken, MAX_AGE)
    },
    logout(): void {
      this.token = ''
      this.refreshToken = ''
      this.profile = null
      storage.remove(TOKEN_KEY)
      storage.remove(REFRESH_TOKEN_KEY)
      storage.remove(PROFILE_KEY)
    },
  },
})
