<template>
  <div class="store-index">
    <div class="store-header">
      <h1>Store</h1>
      <select v-model="selectedCategory" @change="fetchProducts(1)" class="category-filter">
        <option value="">All Categories</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
      </select>
    </div>

    <div v-if="loading" class="loading">Loading products...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="products.length === 0" class="empty">No products found.</div>
    <div v-else class="product-grid">
      <div
        v-for="product in products"
        :key="product.id"
        class="product-card"
        @click="$router.push(`/store/${product.id}`)"
      >
        <div class="product-image">
          <img
            v-if="product.images && product.images.length"
            :src="product.images[0]"
            :alt="product.name"
          />
          <div v-else class="product-image-placeholder">No Image</div>
          <span v-if="product.is_pre_sale" class="badge badge-presale">Pre-Sale</span>
        </div>
        <div class="product-info">
          <h3 class="product-name">{{ product.name }}</h3>
          <p class="product-price">${{ (product.price / 100).toFixed(2) }}</p>
        </div>
      </div>
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="page === 1" @click="fetchProducts(page - 1)" class="btn btn-ghost">Previous</button>
      <span>Page {{ page }} of {{ totalPages }}</span>
      <button :disabled="page === totalPages" @click="fetchProducts(page + 1)" class="btn btn-ghost">Next</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Product {
  id: string
  name: string
  price: number
  images: string[]
  is_pre_sale: boolean
  category_id: string
}

interface Category {
  id: string
  name: string
}

interface ProductsResponse {
  products: Product[]
  total: number
  page: number
  limit: number
}

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const selectedCategory = ref('')
const loading = ref(false)
const error = ref('')
const page = ref(1)
const totalPages = ref(1)
const limit = 12

async function fetchProducts(p = 1) {
  loading.value = true
  error.value = ''
  try {
    let url = `/api/v1/products?page=${p}&limit=${limit}`
    if (selectedCategory.value) url += `&category_id=${selectedCategory.value}`
    const data = await api.get<ProductsResponse>(url)
    products.value = data.products || []
    page.value = data.page || p
    totalPages.value = Math.ceil((data.total || 0) / limit)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load products'
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  try {
    const data = await api.get<{ categories: Category[] }>('/api/v1/products/categories')
    categories.value = data.categories || []
  } catch {
    // categories optional
  }
}

onMounted(() => {
  fetchProducts()
  fetchCategories()
})
</script>

<style scoped>
.store-index {
  padding: 32px;
  text-align: left;
}

.store-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
}

.store-header h1 {
  margin: 0;
}

.category-filter {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
  cursor: pointer;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}

@media (max-width: 1024px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .product-grid {
    grid-template-columns: 1fr;
  }
  .store-index {
    padding: 20px;
  }
}

.product-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s;
  background: var(--bg);
}

.product-card:hover {
  box-shadow: var(--shadow);
}

.product-image {
  position: relative;
  aspect-ratio: 4/3;
  background: var(--code-bg);
  overflow: hidden;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text);
  font-size: 14px;
}

.badge {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.badge-presale {
  background: var(--accent-bg);
  color: var(--accent);
  border: 1px solid var(--accent-border);
}

.product-info {
  padding: 16px;
}

.product-name {
  margin: 0 0 8px;
  font-size: 16px;
  color: var(--text-h);
  font-weight: 500;
}

.product-price {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--accent);
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 32px;
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
}

.btn:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.loading, .empty {
  color: var(--text);
  padding: 40px;
  text-align: center;
}

.error {
  color: #ef4444;
  padding: 16px;
  border: 1px solid #ef4444;
  border-radius: 6px;
  margin-bottom: 16px;
}
</style>
