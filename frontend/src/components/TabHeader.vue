<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { TopNavValue } from '@/types/video'

interface TabItem {
  label: string
  value: TopNavValue
}

interface Props {
  active: TopNavValue
}

interface Emits {
  (event: 'change', value: TopNavValue): void
  (event: 'search'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const tabs: TabItem[] = [
  { label: '关注', value: 'follow' },
  { label: '推荐', value: 'recommend' },
  { label: '发现', value: 'discover' },
  { label: '同城', value: 'nearby' },
]

// 容器与各 tab 按钮的引用，用于实测位置对齐滑块指示器
const listRef = ref<HTMLElement | null>(null)
const buttonRefs = ref<HTMLButtonElement[]>([])
const setButtonRef = (el: unknown, index: number) => {
  if (el instanceof HTMLButtonElement) buttonRefs.value[index] = el
}

// 指示器的真实像素位置/宽度：从 active 按钮相对容器内容盒实测得到，
// 避免用百分比+gap 估算导致越靠后的 tab 偏移累积越大。
const indicatorStyle = ref<{ width: string; transform: string }>({
  width: '0px',
  transform: 'translateX(0px)',
})

const updateIndicator = () => {
  const list = listRef.value
  const index = tabs.findIndex((t) => t.value === props.active)
  const btn = buttonRefs.value[index >= 0 ? index : 0]
  if (!list || !btn) return
  // offsetLeft 相对容器 border 盒；指示器用 left-0 定位在 padding 盒，
  // 故需减去容器左内边距，得到指示器坐标系下的偏移。
  const paddingLeft = parseFloat(getComputedStyle(list).paddingLeft) || 0
  indicatorStyle.value = {
    width: `${btn.offsetWidth}px`,
    transform: `translateX(${btn.offsetLeft - paddingLeft}px)`,
  }
}

const scheduleUpdate = () => {
  void nextTick(updateIndicator)
}

watch(() => props.active, scheduleUpdate)

let resizeObserver: ResizeObserver | null = null
onMounted(() => {
  scheduleUpdate()
  if (typeof ResizeObserver !== 'undefined' && listRef.value) {
    resizeObserver = new ResizeObserver(updateIndicator)
    resizeObserver.observe(listRef.value)
  }
  window.addEventListener('resize', updateIndicator)
})
onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', updateIndicator)
})
</script>

<template>
  <header
    class="safe-top pointer-events-none absolute inset-x-0 top-0 z-30 flex items-center justify-center px-4 pb-6 pt-3"
  >
    <!-- 顶部渐变背景 -->
    <div class="absolute inset-0 bg-gradient-to-b from-black/80 via-black/40 to-transparent" />

    <!-- Tab 切换区域 - 毛玻璃胶囊设计 -->
    <div
      ref="listRef"
      class="pointer-events-auto relative flex items-center gap-1 rounded-full bg-white/10 p-1 backdrop-blur-xl"
    >
      <!-- 滑动指示器 - 跟随 active tab（实测对齐） -->
      <div
        class="pointer-events-none absolute left-1 top-1 bottom-1 rounded-full bg-gradient-to-br from-pink-500 to-purple-600 shadow-lg transition-all duration-300 ease-out"
        :style="indicatorStyle"
      />

      <button
        v-for="(tab, index) in tabs"
        :key="tab.value"
        :ref="(el) => setButtonRef(el, index)"
        type="button"
        class="relative z-10 rounded-full px-4 py-2 text-sm font-semibold transition-colors duration-300"
        :class="active === tab.value ? 'text-white' : 'text-white/60 hover:text-white/80'"
        @click="emit('change', tab.value)"
      >
        {{ tab.label }}
      </button>
    </div>
  </header>
</template>
