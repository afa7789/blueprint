<script setup lang="ts">
import { RouterView } from 'vue-router'
import DynamicFooter from './components/common/DynamicFooter.vue'
import UpdateToast from './components/common/UpdateToast.vue'
</script>

<template>
  <div id="app">
    <main>
      <RouterView />
    </main>
    <DynamicFooter />
    <UpdateToast />
  </div>
</template>

<style scoped>
#app {
  width: 1126px;
  max-width: 100%;
  margin: 0 auto;
  text-align: center;
  border-inline: 1px solid var(--border);
  min-height: 100svh;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

main {
  flex: 1;
  display: flex;
  flex-direction: column;
}
</style>