<script setup lang="ts">
import { useUserStore } from '@/store/user'
import { useAuthModalStore } from '@/store/authModal'

const userStore = useUserStore()
const authModal = useAuthModalStore()
</script>

<template>
  <slot v-if="userStore.isLoggedIn" />
  <main
    v-else
    class="safe-top flex h-full flex-col items-center justify-center bg-[#0d0d0d] px-8 text-white"
  >
    <div class="flex w-full max-w-[25rem] flex-col items-center text-center">
      <div
        class="flex h-[4.75rem] w-[4.75rem] items-center justify-center rounded-2xl border border-white/15 bg-primary shadow-[0_18px_45px_rgba(254,44,85,0.3)]"
      >
        <i class="fa-solid fa-play ml-1 text-[1.75rem] text-white" />
      </div>
      <h1 class="mt-5 text-[1.75rem] font-bold text-white">闪视</h1>
      <p class="mt-2 text-sm leading-6 text-neutral-400">记录精彩瞬间，发现更多可能</p>

      <button
        type="button"
        class="mt-8 flex h-[3.25rem] w-full items-center justify-center gap-2 rounded-lg bg-primary text-sm font-semibold text-white shadow-[0_12px_28px_rgba(254,44,85,0.28)] transition-all hover:bg-[#ff3e63] active:scale-[0.98]"
        @click="authModal.open('login')"
      >
        <i class="fa-solid fa-arrow-right-to-bracket text-xs" />
        登录
      </button>

      <div class="my-4 flex w-full items-center gap-3 text-[11px] text-neutral-600">
        <span class="h-px flex-1 bg-white/10" />
        <span>还没有账号？</span>
        <span class="h-px flex-1 bg-white/10" />
      </div>

      <button
        type="button"
        class="flex h-[3.25rem] w-full items-center justify-center gap-2 rounded-lg border border-white/15 bg-white/[0.03] text-sm font-medium text-neutral-200 transition-all hover:border-white/25 hover:bg-white/[0.07] active:scale-[0.98]"
        @click="authModal.open('register')"
      >
        <i class="fa-solid fa-user-plus text-xs text-neutral-400" />
        新用户注册
      </button>

      <p
        class="mt-6 flex flex-wrap items-center justify-center gap-x-1 text-[11px] leading-5 text-neutral-600"
      >
        <span>登录即表示同意</span>
        <button type="button" class="transition-colors hover:text-neutral-400">《用户协议》</button>
        <span>与</span>
        <button type="button" class="transition-colors hover:text-neutral-400">《隐私政策》</button>
      </p>
    </div>
  </main>
</template>
