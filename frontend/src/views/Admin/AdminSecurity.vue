<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface Setting {
  id: number
  key: string
  value: string
  description: string | null
  updated_at: string
}

const settings = ref<Setting[]>([])
const editing = ref<string | null>(null)
const editValue = ref('')
const saving = ref(false)

async function load() {
  const data = await api.get<{ data: Setting[] }>('/api/v1/admin/security')
  settings.value = data.data || []
}

function startEdit(s: Setting) {
  editing.value = s.key
  editValue.value = s.value
}

function cancel() {
  editing.value = null
}

async function save(key: string) {
  saving.value = true
  try {
    await api.put(`/api/v1/admin/security/${key}`, { value: editValue.value })
    editing.value = null
    await load()
  } catch (e: any) {
    alert(e.message || 'Failed to save')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <HelperBox title="Rate Limiting" description="Rate limits use Redis for tracking. Without Redis, rate limiting is disabled." envVar="REDIS_URL" />
    <h2>Security Settings</h2>
    <p class="subtitle">Configure rate limiting, authentication, and security policies.</p>

    <table>
      <thead>
        <tr>
          <th>Setting</th>
          <th>Value</th>
          <th>Description</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="s in settings" :key="s.key">
          <td><code>{{ s.key }}</code></td>
          <td>
            <template v-if="editing === s.key">
              <input v-model="editValue" class="edit-input" @keyup.enter="save(s.key)" />
            </template>
            <template v-else>
              <span class="value">{{ s.value }}</span>
            </template>
          </td>
          <td class="desc">{{ s.description || '-' }}</td>
          <td>
            <template v-if="editing === s.key">
              <button class="btn-primary" @click="save(s.key)" :disabled="saving">Save</button>
              <button @click="cancel">Cancel</button>
            </template>
            <template v-else>
              <button @click="startEdit(s)">Edit</button>
            </template>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="info-box">
      <h3>Rate Limit Reference</h3>
      <ul>
        <li><code>rate_limit_api</code> — General API requests per minute per IP</li>
        <li><code>rate_limit_auth</code> — Login attempts per minute per email</li>
        <li><code>rate_limit_register</code> — Registrations per hour per email</li>
        <li><code>rate_limit_forgot</code> — Password reset requests per hour per email</li>
        <li><code>max_login_attempts</code> — Failed logins before temporary lock</li>
        <li><code>login_lock_duration</code> — Lock duration in minutes</li>
      </ul>
      <p>Changes take effect on next request. No server restart needed for DB-backed settings.</p>
      <p>ENV-based rate limits (in .env) require a restart. DB settings override ENV when implemented.</p>
    </div>
  </div>
</template>

<style scoped>
h2 { margin: 0 0 4px; }
.subtitle { color: var(--text); font-size: 14px; margin: 0 0 24px; }

table { width: 100%; border-collapse: collapse; }
th, td { padding: 10px 12px; text-align: left; border-bottom: 1px solid var(--border); font-size: 14px; }
th { font-weight: 500; color: var(--text); font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; }

code { font-size: 13px; padding: 2px 6px; background: var(--code-bg); border-radius: 4px; }
.value { font-family: var(--mono); font-size: 14px; color: var(--accent); font-weight: 500; }
.desc { color: var(--text); font-size: 13px; max-width: 300px; }

.edit-input {
  padding: 6px 10px;
  border: 1px solid var(--accent);
  border-radius: 4px;
  font-size: 14px;
  font-family: var(--mono);
  width: 80px;
  background: var(--bg);
  color: var(--text-h);
}

button {
  padding: 6px 12px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--text-h);
  cursor: pointer;
  font-size: 13px;
  margin-right: 4px;
}
.btn-primary { background: var(--accent); color: white; border-color: var(--accent); }

.info-box {
  margin-top: 32px;
  padding: 20px;
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
}
.info-box h3 { margin: 0 0 12px; font-size: 15px; color: var(--text-h); }
.info-box ul { padding-left: 20px; margin: 0 0 12px; }
.info-box li { font-size: 13px; margin: 4px 0; color: var(--text); }
.info-box p { font-size: 13px; color: var(--text); margin: 8px 0 0; }
</style>
