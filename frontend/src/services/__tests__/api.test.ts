import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock fetch
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

// Need to import after mocking
import { api, ApiError } from '../api'

describe('api service', () => {
  beforeEach(() => {
    mockFetch.mockReset()
  })

  it('GET returns parsed JSON', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: 'test' }),
    })
    const result = await api.get('/test')
    expect(result).toEqual({ data: 'test' })
    expect(mockFetch).toHaveBeenCalledWith(expect.stringContaining('/test'), expect.any(Object))
  })

  it('POST sends JSON body', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ ok: true }),
    })
    await api.post('/test', { email: 'a@b.com' })
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'a@b.com' }),
      })
    )
  })

  it('throws ApiError on non-ok response', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: () => Promise.resolve({ error: 'invalid token' }),
    })
    await expect(api.get('/test')).rejects.toThrow(ApiError)
    await expect(api.get('/test')).rejects.toMatchObject({ status: 401 })
  })

  it('DELETE sends correct method', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ deleted: true }),
    })
    await api.delete('/test/123')
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ method: 'DELETE' })
    )
  })
})
