<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Loading as VanLoading } from 'vant'

import BottomNav from '@/components/BottomNav.vue'
import LeftSidebar from '@/components/layout/LeftSidebar.vue'
import RightPanel from '@/components/layout/RightPanel.vue'
import TopHeader from '@/components/layout/TopHeader.vue'
import { useAppStore } from '@/store/app'

const route = useRoute()
const appStore = useAppStore()
const showBottomNav = computed(() => !route.meta.hideBottomNav)
</script>

<template>
  <div class="flex h-dvh w-full flex-col overflow-hidden bg-[#0d0d12] text-white">
    <!-- Top header (全宽) -->
    <TopHeader />

    <!-- Body: sidebar + content + right panel -->
    <div class="flex flex-1 min-h-0 overflow-hidden">
      <!-- Left sidebar — desktop only -->
      <LeftSidebar class="hidden lg:flex flex-col" />

      <!-- Center content area -->
      <div class="flex-1 min-w-0 relative overflow-hidden">
        <RouterView v-slot="{ Component }">
          <component :is="Component" />
        </RouterView>
      </div>

      <!-- Right panel — large desktop only -->
      <RightPanel class="hidden xl:block" />
    </div>

    <!-- Floating dock nav — always visible (fixed positioned, overlays content) -->
    <BottomNav v-if="showBottomNav" />

    <!-- Global loading overlay -->
    <Transition
      enter-active-class="transition"
      enter-from-class="opacity-0"
      leave-active-class="transition"
      leave-to-class="opacity-0"
    >
      <div
        v-if="appStore.globalLoading"
        class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm"
      >
        <VanLoading color="#7c3aed" size="32px" vertical>加载中</VanLoading>
      </div>
    </Transition>
  </div>
</template>
