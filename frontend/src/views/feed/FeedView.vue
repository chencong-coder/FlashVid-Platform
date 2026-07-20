<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { RecycleScroller } from 'vue-virtual-scroller'

import CommentDrawer from '@/components/CommentDrawer.vue'
import SearchPopup from '@/components/SearchPopup.vue'
import TabHeader from '@/components/TabHeader.vue'
import VideoCard from '@/components/VideoCard.vue'
import { useAppStore } from '@/store/app'
import { useVideoStore } from '@/store/video'
import type { FeedType, VideoItem } from '@/types/video'

interface Props {
  feed: FeedType
}

const props = defineProps<Props>()
const router = useRouter()
const appStore = useAppStore()
const videoStore = useVideoStore()

const feedRoot = ref<HTMLElement | null>(null)
const itemHeight = ref(640)
const currentIndex = ref(0)
const commentVideoId = ref('')
const commentsVisible = ref(false)
const searchVisible = ref(false)
const dragging = ref(false)
let scrollEndTimer: number | undefined
let wheelResetTimer: number | undefined
let wheelUnlockTimer: number | undefined
let wheelDelta = 0
let wheelLocked = false
let wheelTargetIndex = 0
let dragSettleTimer: number | undefined
let dragSettling = false
let dragTargetIndex = 0
let mouseStartY = 0
let mouseStartScrollTop = 0
let mouseStartIndex = 0
let mouseDistance = 0
let mouseHasMoved = false
let suppressClick = false
let resizeObserver: ResizeObserver | undefined
let scrollerElement: HTMLElement | null = null
let touchStartY = 0
let touchStartScrollTop = 0
let touchStartIndex = 0
let touchDistance = 0
let touchHasMoved = false
let touchSettleTimer: number | undefined
let touchSettling = false
let touchTargetIndex = 0

const videos = computed(() => videoStore.feeds[props.feed].items)
const currentVideo = computed(() => videos.value[currentIndex.value])
const commentTotal = computed(() =>
  commentVideoId.value
    ? (videoStore.findVideo(props.feed, commentVideoId.value)?.stats.comments ?? 0)
    : 0,
)

const syncActiveVideo = (index: number): void => {
  const boundedIndex = Math.max(0, Math.min(index, videos.value.length - 1))
  currentIndex.value = boundedIndex
  const video = videos.value[boundedIndex]
  if (video) videoStore.setActiveVideo(video.id)
  if (boundedIndex >= videos.value.length - 2) videoStore.loadMore(props.feed)
}

const getScroller = (): HTMLElement | null =>
  feedRoot.value?.querySelector<HTMLElement>('.feed-scroller') ?? null

const scrollToIndex = (index: number, behavior: ScrollBehavior = 'smooth'): void => {
  const boundedIndex = Math.max(0, Math.min(index, videos.value.length - 1))
  const scroller = getScroller()
  if (!scroller) return
  const targetTop = boundedIndex * itemHeight.value
  if (Math.abs(scroller.scrollTop - targetTop) <= 2) {
    syncActiveVideo(boundedIndex)
    return
  }
  scroller.scrollTo({ top: targetTop, behavior })
}

// 禁用自动吸附功能，避免在滚动动画期间计算出错误的索引导致跳视频
// const snapToNearestVideo = (): void => {
//   const scroller = getScroller()
//   if (!scroller || wheelLocked || dragging.value || dragSettling || touchSettling) return
//   const index = Math.round(scroller.scrollTop / itemHeight.value)
//   console.log('[snapToNearestVideo] scrollTop:', scroller.scrollTop, 'itemHeight:', itemHeight.value, 'calculated index:', index, 'current:', currentIndex.value)
//   const targetTop = index * itemHeight.value
//   if (Math.abs(scroller.scrollTop - targetTop) <= 2) {
//     syncActiveVideo(index)
//     return
//   }
//   scroller.scrollTo({ top: targetTop, behavior: 'smooth' })
// }

const handleScroll = (): void => {
  // 完全禁用自动吸附，只依赖精确的滚轮/拖动/触摸控制
  // 这样可以避免 snapToNearestVideo 在滚动动画期间计算出错误的索引
  return
}

