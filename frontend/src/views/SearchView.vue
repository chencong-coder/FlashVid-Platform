<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { searchVideos } from '@/api/video'
import { searchUsers, type SearchUserItem } from '@/api/user'
import type { FeedVideo } from '@/api/feed'
import { formatCount } from '@/utils/format'
import AddToPlaylistModal from '@/components/AddToPlaylistModal.vue'
import { useAuthModalStore } from '@/store/authModal'

type Tab = 'videos' | 'users'
const tabs: { key: Tab; label: string }[] = [
  { key: 'videos', label: '视频' },
  { key: 'users', label: '用户' },
]

const route = useRoute()
const router = useRouter()
const authModal = useAuthModalStore()

const playlistVideoId = ref<string | null>(null)

const keyword = ref((route.query.q as string) ?? '')
const inputKw = ref(keyword.value)
const activeTab = ref<Tab>('videos')

// ── 视频 state ─────────────────────────────────────────────────
const videos = ref<FeedVideo[]>([])
const videoPage = ref(1)
const videoTotal = ref(0)
const videoPageSize = 20
const videoLoading = ref(false)
const videoHasMore = computed(() => videos.value.length < videoTotal.value)

// ── 用户 state ─────────────────────────────────────────────────
const users = ref<SearchUserItem[]>([])
const usersLoading = ref(false)
const usersLoaded = ref(false)

// ── Fetch ──────────────────────────────────────────────────────
const loadVideos = async (reset = false) => {
  const kw = keyword.value
  if (!kw) return
  if (!reset && (videoLoading.value || !videoHasMore.value)) return
  if (reset) { videos.value = []; videoPage.value = 1; videoTotal.value = 0 }
  videoLoading.value = true
  try {
    const res = await searchVideos({ keyword: kw, page: videoPage.value, pageSize: videoPageSize })
    const data = res.data.data
    videos.value = reset ? data.videos : [...videos.value, ...data.videos]
    videoTotal.value = data.pagination.total
    videoPage.value = reset ? 2 : videoPage.value + 1
  } catch {
    showToast('搜索视频失败')
  } finally {
    videoLoading.value = false
  }
}

const loadUsers = async () => {
  const kw = keyword.value
  if (!kw || usersLoaded.value) return
  usersLoading.value = true
  try {
    const res = await searchUsers(kw, 50)
    users.value = res.data.data.users ?? []
    usersLoaded.value = true
  } catch {
    showToast('搜索用户失败')
  } finally {
    usersLoading.value = false
  }
}

const resetAndLoad = () => {
  users.value = []
  usersLoaded.value = false
  void loadVideos(true)
  if (activeTab.value === 'users') void loadUsers()
}

const switchTab = (tab: Tab) => {
  activeTab.value = tab
  if (tab === 'users' && !usersLoaded.value) void loadUsers()
}

// ── 跳转 ───────────────────────────────────────────────────────
// 没有独立视频详情路由，视频挂在作者主页下，点击跳作者主页
const goUser = (userId: string) => {
  void router.push({ name: 'user-profile', params: { id: userId } })
}
const goVideoAuthor = (video: FeedVideo) => {
  void router.push({ name: 'user-profile', params: { id: video.author.id } })
}

const handleSearch = () => {
  const kw = inputKw.value.trim()
  if (!kw || kw === keyword.value) return
  void router.push({ name: 'search', query: { q: kw } })
}

// ── 关键词随路由变化 ───────────────────────────────────────────
watch(
  () => route.query.q,
  (q) => {
    keyword.value = (q as string) ?? ''
    inputKw.value = keyword.value
    if (keyword.value) resetAndLoad()
  },
)

onMounted(() => {
  if (keyword.value) resetAndLoad()
})
</script>

