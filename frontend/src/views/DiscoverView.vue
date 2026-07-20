<script setup lang="ts">
import { ref } from 'vue'

interface DiscoverItem {
  id: number
  title: string
  cover: string
  views: string
  tag: string
}

const keyword = ref('')
const categories = ['直播', '影视', '音乐', '游戏', '美食', '旅行']
const items: DiscoverItem[] = [
  {
    id: 1,
    title: '在城市里寻找夏天',
    cover:
      'https://images.unsplash.com/photo-1494526585095-c41746248156?q=80&w=800&auto=format&fit=crop',
    views: '128.6万',
    tag: '城市漫游',
  },
  {
    id: 2,
    title: '把日落装进镜头',
    cover:
      'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?q=80&w=800&auto=format&fit=crop',
    views: '76.3万',
    tag: '旅行',
  },
  {
    id: 3,
    title: '今晚的现场太浪漫',
    cover:
      'https://images.unsplash.com/photo-1492684223066-81342ee5ff30?q=80&w=800&auto=format&fit=crop',
    views: '42.9万',
    tag: '音乐现场',
  },
  {
    id: 4,
    title: '周末咖啡店打卡',
    cover:
      'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?q=80&w=800&auto=format&fit=crop',
    views: '31.5万',
    tag: '探店',
  },
]
</script>

<template>
  <main class="no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] text-white">
    <header class="safe-top sticky top-0 z-20 bg-[#0d0d0d]/95 px-4 pb-3 pt-3 backdrop-blur">
      <div class="flex h-10 items-center gap-3 rounded bg-white/10 px-3 text-neutral-400">
        <i class="fa-solid fa-magnifying-glass text-sm" />
        <input
          v-model="keyword"
          type="search"
          placeholder="搜索视频、用户和话题"
          class="min-w-0 flex-1 bg-transparent text-sm text-white outline-none placeholder:text-neutral-500"
        />
      </div>
      <div class="no-scrollbar mt-4 flex gap-5 overflow-x-auto whitespace-nowrap text-sm">
        <button
          v-for="(category, index) in categories"
          :key="category"
          type="button"
          :class="index === 0 ? 'font-semibold text-white' : 'text-neutral-500'"
        >
          {{ category }}
        </button>
      </div>
    </header>

    <section class="px-3 pb-6">
      <div class="mb-3 flex items-center justify-between">
        <h1 class="text-lg font-bold">发现精彩</h1>
        <button type="button" class="text-xs text-neutral-500">
          换一换 <i class="fa-solid fa-rotate ml-1" />
        </button>
      </div>
      <div class="grid grid-cols-2 gap-1.5">
        <article
          v-for="item in items"
          :key="item.id"
          class="relative aspect-[3/4] overflow-hidden bg-neutral-900"
        >
          <img
            :src="item.cover"
            :alt="item.title"
            loading="lazy"
            class="h-full w-full object-cover"
          />
          <div
            class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/90 to-transparent px-2 pb-2 pt-10"
          >
            <p class="line-clamp-2 text-sm font-medium leading-5">{{ item.title }}</p>
            <div class="mt-1 flex items-center justify-between text-[10px] text-neutral-300">
              <span># {{ item.tag }}</span
              ><span><i class="fa-solid fa-play mr-1" />{{ item.views }}</span>
            </div>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
