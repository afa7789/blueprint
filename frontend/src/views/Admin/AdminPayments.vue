<template>
  <div>
    <HelperBox
      title="Payments"
      description="Configure payment methods and monitor transactions. Stripe requires a secret key, PIX manual uses admin approval."
      envVar="STRIPE_KEY, STRIPE_WEBHOOK_SECRET"
    />

    <h2>Payment Methods</h2>

    <!-- Section 1: Payment Methods Status -->
    <div class="methods-grid">
      <div class="method-card">
        <div class="method-header">
          <i class="fas fa-credit-card method-icon"></i>
          <span class="method-name">Stripe</span>
          <span
            class="status-dot"
            :class="stripeStatus"
            :title="stripeStatusLabel"
          ></span>
        </div>
        <div class="method-status-label" :class="stripeStatus">{{ stripeStatusLabel }}</div>
        <div class="method-note">{{ stripeConfigured ? 'STRIPE_KEY is configured' : 'STRIPE_KEY not set — configure in Config / ENV' }}</div>
      </div>

      <div class="method-card">
        <div class="method-header">
          <i class="fas fa-bolt method-icon"></i>
          <span class="method-name">PIX Auto</span>
          <span
            class="status-dot"
            :class="pixAutoFlag ? 'green' : 'gray'"
            :title="pixAutoFlag ? 'Enabled' : 'Disabled'"
          ></span>
        </div>
        <div class="method-status-label" :class="pixAutoFlag ? 'green' : 'gray'">{{ pixAutoFlag ? 'Enabled' : 'Disabled' }}</div>
        <div class="method-note">Requires gateway integration</div>
      </div>

      <div class="method-card">
        <div class="method-header">
          <i class="fas fa-file-invoice method-icon"></i>
          <span class="method-name">PIX Manual</span>
          <span
            class="status-dot"
            :class="pixManualFlag ? 'green' : 'gray'"
            :title="pixManualFlag ? 'Enabled' : 'Disabled'"
          ></span>
        </div>
        <div class="method-status-label" :class="pixManualFlag ? 'green' : 'gray'">{{ pixManualFlag ? 'Enabled' : 'Disabled' }}</div>
        <div class="method-note">Admin approval workflow</div>
      </div>
    </div>

    <!-- Section 2: Quick Actions -->
    <h2>Quick Actions</h2>
    <div class="quick-actions">
      <router-link to="/admin/orders" class="action-card">
        <i class="fas fa-hourglass-half"></i>
        <span>View Orders Awaiting PIX Approval</span>
      </router-link>
      <router-link to="/admin/features" class="action-card">
        <i class="fas fa-toggle-on"></i>
        <span>Manage Feature Flags</span>
      </router-link>
      <router-link to="/admin/config" class="action-card">
        <i class="fas fa-key"></i>
        <span>Configure Stripe Key</span>
      </router-link>
    </div>

    <!-- Section 3: Payment Flags -->
    <h2>Payment Feature Flags</h2>
    <div v-if="flagsError" class="error">{{ flagsError }}</div>
    <div v-if="flagsLoading">Loading flags...</div>
    <div v-else class="flag-list">
      <div v-for="flag in paymentFlags" :key="flag.key" class="flag-row">
        <div class="flag-info">
          <span class="flag-key">{{ flag.key }}</span>
          <span class="flag-status" :class="flag.enabled ? 'enabled' : 'disabled'">
            {{ flag.enabled ? 'Enabled' : 'Disabled' }}
          </span>
        </div>
        <label class="toggle">
          <input type="checkbox" :checked="flag.enabled" @change="toggleFlag(flag, ($event.target as HTMLInputElement).checked)" />
          <span class="slider"></span>
        </label>
      </div>
      <div v-if="paymentFlags.length === 0 && !flagsLoading" class="empty-note">No payment flags found.</div>
    </div>

    <!-- Section 4: Recent Payments -->
    <h2>Recent Payments</h2>
    <div v-if="ordersError" class="error">{{ ordersError }}</div>
    <div v-if="ordersLoading">Loading orders...</div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th>Order ID</th>
          <th>Date</th>
          <th>Payment Method</th>
          <th>Status</th>
          <th>Total</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="recentOrders.length === 0">
          <td colspan="5" class="empty-row">No recent payment activity.</td>
        </tr>
        <tr v-for="order in recentOrders" :key="order.id">
          <td class="order-id">{{ order.id.slice(0, 8) }}...</td>
          <td>{{ formatDate(order.created_at) }}</td>
          <td>{{ order.payment_method || '—' }}</td>
          <td>
            <span class="status-badge" :class="`status-${order.status}`">{{ order.status }}</span>
          </td>
          <td>${{ (order.total / 100).toFixed(2) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface Flag {
  key: string
  enabled: boolean
}

interface Order {
  id: string
  created_at: string
  payment_method: string | null
  status: string
  total: number
}

const PAYMENT_FLAG_KEYS = ['payments_stripe', 'payments_pix', 'pix_auto_enabled', 'pix_manual_enabled']

const allFlags = ref<Flag[]>([])
const flagsLoading = ref(false)
const flagsError = ref('')

const recentOrders = ref<Order[]>([])
const ordersLoading = ref(false)
const ordersError = ref('')

const stripeConfigured = ref(false)

const paymentFlags = computed(() =>
  allFlags.value.filter(f => PAYMENT_FLAG_KEYS.includes(f.key))
)

const stripeFlag = computed(() => allFlags.value.find(f => f.key === 'payments_stripe')?.enabled ?? false)
const pixAutoFlag = computed(() => allFlags.value.find(f => f.key === 'pix_auto_enabled')?.enabled ?? false)
const pixManualFlag = computed(() => allFlags.value.find(f => f.key === 'pix_manual_enabled')?.enabled ?? false)

const stripeStatus = computed(() => {
  if (!stripeFlag.value) return 'gray'
  return stripeConfigured.value ? 'green' : 'yellow'
})

const stripeStatusLabel = computed(() => {
  if (!stripeFlag.value) return 'Disabled'
  return stripeConfigured.value ? 'Enabled & Configured' : 'Enabled (missing STRIPE_KEY)'
})

async function loadFlags() {
  flagsLoading.value = true
  flagsError.value = ''
  try {
    allFlags.value = await api.get<Flag[]>('/api/v1/admin/features')
  } catch (e: unknown) {
    flagsError.value = e instanceof Error ? e.message : 'Failed to load feature flags'
  } finally {
    flagsLoading.value = false
  }
}

async function loadEnv() {
  try {
    const envData = await api.get<{ data: Array<{ key: string; is_set: boolean }> }>('/api/v1/admin/config/env')
    stripeConfigured.value = envData.data.find(e => e.key === 'STRIPE_KEY')?.is_set ?? false
  } catch {
    // non-fatal
  }
}

async function loadOrders() {
  ordersLoading.value = true
  ordersError.value = ''
  try {
    const data = await api.get<{ data: Order[] }>('/api/v1/admin/orders?limit=10')
    const all = data.data || []
    recentOrders.value = all.filter(o => o.payment_method).slice(0, 10)
  } catch (e: unknown) {
    ordersError.value = e instanceof Error ? e.message : 'Failed to load orders'
  } finally {
    ordersLoading.value = false
  }
}

async function toggleFlag(flag: Flag, enabled: boolean) {
  try {
    await api.put(`/api/v1/admin/features/${flag.key}`, { enabled })
    flag.enabled = enabled
  } catch (e: unknown) {
    flagsError.value = e instanceof Error ? e.message : 'Failed to update flag'
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(async () => {
  await Promise.all([loadFlags(), loadEnv(), loadOrders()])
})
</script>

<style scoped>
h2 {
  margin: 28px 0 14px;
  font-size: 16px;
  color: var(--text-h);
}

/* Method Cards */
.methods-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 14px;
  margin-bottom: 8px;
}

.method-card {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 18px;
}

.method-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.method-icon {
  font-size: 16px;
  color: var(--text);
  opacity: 0.7;
  width: 18px;
  text-align: center;
}

.method-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-h);
  flex: 1;
}

.status-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-dot.green { background: #22c55e; }
.status-dot.yellow { background: #f59e0b; }
.status-dot.gray { background: #6b7280; }

.method-status-label {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 6px;
}
.method-status-label.green { color: #22c55e; }
.method-status-label.yellow { color: #f59e0b; }
.method-status-label.gray { color: #6b7280; }

.method-note {
  font-size: 12px;
  color: var(--text);
  opacity: 0.7;
}

/* Quick Actions */
.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 8px;
}

.action-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--accent);
  text-decoration: none;
  font-size: 13px;
  transition: background 0.15s;
}

.action-card:hover {
  background: var(--accent-bg);
  border-color: var(--accent-border);
}

.action-card i {
  font-size: 13px;
  opacity: 0.8;
}

/* Flag list */
.flag-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin-bottom: 8px;
}

.flag-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid var(--border);
}

.flag-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.flag-key {
  font-family: var(--mono);
  font-size: 14px;
  color: var(--text-h);
}

.flag-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}
.flag-status.enabled { background: rgba(72, 187, 120, 0.15); color: #48bb78; }
.flag-status.disabled { background: var(--code-bg); color: var(--text); }

.toggle {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  cursor: pointer;
}
.toggle input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  inset: 0;
  background: var(--border);
  border-radius: 24px;
  transition: background 0.2s;
}
.slider::before {
  content: '';
  position: absolute;
  height: 18px;
  width: 18px;
  left: 3px;
  top: 3px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
}
.toggle input:checked + .slider { background: var(--accent); }
.toggle input:checked + .slider::before { transform: translateX(20px); }

/* Data table */
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
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.data-table tr:hover td { background: var(--social-bg); }

.order-id { font-family: monospace; font-size: 12px; }

.empty-row,
.empty-note {
  text-align: center;
  color: var(--text);
  padding: 24px;
  font-size: 13px;
  opacity: 0.6;
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
.status-awaiting_approval { background: rgba(168, 85, 247, 0.15); color: #9333ea; }

.error {
  color: #ef4444;
  padding: 10px 12px;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 12px;
}
</style>
