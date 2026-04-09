<template>
  <div class="orders-page">
    <h1>My Orders</h1>

    <div v-if="successMessage" class="success-banner">{{ successMessage }}</div>

    <div v-if="loading" class="loading">Loading orders...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="orders.length === 0" class="empty">
      You have no orders yet. <router-link to="/store">Start shopping</router-link>
    </div>
    <div v-else class="orders-list">
      <div v-for="order in orders" :key="order.id" class="order-card">
        <div class="order-header">
          <div>
            <span class="order-id">#{{ order.id.slice(0, 8) }}</span>
            <span class="order-date">{{ formatDate(order.created_at) }}</span>
          </div>
          <span class="order-status" :class="`status-${order.status}`">{{ order.status }}</span>
        </div>
        <div class="order-body">
          <div class="order-total">
            <span>Total:</span>
            <strong>${{ (order.total / 100).toFixed(2) }}</strong>
          </div>
          <div v-if="order.tracking_code" class="tracking">
            Tracking: <code>{{ order.tracking_code }}</code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../../services/api'

interface Order {
  id: string
  created_at: string
  status: string
  total: number
  tracking_code: string | null
}

const route = useRoute()
const orders = ref<Order[]>([])
const loading = ref(false)
const error = ref('')
const successMessage = ref('')

async function fetchOrders() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Order[] }>('/api/v1/orders/me')
    orders.value = data.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load orders'
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  if (route.query.success) {
    successMessage.value = 'Order placed successfully!'
    setTimeout(() => { successMessage.value = '' }, 5000)
  }
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 32px;
  text-align: left;
  max-width: 800px;
  margin: 0 auto;
}

.success-banner {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid #22c55e;
  border-radius: 6px;
  padding: 12px 16px;
  color: #22c55e;
  margin-bottom: 20px;
}

.loading, .empty {
  padding: 40px;
  text-align: center;
  color: var(--text);
}

.empty a {
  color: var(--accent);
}

.error {
  color: #ef4444;
  padding: 16px;
  border: 1px solid #ef4444;
  border-radius: 6px;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.order-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.order-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--code-bg);
  border-bottom: 1px solid var(--border);
}

.order-id {
  font-family: var(--mono);
  font-weight: 600;
  color: var(--text-h);
  margin-right: 12px;
}

.order-date {
  font-size: 13px;
  color: var(--text);
}

.order-status {
  font-size: 13px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 4px;
  text-transform: capitalize;
}

.status-pending { background: rgba(234, 179, 8, 0.15); color: #ca8a04; }
.status-paid { background: rgba(34, 197, 94, 0.15); color: #16a34a; }
.status-shipped { background: rgba(59, 130, 246, 0.15); color: #2563eb; }
.status-delivered { background: rgba(34, 197, 94, 0.2); color: #15803d; }
.status-cancelled { background: rgba(239, 68, 68, 0.15); color: #dc2626; }

.order-body {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.order-total {
  font-size: 15px;
  color: var(--text);
  display: flex;
  gap: 8px;
  align-items: center;
}

.order-total strong {
  color: var(--text-h);
}

.tracking {
  font-size: 13px;
  color: var(--text);
}

@media (max-width: 600px) {
  .orders-page {
    padding: 20px;
  }
}
</style>
