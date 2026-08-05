<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'

import type { FeedVideo } from '@/api/feed'
import type { UserInfo, UpdateUserPayload, PlaylistInfo } from '@/api/user'
import {
  getMyLikes,
  getMyFavorites,
  getMyPlaylists,
  getPlaylistVideos,
  getUserInfo,
  getUserVideos,
  updateUserInfo,
} from '@/api/user'
import { useUserStore } from '@/store/user'
import { useAuthModalStore } from '@/store/authModal'
import { formatCount } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const authModal = useAuthModalStore()

type ProfileTab = 'works' | 'likes' | 'favorites'
const validTabs: ProfileTab[] = ['works', 'likes', 'favorites']
const activeTab = ref<ProfileTab>(
  validTabs.includes(route.query.tab as ProfileTab) ? (route.query.tab as ProfileTab) : 'works',
)

const userInfo = ref<UserInfo | null>(null)
const worksVideos = ref<FeedVideo[]>([])
const likesVideos = ref<FeedVideo[]>([])
const favoritesVideos = ref<FeedVideo[]>([])
const playlists = ref<PlaylistInfo[]>([])
const selectedPlaylist = ref<PlaylistInfo | null>(null)
const playlistVideos = ref<FeedVideo[]>([])
const playlistVideosLoading = ref(false)
const tabLoading = ref(false)

// 编辑资料弹窗
const showEditModal = ref(false)
const editLoading = ref(false)
const editForm = ref<UpdateUserPayload>({
  nickname: '',
  bio: '',
  city: '',
  gender: 0,
  birthday: '',
  email: '',
})

const openEditModal = (): void => {
  const p = profile.value
  editForm.value = {
    nickname: p.nickname,
    bio: p.bio,
    city: userInfo.value?.city ?? '',
    gender: userInfo.value?.gender ?? 0,
    birthday: userInfo.value?.birthday ?? '',
    email: userInfo.value?.email ?? '',
  }
  showEditModal.value = true
}

const saveProfile = async (): Promise<void> => {
  if (!editForm.value.nickname?.trim()) {
    showToast('昵称不能为空')
    return
  }
  editLoading.value = true
  try {
    const payload: UpdateUserPayload = {}
    if (editForm.value.nickname) payload.nickname = editForm.value.nickname.trim()
    if (editForm.value.bio !== undefined) payload.bio = editForm.value.bio
    if (editForm.value.city) payload.city = editForm.value.city
    if (editForm.value.gender !== undefined) payload.gender = editForm.value.gender
    if (editForm.value.birthday) payload.birthday = editForm.value.birthday
    if (editForm.value.email) payload.email = editForm.value.email

    const res = await updateUserInfo(payload)
    if (res.data.code === 0) {
      if (userStore.profile && payload.nickname) userStore.profile.nickname = payload.nickname
      if (userStore.profile && payload.bio !== undefined) userStore.profile.bio = payload.bio
      if (userInfo.value) userInfo.value = { ...userInfo.value, ...res.data.data }
      showToast('保存成功')
      showEditModal.value = false
    } else {
      showToast(res.data.message || '保存失败')
    }
  } catch {
    showToast('网络错误，请重试')
  } finally {
    editLoading.value = false
  }
}

const handleLogout = (): void => {
  userStore.logout()
  showToast('已退出登录')
  void router.push('/login')
}

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
    avatar:
      'https://images.unsplash.com/photo-1534528741775-53994a69daeb?q=80&w=240&auto=format&fit=crop',
    bio: '',
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
  if (!userStore.profile) return
  tabLoading.value = true
  try {
    if (tab === 'works') {
      const res = await getUserVideos(userStore.profile.id, { page: 1, pageSize: 20 })
      worksVideos.value = res.data.data.videos
    } else if (tab === 'likes') {
      const res = await getMyLikes({ page: 1, pageSize: 20 })
      likesVideos.value = res.data.data.videos
    } else if (tab === 'favorites') {
      const res = await getMyFavorites({ page: 1, pageSize: 20 })
      favoritesVideos.value = res.data.data.videos
    }
  } catch {
    /* silently keep empty */
  } finally {
    tabLoading.value = false
  }
}

onMounted(async () => {
  if (!userStore.isLoggedIn || !userStore.profile) return
  try {
    const res = await getUserInfo(userStore.profile.id)
    userInfo.value = res.data.data
  } catch {
    /* keep store cache */
  }
  void loadTab(activeTab.value)
})

