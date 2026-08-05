<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import type { FeedVideo } from '@/api/feed'
import type { UserInfo } from '@/api/user'
import { followUser, getUserInfo, getUserVideos, unfollowUser } from '@/api/user'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import { formatCount } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const authModal = useAuthModalStore()

const userId = computed(() => String(route.params.id))
const userInfo = ref<UserInfo | null>(null)
const videos = ref<FeedVideo[]>([])
const loading = ref(true)
const videosLoading = ref(false)
const following = ref(false)
const followLoading = ref(false)

const isSelf = computed(
  () => userStore.isLoggedIn && userStore.profile?.id === userId.value,
)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getUserInfo(userId.value)
    userInfo.value = res.data.data
    // 初始化关注状态（后端返回 isFollowing 字段）
    following.value = res.data.data.isFollowing ?? false
  } catch {
    showToast('用户不存在')
    void router.back()
    return
  } finally {
    loading.value = false
  }
  void loadVideos()
})

const loadVideos = async () => {
  videosLoading.value = true
  try {
    const res = await getUserVideos(userId.value, { page: 1, pageSize: 20 })
    videos.value = res.data.data.videos
  } catch {
    /* silently keep empty */
  } finally {
    videosLoading.value = false
  }
}

const toggleFollow = async () => {
  if (!authModal.requireLogin()) return
  followLoading.value = true
  try {
    if (following.value) {
      await unfollowUser(userId.value)
      following.value = false
      if (userInfo.value) userInfo.value.followersCount -= 1
    } else {
      await followUser(userId.value)
      following.value = true
      if (userInfo.value) userInfo.value.followersCount += 1
    }
  } catch {
    showToast('操作失败，请重试')
  } finally {
    followLoading.value = false
  }
}
</script>

<template>
  <main class="no-scrollbar relative h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <!-- 返回按钮 -->
    <button
      type="button"
      aria-label="返回"
      class="safe-top absolute left-3 top-3 z-40 flex h-9 w-9 items-center justify-center rounded-full bg-black/40 text-white backdrop-blur-sm transition-colors hover:bg-white/10"
      @click="router.back()"
    >
      <i class="fa-solid fa-chevron-left text-sm" />
    </button>

    <!-- 骨架屏 -->
    <template v-if="loading">
      <div class="h-44 animate-pulse bg-neutral-800" />
      <div class="relative -mt-11 px-4">
        <div class="h-[5.5rem] w-[5.5rem] animate-pulse rounded-full bg-neutral-700" />
        <div class="mt-3 h-5 w-32 animate-pulse rounded bg-neutral-700" />
        <div class="mt-2 h-3 w-48 animate-pulse rounded bg-neutral-700" />
      </div>
    </template>

    <template v-else-if="userInfo">
      <!-- 背景封面 -->
      <div class="relative h-44 bg-neutral-900">
        <img
          src="https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?q=80&w=1080&auto=format&fit=crop"
          alt="背景"
          class="h-full w-full object-cover opacity-50"
        />
        <div class="absolute inset-0 bg-gradient-to-b from-black/10 to-[#0d0d0d]" />
      </div>

      <!-- 用户信息 -->
      <section class="relative -mt-11 px-4">
        <img
          :src="userInfo.avatar"
          :alt="userInfo.nickname"
          class="h-[5.5rem] w-[5.5rem] rounded-full border-4 border-[#0d0d0d] object-cover"
        />
        <div class="mt-3 flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h1 class="text-xl font-bold truncate">{{ userInfo.nickname || userInfo.username }}</h1>
            <p class="mt-0.5 text-xs text-neutral-500">@{{ userInfo.username }}</p>
          </div>
          <!-- 自己的主页不显示关注按钮 -->
          <button
            v-if="!isSelf"
            type="button"
            :disabled="followLoading"
            class="mt-1 shrink-0 rounded-lg px-5 py-2 text-sm font-semibold transition-all disabled:opacity-50"
            :class="
              following
                ? 'border border-white/20 bg-white/10 text-white hover:bg-white/15'
                : 'bg-primary text-white hover:bg-[#ff3e63]'
            "
            @click="toggleFollow"
          >
            {{ following ? '已关注' : '关注' }}
          </button>
          <button
            v-else
            type="button"
            class="mt-1 shrink-0 rounded-lg border border-white/20 bg-white/10 px-5 py-2 text-sm font-semibold text-white hover:bg-white/15"
            @click="router.push({ name: 'profile' })"
          >
            编辑资料
          </button>
        </div>
        <p class="mt-3 text-sm text-neutral-300">{{ userInfo.bio || '暂无简介' }}</p>
        <div class="mt-4 flex gap-6 text-sm">
          <span><b>{{ userInfo.followingCount }}</b> <i class="not-italic text-neutral-500">关注</i></span>
          <span><b>{{ formatCount(userInfo.followersCount) }}</b> <i class="not-italic text-neutral-500">粉丝</i></span>
          <span><b>{{ formatCount(userInfo.likesCount) }}</b> <i class="not-italic text-neutral-500">获赞</i></span>
        </div>
      </section>

      <!-- 作品标题 -->
      <div class="mt-5 border-b border-white/5 px-4 pb-3">
        <span class="text-sm font-medium text-white">作品</span>
        <span class="ml-2 text-xs text-neutral-500">{{ userInfo.videosCount }}</span>
      </div>

      <!-- 视频网格 -->
      <div v-if="videosLoading" class="grid grid-cols-3 gap-0.5 mt-0.5">
        <div v-for="n in 9" :key="n" class="aspect-[3/4] animate-pulse bg-neutral-800" />
      </div>
      <div v-else-if="videos.length" class="grid grid-cols-3 gap-0.5 mt-0.5">
        <div
          v-for="item in videos"
          :key="item.id"
          class="relative aspect-[3/4] overflow-hidden bg-neutral-900"
        >
          <img
            :src="item.coverUrl"
            alt="视频封面"
            loading="lazy"
            class="h-full w-full object-cover"
          />
          <span class="absolute bottom-1 left-1 text-[10px] text-white">
            <i class="fa-solid fa-play mr-1" />{{ formatCount(item.stats.likeCount) }}
          </span>
        </div>
      </div>
      <div v-else class="flex flex-col items-center justify-center py-16 text-neutral-500">
        <i class="fa-regular fa-video mb-3 text-3xl" />
        <p class="text-sm">暂无作品</p>
      </div>
    </template>
  </main>
</template>
