<script setup lang="ts">
import { useServiceWorker } from '../../composables/useServiceWorker'

const { needsUpdate, updateApp } = useServiceWorker()
</script>

<template>
  <div v-if="needsUpdate" class="update-toast" @click="updateApp">
    New version available — click to refresh
  </div>
</template>

<style scoped>
.update-toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--accent);
  color: white;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  box-shadow: var(--shadow);
  z-index: 9999;
  white-space: nowrap;
}
</style>
