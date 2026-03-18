import type { Transaction } from '../api/client'

const CATEGORY_META: Record<string, { icon: string; color: string }> = {
  food:          { icon: '🍕', color: '#ef4444' },
  learning:      { icon: '📖', color: '#34d399' },
  health:        { icon: '💊', color: '#3b82f6' },
  sports:        { icon: '🏋️', color: '#a78bfa' },
  entertainment: { icon: '🎮', color: '#f5c518' },
  shopping:      { icon: '🛍️', color: '#fb923c' },
  transport:     { icon: '🚕', color: '#94a3b8' },
  cafe:          { icon: '☕', color: '#d97706' },
  travel:        { icon: '✈️', color: '#22d3ee' },
  xp:            { icon: '⭐', color: '#f5c518' },
  hp:            { icon: '❤️', color: '#ef4444' },
  mana:          { icon: '💎', color: '#3b82f6' },
  strength:      { icon: '⚔️', color: '#a78bfa' },
  intellect:     { icon: '📖', color: '#34d399' },
  luck:          { icon: '🍀', color: '#fbbf24' },
}

function getCategoryMeta(category?: string) {
  if (!category) return { icon: '💳', color: '#6b7280' }
  const key = category.toLowerCase()
  for (const [k, v] of Object.entries(CATEGORY_META)) {
    if (key.includes(k)) return v
  }
  return { icon: '💳', color: '#6b7280' }
}

function formatAmount(cents: number): string {
  const abs = (Math.abs(cents) / 100).toLocaleString('ru-RU', { minimumFractionDigits: 2 })
  return cents < 0 ? `−${abs} ₽` : `+${abs} ₽`
}

function formatDate(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
}

interface Props {
  transactions: Transaction[]
}

export function TransactionFeed({ transactions }: Props) {
  return (
    <div className="card space-y-3">
      <h2 className="section-label">💳 Последние транзакции</h2>

      {transactions.length === 0 ? (
        <div className="rounded-xl border border-dashed border-gray-700/50 p-8 text-center">
          <p className="text-4xl mb-3">📂</p>
          <p className="text-gray-500 text-sm">Транзакций ещё нет</p>
          <p className="text-gray-700 text-xs mt-1">Импортируй CSV из Тинькофф или Сбер</p>
        </div>
      ) : (
        <ul className="space-y-1.5">
          {transactions.map((tx) => {
            const meta = getCategoryMeta(tx.clean_category)
            const isPending = tx.status === 'pending'
            return (
              <li key={tx.id}
                className="flex items-center gap-3 rounded-xl bg-gray-800/40 border border-gray-800/60 px-3 py-2.5 hover:bg-gray-800/70 hover:border-gray-700/60 transition-all cursor-default"
              >
                {/* Category icon */}
                <div className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 text-lg"
                  style={{ background: `${meta.color}15` }}>
                  {meta.icon}
                </div>

                {/* Description + category */}
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-gray-100 truncate leading-tight">
                    {tx.original_description || 'Операция'}
                  </p>
                  <p className="text-xs mt-0.5">
                    {isPending ? (
                      <span className="text-yellow-500/80 pulse-dot">⚙ Классифицируется…</span>
                    ) : (
                      <span style={{ color: meta.color + 'cc' }}>
                        {tx.clean_category || '—'}
                      </span>
                    )}
                  </p>
                </div>

                {/* Amount + date */}
                <div className="text-right shrink-0">
                  <p className={`text-sm font-semibold font-mono ${tx.amount < 0 ? 'text-red-400' : 'text-green-400'}`}>
                    {formatAmount(tx.amount)}
                  </p>
                  <p className="text-[10px] text-gray-600 mt-0.5">{formatDate(tx.created_at)}</p>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
