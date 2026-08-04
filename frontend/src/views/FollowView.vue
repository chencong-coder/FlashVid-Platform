<script setup lang="ts">
import FeedView from './feed/FeedView.vue'
import LoginGate from '@/components/LoginGate.vue'
</script>

<template>
  <LoginGate>
    <FeedView feed="follow" />
  </LoginGate>
</template>
