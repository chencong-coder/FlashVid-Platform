<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'

import { getTopicById, getTopicVideos, type TopicItem } from '@/api/topic'
import { mapFeedVideo, type FeedVideo } from '@/api/feed'
import { useVideoStore } from '@/store/video'
import { formatCount } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const videoStore = useVideoStore()

const topicId = String(route.params.id)
const topic = ref<TopicItem | null>(null)

type SortMode = 'popular' | 'latest'
const sort = ref<SortMode>('popular')

const videos = ref<FeedVideo[]>([])
const cursor = ref('')
const hasMore = ref(true)
const loading = ref(false)

const loadTopic = async (): Promise<void> => {
  try {
    const res = await getTopicById(topicId)
    topic.value = res.data.data.topic
  } catch {
    /* 标题栏兜底为空 */
  }
}

const loadVideos = async (reset = false): Promise<void> => {
  if (loading.value || (!reset && !hasMore.value)) return
  loading.value = true
  try {
    const res = await getTopicVideos(topicId, {
      sort: sort.value,
      cursor: reset ? undefined : cursor.value || undefined,
      count: 21,
    })
    const page = res.data.data
    videos.value = reset ? page.videos : [...videos.value, ...page.videos]
    cursor.value = page.nextCursorToken
    hasMore.value = page.hasMore
  } catch {
    showToast('加载视频失败')
  } finally {
    loading.value = false
  }
}

const switchSort = (mode: SortMode): void => {
  if (sort.value === mode) return
  sort.value = mode
  void loadVideos(true)
}

// 点封面：把已加载数据灌入 store 的 topic 流，跳全屏播放并定位到该条
const openVideo = (index: number): void => {
  videoStore.startTopicFeed(
    topicId,
    sort.value,
    videos.value.map((v) => mapFeedVideo(v, 'topic')),
    cursor.value,
    hasMore.value,
  )
  void router.push({ name: 'topic-play', params: { id: topicId }, query: { index: String(index) } })
}

onMounted(() => {
  void loadTopic()
  void loadVideos(true)
})
</script>

<template>
  <main class="no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <!-- 话题头部 -->
    <div class="relative h-40 bg-neutral-900">
      <img
        v-if="topic?.coverUrl"
        :src="topic.coverUrl"
        :alt="topic.name"
        class="h-full w-full object-cover opacity-50"
      />
      <div class="absolute inset-0 bg-gradient-to-b from-black/40 to-[#0d0d0d]" />
      <button
        aria-label="返回"
        class="safe-top absolute left-4 top-4 flex h-9 w-9 items-center justify-center rounded-full bg-black/40 text-white backdrop-blur"
        @click="router.back()"
      >
        <i class="fa-solid fa-arrow-left" />
      </button>
      <div class="absolute inset-x-0 bottom-0 px-4 pb-3">
        <h1 class="text-xl font-bold"># {{ topic?.name || '话题' }}</h1>
        <div class="mt-1 flex gap-4 text-xs text-neutral-300">
          <span>{{ formatCount(topic?.videoCount ?? 0) }} 个视频</span>
          <span><i class="fa-solid fa-fire mr-1" />{{ formatCount(topic?.viewCount ?? 0) }} 播放</span>
        </div>
        <p v-if="topic?.description" class="mt-2 line-clamp-2 text-xs text-neutral-400">
          {{ topic.description }}
        </p>
      </div>
    </div>

    <!-- 排序切换 -->
    <div class="flex gap-5 border-b border-white/5 px-4 py-3 text-sm">
      <button
        type="button"
        :class="sort === 'popular' ? 'font-semibold text-white' : 'text-neutral-500'"
        @click="switchSort('popular')"
      >
        最热
      </button>
      <button
        type="button"
        :class="sort === 'latest' ? 'font-semibold text-white' : 'text-neutral-500'"
        @click="switchSort('latest')"
      >
        最新
      </button>
    </div>

    <section class="px-0.5 pb-6">
      <!-- 骨架 -->
      <div v-if="loading && videos.length === 0" class="grid grid-cols-3 gap-0.5">
        <div v-for="n in 9" :key="n" class="aspect-[3/4] animate-pulse bg-neutral-800" />
      </div>

      <!-- 视频封面网格 -->
      <div v-else-if="videos.length" class="grid grid-cols-3 gap-0.5">
        <div
          v-for="(video, index) in videos"
          :key="video.id"
          class="relative aspect-[3/4] cursor-pointer overflow-hidden bg-neutral-900 transition-transform active:scale-[0.98]"
          @click="openVideo(index)"
        >
          <img
            :src="video.coverUrl"
            :alt="video.title"
            loading="lazy"
            class="h-full w-full object-cover"
          />
          <span class="absolute bottom-1 left-1 text-[10px] text-white">
            <i class="fa-solid fa-play mr-1" />{{ formatCount(video.stats.likeCount) }}
          </span>
        </div>
      </div>

      <!-- 空态 -->
      <div v-else class="flex flex-col items-center justify-center py-20 text-neutral-500">
        <i class="fa-regular fa-folder-open mb-3 text-3xl" />
        <p class="text-sm">该话题下暂无视频</p>
      </div>

      <!-- 加载更多 -->
      <div v-if="videos.length" class="py-5 text-center">
        <button
          v-if="hasMore"
          type="button"
          class="text-xs text-neutral-500 hover:text-neutral-300 disabled:opacity-40"
          :disabled="loading"
          @click="void loadVideos(false)"
        >
          {{ loading ? '加载中…' : '加载更多' }}
        </button>
        <span v-else class="text-xs text-neutral-600">没有更多了</span>
      </div>
    </section>
  </main>
</template>

