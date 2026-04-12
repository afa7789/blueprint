<template>
  <div class="admin-products">
    <HelperBox title="Products" description="Manage your store products. Add individually or bulk import via CSV." featureFlag="store_enabled" />
    <div class="page-header">
      <h1>Products ({{ products.length }})</h1>
      <div class="header-actions">
        <button @click="showImport = true" class="btn"><i class="fas fa-file-csv"></i> Import CSV</button>
        <button @click="openCreate" class="btn btn-primary"><i class="fas fa-plus"></i> Add Product</button>
      </div>
    </div>

    <div class="filter-row">
      <select v-model="selectedCategory" class="input input-sm">
        <option value="">All Categories</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
      </select>
    </div>

    <div v-if="loading" class="loading">Loading products...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Category</th>
          <th>Price</th>
          <th>Stock</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="filteredProducts.length === 0">
          <td colspan="6" class="empty-row">No products yet.</td>
        </tr>
        <tr v-for="product in filteredProducts" :key="product.id">
          <td>{{ product.name }}</td>
          <td>{{ categoryName(product.category_id) }}</td>
          <td>{{ formatCurrency(product.price) }}</td>
          <td>{{ product.stock }}</td>
          <td><span class="status-badge" :class="product.is_active ? 'status-active' : 'status-inactive'">{{ product.is_active ? 'Active' : 'Inactive' }}</span></td>
          <td class="actions">
            <button @click="openEdit(product)" class="btn btn-ghost btn-sm">Edit</button>
            <button @click="deleteProduct(product.id)" class="btn btn-danger btn-sm">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Form Modal -->
    <div v-if="showForm" class="modal-overlay" @click.self="closeForm">
      <div class="modal">
        <h2>{{ editingProduct ? 'Edit Product' : 'New Product' }}</h2>

        <div class="form-group">
          <label>Name</label>
          <input v-model="form.name" type="text" class="input" placeholder="Product name" />
        </div>

        <div class="form-group">
          <label>Description</label>
          <textarea v-model="form.description" class="input textarea" rows="3" placeholder="Description..."></textarea>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Price (R$)</label>
            <input v-model="form.priceDisplay" type="text" class="input" placeholder="25,00" />
          </div>
          <div class="form-group">
            <label>Stock</label>
            <input v-model.number="form.stock" type="number" class="input" placeholder="0" />
          </div>
        </div>

        <div class="form-group">
          <label>Category</label>
          <select v-model="form.category_id" class="input">
            <option value="">No category</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
        </div>

        <div class="form-group">
          <label>Image URL</label>
          <input v-model="form.image_url" type="text" class="input" placeholder="https://..." />
        </div>

        <div class="form-group checkbox-group">
          <label>
            <input type="checkbox" v-model="form.active" />
            Active (visible to customers)
          </label>
        </div>

        <p v-if="formError" class="field-error">{{ formError }}</p>

        <div class="modal-actions">
          <button @click="closeForm" class="btn btn-ghost">Cancel</button>
          <button @click="submitForm" class="btn btn-primary" :disabled="formSubmitting">
            {{ formSubmitting ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>
    <!-- CSV Import Modal -->
    <div v-if="showImport" class="modal-overlay" @click.self="showImport = false">
      <div class="modal">
        <h2><i class="fas fa-file-csv"></i> Import Products from CSV</h2>
        <p class="import-hint">Format: <code>name,description,price(centavos)</code> — one product per line</p>
        <pre class="import-example">Cerveja artesanal,IPA local,1500
Camiseta Blueprint,Tamanho único,3500</pre>
        <textarea v-model="csvText" class="input textarea" rows="8" placeholder="Paste CSV here..."></textarea>
        <p v-if="importError" class="field-error">{{ importError }}</p>
        <p v-if="importSuccess" class="import-success">{{ importSuccess }}</p>
        <div class="modal-actions">
          <button @click="showImport = false" class="btn btn-ghost">Cancel</button>
          <button @click="importCSV" class="btn btn-primary" :disabled="!csvText.trim() || importing">
            {{ importing ? 'Importing...' : 'Import' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'
import { formatCurrency } from '../../utils/currency'

interface Category {
  id: string
  name: string
}

interface Product {
  id: string
  name: string
  description: string
  price: number
  stock: number
  category_id: string | null
  images: string[] | null
  is_active: boolean
  emoji?: string
}


const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const loading = ref(false)
const error = ref('')
const showForm = ref(false)
const editingProduct = ref<Product | null>(null)
const formSubmitting = ref(false)
const formError = ref('')
const selectedCategory = ref('')

const showImport = ref(false)
const csvText = ref('')
const importError = ref('')
const importSuccess = ref('')
const importing = ref(false)

const form = ref({
  name: '',
  description: '',
  priceDisplay: '',
  stock: 0,
  category_id: '',
  image_url: '',
  active: true,
})

const filteredProducts = computed(() => {
  if (!selectedCategory.value) return products.value
  return products.value.filter(p => p.category_id === selectedCategory.value)
})

function categoryName(id: string | null) {
  if (!id) return '-'
  return categories.value.find(c => c.id === id)?.name || id
}

async function fetchData() {
  loading.value = true
  error.value = ''
  try {
    const [productsData, categoriesData] = await Promise.all([
      api.get<{ data: Product[] }>('/api/v1/admin/products'),
      api.get<{ data: Category[] }>('/api/v1/admin/categories'),
    ])
    products.value = productsData.data || []
    categories.value = categoriesData.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load data'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingProduct.value = null
  form.value = { name: '', description: '', priceDisplay: '', stock: 0, category_id: '', image_url: '', active: true }
  formError.value = ''
  showForm.value = true
}

function openEdit(product: Product) {
  editingProduct.value = product
  form.value = {
    name: product.name,
    description: product.description,
    priceDisplay: product.price.toFixed(2).replace('.', ','),
    stock: product.stock,
    category_id: product.category_id || '',
    image_url: product.images?.[0] || '',
    active: product.is_active,
  }
  formError.value = ''
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingProduct.value = null
}

function parsePrice(display: string): number {
  return parseFloat(display.replace(',', '.')) || 0
}

async function submitForm() {
  formSubmitting.value = true
  formError.value = ''
  try {
    const price = parsePrice(form.value.priceDisplay)
    const payload = {
      name: form.value.name,
      description: form.value.description || null,
      price,
      stock: form.value.stock,
      is_active: form.value.active,
      category_id: form.value.category_id || null,
      images: form.value.image_url ? [form.value.image_url] : [],
    }
    if (editingProduct.value) {
      await api.put(`/api/v1/admin/products/${editingProduct.value.id}`, payload)
    } else {
      await api.post('/api/v1/admin/products', payload)
    }
    await fetchData()
    closeForm()
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : 'Failed to save product'
  } finally {
    formSubmitting.value = false
  }
}

async function importCSV() {
  importing.value = true
  importError.value = ''
  importSuccess.value = ''
  const lines = csvText.value.trim().split('\n').filter(l => l.trim())
  let imported = 0
  const errors: string[] = []

  for (const line of lines) {
    const parts = line.split(',')
    if (parts.length < 3) { errors.push(`Invalid line: ${line}`); continue }
    const [name, description, priceStr] = parts
    const price = parsePrice(priceStr.trim()) / 100
    if (price <= 0) { errors.push(`Invalid price: ${line}`); continue }
    try {
      await api.post('/api/v1/admin/products', {
        name: name.trim(),
        description: description.trim() || null,
        price,
        stock: 0,
        is_active: true,
        images: [],
      })
      imported++
    } catch { errors.push(`Failed: ${name}`) }
  }

  if (imported > 0) {
    importSuccess.value = `${imported} product(s) imported.`
    await fetchData()
    csvText.value = ''
  }
  if (errors.length) importError.value = errors.join('; ')
  importing.value = false
}

async function deleteProduct(id: string) {
  if (!confirm('Delete this product?')) return
  try {
    await api.delete(`/api/v1/admin/products/${id}`)
    await fetchData()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete product'
  }
}

onMounted(fetchData)
</script>

<style scoped>
.admin-products {
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

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.status-active { background: rgba(34, 197, 94, 0.15); color: #16a34a; }
.status-inactive { background: rgba(107, 114, 128, 0.15); color: #6b7280; }

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
  max-width: 560px;
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

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: normal;
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

.textarea {
  resize: vertical;
  min-height: 80px;
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

.header-actions {
  display: flex;
  gap: 8px;
}

.import-hint {
  font-size: 13px;
  color: var(--text);
  margin: 0 0 8px;
}

.import-hint code {
  font-size: 12px;
  padding: 2px 6px;
  background: var(--code-bg);
  border-radius: 3px;
}

.import-example {
  font-size: 12px;
  background: var(--code-bg);
  padding: 12px;
  border-radius: 6px;
  margin: 0 0 12px;
  overflow-x: auto;
  font-family: var(--mono);
  color: var(--text);
}

.import-success {
  color: #16a34a;
  font-size: 13px;
}
</style>
