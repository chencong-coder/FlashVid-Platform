<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import FeedView from './FeedView.vue'

const route = useRoute()
const startIndex = computed(() => {
  const raw = Number(route.query.index)
  return Number.isFinite(raw) && raw >= 0 ? raw : 0
})
</script>

<template>
  <FeedView feed="profile" :show-tabs="false" :start-index="startIndex" />
</template>
