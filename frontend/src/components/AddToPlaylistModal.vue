<script setup lang="ts">
import { ref, watch } from 'vue'
import { showToast } from 'vant'
import {
  getMyPlaylists,
  addVideoToPlaylist,
  createPlaylist,
  type PlaylistInfo,
} from '@/api/user'

interface Props {
  /** 要添加的视频 ID，null 时弹窗关闭 */
  videoId: string | null
}
interface Emits {
  (event: 'close'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const playlists = ref<PlaylistInfo[]>([])
const loading = ref(false)
const addingId = ref<string | null>(null)
const addedIds = ref<Set<string>>(new Set())

// 每次打开（videoId 变为非空）都重新加载
watch(
  () => props.videoId,
  async (videoId) => {
    if (!videoId) return
    addedIds.value = new Set()
    loading.value = true
    try {
      const res = await getMyPlaylists()
      playlists.value = res.data.data.playlists
    } catch {
      showToast('加载播放列表失败')
    } finally {
      loading.value = false
    }
  },
)

const addTo = async (playlistId: string) => {
  if (!props.videoId) return
  if (addedIds.value.has(playlistId)) return // 已添加，幂等
  addingId.value = playlistId
  try {
    await addVideoToPlaylist(playlistId, props.videoId)
    addedIds.value = new Set([...addedIds.value, playlistId])
    showToast('已添加到播放列表')
  } catch {
    showToast('添加失败，请重试')
  } finally {
    addingId.value = null
  }
}

// ── 内联新建播放列表 ──
const showCreate = ref(false)
const newTitle = ref('')
const creating = ref(false)

const openCreate = () => {
  newTitle.value = ''
  showCreate.value = true
}

const submitCreate = async () => {
  const title = newTitle.value.trim()
  if (!title) { showToast('请输入名称'); return }
  creating.value = true
  try {
    const res = await createPlaylist({ title })
    const newPl = res.data.data.playlist
    playlists.value.unshift(newPl)
    showCreate.value = false
    // 创建后自动添加
    await addTo(newPl.id)
  } catch {
    showToast('创建失败')
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-150 ease-in"
      leave-to-class="opacity-0"
    >
      <div
        v-if="videoId"
        class="fixed inset-0 z-[60] flex items-end justify-center sm:items-center sm:px-4"
        @click.self="emit('close')"
      >
        <!-- 遮罩 -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="emit('close')" />

        <!-- 卡片：移动端 bottom-sheet，桌面居中 -->
        <div
          class="relative z-10 w-full max-w-sm rounded-t-2xl bg-[#1a1a22] shadow-2xl sm:rounded-2xl"
          @click.stop
        >
          <!-- 标题栏 -->
          <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-4">
            <span class="text-sm font-semibold text-white">加入播放列表</span>
            <button
              type="button"
              class="flex h-7 w-7 items-center justify-center rounded-full text-gray-500 hover:bg-white/10 hover:text-white"
              @click="emit('close')"
            >
              <i class="fa-solid fa-xmark text-xs" />
            </button>
          </div>

          <!-- 播放列表 -->
          <div class="max-h-72 overflow-y-auto overscroll-contain">
            <!-- 加载中骨架 -->
            <div v-if="loading" class="space-y-0.5 p-2">
              <div v-for="i in 4" :key="i" class="flex items-center gap-3 rounded-xl px-3 py-2.5">
                <div class="h-10 w-10 animate-pulse rounded-lg bg-white/[0.06] shrink-0" />
                <div class="flex-1 space-y-1.5">
                  <div class="h-3 w-2/3 animate-pulse rounded bg-white/[0.06]" />
                  <div class="h-2 w-1/4 animate-pulse rounded bg-white/[0.06]" />
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-else-if="playlists.length === 0" class="py-6 text-center text-xs text-gray-600">
              还没有播放列表，新建一个吧
            </div>

            <!-- 列表项 -->
            <ul v-else class="space-y-0.5 p-2">
              <li v-for="pl in playlists" :key="pl.id">
                <button
                  type="button"
                  class="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 transition-colors hover:bg-white/[0.06]"
                  :disabled="addingId === pl.id"
                  @click="addTo(pl.id)"
                >
                  <img
                    :src="pl.coverUrl || `https://picsum.photos/40/40?random=${pl.id}`"
                    :alt="pl.title"
                    class="h-10 w-10 shrink-0 rounded-lg object-cover"
                  />
                  <div class="min-w-0 flex-1 text-left">
                    <p class="truncate text-sm font-medium text-white">{{ pl.title }}</p>
                    <p class="text-[11px] text-gray-500">{{ pl.videoCount }} 个视频</p>
                  </div>
                  <!-- 状态图标 -->
                  <span class="shrink-0 text-sm">
                    <i
                      v-if="addedIds.has(pl.id)"
                      class="fa-solid fa-circle-check text-green-400"
                    />
                    <i
                      v-else-if="addingId === pl.id"
                      class="fa-solid fa-spinner fa-spin text-gray-400"
                    />
                    <i v-else class="fa-solid fa-plus text-gray-600" />
                  </span>
                </button>
              </li>
            </ul>
          </div>

          <!-- 新建播放列表区域 -->
          <div class="border-t border-white/[0.06] px-4 py-3">
            <div v-if="!showCreate">
              <button
                type="button"
                class="flex w-full items-center gap-2 rounded-xl px-3 py-2 text-sm text-gray-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                @click="openCreate"
              >
                <i class="fa-solid fa-plus text-xs" />
                新建播放列表
              </button>
            </div>
            <div v-else class="flex gap-2">
              <input
                v-model="newTitle"
                type="text"
                maxlength="50"
                placeholder="播放列表名称"
                class="min-w-0 flex-1 rounded-lg bg-white/[0.06] px-3 py-2 text-sm text-white outline-none placeholder:text-gray-600 focus:ring-1 focus:ring-violet-500/50"
                autofocus
                @keydown.enter="submitCreate"
                @keydown.esc="showCreate = false"
              />
              <button
                type="button"
                :disabled="creating || !newTitle.trim()"
                class="shrink-0 rounded-lg bg-violet-600 px-3 py-2 text-sm font-medium text-white hover:bg-violet-500 disabled:opacity-40 transition-colors"
                @click="submitCreate"
              >
                {{ creating ? '…' : '创建' }}
              </button>
              <button
                type="button"
                class="shrink-0 rounded-lg bg-white/[0.06] px-3 py-2 text-sm text-gray-400 hover:text-white transition-colors"
                @click="showCreate = false"
              >取消</button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
