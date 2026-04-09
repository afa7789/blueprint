<script setup lang="ts">
defineProps<{
  title: string
  description: string
  envVar?: string
  featureFlag?: string
}>()
</script>

<template>
  <div class="helper-box">
    <h4>{{ title }}</h4>
    <p>{{ description }}</p>
    <div v-if="envVar" class="env-notice">
      Requires: <code>{{ envVar }}</code>
    </div>
    <div v-if="featureFlag" class="flag-notice">
      Feature flag: <code>{{ featureFlag }}</code>
    </div>
  </div>
</template>

<style scoped>
.helper-box {
  border-left: 3px solid var(--accent);
  background: var(--code-bg, #f8f8f8);
  padding: 12px 16px;
  border-radius: 0 8px 8px 0;
  margin-bottom: 20px;
  font-size: 13px;
}

.helper-box h4 {
  margin: 0 0 4px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-h);
}

.helper-box p {
  margin: 0 0 8px;
  color: var(--text);
  line-height: 1.5;
}

.helper-box p:last-child {
  margin-bottom: 0;
}

.env-notice,
.flag-notice {
  margin-top: 6px;
  color: var(--text);
}

.env-notice code,
.flag-notice code {
  background: var(--bg);
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 12px;
  color: var(--accent);
}
</style>
