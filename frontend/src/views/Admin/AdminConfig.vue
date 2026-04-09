<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface EnvVar {
  key: string
  value: string
  is_set: boolean
  is_secret: boolean
  category: string
  description: string
  required: boolean
}

const envVars = ref<EnvVar[]>([])
const filterCategory = ref('all')
const importText = ref('')
const importResult = ref<{ updated: number; skipped: number; message: string; errors?: string[] } | null>(null)
const showImport = ref(false)

const categories = computed(() => {
  const cats = new Set(envVars.value.map(v => v.category))
  return ['all', ...Array.from(cats).sort()]
})

const filtered = computed(() => {
  if (filterCategory.value === 'all') return envVars.value
  return envVars.value.filter(v => v.category === filterCategory.value)
})

const missingRequired = computed(() => envVars.value.filter(v => v.required && !v.is_set))
const missingOptional = computed(() => envVars.value.filter(v => !v.required && !v.is_set && !v.is_secret))

async function load() {
  const data = await api.get<{ data: EnvVar[] }>('/api/v1/admin/config/env')
  envVars.value = data.data || []
}

function exportEnv() {
  window.open('/api/v1/admin/config/export', '_blank')
}

async function doImport() {
  if (!importText.value.trim()) return
  try {
    const API_URL = import.meta.env.VITE_API_URL ?? ''
    const res = await fetch(`${API_URL}/api/v1/admin/config/import`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'text/plain' },
      body: importText.value,
    })
    const result = await res.json() as { updated: number; skipped: number; message: string; errors?: string[] }
    importResult.value = result
    await load()
  } catch (e: any) {
    importResult.value = { updated: 0, skipped: 0, message: e.message || 'Import failed' }
  }
}

async function importFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  importText.value = await file.text()
  input.value = ''
}

onMounted(load)
</script>

<template>
  <div>
    <HelperBox
      title="Configuration Panel"
      description="View environment variable status, export DB settings as .env, and import settings from .env files. ENV vars require server restart. DB-backed settings (security, feature flags) take effect immediately."
    />

    <div class="header">
      <h2>Configuration</h2>
      <div class="actions">
        <button @click="exportEnv" class="btn">Export .env</button>
        <button @click="showImport = !showImport" class="btn">{{ showImport ? 'Cancel Import' : 'Import .env' }}</button>
      </div>
    </div>

    <!-- Missing required warning -->
    <div v-if="missingRequired.length" class="alert alert-error">
      <strong>Missing required ENV vars:</strong>
      <span v-for="v in missingRequired" :key="v.key" class="badge badge-error">{{ v.key }}</span>
    </div>

    <!-- Missing optional info -->
    <div v-if="missingOptional.length" class="alert alert-warn">
      <strong>Unconfigured optional features:</strong>
      <span v-for="v in missingOptional" :key="v.key" class="badge badge-warn">{{ v.key }}</span>
    </div>

    <!-- Import panel -->
    <div v-if="showImport" class="import-panel">
      <h3>Import Settings</h3>
      <p class="hint">Paste .env content or upload a file. Only DB-backed settings (security, feature flags) will be updated. ENV vars are shown for reference but require server restart.</p>
      <div class="import-actions">
        <label class="btn file-btn">
          Upload .env file
          <input type="file" accept=".env,.txt" @change="importFile" hidden />
        </label>
      </div>
      <textarea v-model="importText" rows="10" placeholder="# Paste .env content here&#10;RATE_LIMIT_API=60&#10;FLAG_STORE_ENABLED=true"></textarea>
      <button @click="doImport" class="btn btn-primary" :disabled="!importText.trim()">Apply Import</button>
      <div v-if="importResult" class="import-result" :class="{ success: importResult.updated > 0 }">
        <p>{{ importResult.message }}</p>
        <ul v-if="importResult.errors?.length">
          <li v-for="(err, i) in importResult.errors" :key="i" class="error-line">{{ err }}</li>
        </ul>
      </div>
    </div>

    <!-- Category filter -->
    <div class="filter-bar">
      <button
        v-for="cat in categories"
        :key="cat"
        :class="['filter-btn', { active: filterCategory === cat }]"
        @click="filterCategory = cat"
      >
        {{ cat }}
      </button>
    </div>

    <!-- ENV vars table -->
    <table>
      <thead>
        <tr>
          <th>Variable</th>
          <th>Status</th>
          <th>Value</th>
          <th>Description</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="v in filtered" :key="v.key" :class="{ missing: v.required && !v.is_set }">
          <td>
            <code>{{ v.key }}</code>
            <span v-if="v.required" class="required-dot" title="Required">*</span>
          </td>
          <td>
            <span :class="['status-dot', v.is_set ? 'set' : 'unset']"></span>
            {{ v.is_set ? 'Set' : 'Not set' }}
          </td>
          <td class="value-cell">
            <template v-if="v.is_secret && v.is_set">
              <span class="secret">***</span>
            </template>
            <template v-else-if="v.value">
              <span class="val">{{ v.value }}</span>
            </template>
            <template v-else>
              <span class="empty">—</span>
            </template>
          </td>
          <td class="desc">{{ v.description }}</td>
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
}
h2 { margin: 0; }
.actions { display: flex; gap: 8px; }

