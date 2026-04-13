import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock fetch
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

// Need to import after mocking
import { api, ApiError } from '../api'

// Build a minimal Response-like stub. The real `api.request` reads
// `res.headers.get('content-length')` to decide whether to parse a body,
// so mocks must expose a `headers.get` function.
function mockResponse(body: {
  ok: boolean
  status?: number
  statusText?: string
  json?: unknown
  contentLength?: string | null
}) {
  return {
    ok: body.ok,
    status: body.status ?? (body.ok ? 200 : 500),
    statusText: body.statusText ?? '',
    headers: {
      get: (name: string) =>
        name.toLowerCase() === 'content-length' ? (body.contentLength ?? null) : null,
    },
    json: () => Promise.resolve(body.json ?? {}),
  }
}

describe('api service', () => {
  beforeEach(() => {
    mockFetch.mockReset()
  })

  it('GET returns parsed JSON', async () => {
    mockFetch.mockResolvedValue(mockResponse({ ok: true, json: { data: 'test' } }))
    const result = await api.get('/test')
    expect(result).toEqual({ data: 'test' })
    expect(mockFetch).toHaveBeenCalledWith(expect.stringContaining('/test'), expect.any(Object))
  })

  it('POST sends JSON body', async () => {
    mockFetch.mockResolvedValue(mockResponse({ ok: true, json: { ok: true } }))
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
    mockFetch.mockResolvedValue(
      mockResponse({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        json: { error: 'invalid token' },
      })
    )
    await expect(api.get('/test')).rejects.toThrow(ApiError)
    await expect(api.get('/test')).rejects.toMatchObject({ status: 401 })
  })

  it('DELETE sends correct method', async () => {
    mockFetch.mockResolvedValue(mockResponse({ ok: true, json: { deleted: true } }))
    await api.delete('/test/123')
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ method: 'DELETE' })
    )
  })
})
