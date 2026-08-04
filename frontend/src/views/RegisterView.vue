<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import { register, login } from '@/api/auth'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()

const username = ref('')
const phone = ref('')
const password = ref('')
const showPwd = ref(false)
const code = ref('')
const loading = ref(false)
const codeSent = ref(false)
const codeCooldown = ref(0)
let cooldownTimer: ReturnType<typeof setInterval> | null = null

const sendCode = (): void => {
  if (!phone.value || !/^1[3-9]\d{9}$/.test(phone.value)) {
    showToast('请输入正确的手机号')
    return
  }
  codeSent.value = true
  codeCooldown.value = 60
  showToast('验证码已发送（演示：填 123456）')
  cooldownTimer = setInterval(() => {
    codeCooldown.value--
    if (codeCooldown.value <= 0) {
      clearInterval(cooldownTimer!)
      codeSent.value = false
    }
  }, 1000)
}

const handleRegister = async (): Promise<void> => {
  if (!username.value.trim() || !phone.value || !password.value || !code.value) {
    showToast('请填写所有必填项')
    return
  }
  if (username.value.length < 4 || username.value.length > 32) {
    showToast('用户名须 4-32 个字符')
    return
  }
  if (password.value.length < 6 || password.value.length > 20) {
    showToast('密码须 6-20 个字符')
    return
  }
  if (!/^\d{6}$/.test(code.value)) {
    showToast('请输入6位数字验证码')
    return
  }
  loading.value = true
  try {
    const res = await register({
      username: username.value.trim(),
      password: password.value,
      phone: phone.value,
      code: code.value,
    })
    if (res.data.code === 0) {
      const loginRes = await login({ account: username.value.trim(), password: password.value })
      if (loginRes.data.code === 0) userStore.setSession(loginRes.data.data)
      showToast('注册成功，欢迎加入闪视！')
      await router.replace('/')
    } else {
      showToast(res.data.message || '注册失败')
    }
  } catch (error: unknown) {
    showToast(error instanceof Error ? error.message : '网络错误，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="no-scrollbar flex h-full flex-col overflow-y-auto bg-[#0d0d0d] text-white">
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

    <!-- Logo -->
    <section class="flex flex-col items-center px-8 pt-2 pb-8">
      <div
        class="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary shadow-lg shadow-primary/30"
      >
        <i class="fa-solid fa-play text-2xl text-white" />
      </div>
      <h1 class="mt-4 text-2xl font-bold">加入闪视</h1>
      <p class="mt-1 text-sm text-neutral-400">开始记录你的精彩瞬间</p>
    </section>

    <!-- 表单 -->
    <section class="flex-1 px-8 pb-10 space-y-0">
      <!-- 用户名 -->
      <div
        class="group flex items-center gap-3 border-b border-white/10 py-3 focus-within:border-primary/70 transition-colors"
      >
        <i
          class="fa-regular fa-user w-4 text-center text-neutral-500 group-focus-within:text-primary transition-colors"
        />
        <input
          v-model="username"
          type="text"
          autocomplete="username"
          placeholder="用户名（4-32 字符）"
          class="flex-1 bg-transparent text-sm outline-none placeholder:text-neutral-600"
        />
      </div>

      <!-- 手机号 -->
      <div
        class="group flex items-center gap-3 border-b border-white/10 py-3 focus-within:border-primary/70 transition-colors"
      >
        <i
          class="fa-solid fa-mobile-screen w-4 text-center text-neutral-500 group-focus-within:text-primary transition-colors"
        />
        <input
          v-model="phone"
          type="tel"
          autocomplete="tel"
          placeholder="手机号"
          maxlength="11"
          class="flex-1 bg-transparent text-sm outline-none placeholder:text-neutral-600"
        />
      </div>

      <!-- 验证码 -->
      <div
        class="group flex items-center gap-3 border-b border-white/10 py-3 focus-within:border-primary/70 transition-colors"
      >
        <i
          class="fa-solid fa-shield-halved w-4 text-center text-neutral-500 group-focus-within:text-primary transition-colors"
        />
        <input
          v-model="code"
          type="text"
          inputmode="numeric"
          placeholder="验证码"
          maxlength="6"
          class="flex-1 bg-transparent text-sm outline-none placeholder:text-neutral-600"
        />
        <button
          type="button"
          :disabled="codeSent"
          class="shrink-0 text-xs font-medium text-primary disabled:text-neutral-500 disabled:cursor-not-allowed"
          @click="sendCode"
        >
          {{ codeSent ? `${codeCooldown}s` : '发送验证码' }}
        </button>
      </div>

      <!-- 密码 -->
      <div
        class="group flex items-center gap-3 border-b border-white/10 py-3 focus-within:border-primary/70 transition-colors"
      >
        <i
          class="fa-solid fa-lock w-4 text-center text-neutral-500 group-focus-within:text-primary transition-colors"
        />
        <input
          v-model="password"
          :type="showPwd ? 'text' : 'password'"
          autocomplete="new-password"
          placeholder="密码（6-20 字符）"
          class="flex-1 bg-transparent text-sm outline-none placeholder:text-neutral-600"
          @keyup.enter="handleRegister"
        />
        <button
          type="button"
          class="text-neutral-600 active:text-white"
          @click="showPwd = !showPwd"
        >
          <i :class="showPwd ? 'fa-regular fa-eye' : 'fa-regular fa-eye-slash'" />
        </button>
      </div>

      <button
        type="button"
        :disabled="loading"
        class="mt-10 flex h-12 w-full items-center justify-center rounded-full bg-primary text-sm font-semibold shadow-lg shadow-primary/30 disabled:opacity-50 active:scale-[0.98] transition-transform"
        @click="handleRegister"
      >
        <span v-if="loading" class="flex items-center gap-2">
          <i class="fa-solid fa-circle-notch animate-spin" />注册中...
        </span>
        <span v-else>注册</span>
      </button>

      <div class="mt-8 flex items-center gap-4 text-xs text-neutral-600">
        <div class="flex-1 border-t border-white/5" />
        已有账号？
        <div class="flex-1 border-t border-white/5" />
      </div>

      <router-link
        to="/login"
        class="mt-4 flex h-12 w-full items-center justify-center rounded-full border border-white/15 text-sm text-neutral-300 active:bg-white/5 transition-colors"
      >
        立即登录
      </router-link>
    </section>

    <p class="pb-8 text-center text-xs text-neutral-600">
      注册即表示同意《用户协议》与《隐私政策》
    </p>
  </main>
</template>

<style scoped>
input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus,
input:-webkit-autofill:active {
  -webkit-box-shadow: 0 0 0 100px #0d0d0d inset !important;
  -webkit-text-fill-color: #ffffff !important;
  caret-color: #ffffff;
}
</style>
