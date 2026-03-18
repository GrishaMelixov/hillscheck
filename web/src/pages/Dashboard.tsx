import { useState, useCallback } from 'react'
import { Avatar } from '../components/Avatar'
import { TransactionFeed } from '../components/TransactionFeed'
import { QuestList } from '../components/QuestList'
import { useWebSocket } from '../hooks/useWebSocket'

const DEMO_PROFILE = {
  level: 3,
  xp: 540,
  hp: 72,
  mana: 45,
  strength: 14,
  intellect: 22,
  luck: 11,
}

const DEMO_QUESTS = [
  {
    id: 'quest-read-books',
    title: 'Scholar',
    description: 'Buy 3 books or courses this month',
    attribute: 'intellect',
    reward: 50,
    progress: 66,
  },
  {
    id: 'quest-gym',
    title: 'Warrior',
    description: 'Visit the gym 3 times',
    attribute: 'strength',
    reward: 40,
    progress: 33,
  },
]

export default function Dashboard() {
  const [profile, setProfile] = useState(DEMO_PROFILE)
  const [transactions] = useState<never[]>([])
  const [quests] = useState(DEMO_QUESTS)

  // Replace 'demo-token' with real auth token once auth is implemented.
  const token = 'demo-token'

  const handleWsMessage = useCallback((type: string, payload: unknown) => {
    if (type === 'profile_update') {
      setProfile(payload as typeof DEMO_PROFILE)
    }
  }, [])

  useWebSocket(token, handleWsMessage)

  return (
    <div className="min-h-screen bg-gray-950">
      {/* Header */}
      <header className="border-b border-gray-800 px-6 py-3 flex items-center justify-between">
        <span className="text-rpg-gold font-bold tracking-widest text-lg">⚔ HILLSCHECK</span>
        <span className="text-xs text-gray-500">Level {profile.level} Adventurer</span>
      </header>

      {/* Main grid */}
      <main className="max-w-5xl mx-auto px-4 py-6 grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Left column */}
        <div className="md:col-span-1 space-y-4">
          <Avatar profile={profile} />
          <QuestList quests={quests} />
        </div>

        {/* Right column */}
        <div className="md:col-span-2 space-y-4">
          <div className="card">
            <h2 className="text-sm text-gray-400 uppercase tracking-wider mb-4">Skill Tree</h2>
            <div className="grid grid-cols-3 gap-3 text-center">
              {[
                { name: 'Food', icon: '🍕', value: 12, color: '#dc2626' },
                { name: 'Learning', icon: '📖', value: 28, color: '#16a34a' },
                { name: 'Health', icon: '💊', value: 8, color: '#2563eb' },
                { name: 'Sports', icon: '🏋️', value: 15, color: '#7c3aed' },
                { name: 'Entertainment', icon: '🎮', value: 6, color: '#f5c518' },
                { name: 'Shopping', icon: '🛍', value: 22, color: '#ea580c' },
              ].map((skill) => (
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
        </div>
      </main>
    </div>
  )
}
