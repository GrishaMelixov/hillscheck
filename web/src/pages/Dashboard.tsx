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

const SKILL_CATEGORIES = [
  { name: 'Food',          icon: '🍕', color: '#dc2626', key: 'food' },
  { name: 'Learning',      icon: '📖', color: '#16a34a', key: 'learning' },
  { name: 'Health',        icon: '💊', color: '#2563eb', key: 'health' },
  { name: 'Sports',        icon: '🏋️', color: '#7c3aed', key: 'sports' },
  { name: 'Entertainment', icon: '🎮', color: '#f5c518', key: 'entertainment' },
  { name: 'Shopping',      icon: '🛍',  color: '#ea580c', key: 'shopping' },
]

export default function Dashboard() {
  const { logout, user } = useAuth()
  const navigate = useNavigate()

  const [profile, setProfile]           = useState<Profile | null>(null)
  const [quests, setQuests]             = useState<Quest[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [defaultAccountId, setDefaultAccountId] = useState<string | null>(null)
  const [loadingProfile, setLoadingProfile] = useState(true)
  const [error, setError]               = useState('')
  const [showImport, setShowImport]     = useState(false)

  const loadTransactions = useCallback(async (accountId: string) => {
    try {
      const t = await fetchTransactions(accountId)
      setTransactions(t.transactions ?? [])
    } catch {
      // Non-fatal — transactions may just be empty
    }
  }, [])

  const loadAll = useCallback(async () => {
    try {
      const [p, q, accounts] = await Promise.all([fetchProfile(), fetchQuests(), fetchAccounts()])
      setProfile(p)
      setQuests(q.quests ?? [])
      const firstAccount = accounts.accounts?.[0]
      if (firstAccount) {
        setDefaultAccountId(firstAccount.id)
        await loadTransactions(firstAccount.id)
      }
    } catch {
      setError('Не удалось загрузить данные')
    } finally {
      setLoadingProfile(false)
    }
  }, [loadTransactions])

  useEffect(() => { loadAll() }, [loadAll])

  // Real-time profile updates via WebSocket
  const handleWsMessage = useCallback((type: string, payload: unknown) => {
    if (type === 'profile_update') setProfile(payload as Profile)
  }, [])
  useWebSocket(getAccessToken(), handleWsMessage)

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const handleImported = () => {
    if (defaultAccountId) loadTransactions(defaultAccountId)
  }

  // Build skill-tree counts from real transactions
  const skillCounts = SKILL_CATEGORIES.map(skill => ({
    ...skill,
    value: transactions.filter(tx =>
      tx.clean_category?.toLowerCase().includes(skill.key)
    ).length,
  }))

  if (loadingProfile) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <p className="text-gray-500 text-sm animate-pulse">Загружаем персонажа...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <p className="text-red-400 text-sm">{error}</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800 px-6 py-3 flex items-center justify-between">
        <span className="text-rpg-gold font-bold tracking-widest text-lg">⚔ HILLSCHECK</span>
        <div className="flex items-center gap-4">
          {profile && (
            <span className="text-xs text-gray-500">Level {profile.level} Adventurer</span>
          )}
          <button
            onClick={() => setShowImport(true)}
            className="text-xs bg-rpg-gold text-gray-950 font-bold px-3 py-1.5 rounded-lg hover:brightness-110 transition"
          >
            + Импорт
          </button>
          <button
            onClick={handleLogout}
            className="text-xs text-gray-500 hover:text-gray-300 transition"
          >
            Выйти {user?.name ? `(${user.name})` : ''}
          </button>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6 grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="md:col-span-1 space-y-4">
          {profile && <Avatar profile={profile} />}
          <QuestList quests={quests} />
        </div>

        <div className="md:col-span-2 space-y-4">
          <div className="card">
            <h2 className="text-sm text-gray-400 uppercase tracking-wider mb-4">Skill Tree</h2>
            <div className="grid grid-cols-3 gap-3 text-center">
              {skillCounts.map((skill) => (
                <div key={skill.name} className="bg-gray-800 rounded-lg p-3">
                  <div className="text-2xl mb-1">{skill.icon}</div>
                  <div className="text-xs text-gray-400 mb-1">{skill.name}</div>
                  <div className="w-full h-1.5 rounded-full bg-gray-700">
                    <div
                      className="h-1.5 rounded-full"
                      style={{ width: `${Math.min(skill.value * 3, 100)}%`, backgroundColor: skill.color }}
                    />
                  </div>
                  <div className="text-xs font-bold mt-1">{skill.value} ops</div>
                </div>
              ))}
            </div>
          </div>

          <TransactionFeed transactions={transactions} />

          {transactions.length === 0 && !loadingProfile && (
            <div className="card text-center py-10">
              <p className="text-gray-500 text-sm mb-3">Транзакций пока нет</p>
              <button
                onClick={() => setShowImport(true)}
                className="text-xs bg-rpg-gold text-gray-950 font-bold px-4 py-2 rounded-lg hover:brightness-110 transition"
              >
                Импортировать CSV
              </button>
            </div>
          )}
        </div>
      </main>

      {showImport && (
        <ImportModal
          onClose={() => setShowImport(false)}
          onImported={handleImported}
        />
      )}
    </div>
  )
}
