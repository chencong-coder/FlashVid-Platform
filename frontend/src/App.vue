<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Loading as VanLoading } from 'vant'

import BottomNav from '@/components/BottomNav.vue'
import LeftSidebar from '@/components/layout/LeftSidebar.vue'
import RightPanel from '@/components/layout/RightPanel.vue'
import TopHeader from '@/components/layout/TopHeader.vue'
import AuthModal from '@/components/AuthModal.vue'
import { useAppStore } from '@/store/app'
import { useAuthModalStore } from '@/store/authModal'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authModal = useAuthModalStore()
const showBottomNav = computed(() => !route.meta.hideBottomNav)

// 拦截 /login 和 /register 路由 → 弹出 Modal 而不是整页跳转
watch(
  route,
  (to) => {
    if (to.name === 'login') {
      const redirect = to.query.redirect as string | undefined
      authModal.open('prompt', redirect)
      void router.replace({ name: 'profile' })
    } else if (to.name === 'register') {
      authModal.open('prompt')
      void router.replace({ name: 'profile' })
    }
  },
  { immediate: true },
)
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

    <!-- 登录/注册弹窗（全局单例） -->
    <AuthModal />
  </div>
</template>
