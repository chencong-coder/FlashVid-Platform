<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Uploader as VanUploader,
  showToast,
  type UploaderBeforeRead,
  type UploaderFileListItem,
} from 'vant'

import { uploadFile } from '@/api/upload'
import { createVideo } from '@/api/video'
import { getMusicList, searchMusic, createMusic, type MusicItem } from '@/api/music'
import { getTopics, searchTopics, type TopicItem } from '@/api/topic'
import { useAuthModalStore } from '@/store/authModal'

const router = useRouter()
const authModal = useAuthModalStore()

const files = ref<UploaderFileListItem[]>([])
const coverFiles = ref<UploaderFileListItem[]>([])
const title = ref('')
const caption = ref('')
const publishing = ref(false)
const uploadProgress = ref(0)
const statusText = ref('')

const MAX_VIDEO_FILE_SIZE = 500 * 1024 * 1024
const MAX_COVER_FILE_SIZE = 10 * 1024 * 1024
const SUPPORTED_VIDEO_EXTENSIONS = new Set(['mp4', 'mov', 'avi', 'mkv'])

const beforeReadVideo: UploaderBeforeRead = (file) => {
  const selectedFiles = Array.isArray(file) ? file : [file]

  for (const selectedFile of selectedFiles) {
    const extension = selectedFile.name.split('.').pop()?.toLowerCase() ?? ''
    if (!SUPPORTED_VIDEO_EXTENSIONS.has(extension)) {
      showToast('仅支持 MP4、MOV、AVI、MKV 格式')
      return false
    }
    if (selectedFile.size > MAX_VIDEO_FILE_SIZE) {
      showToast('视频大小不能超过 500 MB')
      return false
    }
  }

  return true
}

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

// ---- 背景音乐选择 ----
const selectedMusic = ref<MusicItem | null>(null)
const musicPickerVisible = ref(false)
const musicList = ref<MusicItem[]>([])
const musicLoading = ref(false)
const musicKeyword = ref('')

const openMusicPicker = async (): Promise<void> => {
  musicPickerVisible.value = true
  if (musicList.value.length === 0) await loadMusic()
}

const loadMusic = async (): Promise<void> => {
  musicLoading.value = true
  try {
    const keyword = musicKeyword.value.trim()
    const res = keyword
      ? await searchMusic(keyword, 1, 50)
      : await getMusicList({ sort: 'hot', page: 1, pageSize: 50 })
    musicList.value = res.data.data.list ?? []
  } catch {
    musicList.value = []
    showToast('音乐加载失败')
  } finally {
    musicLoading.value = false
  }
}

const chooseMusic = (music: MusicItem): void => {
  selectedMusic.value = music
  musicPickerVisible.value = false
}

const clearMusic = (): void => {
  selectedMusic.value = null
}

const localMusicUploading = ref(false)

const uploadLocalMusic = async (event: Event): Promise<void> => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  // 重置 input，允许重复选同一文件
  input.value = ''

  localMusicUploading.value = true
  try {
    const uploadRes = await uploadFile(file, 'audio')
    const musicUrl: string = uploadRes.data.data?.url ?? uploadRes.data.data
    const name = file.name.replace(/\.[^.]+$/, '') // 去掉扩展名
    const res = await createMusic({ name, musicUrl })
    const newMusic = res.data.data?.music
    if (!newMusic) throw new Error('创建失败')
    chooseMusic(newMusic)
    showToast('音乐添加成功')
  } catch {
    showToast('上传失败，请重试')
  } finally {
    localMusicUploading.value = false
  }
}

const musicLabel = computed(() =>
  selectedMusic.value ? `${selectedMusic.value.name} - ${selectedMusic.value.artist}` : '添加背景音乐',
)

// ---- 话题选择 ----
const MAX_TOPICS = 5
const selectedTopics = ref<TopicItem[]>([])
const topicPickerVisible = ref(false)
const topicList = ref<TopicItem[]>([])
const topicLoading = ref(false)
const topicKeyword = ref('')

const openTopicPicker = async (): Promise<void> => {
  topicPickerVisible.value = true
  if (topicList.value.length === 0) await loadTopicList()
}

