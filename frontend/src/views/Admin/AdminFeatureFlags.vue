<template>
  <div>
    <h2>Feature Flags</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>
    <template v-else>
      <div class="flag-list">
        <div v-for="flag in flags" :key="flag.key" class="flag-row">
          <div class="flag-info">
            <span class="flag-key">{{ flag.key }}</span>
            <span class="flag-status" :class="flag.enabled ? 'enabled' : 'disabled'">
              {{ flag.enabled ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
          <label class="toggle">
            <input type="checkbox" :checked="flag.enabled" @change="toggle(flag, ($event.target as HTMLInputElement).checked)" />
            <span class="slider"></span>
          </label>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Flag {
  key: string
  enabled: boolean
}

const flags = ref<Flag[]>([])
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ features: Flag[] }>('/api/v1/admin/features')
    flags.value = data.features
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load feature flags'
  } finally {
    loading.value = false
  }
}

async function toggle(flag: Flag, enabled: boolean) {
  try {
    await api.put(`/api/v1/admin/features/${flag.key}`, { enabled })
    flag.enabled = enabled
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update flag'
  }
}

onMounted(load)
</script>

<style scoped>
.flag-list {
  display: flex;
  flex-direction: column;
  gap: 0;
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

.flag-status.enabled {
  background: rgba(72, 187, 120, 0.15);
  color: #48bb78;
}

.flag-status.disabled {
  background: var(--code-bg);
  color: var(--text);
}

.toggle {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  cursor: pointer;
}

.toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

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

.toggle input:checked + .slider {
  background: var(--accent);
}

.toggle input:checked + .slider::before {
  transform: translateX(20px);
}

.error {
  color: #e53e3e;
  margin-bottom: 12px;
  font-size: 14px;
}
</style>
