<script setup lang="ts">
import { ref, watch } from 'vue'
import { Popup as VanPopup, showToast } from 'vant'

import { getVideoComments, likeComment, postComment, unlikeComment } from '@/api/video'
import { useAuthModalStore } from '@/store/authModal'
import type { CommentItem } from '@/types/video'
import { formatCount } from '@/utils/format'

interface Props {
  show: boolean
  videoId: string
  total: number
}

interface Emits {
  (event: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const authModal = useAuthModalStore()

const content = ref('')
const comments = ref<CommentItem[]>([])
const cursor = ref('')
const hasMore = ref(false)
const loading = ref(false)
const submitting = ref(false)

const fetchComments = async (reset = false): Promise<void> => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getVideoComments(props.videoId, {
      cursor: reset ? undefined : cursor.value || undefined,
      count: 20,
    })
    const page = res.data.data
    if (reset) {
      comments.value = page.comments
    } else {
      comments.value.push(...page.comments)
    }
    cursor.value = page.nextCursorToken
    hasMore.value = page.hasMore
  } catch {
    showToast('加载评论失败')
  } finally {
    loading.value = false
  }
}

const toggleLike = async (comment: CommentItem): Promise<void> => {
  if (!authModal.requireLogin()) return
  const next = !comment.isLiked
  comment.isLiked = next
  comment.likeCount += next ? 1 : -1
  try {
    const res = next ? await likeComment(comment.id) : await unlikeComment(comment.id)
    comment.isLiked = res.data.data.isLiked
    comment.likeCount = res.data.data.likeCount
  } catch {
    comment.isLiked = !next
    comment.likeCount += next ? -1 : 1
  }
}

const submit = async (): Promise<void> => {
  const value = content.value.trim()
  if (!value || submitting.value) return
  if (!authModal.requireLogin()) return
  submitting.value = true
  try {
    const res = await postComment(props.videoId, value)
    // 一级评论返回 comment；此处只发一级评论
    if (res.data.data.comment) comments.value.unshift(res.data.data.comment)
    content.value = ''
  } catch {
    showToast('评论发布失败')
  } finally {
    submitting.value = false
  }
}

// Reload whenever the drawer opens or the target video changes while open
watch(
  () => [props.show, props.videoId] as const,
  ([show]) => {
    if (show) void fetchComments(true)
  },
)
</script>

<template>
  <van-popup
    :show="show"
    position="bottom"
    round
    class="h-[72dvh] overflow-hidden bg-panel text-white"
    @update:show="emit('update:show', $event)"
  >
    <section class="flex h-full flex-col bg-panel">
      <header
        class="relative flex h-12 shrink-0 items-center justify-center border-b border-white/5 text-sm font-semibold"
      >
        {{ formatCount(total) }} 条评论
        <button
          type="button"
          aria-label="关闭"
          class="absolute right-4 text-lg text-neutral-400"
          @click="emit('update:show', false)"
        >
          <i class="fa-solid fa-xmark" />
        </button>
      </header>

      <div class="no-scrollbar flex-1 overflow-y-auto px-4 py-2">
        <!-- Loading skeleton -->
        <div v-if="loading && comments.length === 0" class="flex flex-col gap-4 py-4">
          <div v-for="n in 5" :key="n" class="flex gap-3 animate-pulse">
            <div class="h-9 w-9 shrink-0 rounded-full bg-white/10" />
            <div class="flex-1 space-y-2 pt-1">
              <div class="h-3 w-24 rounded bg-white/10" />
              <div class="h-3 w-full rounded bg-white/10" />
            </div>
          </div>
        </div>

        <article v-for="comment in comments" :key="comment.id" class="flex gap-3 py-3">
          <img
            :src="comment.user.avatar"
            :alt="comment.user.nickname || comment.user.username"
            loading="lazy"
            class="h-9 w-9 shrink-0 rounded-full object-cover"
          />
          <div class="min-w-0 flex-1">
            <div class="text-xs text-neutral-500">
              {{ comment.user.nickname || comment.user.username }}
            </div>
            <p class="mt-1 text-sm leading-5 text-neutral-100">{{ comment.content }}</p>
            <div class="mt-1 text-[11px] text-neutral-500">{{ comment.createdAt }}</div>
          </div>
          <button
            type="button"
            class="flex w-8 flex-col items-center gap-1 transition-colors"
            :class="comment.isLiked ? 'text-red-400' : 'text-neutral-500'"
            @click="void toggleLike(comment)"
          >
            <i :class="comment.isLiked ? 'fa-solid fa-heart' : 'fa-regular fa-heart'" />
            <span class="text-[10px]">{{ comment.likeCount || '' }}</span>
          </button>
        </article>

        <!-- Load more -->
        <div class="py-4 text-center">
          <button
            v-if="hasMore"
            type="button"
            class="text-xs text-neutral-500 hover:text-neutral-300 disabled:opacity-40"
            :disabled="loading"
            @click="void fetchComments(false)"
          >
            {{ loading ? '加载中…' : '加载更多' }}
          </button>
          <span v-else-if="comments.length > 0" class="text-xs text-neutral-600">没有更多了</span>
        </div>
      </div>

      <footer
        class="safe-bottom flex shrink-0 items-center gap-3 border-t border-white/5 bg-panel px-4 pb-3 pt-3"
      >
        <input
          v-model="content"
          type="text"
          placeholder="留下你的精彩评论"
          class="h-10 min-w-0 flex-1 rounded-full bg-white/10 px-4 text-sm text-white outline-none placeholder:text-neutral-500"
          @keyup.enter="void submit()"
        />
        <button
          type="button"
          class="text-sm font-semibold text-primary disabled:opacity-40"
          :disabled="!content.trim() || submitting"
          @click="void submit()"
        >
          发送
        </button>
      </footer>
    </section>
  </van-popup>
</template>
