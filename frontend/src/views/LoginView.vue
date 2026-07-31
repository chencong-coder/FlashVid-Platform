<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import { login } from '@/api/auth'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()

const account = ref('')
const password = ref('')
const loading = ref(false)
const showPwd = ref(false)

const handleLogin = async (): Promise<void> => {
  if (!account.value.trim() || !password.value.trim()) {
    showToast('请填写账号和密码')
    return
  }
  loading.value = true
  try {
    const res = await login({ account: account.value.trim(), password: password.value })
    if (res.data.code === 0) {
      userStore.setSession(res.data.data)
      const redirect = (router.currentRoute.value.query.redirect as string) || '/'
      await router.replace(redirect)
    } else {
      showToast(res.data.message || '账号或密码错误')
    }
  } catch (error: unknown) {
    showToast(error instanceof Error ? error.message : '网络错误，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="flex h-full flex-col bg-[#0d0d0d] text-white">
    <!-- 顶部返回 -->
    <header class="safe-top flex h-14 shrink-0 items-center px-4">
      <button
        type="button"
        aria-label="返回"
        class="flex h-9 w-9 items-center justify-center rounded-full text-neutral-400 active:bg-white/10"
        @click="router.back()"
      >
        <i class="fa-solid fa-chevron-left text-lg" />
      </button>
    </header>

    <!-- Logo 区域 -->
    <section class="flex flex-col items-center px-8 pt-6 pb-10">
      <div
        class="flex h-20 w-20 items-center justify-center rounded-3xl bg-primary shadow-lg shadow-primary/30"
      >
        <i class="fa-solid fa-play text-3xl text-white" />
      </div>
      <h1 class="mt-5 text-2xl font-bold tracking-wide">闪视</h1>
      <p class="mt-2 text-sm text-neutral-400">记录精彩瞬间，发现更多可能</p>
    </section>

    <!-- 表单区域 -->
    <section class="flex-1 px-8">
      <!-- 账号 -->
      <div
        class="group flex items-center gap-3 border-b border-white/10 py-3 focus-within:border-primary/70 transition-colors"
      >
        <i
          class="fa-regular fa-user w-4 text-center text-neutral-500 group-focus-within:text-primary transition-colors"
        />
        <input
          v-model="account"
          type="text"
          autocomplete="username"
          placeholder="用户名 / 手机号"
          class="flex-1 bg-transparent text-sm text-white outline-none placeholder:text-neutral-600"
        />
      </div>
      <!-- 密码 -->
      <div
        class="group flex items-center gap-3 border-b border-white/10 py-3 mt-4 focus-within:border-primary/70 transition-colors"
      >
        <i
          class="fa-solid fa-lock w-4 text-center text-neutral-500 group-focus-within:text-primary transition-colors"
        />
        <input
          v-model="password"
          :type="showPwd ? 'text' : 'password'"
          autocomplete="current-password"
          placeholder="密码"
          class="flex-1 bg-transparent text-sm text-white outline-none placeholder:text-neutral-600"
          @keyup.enter="handleLogin"
        />
        <button
          type="button"
          class="text-neutral-600 active:text-white"
          @click="showPwd = !showPwd"
        >
          <i :class="showPwd ? 'fa-regular fa-eye' : 'fa-regular fa-eye-slash'" />
        </button>
      </div>

      <!-- 登录按钮 -->
      <button
        type="button"
        :disabled="loading"
        class="mt-10 flex h-12 w-full items-center justify-center rounded-full bg-primary text-sm font-semibold shadow-lg shadow-primary/30 disabled:opacity-50 active:scale-[0.98] transition-transform"
        @click="handleLogin"
      >
        <span v-if="loading" class="flex items-center gap-2">
          <i class="fa-solid fa-circle-notch animate-spin" />登录中...
        </span>
        <span v-else>登录</span>
      </button>

      <!-- 分割线 -->
      <div class="mt-8 flex items-center gap-4 text-xs text-neutral-600">
        <div class="flex-1 border-t border-white/5" />
        还没有账号？
        <div class="flex-1 border-t border-white/5" />
      </div>

      <router-link
        to="/register"
        class="mt-4 flex h-12 w-full items-center justify-center rounded-full border border-white/15 text-sm text-neutral-300 active:bg-white/5 transition-colors"
      >
        新用户注册
      </router-link>
    </section>

    <p class="pb-10 text-center text-xs text-neutral-600">
      登录即表示同意《用户协议》与《隐私政策》
    </p>
  </main>
</template>

<style scoped>
/* 覆盖浏览器 autofill 白底 */
input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus,
input:-webkit-autofill:active {
  -webkit-box-shadow: 0 0 0 100px #0d0d0d inset !important;
  -webkit-text-fill-color: #ffffff !important;
  caret-color: #ffffff;
}
</style>
