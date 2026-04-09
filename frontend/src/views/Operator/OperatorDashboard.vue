<template>
  <div class="operator-page">
    <h1>Operator Dashboard</h1>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>

    <template v-else>
      <table v-if="orders.length > 0" class="admin-table">
        <thead>
          <tr>
            <th>Order ID</th>
            <th>Customer</th>
            <th>Total</th>
            <th>Date</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="order in orders" :key="order.id">
            <td class="mono">{{ order.id }}</td>
            <td>{{ order.user_email || '—' }}</td>
            <td>${{ (order.total / 100).toFixed(2) }}</td>
            <td>{{ formatDate(order.created_at) }}</td>
            <td class="actions-cell">
              <template v-if="shippingOrderId === order.id">
                <div class="inline-form">
                  <input
                    v-model="trackingCode"
                    type="text"
                    class="input"
                    placeholder="Tracking code"
                  />
                  <button class="btn btn-primary" :disabled="actionLoading" @click="submitShipped(order.id)">
                    {{ actionLoading ? 'Saving...' : 'Confirm' }}
                  </button>
                  <button class="btn btn-secondary" @click="shippingOrderId = null">Cancel</button>
                </div>
                <div v-if="actionError" class="action-error">{{ actionError }}</div>
              </template>
              <template v-else>
                <button class="btn btn-primary" @click="startShipping(order.id)">
                  Mark Shipped
                </button>
                <button
                  v-if="order.payment_method === 'pix_manual'"
                  class="btn btn-secondary"
                  :disabled="actionLoading"
                  @click="approvePix(order.id)"
                >
                  Approve PIX
                </button>
              </template>
              <span v-if="successOrderId === order.id" class="success-badge">Done</span>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">No paid orders found.</div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Order {
  id: string
  user_email?: string
  total: number
  created_at: string
  payment_method: string
}

const orders = ref<Order[]>([])
const loading = ref(false)
const error = ref('')
const actionLoading = ref(false)
const actionError = ref('')
const shippingOrderId = ref<string | null>(null)
const trackingCode = ref('')
const successOrderId = ref<string | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ orders: Order[] }>('/api/v1/admin/orders?status=paid')
    orders.value = data.orders
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load orders'
  } finally {
    loading.value = false
  }
}

function startShipping(id: string) {
  shippingOrderId.value = id
  trackingCode.value = ''
  actionError.value = ''
}

async function submitShipped(id: string) {
  actionLoading.value = true
  actionError.value = ''
  try {
    await api.put(`/api/v1/admin/orders/${id}/status`, { status: 'shipped' })
    if (trackingCode.value) {
      await api.put(`/api/v1/admin/orders/${id}/tracking`, { tracking_code: trackingCode.value })
    }
    orders.value = orders.value.filter(o => o.id !== id)
    shippingOrderId.value = null
    successOrderId.value = id
    setTimeout(() => { successOrderId.value = null }, 3000)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Failed to update order'
  } finally {
    actionLoading.value = false
  }
}

async function approvePix(id: string) {
  actionLoading.value = true
  actionError.value = ''
  try {
    await api.put(`/api/v1/admin/orders/${id}/approve-pix`)
    orders.value = orders.value.filter(o => o.id !== id)
    successOrderId.value = id
    setTimeout(() => { successOrderId.value = null }, 3000)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Failed to approve PIX'
  } finally {
    actionLoading.value = false
  }
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString()
}

onMounted(load)
</script>

<style scoped>
.operator-page {
  padding: 32px;
  text-align: left;
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
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.inline-form {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.input {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 13px;
}

.btn {
  padding: 6px 14px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent);
  color: white;
}

.btn-secondary {
  background: none;
  border: 1px solid var(--border);
  color: var(--text-h);
}

.success-badge {
  color: #22c55e;
  font-size: 13px;
  font-weight: 500;
}

.action-error {
  color: #ef4444;
  font-size: 13px;
  width: 100%;
}

.empty {
  padding: 40px;
  text-align: center;
  color: var(--text);
}

.error {
  color: #e53e3e;
  margin-bottom: 12px;
  font-size: 14px;
}
</style>
