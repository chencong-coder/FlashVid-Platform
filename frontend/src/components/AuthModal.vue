<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { login, register } from '@/api/auth'
import { useUserStore } from '@/store/user'
import { useAuthModalStore } from '@/store/authModal'

const router = useRouter()
const userStore = useUserStore()
const authModal = useAuthModalStore()

// ─── Login ────────────────────────────────────────
const loginAccount = ref('')
const loginPassword = ref('')
const showLoginPwd = ref(false)
const loginLoading = ref(false)

async function handleLogin(): Promise<void> {
  if (!loginAccount.value.trim() || !loginPassword.value) {
    showToast('请填写账号和密码')
    return
  }
  loginLoading.value = true
  try {
    const res = await login({ account: loginAccount.value.trim(), password: loginPassword.value })
    if (res.data.code === 0) {
      userStore.setSession(res.data.data)
      const redirect = authModal.pendingRedirect
      authModal.close()
      if (redirect) await router.push(redirect)
    } else {
      showToast(res.data.message || '登录失败')
    }
  } catch (error: unknown) {
    showToast(error instanceof Error ? error.message : '网络错误，请重试')
  } finally {
    loginLoading.value = false
  }
}

// ─── Register ─────────────────────────────────────
const regUsername = ref('')
const regPhone = ref('')
const regCode = ref('')
const regPassword = ref('')
const showRegPwd = ref(false)
const regLoading = ref(false)
const codeSent = ref(false)
const codeCooldown = ref(0)

function sendCode(): void {
  if (!/^\d{11}$/.test(regPhone.value)) {
    showToast('请输入11位手机号')
    return
  }
  codeSent.value = true
  codeCooldown.value = 60
  showToast('演示验证码：123456')
  const t = setInterval(() => {
    if (--codeCooldown.value <= 0) {
      clearInterval(t)
      codeSent.value = false
    }
  }, 1000)
}

async function handleRegister(): Promise<void> {
  if (!regUsername.value || !regPhone.value || !regCode.value || !regPassword.value) {
    showToast('请填写所有信息')
    return
  }
  if (!/^\d{6}$/.test(regCode.value)) {
    showToast('请输入6位数字验证码')
    return
  }
  regLoading.value = true
  try {
    const res = await register({
      username: regUsername.value,
      phone: regPhone.value,
      code: regCode.value,
      password: regPassword.value,
    })
    if (res.data.code === 0) {
      const lr = await login({ account: regPhone.value, password: regPassword.value })
      if (lr.data.code === 0) {
        userStore.setSession(lr.data.data)
        const redirect = authModal.pendingRedirect
        authModal.close()
        if (redirect) await router.push(redirect)
      }
    } else {
      showToast(res.data.message || '注册失败')
    }
  } catch (error: unknown) {
    showToast(error instanceof Error ? error.message : '网络错误，请重试')
  } finally {
    regLoading.value = false
  }
}

// 关闭时重置到引导页并清空表单
watch(
  () => authModal.visible,
  (v) => {
    if (!v) {
      authModal.view = 'prompt'
      loginAccount.value = ''
      loginPassword.value = ''
      showLoginPwd.value = false
      regUsername.value = ''
      regPhone.value = ''
      regCode.value = ''
      regPassword.value = ''
      showRegPwd.value = false
    }
  },
)
</script>

