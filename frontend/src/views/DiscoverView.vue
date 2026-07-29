<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import { getTopics, searchTopics, type TopicItem } from '@/api/topic'
import { formatCount } from '@/utils/format'

const router = useRouter()

type SortMode = 'hot' | 'latest'
const sort = ref<SortMode>('hot')
const keyword = ref('')

const topics = ref<TopicItem[]>([])
const cursor = ref('')
const hasMore = ref(true)
const loading = ref(false)
const searching = ref(false)

// 拉取热门/最新话题（游标分页）
const loadTopics = async (reset = false): Promise<void> => {
  if (loading.value || (!reset && !hasMore.value)) return
  loading.value = true
  try {
    const res = await getTopics({
      sort: sort.value,
      cursor: reset ? undefined : cursor.value || undefined,
      count: 20,
    })
    const page = res.data.data
    topics.value = reset ? page.topics : [...topics.value, ...page.topics]
    cursor.value = page.nextCursorToken
    hasMore.value = page.hasMore
  } catch {
    showToast('加载话题失败')
  } finally {
    loading.value = false
  }
}

// 搜索话题（关键词非空时覆盖列表）
const runSearch = async (): Promise<void> => {
  const kw = keyword.value.trim()
  if (!kw) {
    searching.value = false
    void loadTopics(true)
    return
  }
  searching.value = true
  loading.value = true
  try {
    const res = await searchTopics(kw, undefined, 20)
    topics.value = res.data.data.topics
    hasMore.value = false
  } catch {
    showToast('搜索失败')
  } finally {
    loading.value = false
  }
}

const switchSort = (mode: SortMode): void => {
  if (sort.value === mode || searching.value) return
  sort.value = mode
  void loadTopics(true)
}

const openTopic = (topic: TopicItem): void => {
  void router.push({ name: 'topic', params: { id: String(topic.id) } })
}

let searchTimer: number | undefined
watch(keyword, () => {
  window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => void runSearch(), 350)
})

onMounted(() => {
  void loadTopics(true)
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

    <section class="px-3 pb-6 pt-3">
      <!-- 加载骨架 -->
      <div v-if="loading && topics.length === 0" class="grid grid-cols-2 gap-1.5">
        <div v-for="n in 6" :key="n" class="aspect-[3/4] animate-pulse rounded bg-neutral-800" />
      </div>

      <!-- 话题卡片网格 -->
      <div v-else-if="topics.length" class="grid grid-cols-2 gap-1.5">
        <article
          v-for="topic in topics"
          :key="topic.id"
          class="relative aspect-[3/4] cursor-pointer overflow-hidden rounded bg-neutral-900 transition-transform active:scale-[0.98]"
          @click="openTopic(topic)"
        >
          <img
            :src="topic.coverUrl"
            :alt="topic.name"
            loading="lazy"
            class="h-full w-full object-cover"
          />
          <div
            class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/90 to-transparent px-2 pb-2 pt-10"
          >
            <p class="line-clamp-1 text-sm font-semibold leading-5"># {{ topic.name }}</p>
            <div class="mt-1 flex items-center justify-between text-[10px] text-neutral-300">
              <span>{{ formatCount(topic.videoCount) }} 个视频</span>
              <span><i class="fa-solid fa-fire mr-1" />{{ formatCount(topic.viewCount) }}</span>
            </div>
          </div>
        </article>
      </div>

      <!-- 空态 -->
      <div v-else class="flex flex-col items-center justify-center py-20 text-neutral-500">
        <i class="fa-regular fa-compass mb-3 text-3xl" />
        <p class="text-sm">{{ searching ? '没有找到相关话题' : '暂无话题' }}</p>
      </div>

      <!-- 加载更多 -->
      <div v-if="topics.length" class="py-5 text-center">
        <button
          v-if="hasMore && !searching"
          type="button"
          class="text-xs text-neutral-500 hover:text-neutral-300 disabled:opacity-40"
          :disabled="loading"
          @click="void loadTopics(false)"
        >
          {{ loading ? '加载中…' : '加载更多' }}
        </button>
        <span v-else-if="!searching" class="text-xs text-neutral-600">没有更多了</span>
      </div>
    </section>
  </main>
</template>

