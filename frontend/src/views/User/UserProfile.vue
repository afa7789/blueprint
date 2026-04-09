<template>
  <div class="user-profile">
    <h2>Profile</h2>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="fetchError" class="error">{{ fetchError }}</div>

    <form v-else @submit.prevent="save" class="profile-form">
      <div class="field">
        <label>Name</label>
        <input v-model="form.name" type="text" />
      </div>

      <div class="field">
        <label>Email
          <span v-if="profile?.email_verified" class="badge badge-success">Verified</span>
          <span v-else class="badge badge-warn">Unverified</span>
        </label>
        <input :value="profile?.email" type="email" readonly disabled />
      </div>

      <div class="field">
        <label>Phone</label>
        <input v-model="form.phone" type="tel" />
      </div>

      <div class="field">
        <label>Avatar URL</label>
        <input v-model="form.avatar_url" type="url" />
      </div>

      <fieldset class="address-group">
        <legend>Address</legend>
        <div class="field">
          <label>Street</label>
          <input v-model="form.address_street" type="text" />
        </div>
        <div class="field">
          <label>City</label>
          <input v-model="form.address_city" type="text" />
        </div>
        <div class="field">
          <label>State</label>
          <input v-model="form.address_state" type="text" />
        </div>
        <div class="field">
          <label>ZIP</label>
          <input v-model="form.address_zip" type="text" />
        </div>
      </fieldset>

      <div v-if="saveError" class="error">{{ saveError }}</div>
      <div v-if="success" class="success">Profile saved.</div>

      <button type="submit" class="btn btn-primary" :disabled="saving">
        {{ saving ? 'Saving...' : 'Save' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Profile {
  email: string
  email_verified: boolean
  name: string
  phone: string
  avatar_url: string
  address_street: string
  address_city: string
  address_state: string
  address_zip: string
}

const loading = ref(true)
const fetchError = ref('')
const saving = ref(false)
const saveError = ref('')
const success = ref(false)
const profile = ref<Profile | null>(null)

const form = ref({
  name: '',
  phone: '',
  avatar_url: '',
  address_street: '',
  address_city: '',
  address_state: '',
  address_zip: '',
})

onMounted(async () => {
  try {
    const data = await api.get<{ profile: Profile }>('/api/v1/user/profile')
    profile.value = data.profile
    const p = data.profile
    form.value = {
      name: p.name ?? '',
      phone: p.phone ?? '',
      avatar_url: p.avatar_url ?? '',
      address_street: p.address_street ?? '',
      address_city: p.address_city ?? '',
      address_state: p.address_state ?? '',
      address_zip: p.address_zip ?? '',
    }
  } catch (e: unknown) {
    fetchError.value = e instanceof Error ? e.message : 'Failed to load profile'
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  saveError.value = ''
  success.value = false
  try {
    await api.put('/api/v1/user/profile', form.value)
    success.value = true
  } catch (e: unknown) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.user-profile {
  max-width: 520px;
}

h2 {
  margin: 0 0 24px;
}

.profile-form {
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
  display: flex;
  align-items: center;
  gap: 8px;
}

.field input {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
}

.field input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.address-group {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.address-group legend {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text);
  padding: 0 4px;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 500;
}

.badge-success {
  background: #16a34a22;
  color: #16a34a;
}

.badge-warn {
  background: #d9770622;
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