const loadTopicList = async (): Promise<void> => {
  topicLoading.value = true
  try {
    const keyword = topicKeyword.value.trim()
    const res = keyword
      ? await searchTopics(keyword, undefined, 50)
      : await getTopics({ sort: 'hot', count: 50 })
    topicList.value = res.data.data?.topics ?? []
  } catch {
    topicList.value = []
    showToast('话题加载失败')
  } finally {
    topicLoading.value = false
  }
}

const isTopicSelected = (topicId: string): boolean =>
  selectedTopics.value.some((t) => t.id === topicId)

const toggleTopic = (topic: TopicItem): void => {
  const idx = selectedTopics.value.findIndex((t) => t.id === topic.id)
  if (idx !== -1) {
    selectedTopics.value.splice(idx, 1)
  } else {
    if (selectedTopics.value.length >= MAX_TOPICS) {
      showToast(`最多选 ${MAX_TOPICS} 个话题`)
      return
    }
    selectedTopics.value.push(topic)
  }
}

const removeTopic = (topicId: string): void => {
  selectedTopics.value = selectedTopics.value.filter((t) => t.id !== topicId)
}

/** 从描述中提取 #话题 标签 */
const extractTopicsFromCaption = (text: string): string[] => {
  const regex = /#([一-龥a-zA-Z0-9_]{1,20})/g
  const names: string[] = []
  let m: RegExpExecArray | null
  while ((m = regex.exec(text)) !== null) {
    names.push(m[1])
  }
  return [...new Set(names)] // 去重
}

/** 从视频文件截取第一帧作为封面 */
const extractCover = (videoFile: File): Promise<File> =>
  new Promise((resolve, reject) => {
    const video = document.createElement('video')
    const url = URL.createObjectURL(videoFile)
    video.src = url
    video.muted = true
    video.currentTime = 0.5
    video.addEventListener(
      'seeked',
      () => {
        const canvas = document.createElement('canvas')
        canvas.width = video.videoWidth || 720
        canvas.height = video.videoHeight || 1280
        canvas.getContext('2d')!.drawImage(video, 0, 0, canvas.width, canvas.height)
        URL.revokeObjectURL(url)
        canvas.toBlob(
          (blob) => {
            if (blob) resolve(new File([blob], 'cover.jpg', { type: 'image/jpeg' }))
            else reject(new Error('封面生成失败'))
          },
          'image/jpeg',
          0.85,
        )
      },
      { once: true },
    )
    video.addEventListener(
      'error',
      () => {
        URL.revokeObjectURL(url)
        reject(new Error('视频加载失败'))
      },
      { once: true },
    )
    video.load()
  })

/** 获取视频时长（秒） */
const getVideoDuration = (videoFile: File): Promise<number> =>
  new Promise((resolve) => {
    const video = document.createElement('video')
    const url = URL.createObjectURL(videoFile)
    video.src = url
    video.muted = true
    video.addEventListener(
      'loadedmetadata',
      () => {
        URL.revokeObjectURL(url)
        resolve(Math.ceil(video.duration) || 1)
      },
      { once: true },
    )
    video.addEventListener(
      'error',
      () => {
        URL.revokeObjectURL(url)
        resolve(1)
      },
      { once: true },
    )
    video.load()
  })

