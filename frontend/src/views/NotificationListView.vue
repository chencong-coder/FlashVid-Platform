<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getNotifications, markAsRead, type NotificationItem } from '@/api/notification'
import { useNotificationStore } from '@/store/notification'

const route = useRoute()
const router = useRouter()
const notifStore = useNotificationStore()

// 路由类型 → 标题 & actionTypes 映射
const TYPE_CONFIG: Record<string, { title: string; actionTypes: number[]; countKey: 'followers' | 'likesAndFavs' | 'mentions' | 'comments' }> = {
  followers:   { title: '新增粉丝',   actionTypes: [1],    countKey: 'followers' },
  'likes-favs':{ title: '赞和收藏',   actionTypes: [2, 3], countKey: 'likesAndFavs' },
  mentions:    { title: '@我的',       actionTypes: [6],    countKey: 'mentions' },
  comments:    { title: '评论',        actionTypes: [4, 5], countKey: 'comments' },
}

const typeKey = computed(() => String(route.params.type))
const config   = computed(() => TYPE_CONFIG[typeKey.value] ?? TYPE_CONFIG.followers)
const title    = computed(() => config.value.title)

const list      = ref<NotificationItem[]>([])
const total     = ref(0)
const page      = ref(1)
const loading   = ref(false)
const loadedAll = ref(false)

async function load(reset = false) {
  if (loading.value) return
  if (!reset && loadedAll.value) return
  loading.value = true
  if (reset) {
    page.value = 1
    list.value = []
    loadedAll.value = false
  }
  try {
    const res = await getNotifications({
      action_types: config.value.actionTypes,
      page: page.value,
      page_size: 20,
    })
    const data = res.data.data
    if (data) {
      list.value.push(...data.list)
      total.value = data.total
      if (list.value.length >= data.total) loadedAll.value = true
      else page.value++
    }
  } catch {
    // 静默失败
  } finally {
    loading.value = false
  }
}

async function markAndRefresh() {
  try {
    await markAsRead(config.value.actionTypes)
    notifStore.resetTypeCount(config.value.countKey)
  } catch {
    // 静默失败
  }
}

function actionLabel(item: NotificationItem): string {
  switch (item.actionType) {
    case 1: return '关注了你'
    case 2: return '赞了你的视频'
    case 3: return '收藏了你的视频'
    case 4: return '评论了你的视频'
    case 5: return '回复了你的评论'
    case 6: return '@了你'
    default: return '通知了你'
  }
}

function handleItemClick(item: NotificationItem) {
  if (item.actionType === 1) {
    void router.push(`/profile/${item.actorId}`)
  } else {
    void router.push(`/video/${item.targetId}`)
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffDays = Math.floor(diffMs / 86_400_000)
  if (diffDays === 0)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  if (diffDays === 1) return '昨天'
  if (diffDays < 7)
    return `星期${['日', '一', '二', '三', '四', '五', '六'][d.getDay()]}`
  return `${d.getMonth() + 1}/${d.getDate()}`
}

// 切换 type 时重新加载
watch(typeKey, () => {
  void load(true)
  void markAndRefresh()
})

onMounted(() => {
  void load(true)
  void markAndRefresh()
})
</script>

<template>
  <main class="safe-top flex h-full flex-col bg-[#0d0d0d] text-white">
    <!-- 顶部导航 -->
    <header class="flex h-14 shrink-0 items-center gap-3 border-b border-white/5 px-4">
      <button type="button" class="text-xl" @click="router.back()">
        <i class="fa-solid fa-arrow-left" />
      </button>
      <h1 class="flex-1 text-center text-base font-semibold">{{ title }}</h1>
      <!-- 占位，保持标题居中 -->
      <span class="w-6" />
    </header>

    <!-- 列表 -->
    <section class="flex-1 overflow-y-auto">
      <!-- 加载中（首次） -->
      <div v-if="loading && list.length === 0" class="flex justify-center py-16">
        <i class="fa-solid fa-circle-notch animate-spin text-2xl text-neutral-500" />
      </div>

      <!-- 空状态 -->
      <div
        v-else-if="!loading && list.length === 0"
        class="flex flex-col items-center gap-3 py-20 text-neutral-500"
      >
        <i class="fa-regular fa-bell text-4xl" />
        <p class="text-sm">暂无通知</p>
      </div>

      <!-- 通知项 -->
      <template v-else>
        <article
          v-for="item in list"
          :key="item.id"
          class="flex cursor-pointer items-start gap-3 px-4 py-3 active:bg-white/5"
          :class="item.isRead === 0 ? 'bg-white/[0.02]' : ''"
          @click="handleItemClick(item)"
        >
          <!-- 头像 -->
          <img
            :src="item.actorAvatar || 'https://placehold.co/40x40?text=?'"
            :alt="item.actorName"
            class="h-11 w-11 shrink-0 rounded-full object-cover"
          />
          <!-- 正文 -->
          <div class="min-w-0 flex-1">
            <div class="flex items-baseline justify-between gap-2">
              <span class="truncate text-sm font-medium">{{ item.actorName }}</span>
              <time class="shrink-0 text-[10px] text-neutral-600">{{ formatTime(item.createdAt) }}</time>
            </div>
            <p class="mt-0.5 text-xs text-neutral-400">{{ actionLabel(item) }}</p>
            <!-- 评论/回复内容预览 -->
            <p v-if="item.content" class="mt-1 truncate text-xs text-neutral-500">
              "{{ item.content }}"
            </p>
          </div>
          <!-- 视频封面（仅非关注类） -->
          <img
            v-if="item.targetCover && item.actionType !== 1"
            :src="item.targetCover"
            alt=""
            class="h-12 w-9 shrink-0 rounded object-cover"
          />
          <!-- 未读红点 -->
          <span
            v-if="item.isRead === 0"
            class="mt-2 h-2 w-2 shrink-0 rounded-full bg-red-500"
          />
        </article>

        <!-- 底部加载更多 -->
        <div class="flex justify-center py-4">
          <button
            v-if="!loadedAll"
            type="button"
            class="text-sm text-primary disabled:opacity-50"
            :disabled="loading"
            @click="load()"
          >
            <i v-if="loading" class="fa-solid fa-circle-notch animate-spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-xs text-neutral-600">已加载全部</p>
        </div>
      </template>
    </section>
  </main>
</template>
