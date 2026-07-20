<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Switch as VanSwitch,
  Uploader as VanUploader,
  showToast,
  type UploaderFileListItem,
} from 'vant'

const router = useRouter()
const files = ref<UploaderFileListItem[]>([])
const caption = ref('')
const allowComments = ref(true)

const publish = async (): Promise<void> => {
  if (!files.value.length) {
    showToast('请先选择视频')
    return
  }
  showToast('作品已进入发布队列')
  await router.push({ name: 'recommend' })
}
</script>

<template>
  <main class="safe-top no-scrollbar h-full overflow-y-auto bg-[#0d0d0d] px-4 pb-8 text-white">
    <header class="flex h-14 items-center justify-between">
      <button
        type="button"
        aria-label="返回"
        class="h-9 w-9 text-left text-xl"
        @click="router.back()"
      >
        <i class="fa-solid fa-chevron-left" />
      </button>
      <h1 class="text-base font-semibold">发布作品</h1>
      <span class="w-9" />
    </header>

    <section class="mt-4">
      <van-uploader v-model="files" accept="video/*" :max-count="1" :max-size="200 * 1024 * 1024">
        <div
          class="flex aspect-[9/13] w-44 flex-col items-center justify-center rounded-md border border-dashed border-white/30 bg-white/5 text-neutral-400"
        >
          <i class="fa-solid fa-video mb-3 text-3xl" />
          <span class="text-sm font-medium text-white">选择视频</span>
          <span class="mt-2 text-xs">最长 15 分钟</span>
        </div>
      </van-uploader>
    </section>

    <section class="mt-6 border-y border-white/10 py-4">
      <textarea
        v-model="caption"
        rows="4"
        maxlength="200"
        placeholder="添加作品描述，让更多人看见..."
        class="w-full resize-none bg-transparent text-sm leading-6 text-white outline-none placeholder:text-neutral-500"
      />
      <div class="flex gap-4 text-sm font-medium">
        <button type="button"># 话题</button><button type="button">@ 朋友</button>
      </div>
    </section>

    <section class="mt-2 divide-y divide-white/5 text-sm">
      <button type="button" class="flex w-full items-center justify-between py-4">
        <span><i class="fa-solid fa-location-dot mr-3 w-4 text-center" />添加位置</span
        ><i class="fa-solid fa-chevron-right text-xs text-neutral-600" />
      </button>
      <div class="flex items-center justify-between py-4">
        <span><i class="fa-solid fa-comment mr-3 w-4 text-center" />允许评论</span
        ><van-switch v-model="allowComments" size="20px" active-color="#fe2c55" />
      </div>
    </section>

    <button
      type="button"
      class="mt-8 h-12 w-full rounded bg-primary text-sm font-semibold disabled:opacity-50"
      :disabled="!files.length"
      @click="publish"
    >
      发布
    </button>
  </main>
</template>
