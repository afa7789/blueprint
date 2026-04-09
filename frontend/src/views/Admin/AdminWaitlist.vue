<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface WaitlistEntry {
  id: string
  email: string
  name: string | null
  source: string | null
  created_at: string
}

const entries = ref<WaitlistEntry[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')

const filtered = computed(() => {
  if (!search.value) return entries.value
  const q = search.value.toLowerCase()
  return entries.value.filter(e => e.email.toLowerCase().includes(q) || (e.name && e.name.toLowerCase().includes(q)))
})

async function load() {
  loading.value = true
  try {
    entries.value = await api.get<WaitlistEntry[]>('/api/v1/waitlist')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load waitlist'
  } finally {
    loading.value = false
  }
}

function exportCSV() {
  const header = 'email,name,source,created_at'
  const rows = entries.value.map(e => `${e.email},${e.name ?? ''},${e.source ?? ''},${e.created_at}`)
  const csv = [header, ...rows].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `waitlist_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(load)
</script>

<template>
  <div>
    <HelperBox
      title="Waitlist"
      description="Users who signed up via the landing page waitlist form. Export to CSV for email campaigns."
      featureFlag="waitlist_enabled"
    />

    <div class="header">
      <h2>Waitlist <span class="count" v-if="!loading">({{ entries.length }})</span></h2>
      <div class="actions">
        <input v-model="search" type="text" placeholder="Search email or name..." class="search" />
        <button class="btn" @click="exportCSV" :disabled="entries.length === 0">
          <i class="fas fa-download"></i> Export CSV
        </button>
      </div>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading" class="status">Loading...</div>

    <table v-else>
      <thead>
        <tr>
          <th>Email</th>
          <th>Name</th>
          <th>Source</th>
          <th>Signed Up</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="entry in filtered" :key="entry.id">
          <td class="email">{{ entry.email }}</td>
          <td>{{ entry.name || '—' }}</td>
          <td>{{ entry.source || '—' }}</td>
          <td>{{ new Date(entry.created_at).toLocaleDateString() }}</td>
        </tr>
        <tr v-if="filtered.length === 0">
          <td colspan="4" class="empty">{{ search ? 'No matches' : 'No waitlist entries yet' }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

h2 { margin: 0; }

.count {
  font-weight: 400;
  color: var(--text);
  font-size: 18px;
}

.actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.search {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 13px;
  background: var(--bg);
  color: var(--text-h);
  width: 200px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.4;
  cursor: default;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border);
  font-size: 14px;
}

th {
  font-weight: 500;
  color: var(--text);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.email {
  font-family: var(--mono);
  font-size: 13px;
}

.empty {
  text-align: center;
  color: var(--text);
  padding: 32px;
  font-style: italic;
}

.error {
  color: #ef4444;
  margin-bottom: 12px;
  font-size: 14px;
}

.status {
  color: var(--text);
  padding: 20px 0;
}
</style>
