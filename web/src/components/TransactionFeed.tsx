import type { Transaction } from '../api/client'

const ATTR_ICONS: Record<string, string> = {
  xp:        '⭐',
  hp:        '❤️',
  mana:      '💎',
  strength:  '⚔️',
  intellect: '📖',
  luck:      '🍀',
}

function formatAmount(cents: number): string {
  const sign = cents < 0 ? '−' : '+'
  return `${sign}$${(Math.abs(cents) / 100).toFixed(2)}`
}

function statusColor(status: string) {
  if (status === 'processed') return 'text-green-400'
  if (status === 'failed')    return 'text-red-400'
  return 'text-yellow-400'
}

interface Props {
  transactions: Transaction[]
}

export function TransactionFeed({ transactions }: Props) {
  if (transactions.length === 0) {
    return (
      <div className="card text-center text-gray-500 py-8">
        No transactions yet. Import your first batch!
      </div>
    )
  }

  return (
    <div className="card space-y-3">
      <h2 className="text-sm text-gray-400 uppercase tracking-wider">Recent Transactions</h2>
      <ul className="space-y-2">
        {transactions.map((tx) => (
          <li key={tx.id} className="flex items-center gap-3 bg-gray-800 rounded-lg px-3 py-2">
            <span className="text-xl" title={tx.clean_category}>
              {ATTR_ICONS[tx.clean_category?.toLowerCase()] ?? '💳'}
            </span>
            <div className="flex-1 min-w-0">
              <p className="text-sm truncate">{tx.original_description || 'Unknown'}</p>
              <p className="text-xs text-gray-500">{tx.clean_category || 'Classifying…'}</p>
            </div>
            <div className="text-right shrink-0">
              <p className={`text-sm font-semibold ${tx.amount < 0 ? 'text-red-400' : 'text-green-400'}`}>
                {formatAmount(tx.amount)}
              </p>
              <p className={`text-xs ${statusColor(tx.status ?? '')}`}>{tx.status}</p>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
