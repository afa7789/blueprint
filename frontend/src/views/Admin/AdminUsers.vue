<template>
  <div>
    <HelperBox
      title="User Management"
      description="Manage user accounts and roles. Admins can upgrade users to operator or admin roles. Deleting a user removes all their data."
    />
    <h2>Users</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="loading">Loading...</div>
    <template v-else>
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Email</th>
            <th>Name</th>
            <th>Role</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td class="mono">{{ user.id }}</td>
            <td>{{ user.email }}</td>
            <td>{{ user.name }}</td>
            <td>
              <select :value="user.role" @change="changeRole(user.id, ($event.target as HTMLSelectElement).value)">
                <option value="user">user</option>
                <option value="operator">operator</option>
                <option value="admin">admin</option>
              </select>
            </td>
            <td>{{ formatDate(user.created_at) }}</td>
            <td>
              <button class="btn-danger" @click="deleteUser(user.id)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="pagination">
        <button :disabled="page <= 1" @click="page--; load()">Prev</button>
        <span>Page {{ page }}</span>
        <button :disabled="users.length < pageSize" @click="page++; load()">Next</button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface User {
  id: string
  email: string
  name: string
  role: string
  created_at: string
}

const users = ref<User[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = 20

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: User[]; total: number }>(`/api/v1/admin/users?page=${page.value}&limit=${pageSize}`)
    users.value = data.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load users'
  } finally {
    loading.value = false
  }
}

async function changeRole(id: string, role: string) {
  try {
    await api.put(`/api/v1/admin/users/${id}/role`, { role })
    const u = users.value.find(u => u.id === id)
    if (u) u.role = role
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update role'
  }
}

async function deleteUser(id: string) {
  if (!confirm('Delete this user?')) return
  try {
    await api.delete(`/api/v1/admin/users/${id}`)
    users.value = users.value.filter(u => u.id !== id)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete user'
  }
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString()
}

onMounted(load)
</script>

<style scoped>
.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.admin-table th,
.admin-table td {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}

.admin-table th {
  color: var(--text);
  font-weight: 500;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.mono {
  font-family: var(--mono);
  font-size: 12px;
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

select {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 13px;
}

.btn-danger {
  background: none;
  border: 1px solid #e53e3e;
  color: #e53e3e;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.btn-danger:hover {
  background: #e53e3e;
  color: #fff;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  font-size: 14px;
}

.pagination button {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: default;
}

.error {
  color: #e53e3e;
  margin-bottom: 12px;
  font-size: 14px;
}
</style>