const finishWheelNavigation = (): void => {
  const targetIndex = wheelTargetIndex
  const targetTop = targetIndex * itemHeight.value
  if (scrollerElement && Math.abs(scrollerElement.scrollTop - targetTop) > 2) {
    scrollerElement.scrollTo({ top: targetTop, behavior: 'auto' })
  }
  syncActiveVideo(targetIndex)
  window.clearTimeout(scrollEndTimer)
  wheelLocked = false
}

const handleWheel = (event: WheelEvent): void => {
  event.preventDefault()
  if (Math.abs(event.deltaX) > Math.abs(event.deltaY)) return
  if (dragging.value || dragSettling || touchSettling) return

  if (wheelLocked) {
    // 锁定期间重置 wheelDelta，防止累积触发
    wheelDelta = 0
    return
  }

  wheelDelta += event.deltaY
  window.clearTimeout(wheelResetTimer)
  wheelResetTimer = window.setTimeout(() => {
    wheelDelta = 0
  }, 140)

  if (Math.abs(wheelDelta) < 24) return

  const direction = wheelDelta > 0 ? 1 : -1
  const nextIndex = Math.max(0, Math.min(currentIndex.value + direction, videos.value.length - 1))
  wheelDelta = 0
  if (nextIndex === currentIndex.value) return

  // 立即清除任何待处理的定时器，防止旧的 finishWheelNavigation 触发
  window.clearTimeout(wheelUnlockTimer)
  window.clearTimeout(scrollEndTimer)
  window.clearTimeout(wheelResetTimer)

  wheelLocked = true
  wheelTargetIndex = nextIndex
  scrollToIndex(nextIndex)

  // 使用更长的 1200ms 延迟，确保所有滚轮事件都被过滤
  wheelUnlockTimer = window.setTimeout(finishWheelNavigation, 1200)
}

const isInteractiveTarget = (target: EventTarget | null): boolean => {
  if (!(target instanceof Element)) return false
  if (target.closest('.flash-video-player')) return false
  return Boolean(
    target.closest('button, a, input, textarea, aside, header, .van-popup, .van-overlay'),
  )
}

const removeMouseDragListeners = (): void => {
  window.removeEventListener('mousemove', handleMouseMove)
  window.removeEventListener('mouseup', finishMouseDrag)
  window.removeEventListener('blur', finishMouseDrag)
}

const finishDragNavigation = (): void => {
  const targetTop = dragTargetIndex * itemHeight.value
  if (scrollerElement && Math.abs(scrollerElement.scrollTop - targetTop) > 2) {
    scrollerElement.scrollTo({ top: targetTop, behavior: 'auto' })
  }
  syncActiveVideo(dragTargetIndex)
  // 清除任何待处理的 scroll 定时器
  window.clearTimeout(scrollEndTimer)
  dragSettling = false
}

const settleDragAtIndex = (index: number): void => {
  const boundedIndex = Math.max(0, Math.min(index, videos.value.length - 1))
  const scroller = getScroller()
  if (!scroller) return

  dragSettling = true
  dragTargetIndex = boundedIndex
  window.clearTimeout(scrollEndTimer)
  scroller.scrollTo({ top: boundedIndex * itemHeight.value, behavior: 'smooth' })
  window.clearTimeout(dragSettleTimer)
  dragSettleTimer = window.setTimeout(finishDragNavigation, 560)
}

const finishMouseDrag = (): void => {
  if (!dragging.value) return
  dragging.value = false
  removeMouseDragListeners()

  if (mouseHasMoved) {
    suppressClick = true
    const crossedThreshold = Math.abs(mouseDistance) >= itemHeight.value / 2
    const direction = mouseDistance > 0 ? 1 : -1
    settleDragAtIndex(mouseStartIndex + (crossedThreshold ? direction : 0))
    window.setTimeout(() => (suppressClick = false), 0)
  }
}

