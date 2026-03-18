import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { restoreSession, login as apiLogin, logout as apiLogout, register as apiRegister, getAccessToken } from '../api/client'

interface User {
  id: string
  name: string
  email: string
  plan: string
}

interface AuthState {
  user: User | null
  ready: boolean   // true once we've checked for a saved session
  loading: boolean
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ user: null, ready: false, loading: false })

  // On mount: try to restore session from refresh-token cookie
  useEffect(() => {
    restoreSession().then((ok) => {
      // If restored we don't have user info yet — the dashboard will fetch /profile
      // We just signal that the token is available
      setState({ user: ok ? ({ id: '', name: '', email: '', plan: '' }) : null, ready: true, loading: false })
    })
  }, [])

  const login = async (email: string, password: string) => {
    setState(s => ({ ...s, loading: true }))
    const data = await apiLogin(email, password)
    setState({ user: data.user, ready: true, loading: false })
  }

  const register = async (name: string, email: string, password: string) => {
    setState(s => ({ ...s, loading: true }))
    await apiRegister(name, email, password)
    // Auto-login after register
    const data = await apiLogin(email, password)
    setState({ user: data.user, ready: true, loading: false })
  }

  const logout = async () => {
    await apiLogout()
    setState({ user: null, ready: true, loading: false })
  }

  return (
    <AuthContext.Provider value={{ ...state, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}

export function isAuthenticated(): boolean {
  return getAccessToken() !== null
}
