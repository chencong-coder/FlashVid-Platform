<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import {
  getPlaylistVideos,
  updatePlaylist,
  deletePlaylist,
  removeVideoFromPlaylist,
  type PlaylistInfo,
  type UpdatePlaylistPayload,
} from '@/api/user'
import { deleteVideo } from '@/api/video'
import type { FeedVideo } from '@/api/feed'

const route = useRoute()
const router = useRouter()

const playlistId = ref(route.params.id as string)

// ---- 播放列表信息 ----
const playlist = ref<PlaylistInfo | null>(null)
const videos = ref<FeedVideo[]>([])
const nextCursor = ref('')
const hasMore = ref(false)
const loading = ref(false)
const loadingMore = ref(false)

const loadPlaylist = async (reset = true) => {
  if (reset) {
    loading.value = true
    videos.value = []
    nextCursor.value = ''
    hasMore.value = false
  } else {
    loadingMore.value = true
  }
  try {
    const res = await getPlaylistVideos(playlistId.value, {
      cursor: reset ? undefined : nextCursor.value,
      count: 20,
    })
    const data = res.data.data
    playlist.value = data.playlist
    if (reset) {
      videos.value = data.videos ?? []
    } else {
      videos.value.push(...(data.videos ?? []))
    }
    nextCursor.value = data.nextCursor ?? ''
    hasMore.value = data.hasMore ?? false
  } catch {
    showToast('加载失败')
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

onMounted(() => loadPlaylist())
watch(() => route.params.id, (id) => {
  playlistId.value = id as string
  loadPlaylist()
})

// ---- 编辑播放列表弹窗 ----
const showEditModal = ref(false)
const editTitle = ref('')
const editDesc = ref('')
const saving = ref(false)

const openEditModal = () => {
  if (!playlist.value) return
  editTitle.value = playlist.value.title
  editDesc.value = playlist.value.description ?? ''
  showEditModal.value = true
}

const submitEdit = async () => {
  const title = editTitle.value.trim()
  if (!title) { showToast('名称不能为空'); return }
  saving.value = true
  try {
    const payload: UpdatePlaylistPayload = { title }
    if (editDesc.value.trim()) payload.description = editDesc.value.trim()
    await updatePlaylist(playlistId.value, payload)
    showToast('保存成功')
    showEditModal.value = false
    await loadPlaylist()
  } catch {
    showToast('保存失败，请重试')
  } finally {
    saving.value = false
  }
}

// ---- 删除播放列表 ----
const showDeleteConfirm = ref(false)
const deleting = ref(false)

const confirmDelete = async () => {
  deleting.value = true
  try {
    await deletePlaylist(playlistId.value)
    showToast('已删除')
    void router.back()
  } catch {
    showToast('删除失败，请重试')
  } finally {
    deleting.value = false
    showDeleteConfirm.value = false
  }
}

// ---- 管理模式 ----
const manageMode = ref(false)

// ---- 移除视频（从列表） ----
const removingId = ref<string | null>(null)

const removeVideo = async (videoId: string) => {
  removingId.value = videoId
  try {
    await removeVideoFromPlaylist(playlistId.value, videoId)
    videos.value = videos.value.filter(v => v.id !== videoId)
    if (playlist.value) playlist.value.videoCount = Math.max(0, playlist.value.videoCount - 1)
    showToast('已移出列表')
  } catch {
    showToast('移除失败')
  } finally {
    removingId.value = null
  }
}

// ---- 删除视频（彻底删除） ----
const deleteConfirmId = ref<string | null>(null)
const deletingVideoId = ref<string | null>(null)

const confirmDeleteVideo = async () => {
  if (!deleteConfirmId.value) return
  const videoId = deleteConfirmId.value
  deletingVideoId.value = videoId
  try {
    await deleteVideo(videoId)
    videos.value = videos.value.filter(v => v.id !== videoId)
    if (playlist.value) playlist.value.videoCount = Math.max(0, playlist.value.videoCount - 1)
    showToast('视频已删除')
    deleteConfirmId.value = null
  } catch {
    showToast('删除失败，请重试')
  } finally {
    deletingVideoId.value = null
  }
}

const formatDuration = (s: number): string => {
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${sec.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="min-h-screen bg-[#0d0d0d] text-white">
    <!-- 顶部导航 -->
    <div class="sticky top-0 z-10 flex items-center gap-3 bg-[#0d0d0d]/90 backdrop-blur px-4 py-3 border-b border-white/[0.06]">
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-full text-gray-400 hover:bg-white/10 hover:text-white transition-colors"
        @click="router.back()"
      >
        <i class="fa-solid fa-arrow-left text-sm" />
      </button>
      <h1 class="flex-1 truncate text-base font-semibold">{{ playlist?.title ?? '播放列表' }}</h1>

      <!-- 管理模式：仅显示"完成" -->
      <template v-if="manageMode">
        <button
          type="button"
          class="px-3 py-1 rounded-full bg-violet-600 text-xs font-medium text-white hover:bg-violet-500 transition-colors"
          @click="manageMode = false"
        >完成</button>
      </template>

      <!-- 普通模式：编辑 / 管理 / 删除列表 -->
      <template v-else-if="playlist">
        <button
          type="button"
          class="flex h-8 w-8 items-center justify-center rounded-full text-gray-400 hover:bg-white/10 hover:text-white transition-colors"
          title="编辑信息"
          @click="openEditModal"
        >
          <i class="fa-solid fa-pen text-xs" />
        </button>
        <button
          v-if="videos.length > 0"
          type="button"
          class="px-3 py-1 rounded-full bg-white/[0.07] text-xs text-gray-300 hover:bg-white/15 hover:text-white transition-colors"
          @click="manageMode = true"
        >管理</button>
        <button
          type="button"
          class="flex h-8 w-8 items-center justify-center rounded-full text-red-500/70 hover:bg-red-500/10 hover:text-red-400 transition-colors"
          title="删除列表"
          @click="showDeleteConfirm = true"
        >
          <i class="fa-solid fa-trash text-xs" />
        </button>
      </template>
    </div>

    <!-- 播放列表简介 -->
    <div v-if="playlist" class="flex items-start gap-4 px-5 py-5 border-b border-white/[0.05]">
      <img
        :src="playlist.coverUrl || `https://picsum.photos/80/80?random=${playlist.id}`"
        :alt="playlist.title"
        class="w-20 h-20 rounded-xl object-cover shrink-0"
      />
      <div class="min-w-0 flex-1 pt-1">
        <p class="text-lg font-bold leading-snug truncate">{{ playlist.title }}</p>
        <p v-if="playlist.description" class="mt-1 text-sm text-gray-400 line-clamp-2">{{ playlist.description }}</p>
        <p class="mt-2 text-xs text-gray-600">{{ playlist.videoCount }} 个视频</p>
      </div>
    </div>

    <!-- 骨架屏 -->
    <div v-if="loading" class="grid grid-cols-2 gap-3 p-4">
      <div v-for="i in 6" :key="i" class="rounded-xl bg-white/[0.04] overflow-hidden">
        <div class="aspect-[9/16] animate-pulse bg-white/[0.06]" />
        <div class="p-2 space-y-1.5">
          <div class="h-2.5 bg-white/[0.06] rounded animate-pulse w-3/4" />
          <div class="h-2 bg-white/[0.06] rounded animate-pulse w-1/3" />
        </div>
      </div>
    </div>

    <!-- 视频列表 -->
    <div v-else-if="videos.length > 0" class="p-4">
      <div class="grid grid-cols-2 gap-3">
        <div
          v-for="video in videos"
          :key="video.id"
          class="relative rounded-xl overflow-hidden bg-[#1a1a22]"
          :class="manageMode ? 'cursor-default' : 'group cursor-pointer'"
          @click="!manageMode && router.push({ name: 'user-profile', params: { id: video.author.id } })"
        >
          <!-- 封面 -->
          <div class="relative aspect-[9/16] bg-black">
            <img
              :src="video.coverUrl"
              :alt="video.title"
              class="w-full h-full object-cover"
              :class="manageMode ? 'brightness-50' : ''"
            />
            <!-- 时长（非管理模式） -->
            <span
              v-if="video.duration && !manageMode"
              class="absolute bottom-1.5 right-1.5 text-[10px] font-medium text-white bg-black/60 rounded px-1 py-0.5"
            >{{ formatDuration(video.duration) }}</span>
            <!-- 普通模式：悬浮移除按钮 -->
            <button
              v-if="!manageMode"
              type="button"
              class="absolute top-1.5 right-1.5 flex h-6 w-6 items-center justify-center rounded-full bg-black/60 text-white opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-500/80"
              :disabled="removingId === video.id"
              @click.stop="removeVideo(video.id)"
            >
              <i v-if="removingId === video.id" class="fa-solid fa-spinner fa-spin text-[9px]" />
              <i v-else class="fa-solid fa-xmark text-[9px]" />
            </button>
            <!-- 管理模式：操作覆盖层 -->
            <div v-if="manageMode" class="absolute inset-0 flex flex-col items-center justify-center gap-2 p-2">
              <button
                type="button"
                class="w-full flex items-center justify-center gap-1.5 rounded-lg bg-white/15 py-2 text-[11px] font-medium text-white backdrop-blur-sm hover:bg-white/25 transition-colors"
                :disabled="removingId === video.id"
                @click.stop="removeVideo(video.id)"
              >
                <i v-if="removingId === video.id" class="fa-solid fa-spinner fa-spin text-[10px]" />
                <i v-else class="fa-solid fa-minus-circle text-[10px] text-yellow-400" />
                移出列表
              </button>
              <button
                type="button"
                class="w-full flex items-center justify-center gap-1.5 rounded-lg bg-red-500/25 py-2 text-[11px] font-medium text-red-300 backdrop-blur-sm hover:bg-red-500/40 transition-colors"
                :disabled="deletingVideoId === video.id"
                @click.stop="deleteConfirmId = video.id"
              >
                <i v-if="deletingVideoId === video.id" class="fa-solid fa-spinner fa-spin text-[10px]" />
                <i v-else class="fa-solid fa-trash text-[10px]" />
                删除视频
              </button>
            </div>
          </div>
          <!-- 信息 -->
          <div class="p-2">
            <p class="text-xs text-white font-medium leading-tight line-clamp-2">{{ video.title }}</p>
            <p class="mt-1 text-[10px] text-gray-500">{{ video.author.nickname }}</p>
          </div>
        </div>
      </div>

      <!-- 加载更多 -->
      <div class="mt-4 flex justify-center">
        <button
          v-if="hasMore"
          type="button"
          :disabled="loadingMore"
          class="px-6 py-2 rounded-full bg-white/[0.06] text-sm text-gray-400 hover:bg-white/10 hover:text-white transition-colors disabled:opacity-50"
          @click="loadPlaylist(false)"
        >
          {{ loadingMore ? '加载中…' : '加载更多' }}
        </button>
        <p v-else class="text-xs text-gray-600 py-2">已加载全部</p>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!loading" class="flex flex-col items-center justify-center py-20 text-center px-8">
      <div class="w-16 h-16 rounded-full bg-white/[0.04] flex items-center justify-center mb-4">
        <i class="fa-solid fa-film text-2xl text-gray-600" />
      </div>
      <p class="text-sm font-medium text-gray-400">还没有视频</p>
      <p class="mt-1 text-xs text-gray-600">在视频页面点击"添加到播放列表"即可收录</p>
    </div>

    <!-- ===== 编辑弹窗 ===== -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        leave-active-class="transition duration-100 ease-in"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showEditModal"
          class="fixed inset-0 z-50 flex items-center justify-center px-4"
          @click.self="showEditModal = false"
        >
          <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="showEditModal = false" />
          <div class="relative z-10 w-full max-w-sm rounded-2xl bg-[#1a1a22] shadow-2xl" @click.stop>
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-4">
              <span class="text-sm font-semibold text-white">编辑播放列表</span>
              <button
                type="button"
                class="flex h-7 w-7 items-center justify-center rounded-full text-gray-500 hover:bg-white/10 hover:text-white"
                @click="showEditModal = false"
              >
                <i class="fa-solid fa-xmark text-xs" />
              </button>
            </div>
            <div class="space-y-3 px-5 py-4">
              <div>
                <label class="mb-1.5 block text-xs text-gray-400">名称 <span class="text-violet-400">*</span></label>
                <input
                  v-model="editTitle"
                  type="text"
                  maxlength="50"
                  placeholder="播放列表名称"
                  class="w-full rounded-lg bg-white/[0.06] px-3 py-2.5 text-sm text-white outline-none placeholder:text-gray-600 focus:ring-1 focus:ring-violet-500/50"
                  @keydown.enter="submitEdit"
                />
              </div>
              <div>
                <label class="mb-1.5 block text-xs text-gray-400">描述（选填）</label>
                <textarea
                  v-model="editDesc"
                  rows="3"
                  maxlength="200"
                  placeholder="描述一下这个播放列表…"
                  class="w-full resize-none rounded-lg bg-white/[0.06] px-3 py-2.5 text-sm text-white outline-none placeholder:text-gray-600 focus:ring-1 focus:ring-violet-500/50"
                />
              </div>
            </div>
            <div class="flex gap-2 border-t border-white/[0.06] px-5 py-4">
              <button
                type="button"
                class="flex-1 rounded-lg border border-white/10 py-2.5 text-sm text-gray-400 hover:bg-white/5 hover:text-white transition-colors"
                @click="showEditModal = false"
              >取消</button>
              <button
                type="button"
                :disabled="saving || !editTitle.trim()"
                class="flex-1 rounded-lg bg-violet-600 py-2.5 text-sm font-medium text-white hover:bg-violet-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                @click="submitEdit"
              >{{ saving ? '保存中…' : '保存' }}</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ===== 删除确认弹窗 ===== -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        leave-active-class="transition duration-100 ease-in"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showDeleteConfirm"
          class="fixed inset-0 z-50 flex items-center justify-center px-4"
          @click.self="showDeleteConfirm = false"
        >
          <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="showDeleteConfirm = false" />
          <div class="relative z-10 w-full max-w-xs rounded-2xl bg-[#1a1a22] shadow-2xl p-6" @click.stop>
            <div class="text-center mb-5">
              <div class="w-12 h-12 rounded-full bg-red-500/10 flex items-center justify-center mx-auto mb-3">
                <i class="fa-solid fa-trash text-red-400 text-lg" />
              </div>
              <p class="text-sm font-semibold text-white">删除播放列表</p>
              <p class="mt-1.5 text-xs text-gray-500">确定要删除「{{ playlist?.title }}」吗？此操作不可撤销。</p>
            </div>
            <div class="flex gap-2">
              <button
                type="button"
                class="flex-1 rounded-lg border border-white/10 py-2.5 text-sm text-gray-400 hover:bg-white/5 hover:text-white transition-colors"
                @click="showDeleteConfirm = false"
              >取消</button>
              <button
                type="button"
                :disabled="deleting"
                class="flex-1 rounded-lg bg-red-600 py-2.5 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                @click="confirmDelete"
              >{{ deleting ? '删除中…' : '确认删除' }}</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- ===== 视频删除确认弹窗 ===== -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        leave-active-class="transition duration-100 ease-in"
        leave-to-class="opacity-0"
      >
        <div
          v-if="deleteConfirmId"
          class="fixed inset-0 z-50 flex items-center justify-center px-4"
          @click.self="deleteConfirmId = null"
        >
          <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="deleteConfirmId = null" />
          <div class="relative z-10 w-full max-w-xs rounded-2xl bg-[#1a1a22] shadow-2xl p-6" @click.stop>
            <div class="text-center mb-5">
              <div class="w-12 h-12 rounded-full bg-red-500/10 flex items-center justify-center mx-auto mb-3">
                <i class="fa-solid fa-film text-red-400 text-lg" />
              </div>
              <p class="text-sm font-semibold text-white">删除视频</p>
              <p class="mt-1.5 text-xs text-gray-500 leading-relaxed">
                视频将从平台永久删除，且无法恢复。<br/>确认继续？
              </p>
            </div>
            <div class="flex gap-2">
              <button
                type="button"
                class="flex-1 rounded-lg border border-white/10 py-2.5 text-sm text-gray-400 hover:bg-white/5 hover:text-white transition-colors"
                @click="deleteConfirmId = null"
              >取消</button>
              <button
                type="button"
                :disabled="!!deletingVideoId"
                class="flex-1 rounded-lg bg-red-600 py-2.5 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                @click="confirmDeleteVideo"
              >{{ deletingVideoId ? '删除中…' : '确认删除' }}</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>
