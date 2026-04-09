<template>
  <div>
    <HelperBox title="Email System" description="Sending emails requires SMTP configuration." envVar="SMTP_HOST, SMTP_PORT" featureFlag="email_auto_enabled" />
    <h2>Email Groups</h2>
    <div v-if="error" class="error">{{ error }}</div>

    <div class="layout">
      <!-- Groups list -->
      <div class="groups-panel">
        <h3>Groups</h3>
        <form class="create-form" @submit.prevent="createGroup">
          <input v-model="newGroup.name" type="text" placeholder="Group name" required />
          <input v-model="newGroup.description" type="text" placeholder="Description (optional)" />
          <button type="submit" class="btn-primary">Create</button>
        </form>

        <div v-if="loadingGroups">Loading...</div>
        <div v-else class="group-list">
          <div
            v-for="group in groups"
            :key="group.id"
            class="group-row"
            :class="{ active: selectedGroup?.id === group.id }"
            @click="selectGroup(group)"
          >
            <div class="group-info">
              <span class="group-name">{{ group.name }}</span>
              <span v-if="group.description" class="group-desc">{{ group.description }}</span>
            </div>
            <button class="btn-danger-sm" @click.stop="deleteGroup(group.id)">Delete</button>
          </div>
        </div>
      </div>

      <!-- Subscribers panel -->
      <div v-if="selectedGroup" class="subscribers-panel">
        <h3>{{ selectedGroup.name }} — Subscribers</h3>
        <div v-if="loadingSubs">Loading...</div>
        <table v-else class="admin-table">
          <thead>
            <tr>
              <th>Email</th>
              <th>Name</th>
              <th>Subscribed</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sub in subscribers" :key="sub.email">
              <td>{{ sub.email }}</td>
              <td>{{ sub.name || '—' }}</td>
              <td>{{ formatDate(sub.subscribed_at) }}</td>
            </tr>
            <tr v-if="subscribers.length === 0">
              <td colspan="3" style="color: var(--text); font-style: italic;">No subscribers</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface Group {
  id: string
  name: string
  description: string
}

interface Subscriber {
  email: string
  name: string
  subscribed_at: string
}

const groups = ref<Group[]>([])
const subscribers = ref<Subscriber[]>([])
const selectedGroup = ref<Group | null>(null)
const loadingGroups = ref(false)
const loadingSubs = ref(false)
const error = ref('')

const newGroup = reactive({ name: '', description: '' })

async function loadGroups() {
  loadingGroups.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Group[] }>('/api/v1/admin/email-groups')
    groups.value = data.data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load groups'
  } finally {
    loadingGroups.value = false
  }
}

async function selectGroup(group: Group) {
  selectedGroup.value = group
  loadingSubs.value = true
  try {
    const data = await api.get<{ data: Subscriber[] }>(`/api/v1/admin/email-groups/${group.id}/subscribers`)
    subscribers.value = data.data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load subscribers'
  } finally {
    loadingSubs.value = false
  }
}

async function createGroup() {
  error.value = ''
  try {
    const data = await api.post<{ group: Group }>('/api/v1/admin/email-groups', newGroup)
    groups.value.push(data.group)
    newGroup.name = ''
    newGroup.description = ''
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to create group'
  }
}

async function deleteGroup(id: string) {
  if (!confirm('Delete this group and all its subscribers?')) return
  try {
    await api.delete(`/api/v1/admin/email-groups/${id}`)
    groups.value = groups.value.filter(g => g.id !== id)
    if (selectedGroup.value?.id === id) {
      selectedGroup.value = null
      subscribers.value = []
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete group'
  }
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString()
}

onMounted(loadGroups)
</script>

<style scoped>
.layout {
  display: flex;
  gap: 24px;
  margin-top: 12px;
  align-items: flex-start;
}

.groups-panel {
  width: 280px;
  flex-shrink: 0;
}

.groups-panel h3,
.subscribers-panel h3 {
  font-size: 14px;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 12px;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.create-form input {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 7px 10px;
  border-radius: 4px;
  font-size: 14px;
}

.group-list {
  display: flex;
  flex-direction: column;
}

.group-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: background 0.15s;
}

.group-row:hover {
  background: var(--code-bg);
}

.group-row.active {
  background: var(--accent-bg);
  border-color: var(--accent-border);
}

.group-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.group-name {
  font-size: 14px;
  color: var(--text-h);
  font-weight: 500;
}

.group-desc {
  font-size: 12px;
  color: var(--text);
}

.subscribers-panel {
  flex: 1;
}

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

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.btn-danger-sm {
  background: none;
  border: 1px solid #e53e3e;
  color: #e53e3e;
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  flex-shrink: 0;
}

.error {
  color: #e53e3e;
  margin-bottom: 12px;
  font-size: 14px;
}

@media (max-width: 768px) {
  .layout {
    flex-direction: column;
  }
  .groups-panel {
    width: 100%;
  }
}
</style>