const handleMouseMove = (event: MouseEvent): void => {
  if (!dragging.value) return
  const scroller = getScroller()
  if (!scroller) return

  const distance = mouseStartY - event.clientY
  mouseDistance = distance
  if (Math.abs(distance) > 5) mouseHasMoved = true
  if (!mouseHasMoved) return

  event.preventDefault()
  const minTop = Math.max(0, (mouseStartIndex - 1) * itemHeight.value)
  const maxTop = Math.min(
    (videos.value.length - 1) * itemHeight.value,
    (mouseStartIndex + 1) * itemHeight.value,
  )
  scroller.scrollTop = Math.max(minTop, Math.min(mouseStartScrollTop + distance, maxTop))
}

const handleMouseDown = (event: MouseEvent): void => {
  if (
    event.button !== 0 ||
    wheelLocked ||
    dragSettling ||
    touchSettling ||
    isInteractiveTarget(event.target)
  )
    return
  const scroller = getScroller()
  if (!scroller) return

  mouseStartY = event.clientY
  mouseStartScrollTop = scroller.scrollTop
  mouseStartIndex = currentIndex.value
  mouseDistance = 0
  mouseHasMoved = false
  dragging.value = true
  window.addEventListener('mousemove', handleMouseMove, { passive: false })
  window.addEventListener('mouseup', finishMouseDrag, { once: true })
  window.addEventListener('blur', finishMouseDrag, { once: true })
}

const handleClickCapture = (event: MouseEvent): void => {
  if (!suppressClick) return
  event.preventDefault()
  event.stopImmediatePropagation()
  suppressClick = false
}

const finishTouchNavigation = (): void => {
  const targetTop = touchTargetIndex * itemHeight.value
  if (scrollerElement && Math.abs(scrollerElement.scrollTop - targetTop) > 2) {
    scrollerElement.scrollTo({ top: targetTop, behavior: 'auto' })
  }
  syncActiveVideo(touchTargetIndex)
  // 清除任何待处理的 scroll 定时器
  window.clearTimeout(scrollEndTimer)
  touchSettling = false
}

const settleTouchAtIndex = (index: number): void => {
  const boundedIndex = Math.max(0, Math.min(index, videos.value.length - 1))
  const scroller = getScroller()
  if (!scroller) return

  touchSettling = true
  touchTargetIndex = boundedIndex
  window.clearTimeout(scrollEndTimer)
  scroller.scrollTo({ top: boundedIndex * itemHeight.value, behavior: 'smooth' })
  window.clearTimeout(touchSettleTimer)
  touchSettleTimer = window.setTimeout(finishTouchNavigation, 560)
}

const handleTouchStart = (event: TouchEvent): void => {
  if (wheelLocked || dragSettling || touchSettling || isInteractiveTarget(event.target)) return
  const scroller = getScroller()
  const touch = event.touches[0]
  if (!scroller || !touch) return

  touchStartY = touch.clientY
  touchStartScrollTop = scroller.scrollTop
  touchStartIndex = currentIndex.value
  touchDistance = 0
  touchHasMoved = false
}

const handleTouchMove = (event: TouchEvent): void => {
  const scroller = getScroller()
  const touch = event.touches[0]
  if (!scroller || !touch) return

  const currentY = touch.clientY
  const distance = touchStartY - currentY
  touchDistance = distance

  if (Math.abs(distance) > 10) {
    touchHasMoved = true
    event.preventDefault()
  }

  if (!touchHasMoved) return

  // 严格限制只能滚动到相邻视频，不允许跨越多个视频
  const maxDistance = itemHeight.value * 1.5
  const clampedDistance = Math.max(-maxDistance, Math.min(distance, maxDistance))
  const targetScrollTop = touchStartScrollTop + clampedDistance

  scroller.scrollTop = targetScrollTop
}

const handleTouchEnd = (): void => {
  if (!touchHasMoved) return

  suppressClick = true
  const crossedThreshold = Math.abs(touchDistance) >= itemHeight.value / 2
  const direction = touchDistance > 0 ? 1 : -1
  settleTouchAtIndex(touchStartIndex + (crossedThreshold ? direction : 0))
  window.setTimeout(() => (suppressClick = false), 0)

  touchHasMoved = false
  touchDistance = 0
}

