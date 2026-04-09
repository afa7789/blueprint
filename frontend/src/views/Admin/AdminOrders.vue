<template>
  <div class="admin-orders">
    <div class="page-header">
      <h1>Orders</h1>
    </div>

    <div class="filter-row">
      <select v-model="statusFilter" class="input input-sm">
        <option value="">All Statuses</option>
        <option value="pending">Pending</option>
        <option value="paid">Paid</option>
        <option value="shipped">Shipped</option>
        <option value="delivered">Delivered</option>
        <option value="cancelled">Cancelled</option>
      </select>
    </div>

    <div v-if="loading" class="loading">Loading orders...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th>Order ID</th>
          <th>Customer</th>
          <th>Total</th>
          <th>Status</th>
          <th>Date</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="filteredOrders.length === 0">
          <td colspan="6" class="empty-row">No orders found.</td>
        </tr>
        <tr v-for="order in filteredOrders" :key="order.id">
          <td class="order-id">{{ order.id.slice(0, 8) }}...</td>
          <td>{{ order.user_email || '-' }}</td>
          <td>${{ (order.total / 100).toFixed(2) }}</td>
          <td>
            <span class="status-badge" :class="`status-${order.status}`">{{ order.status }}</span>
          </td>
          <td>{{ formatDate(order.created_at) }}</td>
          <td class="actions">
            <select v-model="order.status" @change="updateStatus(order)" class="input input-sm status-select">
              <option value="pending">Pending</option>
              <option value="paid">Paid</option>
              <option value="shipped">Shipped</option>
              <option value="delivered">Delivered</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../services/api'

interface Order {
  id: string
  user_email: string | null
  total: number
  status: string
  created_at: string
}

const orders = ref<Order[]>([])
const loading = ref(false)
const error = ref('')
const statusFilter = ref('')

const filteredOrders = computed(() => {
  if (!statusFilter.value) return orders.value
  return orders.value.filter(o => o.status === statusFilter.value)
})

async function fetchOrders() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Order[] }>('/api/v1/admin/orders')
    orders.value = data.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load orders'
  } finally {
    loading.value = false
  }
}

async function updateStatus(order: Order) {
  try {
    await api.put(`/api/v1/admin/orders/${order.id}/status`, { status: order.status })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update status'
    await fetchOrders()
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(fetchOrders)
</script>

<style scoped>
.admin-orders {
  padding: 32px;
  text-align: left;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.page-header h1 {
  margin: 0;
}

.filter-row {
  margin-bottom: 20px;
}

.input-sm {
  padding: 6px 10px;
  font-size: 13px;
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

.order-id {
  font-family: monospace;
  font-size: 12px;
}

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}

.status-pending { background: rgba(234, 179, 8, 0.15); color: #ca8a04; }
.status-paid { background: rgba(34, 197, 94, 0.15); color: #16a34a; }
.status-shipped { background: rgba(59, 130, 246, 0.15); color: #2563eb; }
.status-delivered { background: rgba(34, 197, 94, 0.2); color: #15803d; }
.status-cancelled { background: rgba(239, 68, 68, 0.15); color: #dc2626; }

.actions {
  display: flex;
  gap: 8px;
}

.status-select {
  width: auto;
}

.input {
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
</style>
