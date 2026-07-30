<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getMessages, sendMessage, markConversationRead, type MessageInfo } from '@/api/message'
import { getUserInfo, type UserInfo } from '@/api/user'
import { useUserStore } from '@/store/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const targetUserId = computed(() => Number(route.params.userId))
const currentUserId = computed(() => Number(userStore.profile?.id ?? 0))

const messages = ref<MessageInfo[]>([])
const targetUser = ref<UserInfo | null>(null)
const cursorToken = ref('')
const hasMore = ref(false)
const loadingOlder = ref(false)
const sendingMessage = ref(false)
const inputText = ref('')
const msgListEl = ref<HTMLElement | null>(null)

async function loadTargetUser() {
  try {
    const res = await getUserInfo(targetUserId.value)
    targetUser.value = res.data.data
  } catch {}
}

async function loadMessages(prepend = false) {
  loadingOlder.value = true
  try {
    const params: { cursor?: string; count: number } = { count: 20 }
    if (prepend && cursorToken.value) params.cursor = cursorToken.value
    const res = await getMessages(targetUserId.value, params)
    const data = res.data.data
    // 后端返回的是倒序（最新在前），翻转后最新在底部
    const ordered = [...data.messages].reverse()
    messages.value = prepend ? [...ordered, ...messages.value] : ordered
    cursorToken.value = data.nextCursorToken
    hasMore.value = data.hasMore
  } catch {
    // 静默失败，不影响其他交互
  } finally {
    loadingOlder.value = false
  }
}

async function markRead() {
  try {
    await markConversationRead(targetUserId.value)
  } catch {}
}

function scrollToBottom() {
  if (msgListEl.value) msgListEl.value.scrollTop = msgListEl.value.scrollHeight
}

async function handleSend() {
  const text = inputText.value.trim()
  if (!text || sendingMessage.value) return
  sendingMessage.value = true
  // 乐观更新：先展示，再确认
  const optimistic: MessageInfo = {
    id: Date.now(),
    fromUserId: currentUserId.value,
    toUserId: targetUserId.value,
    messageType: 1,
    content: text,
    mediaUrl: '',
    isRead: false,
    createdAt: new Date().toISOString(),
  }
  messages.value.push(optimistic)
  inputText.value = ''
  await nextTick()
  scrollToBottom()
  try {
    const res = await sendMessage({ toUserId: targetUserId.value, messageType: 1, content: text })
    const idx = messages.value.findIndex((m) => m.id === optimistic.id)
    if (idx !== -1) messages.value[idx] = res.data.data
  } catch {
    messages.value = messages.value.filter((m) => m.id !== optimistic.id)
    inputText.value = text
  } finally {
    sendingMessage.value = false
  }
}

function formatMsgTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return ''
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

onMounted(async () => {
  await Promise.all([loadTargetUser(), loadMessages()])
  markRead()
  await nextTick()
  scrollToBottom()
})
</script>

<template>
  <main class="safe-top flex h-full flex-col bg-[#0d0d0d] text-white">
    <!-- 顶栏 -->
    <header class="flex h-14 shrink-0 items-center gap-1 border-b border-white/5 px-3">
      <button type="button" class="p-2 text-xl" @click="router.back()">
        <i class="fa-solid fa-chevron-left" />
      </button>
      <span class="flex-1 text-center text-base font-semibold">
        {{ targetUser?.nickname || targetUser?.username || '私信' }}
      </span>
      <div class="w-10" />
    </header>

    <!-- 消息列表 -->
    <section ref="msgListEl" class="flex-1 space-y-3 overflow-y-auto px-3 py-4">
      <div class="flex justify-center">
        <button
          v-if="hasMore"
          type="button"
          :disabled="loadingOlder"
          class="text-xs text-neutral-500 disabled:opacity-50"
          @click="loadMessages(true)"
        >
          {{ loadingOlder ? '加载中…' : '查看更早消息' }}
        </button>
      </div>

      <div
        v-for="msg in messages"
        :key="msg.id"
        class="flex"
        :class="msg.fromUserId === currentUserId ? 'justify-end' : 'justify-start'"
      >
        <!-- 对方消息 -->
        <div v-if="msg.fromUserId !== currentUserId" class="flex max-w-[70%] items-end gap-2">
          <img
            v-if="targetUser?.avatar"
            :src="targetUser.avatar"
            :alt="targetUser.nickname"
            class="h-8 w-8 shrink-0 rounded-full object-cover"
          />
          <div class="rounded-2xl rounded-bl-sm bg-white/10 px-3 py-2 text-sm leading-relaxed">
            <span v-if="msg.messageType === 2">[图片]</span>
            <span v-else-if="msg.messageType === 3">[视频]</span>
            <span v-else>{{ msg.content }}</span>
          </div>
          <time class="shrink-0 text-[10px] text-neutral-600">{{ formatMsgTime(msg.createdAt) }}</time>
        </div>

        <!-- 自己的消息 -->
        <div v-else class="flex max-w-[70%] items-end gap-2">
          <time class="shrink-0 text-[10px] text-neutral-600">{{ formatMsgTime(msg.createdAt) }}</time>
          <div class="rounded-2xl rounded-br-sm bg-primary px-3 py-2 text-sm leading-relaxed">
            <span v-if="msg.messageType === 2">[图片]</span>
            <span v-else-if="msg.messageType === 3">[视频]</span>
            <span v-else>{{ msg.content }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 输入框 -->
    <footer class="shrink-0 border-t border-white/5 px-3 py-2">
      <div class="flex items-center gap-2">
        <input
          v-model="inputText"
          type="text"
          placeholder="发消息…"
          maxlength="1000"
          class="h-10 flex-1 rounded-full bg-white/10 px-4 text-sm text-white placeholder:text-neutral-500 focus:outline-none focus:ring-1 focus:ring-primary"
          @keydown.enter.prevent="handleSend"
        />
        <button
          type="button"
          :disabled="!inputText.trim() || sendingMessage"
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary text-white disabled:opacity-40"
          @click="handleSend"
        >
          <i class="fa-solid fa-paper-plane text-sm" />
        </button>
      </div>
    </footer>
  </main>
</template>
