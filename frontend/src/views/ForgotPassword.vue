<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { ApiError } from '../services/api'

const auth = useAuthStore()

const email = ref('')
const error = ref('')
const success = ref(false)
const loading = ref(false)

async function handleForgotPassword() {
  error.value = ''
  loading.value = true
  try {
    await auth.forgotPassword(email.value)
    success.value = true
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Request failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <h2>Reset password</h2>
      <div v-if="success" class="success">
        <p>Check your email for a password reset link.</p>
        <router-link to="/login">Back to sign in</router-link>
      </div>
      <form v-else @submit.prevent="handleForgotPassword">
        <p class="hint">Enter your email and we'll send you a reset link.</p>
        <div class="field">
          <label for="email">Email</label>
          <input id="email" v-model="email" type="email" placeholder="you@example.com" :disabled="loading" required />
        </div>
        <p v-if="error" class="error">{{ error }}</p>
        <button type="submit" :disabled="loading">{{ loading ? 'Sending...' : 'Send reset link' }}</button>
        <div class="links">
          <router-link to="/login">Back to sign in</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100svh;
  padding: 24px;
  box-sizing: border-box;
}

.auth-card {
  width: 100%;
  max-width: 400px;
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 32px;
  box-sizing: border-box;
  text-align: left;
}

.auth-card h2 {
  margin-bottom: 24px;
}

form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.hint {
  font-size: 14px;
  color: var(--text);
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

label {
  font-size: 14px;
  color: var(--text);
}

input {
  padding: 10px 14px;
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
  padding: 12px;
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

.error {
  color: #ef4444;
  font-size: 14px;
}

.success {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.success p {
  color: #22c55e;
}

.links {
  text-align: center;
  font-size: 14px;
}

.links a,
.success a {
  color: var(--accent);
  text-decoration: none;
  font-size: 14px;
}

.links a:hover,
.success a:hover {
  text-decoration: underline;
}
</style>
