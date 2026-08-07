<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import { getTopics, searchTopics, type TopicItem, type TopicSort } from '@/api/topic'
import { formatCount } from '@/utils/format'

const router = useRouter()

const sort = ref<TopicSort>('hot')
const keyword = ref('')

const topics = ref<TopicItem[]>([])
const cursor = ref('')
const hasMore = ref(true)
const loading = ref(false)
const searching = computed(() => keyword.value.trim().length > 0)

let requestVersion = 0
let searchTimer: number | undefined

// 普通列表和搜索列表共用游标；请求版本用于丢弃关键词或排序变化后的旧响应。
const loadTopics = async (reset = false): Promise<void> => {
  if (!reset && (loading.value || !hasMore.value)) return

  const currentVersion = reset ? ++requestVersion : requestVersion
  const currentKeyword = keyword.value.trim()
  const requestCursor = reset ? undefined : cursor.value || undefined

  if (reset) {
    topics.value = []
    cursor.value = ''
    hasMore.value = true
  }
  loading.value = true

  try {
    const res = currentKeyword
      ? await searchTopics(currentKeyword, requestCursor, 20)
      : await getTopics({
          sort: sort.value,
          cursor: requestCursor,
          count: 20,
        })

    if (currentVersion !== requestVersion || currentKeyword !== keyword.value.trim()) return

    const page = res.data.data
    topics.value = reset ? page.topics : [...topics.value, ...page.topics]
    cursor.value = page.nextCursorToken
    hasMore.value = page.hasMore
  } catch {
    if (currentVersion === requestVersion && currentKeyword === keyword.value.trim()) {
      showToast(currentKeyword ? '搜索失败' : '加载话题失败')
    }
  } finally {
    if (currentVersion === requestVersion && currentKeyword === keyword.value.trim()) {
      loading.value = false
    }
  }
}

const switchSort = (mode: TopicSort): void => {
  if (sort.value === mode || searching.value) return
  sort.value = mode
  void loadTopics(true)
}

const openTopic = (topic: TopicItem): void => {
  void router.push({ name: 'topic', params: { id: String(topic.id) } })
}

// 名次配色：前三名金/银/铜，其余暗灰。搜索结果不强调名次。
const rankClass = (index: number): string => {
  if (searching.value) return 'text-neutral-600'
  if (index === 0) return 'text-amber-400'
  if (index === 1) return 'text-slate-300'
  if (index === 2) return 'text-orange-400'
  return 'text-neutral-600'
}

watch(keyword, () => {
  window.clearTimeout(searchTimer)
  requestVersion += 1
  searchTimer = window.setTimeout(() => void loadTopics(true), 350)
})

onMounted(() => {
  void loadTopics(true)
})

onBeforeUnmount(() => {
  window.clearTimeout(searchTimer)
  requestVersion += 1
})
</script>

<template>
  <main class="no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <header class="safe-top sticky top-0 z-20 bg-[#0d0d0d]/95 px-4 pb-3 pt-3 backdrop-blur">
      <div class="flex h-10 items-center gap-3 rounded bg-white/10 px-3 text-neutral-400">
        <i class="fa-solid fa-magnifying-glass text-sm" />
        <input
          v-model="keyword"
          type="search"
          placeholder="搜索话题"
          class="min-w-0 flex-1 bg-transparent text-sm text-white outline-none placeholder:text-neutral-500"
        />
      </div>
      <!-- 排序切换（搜索时隐藏） -->
      <div v-if="!searching" class="mt-4 flex gap-5 text-sm">
        <button
          type="button"
          :class="sort === 'hot' ? 'font-semibold text-white' : 'text-neutral-500'"
          @click="switchSort('hot')"
        >
          热门话题
        </button>
        <button
          type="button"
          :class="sort === 'latest' ? 'font-semibold text-white' : 'text-neutral-500'"
          @click="switchSort('latest')"
        >
          最新话题
        </button>
      </div>
    </header>

    <section class="px-3 pb-6 pt-2">
      <!-- 加载骨架 -->
      <div v-if="loading && topics.length === 0" class="space-y-1">
        <div
          v-for="n in 8"
          :key="n"
          class="flex items-center gap-3 rounded-lg px-2 py-3"
        >
          <div class="h-6 w-6 shrink-0 animate-pulse rounded bg-neutral-800" />
          <div class="flex-1 space-y-2">
            <div class="h-3.5 w-1/2 animate-pulse rounded bg-neutral-800" />
            <div class="h-2.5 w-1/3 animate-pulse rounded bg-neutral-800/70" />
          </div>
        </div>
      </div>

      <!-- 话题热榜列表 -->
      <ol v-else-if="topics.length" class="space-y-0.5">
        <li
          v-for="(topic, index) in topics"
          :key="topic.id"
          class="group flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2.5 transition-colors hover:bg-white/[0.04] active:bg-white/[0.07]"
          @click="openTopic(topic)"
        >
          <!-- 名次 -->
          <div class="flex w-7 shrink-0 flex-col items-center">
            <span :class="['text-lg font-bold italic leading-none tabular-nums', rankClass(index)]">
              {{ index + 1 }}
            </span>
            <i
              v-if="index < 3 && !searching"
              class="fa-solid fa-fire mt-0.5 text-[9px] text-rose-500/70"
            />
          </div>

          <!-- 话题信息 -->
          <div class="min-w-0 flex-1">
            <p class="truncate text-[15px] font-medium leading-5 text-white group-hover:text-rose-400">
              {{ topic.name }}
            </p>
            <div class="mt-0.5 flex items-center gap-3 text-[11px] text-neutral-500">
              <span>{{ formatCount(topic.videoCount) }} 个视频</span>
              <span class="flex items-center gap-1">
                <i class="fa-solid fa-fire-flame-curved text-[10px] text-orange-500/60" />
                {{ formatCount(topic.viewCount) }} 热度
              </span>
            </div>
          </div>

          <i class="fa-solid fa-chevron-right shrink-0 text-xs text-neutral-700 group-hover:text-neutral-500" />
        </li>
      </ol>

      <!-- 空态 -->
      <div v-else class="flex flex-col items-center justify-center py-20 text-neutral-500">
        <i class="fa-regular fa-compass mb-3 text-3xl" />
        <p class="text-sm">{{ searching ? '没有找到相关话题' : '暂无话题' }}</p>
      </div>

      <!-- 加载更多 -->
      <div v-if="topics.length" class="py-5 text-center">
        <button
          v-if="hasMore"
          type="button"
          class="text-xs text-neutral-500 hover:text-neutral-300 disabled:opacity-40"
          :disabled="loading"
          @click="void loadTopics(false)"
        >
          {{ loading ? '加载中…' : '加载更多' }}
        </button>
        <span v-else class="text-xs text-neutral-600">没有更多了</span>
      </div>
    </section>
  </main>
</template>
