<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'

interface LinktreeItem {
  id: string
  title: string
  url: string
  icon: string | null
  is_active: boolean
  order_index: number
}

const items = ref<LinktreeItem[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const data = await api.get<LinktreeItem[]>('/api/v1/linktree')
    items.value = data.filter(i => i.is_active).sort((a, b) => a.order_index - b.order_index)
  } catch {
    // Linktree items may not be publicly accessible — show empty state
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="linktree">
    <h1>Links</h1>
    <p v-if="loading" class="status">Loading...</p>
    <p v-else-if="items.length === 0" class="status">No links available.</p>
    <div v-else class="links">
      <a
        v-for="item in items"
        :key="item.id"
        :href="item.url"
        target="_blank"
        rel="noopener noreferrer"
        class="link-card"
      >
        <span v-if="item.icon" class="icon">{{ item.icon }}</span>
        <span class="title">{{ item.title }}</span>
      </a>
    </div>
  </div>
</template>

<style scoped>
.linktree {
  padding: 60px 20px;
  max-width: 480px;
  margin: 0 auto;
  text-align: center;
}

.linktree h1 {
  font-size: 32px;
  margin-bottom: 32px;
}

.status {
  color: var(--text);
}

.links {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.link-card {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 16px 24px;
  border: 1px solid var(--border);
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-h);
  font-size: 16px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.link-card:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow);
}

.icon {
  font-size: 20px;
}
</style>