<template>
  <Teleport to="body">
    <Transition name="auth-fade">
      <!-- 遮罩层 + 卡片容器 -->
      <div
        v-if="authModal.visible"
        class="fixed inset-0 z-[4000] flex items-center justify-center px-4"
      >
        <!-- 半透明遮罩：独立绝对层，点击关闭 -->
        <div
          class="absolute inset-0 bg-black/60 backdrop-blur-sm"
          @click="authModal.close()"
        />
        <!-- 居中卡片：z-10 确保在遮罩上方，@click.stop 防止事件冒泡到遮罩 -->
        <div
          class="relative z-10 w-full max-w-[22rem] rounded-2xl bg-[#1c1c1e] px-8 py-8 text-white shadow-2xl"
          @click.stop
        >
          <!-- 关闭按钮 -->
          <button
            type="button"
            aria-label="关闭登录"
            class="absolute right-3 top-3 flex h-8 w-8 items-center justify-center rounded-full text-neutral-400 transition-colors hover:bg-white/10 hover:text-white"
            @click="authModal.close()"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <!-- ── 引导页 prompt ───────────────────────────── -->
          <div
            v-if="authModal.view === 'prompt'"
            class="flex w-full flex-col items-center"
          >
          <div
            class="flex h-20 w-20 items-center justify-center rounded-3xl bg-primary shadow-xl shadow-primary/40"
          >
            <i class="fa-solid fa-play ml-1 text-3xl text-white" />
          </div>
          <h2 class="mt-4 text-2xl font-bold tracking-wide">闪视</h2>
          <p class="mt-2 text-center text-sm leading-relaxed text-neutral-400">
            登录后发现更多精彩内容<br />记录和分享你的精彩瞬间
          </p>
          <button
            class="mt-8 flex h-12 w-full items-center justify-center rounded-full bg-primary text-sm font-semibold shadow-lg shadow-primary/30 transition-transform active:scale-[0.98]"
            @click="authModal.switchView('login')"
          >
            立即登录
          </button>
          <button
            class="mt-3 flex h-12 w-full items-center justify-center rounded-full border border-white/15 text-sm text-neutral-300 transition-colors active:bg-white/5"
            @click="authModal.switchView('register')"
          >
            新用户注册
          </button>
          <p class="mt-6 text-xs text-neutral-600">登录即表示同意《用户协议》与《隐私政策》</p>
        </div>

          <!-- ── 登录表单 login ──────────────────────────── -->
          <div v-else-if="authModal.view === 'login'" class="w-full">
          <div class="mb-6 flex items-center gap-2">
            <button
              class="flex h-8 w-8 items-center justify-center rounded-full text-neutral-400 transition-colors hover:bg-white/10"
              @click="authModal.close()"
            >
              <i class="fa-solid fa-arrow-left text-sm" />
            </button>
            <span class="text-lg font-semibold">登录</span>
          </div>
          <div class="field-row">
            <i class="fa-regular fa-user field-icon" />
            <input
              v-model="loginAccount"
              type="text"
              placeholder="手机号 / 账号"
              class="field-input"
              autocomplete="username"
            />
          </div>
          <div class="field-row mt-2">
            <i class="fa-solid fa-lock field-icon" />
            <input
              v-model="loginPassword"
              :type="showLoginPwd ? 'text' : 'password'"
              placeholder="密码"
              class="field-input"
              autocomplete="current-password"
              @keyup.enter="handleLogin"
            />
            <button
              class="text-neutral-500 transition-colors hover:text-white"
              @click="showLoginPwd = !showLoginPwd"
            >
              <i
                :class="showLoginPwd ? 'fa-regular fa-eye' : 'fa-regular fa-eye-slash'"
                class="text-sm"
              />
            </button>
          </div>
          <button
            :disabled="loginLoading"
            class="mt-8 flex h-12 w-full items-center justify-center rounded-full bg-primary text-sm font-semibold shadow-lg shadow-primary/30 transition-all active:scale-[0.98] disabled:opacity-60"
            @click="handleLogin"
          >
            {{ loginLoading ? '登录中…' : '登录' }}
          </button>
          <p class="mt-4 text-center text-xs text-neutral-500">
            还没有账号？
            <button class="text-primary hover:underline" @click="authModal.switchView('register')">
              去注册
            </button>
          </p>
        </div>

          <!-- ── 注册表单 register ───────────────────────── -->
          <div v-else class="w-full">
          <div class="mb-6 flex items-center gap-2">
            <button
              class="flex h-8 w-8 items-center justify-center rounded-full text-neutral-400 transition-colors hover:bg-white/10"
              @click="authModal.close()"
            >
              <i class="fa-solid fa-arrow-left text-sm" />
            </button>
            <span class="text-lg font-semibold">注册</span>
          </div>
          <div class="field-row">
            <i class="fa-regular fa-user field-icon" />
            <input v-model="regUsername" type="text" placeholder="昵称" class="field-input" />
          </div>
          <div class="field-row mt-2">
            <i class="fa-solid fa-mobile-screen field-icon" />
            <input v-model="regPhone" type="tel" placeholder="手机号" class="field-input" />
            <button
              :disabled="codeSent"
              class="shrink-0 rounded-full border border-primary/60 px-3 py-1 text-xs text-primary transition-opacity disabled:opacity-50"
              @click="sendCode"
            >
              {{ codeSent ? `${codeCooldown}s` : '获取验证码' }}
            </button>
          </div>
          <div class="field-row mt-2">
            <i class="fa-solid fa-shield-halved field-icon" />
            <input
              v-model="regCode"
              type="text"
              inputmode="numeric"
              maxlength="6"
              placeholder="验证码"
              class="field-input"
            />
          </div>
          <div class="field-row mt-2">
            <i class="fa-solid fa-lock field-icon" />
            <input
              v-model="regPassword"
              :type="showRegPwd ? 'text' : 'password'"
              placeholder="设置密码"
              class="field-input"
            />
            <button
              class="text-neutral-500 transition-colors hover:text-white"
              @click="showRegPwd = !showRegPwd"
            >
              <i
                :class="showRegPwd ? 'fa-regular fa-eye' : 'fa-regular fa-eye-slash'"
                class="text-sm"
              />
            </button>
          </div>
          <button
            :disabled="regLoading"
            class="mt-8 flex h-12 w-full items-center justify-center rounded-full bg-primary text-sm font-semibold shadow-lg shadow-primary/30 transition-all active:scale-[0.98] disabled:opacity-60"
            @click="handleRegister"
          >
            {{ regLoading ? '注册中…' : '注册' }}
          </button>
          <p class="mt-4 text-center text-xs text-neutral-500">
            已有账号？
            <button class="text-primary hover:underline" @click="authModal.switchView('login')">
              去登录
            </button>
          </p>
        </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.field-row {
  @apply flex items-center gap-3 border-b border-white/10 py-3 transition-colors focus-within:border-primary/70;
}
.field-icon {
  @apply w-4 shrink-0 text-center text-sm text-neutral-500;
}
.field-input {
  @apply flex-1 bg-transparent text-sm text-white placeholder:text-neutral-500 outline-none;
}

/* autofill 覆盖 */
input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus,
input:-webkit-autofill:active {
  -webkit-box-shadow: 0 0 0 100px #0d0d0d inset !important;
  -webkit-text-fill-color: #ffffff !important;
  caret-color: #ffffff;
}

/* 遮罩过渡 */
.auth-fade-enter-active,
.auth-fade-leave-active {
  transition: opacity 0.25s ease;
}
.auth-fade-enter-from,
.auth-fade-leave-to {
  opacity: 0;
}

/* 弹窗缩放 */
.auth-zoom-enter-active,
.auth-zoom-leave-active {
  transition: all 0.28s cubic-bezier(0.4, 0, 0.2, 1);
}
.auth-zoom-enter-from,
.auth-zoom-leave-to {
  opacity: 0;
  transform: scale(0.92);
}
</style>
