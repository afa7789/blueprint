<template>
  <div>
    <h2>Brand Kit</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="success" class="success">Saved successfully.</div>
    <div v-if="loading">Loading...</div>
    <form v-else class="brand-form" @submit.prevent="save">
      <label>Primary Color
        <div class="color-row">
          <input v-model="form.primary_color" type="color" />
          <input v-model="form.primary_color" type="text" class="color-text" />
        </div>
      </label>
      <label>Secondary Color
        <div class="color-row">
          <input v-model="form.secondary_color" type="color" />
          <input v-model="form.secondary_color" type="text" class="color-text" />
        </div>
      </label>
      <label>Logo URL
        <input v-model="form.logo_url" type="text" />
      </label>
      <label>Favicon URL
        <input v-model="form.favicon_url" type="text" />
      </label>
      <label>Font Family
        <input v-model="form.font_family" type="text" placeholder="e.g. Inter, sans-serif" />
      </label>
      <button type="submit" class="btn-primary">Save Brand Kit</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../services/api'

interface BrandKit {
  primary_color: string
  secondary_color: string
  logo_url: string
  favicon_url: string
  font_family: string
}

const form = reactive<BrandKit>({
  primary_color: '#aa3bff',
  secondary_color: '#08060d',
  logo_url: '',
  favicon_url: '',
  font_family: '',
})

const loading = ref(false)
const error = ref('')
const success = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ brand_kit: BrandKit }>('/api/v1/admin/brand-kit')
    Object.assign(form, data.brand_kit)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load brand kit'
  } finally {
    loading.value = false
  }
}

async function save() {
  error.value = ''
  success.value = false
  try {
    await api.put('/api/v1/admin/brand-kit', form)
    success.value = true
    setTimeout(() => { success.value = false }, 3000)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save brand kit'
  }
}

onMounted(load)
</script>

<style scoped>
.brand-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 480px;
  margin-top: 8px;
}

.brand-form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: var(--text);
}

.brand-form input[type="text"] {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 8px 10px;
  border-radius: 4px;
  font-size: 14px;
}

.color-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.color-row input[type="color"] {
  width: 40px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 2px;
  cursor: pointer;
  background: var(--bg);
}

.color-text {
  flex: 1;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  align-self: flex-start;
}

.error {
  color: #e53e3e;
  font-size: 14px;
  margin-bottom: 8px;
}

.success {
  color: #48bb78;
  font-size: 14px;
  margin-bottom: 8px;
}
</style>