const publish = async (): Promise<void> => {
  if (!authModal.requireLogin(router.currentRoute.value.fullPath)) return
  const videoFile = files.value[0]?.file
  if (!(videoFile instanceof File)) {
    showToast('请先选择视频')
    return
  }
  if (!title.value.trim()) {
    showToast('请填写作品标题')
    return
  }

  publishing.value = true
  uploadProgress.value = 0
  try {
    // 1. 上传视频
    statusText.value = '上传视频...'
    const videoRes = await uploadFile(videoFile, 'video', (p: number) => {
      uploadProgress.value = Math.floor(p * 0.7)
    })
    if (videoRes.data.code !== 0) {
      showToast(videoRes.data.message || '视频上传失败')
      return
    }
    const videoUrl = videoRes.data.data.file_url
    const duration = await getVideoDuration(videoFile)

    // 2. 封面：优先用手动选择的封面，否则从视频截帧
    statusText.value = '处理封面...'
    let coverUrl = videoUrl // 兜底用视频地址
    try {
      const manualCover = coverFiles.value[0]?.file
      const coverFile =
        manualCover instanceof File ? manualCover : await extractCover(videoFile)
      statusText.value = '上传封面...'
      const coverRes = await uploadFile(coverFile, 'image', (p: number) => {
        uploadProgress.value = 70 + Math.floor(p * 0.2)
      })
      if (coverRes.data.code === 0) coverUrl = coverRes.data.data.file_url
    } catch {
      /* 封面失败不阻断发布 */
    }

    // 3. 发布视频
    statusText.value = '发布中...'
    uploadProgress.value = 90

    // 合并手动选择的话题 + 描述中提取的话题标签
    const manualTopics = selectedTopics.value.map((t) => t.name)
    const captionTopics = extractTopicsFromCaption(caption.value)
    const allTopics = [...new Set([...manualTopics, ...captionTopics])].slice(0, 5) // 去重 + 限制 5 个

    const createRes = await createVideo({
      title: title.value.trim(),
      description: caption.value.trim() || undefined,
      videoUrl,
      coverUrl,
      duration,
      musicId: selectedMusic.value?.id,
      topicNames: allTopics.length > 0 ? allTopics : undefined,
    })
    if (createRes.data.code === 0) {
      uploadProgress.value = 100
      showToast('发布成功')
      await router.push('/')
    } else {
      showToast(createRes.data.message || '发布失败')
    }
  } catch (error: unknown) {
    showToast(error instanceof Error ? error.message : '发布失败，请重试')
  } finally {
    publishing.value = false
    statusText.value = ''
    uploadProgress.value = 0
  }
}
</script>

