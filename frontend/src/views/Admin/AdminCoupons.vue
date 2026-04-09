<template>
  <div class="admin-coupons">
    <div class="page-header">
      <h1>Coupons</h1>
      <button @click="openCreate" class="btn btn-primary">+ New Coupon</button>
    </div>

    <div v-if="loading" class="loading">Loading coupons...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th>Code</th>
          <th>Type</th>
          <th>Value</th>
          <th>Uses</th>
          <th>Expires</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="coupons.length === 0">
          <td colspan="6" class="empty-row">No coupons yet.</td>
        </tr>
        <tr v-for="coupon in coupons" :key="coupon.id">
          <td class="coupon-code">{{ coupon.code }}</td>
          <td>{{ coupon.type }}</td>
          <td>{{ coupon.type === 'percent' ? `${coupon.value}%` : `$${(coupon.value / 100).toFixed(2)}` }}</td>
          <td>{{ coupon.uses_count }} / {{ coupon.max_uses ?? '∞' }}</td>
          <td>{{ coupon.expires_at ? formatDate(coupon.expires_at) : 'Never' }}</td>
          <td class="actions">
            <button @click="deleteCoupon(coupon.id)" class="btn btn-danger btn-sm">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Create Modal -->
    <div v-if="showForm" class="modal-overlay" @click.self="closeForm">
      <div class="modal">
        <h2>New Coupon</h2>

        <div class="form-group">
          <label>Code</label>
          <input v-model="form.code" type="text" class="input" placeholder="SAVE10" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Type</label>
            <select v-model="form.type" class="input">
              <option value="percent">Percent off</option>
              <option value="fixed">Fixed amount (cents)</option>
            </select>
          </div>
          <div class="form-group">
            <label>Value</label>
            <input v-model.number="form.value" type="number" class="input" :placeholder="form.type === 'percent' ? '10' : '500'" />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Max Uses (blank = unlimited)</label>
            <input v-model.number="form.max_uses" type="number" class="input" placeholder="100" />
          </div>
          <div class="form-group">
            <label>Expires At (optional)</label>
            <input v-model="form.expires_at" type="date" class="input" />
          </div>
        </div>

        <p v-if="formError" class="field-error">{{ formError }}</p>

        <div class="modal-actions">
          <button @click="closeForm" class="btn btn-ghost">Cancel</button>
          <button @click="submitForm" class="btn btn-primary" :disabled="formSubmitting">
            {{ formSubmitting ? 'Saving...' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Coupon {
  id: string
  code: string
  type: 'percent' | 'fixed'
  value: number
  uses_count: number
  max_uses: number | null
  expires_at: string | null
}

const coupons = ref<Coupon[]>([])
const loading = ref(false)
const error = ref('')
const showForm = ref(false)
const formSubmitting = ref(false)
const formError = ref('')

const form = ref({
  code: '',
  type: 'percent' as 'percent' | 'fixed',
  value: 0,
  max_uses: null as number | null,
  expires_at: '',
})

async function fetchCoupons() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Coupon[] }>('/api/v1/admin/coupons')
    coupons.value = data.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load coupons'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = { code: '', type: 'percent', value: 0, max_uses: null, expires_at: '' }
  formError.value = ''
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

async function submitForm() {
  formSubmitting.value = true
  formError.value = ''
  try {
    const payload = {
      ...form.value,
      max_uses: form.value.max_uses || null,
      expires_at: form.value.expires_at || null,
    }
    await api.post('/api/v1/admin/coupons', payload)
    await fetchCoupons()
    closeForm()
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : 'Failed to create coupon'
  } finally {
    formSubmitting.value = false
  }
}

async function deleteCoupon(id: string) {
  if (!confirm('Delete this coupon?')) return
  try {
    await api.delete(`/api/v1/admin/coupons/${id}`)
    await fetchCoupons()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete coupon'
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(fetchCoupons)
</script>

<style scoped>
.admin-coupons {
  padding: 32px;
  text-align: left;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
}

.loading, .empty-row {
  text-align: center;
  color: var(--text);
  padding: 32px;
}

.error {
  color: #ef4444;
  padding: 12px;
  border: 1px solid #ef4444;
  border-radius: 6px;
  margin-bottom: 16px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table th,
.data-table td {
  text-align: left;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

.data-table th {
  color: var(--text);
  font-weight: 500;
  background: var(--code-bg);
}

.data-table tr:hover td {
  background: var(--social-bg);
}

.coupon-code {
  font-family: monospace;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--border);
  cursor: pointer;
  font-size: 14px;
  background: var(--bg);
  color: var(--text);
  transition: box-shadow 0.2s;
  font-family: inherit;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 13px;
}

.btn-primary {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
}

.btn-ghost {
  background: transparent;
}

.btn:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 32px 16px;
  z-index: 100;
  overflow-y: auto;
}

.modal {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 32px;
  width: 100%;
  max-width: 480px;
}

.modal h2 {
  margin: 0 0 24px;
}

.form-group {
  margin-bottom: 16px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 6px;
}

.input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 14px;
  box-sizing: border-box;
  font-family: inherit;
}

.input:focus {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.field-error {
  color: #ef4444;
  font-size: 13px;
  margin: 8px 0 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  border-top: 1px solid var(--border);
  padding-top: 20px;
}
</style>
