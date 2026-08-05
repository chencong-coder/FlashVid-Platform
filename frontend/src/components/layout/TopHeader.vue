<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import SearchPopup from '@/components/SearchPopup.vue'
import SearchDropdown from '@/components/SearchDropdown.vue'
import { useUserStore } from '@/store/user'
import { getUnreadCount } from '@/api/message'

const router = useRouter()
const userStore = useUserStore()
const searchVisible = ref(false)
const unreadCount = ref(0)

const loadUnreadCount = async () => {
  if (!userStore.isLoggedIn) { unreadCount.value = 0; return }
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data.data.unreadCount
  } catch {
    unreadCount.value = 0
  }
}

onMounted(loadUnreadCount)
watch(() => userStore.isLoggedIn, loadUnreadCount)
</script>

<template>
  <header
    class="h-16 shrink-0 flex items-center gap-4 px-5 bg-[#0d0d10] border-b border-white/[0.06] z-50"
  >
    <!-- Logo (mobile only — desktop sees it in sidebar) -->
    <div class="lg:hidden shrink-0">
      <span
        class="text-xl font-black bg-gradient-to-r from-violet-400 via-purple-300 to-fuchsia-300 bg-clip-text text-transparent tracking-tight"
      >
        闪视
      </span>
    </div>

    <!-- Search dropdown (desktop only — inline with dropdown panel) -->
    <div class="hidden lg:flex flex-1 justify-center">
      <SearchDropdown />
    </div>

    <!-- Right actions -->
    <div class="flex items-center gap-2 ml-auto shrink-0">
      <!-- Mobile search icon -->
      <button
        type="button"
        class="lg:hidden w-9 h-9 flex items-center justify-center rounded-full hover:bg-white/5 text-gray-400 hover:text-white transition-colors"
        @click="searchVisible = true"
      >
        <i class="fa-solid fa-magnifying-glass" />
      </button>

      <!-- Notification bell with badge -->
      <button
        type="button"
        aria-label="消息通知"
        class="relative w-10 h-10 flex items-center justify-center rounded-full hover:bg-white/5 text-gray-300 hover:text-white transition-colors"
        @click="router.push({ name: 'messages' })"
      >
        <i class="fa-regular fa-bell text-[18px]" />
        <span
          v-if="userStore.isLoggedIn && unreadCount > 0"
          class="absolute top-1.5 right-1.5 w-[18px] h-[18px] flex items-center justify-center rounded-full bg-violet-600 text-[9px] font-bold text-white leading-none"
          >{{ unreadCount > 99 ? '99+' : unreadCount }}</span
        >
      </button>

      <!-- User avatar -->
      <button
        type="button"
        class="w-9 h-9 rounded-full overflow-hidden ring-2 ring-violet-600/50 hover:ring-violet-500 transition-all"
        @click="router.push({ name: 'profile' })"
      >
        <img
          v-if="userStore.profile?.avatar"
          :src="userStore.profile.avatar"
          alt="avatar"
          class="w-full h-full object-cover"
        />
        <div
          v-else
          class="w-full h-full bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center text-white text-sm font-bold"
        >
          {{ userStore.profile?.nickname?.[0]?.toUpperCase() ?? 'U' }}
        </div>
      </button>
    </div>

    <!-- Mobile fullscreen search popup -->
    <SearchPopup
      :show="searchVisible"
      @update:show="(v) => (searchVisible = v)"
    />
  </header>
</template>
