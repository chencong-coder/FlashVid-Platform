<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getTopics, type TopicItem } from '@/api/topic'
import { getRecommendUsers, followUser, unfollowUser, type RecommendUser } from '@/api/user'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import { useVideoStore } from '@/store/video'

const router = useRouter()
const authModal = useAuthModalStore()
const userStore = useUserStore()
const videoStore = useVideoStore()

// ---- 正在流行 ----
const trendingTopics = ref<TopicItem[]>([])
const loading = ref(true)

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
    trendingTopics.value = []
  } finally {
    loading.value = false
  }
}

// ---- 为你推荐 ----
const suggestedUsers = ref<RecommendUser[]>([])
const suggestLoading = ref(true)
const followingSet = ref(new Set<string>())

const fmtFollowers = (n: number): string => {
  if (n >= 100_000_000) return `${(n / 100_000_000).toFixed(1)}亿粉丝`
  if (n >= 10_000) return `${(n / 10_000).toFixed(0)}万粉丝`
  return `${n} 粉丝`
}

const loadSuggestions = async () => {
  suggestLoading.value = true
  try {
    const res = await getRecommendUsers(5)
    suggestedUsers.value = res.data.data?.users ?? []
  } catch {
    suggestedUsers.value = []
  } finally {
    suggestLoading.value = false
  }
}

const toggleFollow = async (userId: string) => {
  if (!userStore.isLoggedIn) {
    authModal.requireLogin()
    return
  }
  try {
    if (followingSet.value.has(userId)) {
      await unfollowUser(userId)
      followingSet.value = new Set([...followingSet.value].filter((id) => id !== userId))
      // 同步到自己的关注数
      if (userStore.profile) {
        userStore.profile.following = Math.max(0, userStore.profile.following - 1)
      }
      // 同步视频流里的关注按钮状态
      videoStore.syncAuthorFollowed(userId, false)
      // 取关后重新加载推荐列表（让被取关的用户有机会重新出现）
      void loadSuggestions()
    } else {
      await followUser(userId)
      followingSet.value = new Set([...followingSet.value, userId])
      // 同步到自己的关注数
      if (userStore.profile) {
        userStore.profile.following = userStore.profile.following + 1
      }
      // 同步视频流里的关注按钮状态
      videoStore.syncAuthorFollowed(userId, true)
      // 关注成功后从推荐列表移除该用户，并补充新推荐
      suggestedUsers.value = suggestedUsers.value.filter((u) => u.userId !== userId)
      // 补充一个新推荐用户
      try {
        const res = await getRecommendUsers(1)
        const newUsers = res.data.data?.users ?? []
        if (newUsers.length > 0) {
          suggestedUsers.value.push(newUsers[0])
        }
      } catch {
        // 静默失败，不影响关注操作
      }
    }
  } catch {
    showToast('操作失败，请重试')
  }
}

const refresh = () => {
  loadTopics()
  loadSuggestions()
}

// 监听视频流里的关注/取关操作，及时同步推荐列表
watch(
  () => videoStore.lastFollowChange,
  (change) => {
    if (!change) return
    if (!change.followed) {
      // 取关：重新拉推荐列表（让被取关的用户有机会重新出现）
      void loadSuggestions()
      return
    }
    // 关注：从推荐列表移除，并补充一个新推荐
    const idx = suggestedUsers.value.findIndex((u) => u.userId === change.authorId)
    if (idx === -1) return
    suggestedUsers.value.splice(idx, 1)
    getRecommendUsers(1)
      .then((res) => {
        const newUsers = res.data.data?.users ?? []
        if (newUsers.length > 0) suggestedUsers.value.push(newUsers[0])
      })
      .catch(() => {})
  },
)

onMounted(() => {
  loadTopics()
  loadSuggestions()
})
</script>

<template>
  <aside
    class="w-[288px] shrink-0 h-full overflow-y-auto bg-[#0d0d10] px-4 py-4 space-y-3 scrollbar-none"
  >
    <!-- ① 正在流行 -->
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

      <!-- 空状态 -->
      <template v-else-if="trendingTopics.length === 0">
        <p class="text-xs text-gray-600 text-center py-3">暂无热门话题</p>
      </template>

      <!-- 列表 -->
      <template v-else>
        <button
          v-for="(topic, i) in trendingTopics"
          :key="topic.id"
          type="button"
          class="w-full flex items-center gap-3 py-2 px-1 rounded-lg hover:bg-white/[0.04] transition-colors"
          @click="router.push({ name: 'topic', params: { id: topic.id } })"
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

    <!-- ② 为你推荐 -->
    <div class="bg-[#111115] rounded-2xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white">为你推荐</h3>
        <button
          type="button"
          class="w-7 h-7 flex items-center justify-center rounded-full hover:bg-white/10 text-gray-500 hover:text-white transition-colors"
          @click="refresh"
        >
          <i class="fa-solid fa-rotate-right text-xs" />
        </button>
      </div>

      <!-- Skeleton -->
      <template v-if="suggestLoading">
        <div v-for="i in 5" :key="i" class="flex items-center gap-3 py-1.5">
          <div class="w-9 h-9 bg-white/5 rounded-full animate-pulse shrink-0" />
          <div class="flex-1 space-y-1">
            <div class="h-3 bg-white/5 rounded animate-pulse w-2/3" />
            <div class="h-2 bg-white/5 rounded animate-pulse w-1/3" />
          </div>
          <div class="w-12 h-7 bg-white/5 rounded-lg animate-pulse shrink-0" />
        </div>
      </template>

      <!-- 空状态 -->
      <template v-else-if="suggestedUsers.length === 0">
        <p class="text-xs text-gray-600 text-center py-3">暂无推荐用户</p>
      </template>

      <!-- 列表 -->
      <template v-else>
        <div
          v-for="u in suggestedUsers"
          :key="u.userId"
          class="flex items-center gap-3 py-1.5"
        >
          <img
            :src="u.avatar || `https://picsum.photos/40/40?random=${u.userId}`"
            :alt="u.nickname"
            class="w-9 h-9 rounded-full object-cover shrink-0"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm text-white font-medium truncate">{{ u.nickname }}</p>
            <p class="text-xs text-gray-500">{{ fmtFollowers(u.followerCount) }}</p>
          </div>
          <button
            type="button"
            class="shrink-0 text-xs px-3 py-1.5 rounded-lg border transition-all"
            :class="
              followingSet.has(u.userId)
                ? 'bg-violet-600/20 border-violet-500/40 text-violet-300 hover:bg-red-500/20 hover:border-red-500/40 hover:text-red-300'
                : 'bg-white/[0.06] border-white/[0.1] text-gray-300 hover:bg-white/[0.12] hover:text-white'
            "
            @click="toggleFollow(u.userId)"
          >
            {{ followingSet.has(u.userId) ? '已关注' : '关注' }}
          </button>
        </div>
      </template>
    </div>
  </aside>
</template>
