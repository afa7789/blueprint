<template>
  <div>
    <HelperBox
      title="Banner System"
      description="Create banners with scheduling (start/end dates), profile targeting, and display duration. Banners show across the site based on active status and target audience."
    />
    <h2>Banners</h2>
    <div v-if="error" class="error">{{ error }}</div>

    <button class="btn-primary" @click="openCreate">+ New Banner</button>

    <div v-if="loading">Loading...</div>
    <table v-else class="admin-table">
      <thead>
        <tr>
          <th>Preview</th>
          <th>Title</th>
          <th>Target Profile</th>
          <th>Order</th>
          <th>Active</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="banner in banners" :key="banner.id">
          <td>
            <img v-if="banner.image_url" :src="banner.image_url" class="preview-img" alt="banner" />
            <span v-else>—</span>
          </td>
          <td>{{ banner.title }}</td>
          <td>{{ banner.target_profile || '—' }}</td>
          <td>{{ banner.order_index }}</td>
          <td>
            <span :class="banner.is_active ? 'badge-active' : 'badge-inactive'">
              {{ banner.is_active ? 'Active' : 'Inactive' }}
            </span>
          </td>
          <td class="actions">
            <button class="btn-sm" @click="openEdit(banner)">Edit</button>
            <button class="btn-danger-sm" @click="deleteBanner(banner.id)">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ form.id ? 'Edit Banner' : 'New Banner' }}</h3>
        <label>Title
          <input v-model="form.title" type="text" />
        </label>
        <label>Image URL
          <input v-model="form.image_url" type="text" />
        </label>
        <label>Link URL
          <input v-model="form.link_url" type="text" />
        </label>
        <label>Target Profile
          <input v-model="form.target_profile" type="text" placeholder="leave empty for all" />
        </label>
        <label>Order Index
          <input v-model.number="form.order_index" type="number" />
        </label>
        <label class="checkbox-label">
          <input v-model="form.is_active" type="checkbox" />
          Active
        </label>
        <div class="modal-actions">
          <button class="btn-primary" @click="save">Save</button>
          <button @click="showModal = false">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface Banner {
  id: string
  title: string
  image_url: string
  link_url: string
  target_profile: string
  is_active: boolean
  order_index: number
}

const banners = ref<Banner[]>([])
const loading = ref(false)
const error = ref('')
const showModal = ref(false)

const emptyForm = () => ({
  id: '',
  title: '',
  image_url: '',
  link_url: '',
  target_profile: '',
  is_active: true,
  order_index: 0,
})

const form = reactive(emptyForm())

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Banner[] }>('/api/v1/admin/banners')
    banners.value = data.data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load banners'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

function openEdit(banner: Banner) {
  Object.assign(form, banner)
  showModal.value = true
}

async function save() {
  error.value = ''
  try {
    if (form.id) {
      const data = await api.put<Banner>(`/api/v1/admin/banners/${form.id}`, form)
      const idx = banners.value.findIndex(b => b.id === form.id)
      if (idx !== -1) banners.value[idx] = data
    } else {
      const data = await api.post<Banner>('/api/v1/admin/banners', form)
      banners.value.push(data)
    }
    showModal.value = false
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save banner'
  }
}

async function deleteBanner(id: string) {
  if (!confirm('Delete this banner?')) return
  try {
    await api.delete(`/api/v1/admin/banners/${id}`)
    banners.value = banners.value.filter(b => b.id !== id)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete banner'
  }
}

onMounted(load)
</script>

<style scoped>
.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
  margin-top: 16px;
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

.preview-img {
  width: 60px;
  height: 36px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid var(--border);
}

.badge-active {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(72, 187, 120, 0.15);
  color: #48bb78;
}

.badge-inactive {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--code-bg);
  color: var(--text);
}

.actions {
  display: flex;
  gap: 6px;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.btn-sm {
  background: none;
  border: 1px solid var(--border);
  color: var(--text-h);
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.btn-danger-sm {
  background: none;
  border: 1px solid #e53e3e;
  color: #e53e3e;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
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
  width: 440px;
  max-width: 95vw;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal h3 {
  margin: 0 0 4px;
  font-size: 16px;
  color: var(--text-h);
}

.modal label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--text);
}

.modal input[type="text"],
.modal input[type="number"] {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 7px 10px;
  border-radius: 4px;
  font-size: 14px;
}

.checkbox-label {
  flex-direction: row !important;
  align-items: center;
  gap: 8px !important;
}

.modal-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.modal-actions button:last-child {
  background: none;
  border: 1px solid var(--border);
  color: var(--text-h);
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.error {
  color: #e53e3e;
  margin-bottom: 12px;
  font-size: 14px;
}
</style>
