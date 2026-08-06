import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { getUnreadCounts, type UnreadCounts } from '@/api/notification'

export const useNotificationStore = defineStore('notification', () => {
  const unreadCounts = ref<UnreadCounts>({
    followers: 0,
    likesAndFavs: 0,
    mentions: 0,
    comments: 0,
  })

  const totalUnread = computed(
    () =>
      unreadCounts.value.followers +
      unreadCounts.value.likesAndFavs +
      unreadCounts.value.mentions +
      unreadCounts.value.comments,
  )

  const fetchUnreadCounts = async () => {
    try {
      const res = await getUnreadCounts()
      if (res.data.code === 0 && res.data.data) {
        unreadCounts.value = res.data.data
      }
    } catch {
      // 静默失败，不影响主流程
    }
  }

  const resetTypeCount = (type: keyof UnreadCounts) => {
    unreadCounts.value[type] = 0
  }

  return { unreadCounts, totalUnread, fetchUnreadCounts, resetTypeCount }
})
