<script setup lang="ts">
interface MessageItem {
  name: string
  preview: string
  time: string
  avatar: string
  unread: number
}

const shortcuts = [
  { label: '粉丝', icon: 'fa-user-plus', color: 'bg-primary' },
  { label: '赞和收藏', icon: 'fa-heart', color: 'bg-amber-500' },
  { label: '@我的', icon: 'fa-at', color: 'bg-cyan-500' },
  { label: '评论', icon: 'fa-comment-dots', color: 'bg-emerald-500' },
]

const messages: MessageItem[] = [
  {
    name: '闪视小助手',
    preview: '欢迎来到闪视，记录此刻的精彩。',
    time: '14:20',
    avatar:
      'https://images.unsplash.com/photo-1611162617474-5b21e879e113?q=80&w=200&auto=format&fit=crop',
    unread: 1,
  },
  {
    name: '户外旅行家',
    preview: '分享给你一个视频',
    time: '昨天',
    avatar:
      'https://images.unsplash.com/photo-1544005313-94ddf0286df2?q=80&w=200&auto=format&fit=crop',
    unread: 2,
  },
  {
    name: '可乐要加冰',
    preview: '周末一起去看展吗？',
    time: '星期三',
    avatar:
      'https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=200&auto=format&fit=crop',
    unread: 0,
  },
]
</script>

<template>
  <main class="safe-top h-full bg-[#0d0d0d] text-white">
    <header
      class="flex h-14 items-center justify-center border-b border-white/5 text-base font-semibold"
    >
      消息
    </header>
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
    <section>
      <article
        v-for="message in messages"
        :key="message.name"
        class="flex items-center gap-3 px-4 py-3"
      >
        <div class="relative shrink-0">
          <img
            :src="message.avatar"
            :alt="message.name"
            class="h-12 w-12 rounded-full object-cover"
          /><span
            v-if="message.unread"
            class="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[10px]"
            >{{ message.unread }}</span
          >
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex justify-between">
            <h2 class="text-sm font-medium">{{ message.name }}</h2>
            <time class="text-[10px] text-neutral-600">{{ message.time }}</time>
          </div>
          <p class="mt-1 truncate text-xs text-neutral-500">{{ message.preview }}</p>
        </div>
      </article>
    </section>
  </main>
</template>
