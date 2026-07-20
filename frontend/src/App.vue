<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Loading as VanLoading } from 'vant'

import BottomNav from '@/components/BottomNav.vue'
import { useAppStore } from '@/store/app'

const route = useRoute()
const appStore = useAppStore()
const showBottomNav = computed(() => !route.meta.hideBottomNav)
</script>

<template>
  <div class="flex h-dvh w-full flex-col overflow-hidden bg-black text-white">
    <div class="relative min-h-0 flex-1 overflow-hidden">
      <RouterView v-slot="{ Component }">
        <component :is="Component" />
      </RouterView>
    </div>
    <BottomNav v-if="showBottomNav" />

    <Transition
      enter-active-class="transition"
      enter-from-class="opacity-0"
      leave-active-class="transition"
      leave-to-class="opacity-0"
    >
      <div
        v-if="appStore.globalLoading"
        class="fixed inset-0 z-[100] flex items-center justify-center bg-black/55 backdrop-blur-sm"
      >
        <VanLoading color="#fe2c55" size="32px" vertical>加载中</VanLoading>
      </div>
    </Transition>
  </div>
</template>
