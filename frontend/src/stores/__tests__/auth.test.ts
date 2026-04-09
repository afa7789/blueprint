import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock the api module
vi.mock('../../services/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
  ApiError: class extends Error {
    status: number
    constructor(status: number, message: string) {
      super(message)
      this.status = status
    }
  },
}))

import { useAuthStore } from '../auth'
import { api } from '../../services/api'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(api.get).mockReset()
    vi.mocked(api.post).mockReset()
  })

  it('starts unauthenticated', () => {
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
    expect(auth.isAdmin).toBe(false)
  })

  it('login sets user', async () => {
    vi.mocked(api.post).mockResolvedValue({ user: { id: '1', email: 'a@b.com', name: 'Test', role: 'user' } })
    const auth = useAuthStore()
    await auth.login('a@b.com', 'pass')
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.email).toBe('a@b.com')
  })

  it('logout clears user', async () => {
    vi.mocked(api.post).mockResolvedValue({ user: { id: '1', email: 'a@b.com', name: 'Test', role: 'user' } })
    const auth = useAuthStore()
    await auth.login('a@b.com', 'pass')
    await auth.logout()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('isAdmin true for admin role', async () => {
    vi.mocked(api.post).mockResolvedValue({ user: { id: '1', email: 'a@b.com', name: 'Admin', role: 'admin' } })
    const auth = useAuthStore()
    await auth.login('a@b.com', 'pass')
    expect(auth.isAdmin).toBe(true)
    expect(auth.isOperator).toBe(true) // admin is also operator
  })

  it('isOperator true for operator role', async () => {
    vi.mocked(api.post).mockResolvedValue({ user: { id: '1', email: 'a@b.com', name: 'Op', role: 'operator' } })
    const auth = useAuthStore()
    await auth.login('a@b.com', 'pass')
    expect(auth.isAdmin).toBe(false)
    expect(auth.isOperator).toBe(true)
  })

  it('fetchMe sets user from API', async () => {
    vi.mocked(api.get).mockResolvedValue({ user: { id: '1', email: 'a@b.com', name: 'Me', role: 'user' } })
    const auth = useAuthStore()
    await auth.fetchMe()
    expect(auth.isAuthenticated).toBe(true)
  })

  it('fetchMe clears user on error', async () => {
    vi.mocked(api.get).mockRejectedValue(new Error('401'))
    const auth = useAuthStore()
    await auth.fetchMe()
    expect(auth.isAuthenticated).toBe(false)
  })
})
