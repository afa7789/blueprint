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

interface FeatureFlag {
  id: number
  key: string
  enabled: boolean
}

interface FeatureDep {
  flagKey: string
  requiredEnvs: string[]
}

const FEATURE_DEPS: FeatureDep[] = [
  { flagKey: 'payments_stripe',           requiredEnvs: ['STRIPE_KEY'] },
  { flagKey: 'payments_pix',              requiredEnvs: ['STRIPE_KEY'] },
  { flagKey: 'ai_blog_enabled',           requiredEnvs: ['OPENAI_KEY'] },
  { flagKey: 'email_auto_enabled',        requiredEnvs: ['SMTP_HOST', 'SMTP_PORT'] },
  { flagKey: 'email_verification_required', requiredEnvs: ['SMTP_HOST', 'REDIS_URL'] },
]

const envVars = ref<EnvVar[]>([])
const features = ref<FeatureFlag[]>([])
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

// For each feature dep entry, compute which ENVs are missing
const featureDepsWithStatus = computed(() => {
  const envMap = new Map(envVars.value.map(v => [v.key, v.is_set]))
  return FEATURE_DEPS.map(dep => {
    const flag = features.value.find(f => f.key === dep.flagKey)
    const missingEnvs = dep.requiredEnvs.filter(k => !envMap.get(k))
    return {
      flagKey: dep.flagKey,
      enabled: flag?.enabled ?? false,
      requiredEnvs: dep.requiredEnvs,
      missingEnvs,
      warn: (flag?.enabled ?? false) && missingEnvs.length > 0,
    }
  })
})

async function load() {
  const [envData, featData] = await Promise.all([
    api.get<{ data: EnvVar[] }>('/api/v1/admin/config/env'),
    api.get<FeatureFlag[]>('/api/v1/admin/features').catch(() => [] as FeatureFlag[]),
  ])
  envVars.value = envData.data || []
  features.value = Array.isArray(featData) ? featData : []
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

    <!-- ── Active Feature Flags ── -->
    <section class="section-card">
      <h3 class="section-title">Active Feature Flags</h3>
      <div v-if="features.length === 0" class="empty-note">No feature flags found.</div>
      <div v-else class="flag-grid">
        <span
          v-for="f in features"
          :key="f.id"
          :class="['flag-badge', f.enabled ? 'flag-on' : 'flag-off']"
          :title="f.enabled ? 'Enabled' : 'Disabled'"
        >{{ f.key }}</span>
      </div>
    </section>

    <!-- ── Feature → ENV Dependency Map ── -->
    <section class="section-card">
      <h3 class="section-title">Feature → ENV Dependencies</h3>
      <table class="dep-table">
        <thead>
          <tr>
            <th>Flag</th>
            <th>Flag Status</th>
            <th>Required ENVs</th>
            <th>ENV Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="dep in featureDepsWithStatus" :key="dep.flagKey" :class="{ 'dep-warn': dep.warn }">
            <td><code>{{ dep.flagKey }}</code></td>
            <td>
              <span :class="['status-dot', dep.enabled ? 'set' : 'unset']"></span>
              {{ dep.enabled ? 'Enabled' : 'Disabled' }}
            </td>
            <td>
              <span v-for="env in dep.requiredEnvs" :key="env" class="env-tag">{{ env }}</span>
            </td>
            <td>
              <span v-if="dep.missingEnvs.length === 0" class="env-ok">All set</span>
              <span v-else class="env-missing">
                Missing: {{ dep.missingEnvs.join(', ') }}
                <span v-if="dep.warn" class="warn-icon" title="Feature is enabled but ENV is missing">⚠</span>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

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

    <!-- Missing required / optional alerts (contextual, above the ENV table) -->
    <div v-if="missingRequired.length" class="alert alert-error">
      <strong>Missing required ENV vars:</strong>
      <span v-for="v in missingRequired" :key="v.key" class="badge badge-error">{{ v.key }}</span>
    </div>
    <div v-if="missingOptional.length" class="alert alert-warn">
      <strong>Unconfigured optional features:</strong>
      <span v-for="v in missingOptional" :key="v.key" class="badge badge-warn">{{ v.key }}</span>
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

/* Feature flags & dep map */
.section-card {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}
.section-title { margin: 0 0 14px; font-size: 15px; color: var(--text-h); font-weight: 600; }
.empty-note { font-size: 13px; color: var(--text); opacity: 0.6; }

.flag-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.flag-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 500;
}
.flag-on  { background: rgba(34,197,94,0.15); color: #22c55e; border: 1px solid rgba(34,197,94,0.3); }
.flag-off { background: rgba(107,114,128,0.12); color: #6b7280; border: 1px solid rgba(107,114,128,0.25); }

.dep-table { width: 100%; border-collapse: collapse; }
.dep-table th, .dep-table td { padding: 9px 12px; text-align: left; border-bottom: 1px solid var(--border); font-size: 13px; }
.dep-table th { font-weight: 500; color: var(--text); font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; }
tr.dep-warn { background: rgba(239,68,68,0.05); }

.env-tag {
  display: inline-block;
  margin-right: 4px;
  padding: 2px 7px;
  border-radius: 4px;
  font-family: var(--mono);
  font-size: 11px;
  background: var(--bg);
  border: 1px solid var(--border);
  color: var(--text-h);
}
.env-ok   { color: #22c55e; font-size: 12px; }
.env-missing { color: #ef4444; font-size: 12px; font-family: var(--mono); }
.warn-icon { margin-left: 4px; font-style: normal; }
</style>
