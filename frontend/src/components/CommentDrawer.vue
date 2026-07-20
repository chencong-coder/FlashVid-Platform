<script setup lang="ts">
import { ref } from 'vue'
import { Popup as VanPopup, showToast } from 'vant'

import { mockComments } from '@/data/mockVideos'
import type { CommentItem } from '@/types/video'
import { formatCount } from '@/utils/format'

interface Props {
  show: boolean
  total: number
}

interface Emits {
  (event: 'update:show', value: boolean): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

const content = ref('')
const comments = ref<CommentItem[]>([...mockComments])

const submit = (): void => {
  const value = content.value.trim()
  if (!value) return
  comments.value.unshift({
    id: `local-${Date.now()}`,
    userName: '我',
    avatar:
      'https://images.unsplash.com/photo-1534528741775-53994a69daeb?q=80&w=200&auto=format&fit=crop',
    content: value,
    time: '刚刚',
    likes: 0,
  })
  content.value = ''
  showToast('评论已发布')
}
</script>

<template>
  <van-popup
    :show="show"
    position="bottom"
    round
    class="h-[72dvh] overflow-hidden bg-panel text-white"
    @update:show="emit('update:show', $event)"
  >
    <section class="flex h-full flex-col bg-panel">
      <header
        class="relative flex h-12 shrink-0 items-center justify-center border-b border-white/5 text-sm font-semibold"
      >
        {{ formatCount(total) }} 条评论
        <button
          type="button"
          aria-label="关闭"
          class="absolute right-4 text-lg text-neutral-400"
          @click="emit('update:show', false)"
        >
          <i class="fa-solid fa-xmark" />
        </button>
      </header>
      <div class="no-scrollbar flex-1 overflow-y-auto px-4 py-2">
        <article v-for="comment in comments" :key="comment.id" class="flex gap-3 py-3">
          <img
            :src="comment.avatar"
            :alt="comment.userName"
            loading="lazy"
            class="h-9 w-9 shrink-0 rounded-full object-cover"
          />
          <div class="min-w-0 flex-1">
            <div class="text-xs text-neutral-500">{{ comment.userName }}</div>
            <p class="mt-1 text-sm leading-5 text-neutral-100">{{ comment.content }}</p>
            <div class="mt-1 text-[11px] text-neutral-500">{{ comment.time }}</div>
          </div>
          <button type="button" class="flex w-8 flex-col items-center gap-1 text-neutral-500">
            <i class="fa-regular fa-heart" />
            <span class="text-[10px]">{{ comment.likes || '' }}</span>
          </button>
        </article>
      </div>
      <footer
        class="safe-bottom flex shrink-0 items-center gap-3 border-t border-white/5 bg-panel px-4 pb-3 pt-3"
      >
        <input
          v-model="content"
          type="text"
          placeholder="留下你的精彩评论"
          class="h-10 min-w-0 flex-1 rounded-full bg-white/10 px-4 text-sm text-white outline-none placeholder:text-neutral-500"
          @keyup.enter="submit"
        />
        <button
          type="button"
          class="text-sm font-semibold text-primary disabled:opacity-40"
          :disabled="!content.trim()"
          @click="submit"
        >
          发送
        </button>
      </footer>
    </section>
  </van-popup>
</template>
