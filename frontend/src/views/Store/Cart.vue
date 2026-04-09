<template>
  <div class="cart-page">
    <h1>Cart</h1>

    <div v-if="cart.items.length === 0" class="empty-cart">
      Your cart is empty. <router-link to="/store">Continue shopping</router-link>
    </div>

    <div v-else class="cart-layout">
      <div class="cart-items">
        <div v-for="item in cart.items" :key="item.productId" class="cart-item">
          <img v-if="item.imageUrl" :src="item.imageUrl" :alt="item.name" class="item-image" />
          <div v-else class="item-image-placeholder"></div>
          <div class="item-details">
            <p class="item-name">{{ item.name }}</p>
            <p class="item-unit-price">${{ (item.price / 100).toFixed(2) }} each</p>
          </div>
          <div class="item-qty">
            <button @click="cart.updateQuantity(item.productId, item.quantity - 1)" class="qty-btn">-</button>
            <span>{{ item.quantity }}</span>
            <button @click="cart.updateQuantity(item.productId, item.quantity + 1)" class="qty-btn">+</button>
          </div>
          <p class="item-total">${{ (item.price * item.quantity / 100).toFixed(2) }}</p>
          <button @click="cart.removeItem(item.productId)" class="remove-btn">Remove</button>
        </div>
      </div>

      <div class="cart-summary">
        <div class="coupon-section">
          <h3>Coupon Code</h3>
          <div class="coupon-row">
            <input
              v-model="couponInput"
              type="text"
              placeholder="Enter coupon code"
              class="input"
              :disabled="couponApplied"
            />
            <button @click="applyCoupon" class="btn btn-ghost" :disabled="couponApplied || !couponInput">
              Apply
            </button>
          </div>
          <p v-if="couponError" class="coupon-error">{{ couponError }}</p>
          <p v-if="couponApplied" class="coupon-success">Coupon applied!</p>
        </div>

        <div class="summary-lines">
          <div class="summary-line">
            <span>Subtotal</span>
            <span>${{ (cart.subtotal / 100).toFixed(2) }}</span>
          </div>
          <div v-if="cart.discount > 0" class="summary-line discount">
            <span>Discount</span>
            <span>-${{ (cart.discount / 100).toFixed(2) }}</span>
          </div>
          <div class="summary-line total">
            <span>Total</span>
            <span>${{ (cart.total / 100).toFixed(2) }}</span>
          </div>
        </div>

        <router-link
          v-if="auth.isAuthenticated"
          to="/store/checkout"
          class="btn btn-primary checkout-btn"
        >
          Proceed to Checkout
        </router-link>
        <router-link v-else to="/login" class="btn btn-primary checkout-btn">
          Login to Checkout
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCartStore } from '../../stores/cart'
import { useAuthStore } from '../../stores/auth'
import { api } from '../../services/api'

const cart = useCartStore()
const auth = useAuthStore()

const couponInput = ref('')
const couponError = ref('')
const couponApplied = ref(false)

async function applyCoupon() {
  couponError.value = ''
  try {
    const data = await api.post<{ discount: number; code: string }>('/api/v1/coupons/validate', {
      code: couponInput.value,
    })
    cart.discount = data.discount
    cart.couponCode = data.code
    couponApplied.value = true
  } catch (e: unknown) {
    couponError.value = e instanceof Error ? e.message : 'Invalid coupon code'
  }
}
</script>

<style scoped>
.cart-page {
  padding: 32px;
  text-align: left;
  max-width: 900px;
  margin: 0 auto;
}

.empty-cart {
  padding: 40px;
  text-align: center;
  color: var(--text);
}

.empty-cart a {
  color: var(--accent);
}

.cart-layout {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 32px;
  align-items: start;
}

@media (max-width: 768px) {
  .cart-layout {
    grid-template-columns: 1fr;
  }
  .cart-page {
    padding: 20px;
  }
}

.cart-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
  border-bottom: 1px solid var(--border);
}

.item-image {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--border);
}

.item-image-placeholder {
  width: 64px;
  height: 64px;
  background: var(--code-bg);
  border-radius: 6px;
  flex-shrink: 0;
}

.item-details {
  flex: 1;
}

.item-name {
  margin: 0;
  font-weight: 500;
  color: var(--text-h);
}

.item-unit-price {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text);
}

.item-qty {
  display: flex;
  align-items: center;
  gap: 8px;
}

.qty-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border);
  background: var(--bg);
  border-radius: 4px;
  cursor: pointer;
  color: var(--text-h);
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qty-btn:hover {
  background: var(--code-bg);
}

.item-total {
  margin: 0;
  font-weight: 600;
  color: var(--text-h);
  min-width: 60px;
  text-align: right;
}

.remove-btn {
  background: none;
  border: none;
  color: #ef4444;
  cursor: pointer;
  font-size: 13px;
  padding: 4px 8px;
  border-radius: 4px;
}

.remove-btn:hover {
  background: rgba(239, 68, 68, 0.1);
}

.cart-summary {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
}

.coupon-section {
  margin-bottom: 20px;
}

.coupon-section h3 {
  margin: 0 0 12px;
  font-size: 16px;
  color: var(--text-h);
}

.coupon-row {
  display: flex;
  gap: 8px;
}

.input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 14px;
}

.input:focus {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.coupon-error {
  color: #ef4444;
  font-size: 13px;
  margin: 8px 0 0;
}

.coupon-success {
  color: #22c55e;
  font-size: 13px;
  margin: 8px 0 0;
}

.summary-lines {
  border-top: 1px solid var(--border);
  padding-top: 16px;
  margin-bottom: 20px;
}

.summary-line {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  color: var(--text);
  font-size: 14px;
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
  margin-top: 8px;
}

.btn {
  padding: 10px 20px;
  border-radius: 6px;
  border: 1px solid var(--border);
  cursor: pointer;
  font-size: 14px;
  background: var(--bg);
  color: var(--text);
  transition: box-shadow 0.2s;
  text-decoration: none;
  display: inline-block;
  text-align: center;
}

.btn:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-ghost {
  background: transparent;
}

.checkout-btn {
  width: 100%;
  box-sizing: border-box;
}
</style>
