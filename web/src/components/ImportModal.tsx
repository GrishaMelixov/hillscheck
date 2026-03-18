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

const FORMAT_LABEL: Record<string, { label: string; color: string }> = {
  tinkoff: { label: 'Тинькофф',     color: '#FFD60A' },
  sber:    { label: 'Сбербанк',     color: '#30D158' },
  generic: { label: 'Универсальный', color: 'rgba(255,255,255,0.4)' },
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
      .catch(() => {})
  }, [])

  const processFile = useCallback(async (file: File) => {
    if (!file.name.toLowerCase().endsWith('.csv')) {
      setError('Загрузи CSV-файл')
      return
    }
    setError('')
    const text = await file.text()
    const result = parseCSV(text)
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

  const modalStyle: React.CSSProperties = {
    background: 'rgba(12,12,14,0.92)',
    border: '1px solid rgba(255,255,255,0.10)',
    borderRadius: '28px',
    backdropFilter: 'blur(60px) saturate(1.8)',
    WebkitBackdropFilter: 'blur(60px) saturate(1.8)',
    boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.08), 0 32px 80px rgba(0,0,0,0.8)',
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center px-4"
      style={{ background: 'rgba(0,0,0,0.75)', backdropFilter: 'blur(8px)' }}
      onClick={e => { if (e.target === e.currentTarget) onClose() }}
    >
      <div className="w-full max-w-lg p-7 relative" style={modalStyle}>

        {/* Close */}
        <button
          onClick={onClose}
          className="absolute top-5 right-5 w-8 h-8 flex items-center justify-center rounded-full text-white/40 hover:text-white/70 transition-colors text-xl leading-none"
          style={{ background: 'rgba(255,255,255,0.06)' }}
        >
          ×
        </button>

        <h2 className="text-[17px] font-semibold mb-6" style={{ color: '#F5C518' }}>
          Импорт транзакций
        </h2>

        {/* ── UPLOAD ── */}
        {step === 'upload' && (
          <div
            onDrop={onDrop}
            onDragOver={e => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onClick={() => fileRef.current?.click()}
            className="rounded-2xl p-12 text-center cursor-pointer transition-all"
            style={{
              border: `2px dashed ${dragOver ? 'rgba(245,197,24,0.6)' : 'rgba(255,255,255,0.12)'}`,
              background: dragOver ? 'rgba(245,197,24,0.05)' : 'rgba(255,255,255,0.02)',
            }}
          >
            <div className="text-5xl mb-4 float">📂</div>
            <p className="text-[15px] font-medium text-white/70">Перетащи CSV или кликни</p>
            <p className="text-[12px] mt-2" style={{ color: 'rgba(255,255,255,0.30)' }}>
              Тинькофф · Сбер · Универсальный CSV
            </p>
            {error && (
              <p className="mt-3 text-[13px]" style={{ color: '#FF453A' }}>{error}</p>
            )}
            <input ref={fileRef} type="file" accept=".csv" className="hidden" onChange={onInput} />
          </div>
        )}

        {/* ── PREVIEW ── */}
        {step === 'preview' && parseResult && (
          <div className="space-y-4">
            {/* Format badge + count */}
            <div className="flex items-center gap-2 flex-wrap">
              {(() => {
                const fm = FORMAT_LABEL[parseResult.format] ?? FORMAT_LABEL.generic
                return (
                  <span
                    className="text-[11px] font-semibold px-2.5 py-1 rounded-full"
                    style={{ color: fm.color, background: `${fm.color}18`, border: `1px solid ${fm.color}30` }}
                  >
                    {fm.label}
                  </span>
                )
              })()}
              <span className="text-[13px]" style={{ color: 'rgba(255,255,255,0.40)' }}>
                {parseResult.transactions.length} транзакций
                {parseResult.errors.length > 0 && `, ${parseResult.errors.length} пропущено`}
              </span>
            </div>

            {parseResult.format === 'generic' && (
              <p
                className="text-[12px] rounded-xl p-3"
                style={{ color: '#FF9F0A', background: 'rgba(255,159,10,0.08)', border: '1px solid rgba(255,159,10,0.15)' }}
              >
                Формат не распознан автоматически. Первые колонки: {parseResult.headers.slice(0, 4).join(', ')}
              </p>
            )}

            {/* Preview table */}
            <div
              className="overflow-auto rounded-2xl"
              style={{ maxHeight: '176px', border: '1px solid rgba(255,255,255,0.08)' }}
            >
              <table className="text-[12px] w-full">
                <thead style={{ background: 'rgba(255,255,255,0.05)' }}>
                  <tr>
                    <th className="px-3 py-2 text-left font-medium" style={{ color: 'rgba(255,255,255,0.40)' }}>Дата</th>
                    <th className="px-3 py-2 text-left font-medium" style={{ color: 'rgba(255,255,255,0.40)' }}>Описание</th>
                    <th className="px-3 py-2 text-right font-medium" style={{ color: 'rgba(255,255,255,0.40)' }}>Сумма</th>
                  </tr>
                </thead>
                <tbody>
                  {parseResult.transactions.slice(0, 8).map((tx, i) => (
                    <tr key={i} style={{ borderTop: '1px solid rgba(255,255,255,0.04)' }}>
                      <td className="px-3 py-2 whitespace-nowrap" style={{ color: 'rgba(255,255,255,0.40)' }}>
                        {tx.occurredAt.slice(0, 10)}
                      </td>
                      <td className="px-3 py-2 max-w-[180px] truncate" style={{ color: 'rgba(255,255,255,0.75)' }}>
                        {tx.description || '—'}
                      </td>
                      <td className={`px-3 py-2 text-right font-mono font-semibold ${tx.amountCents < 0 ? 'text-[#FF453A]' : 'text-[#30D158]'}`}>
                        {(tx.amountCents / 100).toFixed(2)}
                      </td>
                    </tr>
                  ))}
                  {parseResult.transactions.length > 8 && (
                    <tr style={{ borderTop: '1px solid rgba(255,255,255,0.04)' }}>
                      <td colSpan={3} className="px-3 py-2 text-center text-[11px]" style={{ color: 'rgba(255,255,255,0.20)' }}>
                        … и ещё {parseResult.transactions.length - 8}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Account selector */}
            {accounts.length > 0 && (
              <div className="space-y-1.5">
                <label className="block text-[11px] font-semibold uppercase tracking-widest" style={{ color: 'rgba(255,255,255,0.35)' }}>
                  Счёт
                </label>
                <select
                  value={accountId}
                  onChange={e => setAccountId(e.target.value)}
                  className="input"
                >
                  {accounts.map(a => (
                    <option key={a.id} value={a.id} style={{ background: '#1a1a1a' }}>
                      {a.name} ({a.currency})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {error && <p className="text-[13px]" style={{ color: '#FF453A' }}>{error}</p>}

            <div className="flex gap-3 pt-1">
              <button onClick={() => setStep('upload')} className="btn-glass flex-1 py-3">
                Назад
              </button>
              <button
                onClick={handleImport}
                disabled={parseResult.transactions.length === 0 || !accountId}
                className="btn-primary flex-1 py-3"
              >
                Импортировать {parseResult.transactions.length}
              </button>
            </div>
          </div>
        )}

        {/* ── IMPORTING ── */}
        {step === 'importing' && (
          <div className="py-16 text-center space-y-4">
            <div className="text-5xl float">⚙️</div>
            <p className="text-[15px] font-medium text-white/70">Обрабатываем транзакции…</p>
            <p className="text-[13px]" style={{ color: 'rgba(255,255,255,0.30)' }}>
              Классифицируем каждую операцию
            </p>
          </div>
        )}

        {/* ── DONE ── */}
        {step === 'done' && importResult && (
          <div className="py-12 text-center space-y-5">
            <div className="text-5xl">✅</div>
            <div>
              <p className="text-[20px] font-bold" style={{ color: '#F5C518' }}>Готово!</p>
              <div className="mt-3 space-y-1 text-[14px]" style={{ color: 'rgba(255,255,255,0.50)' }}>
                <p>Добавлено: <span className="font-bold text-[#30D158]">{importResult.created}</span></p>
                <p>Дубликатов: <span className="font-semibold text-white/30">{importResult.duplicates}</span></p>
              </div>
            </div>
            <button
              onClick={() => { onClose(); onImported() }}
              className="btn-primary px-10 py-3"
            >
              На дашборд
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
