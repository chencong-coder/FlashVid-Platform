<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { videoPlay } from 'vue3-video-play'

import type { VideoItem } from '@/types/video'
import { searchTopics } from '@/api/topic'
import RightAction from './RightAction.vue'

interface Props {
  video: VideoItem
  active: boolean
  muted: boolean
}

interface Emits {
  (event: 'follow', videoId: string): void
  (event: 'profile', videoId: string): void
  (event: 'like', videoId: string): void
  (event: 'comment', videoId: string): void
  (event: 'favorite', videoId: string): void
  (event: 'share', videoId: string): void
  (event: 'playlist', videoId: string, anchor: { right: number; bottom: number } | null): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const router = useRouter()

const TOPIC_RE = /#([一-龥a-zA-Z0-9_]{1,20})/g

/** 把描述拆成 文本 / 话题 片段，话题可点击 */
interface CaptionSegment {
  type: 'text' | 'topic'
  value: string
}
const captionSegments = computed<CaptionSegment[]>(() => {
  const text = props.video.description ?? ''
  const segments: CaptionSegment[] = []
  let lastIndex = 0
  let m: RegExpExecArray | null
  TOPIC_RE.lastIndex = 0
  while ((m = TOPIC_RE.exec(text)) !== null) {
    if (m.index > lastIndex) {
      segments.push({ type: 'text', value: text.slice(lastIndex, m.index) })
    }
    segments.push({ type: 'topic', value: m[1] })
    lastIndex = m.index + m[0].length
  }
  if (lastIndex < text.length) {
    segments.push({ type: 'text', value: text.slice(lastIndex) })
  }
  return segments
})

/** 描述里已出现的话题名（用于去重，避免 video.topics 再渲染一遍） */
const captionTopicNames = computed(
  () => new Set(captionSegments.value.filter((s) => s.type === 'topic').map((s) => s.value)),
)

/** video.topics 中未被描述覆盖的话题（老数据兜底） */
const extraTopics = computed(() =>
  (props.video.topics ?? []).filter((name) => !captionTopicNames.value.has(name)),
)

let navigatingTopic = false

/** 点话题：按名称解析出 topicId 后跳到话题页 */
const goTopic = async (name: string): Promise<void> => {
  if (navigatingTopic) return
  navigatingTopic = true
  try {
    const res = await searchTopics(name, undefined, 20)
    const topics = res.data.data?.topics ?? []
    const match = topics.find((t) => t.name === name) ?? topics[0]
    if (!match) {
      showToast('话题不存在')
      return
    }
    void router.push({ name: 'topic', params: { id: match.id } })
  } catch {
    showToast('话题加载失败')
  } finally {
    navigatingTopic = false
  }
}

const playerRoot = ref<HTMLElement | null>(null)
const playing = ref(false)
const showStatus = ref(false)
let statusTimer: number | undefined

const options = computed(() => ({
  width: '100%',
  height: '100%',
  src: props.video.source,
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

    <!-- 视频信息区域 - 纯文字透明设计 -->
    <div class="absolute bottom-[4.5rem] left-0 right-16 z-20 px-4 pb-1">
      <!-- 作者信息 -->
      <div class="mb-2.5 flex items-center gap-2">
        <div class="text-[16px] font-bold text-white drop-shadow-[0_2px_8px_rgba(0,0,0,0.8)]">
          @{{ video.author.nickname }}
        </div>
        <div
          v-if="video.city"
          class="inline-flex items-center gap-1 rounded-full bg-black/30 px-2.5 py-0.5 text-xs backdrop-blur-sm"
        >
          <i class="fa-solid fa-location-dot text-rose-400" />
          <span class="font-medium text-white/90">{{ video.city }}</span>
        </div>
      </div>

      <!-- 视频描述 - 去掉背景框 -->
      <div class="mb-2.5 max-w-[85%]">
        <p
          class="line-clamp-3 text-[15px] leading-relaxed text-white drop-shadow-[0_2px_6px_rgba(0,0,0,0.9)]"
        >
          <template v-for="(seg, i) in captionSegments">
            <a
              v-if="seg.type === 'topic'"
              :key="`t-${i}`"
              class="cursor-pointer font-semibold text-yellow-300"
              @click.stop="goTopic(seg.value)"
              >#{{ seg.value }}</a
            >
            <span v-else :key="`s-${i}`">{{ seg.value }}</span>
          </template>
          <a
            v-for="topic in extraTopics"
            :key="topic"
            class="ml-1 cursor-pointer font-semibold text-yellow-300"
            @click.stop="goTopic(topic)"
          >
            #{{ topic }}
          </a>
        </p>
      </div>

      <!-- 音乐信息 - 极简透明设计 -->
      <div
        class="inline-flex max-w-[75%] items-center gap-2 rounded-full bg-black/30 px-3 py-1 backdrop-blur-sm"
      >
        <i class="fa-solid fa-music text-xs text-white/80" />
        <div class="overflow-hidden">
          <span
            class="inline-block animate-[marquee_12s_linear_infinite] text-xs font-medium text-white/90"
          >
            {{ video.music }}
          </span>
        </div>
      </div>
    </div>

    <RightAction
      :video="video"
      :playing="playing"
      @follow="emit('follow', video.id)"
      @profile="emit('profile', video.id)"
      @like="emit('like', video.id)"
      @comment="emit('comment', video.id)"
      @favorite="emit('favorite', video.id)"
      @share="emit('share', video.id)"
      @playlist="(anchor) => emit('playlist', video.id, anchor)"
    />
  </article>
</template>
