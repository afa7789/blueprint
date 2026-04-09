<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface LegalPage {
  id: string
  slug: string
  title: string
  content: string
  is_active: boolean
  updated_at: string
}

const pages = ref<LegalPage[]>([])
const editing = ref<LegalPage | null>(null)
const creating = ref(false)
const form = ref({ slug: '', title: '', content: '', is_active: true })
const saving = ref(false)

async function load() {
  const data = await api.get<{ data: LegalPage[] }>('/api/v1/admin/legal')
  pages.value = data.data || []
}

function startCreate() {
  form.value = { slug: '', title: '', content: '', is_active: true }
  creating.value = true
  editing.value = null
}

function startEdit(page: LegalPage) {
  form.value = { slug: page.slug, title: page.title, content: page.content, is_active: page.is_active }
  editing.value = page
  creating.value = false
}

function cancel() {
  editing.value = null
  creating.value = false
}

async function save() {
  saving.value = true
  try {
    if (creating.value) {
      await api.post('/api/v1/admin/legal', form.value)
    } else if (editing.value) {
      await api.put(`/api/v1/admin/legal/${editing.value.id}`, form.value)
    }
    cancel()
    await load()
  } catch (e: any) {
    alert(e.message || 'Failed to save')
  } finally {
    saving.value = false
  }
}

async function deletePage(id: string) {
  if (!confirm('Delete this page?')) return
  await api.delete(`/api/v1/admin/legal/${id}`)
  await load()
}

onMounted(load)
</script>

<template>
  <div>
    <HelperBox
      title="Legal Pages"
      description="Manage Terms of Service, Privacy Policy, and other legal pages. Active pages automatically appear in the site footer. Content supports HTML."
    />
    <div class="header">
      <h2>Legal Pages</h2>
      <button class="btn-primary" @click="startCreate">+ New Page</button>
    </div>

    <!-- Form -->
    <div v-if="creating || editing" class="form-card">
      <h3>{{ creating ? 'New Page' : 'Edit Page' }}</h3>
      <div class="field">
        <label>Slug</label>
        <input v-model="form.slug" placeholder="e.g. terms, privacy, cookies" :disabled="!!editing" />
      </div>
      <div class="field">
        <label>Title</label>
        <input v-model="form.title" placeholder="Terms of Service" />
      </div>
      <div class="field">
        <label>Content (HTML)</label>
        <textarea v-model="form.content" rows="16" placeholder="<h2>Title</h2><p>Content...</p>"></textarea>
      </div>
      <div class="field checkbox">
        <label><input type="checkbox" v-model="form.is_active" /> Active</label>
      </div>
      <div class="actions">
        <button class="btn-primary" @click="save" :disabled="saving">{{ saving ? 'Saving...' : 'Save' }}</button>
        <button class="btn-secondary" @click="cancel">Cancel</button>
      </div>
    </div>

    <!-- List -->
    <table v-else>
      <thead>
        <tr>
          <th>Slug</th>
          <th>Title</th>
          <th>Active</th>
          <th>Updated</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="page in pages" :key="page.id">
          <td><code>/legal/{{ page.slug }}</code></td>
          <td>{{ page.title }}</td>
          <td>{{ page.is_active ? 'Yes' : 'No' }}</td>
          <td>{{ new Date(page.updated_at).toLocaleDateString() }}</td>
          <td class="actions-cell">
            <button @click="startEdit(page)">Edit</button>
            <button class="btn-danger" @click="deletePage(page.id)">Delete</button>
          </td>
        </tr>
        <tr v-if="pages.length === 0">
          <td colspan="5" class="empty">No legal pages yet.</td>
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
  margin-bottom: 24px;
}

h2 { margin: 0; }

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

code {
  font-size: 13px;
  padding: 2px 6px;
  background: var(--code-bg);
  border-radius: 4px;
}

.actions-cell {
  display: flex;
  gap: 8px;
}

.form-card {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
}

.form-card h3 {
  margin: 0 0 16px;
  color: var(--text-h);
}

.field {
  margin-bottom: 16px;
}

.field label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text);
}

.field input, .field textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 14px;
  font-family: inherit;
  box-sizing: border-box;
}

.field textarea {
  font-family: var(--mono);
  font-size: 13px;
  resize: vertical;
}

.checkbox label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.actions {
  display: flex;
  gap: 8px;
}

button {
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  cursor: pointer;
  font-size: 13px;
}

.btn-primary {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-secondary {
  background: var(--bg);
}

.btn-danger {
  color: #ef4444;
  border-color: #ef4444;
}

.empty {
  text-align: center;
  color: var(--text);
  padding: 32px;
}
</style>
