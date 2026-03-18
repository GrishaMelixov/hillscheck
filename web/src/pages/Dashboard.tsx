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
  { name: 'Еда',          icon: '🍕', color: '#FF453A', key: 'food'          },
  { name: 'Обучение',     icon: '📖', color: '#30D158', key: 'learning'      },
  { name: 'Здоровье',     icon: '💊', color: '#0A84FF', key: 'health'        },
  { name: 'Спорт',        icon: '🏋️', color: '#BF5AF2', key: 'sports'        },
  { name: 'Развлечения',  icon: '🎮', color: '#FFD60A', key: 'entertainment' },
  { name: 'Покупки',      icon: '🛍️', color: '#FF9F0A', key: 'shopping'      },
]

export default function Dashboard() {
  const { logout, user } = useAuth()
  const navigate = useNavigate()

  const [profile, setProfile]           = useState<Profile | null>(null)
  const [quests, setQuests]             = useState<Quest[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [defaultAccountId, setDefaultAccountId] = useState<string | null>(null)
  const [loading, setLoading]           = useState(true)
  const [error, setError]               = useState('')
  const [showImport, setShowImport]     = useState(false)

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

  const handleLogout   = async () => { await logout(); navigate('/login') }
  const handleImported = () => { if (defaultAccountId) loadTransactions(defaultAccountId) }

  const skillCounts = SKILLS.map(s => ({
    ...s,
    count: transactions.filter(tx => tx.clean_category?.toLowerCase().includes(s.key)).length,
  }))
  const maxCount = Math.max(...skillCounts.map(s => s.count), 1)

  /* ── Loading ── */
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="text-5xl float">⚔️</div>
          <p className="text-[13px] uppercase tracking-widest" style={{ color: 'rgba(255,255,255,0.3)' }}>
            Загружаем персонажа…
          </p>
        </div>
      </div>
    )
  }

  /* ── Error ── */
  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center space-y-3">
          <p className="text-[14px]" style={{ color: '#FF453A' }}>{error}</p>
          <button
            onClick={loadAll}
            className="text-[13px] underline transition-opacity hover:opacity-60"
            style={{ color: 'rgba(255,255,255,0.4)' }}
          >
            Попробовать снова
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen">

      {/* ── Header ── */}
      <header
        className="sticky top-0 z-30 px-4 sm:px-6 py-4"
        style={{
          background: 'rgba(0,0,0,0.6)',
          backdropFilter: 'blur(40px) saturate(1.8)',
          WebkitBackdropFilter: 'blur(40px) saturate(1.8)',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
        }}
      >
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span
              className="text-[18px] font-bold tracking-tight"
              style={{ color: '#F5C518', letterSpacing: '-0.01em' }}
            >
              ⚔ HILLSCHECK
            </span>
            {profile && (
              <span
                className="hidden sm:inline-block text-[10px] font-semibold uppercase tracking-widest px-2.5 py-1 rounded-full"
                style={{
                  color: 'rgba(255,255,255,0.35)',
                  background: 'rgba(255,255,255,0.06)',
                  border: '1px solid rgba(255,255,255,0.08)',
                }}
              >
                Lv.{profile.level}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowImport(true)}
              className="btn-primary text-[13px] px-4 py-2"
            >
              <span className="mr-1.5">+</span>
              <span className="hidden sm:inline">Импорт CSV</span>
              <span className="sm:hidden">Импорт</span>
            </button>
            <button
              onClick={handleLogout}
              className="btn-glass text-[13px] px-4 py-2"
            >
              <span className="hidden sm:inline">Выйти </span>
              {user?.name ? `(${user.name})` : '↩'}
            </button>
          </div>
        </div>
      </header>

      {/* ── Main ── */}
      <main className="max-w-5xl mx-auto px-4 py-6 grid grid-cols-1 md:grid-cols-3 gap-4">

        {/* Left column */}
        <div className="md:col-span-1 space-y-4">
          {profile && <Avatar profile={profile} />}
          <QuestList quests={quests} />
        </div>

        {/* Right column */}
        <div className="md:col-span-2 space-y-4">

          {/* Skill Tree */}
          <div className="glass p-5">
            <h2 className="section-label mb-4">Skill Tree</h2>
            <div className="grid grid-cols-3 sm:grid-cols-6 gap-2">
              {skillCounts.map((skill) => {
                const pct = Math.round((skill.count / maxCount) * 100)
                return (
                  <div
                    key={skill.name}
                    className="flex flex-col items-center gap-2 rounded-2xl p-3 transition-all"
                    style={{
                      background: 'rgba(255,255,255,0.03)',
                      border: '1px solid rgba(255,255,255,0.07)',
                    }}
                  >
                    <div className="text-2xl">{skill.icon}</div>
                    <div className="text-[10px] text-center leading-tight" style={{ color: 'rgba(255,255,255,0.3)' }}>
                      {skill.name}
                    </div>
                    {/* Vertical bar */}
                    <div
                      className="w-full rounded-xl overflow-hidden flex items-end"
                      style={{ height: '48px', background: 'rgba(255,255,255,0.05)', padding: '2px' }}
                    >
                      <div
                        className="w-full rounded-lg transition-all duration-700"
                        style={{
                          height: `${Math.max(pct, 4)}%`,
                          backgroundColor: skill.color,
                          boxShadow: skill.count > 0 ? `0 0 10px ${skill.color}50` : 'none',
                        }}
                      />
                    </div>
                    <div className="text-[12px] font-bold" style={{ color: skill.color }}>
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
            <div
              className="rounded-3xl p-10 text-center space-y-4"
              style={{
                border: '1px dashed rgba(255,255,255,0.10)',
                background: 'rgba(255,255,255,0.02)',
              }}
            >
              <div className="text-5xl float">📂</div>
              <div>
                <p className="text-[15px] font-semibold text-white/70">Импортируй первые транзакции</p>
                <p className="text-[13px] mt-1" style={{ color: 'rgba(255,255,255,0.30)' }}>
                  Поддерживается Тинькофф, Сбер и любой CSV
                </p>
              </div>
              <button
                onClick={() => setShowImport(true)}
                className="btn-primary inline-flex"
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
