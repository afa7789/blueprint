<template>
  <div class="checkout-page">
    <h1>Checkout</h1>

    <div v-if="cart.items.length === 0" class="empty">
      Your cart is empty. <router-link to="/store">Go shopping</router-link>
    </div>

    <div v-else class="checkout-layout">
      <form @submit.prevent="placeOrder" class="checkout-form">
        <section class="form-section">
          <h2>Shipping Address</h2>
          <div class="form-row">
            <div class="form-group">
              <label>Full Name</label>
              <input v-model="shipping.name" type="text" class="input" required />
            </div>
          </div>
          <div class="form-group">
            <label>Street Address</label>
            <input v-model="shipping.street" type="text" class="input" required />
          </div>
          <div class="form-row three-col">
            <div class="form-group">
              <label>City</label>
              <input v-model="shipping.city" type="text" class="input" required />
            </div>
            <div class="form-group">
              <label>State</label>
              <input v-model="shipping.state" type="text" class="input" required />
            </div>
            <div class="form-group">
              <label>ZIP Code</label>
              <input v-model="shipping.zip" type="text" class="input" required />
            </div>
          </div>
        </section>

        <section class="form-section">
          <h2>Payment Method</h2>
          <p v-if="paymentMethods.length === 0" class="empty-payment">
            No payment methods are available right now.
          </p>
          <div v-else class="payment-options">
            <label
              v-for="method in paymentMethods"
              :key="method.value"
              class="payment-option"
              :class="{ selected: paymentMethod === method.value }"
            >
              <input
                v-model="paymentMethod"
                type="radio"
                :value="method.value"
                name="payment"
              />
              <span>{{ method.label }}</span>
            </label>
          </div>
        </section>

        <div v-if="orderError" class="error">{{ orderError }}</div>

        <div v-if="orderResult" class="payment-result">
          <template v-if="orderResult.method === 'stripe'">
            <p>Stripe payment was initialized for order <span class="mono">{{ orderResult.orderId }}</span>.</p>
            <p class="payment-note">The visual card checkout still needs the final UI hookup, but the payment intent is already created.</p>
          </template>
          <template v-else-if="orderResult.method === 'pix_manual'">
            <p class="pix-title">PIX Payment</p>
            <p v-if="orderResult.beneficiary">Beneficiary: {{ orderResult.beneficiary }}</p>
            <p v-if="orderResult.amount != null">Amount: {{ formatCurrency(Number(orderResult.amount)) }}</p>
            <div class="pix-qr-container">
              <img v-if="qrImageUrl" :src="qrImageUrl" alt="PIX QR Code" class="pix-qr-img" />
              <p v-else class="pix-qr">Generating QR code...</p>
            </div>
            <p v-if="orderResult.brcode" class="pix-copy">
              <button type="button" class="btn btn-secondary" @click="copyBrcode">Copy PIX code</button>
            </p>
            <p>Transfer the order total via PIX, then upload your payment receipt below.</p>
            <p class="pix-txid">Order ID: <code class="mono">{{ orderResult.orderId }}</code></p>

            <div v-if="!receiptUploaded" class="receipt-upload">
              <h4>Upload Payment Receipt</h4>
              <input type="file" accept="image/*,.pdf" @change="handleReceiptFile" class="input" />
              <button v-if="receiptFile" type="button" @click="uploadReceipt" :disabled="uploadingReceipt" class="btn btn-primary receipt-button">
                {{ uploadingReceipt ? 'Uploading...' : 'Submit Receipt' }}
              </button>
              <p v-if="receiptError" class="receipt-error">{{ receiptError }}</p>
            </div>

            <div v-else class="receipt-success">
              <p class="receipt-success-title">Receipt uploaded! Your order is awaiting admin approval.</p>
              <p>You will be notified when your payment is confirmed.</p>
            </div>
          </template>
        </div>

        <button v-if="!orderResult" type="submit" class="btn btn-primary" :disabled="submitting || !paymentMethod">
          {{ submitting ? 'Placing Order...' : 'Place Order' }}
        </button>
      </form>

      <div class="order-summary">
        <h2>Order Summary</h2>
        <div v-for="item in cart.items" :key="item.productId" class="summary-item">
          <span>{{ item.name }} × {{ item.quantity }}</span>
          <span>{{ formatCurrency(item.price * item.quantity) }}</span>
        </div>
        <div class="summary-divider"></div>
        <div class="summary-line">
          <span>Subtotal</span>
          <span>{{ formatCurrency(cart.subtotal) }}</span>
        </div>
        <div v-if="cart.discount > 0" class="summary-line discount">
          <span>Discount</span>
          <span>-{{ formatCurrency(cart.discount) }}</span>
        </div>
        <div class="summary-line total">
          <span>Total</span>
          <span>{{ formatCurrency(cart.total) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import QRCode from 'qrcode'
import { api } from '../../services/api'
import { fetchFeatureFlags, isFeatureEnabled } from '../../services/featureFlags'
import { formatCurrency } from '../../utils/currency'
import { useCartStore } from '../../stores/cart'

interface PaymentMethodOption {
  value: string
  label: string
  flag: string
  extraFlag?: string
}

const cart = useCartStore()
const router = useRouter()

const shipping = ref({ name: '', street: '', city: '', state: '', zip: '' })
const paymentMethod = ref('')
const submitting = ref(false)
const orderError = ref('')
const orderResult = ref<{ method: string; orderId: string; clientSecret?: string; txId?: string; brcode?: string; beneficiary?: string; amount?: number } | null>(null)
const qrImageUrl = ref('')
const receiptFile = ref<File | null>(null)
const uploadingReceipt = ref(false)
const receiptUploaded = ref(false)
const receiptError = ref('')

const allPaymentMethods: PaymentMethodOption[] = [
  { value: 'stripe', label: 'Credit Card (Stripe)', flag: 'payments_stripe' },
  { value: 'pix_manual', label: 'PIX Manual', flag: 'payments_pix', extraFlag: 'pix_manual_enabled' },
]

const paymentMethods = computed(() =>
  allPaymentMethods.filter(method =>
    isFeatureEnabled(method.flag) && (!method.extraFlag || isFeatureEnabled(method.extraFlag)),
  )
)

onMounted(async () => {
  await fetchFeatureFlags()
  if (paymentMethods.value.length > 0) {
    paymentMethod.value = paymentMethods.value[0].value
  }
})

function handleReceiptFile(e: Event) {
  const input = e.target as HTMLInputElement
  receiptFile.value = input.files?.[0] || null
}

async function uploadReceipt() {
  if (!receiptFile.value || !orderResult.value) return
  uploadingReceipt.value = true
  receiptError.value = ''
  try {
    const formData = new FormData()
    formData.append('receipt', receiptFile.value)
    const API_URL = import.meta.env.VITE_API_URL ?? ''
    const res = await fetch(`${API_URL}/api/v1/payments/pix/${orderResult.value.orderId}/receipt`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    })
    if (!res.ok) throw new Error('Upload failed')
    receiptUploaded.value = true
  } catch (e: unknown) {
    receiptError.value = e instanceof Error ? e.message : 'Failed to upload receipt'
  } finally {
    uploadingReceipt.value = false
  }
}

function copyBrcode() {
  if (orderResult.value?.brcode) {
    navigator.clipboard.writeText(orderResult.value.brcode)
  }
}

async function placeOrder() {
  if (!paymentMethod.value) {
    orderError.value = 'Select a payment method to continue'
    return
  }

  submitting.value = true
  orderError.value = ''
  orderResult.value = null
  qrImageUrl.value = ''
  receiptFile.value = null
  receiptUploaded.value = false
  receiptError.value = ''
  try {
    const data = await api.post<{ id: string }>('/api/v1/orders', {
      items: cart.items.map(i => ({ product_id: i.productId, quantity: i.quantity })),
      shipping_address: shipping.value,
      payment_method: paymentMethod.value,
      coupon_code: cart.couponCode || undefined,
    })
    const orderId = data.id

    if (paymentMethod.value === 'stripe') {
      const stripeRes = await api.post<{ client_secret: string }>('/api/v1/payments/stripe', { order_id: orderId })
      // TODO: Use stripeRes.client_secret to complete payment via Stripe Elements/PaymentElement
      orderResult.value = { method: 'stripe', orderId, clientSecret: stripeRes.client_secret }
      cart.clear()
      return
    }

    if (paymentMethod.value === 'pix_manual') {
      const pix = await api.post<{ tx_id: string; brcode?: string; beneficiary?: string; amount?: number }>('/api/v1/payments/pix', { order_id: orderId })
      orderResult.value = { method: 'pix_manual', orderId, txId: pix.tx_id, brcode: pix.brcode, beneficiary: pix.beneficiary, amount: pix.amount }
      if (pix.brcode) {
        try {
          qrImageUrl.value = await QRCode.toDataURL(pix.brcode, { width: 256, margin: 2 })
        } catch {
          qrImageUrl.value = ''
        }
      }
      cart.clear()
      return
    }

    cart.clear()
    router.push({ path: '/store/orders', query: { success: '1' } })
  } catch (e: unknown) {
    orderError.value = e instanceof Error ? e.message : 'Failed to place order'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.checkout-page {
  padding: 32px;
  text-align: left;
  max-width: 1000px;
  margin: 0 auto;
}

.empty {
  padding: 40px;
  text-align: center;
  color: var(--text);
}

.empty a {
  color: var(--accent);
}

.checkout-layout {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 40px;
  align-items: start;
}

@media (max-width: 768px) {
  .checkout-layout {
    grid-template-columns: 1fr;
  }
  .checkout-page {
    padding: 20px;
  }
}

.form-section {
  margin-bottom: 32px;
}

.form-section h2 {
  margin-bottom: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.form-group label {
  font-size: 14px;
  color: var(--text);
  font-weight: 500;
}

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.form-row.three-col {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
}

.form-group {
  margin-bottom: 16px;
}

.input {
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 14px;
  width: 100%;
  box-sizing: border-box;
}

.input:focus {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.empty-payment,
.payment-note {
  color: var(--text);
  font-size: 14px;
}

.payment-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.payment-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.payment-option.selected {
  border-color: var(--accent);
  background: var(--accent-bg);
}

.btn {
  padding: 12px 24px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: box-shadow 0.2s;
}

.btn-primary {
  background: var(--accent);
  color: white;
  width: 100%;
}

.btn-primary:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.order-summary {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
}

.order-summary h2 {
  margin-bottom: 16px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: var(--text);
  margin-bottom: 8px;
}

.summary-divider {
  border-top: 1px solid var(--border);
  margin: 12px 0;
}

.summary-line {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--text);
}

.summary-line.discount {
  color: #22c55e;
}

.summary-line.total {
  font-weight: 700;
  font-size: 16px;
  color: var(--text-h);
  border-top: 1px solid var(--border);
  padding-top: 8px;
  margin-top: 4px;
}

.error {
  color: #ef4444;
  padding: 12px;
  border: 1px solid #ef4444;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 14px;
}

.payment-result {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  font-size: 14px;
  color: var(--text);
}

.pix-title {
  font-weight: 600;
  font-size: 16px;
  color: var(--text-h);
  margin-bottom: 12px;
}

.pix-qr-container {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;
}

.pix-qr-img {
  width: 256px;
  height: 256px;
  border: 1px solid var(--border);
  border-radius: 8px;
}

.btn-secondary {
  background: var(--code-bg);
  color: var(--text-h);
  border: 1px solid var(--border);
}

.pix-copy {
  margin-bottom: 12px;
}

.pix-qr {
  padding: 24px;
  border: 2px dashed var(--border);
  border-radius: 8px;
  text-align: center;
  color: var(--text);
  margin-bottom: 12px;
}

.pix-txid {
  color: var(--text);
}

.mono {
  font-family: var(--mono);
  font-size: 13px;
}

.receipt-upload {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
}

.receipt-upload h4 {
  margin-bottom: 10px;
  font-size: 14px;
  color: var(--text-h);
}

.receipt-button {
  margin-top: 10px;
}

.receipt-error {
  color: #ef4444;
  font-size: 13px;
  margin-top: 8px;
}

.receipt-success {
  margin-top: 16px;
  padding: 16px;
  border: 1px solid #22c55e;
  border-radius: 6px;
  background: rgba(34, 197, 94, 0.05);
}

.receipt-success-title {
  color: #22c55e;
  font-weight: 600;
}
</style>