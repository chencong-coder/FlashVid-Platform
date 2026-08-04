<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import { getMyPlaylists, type PlaylistInfo } from '@/api/user'

const route = useRoute()
const router = useRouter()
const authModal = useAuthModalStore()
const userStore = useUserStore()

interface NavItem {
  label: string
  icon: string
  name: string
  query?: Record<string, string>
  badge?: number
}

const navItems: NavItem[] = [
  { label: '推荐', icon: 'fa-house', name: 'recommend' },
  { label: '发现', icon: 'fa-compass', name: 'discover' },
  { label: '关注', icon: 'fa-user-group', name: 'follow' },
  { label: '朋友', icon: 'fa-users', name: 'friends' },
  { label: '消息', icon: 'fa-message', name: 'messages' },
  { label: '我的', icon: 'fa-user', name: 'profile' },
]

// ---- 我的播放列表 ----
const playlists = ref<PlaylistInfo[]>([])
const playlistsLoading = ref(false)

const loadPlaylists = async () => {
  if (!userStore.isLoggedIn) return
  playlistsLoading.value = true
  try {
    const res = await getMyPlaylists()
    playlists.value = res.data.data.playlists
  } catch {
    playlists.value = []
  } finally {
    playlistsLoading.value = false
  }
}

// 登录态变化时重新加载
watch(
  () => userStore.isLoggedIn,
  (loggedIn) => {
    if (loggedIn) {
      loadPlaylists()
    } else {
      playlists.value = []
    }
  },
)

onMounted(loadPlaylists)

// ---- 导航 ----
const isActive = (item: NavItem): boolean => {
  if (item.name === 'recommend') return ['recommend', 'nearby'].includes(route.name as string)
  return route.name === item.name
}

const navigate = (item: NavItem) => router.push({ name: item.name, query: item.query })

const onPlaylistClick = () => {
  if (!userStore.isLoggedIn) {
    authModal.requireLogin()
    return
  }
  router.push({ name: 'profile' })
}

const onCreatePlaylist = () => {
  if (!userStore.isLoggedIn) {
    authModal.requireLogin()
    return
  }
  router.push({ name: 'profile' })
}
</script>

<template>
  <aside class="flex flex-col w-60 shrink-0 h-full bg-[#111115] overflow-y-auto scrollbar-none">
    <!-- Logo -->
    <div class="px-5 pt-6 pb-5 shrink-0">
      <span
        class="text-2xl font-black tracking-tight bg-gradient-to-r from-violet-400 via-purple-300 to-fuchsia-300 bg-clip-text text-transparent"
      >
        闪视
      </span>
    </div>

    <!-- Navigation -->
    <nav class="px-3 space-y-0.5 shrink-0">
      <button
        v-for="item in navItems"
        :key="item.label"
        type="button"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-150"
        :class="
          isActive(item)
            ? 'bg-[#3730a3] text-white'
            : 'text-gray-400 hover:bg-white/5 hover:text-white'
        "
        @click="navigate(item)"
      >
        <i class="fa-solid w-5 text-center text-[15px]" :class="item.icon" />
        <span class="flex-1 text-left">{{ item.label }}</span>
        <span
          v-if="item.badge"
          class="text-[10px] font-bold bg-blue-600 text-white rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1"
          >{{ item.badge }}</span
        >
      </button>
    </nav>

    <!-- Divider -->
    <div class="mx-4 my-4 border-t border-white/[0.07] shrink-0" />

    <!-- 我的播放列表 -->
    <div class="px-4 shrink-0">
      <p class="text-[11px] font-semibold text-gray-500 uppercase tracking-widest mb-3 px-1">
        我的播放列表
      </p>

      <!-- 未登录提示 -->
      <template v-if="!userStore.isLoggedIn">
        <p class="text-xs text-gray-600 px-2 py-1">登录后查看播放列表</p>
      </template>

      <!-- 加载骨架 -->
      <template v-else-if="playlistsLoading">
        <div v-for="i in 3" :key="i" class="flex items-center gap-3 px-2 py-2">
          <div class="w-10 h-10 bg-white/5 rounded-lg animate-pulse shrink-0" />
          <div class="flex-1 space-y-1.5">
            <div class="h-3 bg-white/5 rounded animate-pulse w-3/4" />
            <div class="h-2 bg-white/5 rounded animate-pulse w-1/3" />
          </div>
        </div>
      </template>

      <!-- 空状态 -->
      <template v-else-if="playlists.length === 0">
        <p class="text-xs text-gray-600 px-2 py-1">暂无播放列表</p>
      </template>

      <!-- 列表 -->
      <template v-else>
        <div class="space-y-0.5">
          <button
            v-for="pl in playlists"
            :key="pl.id"
            type="button"
            class="w-full flex items-center gap-3 px-2 py-2 rounded-xl hover:bg-white/5 transition-colors"
            @click="onPlaylistClick"
          >
            <img
              :src="pl.coverUrl || `https://picsum.photos/40/40?random=${pl.id}`"
              :alt="pl.title"
              class="w-10 h-10 rounded-lg object-cover shrink-0"
            />
            <div class="min-w-0 text-left">
              <p class="text-sm text-white font-medium truncate">{{ pl.title }}</p>
              <p class="text-xs text-gray-500">{{ pl.videoCount }} 个视频</p>
            </div>
          </button>
        </div>
      </template>
    </div>

    <div class="flex-1" />

    <!-- 创建播放列表 -->
    <div class="px-4 pb-4 shrink-0">
      <button
        type="button"
        class="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-white/[0.06] border border-white/[0.08] text-sm text-gray-400 hover:bg-white/10 hover:text-white transition-all"
        @click="onCreatePlaylist"
      >
        <i class="fa-solid fa-plus text-xs" />
        创建播放列表
      </button>
    </div>

    <!-- Footer links -->
    <div class="px-5 pb-5 shrink-0">
      <div class="flex flex-wrap gap-x-2.5 gap-y-1 text-[11px] text-gray-600">
        <span
          v-for="link in ['关于', '招聘', '帮助', '条款', '隐私']"
          :key="link"
          class="hover:text-gray-400 cursor-pointer transition-colors"
          >{{ link }}</span
        >
      </div>
      <p class="text-[10px] text-gray-700 mt-1">© 2024 闪视</p>
    </div>
  </aside>
</template>