watch(activeTab, (tab) => {
  selectedPlaylist.value = null
  playlistVideos.value = []
  void loadTab(tab)
})

const openPlaylist = async (playlist: PlaylistInfo): Promise<void> => {
  selectedPlaylist.value = playlist
  playlistVideos.value = []
  playlistVideosLoading.value = true
  try {
    const res = await getPlaylistVideos(playlist.id)
    if (res.data.code === 0) {
      playlistVideos.value = res.data.data.videos
    }
  } catch {
    /* silently fail */
  } finally {
    playlistVideosLoading.value = false
  }
}

const closePlaylist = (): void => {
  selectedPlaylist.value = null
  playlistVideos.value = []
}
</script>

<template>
  <main class="no-scrollbar relative h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <!-- 未登录：引导登录（与 LoginGate 保持一致的布局） -->
    <section
      v-if="!userStore.isLoggedIn"
      class="safe-top flex h-full flex-col items-center justify-center px-8 pb-28"
    >
      <div class="flex w-full max-w-[25rem] flex-col items-center text-center">
        <div
          class="flex h-[4.75rem] w-[4.75rem] items-center justify-center rounded-2xl border border-white/15 bg-primary shadow-[0_18px_45px_rgba(254,44,85,0.3)]"
        >
          <i class="fa-solid fa-play ml-1 text-[1.75rem] text-white" />
        </div>
        <h1 class="mt-5 text-[1.75rem] font-bold text-white">闪视</h1>
        <p class="mt-2 text-sm leading-6 text-neutral-400">记录精彩瞬间，发现更多可能</p>

        <button
          type="button"
          class="mt-8 flex h-[3.25rem] w-full items-center justify-center gap-2 rounded-lg bg-primary text-sm font-semibold text-white shadow-[0_12px_28px_rgba(254,44,85,0.28)] transition-all hover:bg-[#ff3e63] active:scale-[0.98]"
          @click="authModal.open('login')"
        >
          <i class="fa-solid fa-arrow-right-to-bracket text-xs" />
          登录
        </button>

        <div class="my-4 flex w-full items-center gap-3 text-[11px] text-neutral-600">
          <span class="h-px flex-1 bg-white/10" />
          <span>还没有账号？</span>
          <span class="h-px flex-1 bg-white/10" />
        </div>

        <button
          type="button"
          class="flex h-[3.25rem] w-full items-center justify-center gap-2 rounded-lg border border-white/15 bg-white/[0.03] text-sm font-medium text-neutral-200 transition-all hover:border-white/25 hover:bg-white/[0.07] active:scale-[0.98]"
          @click="authModal.open('register')"
        >
          <i class="fa-solid fa-user-plus text-xs text-neutral-400" />
          新用户注册
        </button>

        <p
          class="mt-6 flex flex-wrap items-center justify-center gap-x-1 text-[11px] leading-5 text-neutral-600"
        >
          <span>登录即表示同意</span>
          <button type="button" class="transition-colors hover:text-neutral-400">《用户协议》</button>
          <span>与</span>
          <button type="button" class="transition-colors hover:text-neutral-400">《隐私政策》</button>
        </p>
      </div>
    </section>

    <!-- 已登录：完整个人资料 -->
    <template v-else>
    <div class="relative h-44 bg-neutral-900">
      <img
        src="https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?q=80&w=1080&auto=format&fit=crop"
        alt="主页背景"
        class="h-full w-full object-cover opacity-60"
      />
      <div class="absolute inset-0 bg-gradient-to-b from-black/20 to-[#0d0d0d]" />
      <div class="safe-top absolute inset-x-0 top-0 z-30 flex justify-end gap-2 px-4 pt-4 text-lg">
        <button
          type="button"
          aria-label="退出登录"
          title="退出登录"
          class="flex h-9 w-9 items-center justify-center rounded-full bg-black/20 text-white/90 backdrop-blur-sm transition-colors hover:bg-white/10 hover:text-white"
          @click="handleLogout"
        >
          <i class="fa-solid fa-right-from-bracket" />
        </button>
        <button
          type="button"
          aria-label="更多"
          class="flex h-9 w-9 items-center justify-center rounded-full bg-black/20 text-white/90 backdrop-blur-sm transition-colors hover:bg-white/10 hover:text-white"
        >
          <i class="fa-solid fa-ellipsis text-sm" />
        </button>
      </div>
    </div>

    <!-- 已登录：个人资料 + 内容（接上方 v-else 的 template） -->
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
          <span
            ><b>{{ profile.followingCount }}</b>
            <i class="not-italic text-neutral-500">关注</i></span
          >
          <span
            ><b>{{ formatCount(profile.followersCount) }}</b>
            <i class="not-italic text-neutral-500">粉丝</i></span
          >
          <span
            ><b>{{ formatCount(profile.likesCount) }}</b>
            <i class="not-italic text-neutral-500">获赞</i></span
          >
        </div>
        <div class="mt-4 flex gap-2">
          <button class="h-9 flex-1 rounded bg-white/10 text-sm font-medium" @click="openEditModal">
            编辑资料
          </button>
          <button aria-label="分享主页" class="h-9 w-11 rounded bg-white/10">
            <i class="fa-solid fa-share" />
          </button>
        </div>
      </section>

      <!-- 视频 tab -->
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
            />
          </button>
          <button
            type="button"
            class="relative"
            :class="activeTab === 'likes' ? 'text-white' : 'text-neutral-500'"
            @click="activeTab = 'likes'"
          >
            喜欢<span
              v-if="activeTab === 'likes'"
              class="absolute bottom-0 left-1/2 h-0.5 w-6 -translate-x-1/2 bg-white"
            />
          </button>
          <button
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
        <!-- 作品 / 喜欢：3列视频网格 -->
        <template v-if="activeTab !== 'favorites'">
          <div v-if="tabLoading && !gridItems.length" class="grid grid-cols-3 gap-0.5">
            <div v-for="n in 9" :key="n" class="aspect-[3/4] animate-pulse bg-neutral-800" />
          </div>
          <div v-else-if="gridItems.length" class="grid grid-cols-3 gap-0.5">
            <div
              v-for="item in gridItems"
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
            <i class="fa-regular fa-folder-open mb-3 text-3xl" />
            <p class="text-sm">暂无内容</p>
          </div>
        </template>

        <!-- 收藏：收藏的视频网格 -->
        <template v-else-if="activeTab === 'favorites'">
          <div v-if="tabLoading && !gridItems.length" class="grid grid-cols-3 gap-0.5">
            <div v-for="n in 9" :key="n" class="aspect-[3/4] animate-pulse bg-neutral-800" />
          </div>
          <div v-else-if="gridItems.length" class="grid grid-cols-3 gap-0.5">
            <div
              v-for="item in gridItems"
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
            <i class="fa-regular fa-star mb-3 text-3xl" />
            <p class="text-sm">暂无收藏</p>
          </div>
        </template>

        <!-- 播放列表（已移除，收藏现在显示收藏的视频） -->
        <template v-else>
          <!-- 播放列表详情 -->
          <div v-if="selectedPlaylist">
            <div class="flex items-center gap-3 border-b border-white/5 px-4 py-3">
              <button
                type="button"
                class="flex h-7 w-7 items-center justify-center rounded-full bg-white/10"
                @click="closePlaylist"
              >
                <i class="fa-solid fa-chevron-left text-xs" />
              </button>
              <span class="text-sm font-medium">{{ selectedPlaylist.title }}</span>
              <span class="ml-auto text-xs text-neutral-500"
                >{{ selectedPlaylist.videoCount }} 个视频</span
              >
            </div>
            <div v-if="playlistVideosLoading" class="grid grid-cols-3 gap-0.5">
              <div v-for="n in 6" :key="n" class="aspect-[3/4] animate-pulse bg-neutral-800" />
            </div>
            <div v-else-if="playlistVideos.length" class="grid grid-cols-3 gap-0.5">
              <div
                v-for="item in playlistVideos"
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
              <i class="fa-regular fa-film mb-3 text-3xl" />
              <p class="text-sm">该列表暂无视频</p>
            </div>
          </div>

          <!-- 播放列表卡片网格 -->
          <div v-else>
            <div v-if="tabLoading && !playlists.length" class="grid grid-cols-2 gap-3 p-3">
              <div v-for="n in 4" :key="n" class="overflow-hidden rounded-lg bg-neutral-800">
                <div class="aspect-square animate-pulse bg-neutral-700" />
                <div class="space-y-1.5 px-2 py-2">
                  <div class="h-2.5 w-3/4 animate-pulse rounded bg-neutral-700" />
                  <div class="h-2 w-1/2 animate-pulse rounded bg-neutral-700" />
                </div>
              </div>
            </div>
            <div v-else-if="playlists.length" class="grid grid-cols-2 gap-3 p-3">
              <button
                v-for="pl in playlists"
                :key="pl.id"
                type="button"
                class="overflow-hidden rounded-lg bg-neutral-800 text-left transition-opacity active:opacity-70"
                @click="openPlaylist(pl)"
              >
                <div class="relative aspect-square bg-neutral-700">
                  <img
                    v-if="pl.coverUrl"
                    :src="pl.coverUrl"
                    alt="封面"
                    class="h-full w-full object-cover"
                  />
                  <div v-else class="flex h-full w-full items-center justify-center">
                    <i class="fa-solid fa-film text-2xl text-neutral-600" />
                  </div>
                </div>
                <div class="px-2 py-2">
                  <p class="truncate text-xs font-medium text-white">{{ pl.title }}</p>
                  <p class="mt-0.5 text-[10px] text-neutral-500">{{ pl.videoCount }} 个视频</p>
                </div>
              </button>
            </div>
            <div v-else class="flex flex-col items-center justify-center py-16 text-neutral-500">
              <i class="fa-regular fa-bookmark mb-3 text-3xl" />
              <p class="text-sm">暂无收藏列表</p>
            </div>
          </div>
        </template>
      </section>
    </template>

    <!-- 编辑资料弹窗（底部弹出） -->
    <transition name="fade">
      <div
        v-if="showEditModal"
        class="fixed inset-0 z-50 flex flex-col"
        @click.self="showEditModal = false"
      >
        <div class="flex-1 bg-black/60" @click="showEditModal = false" />
        <div
          class="no-scrollbar max-h-[82vh] overflow-y-auto rounded-t-2xl bg-[#1a1a1a] px-4 pb-10"
        >
          <div class="flex h-14 items-center justify-between text-white">
            <button type="button" class="text-sm text-neutral-400" @click="showEditModal = false">
              取消
            </button>
            <span class="text-sm font-semibold">编辑资料</span>
            <button
              type="button"
              :disabled="editLoading"
              class="text-sm font-medium text-primary disabled:opacity-50"
              @click="saveProfile"
            >
              {{ editLoading ? '保存中...' : '保存' }}
            </button>
          </div>
          <div class="space-y-3 text-white">
            <div class="rounded-lg bg-white/5 px-4 py-3">
              <label class="block text-xs text-neutral-400">昵称</label>
              <input
                v-model="editForm.nickname"
                type="text"
                maxlength="30"
                placeholder="请输入昵称"
                class="mt-1 w-full bg-transparent text-sm outline-none placeholder:text-neutral-600"
              />
            </div>
            <div class="rounded-lg bg-white/5 px-4 py-3">
              <label class="block text-xs text-neutral-400">简介</label>
              <textarea
                v-model="editForm.bio"
                rows="2"
                maxlength="100"
                placeholder="添加个人简介"
                class="mt-1 w-full resize-none bg-transparent text-sm outline-none placeholder:text-neutral-600"
              />
            </div>
            <div class="rounded-lg bg-white/5 px-4 py-3">
              <label class="block text-xs text-neutral-400">城市</label>
              <input
                v-model="editForm.city"
                type="text"
                maxlength="20"
                placeholder="所在城市"
                class="mt-1 w-full bg-transparent text-sm outline-none placeholder:text-neutral-600"
              />
            </div>
            <div class="rounded-lg bg-white/5 px-4 py-3">
              <label class="block text-xs text-neutral-400">邮箱</label>
              <input
                v-model="editForm.email"
                type="email"
                placeholder="邮箱（选填）"
                class="mt-1 w-full bg-transparent text-sm outline-none placeholder:text-neutral-600"
              />
            </div>
            <div class="rounded-lg bg-white/5 px-4 py-3">
              <label class="block text-xs text-neutral-400">生日</label>
              <input
                v-model="editForm.birthday"
                type="date"
                class="mt-1 w-full bg-transparent text-sm outline-none"
              />
            </div>
          </div>
        </div>
      </div>
    </transition>
  </main>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
