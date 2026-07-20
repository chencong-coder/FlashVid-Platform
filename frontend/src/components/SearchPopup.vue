<script setup lang="ts">
import { computed, ref } from 'vue'
import { Popup as VanPopup } from 'vant'

interface Props {
  show: boolean
}

interface Emits {
  (event: 'update:show', value: boolean): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()
const keyword = ref('')

const hotWords = ['夏日旅行', '城市夜景', '一口气看完', '松弛感生活', '今日穿搭']
const results = computed(() =>
  keyword.value.trim()
    ? [`搜索“${keyword.value.trim()}”相关视频`, `寻找用户 @${keyword.value.trim()}`]
    : [],
)
</script>

<template>
  <van-popup
    :show="show"
    position="top"
    class="h-full bg-black text-white"
    @update:show="emit('update:show', $event)"
  >
    <div class="safe-top flex h-full flex-col bg-black px-4 pt-3">
      <div class="flex items-center gap-3">
        <div class="flex h-10 min-w-0 flex-1 items-center gap-2 rounded bg-white/10 px-3">
          <i class="fa-solid fa-magnifying-glass text-sm text-neutral-400" />
          <input
            v-model="keyword"
            autofocus
            type="search"
            placeholder="搜索你感兴趣的内容"
            class="min-w-0 flex-1 bg-transparent text-sm text-white outline-none placeholder:text-neutral-500"
          />
          <button
            v-if="keyword"
            type="button"
            aria-label="清空"
            class="text-neutral-500"
            @click="keyword = ''"
          >
            <i class="fa-solid fa-circle-xmark" />
          </button>
        </div>
        <button type="button" class="shrink-0 text-sm" @click="emit('update:show', false)">
          取消
        </button>
      </div>

      <div v-if="!keyword" class="pt-7">
        <h2 class="text-sm font-semibold">闪视热榜</h2>
        <button
          v-for="(word, index) in hotWords"
          :key="word"
          type="button"
          class="flex w-full items-center gap-3 border-b border-white/5 py-4 text-left text-sm"
          @click="keyword = word"
        >
          <span
            class="w-5 text-center font-bold"
            :class="index < 3 ? 'text-primary' : 'text-neutral-500'"
            >{{ index + 1 }}</span
          >
          <span>{{ word }}</span>
          <span v-if="index < 2" class="rounded-sm bg-primary/20 px-1 text-[10px] text-primary"
            >热</span
          >
        </button>
      </div>

      <div v-else class="pt-4">
        <button
          v-for="result in results"
          :key="result"
          type="button"
          class="flex w-full items-center gap-3 border-b border-white/5 py-4 text-left text-sm"
        >
          <i class="fa-solid fa-magnifying-glass text-neutral-500" />
          {{ result }}
        </button>
      </div>
    </div>
  </van-popup>
</template>
