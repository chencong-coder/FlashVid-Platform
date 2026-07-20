<script setup lang="ts">
import type { FeedType } from '@/types/video'

interface TabItem {
  label: string
  value: FeedType
}

interface Props {
  active: FeedType
}

interface Emits {
  (event: 'change', value: FeedType): void
  (event: 'search'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

const tabs: TabItem[] = [
  { label: '关注', value: 'follow' },
  { label: '推荐', value: 'recommend' },
  { label: '同城', value: 'nearby' },
]
</script>

<template>
  <header
    class="safe-top pointer-events-none absolute inset-x-0 top-0 z-30 flex items-center justify-center px-4 pb-6 pt-3"
  >
    <!-- 顶部渐变背景 -->
    <div class="absolute inset-0 bg-gradient-to-b from-black/80 via-black/40 to-transparent" />

    <!-- Tab 切换区域 - 毛玻璃胶囊设计 -->
    <div
      class="pointer-events-auto relative flex items-center gap-1 rounded-full bg-white/10 p-1 backdrop-blur-xl"
    >
      <button
        v-for="tab in tabs"
        :key="tab.value"
        type="button"
        class="relative z-10 rounded-full px-5 py-2 text-sm font-semibold transition-all duration-300"
        :class="active === tab.value ? 'text-white' : 'text-white/60 hover:text-white/80'"
        @click="emit('change', tab.value)"
      >
        {{ tab.label }}
      </button>

      <!-- 滑动指示器 - 跟随 active tab -->
      <div
        class="absolute left-1 top-1 bottom-1 rounded-full bg-gradient-to-br from-pink-500 to-purple-600 shadow-lg transition-all duration-300"
        :style="{
          width: 'calc(33.333% - 0.25rem)',
          transform: `translateX(${tabs.findIndex((t) => t.value === active) * 100}%)`,
        }"
      />
    </div>

    <!-- 搜索按钮 - 圆形毛玻璃 -->
    <button
      type="button"
      aria-label="搜索"
      class="pointer-events-auto absolute right-4 top-[calc(env(safe-area-inset-top)+0.75rem)] flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-lg text-white backdrop-blur-xl transition-all duration-300 hover:bg-white/20 hover:scale-110 active:scale-95"
      @click="emit('search')"
    >
      <i class="fa-solid fa-magnifying-glass" />
    </button>
  </header>
</template>
