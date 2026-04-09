<template>
  <div>
    <HelperBox
      title="Audit Trail"
      description="All admin POST/PUT/DELETE actions are logged automatically. Filter by user, action type, resource, or date range to investigate changes."
    />
    <h2>Audit Trail</h2>

    <!-- Filters -->
    <div class="filter-bar">
      <input v-model="filters.user_id" placeholder="User ID" @keyup.enter="loadAudit" />
      <input v-model="filters.action" placeholder="Action" @keyup.enter="loadAudit" />
      <input v-model="filters.resource" placeholder="Resource" @keyup.enter="loadAudit" />
      <input type="datetime-local" v-model="filters.from" @change="loadAudit" />
      <input type="datetime-local" v-model="filters.to" @change="loadAudit" />
      <button @click="loadAudit">Apply</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>

    <template v-else>
      <table class="admin-table">
        <thead>
          <tr>
            <th>Timestamp</th>
            <th>User ID</th>
            <th>Action</th>
            <th>Resource</th>
            <th>Resource ID</th>
            <th>IP Address</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="entry in entries" :key="entry.id">
            <tr class="audit-row" @click="toggleExpand(entry.id)" :class="{ expanded: expandedId === entry.id }">
              <td>{{ formatDate(entry.timestamp) }}</td>
              <td class="mono">{{ entry.user_id }}</td>
              <td><span class="action-badge">{{ entry.action }}</span></td>
              <td>{{ entry.resource }}</td>
              <td class="mono">{{ entry.resource_id || '—' }}</td>
              <td class="mono">{{ entry.ip_address || '—' }}</td>
            </tr>
            <tr v-if="expandedId === entry.id">
              <td colspan="6" class="details-cell">
                <pre>{{ JSON.stringify(entry.details, null, 2) }}</pre>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
      <div v-if="!entries.length" class="empty">No audit entries found.</div>

      <div class="pagination">
        <button :disabled="page <= 1" @click="page--; loadAudit()">Prev</button>
        <span>Page {{ page }}</span>
        <button :disabled="entries.length < pageSize" @click="page++; loadAudit()">Next</button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface AuditEntry {
  id: string
  timestamp: string
  user_id: string
  action: string
  resource: string
  resource_id: string | null
  ip_address: string | null
  details: unknown
}

const entries = ref<AuditEntry[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = 20
const expandedId = ref<string | null>(null)

const filters = ref({ user_id: '', action: '', resource: '', from: '', to: '' })

async function loadAudit() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams()
    params.set('page', String(page.value))
    params.set('limit', String(pageSize))
    if (filters.value.user_id) params.set('user_id', filters.value.user_id)
    if (filters.value.action) params.set('action', filters.value.action)
    if (filters.value.resource) params.set('resource', filters.value.resource)
    if (filters.value.from) params.set('from', filters.value.from)
    if (filters.value.to) params.set('to', filters.value.to)
    const data = await api.get<{ data: AuditEntry[] }>(`/api/v1/admin/audit?${params}`)
    entries.value = data.data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load audit log'
  } finally {
    loading.value = false
  }
}

function toggleExpand(id: string) {
  expandedId.value = expandedId.value === id ? null : id
}

function formatDate(d: string) {
  return new Date(d).toLocaleString()
}

onMounted(loadAudit)
</script>

<style scoped>
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.filter-bar input,
.filter-bar select {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 13px;
}

.filter-bar button {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.admin-table th,
.admin-table td {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}

.admin-table th {
  color: var(--text);
  font-weight: 500;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.audit-row {
  cursor: pointer;
  transition: background 0.1s;
}

.audit-row:hover,
.audit-row.expanded {
  background: var(--code-bg);
}

.mono {
  font-family: var(--mono);
  font-size: 12px;
}

.action-badge {
  display: inline-block;
  background: var(--code-bg);
  color: var(--text-h);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-family: var(--mono);
}

.details-cell {
  background: var(--code-bg);
  padding: 0;
}

.details-cell pre {
  margin: 0;
  padding: 12px 20px;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-h);
  white-space: pre-wrap;
  word-break: break-all;
}

.empty {
  color: var(--text);
  font-size: 14px;
  margin-top: 16px;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  font-size: 14px;
}

.pagination button {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: default;
}

.error {
  color: #ef4444;
  font-size: 14px;
  margin-bottom: 12px;
}
</style>
