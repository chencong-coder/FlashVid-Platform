<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { BottomTab } from '@/types/app'

interface NavItem {
  label: string
  icon: string
  name: string
  tab: BottomTab
}

const route = useRoute()
const router = useRouter()

const items: NavItem[] = [
  { label: '首页', icon: 'fa-house', name: 'recommend', tab: 'home' },
  { label: '发现', icon: 'fa-compass', name: 'discover', tab: 'discover' },
  { label: '消息', icon: 'fa-message', name: 'messages', tab: 'messages' },
  { label: '我的', icon: 'fa-user', name: 'profile', tab: 'profile' },
]

const activeTab = computed<BottomTab>(() => route.meta.bottomTab ?? 'home')

const navigate = async (name: string): Promise<void> => {
  await router.push({ name })
}
</script>

<template>
  <nav
    class="safe-bottom z-50 flex h-[calc(4rem+env(safe-area-inset-bottom))] shrink-0 items-start justify-around bg-[#121212] px-1 pt-2 text-[11px]"
  >
    <button
      v-for="item in items.slice(0, 2)"
      :key="item.name"
      type="button"
      class="flex h-12 min-w-14 flex-col items-center justify-center gap-1"
      :class="activeTab === item.tab ? 'text-white' : 'text-neutral-500'"
      @click="navigate(item.name)"
    >
      <i class="fa-solid text-lg" :class="item.icon" />
      <span>{{ item.label }}</span>
    </button>

    <button
      type="button"
      aria-label="发布视频"
      class="mt-1 flex h-9 w-14 items-center justify-center rounded-md bg-primary text-xl text-white shadow-lg active:scale-95"
      @click="navigate('publish')"
    >
      <i class="fa-solid fa-plus" />
    </button>

    <button
      v-for="item in items.slice(2)"
      :key="item.name"
      type="button"
      class="flex h-12 min-w-14 flex-col items-center justify-center gap-1"
      :class="activeTab === item.tab ? 'text-white' : 'text-neutral-500'"
      @click="navigate(item.name)"
    >
      <i class="fa-solid text-lg" :class="item.icon" />
      <span>{{ item.label }}</span>
    </button>
  </nav>
</template>
