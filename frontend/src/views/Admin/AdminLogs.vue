<template>
  <div>
    <HelperBox
      title="Application Logs"
      description="View and search application logs. Real-time streaming requires an active connection. Configure retention period and minimum log level in settings below."
    />
    <div class="page-header">
      <h2>Logs</h2>
      <button class="btn-primary" :class="{ 'btn-active': streaming }" @click="toggleStream">
        {{ streaming ? 'Stop Stream' : 'Stream' }}
      </button>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <select v-model="filters.level" @change="loadLogs">
        <option value="">All Levels</option>
        <option value="debug">Debug</option>
        <option value="info">Info</option>
        <option value="warn">Warn</option>
        <option value="error">Error</option>
      </select>
      <input v-model="filters.source" placeholder="Source" @keyup.enter="loadLogs" />
      <input v-model="filters.search" placeholder="Search..." @keyup.enter="loadLogs" />
      <input type="datetime-local" v-model="filters.from" @change="loadLogs" />
      <input type="datetime-local" v-model="filters.to" @change="loadLogs" />
      <button @click="loadLogs">Apply</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>

    <div class="log-list" ref="logContainer">
      <div v-for="(log, idx) in logs" :key="idx" class="log-line">
        <span class="log-time">{{ formatDate(log.created_at) }}</span>
        <span class="badge" :class="levelClass(log.level)">{{ log.level }}</span>
        <span class="log-source">{{ log.source }}</span>
        <span class="log-message">{{ log.message }}</span>
      </div>
      <div v-if="!logs.length && !loading" class="empty">No log entries.</div>
    </div>

    <div v-if="!streaming" class="pagination">
      <button :disabled="page <= 1" @click="page--; loadLogs()">Prev</button>
      <span>Page {{ page }}</span>
      <button :disabled="logs.length < pageSize" @click="page++; loadLogs()">Next</button>
    </div>

    <!-- Settings section -->
    <div class="settings-section">
      <h3>Log Settings</h3>
      <div v-if="configError" class="error">{{ configError }}</div>
      <div class="settings-row">
        <label>Retention Days
          <input type="number" v-model.number="config.retention_days" min="1" />
        </label>
        <label>Min Level
          <select v-model="config.min_level">
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warn">Warn</option>
            <option value="error">Error</option>
          </select>
        </label>
        <button class="btn-primary" @click="saveConfig">Save</button>
        <button @click="cleanupNow">Clear Logs</button>
      </div>
      <div v-if="cleanupMsg" class="success">{{ cleanupMsg }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

interface LogEntry {
  created_at: string
  level: string
  source: string
  message: string
}

interface LogConfig {
  retention_days: number
  min_level: string
}

const logs = ref<LogEntry[]>([])
const loading = ref(false)
const error = ref('')
const streaming = ref(false)
const page = ref(1)
const pageSize = 50
const logContainer = ref<HTMLElement | null>(null)
let eventSource: EventSource | null = null

const filters = ref({ level: '', source: '', search: '', from: '', to: '' })
const config = ref<LogConfig>({ retention_days: 30, min_level: 'debug' })
const configError = ref('')
const cleanupMsg = ref('')

async function loadLogs() {
  if (streaming.value) return
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams()
    params.set('page', String(page.value))
    params.set('limit', String(pageSize))
    if (filters.value.level) params.set('level', filters.value.level)
    if (filters.value.source) params.set('source', filters.value.source)
    if (filters.value.search) params.set('search', filters.value.search)
    if (filters.value.from) params.set('from', new Date(filters.value.from).toISOString())
    if (filters.value.to) params.set('to', new Date(filters.value.to).toISOString())
    const data = await api.get<{ data: LogEntry[] }>(`/api/v1/admin/logs?${params}`)
    logs.value = data.data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load logs'
  } finally {
    loading.value = false
  }
}

async function loadConfig() {
  try {
    const data = await api.get<LogConfig>('/api/v1/admin/logs/config')
    config.value = data
  } catch {
    // non-critical
  }
}

async function saveConfig() {
  configError.value = ''
  try {
    await api.put('/api/v1/admin/logs/config', config.value)
  } catch (e: unknown) {
    configError.value = e instanceof Error ? e.message : 'Failed to save config'
  }
}

async function cleanupNow() {
  cleanupMsg.value = ''
  try {
    const data = await api.post<{ deleted: number }>('/api/v1/admin/logs/cleanup', { all: true })
    cleanupMsg.value = `Deleted ${data.deleted} log entries.`
    await loadLogs()
  } catch (e: unknown) {
    configError.value = e instanceof Error ? e.message : 'Cleanup failed'
  }
}

function toggleStream() {
  if (streaming.value) {
    stopStream()
  } else {
    startStream()
  }
}

function startStream() {
  streaming.value = true
  logs.value = []
  const url = new URL(`${API_URL}/api/v1/admin/logs/stream`)
  if (filters.value.level) url.searchParams.set('level', filters.value.level)
  if (filters.value.source) url.searchParams.set('source', filters.value.source)
  eventSource = new EventSource(url.toString(), { withCredentials: true })
  eventSource.onmessage = async (event) => {
    try {
      const log = JSON.parse(event.data) as LogEntry
      logs.value.push(log)
      await nextTick()
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    } catch {
      // ignore parse errors
    }
  }
  eventSource.onerror = () => {
    stopStream()
    error.value = 'Stream disconnected'
  }
}

function stopStream() {
  streaming.value = false
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
}

function levelClass(level: string) {
  const map: Record<string, string> = {
    error: 'badge-red',
    warn: 'badge-yellow',
    info: 'badge-blue',
    debug: 'badge-gray',
  }
  return map[level] || ''
}

function formatDate(d: string) {
  return new Date(d).toLocaleString()
}

onMounted(() => {
  loadLogs()
  loadConfig()
})

onUnmounted(() => {
  if (eventSource) eventSource.close()
})
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.page-header h2 { margin: 0; }

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

.log-list {
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--code-bg);
  max-height: 500px;
  overflow-y: auto;
  font-family: var(--mono);
  font-size: 12px;
}

.log-line {
  display: flex;
  gap: 8px;
  align-items: baseline;
  padding: 6px 12px;
  border-bottom: 1px solid var(--border);
}

.log-line:last-child { border-bottom: none; }

.log-time {
  color: var(--text);
  white-space: nowrap;
  flex-shrink: 0;
}

.log-source {
  color: var(--text);
  white-space: nowrap;
  flex-shrink: 0;
}

.log-message {
  color: var(--text-h);
  word-break: break-all;
}

.badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  flex-shrink: 0;
}

.badge-red { background: #fee2e2; color: #ef4444; }
.badge-yellow { background: #fef9c3; color: #ca8a04; }
.badge-blue { background: #dbeafe; color: #3b82f6; }
.badge-gray { background: #f3f4f6; color: #6b7280; }

.empty {
  padding: 16px;
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
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

.settings-section {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--border);
}

.settings-section h3 {
  margin: 0 0 16px;
  font-size: 15px;
}

.settings-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: flex-end;
}

.settings-row label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--text);
}

.settings-row input,
.settings-row select {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 14px;
}

.settings-row button {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background: var(--accent) !important;
  color: #fff !important;
  border-color: var(--accent) !important;
}

.btn-active {
  background: #ef4444 !important;
  border-color: #ef4444 !important;
}

.error {
  color: #ef4444;
  font-size: 14px;
  margin-bottom: 8px;
}

.success {
  color: #22c55e;
  font-size: 13px;
  margin-top: 8px;
}
</style>