.btn {
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  cursor: pointer;
  font-size: 13px;
}
.btn-primary { background: var(--accent); color: white; border-color: var(--accent); }
.file-btn { cursor: pointer; }

.alert {
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 13px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.alert-error { background: rgba(239,68,68,0.1); border: 1px solid rgba(239,68,68,0.3); color: #ef4444; }
.alert-warn { background: rgba(245,158,11,0.1); border: 1px solid rgba(245,158,11,0.3); color: #f59e0b; }

.badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-family: var(--mono);
  font-size: 12px;
}
.badge-error { background: rgba(239,68,68,0.15); }
.badge-warn { background: rgba(245,158,11,0.15); }

.import-panel {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}
.import-panel h3 { margin: 0 0 8px; font-size: 16px; color: var(--text-h); }
.hint { font-size: 13px; color: var(--text); margin: 0 0 12px; }
.import-actions { margin-bottom: 12px; }
.import-panel textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-family: var(--mono);
  font-size: 13px;
  resize: vertical;
  box-sizing: border-box;
  margin-bottom: 12px;
}
.import-result { margin-top: 12px; font-size: 13px; padding: 10px; border-radius: 6px; background: var(--bg); }
.import-result.success { border-left: 3px solid #22c55e; }
.error-line { color: #ef4444; font-size: 12px; }

.filter-bar {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.filter-btn {
  padding: 6px 12px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--text);
  cursor: pointer;
  font-size: 12px;
  text-transform: capitalize;
}
.filter-btn.active {
  background: var(--accent-bg);
  color: var(--accent);
  border-color: var(--accent-border);
}

table { width: 100%; border-collapse: collapse; }
th, td { padding: 10px 12px; text-align: left; border-bottom: 1px solid var(--border); font-size: 13px; }
th { font-weight: 500; color: var(--text); font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; }

tr.missing { background: rgba(239,68,68,0.05); }

code { font-size: 12px; padding: 2px 6px; background: var(--code-bg); border-radius: 3px; }
.required-dot { color: #ef4444; margin-left: 4px; font-weight: bold; }

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}
.status-dot.set { background: #22c55e; }
.status-dot.unset { background: #6b7280; }

.value-cell { font-family: var(--mono); font-size: 12px; max-width: 200px; overflow: hidden; text-overflow: ellipsis; }
.secret { color: var(--text); }
.val { color: var(--accent); }
.empty { color: var(--text); opacity: 0.4; }
.desc { color: var(--text); font-size: 12px; max-width: 250px; }
</style>
