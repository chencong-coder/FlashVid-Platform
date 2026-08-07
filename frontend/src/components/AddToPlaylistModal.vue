<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  Uploader as VanUploader,
  showToast,
  type UploaderBeforeRead,
  type UploaderFileListItem,
} from 'vant'
import {
  getMyPlaylists,
  addVideoToPlaylist,
  createPlaylist,
  type PlaylistInfo,
} from '@/api/user'
import { uploadFile } from '@/api/upload'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

interface Props {
  /** 要添加的视频 ID，null 时面板关闭 */
  videoId: string | null
  /** 触发按钮的锚点位置（fixed坐标），用于精确贴靠操作栏 */
  anchor?: { right: number; bottom: number } | null
}
interface Emits {
  (event: 'close'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const playlists   = ref<PlaylistInfo[]>([])
const loading     = ref(false)
const selectedIds = ref<Set<string>>(new Set())
const saving      = ref(false)

// ── 新建播放列表 ──
const showCreate  = ref(false)
const newTitle    = ref('')
const creating    = ref(false)
const coverFiles  = ref<UploaderFileListItem[]>([])

const MAX_COVER_FILE_SIZE = 10 * 1024 * 1024

const beforeReadCover: UploaderBeforeRead = (file) => {
  const selectedFiles = Array.isArray(file) ? file : [file]
  for (const selectedFile of selectedFiles) {
    if (!selectedFile.type.startsWith('image/')) {
      showToast('封面仅支持图片格式')
      return false
    }
    if (selectedFile.size > MAX_COVER_FILE_SIZE) {
      showToast('封面大小不能超过 10 MB')
      return false
    }
  }
  return true
}

// 每次打开（videoId 变为非空）都重新加载列表、重置选中
watch(
  () => props.videoId,
  async (videoId) => {
    if (!videoId) {
      showCreate.value = false
      return
    }
    selectedIds.value = new Set()
    showCreate.value  = false
    loading.value     = true
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

const toggleSelect = (id: string) => {
  const s = new Set(selectedIds.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selectedIds.value = s
}

/** 保存：把视频加入所有已选播放列表 */
const save = async () => {
  if (!props.videoId) { emit('close'); return }
  if (selectedIds.value.size === 0) { emit('close'); return }

  saving.value = true
  let ok = 0
  let lastErrMsg = ''
  for (const plId of selectedIds.value) {
    try {
      await addVideoToPlaylist(plId, props.videoId)
      ok++
    } catch (e: unknown) {
      lastErrMsg = (e instanceof Error ? e.message : '') || '添加失败'
    }
  }
  saving.value = false
  if (ok > 0) {
    showToast(ok === 1 ? '已添加到播放列表' : `已添加到 ${ok} 个播放列表`)
    // 通知侧边栏刷新播放列表计数
    userStore.bumpPlaylistVersion()
  } else if (lastErrMsg) {
    showToast(lastErrMsg)
  }
  emit('close')
}

const openCreate = () => {
  newTitle.value   = ''
  coverFiles.value = []
  showCreate.value = true
}

const submitCreate = async () => {
  const title = newTitle.value.trim()
  if (!title) { showToast('请输入名称'); return }
  creating.value = true
  try {
    // 上传封面（可选）
    let coverUrl: string | undefined
    const coverFile = coverFiles.value[0]?.file
    if (coverFile instanceof File) {
      try {
        const coverRes = await uploadFile(coverFile, 'image')
        if (coverRes.data.code === 0) coverUrl = coverRes.data.data.file_url
      } catch {
        /* 封面失败不阻断创建 */
      }
    }
    const res   = await createPlaylist({ title, coverUrl })
    const newPl = res.data.data.playlist
    playlists.value.unshift(newPl)
    // 创建后自动选中
    const s = new Set(selectedIds.value)
    s.add(newPl.id)
    selectedIds.value = s
    showCreate.value  = false
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
      enter-from-class="opacity-0 translate-x-3"
      leave-active-class="transition duration-150 ease-in"
      leave-to-class="opacity-0 translate-x-3"
    >
      <!--
        面板定位：贴左边（left-4），底部留出底部导航（bottom-24）
        避免与右侧操作栏重叠
      -->
      <div
        v-if="videoId"
        class="fixed z-[60] w-[278px] overflow-hidden rounded-2xl bg-[#1c1c24] shadow-2xl"
        :style="
          anchor
            ? { right: anchor.right + 'px', bottom: anchor.bottom + 'px' }
            : { right: '80px', bottom: '96px' }
        "
        style="max-height: min(72vh, 460px)"
        @click.stop
      >
        <!-- ── 标题栏 ── -->
        <div class="flex items-center justify-between px-4 py-3.5">
          <span class="text-sm font-semibold text-white">选择收藏夹</span>
          <button
            type="button"
            class="flex items-center gap-1.5 text-xs text-white/75 transition-colors hover:text-white"
            @click="openCreate"
          >
            <i class="fa-solid fa-circle-plus text-sm" />
            <span>新建</span>
          </button>
        </div>

        <!-- ── 播放列表 ── -->
        <div class="overflow-y-auto overscroll-contain" style="max-height: 290px">
          <!-- 骨架屏 -->
          <div v-if="loading" class="space-y-1 px-2 pb-2">
            <div
              v-for="i in 3"
              :key="i"
              class="flex items-center gap-3 rounded-xl px-3 py-2"
            >
              <div class="h-11 w-11 shrink-0 animate-pulse rounded-lg bg-white/[0.07]" />
              <div class="flex-1 space-y-2">
                <div class="h-3 w-3/4 animate-pulse rounded bg-white/[0.07]" />
                <div class="h-2 w-1/3 animate-pulse rounded bg-white/[0.07]" />
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <div
            v-else-if="playlists.length === 0"
            class="py-10 text-center text-xs text-gray-500"
          >
            还没有播放列表，点「新建」创建一个
          </div>

          <!-- 列表 -->
          <ul v-else class="px-2 pb-2">
            <li v-for="pl in playlists" :key="pl.id">
              <button
                type="button"
                class="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 transition-colors hover:bg-white/[0.05] active:bg-white/[0.08]"
                @click="toggleSelect(pl.id)"
              >
                <!-- 封面 -->
                <div class="h-11 w-11 shrink-0 rounded-lg overflow-hidden bg-white/[0.07]">
                  <img
                    v-if="pl.coverUrl"
                    :src="pl.coverUrl"
                    :alt="pl.title"
                    class="h-full w-full object-cover"
                  />
                  <div
                    v-else
                    class="h-full w-full flex items-center justify-center"
                    :style="{ background: `hsl(${Number(pl.id.slice(-6)) % 360}, 38%, 28%)` }"
                  >
                    <i class="fa-solid fa-list text-white/60 text-sm" />
                  </div>
                </div>
                <!-- 标题 + 数量 -->
                <div class="min-w-0 flex-1 text-left">
                  <p class="truncate text-sm font-medium text-white">{{ pl.title }}</p>
                  <p class="mt-0.5 text-[11px] text-gray-500">{{ pl.videoCount }} 个视频</p>
                </div>
                <!-- 复选框 -->
                <div
                  class="flex h-5 w-5 shrink-0 items-center justify-center rounded border-2 transition-all"
                  :class="
                    selectedIds.has(pl.id)
                      ? 'border-rose-500 bg-rose-500'
                      : 'border-gray-500 bg-transparent'
                  "
                >
                  <i
                    v-if="selectedIds.has(pl.id)"
                    class="fa-solid fa-check text-[9px] text-white"
                  />
                </div>
              </button>
            </li>
          </ul>
        </div>

        <!-- ── 新建输入区 ── -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="-translate-y-2 opacity-0"
          leave-active-class="transition duration-150 ease-in"
          leave-to-class="-translate-y-2 opacity-0"
        >
          <div
            v-if="showCreate"
            class="border-t border-white/[0.06] px-4 py-3"
          >
            <!-- 封面选择（可选） -->
            <div class="mb-2 flex items-center gap-3">
              <van-uploader
                v-model="coverFiles"
                accept="image/*"
                :before-read="beforeReadCover"
                :max-count="1"
                :max-size="MAX_COVER_FILE_SIZE"
              >
                <div
                  class="flex h-14 w-14 items-center justify-center overflow-hidden rounded-lg border border-dashed border-white/25 bg-white/[0.05] text-gray-400"
                >
                  <img
                    v-if="coverFiles[0]?.url"
                    :src="coverFiles[0]?.url"
                    alt="封面"
                    class="h-full w-full object-cover"
                  />
                  <i v-else class="fa-solid fa-image text-lg" />
                </div>
              </van-uploader>
              <span class="text-xs text-gray-500">选择封面（可选）</span>
            </div>
            <div class="flex gap-2">
              <input
                v-model="newTitle"
                type="text"
                maxlength="50"
                placeholder="播放列表名称"
                class="min-w-0 flex-1 rounded-lg bg-white/[0.07] px-3 py-2 text-sm text-white outline-none placeholder:text-gray-600 focus:ring-1 focus:ring-rose-500/50"
                autofocus
                @keydown.enter="submitCreate"
                @keydown.esc="showCreate = false"
              />
              <button
                type="button"
                :disabled="creating || !newTitle.trim()"
                class="shrink-0 rounded-lg bg-rose-500 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-rose-400 disabled:opacity-40"
                @click="submitCreate"
              >
                {{ creating ? '…' : '创建' }}
              </button>
              <button
                type="button"
                class="shrink-0 rounded-lg bg-white/[0.06] px-3 py-2 text-sm text-gray-400 transition-colors hover:text-white"
                @click="showCreate = false"
              >
                取消
              </button>
            </div>
          </div>
        </Transition>

        <!-- ── 保存按钮 ── -->
        <div class="px-4 pb-4 pt-2">
          <button
            type="button"
            :disabled="saving"
            class="w-full rounded-full bg-rose-500 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-rose-400 active:bg-rose-600 disabled:opacity-50"
            @click="save"
          >
            {{ saving ? '保存中…' : '保存' }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- 遮罩（右侧留出操作栏，点遮罩关闭面板）-->
    <Transition
      enter-active-class="transition duration-200"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-150"
      leave-to-class="opacity-0"
    >
      <div
        v-if="videoId"
        class="fixed inset-0 z-[59] bg-transparent"
        @click="emit('close')"
      />
    </Transition>
  </Teleport>
</template>
