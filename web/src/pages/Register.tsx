import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function Register() {
  const { register, loading } = useAuth()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (password.length < 8) {
      setError('Пароль должен быть не менее 8 символов')
      return
    }
    try {
      await register(name, email, password)
      navigate('/')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : ''
      setError(msg.includes('409') || msg.toLowerCase().includes('already')
        ? 'Этот email уже зарегистрирован'
        : 'Ошибка регистрации. Попробуй ещё раз.')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-[380px]">

        {/* Logo */}
        <div className="text-center mb-10">
          <div
            className="inline-flex items-center justify-center w-16 h-16 rounded-[22px] text-3xl mb-5"
            style={{
              background: 'linear-gradient(135deg, rgba(245,197,24,0.2), rgba(10,132,255,0.15))',
              border: '1px solid rgba(245,197,24,0.25)',
              boxShadow: '0 0 40px rgba(245,197,24,0.12)',
            }}
          >
            ⚔️
          </div>
          <h1
            className="text-[28px] font-bold tracking-tight leading-none"
            style={{ color: '#F5C518', letterSpacing: '-0.02em' }}
          >
            HILLSCHECK
          </h1>
          <p className="text-[14px] mt-2" style={{ color: 'rgba(255,255,255,0.35)' }}>
            Создай персонажа
          </p>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className="space-y-4"
          style={{
            background: 'rgba(255,255,255,0.05)',
            border: '1px solid rgba(255,255,255,0.10)',
            borderRadius: '28px',
            padding: '28px',
            backdropFilter: 'blur(48px) saturate(1.8)',
            WebkitBackdropFilter: 'blur(48px) saturate(1.8)',
            boxShadow: 'inset 0 1px 0 rgba(245,197,24,0.10), 0 24px 64px rgba(0,0,0,0.6)',
          }}
        >
          {error && (
            <div
              className="text-[13px] text-center rounded-2xl py-2.5 px-4"
              style={{ color: '#FF453A', background: 'rgba(255,69,58,0.12)', border: '1px solid rgba(255,69,58,0.2)' }}
            >
              {error}
            </div>
          )}

          <div className="space-y-1.5">
            <label className="block text-[11px] font-semibold uppercase tracking-widest" style={{ color: 'rgba(255,255,255,0.35)' }}>
              Имя
            </label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              required
              className="input"
              placeholder="Griша"
            />
          </div>

          <div className="space-y-1.5">
            <label className="block text-[11px] font-semibold uppercase tracking-widest" style={{ color: 'rgba(255,255,255,0.35)' }}>
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              className="input"
              placeholder="grisha@example.com"
            />
          </div>

          <div className="space-y-1.5">
            <label className="block text-[11px] font-semibold uppercase tracking-widest" style={{ color: 'rgba(255,255,255,0.35)' }}>
              Пароль
            </label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              minLength={8}
              className="input"
              placeholder="минимум 8 символов"
            />
          </div>

          <button type="submit" disabled={loading} className="btn-primary w-full mt-2">
            {loading ? 'Создаём…' : 'Создать аккаунт'}
          </button>

          <p className="text-center text-[13px]" style={{ color: 'rgba(255,255,255,0.30)' }}>
            Уже есть аккаунт?{' '}
            <Link to="/login" className="font-semibold transition-opacity hover:opacity-80" style={{ color: '#F5C518' }}>
              Войти
            </Link>
          </p>
        </form>
      </div>
    </div>
  )
}
