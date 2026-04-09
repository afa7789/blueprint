<template>
  <div class="product-detail">
    <router-link to="/store" class="back-link">← Back to Store</router-link>

    <div v-if="loading" class="loading">Loading product...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="product" class="product-layout">
      <div class="product-images">
        <img
          v-if="product.images && product.images.length"
          :src="product.images[0]"
          :alt="product.name"
          class="main-image"
        />
        <div v-else class="image-placeholder">No Image</div>
      </div>
      <div class="product-content">
        <div class="product-badges">
          <span v-if="product.is_pre_sale" class="badge badge-presale">Pre-Sale</span>
        </div>
        <h1>{{ product.name }}</h1>
        <p class="price">${{ (product.price / 100).toFixed(2) }}</p>

        <div v-if="product.is_pre_sale && product.available_at" class="presale-info">
          Available from: {{ formatDate(product.available_at) }}
        </div>

        <div class="stock-status">
          <span v-if="product.stock > 0" class="in-stock">In Stock ({{ product.stock }} available)</span>
          <span v-else-if="product.is_pre_sale" class="pre-sale-available">Available for Pre-Sale</span>
          <span v-else class="out-of-stock">Out of Stock</span>
        </div>

        <p class="description">{{ product.description }}</p>

        <button
          @click="addToCart"
          class="btn btn-primary"
          :disabled="product.stock === 0 && !product.is_pre_sale"
        >
          {{ product.stock === 0 && !product.is_pre_sale ? 'Out of Stock' : 'Add to Cart' }}
        </button>

        <div v-if="addedMessage" class="added-message">{{ addedMessage }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../../services/api'
import { useCartStore } from '../../stores/cart'

interface Product {
  id: string
  name: string
  description: string
  price: number
  stock: number
  images: string[]
  is_pre_sale: boolean
  available_at: string | null
}

const route = useRoute()
const cart = useCartStore()

const product = ref<Product | null>(null)
const loading = ref(false)
const error = ref('')
const addedMessage = ref('')

async function fetchProduct() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<Product>(`/api/v1/products/${route.params.id}`)
    product.value = data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load product'
  } finally {
    loading.value = false
  }
}

function addToCart() {
  if (!product.value) return
  cart.addItem({
    id: product.value.id,
    name: product.value.name,
    price: product.value.price,
    imageUrl: product.value.images?.[0] || '',
  })
  addedMessage.value = 'Added to cart!'
  setTimeout(() => { addedMessage.value = '' }, 2000)
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(fetchProduct)
</script>

<style scoped>
.product-detail {
  padding: 32px;
  text-align: left;
  max-width: 900px;
  margin: 0 auto;
}

.back-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 14px;
  display: inline-block;
  margin-bottom: 24px;
}

.back-link:hover {
  text-decoration: underline;
}

.product-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  align-items: start;
}

@media (max-width: 700px) {
  .product-layout {
    grid-template-columns: 1fr;
  }
  .product-detail {
    padding: 20px;
  }
}

.product-images .main-image {
  width: 100%;
  border-radius: 8px;
  border: 1px solid var(--border);
}

.image-placeholder {
  aspect-ratio: 1;
  background: var(--code-bg);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text);
}

.product-badges {
  margin-bottom: 8px;
}

.badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.badge-presale {
  background: var(--accent-bg);
  color: var(--accent);
  border: 1px solid var(--accent-border);
}

.price {
  font-size: 28px;
  font-weight: 700;
  color: var(--accent);
  margin: 8px 0 16px;
}

.presale-info {
  font-size: 14px;
  color: var(--text);
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--accent-bg);
  border-radius: 6px;
}

.stock-status {
  margin-bottom: 16px;
  font-size: 14px;
}

.in-stock { color: #22c55e; }
.out-of-stock { color: #ef4444; }
.pre-sale-available { color: var(--accent); }

.description {
  color: var(--text);
  line-height: 1.6;
  margin-bottom: 24px;
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
}

.btn-primary:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.added-message {
  margin-top: 12px;
  color: #22c55e;
  font-size: 14px;
}

.loading, .error {
  padding: 40px;
  text-align: center;
}

.error {
  color: #ef4444;
  border: 1px solid #ef4444;
  border-radius: 6px;
}
</style>
