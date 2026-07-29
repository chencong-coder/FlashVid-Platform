<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import FeedView from './FeedView.vue'

const route = useRoute()
// 起始索引由话题页跳转时通过 query 传入
const startIndex = computed(() => {
  const raw = Number(route.query.index)
  return Number.isFinite(raw) && raw >= 0 ? raw : 0
})
</script>

<template>
  <FeedView feed="topic" :show-tabs="false" :start-index="startIndex" />
</template>
