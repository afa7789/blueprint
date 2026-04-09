<template>
  <div>
    <HelperBox
      title="Linktree"
      description="Manage links displayed on the public Linktree page. Drag to reorder. Only active items are shown publicly."
      featureFlag="linktree_enabled"
    />
    <h2>Linktree</h2>
    <div v-if="error" class="error">{{ error }}</div>

    <button class="btn-primary" @click="openCreate">+ New Item</button>

    <div v-if="loading">Loading...</div>
    <div v-else class="item-list">
      <div v-for="(item, idx) in items" :key="item.id" class="item-row">
        <div class="item-order">
          <button :disabled="idx === 0" @click="moveUp(idx)">&#8593;</button>
          <button :disabled="idx === items.length - 1" @click="moveDown(idx)">&#8595;</button>
        </div>
        <div class="item-info">
          <span class="item-title">{{ item.title }}</span>
          <span class="item-url">{{ item.url }}</span>
        </div>
        <div class="item-meta">
          <span v-if="item.icon" class="item-icon">{{ item.icon }}</span>
          <span :class="item.is_active ? 'badge-active' : 'badge-inactive'">
            {{ item.is_active ? 'Active' : 'Inactive' }}
          </span>
        </div>
        <div class="item-actions">
          <button class="btn-sm" @click="openEdit(item)">Edit</button>
          <button class="btn-danger-sm" @click="deleteItem(item.id)">Delete</button>
        </div>
      </div>
    </div>

    <button v-if="items.length" class="btn-primary" style="margin-top:16px" @click="saveOrder">Save Order</button>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ form.id ? 'Edit Item' : 'New Item' }}</h3>
        <label>Title
          <input v-model="form.title" type="text" />
        </label>
        <label>URL
          <input v-model="form.url" type="text" />
        </label>
        <label>Icon
          <input v-model="form.icon" type="text" placeholder="emoji or icon name" />
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

interface LinkItem {
  id: string
  title: string
  url: string
  icon: string
  is_active: boolean
}

const items = ref<LinkItem[]>([])
const loading = ref(false)
const error = ref('')
const showModal = ref(false)

const emptyForm = () => ({ id: '', title: '', url: '', icon: '', is_active: true })
const form = reactive(emptyForm())

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ items: LinkItem[] }>('/api/v1/admin/linktree')
    items.value = data.items
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load linktree'
  } finally {
    loading.value = false
  }
}

function moveUp(idx: number) {
  if (idx === 0) return
  const arr = items.value
  ;[arr[idx - 1], arr[idx]] = [arr[idx], arr[idx - 1]]
}

function moveDown(idx: number) {
  const arr = items.value
  if (idx >= arr.length - 1) return
  ;[arr[idx + 1], arr[idx]] = [arr[idx], arr[idx + 1]]
}

async function saveOrder() {
  error.value = ''
  try {
    await api.put('/api/v1/admin/linktree/reorder', { ids: items.value.map(i => i.id) })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save order'
  }
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

function openEdit(item: LinkItem) {
  Object.assign(form, item)
  showModal.value = true
}

async function save() {
  error.value = ''
  try {
    if (form.id) {
      const data = await api.put<{ item: LinkItem }>(`/api/v1/admin/linktree/${form.id}`, form)
      const idx = items.value.findIndex(i => i.id === form.id)
      if (idx !== -1) items.value[idx] = data.item
    } else {
      const data = await api.post<{ item: LinkItem }>('/api/v1/admin/linktree', form)
      items.value.push(data.item)
    }
    showModal.value = false
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save item'
  }
}

async function deleteItem(id: string) {
  if (!confirm('Delete this item?')) return
  try {
    await api.delete(`/api/v1/admin/linktree/${id}`)
    items.value = items.value.filter(i => i.id !== id)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete item'
  }
}

onMounted(load)
</script>

<style scoped>
.item-list {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
}

.item-order {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-order button {
  background: none;
  border: 1px solid var(--border);
  border-radius: 3px;
  padding: 2px 6px;
  cursor: pointer;
  color: var(--text);
  line-height: 1;
}

.item-order button:disabled {
  opacity: 0.3;
  cursor: default;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-title {
  font-size: 14px;
  color: var(--text-h);
  font-weight: 500;
}

.item-url {
  font-size: 12px;
  color: var(--text);
  font-family: var(--mono);
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-icon {
  font-size: 18px;
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

.item-actions {
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
  width: 400px;
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

.modal input[type="text"] {
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
