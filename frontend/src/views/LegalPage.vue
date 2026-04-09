<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../services/api'

const route = useRoute()
const title = ref('')
const content = ref('')
const loading = ref(true)
const error = ref('')

async function loadPage(slug: string) {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ title: string; content: string; slug: string }>(`/api/v1/legal/${slug}`)
    title.value = data.title
    content.value = data.content
  } catch {
    error.value = 'Page not found.'
  } finally {
    loading.value = false
  }
}

onMounted(() => loadPage(route.params.slug as string))
watch(() => route.params.slug, (slug) => { if (slug) loadPage(slug as string) })
</script>

<template>
  <div class="legal-page">
    <p v-if="loading" class="status">Loading...</p>
    <p v-else-if="error" class="status error">{{ error }}</p>
    <template v-else>
      <h1>{{ title }}</h1>
      <div class="content" v-html="content"></div>
    </template>
    <div class="back">
      <router-link to="/">&larr; Back to home</router-link>
    </div>
  </div>
</template>

<style scoped>
.legal-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 48px 20px;
  text-align: left;
}

.legal-page h1 {
  font-size: 32px;
  margin-bottom: 24px;
}

.content {
  line-height: 1.7;
  color: var(--text);
}

.content :deep(h2) {
  font-size: 24px;
  margin: 32px 0 12px;
  color: var(--text-h);
}

.content :deep(h3) {
  font-size: 18px;
  margin: 24px 0 8px;
  color: var(--text-h);
}

.content :deep(p) {
  margin: 8px 0;
}

.content :deep(ul) {
  padding-left: 24px;
  margin: 8px 0;
}

.content :deep(li) {
  margin: 4px 0;
}

.status {
  text-align: center;
  padding: 48px 0;
  color: var(--text);
}

.status.error {
  color: #ef4444;
}

.back {
  margin-top: 48px;
  padding-top: 24px;
  border-top: 1px solid var(--border);
}

.back a {
  color: var(--accent);
  text-decoration: none;
  font-size: 14px;
}
</style>
