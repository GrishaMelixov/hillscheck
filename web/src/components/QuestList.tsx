
interface Quest {
  id: string
  title: string
  description: string
  attribute: string
  reward: number
  progress: number
}

const ATTR_COLORS: Record<string, string> = {
  xp:        '#f5c518',
  hp:        '#dc2626',
  mana:      '#2563eb',
  strength:  '#7c3aed',
  intellect: '#16a34a',
  luck:      '#ea580c',
}

interface Props {
  quests: Quest[]
}

export function QuestList({ quests }: Props) {
  return (
    <div className="card space-y-3">
      <h2 className="text-sm text-gray-400 uppercase tracking-wider">Active Quests</h2>
      {quests.length === 0 && (
        <p className="text-gray-500 text-sm">No active quests. Keep spending wisely!</p>
      )}
      <ul className="space-y-3">
        {quests.map((q) => {
          const color = ATTR_COLORS[q.attribute] ?? '#6b7280'
          return (
            <li key={q.id} className="bg-gray-800 rounded-lg p-3 space-y-2">
              <div className="flex justify-between items-center">
                <span className="font-semibold text-sm">{q.title}</span>
                <span className="badge" style={{ backgroundColor: color + '22', color }}>
                  +{q.reward} XP
                </span>
              </div>
              <p className="text-xs text-gray-400">{q.description}</p>
              <div className="stat-bar">
                <div
                  className="stat-bar-fill"
                  style={{ width: `${q.progress}%`, backgroundColor: color }}
                />
              </div>
              <p className="text-right text-xs text-gray-500">{q.progress}%</p>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
