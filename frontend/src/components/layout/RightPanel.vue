<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getTopics, type TopicItem } from '@/api/topic'
import { useAuthModalStore } from '@/store/authModal'

const router = useRouter()
const authModal = useAuthModalStore()
const trendingTopics = ref<TopicItem[]>([])
const loading = ref(true)

// API 无数据时的兜底 mock
const mockTrending = [
  { id: 1, name: '慵懒周末', videoCount: 2_400_000 },
  { id: 2, name: '晨间routine', videoCount: 1_800_000 },
  { id: 3, name: '旅行日记', videoCount: 1_200_000 },
  { id: 4, name: '舞蹈挑战', videoCount: 965_000 },
  { id: 5, name: '美食ASMR', videoCount: 872_000 },
] as TopicItem[]

const suggestions = [
  { name: '林朴', cat: '旅行', avatar: 'https://picsum.photos/40/40?random=31' },
  { name: '李杰森', cat: '搞笑', avatar: 'https://picsum.photos/40/40?random=32' },
  { name: '阮索菲', cat: '生活', avatar: 'https://picsum.photos/40/40?random=33' },
  { name: '瑞麦克', cat: '健身', avatar: 'https://picsum.photos/40/40?random=34' },
  { name: '陈艾娃', cat: '艺术', avatar: 'https://picsum.photos/40/40?random=35' },
]

const fmtVideos = (n: number): string => {
  if (n >= 100_000_000) return `${(n / 100_000_000).toFixed(1)}亿个视频`
  if (n >= 10_000) return `${(n / 10_000).toFixed(0)}万个视频`
  return `${n} 个视频`
}

const loadTopics = async () => {
  loading.value = true
  try {
    const res = await getTopics({ sort: 'hot', count: 5 })
    trendingTopics.value = res.data.data?.topics ?? []
  } catch {
    /* use mock */
  } finally {
    loading.value = false
  }
}

onMounted(loadTopics)

const displayTopics = () => (trendingTopics.value.length ? trendingTopics.value : mockTrending)
</script>

<template>
  <aside
    class="w-[288px] shrink-0 h-full overflow-y-auto bg-[#0d0d10] px-4 py-4 space-y-3 scrollbar-none"
  >
    <!-- ① Trending Now -->
    <div class="bg-[#111115] rounded-2xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-orange-500 shrink-0 animate-pulse" />
          正在流行
        </h3>
        <button
          type="button"
          class="text-xs text-violet-400 hover:text-violet-300 transition-colors"
          @click="router.push({ name: 'discover' })"
        >
          查看全部
        </button>
      </div>

      <!-- Skeleton -->
      <template v-if="loading">
        <div v-for="i in 5" :key="i" class="flex items-center gap-3 py-1.5">
          <div class="w-4 h-3 bg-white/5 rounded animate-pulse shrink-0" />
          <div class="flex-1 space-y-1">
            <div class="h-3 bg-white/5 rounded animate-pulse w-3/4" />
            <div class="h-2 bg-white/5 rounded animate-pulse w-1/2" />
          </div>
        </div>
      </template>

      <!-- List -->
      <template v-else>
        <button
          v-for="(topic, i) in displayTopics()"
          :key="topic.id"
          type="button"
          class="w-full flex items-center gap-3 py-2 px-1 rounded-lg hover:bg-white/[0.04] transition-colors"
          @click="router.push({ name: 'discover' })"
        >
          <span
            class="text-sm font-bold w-4 text-center shrink-0"
            :class="i === 0 ? 'text-orange-400' : 'text-gray-600'"
            >{{ i + 1 }}</span
          >
          <div class="flex-1 min-w-0 text-left">
            <p class="text-sm text-white font-medium truncate">{{ topic.name }}</p>
            <p class="text-xs text-gray-500">{{ fmtVideos(topic.videoCount) }}</p>
          </div>
        </button>
      </template>
    </div>

    <!-- ② Suggested for you -->
    <div class="bg-[#111115] rounded-2xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white">为你推荐</h3>
        <button
          type="button"
          class="w-7 h-7 flex items-center justify-center rounded-full hover:bg-white/10 text-gray-500 hover:text-white transition-colors"
          @click="loadTopics"
        >
          <i class="fa-solid fa-rotate-right text-xs" />
        </button>
      </div>

      <div class="space-y-1">
        <div v-for="user in suggestions" :key="user.name" class="flex items-center gap-3 py-1.5">
          <img
            :src="user.avatar"
            :alt="user.name"
            class="w-9 h-9 rounded-full object-cover shrink-0"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm text-white font-medium truncate">{{ user.name }}</p>
            <p class="text-xs text-gray-500">{{ user.cat }}</p>
          </div>
          <button
            type="button"
            class="shrink-0 text-xs px-3 py-1.5 rounded-lg bg-white/[0.06] border border-white/[0.1] text-gray-300 hover:bg-white/[0.12] hover:text-white transition-all"
            @click="authModal.requireLogin()"
          >
            关注
          </button>
        </div>
      </div>
    </div>

    <!-- ③ Go Premium -->
    <div
      class="rounded-2xl bg-gradient-to-br from-[#1e1b4b] via-[#2e1d6b] to-[#1a0d3e] p-5 relative overflow-hidden"
    >
      <!-- Subtle glow blobs -->
      <div
        class="absolute -top-6 -right-6 w-24 h-24 rounded-full bg-violet-600/20 blur-2xl pointer-events-none"
      />
      <div
        class="absolute -bottom-4 -left-4 w-20 h-20 rounded-full bg-indigo-600/15 blur-2xl pointer-events-none"
      />

      <div class="relative">
        <!-- Lightning icon badge -->
        <div class="w-10 h-10 rounded-xl bg-yellow-400/15 flex items-center justify-center mb-3">
          <i class="fa-solid fa-bolt text-yellow-400 text-lg" />
        </div>
        <h3 class="text-white font-bold text-base mb-1">开通会员</h3>
        <p class="text-gray-300 text-xs mb-4 leading-relaxed">解锁专属内容与特效。</p>
        <button
          type="button"
          class="w-full py-2.5 rounded-xl bg-white/[0.12] hover:bg-white/[0.2] active:bg-white/[0.08] text-white text-sm font-medium border border-white/[0.15] transition-all"
        >
          免费试用
        </button>
      </div>
    </div>
  </aside>
</template>
