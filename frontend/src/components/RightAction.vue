<script setup lang="ts">
import type { VideoItem } from '@/types/video'
import { formatCount } from '@/utils/format'

interface Props {
  video: VideoItem
  playing: boolean
}

interface Emits {
  (event: 'follow'): void
  (event: 'like'): void
  (event: 'comment'): void
  (event: 'favorite'): void
  (event: 'share'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()
</script>

<template>
  <aside class="absolute bottom-6 right-3 z-20 flex w-12 flex-col items-center gap-5 text-white">
    <button
      type="button"
      aria-label="关注作者"
      class="relative mb-1 h-12 w-12"
      @click.stop="emit('follow')"
    >
      <img
        :src="video.author.avatar"
        :alt="video.author.nickname"
        loading="lazy"
        class="h-12 w-12 rounded-full border-2 border-white object-cover"
      />
      <span
        v-if="!video.author.followed"
        class="absolute -bottom-2 left-1/2 flex h-5 w-5 -translate-x-1/2 items-center justify-center rounded-full bg-primary text-[10px]"
      >
        <i class="fa-solid fa-plus" />
      </span>
      <span
        v-else
        class="absolute -bottom-2 left-1/2 flex h-5 w-5 -translate-x-1/2 items-center justify-center rounded-full bg-white text-[10px] text-primary"
      >
        <i class="fa-solid fa-check" />
      </span>
    </button>

    <button
      type="button"
      class="flex w-12 flex-col items-center gap-0.5"
      @click.stop="emit('like')"
    >
      <i
        class="fa-solid fa-heart text-[30px] drop-shadow-md transition active:scale-90"
        :class="video.liked ? 'text-primary' : 'text-white'"
      />
      <span class="text-[11px] font-medium text-shadow-video">{{
        formatCount(video.stats.likes)
      }}</span>
    </button>

    <button
      type="button"
      class="flex w-12 flex-col items-center gap-0.5"
      @click.stop="emit('comment')"
    >
      <i class="fa-solid fa-comment-dots text-[28px] drop-shadow-md" />
      <span class="text-[11px] font-medium text-shadow-video">{{
        formatCount(video.stats.comments)
      }}</span>
    </button>

    <button
      type="button"
      class="flex w-12 flex-col items-center gap-0.5"
      @click.stop="emit('favorite')"
    >
      <i
        class="fa-solid fa-star text-[28px] drop-shadow-md transition active:scale-90"
        :class="video.favorited ? 'text-amber-400' : 'text-white'"
      />
      <span class="text-[11px] font-medium text-shadow-video">{{
        formatCount(video.stats.favorites)
      }}</span>
    </button>

    <button
      type="button"
      class="flex w-12 flex-col items-center gap-0.5"
      @click.stop="emit('share')"
    >
      <i class="fa-solid fa-share text-[27px] drop-shadow-md" />
      <span class="text-[11px] font-medium text-shadow-video">{{
        formatCount(video.stats.shares)
      }}</span>
    </button>

    <div
      class="mt-1 flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-neutral-700 via-black to-neutral-700 p-1.5 shadow-overlay"
      :class="playing ? 'animate-disc-spin' : ''"
    >
      <img
        :src="video.discCover"
        alt="音乐封面"
        loading="lazy"
        class="h-full w-full rounded-full object-cover"
      />
    </div>
  </aside>
</template>
