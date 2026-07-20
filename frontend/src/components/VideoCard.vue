<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { videoPlay } from 'vue3-video-play'

import type { VideoItem } from '@/types/video'
import RightAction from './RightAction.vue'

interface Props {
  video: VideoItem
  active: boolean
  muted: boolean
}

interface Emits {
  (event: 'follow', videoId: string): void
  (event: 'like', videoId: string): void
  (event: 'comment', videoId: string): void
  (event: 'favorite', videoId: string): void
  (event: 'share', videoId: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const playerRoot = ref<HTMLElement | null>(null)
const playing = ref(false)
const showStatus = ref(false)
let statusTimer: number | undefined

const options = computed(() => ({
  width: '100%',
  height: '100%',
  src: props.video.source,
  poster: props.video.poster,
  muted: props.muted,
  autoPlay: false,
  loop: true,
  volume: 0.7,
  control: false,
  playsinline: true,
  preload: 'auto', // 自动预加载视频
}))

const getVideoElement = (): HTMLVideoElement | null =>
  playerRoot.value?.querySelector('video') ?? null

const play = async (): Promise<void> => {
  const video = getVideoElement()
  if (!video) return
  video.muted = props.muted
  try {
    await video.play()
    playing.value = true
  } catch {
    playing.value = false
  }
}

const pause = (): void => {
  const video = getVideoElement()
  video?.pause()
  playing.value = false
}

const togglePlayback = async (): Promise<void> => {
  if (playing.value) pause()
  else await play()
  showStatus.value = true
  window.clearTimeout(statusTimer)
  statusTimer = window.setTimeout(() => (showStatus.value = false), 650)
}

watch(
  () => props.active,
  async (active) => {
    await nextTick()
    if (active) await play()
    else pause()
  },
  { immediate: true },
)

watch(
  () => props.muted,
  (muted) => {
    const video = getVideoElement()
    if (video) video.muted = muted
  },
)

onBeforeUnmount(() => {
  pause()
  window.clearTimeout(statusTimer)
})
</script>

<template>
  <article class="relative h-full w-full overflow-hidden bg-black" @click="togglePlayback">
    <!-- 始终渲染视频播放器，实现预加载 -->
    <div ref="playerRoot" class="flash-video-player absolute inset-0 h-full w-full">
      <videoPlay v-bind="options" />
    </div>

    <div class="video-gradient pointer-events-none absolute inset-0" />

    <Transition
      enter-active-class="transition duration-150"
      enter-from-class="scale-75 opacity-0"
      leave-active-class="transition duration-200"
      leave-to-class="scale-125 opacity-0"
    >
      <div
        v-if="showStatus"
        class="pointer-events-none absolute left-1/2 top-1/2 z-20 flex h-20 w-20 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full bg-white/20 text-3xl backdrop-blur-xl"
      >
        <i class="fa-solid" :class="playing ? 'fa-play' : 'fa-pause'" />
      </div>
    </Transition>

    <!-- 视频信息区域 - 更现代的毛玻璃设计 -->
    <div class="absolute bottom-20 left-0 right-16 z-20 px-4 pb-2">
      <!-- 作者信息 -->
      <div class="mb-3 flex items-center gap-2">
        <div class="text-[16px] font-bold text-white drop-shadow-lg">
          @{{ video.author.nickname }}
        </div>
        <div
          v-if="video.city"
          class="inline-flex items-center gap-1 rounded-full bg-white/15 px-2.5 py-1 text-xs backdrop-blur-md"
        >
          <i class="fa-solid fa-location-dot text-rose-400" />
          <span class="font-medium">{{ video.city }}</span>
        </div>
      </div>

      <!-- 视频描述 -->
      <div
        class="mb-3 max-w-[85%] rounded-2xl bg-gradient-to-br from-black/40 to-black/20 p-3 backdrop-blur-md"
      >
        <p class="line-clamp-3 text-[15px] leading-relaxed text-white">
          {{ video.description }}
          <span v-for="topic in video.topics" :key="topic" class="ml-1 font-semibold text-cyan-400">
            #{{ topic }}
          </span>
        </p>
      </div>

      <!-- 音乐信息 - 带动画的卡片 -->
      <div
        class="inline-flex max-w-[75%] items-center gap-2 rounded-full bg-gradient-to-r from-purple-500/20 to-pink-500/20 px-3 py-1.5 backdrop-blur-md"
      >
        <i class="fa-solid fa-music text-sm text-purple-300" />
        <div class="overflow-hidden">
          <span class="inline-block animate-[marquee_12s_linear_infinite] text-sm font-medium">
            {{ video.music }}
          </span>
        </div>
      </div>
    </div>

    <RightAction
      :video="video"
      :playing="playing"
      @follow="emit('follow', video.id)"
      @like="emit('like', video.id)"
      @comment="emit('comment', video.id)"
      @favorite="emit('favorite', video.id)"
      @share="emit('share', video.id)"
    />
  </article>
</template>