<template>
  <main class="safe-top no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] px-4 pb-8 text-white">
    <header class="flex h-14 items-center justify-between">
      <button
        type="button"
        aria-label="返回"
        class="h-9 w-9 text-left text-xl"
        @click="router.back()"
      >
        <i class="fa-solid fa-chevron-left" />
      </button>
      <h1 class="text-base font-semibold">发布作品</h1>
      <span class="w-9" />
    </header>

    <!-- 视频 + 封面选择 -->
    <section class="mt-4 flex gap-3">
      <van-uploader
        v-model="files"
        accept=".mp4,.mov,.avi,.mkv"
        :before-read="beforeReadVideo"
        :max-count="1"
        :max-size="MAX_VIDEO_FILE_SIZE"
      >
        <div
          class="flex aspect-[9/13] w-44 flex-col items-center justify-center rounded-md border border-dashed border-white/30 bg-white/5 text-neutral-400"
        >
          <template v-if="files[0]?.url">
            <video :src="files[0]?.url" class="h-full w-full rounded-md object-cover" muted />
          </template>
          <template v-else>
            <i class="fa-solid fa-video mb-3 text-3xl" />
            <span class="text-sm font-medium text-white">选择视频</span>
            <span class="mt-2 text-xs">最大 500 MB</span>
          </template>
        </div>
      </van-uploader>

      <!-- 封面（可选，手动选择；不选则自动截帧） -->
      <van-uploader
        v-model="coverFiles"
        accept="image/*"
        :before-read="beforeReadCover"
        :max-count="1"
        :max-size="MAX_COVER_FILE_SIZE"
      >
        <div
          class="flex aspect-[9/13] w-28 flex-col items-center justify-center rounded-md border border-dashed border-white/30 bg-white/5 text-neutral-400"
        >
          <template v-if="coverFiles[0]?.url">
            <img :src="coverFiles[0]?.url" class="h-full w-full rounded-md object-cover" alt="封面" />
          </template>
          <template v-else>
            <i class="fa-solid fa-image mb-2 text-2xl" />
            <span class="text-xs font-medium text-white">选择封面</span>
            <span class="mt-1 text-[10px] leading-4 text-center">可选<br />默认取首帧</span>
          </template>
        </div>
      </van-uploader>
    </section>

    <!-- 标题（必填） -->
    <section class="mt-4">
      <div class="rounded-lg bg-white/5 px-4 py-3">
        <input
          v-model="title"
          type="text"
          maxlength="100"
          placeholder="添加作品标题（必填）"
          class="w-full bg-transparent text-sm text-white outline-none placeholder:text-neutral-500"
        />
      </div>
    </section>

    <section class="mt-3 border-y border-white/10 py-4">
      <textarea
        v-model="caption"
        rows="3"
        maxlength="500"
        placeholder="添加作品描述，让更多人看见..."
        class="w-full resize-none bg-transparent text-sm leading-6 text-white outline-none placeholder:text-neutral-500"
      />
      <div class="flex gap-4 text-sm font-medium">
        <button type="button" class="text-primary" @click="openTopicPicker"># 话题</button>
        <button type="button">@ 朋友</button>
      </div>

      <!-- 已选话题 -->
      <div v-if="selectedTopics.length > 0" class="mt-3 flex flex-wrap gap-2">
        <button
          v-for="topic in selectedTopics"
          :key="topic.id"
          type="button"
          class="flex items-center gap-1.5 rounded-full bg-primary/20 px-3 py-1 text-xs font-medium text-primary"
          @click="removeTopic(topic.id)"
        >
          <span># {{ topic.name }}</span>
          <i class="fa-solid fa-xmark text-[10px]" />
        </button>
      </div>

      <!-- 话题下拉面板 -->
      <div v-if="topicPickerVisible" class="fixed inset-0 z-10" @click="topicPickerVisible = false" />
      <div
        v-if="topicPickerVisible"
        class="relative z-20 mt-2 overflow-hidden rounded-xl border border-white/10 bg-[#1a1a2e] text-white"
      >
        <!-- 搜索栏 -->
        <div class="flex items-center gap-2 border-b border-white/10 px-3 py-2">
          <i class="fa-solid fa-magnifying-glass text-xs text-neutral-500" />
          <input
            v-model="topicKeyword"
            type="text"
            placeholder="搜索话题"
            class="flex-1 bg-transparent text-sm outline-none placeholder:text-neutral-500"
            @keyup.enter="loadTopicList"
          />
          <button
            v-if="topicKeyword"
            type="button"
            class="text-xs text-neutral-400"
            @click="loadTopicList"
          >
            搜索
          </button>
          <button type="button" class="ml-1 text-neutral-500" @click="topicPickerVisible = false">
            <i class="fa-solid fa-xmark text-xs" />
          </button>
        </div>
        <!-- 列表 -->
        <div class="no-scrollbar max-h-64 overflow-y-auto">
          <div v-if="topicLoading" class="flex justify-center py-6">
            <i class="fa-solid fa-circle-notch animate-spin text-xl text-neutral-500" />
          </div>
          <div v-else-if="topicList.length === 0" class="py-8 text-center text-sm text-neutral-500">
            暂无话题
          </div>
          <button
            v-for="(topic, idx) in topicList"
            v-else
            :key="topic.id"
            type="button"
            class="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-white/5"
            :class="isTopicSelected(topic.id) ? 'bg-white/5' : ''"
            @click="toggleTopic(topic)"
          >
            <span class="w-5 text-center text-sm font-bold" :class="idx < 3 ? 'text-primary' : 'text-neutral-500'">
              {{ idx + 1 }}
            </span>
            <span class="flex-1 truncate text-sm">{{ topic.name }}</span>
            <span class="text-xs text-neutral-500">{{ topic.videoCount }} 视频</span>
            <i v-if="isTopicSelected(topic.id)" class="fa-solid fa-circle-check ml-2 text-primary" />
          </button>
        </div>
        <div class="border-t border-white/10 px-4 py-2 text-xs text-neutral-500">
          已选 {{ selectedTopics.length }}/{{ MAX_TOPICS }}
        </div>
      </div>
    </section>

    <!-- 背景音乐 -->
    <section class="text-sm">
      <button
        type="button"
        class="flex w-full items-center justify-between py-4"
        @click="openMusicPicker"
      >
        <span class="flex min-w-0 items-center">
          <i class="fa-solid fa-music mr-3 w-4 text-center" />
          <span class="truncate" :class="selectedMusic ? 'text-white' : ''">{{ musicLabel }}</span>
        </span>
        <span class="flex items-center gap-2 text-neutral-600">
          <i
            v-if="selectedMusic"
            class="fa-solid fa-xmark px-1"
            @click.stop="clearMusic"
          />
          <i class="fa-solid fa-chevron-right text-xs" />
        </span>
      </button>

      <!-- 音乐下拉面板 -->
      <div v-if="musicPickerVisible" class="fixed inset-0 z-10" @click="musicPickerVisible = false" />
      <div
        v-if="musicPickerVisible"
        class="relative z-20 mb-2 overflow-hidden rounded-xl border border-white/10 bg-[#1a1a2e] text-white"
      >
        <!-- 搜索栏 -->
        <div class="flex items-center gap-2 border-b border-white/10 px-3 py-2">
          <i class="fa-solid fa-magnifying-glass text-xs text-neutral-500" />
          <input
            v-model="musicKeyword"
            type="text"
            placeholder="搜索歌曲或艺术家"
            class="flex-1 bg-transparent text-sm outline-none placeholder:text-neutral-500"
            @keyup.enter="loadMusic"
          />
          <button
            v-if="musicKeyword"
            type="button"
            class="text-xs text-neutral-400"
            @click="loadMusic"
          >
            搜索
          </button>
          <button type="button" class="ml-1 text-neutral-500" @click="musicPickerVisible = false">
            <i class="fa-solid fa-xmark text-xs" />
          </button>
        </div>
        <!-- 本地上传 -->
        <div class="border-b border-white/10 px-3 py-2">
          <label
            class="flex cursor-pointer items-center gap-2 text-sm hover:opacity-80"
            :class="localMusicUploading ? 'pointer-events-none text-neutral-500' : 'text-primary'"
          >
            <i
              class="fa-solid text-xs"
              :class="localMusicUploading ? 'fa-circle-notch animate-spin' : 'fa-upload'"
            />
            <span>{{ localMusicUploading ? '上传中…' : '从本地选取' }}</span>
            <input
              type="file"
              accept=".mp3,.wav,.aac,.ogg,.m4a"
              class="hidden"
              :disabled="localMusicUploading"
              @change="uploadLocalMusic"
            />
          </label>
        </div>
        <!-- 列表 -->
        <div class="no-scrollbar max-h-64 overflow-y-auto">
          <div v-if="musicLoading" class="flex justify-center py-6">
            <i class="fa-solid fa-circle-notch animate-spin text-xl text-neutral-500" />
          </div>
          <div v-else-if="musicList.length === 0" class="py-8 text-center text-sm text-neutral-500">
            暂无音乐
          </div>
          <button
            v-for="music in musicList"
            v-else
            :key="music.id"
            type="button"
            class="flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-white/5"
            :class="selectedMusic?.id === music.id ? 'bg-white/5' : ''"
            @click="chooseMusic(music)"
          >
            <img
              :src="music.coverUrl || `https://picsum.photos/40/40?random=${music.id}`"
              :alt="music.name"
              class="h-10 w-10 shrink-0 rounded-md object-cover"
            />
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium">{{ music.name }}</p>
              <p class="truncate text-xs text-neutral-500">{{ music.artist }}</p>
            </div>
            <i v-if="selectedMusic?.id === music.id" class="fa-solid fa-circle-check text-primary" />
          </button>
        </div>
      </div>
    </section>

    <section class="text-sm">
      <button type="button" class="flex w-full items-center justify-between border-t border-white/10 py-4">
        <span><i class="fa-solid fa-location-dot mr-3 w-4 text-center" />添加位置</span>
        <i class="fa-solid fa-chevron-right text-xs text-neutral-600" />
      </button>
    </section>

    <!-- 上传进度 -->
    <div v-if="publishing" class="mt-6 space-y-2">
      <div class="flex justify-between text-xs text-neutral-400">
        <span>{{ statusText }}</span>
        <span>{{ uploadProgress }}%</span>
      </div>
      <div class="h-1 w-full overflow-hidden rounded-full bg-white/10">
        <div
          class="h-full rounded-full bg-primary transition-all duration-300"
          :style="{ width: `${uploadProgress}%` }"
        />
      </div>
    </div>

    <button
      type="button"
      class="mt-6 h-12 w-full rounded bg-primary text-sm font-semibold disabled:opacity-50"
      :disabled="publishing || !files.length || !title.trim()"
      @click="publish"
    >
      {{ publishing ? statusText || '发布中...' : '发布' }}
    </button>

  </main>
</template>
