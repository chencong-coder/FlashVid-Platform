<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import type { FeedVideo } from '@/api/feed'
import type { UserInfo } from '@/api/user'
import { getMyFavorites, getMyLikes, getUserInfo, getUserVideos } from '@/api/user'
import { useUserStore } from '@/store/user'
import { formatCount } from '@/utils/format'

const route = useRoute()
const userStore = useUserStore()

type ProfileTab = 'works' | 'likes' | 'favorites'
const validTabs: ProfileTab[] = ['works', 'likes', 'favorites']
const activeTab = ref<ProfileTab>(
  validTabs.includes(route.query.tab as ProfileTab) ? (route.query.tab as ProfileTab) : 'works',
)

const userInfo = ref<UserInfo | null>(null)
const worksVideos = ref<FeedVideo[]>([])
const likesVideos = ref<FeedVideo[]>([])
const favoritesVideos = ref<FeedVideo[]>([])
const loadedTabs = ref<Set<ProfileTab>>(new Set())
const tabLoading = ref(false)

// Merges fresh API data with cached store profile as fallback
const profile = computed(() => {
  if (userInfo.value) {
    return {
      nickname: userInfo.value.nickname || userInfo.value.username,
      avatar: userInfo.value.avatar,
      bio: userInfo.value.bio || '',
      followingCount: userInfo.value.followingCount,
      followersCount: userInfo.value.followersCount,
      likesCount: userInfo.value.likesCount,
    }
  }
  if (userStore.profile) {
    return {
      nickname: userStore.profile.nickname,
      avatar: userStore.profile.avatar,
      bio: userStore.profile.bio || '',
      followingCount: userStore.profile.following,
      followersCount: userStore.profile.followers,
      likesCount: userStore.profile.likes,
    }
  }
  return {
    nickname: '登录后发现更多精彩',
    avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?q=80&w=240&auto=format&fit=crop',
    bio: '点击编辑个人简介',
    followingCount: 0,
    followersCount: 0,
    likesCount: 0,
  }
})

const gridItems = computed<FeedVideo[]>(() => {
  if (activeTab.value === 'likes') return likesVideos.value
  if (activeTab.value === 'favorites') return favoritesVideos.value
  return worksVideos.value
})

const loadTab = async (tab: ProfileTab): Promise<void> => {
  if (loadedTabs.value.has(tab) || !userStore.profile) return
  tabLoading.value = true
  try {
    if (tab === 'works') {
      const res = await getUserVideos(userStore.profile.id, { page: 1, pageSize: 20 })
      worksVideos.value = res.data.data.videos
    } else if (tab === 'likes') {
      const res = await getMyLikes({ page: 1, pageSize: 20 })
      likesVideos.value = res.data.data.videos
    } else {
      const res = await getMyFavorites({ page: 1, pageSize: 20 })
      favoritesVideos.value = res.data.data.videos
    }
    loadedTabs.value.add(tab)
  } catch { /* silently keep empty */ } finally {
    tabLoading.value = false
  }
}

onMounted(async () => {
  if (!userStore.isLoggedIn || !userStore.profile) return
  try {
    const res = await getUserInfo(userStore.profile.id)
    userInfo.value = res.data.data
  } catch { /* keep store cache */ }
  void loadTab(activeTab.value)
})

watch(activeTab, (tab) => { void loadTab(tab) })
</script>

<template>
  <main class="no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <div class="relative h-44 bg-neutral-900">
      <img
        src="https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?q=80&w=1080&auto=format&fit=crop"
        alt="主页背景"
        class="h-full w-full object-cover opacity-60"
      />
      <div class="absolute inset-0 bg-gradient-to-b from-black/20 to-[#0d0d0d]" />
      <div class="safe-top absolute inset-x-0 top-0 flex justify-end gap-5 px-4 pt-4 text-lg">
        <button aria-label="添加朋友"><i class="fa-solid fa-user-plus" /></button
        ><button aria-label="更多"><i class="fa-solid fa-ellipsis" /></button>
      </div>
    </div>
    <section class="relative -mt-11 px-4">
      <img
        :src="profile.avatar"
        :alt="profile.nickname"
        class="h-[5.5rem] w-[5.5rem] rounded-full border-4 border-[#0d0d0d] object-cover"
      />
      <h1 class="mt-3 text-xl font-bold">{{ profile.nickname }}</h1>
      <p class="mt-1 text-xs text-neutral-500">闪视号：flashvid_2026</p>
      <p class="mt-3 text-sm text-neutral-300">{{ profile.bio || '暂无简介' }}</p>
      <div class="mt-4 flex gap-6 text-sm">
        <span><b>{{ profile.followingCount }}</b> <i class="not-italic text-neutral-500">关注</i></span
        ><span><b>{{ formatCount(profile.followersCount) }}</b> <i class="not-italic text-neutral-500">粉丝</i></span
        ><span><b>{{ formatCount(profile.likesCount) }}</b> <i class="not-italic text-neutral-500">获赞</i></span>
      </div>
      <div class="mt-4 flex gap-2">
        <button class="h-9 flex-1 rounded bg-white/10 text-sm font-medium">编辑资料</button
        ><button aria-label="分享主页" class="h-9 w-11 rounded bg-white/10">
          <i class="fa-solid fa-share" />
        </button>
      </div>
    </section>
    <section class="mt-5">
      <div class="grid h-12 grid-cols-3 border-b border-white/5 text-sm">
        <button
          type="button"
          class="relative"
          :class="activeTab === 'works' ? 'text-white' : 'text-neutral-500'"
          @click="activeTab = 'works'"
        >
          作品<span
            v-if="activeTab === 'works'"
            class="absolute bottom-0 left-1/2 h-0.5 w-6 -translate-x-1/2 bg-white"
          /></button
        ><button
          type="button"
          class="relative"
          :class="activeTab === 'likes' ? 'text-white' : 'text-neutral-500'"
          @click="activeTab = 'likes'"
        >
          喜欢<span
            v-if="activeTab === 'likes'"
            class="absolute bottom-0 left-1/2 h-0.5 w-6 -translate-x-1/2 bg-white"
          /></button
        ><button
          type="button"
          class="relative"
          :class="activeTab === 'favorites' ? 'text-white' : 'text-neutral-500'"
          @click="activeTab = 'favorites'"
        >
          收藏<span
            v-if="activeTab === 'favorites'"
            class="absolute bottom-0 left-1/2 h-0.5 w-6 -translate-x-1/2 bg-white"
          />
        </button>
      </div>
      <!-- Video grid -->
      <div v-if="tabLoading && gridItems.length === 0" class="grid grid-cols-3 gap-0.5">
        <div
          v-for="n in 9"
          :key="n"
          class="aspect-[3/4] animate-pulse bg-neutral-800"
        />
      </div>
      <div v-else-if="gridItems.length" class="grid grid-cols-3 gap-0.5">
        <div
          v-for="item in gridItems"
          :key="item.id"
          class="relative aspect-[3/4] overflow-hidden bg-neutral-900"
        >
          <img :src="item.coverUrl" alt="视频封面" loading="lazy" class="h-full w-full object-cover" />
          <span class="absolute bottom-1 left-1 text-[10px] text-white">
            <i class="fa-solid fa-play mr-1" />{{ formatCount(item.stats.likeCount) }}
          </span>
        </div>
      </div>
      <div v-else class="flex flex-col items-center justify-center py-16 text-neutral-500">
        <i class="fa-regular fa-folder-open mb-3 text-3xl" />
        <p class="text-sm">暂无内容</p>
      </div>
    </section>
  </main>
</template>
