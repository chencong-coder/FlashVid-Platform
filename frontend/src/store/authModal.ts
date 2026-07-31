import { defineStore } from 'pinia'
import { ref } from 'vue'

import { useUserStore } from '@/store/user'

type AuthView = 'prompt' | 'login' | 'register'

export const useAuthModalStore = defineStore('authModal', () => {
  const visible = ref(false)
  const view = ref<AuthView>('prompt')
  const pendingRedirect = ref<string | null>(null)

  function open(initialView: AuthView = 'prompt', redirect?: string): void {
    view.value = initialView
    pendingRedirect.value = redirect ?? null
    visible.value = true
  }

  function switchView(newView: AuthView): void {
    view.value = newView
  }

  function close(): void {
    visible.value = false
    view.value = 'prompt'
    pendingRedirect.value = null
  }

  function requireLogin(redirect?: string): boolean {
    if (useUserStore().isLoggedIn) return true
    open('prompt', redirect)
    return false
  }

  return { visible, view, pendingRedirect, open, switchView, close, requireLogin }
})
