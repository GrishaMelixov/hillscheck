import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  setAccessToken,
  getAccessToken,
  get,
  post,
  ApiError,
} from './client'

// ── fetch mock helpers ────────────────────────────────────────────────────────

function mockFetch(status: number, body: unknown) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
    text: () => Promise.resolve(typeof body === 'string' ? body : JSON.stringify(body)),
  })
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('api/client — token management', () => {
  beforeEach(() => setAccessToken(null))

  it('stores and retrieves the access token', () => {
    setAccessToken('tok-123')
    expect(getAccessToken()).toBe('tok-123')
  })

  it('returns null when no token is set', () => {
    expect(getAccessToken()).toBeNull()
  })

  it('clears the token when set to null', () => {
    setAccessToken('tok-abc')
    setAccessToken(null)
    expect(getAccessToken()).toBeNull()
  })
})

describe('api/client — get()', () => {
  beforeEach(() => setAccessToken(null))
  afterEach(() => vi.restoreAllMocks())

  it('sends Authorization header when token is set', async () => {
    setAccessToken('my-token')
    const fetchMock = mockFetch(200, { ok: true })
    vi.stubGlobal('fetch', fetchMock)

    await get('/api/v1/profile')

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect((init.headers as Record<string, string>)['Authorization']).toBe('Bearer my-token')
  })

  it('does not send Authorization header when no token', async () => {
    const fetchMock = mockFetch(200, { ok: true })
    vi.stubGlobal('fetch', fetchMock)

    await get('/api/v1/profile')

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect((init.headers as Record<string, string>)['Authorization']).toBeUndefined()
  })

  it('throws ApiError on non-ok response', async () => {
    vi.stubGlobal('fetch', mockFetch(404, 'not found'))

    await expect(get('/api/v1/missing')).rejects.toBeInstanceOf(ApiError)
  })

  it('ApiError carries the correct status code', async () => {
    vi.stubGlobal('fetch', mockFetch(500, 'internal error'))

    try {
      await get('/api/v1/fail')
    } catch (e) {
      expect((e as ApiError).status).toBe(500)
    }
  })
})

describe('api/client — post()', () => {
  afterEach(() => vi.restoreAllMocks())

  it('serialises body as JSON', async () => {
    const fetchMock = mockFetch(200, {})
    vi.stubGlobal('fetch', fetchMock)

    await post('/auth/login', { email: 'a@b.com', password: 'pass1234' })

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(init.method).toBe('POST')
    expect(JSON.parse(init.body as string)).toEqual({ email: 'a@b.com', password: 'pass1234' })
  })

  it('returns undefined on 204 No Content', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
      json: () => Promise.reject(new Error('no body')),
      text: () => Promise.resolve(''),
    }))

    const result = await post('/auth/logout', {})
    expect(result).toBeUndefined()
  })
})
