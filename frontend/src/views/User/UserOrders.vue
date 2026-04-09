<template>
  <div class="user-orders">
    <h2>Orders</h2>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <table v-else-if="orders.length" class="orders-table">
      <thead>
        <tr>
          <th>Order ID</th>
          <th>Date</th>
          <th>Status</th>
          <th>Total</th>
          <th>Tracking</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="order in orders" :key="order.id">
          <td class="mono">#{{ order.id }}</td>
          <td>{{ formatDate(order.created_at) }}</td>
          <td><span class="badge" :class="`badge-${order.status}`">{{ order.status }}</span></td>
          <td>${{ order.total.toFixed(2) }}</td>
          <td>
            <a
              v-if="isUrl(order.tracking_code)"
              :href="order.tracking_code ?? ''"
              target="_blank"
              rel="noopener noreferrer"
              class="tracking-link"
            >{{ order.tracking_code }}</a>
            <span v-else>{{ order.tracking_code || '—' }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <p v-else class="empty">No orders yet.</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Order {
  id: number
  created_at: string
  status: string
  total: number
  tracking_code: string | null | undefined
}

const loading = ref(true)
const error = ref('')
const orders = ref<Order[]>([])

onMounted(async () => {
  try {
    const data = await api.get<{ data: Order[]; total: number }>('/api/v1/orders/me')
    orders.value = data.data ?? []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load orders'
  } finally {
    loading.value = false
  }
})

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString()
}

function isUrl(val: string | null | undefined): boolean {
  if (!val) return false
  try {
    const u = new URL(val)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}
</script>

<style scoped>
.user-orders h2 {
  margin: 0 0 24px;
}

.orders-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.orders-table th,
.orders-table td {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}

.orders-table th {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text);
}

.mono {
  font-family: monospace;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 500;
}

.badge-pending  { background: #6b728022; color: #6b7280; }
.badge-paid     { background: #3b82f622; color: #3b82f6; }
.badge-shipped  { background: #f59e0b22; color: #d97706; }
.badge-delivered { background: #16a34a22; color: #16a34a; }
.badge-cancelled { background: #dc262622; color: #dc2626; }

.tracking-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 13px;
}

.tracking-link:hover {
  text-decoration: underline;
}

.empty {
  color: var(--text);
  font-size: 14px;
}

.error {
  color: var(--error, #dc2626);
  font-size: 14px;
}
</style>
