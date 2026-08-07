<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { Popup as VanPopup, showToast } from 'vant'

import {
  deleteComment,
  getCommentReplies,
  getVideoComments,
  likeComment,
  postComment,
  unlikeComment,
} from '@/api/video'
import { useAuthModalStore } from '@/store/authModal'
import { useUserStore } from '@/store/user'
import type { CommentItem, ReplyItem } from '@/types/video'
import { formatCount } from '@/utils/format'

interface Props {
  show: boolean
  videoId: string
  total: number
  // true → 作为 flex 兄弟内联渲染（桌面侧边栏），false → van-popup 底部抽屉（移动端）
  sidebar?: boolean
}

interface Emits {
  (event: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const authModal = useAuthModalStore()
const userStore = useUserStore()

const content = ref('')
const comments = ref<CommentItem[]>([])
const cursor = ref('')
const hasMore = ref(false)
const loading = ref(false)
const submitting = ref(false)

// 展开/收起回复的评论 id 集合
const expanded = ref<Set<string>>(new Set())
// 每条评论独立的回复加载中状态
const loadingReplies = ref<Set<string>>(new Set())

// 输入框 ref（点击「回复」时聚焦）
const inputRef = ref<HTMLInputElement | null>(null)
// 正在回复的目标（parentId + userId + 展示名）
const replyTarget = ref<{ commentId: string; userId: string; name: string } | null>(null)

// sidebar=true → <aside> 内联；sidebar=false → van-popup
const wrapperTag = computed(() => (props.sidebar ? 'aside' : VanPopup))
const wrapperProps = computed(() =>
  props.sidebar
    ? { class: 'flex h-full w-[380px] flex-none flex-col border-l border-white/5 bg-panel text-white' }
    : {
        show: props.show,
        position: 'bottom' as const,
        round: true,
        class: 'h-[72dvh] overflow-hidden bg-panel text-white',
        'onUpdate:show': (v: boolean) => emit('update:show', v),
      },
)

// ── 数据加载 ─────────────────────────────────────────────────────────
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
      expanded.value = new Set()
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

const loadReplies = async (comment: CommentItem): Promise<void> => {
  if (loadingReplies.value.has(comment.id)) return
  loadingReplies.value = new Set([...loadingReplies.value, comment.id])
  try {
    const res = await getCommentReplies(comment.id)
    comment.replies = res.data.data.replies
  } catch {
    showToast('加载回复失败')
  } finally {
    const next = new Set(loadingReplies.value)
    next.delete(comment.id)
    loadingReplies.value = next
  }
}

const toggleExpand = async (comment: CommentItem): Promise<void> => {
  const next = new Set(expanded.value)
  if (next.has(comment.id)) {
    next.delete(comment.id)
  } else {
    next.add(comment.id)
    // 首次展开时才拉取回复（预置的 replies 可能为空或不完整）
    if (comment.replies.length === 0 && comment.replyCount > 0) {
      await loadReplies(comment)
    }
  }
  expanded.value = next
}

// ── 回复交互 ─────────────────────────────────────────────────────────
const startReply = (commentId: string, userId: string, name: string): void => {
  replyTarget.value = { commentId, userId, name }
  void nextTick(() => inputRef.value?.focus())
}

const clearReply = (): void => {
  if (!content.value.trim()) replyTarget.value = null
}

// ── 点赞 ─────────────────────────────────────────────────────────────
const toggleLike = async (target: CommentItem | ReplyItem): Promise<void> => {
  if (!authModal.requireLogin()) return
  const next = !target.isLiked
  target.isLiked = next
  target.likeCount += next ? 1 : -1
  try {
    const res = next ? await likeComment(target.id) : await unlikeComment(target.id)
    target.isLiked = res.data.data.isLiked
    target.likeCount = res.data.data.likeCount
  } catch {
    target.isLiked = !next
    target.likeCount += next ? -1 : 1
  }
}

// ── 删除 ─────────────────────────────────────────────────────────────
const handleDelete = async (comment: CommentItem, reply?: ReplyItem): Promise<void> => {
  const id = reply ? reply.id : comment.id
  try {
    await deleteComment(id)
    if (reply) {
      comment.replies = comment.replies.filter((r: ReplyItem) => r.id !== reply.id)
      comment.replyCount = Math.max(0, comment.replyCount - 1)
    } else {
      comments.value = comments.value.filter((c) => c.id !== comment.id)
    }
  } catch {
    showToast('删除失败')
  }
}

// ── 发送 ─────────────────────────────────────────────────────────────
const submit = async (): Promise<void> => {
  const value = content.value.trim()
  if (!value || submitting.value) return
  if (!authModal.requireLogin()) return
  submitting.value = true
  const rt = replyTarget.value
  try {
    const res = await postComment(
      props.videoId,
      value,
      rt?.commentId ?? '0',
      rt?.userId ?? '0',
    )
    if (rt) {
      // 追加到对应评论的回复列表
      const parent = comments.value.find((c) => c.id === rt.commentId)
      if (parent && res.data.data.reply) {
        parent.replies.push(res.data.data.reply)
        parent.replyCount++
        // 确保该评论已展开
        const next = new Set(expanded.value)
        next.add(rt.commentId)
        expanded.value = next
      }
    } else if (res.data.data.comment) {
      comments.value.unshift(res.data.data.comment)
    }
    content.value = ''
    replyTarget.value = null
  } catch {
    showToast('评论发布失败')
  } finally {
    submitting.value = false
  }
}

watch(
  () => [props.show, props.videoId] as const,
  ([show]) => {
    if (show) void fetchComments(true)
  },
)
</script>

<template>
  <component :is="wrapperTag" v-bind="wrapperProps">
    <section class="flex h-full flex-col bg-panel">
      <!-- 标题栏 -->
      <header
        class="relative flex h-12 shrink-0 items-center justify-center border-b border-white/5 text-sm font-semibold"
      >
        全部评论({{ formatCount(total) }})
        <button
          type="button"
          aria-label="关闭"
          class="absolute right-4 text-lg text-neutral-400 transition-colors hover:text-white"
          @click="emit('update:show', false)"
        >
          <i class="fa-solid fa-xmark" />
        </button>
      </header>

      <!-- 评论列表 -->
      <div class="no-scrollbar flex-1 overflow-y-auto px-4 py-2">
        <!-- 骨架屏 -->
        <div v-if="loading && comments.length === 0" class="flex flex-col gap-4 py-4">
          <div v-for="n in 5" :key="n" class="flex animate-pulse gap-3">
            <div class="h-9 w-9 shrink-0 rounded-full bg-white/10" />
            <div class="flex-1 space-y-2 pt-1">
              <div class="h-3 w-24 rounded bg-white/10" />
              <div class="h-3 w-full rounded bg-white/10" />
              <div class="h-3 w-2/3 rounded bg-white/10" />
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div
          v-else-if="!loading && comments.length === 0"
          class="flex flex-col items-center gap-3 py-16 text-neutral-500"
        >
          <i class="fa-regular fa-comment-dots text-4xl" />
          <span class="text-sm">还没有评论，快来抢沙发</span>
        </div>

        <!-- 评论行 -->
        <article v-for="comment in comments" :key="comment.id" class="flex gap-3 py-3">
          <img
            :src="comment.user.avatar"
            :alt="comment.user.nickname || comment.user.username"
            loading="lazy"
            class="h-9 w-9 shrink-0 rounded-full object-cover"
          />

          <div class="min-w-0 flex-1">
            <!-- 用户名 + 作者标签 -->
            <div class="flex items-center gap-2 text-xs text-neutral-500">
              <span>{{ comment.user.nickname || comment.user.username }}</span>
              <span
                v-if="comment.isAuthored"
                class="rounded bg-primary/20 px-1.5 py-0.5 text-[10px] font-medium text-primary"
              >
                作者
              </span>
            </div>

            <!-- 正文 -->
            <p class="mt-1 text-sm leading-5 text-neutral-100">{{ comment.content }}</p>

            <!-- 时间 + 回复按钮 -->
            <div class="mt-1.5 flex items-center gap-4 text-[11px] text-neutral-500">
              <span>{{ comment.createdAt }}</span>
              <button
                type="button"
                class="transition-colors hover:text-neutral-300"
                @click="startReply(comment.id, comment.user.id, comment.user.nickname || comment.user.username)"
              >
                回复
              </button>
              <button
                v-if="comment.user.id === userStore.profile?.id"
                type="button"
                class="transition-colors hover:text-red-400"
                @click="void handleDelete(comment)"
              >
                删除
              </button>
            </div>

            <!-- 展开/收起回复 -->
            <button
              v-if="comment.replyCount > 0"
              type="button"
              class="mt-2 flex items-center gap-1.5 text-[11px] text-sky-400/80 transition-colors hover:text-sky-300 disabled:opacity-50"
              :disabled="loadingReplies.has(comment.id)"
              @click="void toggleExpand(comment)"
            >
              <span class="h-px w-4 bg-white/20" />
              <span v-if="loadingReplies.has(comment.id)">加载中…</span>
              <template v-else>
                {{
                  expanded.has(comment.id)
                    ? '收起回复'
                    : `展开 ${formatCount(comment.replyCount)} 条回复`
                }}
                <i
                  class="fa-solid text-[9px]"
                  :class="expanded.has(comment.id) ? 'fa-chevron-up' : 'fa-chevron-down'"
                />
              </template>
            </button>

            <!-- 回复列表（展开时显示） -->
            <div
              v-if="expanded.has(comment.id) && comment.replies.length > 0"
              class="mt-2 flex flex-col gap-2.5 rounded-lg bg-white/5 p-2.5"
            >
              <div v-for="reply in comment.replies" :key="reply.id" class="flex gap-2">
                <img
                  :src="reply.user.avatar"
                  :alt="reply.user.nickname || reply.user.username"
                  loading="lazy"
                  class="h-7 w-7 shrink-0 rounded-full object-cover"
                />
                <div class="min-w-0 flex-1">
                  <div class="text-xs text-neutral-500">
                    {{ reply.user.nickname || reply.user.username }}
                  </div>
                  <p class="mt-0.5 text-sm leading-5 text-neutral-200">
                    <span v-if="reply.replyTo?.nickname" class="mr-0.5 text-sky-400/80">
                      回复 @{{ reply.replyTo.nickname }}：
                    </span>{{ reply.content }}
                  </p>
                  <div class="mt-1 flex items-center gap-4 text-[11px] text-neutral-500">
                    <span>{{ reply.createdAt }}</span>
                    <button
                      type="button"
                      class="transition-colors hover:text-neutral-300"
                      @click="startReply(comment.id, reply.user.id, reply.user.nickname || reply.user.username)"
                    >
                      回复
                    </button>
                    <button
                      v-if="reply.user.id === userStore.profile?.id"
                      type="button"
                      class="transition-colors hover:text-red-400"
                      @click="void handleDelete(comment, reply)"
                    >
                      删除
                    </button>
                  </div>
                </div>
                <button
                  type="button"
                  class="flex w-8 flex-col items-center gap-1 pt-1 transition-colors"
                  :class="reply.isLiked ? 'text-red-400' : 'text-neutral-500'"
                  @click="void toggleLike(reply)"
                >
                  <i :class="reply.isLiked ? 'fa-solid fa-heart' : 'fa-regular fa-heart'" />
                  <span class="text-[10px]">{{ reply.likeCount || '' }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- 点赞按钮（右侧） -->
          <button
            type="button"
            class="flex w-8 flex-col items-center gap-1 pt-1 transition-colors"
            :class="comment.isLiked ? 'text-red-400' : 'text-neutral-500'"
            @click="void toggleLike(comment)"
          >
            <i :class="comment.isLiked ? 'fa-solid fa-heart' : 'fa-regular fa-heart'" />
            <span class="text-[10px]">{{ comment.likeCount || '' }}</span>
          </button>
        </article>

        <!-- 加载更多 -->
        <div v-if="comments.length > 0" class="py-4 text-center">
          <button
            v-if="hasMore"
            type="button"
            class="text-xs text-neutral-500 hover:text-neutral-300 disabled:opacity-40"
            :disabled="loading"
            @click="void fetchComments(false)"
          >
            {{ loading ? '加载中…' : '加载更多' }}
          </button>
          <span v-else class="text-xs text-neutral-600">没有更多了</span>
        </div>
      </div>

      <!-- 输入栏 -->
      <footer class="safe-bottom shrink-0 border-t border-white/5 bg-panel">
        <!-- 回复目标提示条 -->
        <div
          v-if="replyTarget"
          class="flex items-center justify-between px-4 pt-2 text-xs text-neutral-400"
        >
          <span>回复 @{{ replyTarget.name }}</span>
          <button
            type="button"
            aria-label="取消回复"
            class="text-neutral-500 hover:text-white"
            @click="replyTarget = null; content = ''"
          >
            <i class="fa-solid fa-xmark" />
          </button>
        </div>
        <div class="flex items-center gap-3 px-4 pb-3 pt-2">
          <input
            ref="inputRef"
            v-model="content"
            type="text"
            :placeholder="replyTarget ? `回复 @${replyTarget.name}` : '留下你的精彩评论'"
            class="h-10 min-w-0 flex-1 rounded-full bg-white/10 px-4 text-sm text-white outline-none placeholder:text-neutral-500"
            @blur="clearReply"
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
        </div>
      </footer>
    </section>
  </component>
</template>
