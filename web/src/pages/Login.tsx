import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function Login() {
  const { login, loading } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await login(email, password)
      navigate('/')
    } catch {
      setError('Неверный email или пароль')
    }
  }

  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <span className="text-rpg-gold font-bold tracking-widest text-2xl">⚔ HILLSCHECK</span>
          <p className="text-gray-500 text-sm mt-2">Войди в свой аккаунт</p>
        </div>

        <form onSubmit={handleSubmit} className="card space-y-4">
          {error && (
            <div className="text-red-400 text-sm text-center bg-red-900/20 rounded-lg py-2 px-3">
              {error}
            </div>
          )}

          <div>
            <label className="block text-xs text-gray-400 mb-1 uppercase tracking-wider">Email</label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-rpg-gold"
              placeholder="grisha@example.com"
            />
          </div>

          <div>
            <label className="block text-xs text-gray-400 mb-1 uppercase tracking-wider">Пароль</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              minLength={8}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-rpg-gold"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-rpg-gold text-gray-950 font-bold py-2 rounded-lg text-sm hover:brightness-110 transition disabled:opacity-50"
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>

          <p className="text-center text-xs text-gray-500">
            Нет аккаунта?{' '}
            <Link to="/register" className="text-rpg-gold hover:underline">
              Зарегистрироваться
            </Link>
          </p>
        </form>
      </div>
    </div>
  )
}
