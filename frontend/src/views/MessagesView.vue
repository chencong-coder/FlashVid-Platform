<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import LoginGate from '@/components/LoginGate.vue'
import { getConversations, type ConversationInfo } from '@/api/message'
import { parseApiDate } from '@/utils/date'
import { useUserStore } from '@/store/user'
import { useNotificationStore } from '@/store/notification'

const userStore = useUserStore()
const notifStore = useNotificationStore()
const router = useRouter()

const shortcuts = computed(() => [
  {
    label: '粉丝',
    icon: 'fa-user-plus',
    color: 'bg-primary',
    count: notifStore.unreadCounts.followers,
    route: '/notifications/followers',
  },
  {
    label: '赞和收藏',
    icon: 'fa-heart',
    color: 'bg-amber-500',
    count: notifStore.unreadCounts.likesAndFavs,
    route: '/notifications/likes-favs',
  },
  {
    label: '@我的',
    icon: 'fa-at',
    color: 'bg-cyan-500',
    count: notifStore.unreadCounts.mentions,
    route: '/notifications/mentions',
  },
  {
    label: '评论',
    icon: 'fa-comment-dots',
    color: 'bg-emerald-500',
    count: notifStore.unreadCounts.comments,
    route: '/notifications/comments',
  },
])

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

onMounted(() => {
  if (userStore.isLoggedIn) {
    void loadConversations()
    void notifStore.fetchUnreadCounts()
  }
})
</script>

<template>
  <LoginGate>
    <main class="safe-top h-full bg-[#0d0d0d] text-white">
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
        @click="router.push(item.route)"
      >
        <span class="relative">
          <span
            class="flex h-11 w-11 items-center justify-center rounded-full text-lg text-white"
            :class="item.color"
            ><i class="fa-solid" :class="item.icon"
          /></span>
          <span
            v-if="item.count > 0"
            class="absolute -right-1 -top-1 flex min-w-[18px] items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-medium leading-[18px] text-white"
          >{{ item.count > 99 ? '+99+' : `+${item.count}` }}</span>
        </span>
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
  </LoginGate>
</template>
