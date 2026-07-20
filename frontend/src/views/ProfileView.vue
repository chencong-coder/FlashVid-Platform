<script setup lang="ts">
import { computed, ref } from 'vue'

import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const activeTab = ref<'works' | 'likes'>('works')
const profile = computed(
  () =>
    userStore.profile ?? {
      id: 'guest',
      nickname: '登录后发现更多精彩',
      avatar:
        'https://images.unsplash.com/photo-1534528741775-53994a69daeb?q=80&w=240&auto=format&fit=crop',
      bio: '点击编辑个人简介',
      following: 128,
      followers: 32600,
      likes: 89000,
    },
)

const works = [
  'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?q=80&w=500&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?q=80&w=500&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1492684223066-81342ee5ff30?q=80&w=500&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1514525253161-7a46d19cd819?q=80&w=500&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1536440136628-849c177e76a1?q=80&w=500&auto=format&fit=crop',
  'https://images.unsplash.com/photo-1578662996442-48f60103fc96?q=80&w=500&auto=format&fit=crop',
]
</script>

<template>
  <main class="no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <div class="relative h-44 bg-neutral-900">
      <img
        src="https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?q=80&w=1080&auto=format&fit=crop"
        alt="主页背景"
        class="h-full w-full object-cover opacity-60"
      />
      <div class="absolute inset-0 bg-gradient-to-b from-black/20 to-[#0d0d0d]" />
      <div class="safe-top absolute inset-x-0 top-0 flex justify-end gap-5 px-4 pt-4 text-lg">
        <button aria-label="添加朋友"><i class="fa-solid fa-user-plus" /></button
        ><button aria-label="更多"><i class="fa-solid fa-ellipsis" /></button>
      </div>
    </div>
    <section class="relative -mt-11 px-4">
      <img
        :src="profile.avatar"
        :alt="profile.nickname"
        class="h-[5.5rem] w-[5.5rem] rounded-full border-4 border-[#0d0d0d] object-cover"
      />
      <h1 class="mt-3 text-xl font-bold">{{ profile.nickname }}</h1>
      <p class="mt-1 text-xs text-neutral-500">闪视号：flashvid_2026</p>
      <p class="mt-3 text-sm text-neutral-300">{{ profile.bio }}</p>
      <div class="mt-4 flex gap-6 text-sm">
        <span
          ><b>{{ profile.following }}</b> <i class="not-italic text-neutral-500">关注</i></span
        ><span><b>3.2万</b> <i class="not-italic text-neutral-500">粉丝</i></span
        ><span><b>8.9万</b> <i class="not-italic text-neutral-500">获赞</i></span>
      </div>
      <div class="mt-4 flex gap-2">
        <button class="h-9 flex-1 rounded bg-white/10 text-sm font-medium">编辑资料</button
        ><button aria-label="分享主页" class="h-9 w-11 rounded bg-white/10">
          <i class="fa-solid fa-share" />
        </button>
      </div>
    </section>
    <section class="mt-5">
      <div class="grid h-12 grid-cols-2 border-b border-white/5 text-sm">
        <button
          type="button"
          class="relative"
          :class="activeTab === 'works' ? 'text-white' : 'text-neutral-500'"
          @click="activeTab = 'works'"
        >
          作品<span
            v-if="activeTab === 'works'"
            class="absolute bottom-0 left-1/2 h-0.5 w-6 -translate-x-1/2 bg-white"
          /></button
        ><button
          type="button"
          class="relative"
          :class="activeTab === 'likes' ? 'text-white' : 'text-neutral-500'"
          @click="activeTab = 'likes'"
        >
          喜欢<span
            v-if="activeTab === 'likes'"
            class="absolute bottom-0 left-1/2 h-0.5 w-6 -translate-x-1/2 bg-white"
          />
        </button>
      </div>
      <div class="grid grid-cols-3 gap-0.5">
        <div
          v-for="(work, index) in works"
          :key="work"
          class="relative aspect-[3/4] overflow-hidden bg-neutral-900"
        >
          <img :src="work" alt="视频作品" loading="lazy" class="h-full w-full object-cover" /><span
            class="absolute bottom-1 left-1 text-[10px] text-white"
            ><i class="fa-solid fa-play mr-1" />{{ 12 + index * 7 }}.6万</span
          >
        </div>
      </div>
    </section>
  </main>
</template>
