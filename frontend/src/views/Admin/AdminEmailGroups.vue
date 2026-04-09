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
              <span class="group-count">{{ subscriberCounts[group.id] ?? '—' }} subscribers</span>
            </div>
            <button class="btn-danger-sm" @click.stop="deleteGroup(group.id)">Delete</button>
          </div>
          <div v-if="groups.length === 0" class="empty-state">No groups yet</div>
        </div>
      </div>

      <!-- Subscribers panel -->
      <div v-if="selectedGroup" class="subscribers-panel">
        <h3>{{ selectedGroup.name }} — Subscribers</h3>

        <!-- Add subscriber form -->
        <form class="add-sub-form" @submit.prevent="addSubscriber">
          <input
            v-model="newSubEmail"
            type="email"
            placeholder="Add subscriber email"
            required
          />
          <button type="submit" class="btn-primary">Add</button>
        </form>

        <div v-if="loadingSubs">Loading...</div>
        <table v-else class="admin-table">
          <thead>
            <tr>
              <th>Email</th>
              <th>Status</th>
              <th>Subscribed</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sub in subscribers" :key="sub.id">
              <td>{{ sub.email }}</td>
              <td>
                <span :class="sub.is_active ? 'badge-active' : 'badge-inactive'">
                  {{ sub.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td>{{ formatDate(sub.subscribed_at) }}</td>
              <td>
                <button
                  v-if="sub.is_active"
                  class="btn-danger-sm"
                  @click="deactivateSub(sub.email)"
                >
                  Deactivate
                </button>
              </td>
            </tr>
            <tr v-if="subscribers.length === 0">
              <td colspan="4" class="empty-cell">No subscribers</td>
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
  description: string | null
}

interface Subscriber {
  id: string
  email: string
  group_id: string | null
  is_active: boolean
  subscribed_at: string
}

const groups = ref<Group[]>([])
const subscribers = ref<Subscriber[]>([])
const selectedGroup = ref<Group | null>(null)
const subscriberCounts = ref<Record<string, number>>({})
const loadingGroups = ref(false)
const loadingSubs = ref(false)
const error = ref('')

const newGroup = reactive({ name: '', description: '' })
const newSubEmail = ref('')

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
  newSubEmail.value = ''
  loadingSubs.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Subscriber[] }>(`/api/v1/admin/email-groups/${group.id}/subscribers`)
    subscribers.value = data.data
    subscriberCounts.value = { ...subscriberCounts.value, [group.id]: data.data.length }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load subscribers'
  } finally {
    loadingSubs.value = false
  }
}

async function createGroup() {
  error.value = ''
  try {
    const group = await api.post<Group>('/api/v1/admin/email-groups', newGroup)
    groups.value.push(group)
    newGroup.name = ''
    newGroup.description = ''
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to create group'
  }
}

async function deleteGroup(id: string) {
  if (!confirm('Delete this group and all its subscribers?')) return
  error.value = ''
  try {
    await api.delete(`/api/v1/admin/email-groups/${id}`)
    groups.value = groups.value.filter(g => g.id !== id)
    if (selectedGroup.value?.id === id) {
      selectedGroup.value = null
      subscribers.value = []
    }
    const counts = { ...subscriberCounts.value }
    delete counts[id]
    subscriberCounts.value = counts
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete group'
  }
}

async function addSubscriber() {
  if (!selectedGroup.value) return
  error.value = ''
  try {
    const sub = await api.post<Subscriber>('/api/v1/admin/email-subscriptions', {
      email: newSubEmail.value,
      group_id: selectedGroup.value.id,
      is_active: true,
    })
    subscribers.value.push(sub)
    subscriberCounts.value = {
      ...subscriberCounts.value,
      [selectedGroup.value.id]: (subscriberCounts.value[selectedGroup.value.id] ?? 0) + 1,
    }
    newSubEmail.value = ''
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to add subscriber'
  }
}

async function deactivateSub(email: string) {
  if (!confirm(`Deactivate ${email}?`)) return
  error.value = ''
  try {
    await api.put(`/api/v1/admin/email-subscriptions/${encodeURIComponent(email)}/deactivate`, {})
    const sub = subscribers.value.find(s => s.email === email)
    if (sub) sub.is_active = false
    if (selectedGroup.value) {
      const activeCount = subscribers.value.filter(s => s.is_active).length
      subscriberCounts.value = { ...subscriberCounts.value, [selectedGroup.value.id]: activeCount }
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to deactivate subscriber'
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

.group-count {
  font-size: 11px;
  color: var(--text);
  opacity: 0.7;
}

.empty-state {
  font-size: 13px;
  color: var(--text);
  font-style: italic;
  padding: 8px 12px;
}

.subscribers-panel {
  flex: 1;
}

.add-sub-form {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.add-sub-form input {
  flex: 1;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 7px 10px;
  border-radius: 4px;
  font-size: 14px;
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

.empty-cell {
  color: var(--text);
  font-style: italic;
}

.badge-active {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  background: #c6f6d5;
  color: #276749;
}

.badge-inactive {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  background: #fed7d7;
  color: #9b2c2c;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  white-space: nowrap;
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
