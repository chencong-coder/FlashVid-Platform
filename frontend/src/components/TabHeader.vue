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
    class="safe-top pointer-events-none absolute inset-x-0 top-0 z-30 flex items-center justify-center bg-gradient-to-b from-black/70 to-transparent px-4 pb-8 pt-3"
  >
    <div class="pointer-events-auto flex items-center gap-7">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        type="button"
        class="relative h-8 text-base font-medium text-white transition-opacity"
        :class="active === tab.value ? 'opacity-100' : 'opacity-60'"
        @click="emit('change', tab.value)"
      >
        {{ tab.label }}
        <span
          v-if="active === tab.value"
          class="absolute bottom-0 left-1/2 h-0.5 w-5 -translate-x-1/2 bg-white"
        />
      </button>
    </div>
    <button
      type="button"
      aria-label="搜索"
      class="pointer-events-auto absolute right-4 top-[calc(env(safe-area-inset-top)+0.75rem)] flex h-8 w-8 items-center justify-center text-lg text-white"
      @click="emit('search')"
    >
      <i class="fa-solid fa-magnifying-glass" />
    </button>
  </header>
</template>
