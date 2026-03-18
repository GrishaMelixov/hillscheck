
interface Quest {
  id: string
  title: string
  description: string
  attribute: string
  reward: number
  progress: number
}

const ATTR_META: Record<string, { icon: string; color: string; bg: string }> = {
  xp:        { icon: '✨', color: '#f5c518', bg: 'rgba(245,197,24,0.08)'  },
  hp:        { icon: '❤️', color: '#ef4444', bg: 'rgba(239,68,68,0.08)'  },
  mana:      { icon: '💎', color: '#3b82f6', bg: 'rgba(59,130,246,0.08)'  },
  strength:  { icon: '⚔️', color: '#a78bfa', bg: 'rgba(167,139,250,0.08)' },
  intellect: { icon: '📖', color: '#34d399', bg: 'rgba(52,211,153,0.08)'  },
  luck:      { icon: '🍀', color: '#fbbf24', bg: 'rgba(251,191,36,0.08)'  },
}
const DEFAULT_META = { icon: '🎯', color: '#6b7280', bg: 'rgba(107,114,128,0.08)' }

interface Props {
  quests: Quest[]
}

export function QuestList({ quests }: Props) {
  return (
    <div className="card space-y-3">
      <h2 className="section-label">🗺 Active Quests</h2>

      {quests.length === 0 ? (
        <div className="rounded-xl border border-dashed border-gray-700/50 p-5 text-center">
          <p className="text-gray-600 text-sm">Квестов нет</p>
          <p className="text-gray-700 text-xs mt-1">Импортируй транзакции, чтобы получить цели</p>
        </div>
      ) : (
        <ul className="space-y-2">
          {quests.map((q) => {
            const meta = ATTR_META[q.attribute] ?? DEFAULT_META
            return (
              <li key={q.id}
                className="rounded-xl border border-gray-800/60 p-3 space-y-2.5 hover:border-gray-700/60 transition-colors"
                style={{ background: meta.bg }}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="text-lg shrink-0">{meta.icon}</span>
                    <span className="text-sm font-semibold text-gray-100 leading-tight">{q.title}</span>
                  </div>
                  <span
                    className="badge shrink-0 text-xs"
                    style={{ backgroundColor: meta.bg, color: meta.color, border: `1px solid ${meta.color}30` }}
                  >
                    +{q.reward}
                  </span>
                </div>

                <p className="text-xs text-gray-500 leading-relaxed">{q.description}</p>

                <div className="space-y-1">
                  <div className="relative h-1.5 w-full rounded-full bg-black/30 overflow-hidden">
                    <div
                      className="h-full rounded-full transition-all duration-700"
                      style={{
                        width: `${q.progress}%`,
                        backgroundColor: meta.color,
                        boxShadow: `0 0 8px ${meta.color}60`,
                      }}
                    />
                  </div>
                  <p className="text-right text-[10px] text-gray-600">{q.progress}%</p>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
