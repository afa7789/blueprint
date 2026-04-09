<template>
  <div>
    <div class="page-header">
      <h2>Tools Hub</h2>
      <button class="btn-primary" @click="openCreate">Add Tool</button>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>

    <template v-else>
      <div v-for="(group, category) in grouped" :key="category" class="category-group">
        <h3 class="category-title">{{ categoryEmoji(String(category)) }} {{ category }}</h3>
        <div class="tools-grid">
          <div v-for="tool in group" :key="tool.id" class="tool-card">
            <div class="card-header">
              <span class="tool-name">{{ tool.name }}</span>
              <span class="status-dot" :class="pingStatus[tool.id] === 'up' ? 'dot-green' : pingStatus[tool.id] === 'down' ? 'dot-red' : 'dot-gray'"></span>
            </div>
            <p class="tool-desc">{{ tool.description }}</p>
            <div class="card-footer">
              <a v-if="tool.url" :href="tool.url" target="_blank" class="link-open">Open</a>
              <button @click="openEdit(tool)">Edit</button>
              <button class="btn-danger" @click="deleteTool(tool.id)">Delete</button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!Object.keys(grouped).length" class="empty">No tools configured.</div>
    </template>

    <!-- Add/Edit Modal -->
    <div v-if="showForm" class="modal-overlay" @click.self="showForm = false">
      <div class="modal">
        <h3>{{ editingId ? 'Edit Tool' : 'Add Tool' }}</h3>
        <div v-if="formError" class="error">{{ formError }}</div>
        <label>Name
          <input v-model="form.name" placeholder="My Tool" />
        </label>
        <label>Description
          <input v-model="form.description" placeholder="Brief description" />
        </label>
        <label>Category
          <input v-model="form.category" placeholder="database" />
        </label>
        <label>URL
          <input v-model="form.url" placeholder="https://..." />
        </label>
        <div class="modal-actions">
          <button @click="showForm = false">Cancel</button>
          <button class="btn-primary" @click="saveTool">{{ editingId ? 'Save' : 'Add' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../services/api'

interface Tool {
  id: string
  name: string
  description: string
  category: string
  url: string
}

const tools = ref<Tool[]>([])
const loading = ref(false)
const error = ref('')
const pingStatus = ref<Record<string, 'up' | 'down' | 'pending'>>({})
const showForm = ref(false)
const formError = ref('')
const editingId = ref<string | null>(null)
const form = ref({ name: '', description: '', category: '', url: '' })

const grouped = computed(() => {
  const g: Record<string, Tool[]> = {}
  for (const t of tools.value) {
    const cat = t.category || 'other'
    if (!g[cat]) g[cat] = []
    g[cat].push(t)
  }
  return g
})

function categoryEmoji(cat: string) {
  const map: Record<string, string> = {
    database: '🗄️',
    cache: '⚡',
    monitoring: '📊',
    storage: '📦',
  }
  return map[cat.toLowerCase()] || '🔧'
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ tools: Tool[] }>('/api/v1/admin/tools')
    tools.value = data.tools
    pingAll()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load tools'
  } finally {
    loading.value = false
  }
}

function pingAll() {
  for (const tool of tools.value) {
    pingStatus.value[tool.id] = 'pending'
    api.get(`/api/v1/admin/tools/${tool.id}/ping`)
      .then(() => { pingStatus.value[tool.id] = 'up' })
      .catch(() => { pingStatus.value[tool.id] = 'down' })
  }
}

function openCreate() {
  editingId.value = null
  form.value = { name: '', description: '', category: '', url: '' }
  formError.value = ''
  showForm.value = true
}

function openEdit(tool: Tool) {
  editingId.value = tool.id
  form.value = { name: tool.name, description: tool.description, category: tool.category, url: tool.url }
  formError.value = ''
  showForm.value = true
}

async function saveTool() {
  formError.value = ''
  if (!form.value.name) { formError.value = 'Name required'; return }
  try {
    if (editingId.value) {
      const data = await api.put<{ tool: Tool }>(`/api/v1/admin/tools/${editingId.value}`, form.value)
      const idx = tools.value.findIndex(t => t.id === editingId.value)
      if (idx !== -1) tools.value[idx] = data.tool
    } else {
      const data = await api.post<{ tool: Tool }>('/api/v1/admin/tools', form.value)
      tools.value.push(data.tool)
      pingStatus.value[data.tool.id] = 'pending'
      api.get(`/api/v1/admin/tools/${data.tool.id}/ping`)
        .then(() => { pingStatus.value[data.tool.id] = 'up' })
        .catch(() => { pingStatus.value[data.tool.id] = 'down' })
    }
    showForm.value = false
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : 'Failed to save'
  }
}

async function deleteTool(id: string) {
  if (!confirm('Delete this tool?')) return
  try {
    await api.delete(`/api/v1/admin/tools/${id}`)
    tools.value = tools.value.filter(t => t.id !== id)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete'
  }
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-header h2 { margin: 0; }

.category-group { margin-bottom: 32px; }

.category-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 12px;
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.tool-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
  background: var(--bg);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tool-name {
  font-weight: 500;
  font-size: 14px;
  color: var(--text-h);
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-green { background: #22c55e; }
.dot-red { background: #ef4444; }
.dot-gray { background: #9ca3af; }

.tool-desc {
  font-size: 13px;
  color: var(--text);
  margin: 0;
  flex: 1;
}

.card-footer {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
  margin-top: 4px;
}

.link-open {
  border: 1px solid var(--accent);
  color: var(--accent);
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  text-decoration: none;
}

.card-footer button {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.card-footer button:hover {
  background: var(--code-bg);
}

.btn-danger {
  border-color: #ef4444 !important;
  color: #ef4444 !important;
}

.btn-danger:hover {
  background: #ef4444 !important;
  color: #fff !important;
}

.empty {
  color: var(--text);
  font-size: 14px;
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

.modal h3 { margin: 0 0 4px; }

.modal label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--text);
}

.modal input {
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

.btn-primary:hover { opacity: 0.9; }

.error {
  color: #ef4444;
  font-size: 14px;
}
</style>
