<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { RecycleScroller } from 'vue-virtual-scroller'

import AddToPlaylistModal from '@/components/AddToPlaylistModal.vue'
import CommentDrawer from '@/components/CommentDrawer.vue'
import SearchPopup from '@/components/SearchPopup.vue'
import TabHeader from '@/components/TabHeader.vue'
import VideoCard from '@/components/VideoCard.vue'
import { useAppStore } from '@/store/app'
import { useAuthModalStore } from '@/store/authModal'
import { useVideoStore } from '@/store/video'
import type { FeedType, TopNavValue, VideoItem } from '@/types/video'

interface Props {
  feed: FeedType
  // friends 等独立流不显示顶部 tab 切换栏
  showTabs?: boolean
  // 进入时定位到的起始视频索引（话题流点封面进入用）
  startIndex?: number
}

const props = withDefaults(defineProps<Props>(), {
  showTabs: true,
  startIndex: 0,
})
const router = useRouter()
const appStore = useAppStore()
const authModal = useAuthModalStore()
const videoStore = useVideoStore()

const feedRoot = ref<HTMLElement | null>(null)
const itemHeight = ref(640)
const currentIndex = ref(0)
const commentVideoId = ref('')
const commentsVisible = ref(false)
const searchVisible = ref(false)
const playlistVideoId = ref<string | null>(null)
const dragging = ref(false)
const wheelLocked = ref(false)
let scrollEndTimer: number | undefined
let wheelResetTimer: number | undefined
let wheelUnlockTimer: number | undefined
let wheelDelta = 0
let wheelTargetIndex = 0
let lastWheelTime = 0 // 上次滚轮事件的时间
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

// TabHeader 只在带 tab 的常规流（follow/recommend/nearby）渲染，
// 'topic'/'profile' 属于独立流不会走到这里；收窄类型以匹配 TopNavValue
const activeTab = computed<TopNavValue>(() =>
  props.feed === 'topic' || props.feed === 'profile' ? 'recommend' : props.feed,
)

const syncActiveVideo = (index: number): void => {
  const boundedIndex = Math.max(0, Math.min(index, videos.value.length - 1))
  currentIndex.value = boundedIndex
  const video = videos.value[boundedIndex]
  if (video) videoStore.setActiveVideo(video.id)
  if (boundedIndex >= videos.value.length - 2) void videoStore.loadMore(props.feed)
}

// 获取定位（附近流需要），失败则跳过
const ensureLocation = async (): Promise<void> => {
  if (props.feed !== 'nearby' || videoStore.location) return
  if (!navigator.geolocation) return
  await new Promise<void>((resolve) => {
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        videoStore.setLocation({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
        })
        resolve()
      },
      () => resolve(),
      { timeout: 5000 },
    )
  })
}

// 初始化当前 feed 数据并定位到起始条
const initCurrentFeed = async (): Promise<void> => {
  await ensureLocation()
  await videoStore.initFeed(props.feed)
  await nextTick()
  const start = Math.max(0, Math.min(props.startIndex, videos.value.length - 1))
  currentIndex.value = start
  scrollToIndex(start, 'auto')
  syncActiveVideo(start)
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
  wheelLocked.value = false
}

