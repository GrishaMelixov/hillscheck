// API client — stores access token in memory (never localStorage).
// Refresh token lives in httpOnly cookie, handled by the browser automatically.

let accessToken: string | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken(): string | null {
  return accessToken
}

// Attempt to restore a session using the refresh token cookie.
// Call once on app startup.
export async function restoreSession(): Promise<boolean> {
  try {
    const data = await post<{ access_token: string }>('/auth/refresh', {})
    setAccessToken(data.access_token)
    return true
  } catch {
    return false
  }
}

// ── Core fetch wrapper ────────────────────────────────────────────────────────

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string>),
  }
  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`
  }

  const res = await fetch(path, { ...init, headers })

  // Access token expired — try refresh once, then retry
  if (res.status === 401 && accessToken) {
    const refreshed = await restoreSession()
    if (refreshed) {
      headers['Authorization'] = `Bearer ${accessToken}`
      const retry = await fetch(path, { ...init, headers })
      if (!retry.ok) throw new ApiError(retry.status, await retry.text())
      return retry.json() as Promise<T>
    }
    // Refresh failed — clear token so app redirects to login
    setAccessToken(null)
    throw new ApiError(401, 'session expired')
  }

  if (!res.ok) {
    throw new ApiError(res.status, await res.text())
  }

  // 204 No Content
  if (res.status === 204) return undefined as unknown as T

  return res.json() as Promise<T>
}

export function get<T>(path: string): Promise<T> {
  return request<T>(path)
}

export function post<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

// ── Error type ────────────────────────────────────────────────────────────────

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
  }
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export interface LoginResponse {
  access_token: string
  user: { id: string; name: string; email: string; plan: string }
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const data = await post<LoginResponse>('/auth/login', { email, password })
  setAccessToken(data.access_token)
  return data
}

export async function register(name: string, email: string, password: string) {
  return post<{ id: string; name: string; email: string; plan: string }>(
    '/auth/register',
    { name, email, password },
  )
}

export async function logout() {
  await post('/auth/logout', {})
  setAccessToken(null)
}

// ── Game ──────────────────────────────────────────────────────────────────────

export interface Profile {
  user_id: string
  level: number
  xp: number
  hp: number
  mana: number
  strength: number
  intellect: number
  luck: number
}

export interface Quest {
  id: string
  title: string
  description: string
  attribute: string
  reward: number
  progress: number
}

export interface Transaction {
  id: string
  amount: number
  clean_category: string
  original_description: string
  status: string
  created_at: string
}

export const fetchProfile = () => get<Profile>('/api/v1/profile')
export const fetchQuests = () => get<{ quests: Quest[] }>('/api/v1/quests')
export const fetchTransactions = () => get<{ transactions: Transaction[] }>('/api/v1/transactions')
