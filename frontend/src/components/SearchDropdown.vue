<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { getTopics, type TopicItem } from '@/api/topic'
import { formatCount } from '@/utils/format'

const router = useRouter()
const wrapperRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const keyword = ref('')
const open = ref(false)

// 热榜（无输入时）
const hotTopics = ref<TopicItem[]>([])

// ── 加载热榜 ──────────────────────────────────────────────────
const loadHot = async () => {
  try {
    const res = await getTopics({ sort: 'hot', count: 10 })
    hotTopics.value = res.data.data.topics
  } catch {
    hotTopics.value = []
  }
}

onMounted(() => {
  void loadHot()
  document.addEventListener('mousedown', onClickOutside)
  document.addEventListener('keydown', onGlobalKey)
})
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onClickOutside)
  document.removeEventListener('keydown', onGlobalKey)
})

// ── 点击外部关闭 ──────────────────────────────────────────────
const onClickOutside = (e: MouseEvent) => {
  if (wrapperRef.value && !wrapperRef.value.contains(e.target as Node)) open.value = false
}

// ── 全局快捷键 ────────────────────────────────────────────────
const onGlobalKey = (e: KeyboardEvent) => {
  if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
    e.preventDefault()
    open.value = true
    inputRef.value?.focus()
  }
  if (e.key === 'Escape') {
    open.value = false
    inputRef.value?.blur()
  }
}

// ── 跳转 ──────────────────────────────────────────────────────
const goTopic = (topicId: string) => {
  open.value = false
  keyword.value = ''
  void router.push({ name: 'topic', params: { id: topicId } })
}
const handleEnter = () => {
  const kw = keyword.value.trim()
  if (!kw) return
  open.value = false
  keyword.value = ''
  inputRef.value?.blur()
  void router.push({ name: 'search', query: { q: kw } })
}
</script>

<template>
  <div ref="wrapperRef" class="relative w-full max-w-md">
    <!-- 搜索输入框 -->
    <div
      class="flex items-center gap-3 w-full px-4 py-2.5 rounded-full bg-[#1c1c22] border transition-all"
      :class="open ? 'border-violet-500/60' : 'border-white/[0.07] hover:border-white/[0.15]'"
    >
      <i class="fa-solid fa-magnifying-glass text-xs shrink-0 text-gray-500" />
      <input
        ref="inputRef"
        v-model="keyword"
        type="search"
        placeholder="搜索你感兴趣的内容"
        class="flex-1 min-w-0 bg-transparent text-sm text-white outline-none placeholder:text-gray-500"
        @focus="open = true"
        @keydown.enter="handleEnter"
      />
      <button
        v-if="keyword"
        type="button"
        aria-label="清空"
        class="shrink-0 text-gray-500 hover:text-gray-300"
        @click="keyword = ''; inputRef?.focus()"
      >
        <i class="fa-solid fa-circle-xmark text-xs" />
      </button>
      <kbd
        v-else
        class="shrink-0 text-[10px] bg-white/[0.06] border border-white/10 rounded px-1.5 py-0.5 font-mono text-gray-600"
        >⌘ K</kbd
      >
    </div>

    <!-- 下拉面板：闪视热榜 -->
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      leave-active-class="transition duration-100 ease-in"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div
        v-if="open"
        class="absolute top-full left-0 right-0 mt-2 rounded-xl bg-[#1a1a22] border border-white/[0.08] shadow-2xl shadow-black/60 overflow-hidden z-50"
      >
        <div class="px-4 pt-3 pb-1.5">
          <span class="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">闪视热榜</span>
        </div>
        <ul v-if="hotTopics.length" class="pb-2 max-h-[60vh] overflow-y-auto">
          <li v-for="(topic, i) in hotTopics" :key="topic.id">
            <button
              type="button"
              class="w-full flex items-center gap-3 px-4 py-2.5 text-left text-sm hover:bg-white/5 transition-colors"
              @mousedown.prevent="goTopic(String(topic.id))"
            >
              <span
                class="w-4 text-center text-xs font-bold shrink-0"
                :class="i < 3 ? 'text-violet-400' : 'text-gray-600'"
                >{{ i + 1 }}</span
              >
              <span class="flex-1 truncate text-white">{{ topic.name }}</span>
              <span class="text-[11px] text-gray-500 shrink-0">{{ formatCount(topic.videoCount) }} 视频</span>
              <span
                v-if="i < 2"
                class="text-[10px] bg-violet-500/20 text-violet-300 rounded px-1 shrink-0"
                >热</span
              >
            </button>
          </li>
        </ul>
        <div v-else class="px-4 py-6 text-center text-sm text-gray-600">暂无热门话题</div>
      </div>
    </Transition>
  </div>
</template>
