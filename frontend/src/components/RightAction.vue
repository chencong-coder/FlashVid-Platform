<script setup lang="ts">
import { ref } from 'vue'
import type { VideoItem } from '@/types/video'
import { formatCount } from '@/utils/format'

interface Props {
  video: VideoItem
  playing: boolean
}

interface Emits {
  (event: 'follow'): void
  (event: 'profile'): void
  (event: 'like'): void
  (event: 'comment'): void
  (event: 'favorite'): void
  (event: 'share'): void
  (event: 'playlist'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

const moreOpen = ref(false)
const openPlaylist = () => {
  moreOpen.value = false
  emit('playlist')
}
</script>

<template>
  <aside class="absolute bottom-6 right-3 z-20 flex w-14 flex-col items-center gap-4 text-white">
    <!-- 作者头像 - 更大的圆形设计，带发光效果 -->
    <div class="action-item relative mb-1 h-14 w-14">
      <button
        type="button"
        aria-label="查看用户主页"
        class="h-14 w-14 rounded-full bg-gradient-to-br from-pink-500 to-purple-600 p-0.5 transition-transform duration-300 hover:scale-110 active:scale-95"
        @click.stop="emit('profile')"
      >
        <img
          :src="video.author.avatar"
          :alt="video.author.nickname"
          loading="lazy"
          class="h-full w-full rounded-full border-2 border-black object-cover"
        />
      </button>
      <button
        v-if="!video.author.followed"
        type="button"
        aria-label="关注作者"
        class="absolute -bottom-1 left-1/2 flex h-6 w-6 -translate-x-1/2 items-center justify-center rounded-full bg-gradient-to-br from-pink-500 to-rose-500 text-xs shadow-lg transition-transform duration-300 hover:scale-110"
        @click.stop="emit('follow')"
      >
        <i class="fa-solid fa-plus" />
      </button>
      <span
        v-else
        class="absolute -bottom-1 left-1/2 flex h-6 w-6 -translate-x-1/2 items-center justify-center rounded-full bg-white text-xs text-rose-500 shadow-lg"
      >
        <i class="fa-solid fa-check" />
      </span>
    </div>

    <!-- 点赞按钮 - 带心跳动画 -->
    <button
      type="button"
      :aria-label="video.liked ? '取消点赞' : '点赞'"
      class="action-item flex w-14 flex-col items-center gap-1"
      @click.stop="emit('like')"
    >
      <div class="relative">
        <i
          class="fa-solid fa-heart text-[32px] drop-shadow-2xl transition-all duration-300 active:scale-75"
          :class="
            video.liked
              ? 'animate-heart-beat text-rose-500'
              : 'text-white hover:scale-110 hover:text-rose-200'
          "
        />
        <!-- 点赞后的发光效果 -->
        <div
          v-if="video.liked"
          class="absolute inset-0 -z-10 animate-ping rounded-full bg-rose-500/50 blur-xl"
        />
      </div>
      <span class="text-xs font-semibold text-shadow-strong">{{
        formatCount(video.stats.likes)
      }}</span>
    </button>

    <!-- 评论按钮 -->
    <button
      type="button"
      aria-label="评论"
      class="action-item flex w-14 flex-col items-center gap-1"
      @click.stop="emit('comment')"
    >
      <i
        class="fa-solid fa-comment-dots text-[30px] drop-shadow-lg transition-all duration-300 hover:scale-110 hover:text-violet-300 active:scale-90"
      />
      <span class="text-xs font-semibold text-shadow-strong">{{
        formatCount(video.stats.comments)
      }}</span>
    </button>

    <!-- 收藏按钮 - 带星星闪烁 -->
    <button
      type="button"
      :aria-label="video.favorited ? '取消收藏' : '收藏'"
      class="action-item flex w-14 flex-col items-center gap-1"
      @click.stop="emit('favorite')"
    >
      <div class="relative">
        <i
          class="fa-solid fa-star text-[30px] drop-shadow-2xl transition-all duration-300 active:scale-75"
          :class="
            video.favorited
              ? 'animate-pulse text-amber-400'
              : 'text-white hover:scale-110 hover:text-amber-200'
          "
        />
        <div
          v-if="video.favorited"
          class="absolute inset-0 -z-10 animate-ping rounded-full bg-amber-400/50 blur-xl"
        />
      </div>
      <span class="text-xs font-semibold text-shadow-strong">{{
        formatCount(video.stats.favorites)
      }}</span>
    </button>

    <!-- 分享按钮 -->
    <button
      type="button"
      aria-label="分享"
      class="action-item flex w-14 flex-col items-center gap-1"
      @click.stop="emit('share')"
    >
      <i
        class="fa-solid fa-share text-[30px] drop-shadow-lg transition-all duration-300 hover:scale-110 hover:rotate-12 hover:text-violet-300 active:scale-90"
      />
      <span class="text-xs font-semibold text-shadow-strong">{{
        formatCount(video.stats.shares)
      }}</span>
    </button>

    <!-- 更多操作 -->
    <Teleport to="body">
      <div v-if="moreOpen" class="fixed inset-0 z-[45]" @click="moreOpen = false" />
    </Teleport>
    <div class="relative z-[46]">
      <button
        type="button"
        aria-label="更多"
        class="action-item flex w-14 flex-col items-center gap-1"
        @click.stop="moreOpen = !moreOpen"
      >
        <i
          class="fa-solid fa-ellipsis text-[26px] drop-shadow-lg transition-all duration-300 hover:scale-110 hover:text-violet-300 active:scale-90"
        />
        <span class="text-xs font-semibold text-shadow-strong">更多</span>
      </button>
      <!-- 弹出菜单 -->
      <div
        v-if="moreOpen"
        class="absolute bottom-full right-0 mb-3 w-44 overflow-hidden rounded-2xl bg-[#1a1a22] shadow-2xl ring-1 ring-white/10"
      >
        <button
          type="button"
          class="flex w-full items-center gap-3 px-4 py-3.5 text-sm text-white transition-colors hover:bg-white/10 active:bg-white/20"
          @click.stop="openPlaylist"
        >
          <i class="fa-solid fa-list-ul w-5 text-center text-violet-400" />
          加入播放列表
        </button>
      </div>
    </div>

    <!-- 音乐唱片 - 3D 旋转效果 -->
    <div
      class="mt-2 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-neutral-800 via-neutral-900 to-black p-1.5 shadow-2xl ring-2 ring-white/20 transition-transform duration-300 hover:scale-110"
      :class="playing ? 'animate-disc-spin' : ''"
      style="transform-style: preserve-3d"
    >
      <div class="relative h-full w-full rounded-full">
        <img
          :src="video.discCover"
          alt="音乐封面"
          loading="lazy"
          class="h-full w-full rounded-full object-cover"
        />
        <!-- 中心圆点 -->
        <div
          class="absolute left-1/2 top-1/2 h-2 w-2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white/80"
        />
      </div>
    </div>
  </aside>
</template>
