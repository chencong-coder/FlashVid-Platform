<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import { getMyPlaylists, createPlaylist, type PlaylistInfo } from '@/api/user'

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

watch(
  () => userStore.isLoggedIn,
  (loggedIn) => {
    if (loggedIn) loadPlaylists()
    else playlists.value = []
  },
)

onMounted(loadPlaylists)

// ---- 创建播放列表弹窗 ----
const showCreateModal = ref(false)
const createTitle = ref('')
const createDesc = ref('')
const creating = ref(false)

const openCreateModal = () => {
  if (!userStore.isLoggedIn) {
    authModal.requireLogin()
    return
  }
  createTitle.value = ''
  createDesc.value = ''
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const submitCreate = async () => {
  const title = createTitle.value.trim()
  if (!title) {
    showToast('请输入播放列表名称')
    return
  }
  creating.value = true
  try {
    await createPlaylist({ title, description: createDesc.value.trim() || undefined })
    showToast('创建成功')
    showCreateModal.value = false
    await loadPlaylists()
  } catch {
    showToast('创建失败，请重试')
  } finally {
    creating.value = false
  }
}

// ---- 导航 ----
const isActive = (item: NavItem): boolean => {
  if (item.name === 'recommend') return ['recommend', 'nearby'].includes(route.name as string)
  return route.name === item.name
}

const navigate = (item: NavItem) => router.push({ name: item.name, query: item.query })
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
            @click="router.push({ name: 'playlist-detail', params: { id: pl.id } })"
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
        @click="openCreateModal"
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
      <p class="text-[10px] text-gray-700 mt-1">© 2026 闪视</p>
    </div>
    <!-- 创建播放列表弹窗（Teleport 至 body，放在 aside 内不影响布局） -->
    <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-100 ease-in"
      leave-to-class="opacity-0"
    >
      <div
        v-if="showCreateModal"
        class="fixed inset-0 z-50 flex items-center justify-center px-4"
        @click.self="closeCreateModal"
      >
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="closeCreateModal" />
        <div class="relative z-10 w-full max-w-sm rounded-2xl bg-[#1a1a22] shadow-2xl" @click.stop>
          <!-- 标题栏 -->
          <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-4">
            <span class="text-sm font-semibold text-white">新建播放列表</span>
            <button
              type="button"
              class="flex h-7 w-7 items-center justify-center rounded-full text-gray-500 hover:bg-white/10 hover:text-white"
              @click="closeCreateModal"
            >
              <i class="fa-solid fa-xmark text-xs" />
            </button>
          </div>
          <!-- 表单 -->
          <div class="space-y-3 px-5 py-4">
            <div>
              <label class="mb-1.5 block text-xs text-gray-400">名称 <span class="text-primary">*</span></label>
              <input
                v-model="createTitle"
                type="text"
                maxlength="50"
                placeholder="给播放列表起个名字"
                class="w-full rounded-lg bg-white/[0.06] px-3 py-2.5 text-sm text-white outline-none placeholder:text-gray-600 focus:ring-1 focus:ring-violet-500/50"
                @keydown.enter="submitCreate"
              />
            </div>
            <div>
              <label class="mb-1.5 block text-xs text-gray-400">描述（选填）</label>
              <textarea
                v-model="createDesc"
                rows="2"
                maxlength="200"
                placeholder="描述一下这个播放列表…"
                class="w-full resize-none rounded-lg bg-white/[0.06] px-3 py-2.5 text-sm text-white outline-none placeholder:text-gray-600 focus:ring-1 focus:ring-violet-500/50"
              />
            </div>
          </div>
          <!-- 操作按钮 -->
          <div class="flex gap-2 border-t border-white/[0.06] px-5 py-4">
            <button
              type="button"
              class="flex-1 rounded-lg border border-white/10 py-2.5 text-sm text-gray-400 transition-colors hover:bg-white/5 hover:text-white"
              @click="closeCreateModal"
            >
              取消
            </button>
            <button
              type="button"
              :disabled="creating || !createTitle.trim()"
              class="flex-1 rounded-lg bg-violet-600 py-2.5 text-sm font-medium text-white transition-colors hover:bg-violet-500 disabled:cursor-not-allowed disabled:opacity-40"
              @click="submitCreate"
            >
              {{ creating ? '创建中…' : '创建' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</aside>
</template>
