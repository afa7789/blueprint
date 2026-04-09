<script setup lang="ts">
import { ref } from 'vue'
import { api, ApiError } from '../../services/api'

const email = ref('')
const status = ref<'idle' | 'loading' | 'success' | 'error'>('idle')
const message = ref('')

async function submitWaitlist() {
  if (!email.value) {
    status.value = 'error'
    message.value = 'Please enter your email'
    return
  }

  status.value = 'loading'
  message.value = ''

  try {
    await api.post('/api/v1/waitlist', { email: email.value })
    status.value = 'success'
    message.value = 'You have been added to the waitlist!'
    email.value = ''
  } catch (e) {
    status.value = 'error'
    message.value = e instanceof ApiError ? e.message : 'Failed to connect to server'
  }
}
</script>

<template>
  <form class="waitlist-form" @submit.prevent="submitWaitlist">
    <div class="input-group">
      <input
        v-model="email"
        type="email"
        placeholder="Enter your email"
        :disabled="status === 'loading'"
      />
      <button type="submit" :disabled="status === 'loading'">
        {{ status === 'loading' ? 'Joining...' : 'Join Waitlist' }}
      </button>
    </div>
    <p v-if="message" :class="['message', status]">{{ message }}</p>
  </form>
</template>

<style scoped>
.waitlist-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 400px;
}

.input-group {
  display: flex;
  gap: 8px;
}

input {
  flex: 1;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 16px;
  background: var(--bg);
  color: var(--text-h);
}

input:focus {
  outline: none;
  border-color: var(--accent);
}

button {
  padding: 12px 20px;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  cursor: pointer;
  transition: opacity 0.2s;
}

button:hover:not(:disabled) {
  opacity: 0.9;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.message {
  font-size: 14px;
}

.message.success {
  color: #22c55e;
}

.message.error {
  color: #ef4444;
}
</style>