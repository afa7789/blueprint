<template>
  <div>
    <div class="page-header">
      <h2>Jobs</h2>
      <button class="btn-primary" @click="showCreate = true">Create Job</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>

    <template v-else>
      <table class="admin-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Schedule</th>
            <th>Handler</th>
            <th>Status</th>
            <th>Last Run</th>
            <th>Next Run</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="job in jobs" :key="job.id">
            <tr>
              <td>{{ job.name }}</td>
              <td class="mono">{{ job.schedule }}</td>
              <td class="mono">{{ job.handler }}</td>
              <td>
                <span class="badge" :class="job.is_active ? 'badge-green' : 'badge-yellow'">
                  {{ job.is_active ? 'active' : 'paused' }}
                </span>
              </td>
              <td>{{ job.last_run ? formatDate(job.last_run) : '—' }}</td>
              <td>{{ job.next_run ? formatDate(job.next_run) : '—' }}</td>
              <td class="actions">
                <button v-if="job.is_active" @click="pauseJob(job.id)">Pause</button>
                <button v-else @click="resumeJob(job.id)">Resume</button>
                <button @click="runNow(job.id)">Run Now</button>
                <button @click="toggleExecutions(job.id)">
                  {{ expandedJob === job.id ? 'Hide' : 'View' }} Executions
                </button>
              </td>
            </tr>
            <tr v-if="expandedJob === job.id">
              <td colspan="7" class="executions-cell">
                <div v-if="execLoading">Loading executions...</div>
                <table v-else-if="executions.length" class="admin-table inner-table">
                  <thead>
                    <tr>
                      <th>Status</th>
                      <th>Started</th>
                      <th>Duration</th>
                      <th>Error</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="ex in executions" :key="ex.id">
                      <td>
                        <span class="badge" :class="execBadgeClass(ex.status)">{{ ex.status }}</span>
                      </td>
                      <td>{{ formatDate(ex.started_at) }}</td>
                      <td>{{ ex.duration_ms != null ? ex.duration_ms + 'ms' : '—' }}</td>
                      <td class="error-cell">{{ ex.error || '—' }}</td>
                      <td>
                        <button v-if="ex.status === 'failed'" @click="retryExecution(job.id, ex.id)">Retry</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <div v-else class="empty">No executions found.</div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </template>

    <!-- Create Job Modal -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal">
        <h3>Create Job</h3>
        <div v-if="createError" class="error">{{ createError }}</div>
        <label>Name
          <input v-model="form.name" placeholder="my-job" />
        </label>
        <label>Schedule (cron)
          <input v-model="form.schedule" placeholder="0 * * * *" />
        </label>
        <label>Handler
          <select v-model="form.handler">
            <option value="">Select handler</option>
            <option v-for="h in handlers" :key="h" :value="h">{{ h }}</option>
          </select>
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="form.is_active" />
          Active
        </label>
        <div class="modal-actions">
          <button @click="showCreate = false">Cancel</button>
          <button class="btn-primary" @click="createJob">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Job {
  id: string
  name: string
  schedule: string
  handler: string
  is_active: boolean
  last_run: string | null
  next_run: string | null
}

interface Execution {
  id: string
  status: string
  started_at: string
  duration_ms: number | null
  error: string | null
}

const jobs = ref<Job[]>([])
const loading = ref(false)
const error = ref('')
const executions = ref<Execution[]>([])
const execLoading = ref(false)
const expandedJob = ref<string | null>(null)
const showCreate = ref(false)
const createError = ref('')
const handlers = ref<string[]>([])

const form = ref({ name: '', schedule: '', handler: '', is_active: true })

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ jobs: Job[] }>('/api/v1/admin/jobs')
    jobs.value = data.jobs
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load jobs'
  } finally {
    loading.value = false
  }
}

async function loadHandlers() {
  try {
    const data = await api.get<{ handlers: string[] }>('/api/v1/admin/jobs/handlers')
    handlers.value = data.handlers
  } catch {
    // non-critical
  }
}

async function pauseJob(id: string) {
  try {
    await api.put(`/api/v1/admin/jobs/${id}/pause`)
    const j = jobs.value.find(j => j.id === id)
    if (j) j.is_active = false
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to pause'
  }
}

async function resumeJob(id: string) {
  try {
    await api.put(`/api/v1/admin/jobs/${id}/resume`)
    const j = jobs.value.find(j => j.id === id)
    if (j) j.is_active = true
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to resume'
  }
}

async function runNow(id: string) {
  try {
    await api.post(`/api/v1/admin/jobs/${id}/run`)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to run job'
  }
}

async function toggleExecutions(id: string) {
  if (expandedJob.value === id) {
    expandedJob.value = null
    return
  }
  expandedJob.value = id
  execLoading.value = true
  executions.value = []
  try {
    const data = await api.get<{ executions: Execution[] }>(`/api/v1/admin/jobs/${id}/executions?page=1&limit=20`)
    executions.value = data.executions
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load executions'
  } finally {
    execLoading.value = false
  }
}

async function retryExecution(jobId: string, eid: string) {
  try {
    await api.post(`/api/v1/admin/jobs/${jobId}/executions/${eid}/retry`)
    await toggleExecutions(jobId)
    await toggleExecutions(jobId)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to retry'
  }
}

async function createJob() {
  createError.value = ''
  if (!form.value.name || !form.value.schedule || !form.value.handler) {
    createError.value = 'All fields required'
    return
  }
  try {
    const data = await api.post<{ job: Job }>('/api/v1/admin/jobs', form.value)
    jobs.value.push(data.job)
    showCreate.value = false
    form.value = { name: '', schedule: '', handler: '', is_active: true }
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : 'Failed to create job'
  }
}

function execBadgeClass(status: string) {
  if (status === 'success') return 'badge-green'
  if (status === 'failed') return 'badge-red'
  if (status === 'running') return 'badge-blue'
  return ''
}

function formatDate(d: string) {
  return new Date(d).toLocaleString()
}

onMounted(() => {
  load()
  loadHandlers()
})
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
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

.mono {
  font-family: var(--mono);
  font-size: 12px;
}

.actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.actions button, .actions button {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.actions button:hover {
  background: var(--code-bg);
}

.badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.badge-green { background: #dcfce7; color: #22c55e; }
.badge-yellow { background: #fef9c3; color: #ca8a04; }
.badge-red { background: #fee2e2; color: #ef4444; }
.badge-blue { background: #dbeafe; color: #3b82f6; }

.executions-cell {
  background: var(--code-bg);
  padding: 12px 20px;
}

.inner-table {
  background: var(--bg);
}

.error-cell {
  color: #ef4444;
  font-size: 12px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty {
  color: var(--text);
  font-size: 14px;
  padding: 8px 0;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.modal {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
  width: 400px;
  max-width: 90vw;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal h3 {
  margin: 0 0 4px;
}

.modal label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--text);
}

.checkbox-label {
  flex-direction: row !important;
  align-items: center;
  gap: 8px !important;
}

.modal input, .modal select {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 14px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

.modal-actions button {
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

.btn-primary:hover {
  opacity: 0.9;
}

.error {
  color: #ef4444;
  font-size: 14px;
  margin-bottom: 8px;
}
</style>