const bindScrollerEvents = (): void => {
  scrollerElement = getScroller()
  scrollerElement?.addEventListener('wheel', handleWheel, { passive: false })
  scrollerElement?.addEventListener('mousedown', handleMouseDown)
  scrollerElement?.addEventListener('touchstart', handleTouchStart, { passive: true })
  scrollerElement?.addEventListener('touchmove', handleTouchMove, { passive: false })
  scrollerElement?.addEventListener('touchend', handleTouchEnd, { passive: true })
}

const unbindScrollerEvents = (): void => {
  scrollerElement?.removeEventListener('wheel', handleWheel)
  scrollerElement?.removeEventListener('mousedown', handleMouseDown)
  scrollerElement?.removeEventListener('touchstart', handleTouchStart)
  scrollerElement?.removeEventListener('touchmove', handleTouchMove)
  scrollerElement?.removeEventListener('touchend', handleTouchEnd)
  scrollerElement = null
}

const switchFeed = async (feed: FeedType): Promise<void> => {
  appStore.setTopTab(feed)
  const routes: Record<FeedType, string> = {
    recommend: 'recommend',
    follow: 'follow',
    nearby: 'nearby',
  }
  await router.replace({ name: routes[feed] })
}

const openComments = (videoId: string): void => {
  commentVideoId.value = videoId
  commentsVisible.value = true
}

const share = async (): Promise<void> => {
  if (navigator.share && currentVideo.value) {
    await navigator
      .share({
        title: `闪视 · ${currentVideo.value.author.nickname}`,
        text: currentVideo.value.description,
        url: window.location.href,
      })
      .catch(() => undefined)
    return
  }
  await navigator.clipboard?.writeText(window.location.href)
  showToast('分享链接已复制')
}

onMounted(async () => {
  await nextTick()
  bindScrollerEvents()
  if (feedRoot.value) {
    const setHeight = (): void => {
      itemHeight.value = feedRoot.value?.clientHeight || 640
      scrollToIndex(currentIndex.value, 'auto')
    }
    setHeight()
    resizeObserver = new ResizeObserver(setHeight)
    resizeObserver.observe(feedRoot.value)
  }
  appStore.setTopTab(props.feed)
  syncActiveVideo(0)
})

watch(
  () => props.feed,
  () => syncActiveVideo(0),
)

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  unbindScrollerEvents()
  window.clearTimeout(scrollEndTimer)
  window.clearTimeout(wheelResetTimer)
  window.clearTimeout(wheelUnlockTimer)
  window.clearTimeout(dragSettleTimer)
  window.clearTimeout(touchSettleTimer)
  removeMouseDragListeners()
  dragging.value = false
  dragSettling = false
  touchSettling = false
  videoStore.setActiveVideo('')
})
</script>

<template>
  <main
    ref="feedRoot"
    class="relative h-full w-full overflow-hidden bg-black"
    :class="dragging ? 'cursor-grabbing' : 'cursor-grab'"
    @click.capture="handleClickCapture"
    @dragstart.prevent
  >
    <RecycleScroller
      class="feed-scroller h-full w-full"
      :items="videos"
      :item-size="itemHeight"
      key-field="id"
      :buffer="itemHeight"
      @scroll="handleScroll"
    >
      <template #default="{ item, index }: { item: VideoItem; index: number }">
        <VideoCard
          :video="item"
          :active="index === currentIndex"
          :muted="videoStore.muted"
          @follow="videoStore.toggleFollow(feed, $event)"
          @like="videoStore.toggleLike(feed, $event)"
          @comment="openComments"
          @favorite="videoStore.toggleFavorite(feed, $event)"
          @share="share"
        />
      </template>
    </RecycleScroller>

    <TabHeader :active="feed" @change="switchFeed" @search="searchVisible = true" />
    <button
      type="button"
      :aria-label="videoStore.muted ? '打开声音' : '静音'"
      class="safe-top absolute left-3 top-3 z-40 flex h-8 w-8 items-center justify-center rounded-full bg-black/30 text-xs text-white backdrop-blur-sm"
      @click="videoStore.toggleMuted"
    >
      <i class="fa-solid" :class="videoStore.muted ? 'fa-volume-xmark' : 'fa-volume-high'" />
    </button>

    <CommentDrawer v-model:show="commentsVisible" :total="commentTotal" />
    <SearchPopup v-model:show="searchVisible" />
  </main>
</template>