const handleWheel = (event: WheelEvent): void => {
  event.preventDefault()
  if (Math.abs(event.deltaX) > Math.abs(event.deltaY)) return
  if (dragging.value || dragSettling || touchSettling) return

  const now = Date.now()
  const timeSinceLastWheel = now - lastWheelTime

  // 如果距离上次滚轮事件小于 300ms，忽略（防止连续滚动）
  if (timeSinceLastWheel < 300 && lastWheelTime > 0) {
    return
  }

  if (wheelLocked.value) {
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

  // 记录本次滚轮事件时间
  lastWheelTime = now

  // 立即清除任何待处理的定时器，防止旧的 finishWheelNavigation 触发
  window.clearTimeout(wheelUnlockTimer)
  window.clearTimeout(scrollEndTimer)
  window.clearTimeout(wheelResetTimer)

  wheelLocked.value = true
  wheelTargetIndex = nextIndex
  scrollToIndex(nextIndex)

  // 使用 1500ms 延迟，在滚动动画完成和用户可能的后续滚动之间取得平衡
  wheelUnlockTimer = window.setTimeout(finishWheelNavigation, 1500)
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
    wheelLocked.value ||
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
  if (wheelLocked.value || dragSettling || touchSettling || isInteractiveTarget(event.target))
    return
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

const switchFeed = async (value: TopNavValue): Promise<void> => {
  if (value === 'discover') {
    await router.push({ name: 'discover' })
    return
  }
  const routes: Record<Exclude<TopNavValue, 'discover'>, string> = {
    recommend: 'recommend',
    follow: 'follow',
    nearby: 'nearby',
    friends: 'friends',
  }
  const target = { name: routes[value] }
  if (
    (value === 'follow' || value === 'friends') &&
    !authModal.requireLogin(router.resolve(target).fullPath)
  ) {
    return
  }
  appStore.setTopTab(value)
  await router.replace(target)
}

const requireInteractionLogin = (): boolean =>
  authModal.requireLogin(router.currentRoute.value.fullPath)

const toggleFollow = (videoId: string): void => {
  if (!requireInteractionLogin()) return
  void videoStore.toggleFollow(props.feed, videoId)
}

const goToUserProfile = (videoId: string): void => {
  const video = videoStore.findVideo(props.feed, videoId)
  if (!video) return
  void router.push({ name: 'user-profile', params: { id: video.author.id } })
}

const toggleLike = (videoId: string): void => {
  if (!requireInteractionLogin()) return
  void videoStore.toggleLike(props.feed, videoId)
}

const toggleFavorite = (videoId: string): void => {
  if (!requireInteractionLogin()) return
  void videoStore.toggleFavorite(props.feed, videoId)
}

const openComments = (videoId: string): void => {
  commentVideoId.value = videoId
  commentsVisible.value = true
}

const handlePlaylist = (videoId: string): void => {
  if (!requireInteractionLogin()) return
  playlistVideoId.value = videoId
}

const share = async (): Promise<void> => {
  if (!currentVideo.value) return
  if (!requireInteractionLogin()) return
  const shareUrl = await videoStore.shareCurrent(props.feed, currentVideo.value.id)
  if (navigator.share) {
    await navigator
      .share({
        title: `闪视 · ${currentVideo.value.author.nickname}`,
        text: currentVideo.value.description,
        url: shareUrl,
      })
      .catch(() => undefined)
    return
  }
  await navigator.clipboard?.writeText(shareUrl)
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
  await initCurrentFeed()
})

watch(
  () => props.feed,
  () => {
    void initCurrentFeed()
  },
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
          @follow="toggleFollow"
          @profile="goToUserProfile"
          @like="toggleLike"
          @comment="openComments"
          @favorite="toggleFavorite"
          @share="share"
          @playlist="handlePlaylist"
        />
      </template>
    </RecycleScroller>

    <!-- 常规 feed：顶部 tab 切换栏 -->
    <TabHeader
      v-if="showTabs"
      :active="activeTab"
      @change="switchFeed"
      @search="searchVisible = true"
    />

    <!-- 独立流（朋友/话题/个人主页）：仅显示标题栏或返回按钮 -->
    <header
      v-else
      class="safe-top pointer-events-none absolute inset-x-0 top-0 z-30 flex items-center px-4 pb-6 pt-3"
      :class="feed === 'profile' ? 'justify-start' : 'justify-center'"
    >
      <div class="absolute inset-0 bg-gradient-to-b from-black/80 via-black/40 to-transparent" />
      <!-- 个人主页流：左上角返回按钮 -->
      <button
        v-if="feed === 'profile'"
        type="button"
        aria-label="返回"
        class="pointer-events-auto relative flex h-9 w-9 items-center justify-center rounded-full bg-black/30 text-white backdrop-blur-sm"
        @click="router.back()"
      >
        <i class="fa-solid fa-chevron-left text-sm" />
      </button>
      <!-- 朋友流：居中标签 -->
      <div
        v-else
        class="pointer-events-auto relative flex items-center gap-2 rounded-full bg-white/10 px-5 py-2 text-sm font-semibold text-white backdrop-blur-xl"
      >
        <i class="fa-solid fa-user-group text-xs" />
        朋友
      </div>
    </header>

    <!-- 静音按钮 - 更现代的设计 -->
    <button
      type="button"
      :aria-label="videoStore.muted ? '打开声音' : '静音'"
      class="safe-top absolute top-4 z-40 flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-white/20 to-white/5 text-base text-white backdrop-blur-xl transition-all duration-300 hover:scale-110 hover:from-white/30 hover:to-white/10 active:scale-95"
      :class="feed === 'profile' ? 'right-4' : 'left-4'"
      @click="videoStore.toggleMuted"
    >
      <i
        class="fa-solid transition-transform duration-300"
        :class="[
          videoStore.muted ? 'fa-volume-xmark' : 'fa-volume-high',
          videoStore.muted ? '' : 'animate-pulse',
        ]"
      />
    </button>

    <CommentDrawer
      v-model:show="commentsVisible"
      :video-id="commentVideoId"
      :total="commentTotal"
    />
    <SearchPopup v-model:show="searchVisible" />
    <AddToPlaylistModal :video-id="playlistVideoId" @close="playlistVideoId = null" />
  </main>
</template>
