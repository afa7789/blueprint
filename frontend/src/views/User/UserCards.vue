<template>
  <div class="user-cards">
    <h2>Saved Cards</h2>

    <div v-if="envNotice" class="notice">
      Card management requires <code>STRIPE_KEY</code> to be configured.
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="fetchError" class="error">{{ fetchError }}</div>

    <ul v-else-if="cards.length" class="cards-list">
      <li v-for="card in cards" :key="card.id" class="card-item">
        <span class="card-brand">{{ card.brand }}</span>
        <span class="card-number">**** {{ card.last4 }}</span>
        <span class="card-exp">{{ card.exp_month }}/{{ card.exp_year }}</span>
        <button class="btn btn-danger" @click="deleteCard(card.id)" :disabled="deletingId === card.id">
          {{ deletingId === card.id ? 'Removing...' : 'Remove' }}
        </button>
      </li>
    </ul>
    <p v-else class="empty">No saved cards.</p>

    <div v-if="addMessage" class="info">{{ addMessage }}</div>
    <div v-if="deleteError" class="error">{{ deleteError }}</div>

    <button class="btn btn-primary" @click="addCard">+ Add Card</button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, ApiError } from '../../services/api'

interface Card {
  id: string
  brand: string
  last4: string
  exp_month: number
  exp_year: number
}

const loading = ref(true)
const fetchError = ref('')
const envNotice = ref(false)
const cards = ref<Card[]>([])
const deletingId = ref<string | null>(null)
const deleteError = ref('')
const addMessage = ref('')

onMounted(async () => {
  try {
    const data = await api.get<{ cards: Card[] }>('/api/v1/user/saved-cards')
    cards.value = data.cards ?? []
  } catch (e: unknown) {
    if (e instanceof ApiError && (e.message?.includes('env_required') || e.status === 503)) {
      envNotice.value = true
    } else {
      fetchError.value = e instanceof Error ? e.message : 'Failed to load cards'
    }
  } finally {
    loading.value = false
  }
})

async function addCard() {
  try {
    await api.post('/api/v1/user/saved-cards', {})
    addMessage.value = 'Stripe integration - coming soon'
  } catch (e: unknown) {
    if (e instanceof ApiError && (e.message?.includes('env_required') || e.status === 503)) {
      envNotice.value = true
    } else {
      addMessage.value = 'Stripe integration - coming soon'
    }
  }
}

async function deleteCard(id: string) {
  deletingId.value = id
  deleteError.value = ''
  try {
    await api.delete(`/api/v1/user/saved-cards/${id}`)
    cards.value = cards.value.filter(c => c.id !== id)
  } catch (e: unknown) {
    deleteError.value = e instanceof Error ? e.message : 'Failed to remove card'
  } finally {
    deletingId.value = null
  }
}
</script>

<style scoped>
.user-cards {
  max-width: 520px;
}

h2 {
  margin: 0 0 24px;
}

.cards-list {
  list-style: none;
  padding: 0;
  margin: 0 0 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
}

.card-brand {
  font-weight: 600;
  font-size: 13px;
  text-transform: capitalize;
  min-width: 60px;
}

.card-number {
  flex: 1;
  font-family: monospace;
  font-size: 14px;
}

.card-exp {
  font-size: 13px;
  color: var(--text);
}

.empty {
  color: var(--text);
  font-size: 14px;
  margin: 0 0 16px;
}

.notice {
  background: #fef3c722;
  border-left: 3px solid #d97706;
  padding: 10px 14px;
  border-radius: 0 6px 6px 0;
  font-size: 14px;
  margin-bottom: 16px;
}

.info {
  color: var(--text);
  font-size: 14px;
  margin-bottom: 8px;
}

.error {
  color: var(--error, #dc2626);
  font-size: 14px;
  margin-bottom: 8px;
}

.btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
}

.btn-danger {
  background: #dc262622;
  color: #dc2626;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