<template>
  <main class="flex h-full flex-col overflow-hidden bg-[#0d0d0d] text-white">
    <!-- 头部：返回 + 搜索框 -->
    <header
      class="safe-top flex shrink-0 items-center gap-2 bg-[#0d0d0d]/95 px-3 pb-2 pt-3 backdrop-blur"
    >
      <button
        type="button"
        aria-label="返回"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-gray-400 hover:bg-white/5"
        @click="router.back()"
      >
        <i class="fa-solid fa-arrow-left" />
      </button>
      <div class="flex h-9 flex-1 items-center gap-2 rounded-full bg-white/10 px-3 text-sm">
        <i class="fa-solid fa-magnifying-glass shrink-0 text-xs text-gray-500" />
        <input
          v-model="inputKw"
          type="search"
          placeholder="搜索你感兴趣的内容"
          class="min-w-0 flex-1 bg-transparent text-white outline-none placeholder:text-gray-500"
          @keydown.enter="handleSearch"
        />
        <button
          v-if="inputKw"
          type="button"
          aria-label="清空"
          class="shrink-0 text-gray-500"
          @click="inputKw = ''"
        >
          <i class="fa-solid fa-circle-xmark text-xs" />
        </button>
      </div>
      <button
        type="button"
        class="shrink-0 px-1 text-sm font-medium text-violet-400"
        @click="handleSearch"
      >
        搜索
      </button>
    </header>

    <!-- Tab 切换 -->
    <div class="flex shrink-0 border-b border-white/[0.06] px-4">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="-mb-px mr-6 border-b-2 py-3 text-sm font-medium transition-colors"
        :class="
          activeTab === tab.key
            ? 'border-violet-500 text-white'
            : 'border-transparent text-gray-500 hover:text-gray-300'
        "
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 内容区 -->
    <div class="no-scrollbar flex-1 overflow-y-auto">
      <!-- ===== 视频 tab ===== -->
      <template v-if="activeTab === 'videos'">
        <div v-if="videoLoading && videos.length === 0" class="grid grid-cols-2 gap-px pt-px">
          <div v-for="n in 8" :key="n" class="aspect-[3/4] animate-pulse bg-neutral-800" />
        </div>

        <div v-else-if="videos.length" class="grid grid-cols-2 gap-px pt-px">
          <button
            v-for="video in videos"
            :key="video.id"
            type="button"
            class="relative aspect-[3/4] overflow-hidden bg-neutral-900 text-left"
            @click="goVideoAuthor(video)"
          >
            <img
              v-if="video.coverUrl"
              :src="video.coverUrl"
              :alt="video.title"
              loading="lazy"
              class="h-full w-full object-cover"
            />
            <!-- 加入播放列表按钮 -->
            <button
              type="button"
              aria-label="加入播放列表"
              class="absolute right-1.5 top-1.5 flex h-7 w-7 items-center justify-center rounded-full bg-black/50 text-white backdrop-blur-sm"
              @click.stop="authModal.requireLogin() && (playlistVideoId = video.id)"
            >
              <i class="fa-solid fa-list-ul text-[11px]" />
            </button>
            <div
              class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent px-2 pb-2 pt-8"
            >
              <p class="line-clamp-2 text-xs leading-4 text-white">{{ video.title }}</p>
              <p class="mt-0.5 flex items-center gap-1 text-[10px] text-neutral-400">
                <i class="fa-solid fa-play text-[9px]" />{{ formatCount(video.stats.viewCount) }}
              </p>
            </div>
          </button>
        </div>

        <div v-if="videos.length" class="py-5 text-center">
          <button
            v-if="videoHasMore"
            type="button"
            class="text-xs text-gray-500 hover:text-gray-300 disabled:opacity-40"
            :disabled="videoLoading"
            @click="void loadVideos(false)"
          >
            {{ videoLoading ? '加载中…' : '加载更多' }}
          </button>
          <span v-else class="text-xs text-gray-700">没有更多了</span>
        </div>

        <div
          v-else-if="!videoLoading"
          class="flex flex-col items-center justify-center py-20 text-gray-600"
        >
          <i class="fa-regular fa-circle-play mb-3 text-3xl" />
          <p class="text-sm">{{ keyword ? `没有找到“${keyword}”相关视频` : '输入关键词搜索' }}</p>
        </div>
      </template>

      <!-- ===== 用户 tab ===== -->
      <template v-else>
        <div v-if="usersLoading" class="space-y-3 px-4 pt-3">
          <div v-for="n in 6" :key="n" class="flex items-center gap-3">
            <div class="h-12 w-12 shrink-0 animate-pulse rounded-full bg-neutral-800" />
            <div class="flex-1 space-y-1.5">
              <div class="h-3.5 w-1/3 animate-pulse rounded bg-neutral-800" />
              <div class="h-3 w-2/3 animate-pulse rounded bg-neutral-800" />
            </div>
          </div>
        </div>

        <ul v-else-if="users.length" class="divide-y divide-white/[0.05] px-4">
          <li v-for="user in users" :key="user.userId">
            <button
              type="button"
              class="flex w-full items-center gap-3 py-3 text-left transition-colors hover:bg-white/[0.03]"
              @click="goUser(user.userId)"
            >
              <img
                v-if="user.avatar"
                :src="user.avatar"
                :alt="user.nickname"
                class="h-12 w-12 shrink-0 rounded-full object-cover"
              />
              <div
                v-else
                class="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-violet-500 to-purple-600 font-bold text-white"
              >
                {{ user.nickname?.[0]?.toUpperCase() ?? 'U' }}
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-white">{{ user.nickname }}</p>
                <p class="mt-0.5 text-[12px] text-gray-500">
                  {{ formatCount(user.followerCount) }} 粉丝
                </p>
                <p v-if="user.bio" class="mt-0.5 truncate text-xs text-gray-600">{{ user.bio }}</p>
              </div>
              <i class="fa-solid fa-chevron-right shrink-0 text-xs text-gray-700" />
            </button>
          </li>
        </ul>

        <div
          v-else-if="!usersLoading"
          class="flex flex-col items-center justify-center py-20 text-gray-600"
        >
          <i class="fa-regular fa-user mb-3 text-3xl" />
          <p class="text-sm">{{ keyword ? `没有找到“${keyword}”相关用户` : '输入关键词搜索' }}</p>
        </div>
      </template>
    </div>
    <AddToPlaylistModal :video-id="playlistVideoId" @close="playlistVideoId = null" />
  </main>
</template>
