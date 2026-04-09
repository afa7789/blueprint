import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../services/api'

interface User {
  id: string
  email: string
  name: string
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isOperator = computed(() => user.value?.role === 'operator' || isAdmin.value)

  async function login(email: string, password: string) {
    const data = await api.post<{ user: User }>('/api/v1/auth/login', { email, password })
    user.value = data.user
  }

  async function register(email: string, password: string, name: string) {
    const data = await api.post<{ user: User }>('/api/v1/auth/register', { email, password, name })
    user.value = data.user
  }

  async function logout() {
    await api.post('/api/v1/auth/logout').catch(() => {})
    user.value = null
  }

  async function fetchMe() {
    try {
      const data = await api.get<User>('/api/v1/auth/me')
      user.value = data
    } catch {
      user.value = null
    }
  }

  async function forgotPassword(email: string) {
    await api.post('/api/v1/auth/forgot-password', { email })
  }

  async function resetPassword(token: string, password: string) {
    await api.post('/api/v1/auth/reset-password', { token, password })
  }

  return { user, isAuthenticated, isAdmin, isOperator, login, register, logout, fetchMe, forgotPassword, resetPassword }
})
