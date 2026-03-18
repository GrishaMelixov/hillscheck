import type { Transaction } from '../api/client'

const CATEGORY_META: Record<string, { icon: string; color: string }> = {
  food:          { icon: '🍕', color: '#FF453A' },
  learning:      { icon: '📖', color: '#30D158' },
  health:        { icon: '💊', color: '#0A84FF' },
  sports:        { icon: '🏋️', color: '#BF5AF2' },
  entertainment: { icon: '🎮', color: '#FFD60A' },
  shopping:      { icon: '🛍️', color: '#FF9F0A' },
  transport:     { icon: '🚕', color: '#5AC8FA' },
  cafe:          { icon: '☕', color: '#FF9F0A' },
  travel:        { icon: '✈️', color: '#5AC8FA' },
  xp:            { icon: '⭐', color: '#FFD60A' },
  hp:            { icon: '❤️', color: '#FF453A' },
  mana:          { icon: '💎', color: '#0A84FF' },
  strength:      { icon: '⚔️', color: '#BF5AF2' },
  intellect:     { icon: '📖', color: '#30D158' },
  luck:          { icon: '🍀', color: '#FFD60A' },
}

function getCategoryMeta(category?: string) {
  if (!category) return { icon: '💳', color: 'rgba(255,255,255,0.3)' }
  const key = category.toLowerCase()
  for (const [k, v] of Object.entries(CATEGORY_META)) {
    if (key.includes(k)) return v
  }
  return { icon: '💳', color: 'rgba(255,255,255,0.3)' }
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
    <div className="glass p-5 space-y-3">
      <h2 className="section-label">Transactions</h2>

      {transactions.length === 0 ? (
        <div
          className="rounded-2xl p-8 text-center"
          style={{ border: '1px dashed rgba(255,255,255,0.10)', background: 'rgba(255,255,255,0.02)' }}
        >
          <p className="text-3xl mb-3 float">📂</p>
          <p className="text-[13px] text-white/40">Транзакций ещё нет</p>
          <p className="text-[11px] text-white/20 mt-1">Импортируй CSV из Тинькофф или Сбер</p>
        </div>
      ) : (
        <ul className="space-y-1">
          {transactions.map((tx) => {
            const meta = getCategoryMeta(tx.clean_category)
            const isPending = tx.status === 'pending'
            const isDebit = tx.amount < 0
            return (
              <li
                key={tx.id}
                className="flex items-center gap-3 rounded-2xl px-3.5 py-3 transition-all cursor-default group"
                style={{ background: 'rgba(255,255,255,0.03)' }}
                onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.06)')}
                onMouseLeave={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.03)')}
              >
                {/* Icon */}
                <div
                  className="w-9 h-9 rounded-[11px] flex items-center justify-center shrink-0 text-[18px]"
                  style={{ background: `${meta.color}18` }}
                >
                  {meta.icon}
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <p className="text-[14px] font-medium text-white/85 truncate leading-tight">
                    {tx.original_description || 'Операция'}
                  </p>
                  <p className="text-[11px] mt-0.5">
                    {isPending ? (
                      <span className="pulse-dot" style={{ color: '#FF9F0A' }}>⚙ Классифицируется…</span>
                    ) : (
                      <span style={{ color: `${meta.color}99` }}>{tx.clean_category || '—'}</span>
                    )}
                  </p>
                </div>

                {/* Amount */}
                <div className="text-right shrink-0">
                  <p className={`text-[14px] font-semibold font-mono ${isDebit ? 'text-[#FF453A]' : 'text-[#30D158]'}`}>
                    {formatAmount(tx.amount)}
                  </p>
                  <p className="text-[10px] mt-0.5" style={{ color: 'rgba(255,255,255,0.25)' }}>
                    {formatDate(tx.created_at)}
                  </p>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
