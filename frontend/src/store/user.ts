import { defineStore } from 'pinia'

import type { LoginResult, UserProfile } from '@/types/user'
import { storage } from '@/utils/storage'

interface UserState {
  token: string
  profile: UserProfile | null
}

const TOKEN_KEY = 'flashvid_token'
const PROFILE_KEY = 'flashvid_profile'

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: storage.get<string>(TOKEN_KEY) ?? '',
    profile: storage.get<UserProfile>(PROFILE_KEY),
  }),
  getters: {
    isLoggedIn: (state): boolean => Boolean(state.token),
  },
  actions: {
    setSession(result: LoginResult): void {
      this.token = result.token
      this.profile = result.user
      storage.set(TOKEN_KEY, result.token, 30 * 24 * 60 * 60 * 1000)
      storage.set(PROFILE_KEY, result.user, 30 * 24 * 60 * 60 * 1000)
    },
    logout(): void {
      this.token = ''
      this.profile = null
      storage.remove(TOKEN_KEY)
      storage.remove(PROFILE_KEY)
    },
  },
})
