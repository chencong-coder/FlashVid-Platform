<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { getConversations, type ConversationInfo } from '@/api/message'
import { parseApiDate } from '@/utils/date'
import { useUserStore } from '@/store/user'
import { useAuthModalStore } from '@/store/authModal'

const userStore = useUserStore()
const authModal = useAuthModalStore()

const router = useRouter()

const shortcuts = [
  { label: '粉丝', icon: 'fa-user-plus', color: 'bg-primary' },
  { label: '赞和收藏', icon: 'fa-heart', color: 'bg-amber-500' },
  { label: '@我的', icon: 'fa-at', color: 'bg-cyan-500' },
  { label: '评论', icon: 'fa-comment-dots', color: 'bg-emerald-500' },
]

const conversations = ref<ConversationInfo[]>([])
const loading = ref(false)
const error = ref('')

async function loadConversations() {
  loading.value = true
  error.value = ''
  try {
    const res = await getConversations({ page: 1, pageSize: 30 })
    conversations.value = res.data.data.list ?? []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

/** 将后端时间字符串格式化为会话列表时间显示 */
function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = parseApiDate(dateStr)
  if (!d) return dateStr
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffDays = Math.floor(diffMs / 86_400_000)
  if (diffDays === 0)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  if (diffDays === 1) return '昨天'
  if (diffDays < 7)
    return ['日', '一', '二', '三', '四', '五', '六'][d.getDay()]
      ? `星期${['日', '一', '二', '三', '四', '五', '六'][d.getDay()]}`
      : '昨天'
  return `${d.getMonth() + 1}/${d.getDate()}`
}

/** 根据消息类型生成预览文本 */
function previewText(conv: ConversationInfo): string {
  const msg = conv.lastMessage
  if (!msg || !msg.id) return ''
  if (msg.messageType === 2) return '[图片]'
  if (msg.messageType === 3) return '[视频]'
  return msg.content
}

function openChat(conv: ConversationInfo) {
  void router.push({ name: 'chat', params: { userId: conv.targetUser.id } })
}

onMounted(loadConversations)
</script>

<template>
  <!-- 未登录：引导登录 -->
  <main
    v-if="!userStore.isLoggedIn"
    class="safe-top flex h-full flex-col items-center justify-center bg-[#0d0d0d] px-8 text-white"
  >
    <div class="flex w-full max-w-[25rem] flex-col items-center text-center">
      <div
        class="flex h-[4.75rem] w-[4.75rem] items-center justify-center rounded-2xl border border-white/15 bg-primary shadow-[0_18px_45px_rgba(254,44,85,0.3)]"
      >
        <i class="fa-solid fa-play ml-1 text-[1.75rem] text-white" />
      </div>
      <h1 class="mt-5 text-[1.75rem] font-bold">闪视</h1>
      <p class="mt-2 text-sm leading-6 text-neutral-400">登录后查看消息与私信</p>
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
    </div>
  </main>

  <!-- 已登录：正常消息页面 -->
  <main v-else class="safe-top h-full bg-[#0d0d0d] text-white">
    <header
      class="flex h-14 items-center justify-center border-b border-white/5 text-base font-semibold"
    >
      消息
    </header>

    <!-- 快捷入口 -->
    <section class="grid grid-cols-4 border-b border-white/5 px-3 py-5">
      <button
        v-for="item in shortcuts"
        :key="item.label"
        type="button"
        class="flex flex-col items-center gap-2 text-xs"
      >
        <span
          class="flex h-11 w-11 items-center justify-center rounded-full text-lg text-white"
          :class="item.color"
          ><i class="fa-solid" :class="item.icon"
        /></span>
        {{ item.label }}
      </button>
    </section>

    <!-- 加载中 -->
    <div v-if="loading" class="flex justify-center py-10">
      <i class="fa-solid fa-circle-notch animate-spin text-2xl text-neutral-500" />
    </div>

    <!-- 错误提示 -->
    <div v-else-if="error" class="px-4 py-6 text-center text-sm text-red-400">
      {{ error }}
      <button type="button" class="mt-2 block w-full text-primary" @click="loadConversations">
        重试
      </button>
    </div>

    <!-- 空状态 -->
    <div
      v-else-if="conversations.length === 0"
      class="px-4 py-12 text-center text-sm text-neutral-500"
    >
      暂无私信
    </div>

    <!-- 会话列表 -->
    <section v-else>
      <article
        v-for="conv in conversations"
        :key="conv.targetUser.id"
        class="flex cursor-pointer items-center gap-3 px-4 py-3 active:bg-white/5"
        @click="openChat(conv)"
      >
        <div class="relative shrink-0">
          <img
            :src="conv.targetUser.avatar"
            :alt="conv.targetUser.nickname || conv.targetUser.username"
            class="h-12 w-12 rounded-full object-cover"
          />
          <span
            v-if="conv.unreadCount > 0"
            class="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[10px]"
            >{{ conv.unreadCount > 99 ? '99+' : conv.unreadCount }}</span
          >
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex justify-between">
            <h2 class="text-sm font-medium">
              {{ conv.targetUser.nickname || conv.targetUser.username }}
            </h2>
            <time class="text-[10px] text-neutral-600">{{ formatTime(conv.updatedAt) }}</time>
          </div>
          <p class="mt-1 truncate text-xs text-neutral-500">{{ previewText(conv) }}</p>
        </div>
      </article>
    </section>
  </main>
</template>
