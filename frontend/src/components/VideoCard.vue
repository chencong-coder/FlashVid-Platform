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
        class="pointer-events-none absolute left-1/2 top-1/2 z-20 flex h-16 w-16 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full bg-black/45 text-2xl backdrop-blur-sm"
      >
        <i class="fa-solid" :class="playing ? 'fa-play' : 'fa-pause'" />
      </div>
    </Transition>

    <div class="absolute bottom-6 left-3 z-20 w-[72%] text-white text-shadow-video">
      <div class="mb-2 text-[15px] font-bold">@{{ video.author.nickname }}</div>
      <p class="line-clamp-3 text-[14px] leading-[1.45]">
        {{ video.description }}
        <span v-for="topic in video.topics" :key="topic" class="ml-1 font-semibold"
          >#{{ topic }}</span
        >
      </p>
      <div
        v-if="video.city"
        class="mt-2 inline-flex items-center gap-1 rounded bg-black/30 px-1.5 py-0.5 text-xs"
      >
        <i class="fa-solid fa-location-dot" />
        {{ video.city }}
      </div>
      <div class="mt-2 flex items-center gap-2 text-xs">
        <i class="fa-solid fa-music" />
        <div class="max-w-[14rem] overflow-hidden whitespace-nowrap">
          <span class="inline-block animate-[marquee_8s_linear_infinite]">{{ video.music }}</span>
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
