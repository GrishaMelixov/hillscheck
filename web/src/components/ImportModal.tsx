import { useState, useCallback, useRef, useEffect } from 'react'
import { fetchAccounts, importTransactions } from '../api/client'
import { parseCSV } from '../utils/csvParser'
import type { Account } from '../api/client'
import type { ParseResult } from '../utils/csvParser'

interface Props {
  onClose: () => void
  onImported: () => void
}

type Step = 'upload' | 'preview' | 'importing' | 'done'

const FORMAT_LABEL: Record<string, string> = {
  tinkoff: '🟡 Тинькофф',
  sber:    '🟢 Сбер',
  generic: '⚪ Универсальный',
}

export default function ImportModal({ onClose, onImported }: Props) {
  const [step, setStep]               = useState<Step>('upload')
  const [dragOver, setDragOver]       = useState(false)
  const [parseResult, setParseResult] = useState<ParseResult | null>(null)
  const [accounts, setAccounts]       = useState<Account[]>([])
  const [accountId, setAccountId]     = useState('')
  const [importResult, setImportResult] = useState<{ created: number; duplicates: number } | null>(null)
  const [error, setError]             = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    fetchAccounts()
      .then(d => {
        setAccounts(d.accounts ?? [])
        if (d.accounts?.length) setAccountId(d.accounts[0].id)
      })
      .catch(() => {/* handled below during import */})
  }, [])

  const processFile = useCallback(async (file: File) => {
    if (!file.name.toLowerCase().endsWith('.csv')) {
      setError('Загрузи CSV-файл')
      return
    }
    setError('')
    const text = await file.text()
    const result = parseCSV(text)
    if (result.format === 'generic' && result.transactions.length === 0) {
      // Generic format — we still show preview but warn
    }
    setParseResult(result)
    setStep('preview')
  }, [])

  const onDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) processFile(file)
  }, [processFile])

  const onInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) processFile(file)
  }, [processFile])

  const handleImport = async () => {
    if (!parseResult || !accountId) return
    setStep('importing')
    try {
      const rows = parseResult.transactions.map(tx => ({
        external_id:          tx.externalId,
        amount:               tx.amountCents,
        mcc:                  tx.mcc,
        original_description: tx.description,
        occurred_at:          tx.occurredAt,
      }))
      const res = await importTransactions(accountId, rows)
      setImportResult(res)
      setStep('done')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Ошибка импорта')
      setStep('preview')
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 px-4">
      <div className="bg-gray-900 border border-gray-700 rounded-xl w-full max-w-lg p-6 relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-500 hover:text-gray-300 text-xl leading-none"
        >
          ×
        </button>

        <h2 className="text-rpg-gold font-bold tracking-wider text-lg mb-5">⚔ Импорт транзакций</h2>

        {/* ── UPLOAD ── */}
        {step === 'upload' && (
          <div
            onDrop={onDrop}
            onDragOver={e => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onClick={() => fileRef.current?.click()}
            className={`border-2 border-dashed rounded-xl p-12 text-center cursor-pointer transition ${
              dragOver ? 'border-rpg-gold bg-yellow-900/10' : 'border-gray-700 hover:border-gray-600'
            }`}
          >
            <div className="text-5xl mb-4">📂</div>
            <p className="text-gray-200 text-sm font-medium">Перетащи CSV или кликни для выбора</p>
            <p className="text-gray-500 text-xs mt-2">Поддерживается: Тинькофф, Сбер, универсальный CSV</p>
            {error && <p className="text-red-400 text-xs mt-3">{error}</p>}
            <input ref={fileRef} type="file" accept=".csv" className="hidden" onChange={onInput} />
          </div>
        )}

        {/* ── PREVIEW ── */}
        {step === 'preview' && parseResult && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-xs px-2 py-0.5 rounded bg-gray-800 text-gray-300 tracking-wider">
                {FORMAT_LABEL[parseResult.format]}
              </span>
              <span className="text-gray-500 text-xs">
                {parseResult.transactions.length} транзакций
                {parseResult.errors.length > 0 && `, ${parseResult.errors.length} пропущено`}
              </span>
            </div>

            {parseResult.format === 'generic' && (
              <p className="text-yellow-400 text-xs bg-yellow-900/20 rounded p-2">
                Формат не распознан автоматически. Первые колонки: {parseResult.headers.slice(0, 4).join(', ')}
              </p>
            )}

            {/* Preview table */}
            <div className="overflow-auto max-h-44 rounded-lg border border-gray-700">
              <table className="text-xs w-full">
                <thead className="bg-gray-800 text-gray-400 sticky top-0">
                  <tr>
                    <th className="px-3 py-2 text-left">Дата</th>
                    <th className="px-3 py-2 text-left">Описание</th>
                    <th className="px-3 py-2 text-right">Сумма</th>
                  </tr>
                </thead>
                <tbody>
                  {parseResult.transactions.slice(0, 8).map((tx, i) => (
                    <tr key={i} className="border-t border-gray-800 hover:bg-gray-800/50">
                      <td className="px-3 py-1.5 text-gray-400 whitespace-nowrap">
                        {tx.occurredAt.slice(0, 10)}
                      </td>
                      <td className="px-3 py-1.5 text-gray-200 max-w-[180px] truncate">
                        {tx.description || '—'}
                      </td>
                      <td className={`px-3 py-1.5 text-right font-mono ${
                        tx.amountCents < 0 ? 'text-red-400' : 'text-green-400'
                      }`}>
                        {(tx.amountCents / 100).toFixed(2)}
                      </td>
                    </tr>
                  ))}
                  {parseResult.transactions.length > 8 && (
                    <tr className="border-t border-gray-800">
                      <td colSpan={3} className="px-3 py-1.5 text-gray-600 text-center text-xs">
                        … и ещё {parseResult.transactions.length - 8}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Account selector */}
            {accounts.length > 0 && (
              <div>
                <label className="block text-xs text-gray-400 mb-1 uppercase tracking-wider">Счёт</label>
                <select
                  value={accountId}
                  onChange={e => setAccountId(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-rpg-gold"
                >
                  {accounts.map(a => (
                    <option key={a.id} value={a.id}>
                      {a.name} ({a.currency})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {error && <p className="text-red-400 text-xs">{error}</p>}

            <div className="flex gap-3">
              <button
                onClick={() => setStep('upload')}
                className="flex-1 py-2 rounded-lg border border-gray-700 text-sm text-gray-400 hover:text-gray-200 hover:border-gray-500 transition"
              >
                Назад
              </button>
              <button
                onClick={handleImport}
                disabled={parseResult.transactions.length === 0 || !accountId}
                className="flex-1 py-2 rounded-lg bg-rpg-gold text-gray-950 font-bold text-sm hover:brightness-110 transition disabled:opacity-50"
              >
                Импортировать {parseResult.transactions.length}
              </button>
            </div>
          </div>
        )}

        {/* ── IMPORTING ── */}
        {step === 'importing' && (
          <div className="py-14 text-center">
            <div className="text-5xl mb-4 animate-bounce">⚙️</div>
            <p className="text-gray-300 text-sm">Обрабатываем транзакции…</p>
            <p className="text-gray-600 text-xs mt-2">AI классифицирует каждую</p>
          </div>
        )}

        {/* ── DONE ── */}
        {step === 'done' && importResult && (
          <div className="py-10 text-center space-y-4">
            <div className="text-5xl">⚔️</div>
            <p className="text-rpg-gold font-bold text-xl tracking-wider">Готово!</p>
            <div className="text-sm text-gray-400 space-y-1">
              <p>Добавлено: <span className="text-green-400 font-bold">{importResult.created}</span></p>
              <p>Дубликатов: <span className="text-gray-500">{importResult.duplicates}</span></p>
            </div>
            <button
              onClick={() => { onClose(); onImported() }}
              className="mt-2 px-8 py-2 bg-rpg-gold text-gray-950 font-bold rounded-lg text-sm hover:brightness-110 transition"
            >
              На дашборд
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
