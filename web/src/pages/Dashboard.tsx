import { useState, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Avatar } from '../components/Avatar'
import { TransactionFeed } from '../components/TransactionFeed'
import { QuestList } from '../components/QuestList'
import ImportModal from '../components/ImportModal'
import { useWebSocket } from '../hooks/useWebSocket'
import { useAuth } from '../context/AuthContext'
import { fetchProfile, fetchQuests, fetchAccounts, fetchTransactions, getAccessToken } from '../api/client'
import type { Profile, Quest, Transaction } from '../api/client'

const SKILLS = [
  { name: 'Еда',         icon: '🍕', color: '#ef4444', key: 'food'          },
  { name: 'Обучение',    icon: '📖', color: '#34d399', key: 'learning'      },
  { name: 'Здоровье',   icon: '💊', color: '#3b82f6', key: 'health'        },
  { name: 'Спорт',      icon: '🏋️', color: '#a78bfa', key: 'sports'        },
  { name: 'Развлечения', icon: '🎮', color: '#f5c518', key: 'entertainment' },
  { name: 'Покупки',    icon: '🛍️', color: '#fb923c', key: 'shopping'      },
]

export default function Dashboard() {
  const { logout, user } = useAuth()
  const navigate = useNavigate()

  const [profile, setProfile]     = useState<Profile | null>(null)
  const [quests, setQuests]       = useState<Quest[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [defaultAccountId, setDefaultAccountId] = useState<string | null>(null)
  const [loading, setLoading]     = useState(true)
  const [error, setError]         = useState('')
  const [showImport, setShowImport] = useState(false)

  const loadTransactions = useCallback(async (accountId: string) => {
    try {
      const t = await fetchTransactions(accountId)
      setTransactions(t.transactions ?? [])
    } catch { /* non-fatal */ }
  }, [])

  const loadAll = useCallback(async () => {
    try {
      const [p, q, accs] = await Promise.all([fetchProfile(), fetchQuests(), fetchAccounts()])
      setProfile(p)
      setQuests(q.quests ?? [])
      const first = accs.accounts?.[0]
      if (first) {
        setDefaultAccountId(first.id)
        await loadTransactions(first.id)
      }
    } catch {
      setError('Не удалось загрузить данные')
    } finally {
      setLoading(false)
    }
  }, [loadTransactions])

  useEffect(() => { loadAll() }, [loadAll])

  const handleWsMessage = useCallback((type: string, payload: unknown) => {
    if (type === 'profile_update') setProfile(payload as Profile)
  }, [])
  useWebSocket(getAccessToken(), handleWsMessage)

  const handleLogout = async () => { await logout(); navigate('/login') }
  const handleImported = () => { if (defaultAccountId) loadTransactions(defaultAccountId) }

  const skillCounts = SKILLS.map(s => ({
    ...s,
    count: transactions.filter(tx => tx.clean_category?.toLowerCase().includes(s.key)).length,
  }))
  const maxCount = Math.max(...skillCounts.map(s => s.count), 1)

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <div className="text-center space-y-3">
          <div className="text-4xl animate-pulse">⚔️</div>
          <p className="text-gray-500 text-sm tracking-widest uppercase">Загружаем персонажа…</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <div className="text-center space-y-2">
          <p className="text-red-400 text-sm">{error}</p>
          <button onClick={loadAll} className="text-xs text-gray-500 hover:text-gray-300 underline">
            Попробовать снова
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      {/* ── Header ── */}
      <header className="sticky top-0 z-30 border-b border-gray-800/60 bg-gray-950/80 backdrop-blur-md px-4 sm:px-6 py-3">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-rpg-gold font-bold tracking-widest text-base sm:text-lg">⚔ HILLSCHECK</span>
            {profile && (
              <span className="hidden sm:inline-block ml-2 text-[10px] text-gray-600 border border-gray-800 rounded-full px-2 py-0.5 uppercase tracking-wider">
                Lv.{profile.level} Adventurer
              </span>
            )}
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            <button
              onClick={() => setShowImport(true)}
              className="flex items-center gap-1.5 text-xs bg-rpg-gold text-gray-950 font-bold px-3 py-1.5 rounded-lg hover:brightness-110 active:scale-95 transition-all"
            >
              <span>+</span>
              <span className="hidden sm:inline">Импорт CSV</span>
              <span className="sm:hidden">Импорт</span>
            </button>
            <button
              onClick={handleLogout}
              className="text-xs text-gray-500 hover:text-gray-300 transition px-2 py-1.5 rounded-lg hover:bg-gray-800/50"
            >
              <span className="hidden sm:inline">Выйти </span>
              {user?.name ? `(${user.name})` : '↩'}
            </button>
          </div>
        </div>
      </header>

      {/* ── Main ── */}
      <main className="max-w-5xl mx-auto px-4 py-5 grid grid-cols-1 md:grid-cols-3 gap-4">

        {/* Left column */}
        <div className="md:col-span-1 space-y-4">
          {profile && <Avatar profile={profile} />}
          <QuestList quests={quests} />
        </div>

        {/* Right column */}
        <div className="md:col-span-2 space-y-4">

          {/* Skill Tree */}
          <div className="card">
            <h2 className="section-label mb-4">🌳 Skill Tree</h2>
            <div className="grid grid-cols-3 sm:grid-cols-6 gap-2">
              {skillCounts.map((skill) => {
                const pct = Math.round((skill.count / maxCount) * 100)
                return (
                  <div key={skill.name}
                    className="flex flex-col items-center gap-1.5 bg-gray-800/50 rounded-xl p-2.5 border border-gray-700/30 hover:border-gray-600/50 transition-colors"
                  >
                    <div className="text-2xl">{skill.icon}</div>
                    <div className="text-[10px] text-gray-500 text-center leading-tight">{skill.name}</div>
                    {/* vertical bar */}
                    <div className="w-full h-12 rounded-lg bg-gray-900/60 flex items-end overflow-hidden p-0.5">
                      <div
                        className="w-full rounded-md transition-all duration-700"
                        style={{
                          height: `${Math.max(pct, 4)}%`,
                          backgroundColor: skill.color,
                          boxShadow: skill.count > 0 ? `0 0 8px ${skill.color}50` : 'none',
                        }}
                      />
                    </div>
                    <div className="text-xs font-bold" style={{ color: skill.color }}>
                      {skill.count}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Transaction feed */}
          <TransactionFeed transactions={transactions} />

          {/* Empty state CTA */}
          {transactions.length === 0 && (
            <div className="card border-dashed border-gray-700/50 text-center py-10 space-y-3">
              <div className="text-5xl">📂</div>
              <p className="text-gray-400 text-sm font-medium">Импортируй первые транзакции</p>
              <p className="text-gray-600 text-xs">Поддерживается Тинькофф, Сбер и любой CSV</p>
              <button
                onClick={() => setShowImport(true)}
                className="mt-2 inline-flex items-center gap-2 bg-rpg-gold text-gray-950 font-bold text-sm px-5 py-2 rounded-lg hover:brightness-110 transition"
              >
                + Загрузить CSV
              </button>
            </div>
          )}
        </div>
      </main>

      {showImport && (
        <ImportModal onClose={() => setShowImport(false)} onImported={handleImported} />
      )}
    </div>
  )
}
