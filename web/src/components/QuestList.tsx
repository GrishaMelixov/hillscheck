
interface Quest {
  id: string
  title: string
  description: string
  attribute: string
  reward: number
  progress: number
}

const ATTR_META: Record<string, { icon: string; color: string }> = {
  xp:        { icon: '✨', color: '#FFD60A' },
  hp:        { icon: '❤️', color: '#FF453A' },
  mana:      { icon: '💎', color: '#0A84FF' },
  strength:  { icon: '⚔️', color: '#BF5AF2' },
  intellect: { icon: '📖', color: '#30D158' },
  luck:      { icon: '🍀', color: '#FF9F0A' },
}
const DEFAULT_META = { icon: '🎯', color: 'rgba(255,255,255,0.4)' }

interface Props {
  quests: Quest[]
}

export function QuestList({ quests }: Props) {
  return (
    <div className="glass p-5 space-y-4">
      <h2 className="section-label">Active Quests</h2>

      {quests.length === 0 ? (
        <div
          className="rounded-2xl p-6 text-center"
          style={{ border: '1px dashed rgba(255,255,255,0.10)', background: 'rgba(255,255,255,0.02)' }}
        >
          <p className="text-2xl mb-2 float">🗺️</p>
          <p className="text-[13px] text-white/40">Квестов пока нет</p>
          <p className="text-[11px] text-white/20 mt-1">Импортируй транзакции</p>
        </div>
      ) : (
        <ul className="space-y-2.5">
          {quests.map((q) => {
            const meta = ATTR_META[q.attribute] ?? DEFAULT_META
            return (
              <li
                key={q.id}
                className="rounded-2xl p-4 space-y-3 transition-all"
                style={{
                  background: 'rgba(255,255,255,0.04)',
                  border: '1px solid rgba(255,255,255,0.08)',
                }}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <span className="text-lg shrink-0">{meta.icon}</span>
                    <span className="text-[14px] font-semibold text-white/90 leading-snug truncate">{q.title}</span>
                  </div>
                  <span
                    className="shrink-0 text-[11px] font-bold px-2 py-0.5 rounded-full"
                    style={{
                      color: meta.color,
                      background: `${meta.color}18`,
                      border: `1px solid ${meta.color}30`,
                    }}
                  >
                    +{q.reward}
                  </span>
                </div>

                <p className="text-[12px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.4)' }}>
                  {q.description}
                </p>

                <div className="space-y-1">
                  <div className="relative h-[3px] w-full rounded-full overflow-hidden" style={{ background: 'rgba(255,255,255,0.07)' }}>
                    <div
                      className="h-full rounded-full transition-all duration-700"
                      style={{
                        width: `${q.progress}%`,
                        backgroundColor: meta.color,
                        boxShadow: `0 0 8px ${meta.color}60`,
                      }}
                    />
                  </div>
                  <div className="flex justify-end">
                    <span className="text-[10px]" style={{ color: 'rgba(255,255,255,0.25)' }}>{q.progress}%</span>
                  </div>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
