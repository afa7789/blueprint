<template>
  <div class="user-password">
    <h2>Security</h2>

    <form @submit.prevent="submit" class="password-form">
      <div class="field">
        <label>Current Password</label>
        <input v-model="form.current_password" type="password" required />
      </div>

      <div class="field">
        <label>New Password</label>
        <input v-model="form.new_password" type="password" required />
        <span v-if="form.new_password && form.new_password.length < 6" class="hint">
          Minimum 6 characters
        </span>
      </div>

      <div class="field">
        <label>Confirm New Password</label>
        <input v-model="form.confirm_password" type="password" required />
        <span v-if="form.confirm_password && form.new_password !== form.confirm_password" class="hint">
          Passwords do not match
        </span>
      </div>

      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="success" class="success">Password updated successfully.</div>

      <button type="submit" class="btn btn-primary" :disabled="!canSubmit || saving">
        {{ saving ? 'Saving...' : 'Update Password' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '../../services/api'

const form = ref({
  current_password: '',
  new_password: '',
  confirm_password: '',
})

const saving = ref(false)
const error = ref('')
const success = ref(false)

const canSubmit = computed(() =>
  form.value.current_password.length > 0 &&
  form.value.new_password.length >= 6 &&
  form.value.new_password === form.value.confirm_password
)

async function submit() {
  if (!canSubmit.value) return
  saving.value = true
  error.value = ''
  success.value = false
  try {
    await api.put('/api/v1/user/password', {
      current_password: form.value.current_password,
      new_password: form.value.new_password,
    })
    success.value = true
    form.value = { current_password: '', new_password: '', confirm_password: '' }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update password'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.user-password {
  max-width: 400px;
}

h2 {
  margin: 0 0 24px;
}

.password-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-h);
}

.field input {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
}

.hint {
  font-size: 12px;
  color: #d97706;
}

.error {
  color: var(--error, #dc2626);
  font-size: 14px;
}

.success {
  color: #16a34a;
  font-size: 14px;
}

.btn {
  padding: 8px 20px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
