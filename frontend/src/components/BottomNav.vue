<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { BottomTab } from '@/types/app'

interface NavItem {
  icon: string
  name: string
  tab: BottomTab
}

const route = useRoute()
const router = useRouter()

const leftItems: NavItem[] = [
  { icon: 'fa-house',   name: 'recommend', tab: 'home' },
  { icon: 'fa-compass', name: 'discover',  tab: 'discover' },
]
const rightItems: NavItem[] = [
  { icon: 'fa-message', name: 'messages', tab: 'messages' },
  { icon: 'fa-user',    name: 'profile',  tab: 'profile' },
]

const activeTab = computed<BottomTab>(() => route.meta.bottomTab ?? 'home')

const navigate = (name: string) => router.push({ name })
</script>

<template>
  <!-- Floating pill dock — always visible, floats above content -->
  <nav
    aria-label="底部导航"
    class="fixed bottom-5 left-1/2 -translate-x-1/2 z-50
           flex items-center gap-1 px-3 py-2
           rounded-2xl bg-[#1c1c24]/90 backdrop-blur-2xl
           border border-white/[0.08] shadow-2xl shadow-black/60"
  >
    <!-- Left two items -->
    <button
      v-for="item in leftItems"
      :key="item.name"
      type="button"
      class="w-11 h-10 flex items-center justify-center rounded-xl transition-all duration-200"
      :class="activeTab === item.tab
        ? 'text-white'
        : 'text-gray-500 hover:text-gray-300'"
      @click="navigate(item.name)"
    >
      <i class="fa-solid text-[20px]" :class="item.icon" />
    </button>

    <!-- Center publish button (prominent purple circle) -->
    <button
      type="button"
      aria-label="发布"
      class="mx-1 w-12 h-12 flex items-center justify-center rounded-full
             bg-violet-600 text-white text-xl
             shadow-lg shadow-violet-600/50
             transition-all duration-200 hover:bg-violet-500 hover:shadow-violet-500/60 hover:scale-105 active:scale-90"
      @click="navigate('publish')"
    >
      <i class="fa-solid fa-plus" />
    </button>

    <!-- Right two items -->
    <button
      v-for="item in rightItems"
      :key="item.name"
      type="button"
      class="w-11 h-10 flex items-center justify-center rounded-xl transition-all duration-200"
      :class="activeTab === item.tab
        ? 'text-white'
        : 'text-gray-500 hover:text-gray-300'"
      @click="navigate(item.name)"
    >
      <i class="fa-solid text-[20px]" :class="item.icon" />
    </button>
  </nav>
</template>
