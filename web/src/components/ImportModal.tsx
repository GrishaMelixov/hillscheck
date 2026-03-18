import { useState, useCallback, useRef, useEffect } from 'react'
import { fetchAccounts, importTransactions, parseScreenshot } from '../api/client'
import { parseCSV } from '../utils/csvParser'
import type { Account, VisionTransaction } from '../api/client'
import type { ParseResult } from '../utils/csvParser'

interface Props {
  onClose: () => void
  onImported: () => void
}

type Mode = 'csv' | 'screenshot'
type Step = 'upload' | 'preview' | 'parsing' | 'importing' | 'done'

interface PreviewRow {
  externalId: string
  description: string
  amountCents: number
  mcc: number
  occurredAt: string
}

const FORMAT_LABEL: Record<string, { label: string; color: string }> = {
  tinkoff: { label: 'Тинькофф',     color: '#FFD60A' },
  sber:    { label: 'Сбербанк',     color: '#30D158' },
  generic: { label: 'Универсальный', color: 'rgba(255,255,255,0.4)' },
}

function makeExternalId(tx: VisionTransaction): string {
  return `screenshot-${btoa(`${tx.description}|${tx.amount_cents}|${tx.occurred_at}`).replace(/=/g, '').slice(0, 24)}`
}

export default function ImportModal({ onClose, onImported }: Props) {
  const [mode, setMode]               = useState<Mode>('csv')
  const [step, setStep]               = useState<Step>('upload')
  const [dragOver, setDragOver]       = useState(false)
  const [parseResult, setParseResult] = useState<ParseResult | null>(null)
  const [visionRows, setVisionRows]   = useState<PreviewRow[]>([])
  const [accounts, setAccounts]       = useState<Account[]>([])
  const [accountId, setAccountId]     = useState('')
  const [importResult, setImportResult] = useState<{ created: number; duplicates: number } | null>(null)
  const [error, setError]             = useState('')
  const csvRef = useRef<HTMLInputElement>(null)
  const imgRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    fetchAccounts()
      .then(d => {
        setAccounts(d.accounts ?? [])
        if (d.accounts?.length) setAccountId(d.accounts[0].id)
      })
      .catch(() => {})
  }, [])

  // ── CSV processing ──────────────────────────────────────────────────────────

  const processCSV = useCallback(async (file: File) => {
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

  // ── Screenshot processing ───────────────────────────────────────────────────

  const processImage = useCallback(async (file: File) => {
    const allowed = ['image/jpeg', 'image/png', 'image/webp', 'image/heic', 'image/heif']
    const ext = file.name.toLowerCase()
    const ok = allowed.includes(file.type) ||
      ext.endsWith('.jpg') || ext.endsWith('.jpeg') ||
      ext.endsWith('.png') || ext.endsWith('.webp') ||
      ext.endsWith('.heic') || ext.endsWith('.heif')
    if (!ok) {
      setError('Поддерживаются: JPEG, PNG, WEBP, HEIC')
      return
    }
    setError('')
    setStep('parsing')
    try {
      const data = await parseScreenshot(file)
      if (!data.transactions?.length) {
        setError('Транзакции не найдены на скриншоте')
        setStep('upload')
        return
      }
      const rows: PreviewRow[] = data.transactions.map(tx => ({
        externalId:  makeExternalId(tx),
        description: tx.description,
        amountCents: tx.amount_cents,
        mcc:         tx.mcc,
        occurredAt:  tx.occurred_at,
      }))
      setVisionRows(rows)
      setStep('preview')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Ошибка распознавания')
      setStep('upload')
    }
  }, [])

  // ── Drag & drop ─────────────────────────────────────────────────────────────

  const onDropCSV = useCallback((e: React.DragEvent) => {
    e.preventDefault(); setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) processCSV(file)
  }, [processCSV])

  const onDropImg = useCallback((e: React.DragEvent) => {
    e.preventDefault(); setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) processImage(file)
  }, [processImage])

  // ── Import ──────────────────────────────────────────────────────────────────

  const handleImport = async () => {
    if (!accountId) return
    setStep('importing')
    try {
      const rows = mode === 'csv' && parseResult
        ? parseResult.transactions.map(tx => ({
            external_id:          tx.externalId,
            amount:               tx.amountCents,
            mcc:                  tx.mcc,
            original_description: tx.description,
            occurred_at:          tx.occurredAt,
          }))
        : visionRows.map(tx => ({
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

  const previewRows: PreviewRow[] = mode === 'csv' && parseResult
    ? parseResult.transactions.map(tx => ({
        externalId:  tx.externalId,
        description: tx.description,
        amountCents: tx.amountCents,
        mcc:         tx.mcc,
        occurredAt:  tx.occurredAt,
      }))
    : visionRows

  const rowCount = previewRows.length

  // ── Styles ──────────────────────────────────────────────────────────────────

  const modalStyle: React.CSSProperties = {
    background: 'rgba(12,12,14,0.92)',
    border: '1px solid rgba(255,255,255,0.10)',
    borderRadius: '28px',
    backdropFilter: 'blur(60px) saturate(1.8)',
    WebkitBackdropFilter: 'blur(60px) saturate(1.8)',
    boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.08), 0 32px 80px rgba(0,0,0,0.8)',
  }

  const tabActive: React.CSSProperties = {
    background: 'rgba(255,255,255,0.10)',
    color: '#fff',
  }

  const tabInactive: React.CSSProperties = {
    color: 'rgba(255,255,255,0.35)',
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

        <h2 className="text-[17px] font-semibold mb-5" style={{ color: '#F5C518' }}>
          Импорт транзакций
        </h2>

        {/* ── Mode tabs (only on upload step) ── */}
        {(step === 'upload' || step === 'parsing') && (
          <div
            className="flex rounded-xl p-1 mb-5 gap-1"
            style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.06)' }}
          >
            {([['csv', '📄 CSV'], ['screenshot', '📷 Скриншот']] as [Mode, string][]).map(([m, label]) => (
              <button
                key={m}
                onClick={() => { setMode(m); setError('') }}
                className="flex-1 py-2 text-[13px] font-medium rounded-lg transition-all"
                style={mode === m ? tabActive : tabInactive}
              >
                {label}
              </button>
            ))}
          </div>
        )}

        {/* ── CSV UPLOAD ── */}
        {step === 'upload' && mode === 'csv' && (
          <div
            onDrop={onDropCSV}
            onDragOver={e => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onClick={() => csvRef.current?.click()}
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
            {error && <p className="mt-3 text-[13px]" style={{ color: '#FF453A' }}>{error}</p>}
            <input ref={csvRef} type="file" accept=".csv" className="hidden"
              onChange={e => { const f = e.target.files?.[0]; if (f) processCSV(f) }} />
          </div>
        )}

        {/* ── SCREENSHOT UPLOAD ── */}
        {step === 'upload' && mode === 'screenshot' && (
          <div
            onDrop={onDropImg}
            onDragOver={e => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onClick={() => imgRef.current?.click()}
            className="rounded-2xl p-10 text-center cursor-pointer transition-all"
            style={{
              border: `2px dashed ${dragOver ? 'rgba(10,132,255,0.6)' : 'rgba(255,255,255,0.12)'}`,
              background: dragOver ? 'rgba(10,132,255,0.05)' : 'rgba(255,255,255,0.02)',
            }}
          >
            <div className="text-5xl mb-4 float">📱</div>
            <p className="text-[15px] font-medium text-white/70">Скриншот из банковского приложения</p>
            <p className="text-[12px] mt-2" style={{ color: 'rgba(255,255,255,0.30)' }}>
              T‑Банк · Сбер · Кассовый чек · JPEG · PNG · WEBP · HEIC
            </p>
            <p className="text-[11px] mt-3 px-4" style={{ color: 'rgba(255,255,255,0.20)' }}>
              Скриншот обрабатывается Gemini AI — транзакции распознаются автоматически
            </p>
            {error && <p className="mt-3 text-[13px]" style={{ color: '#FF453A' }}>{error}</p>}
            <input ref={imgRef} type="file" accept="image/jpeg,image/png,image/webp,image/heic,.heic,.heif" className="hidden"
              onChange={e => { const f = e.target.files?.[0]; if (f) processImage(f) }} />
          </div>
        )}

        {/* ── PARSING (AI in progress) ── */}
        {step === 'parsing' && (
          <div className="py-16 text-center space-y-4">
            <div className="text-5xl float">🔍</div>
            <p className="text-[15px] font-medium text-white/70">Распознаём скриншот…</p>
            <p className="text-[13px]" style={{ color: 'rgba(255,255,255,0.30)' }}>
              Gemini AI анализирует изображение
            </p>
          </div>
        )}

        {/* ── PREVIEW ── */}
        {step === 'preview' && (
          <div className="space-y-4">
            {/* Header badge */}
            <div className="flex items-center gap-2 flex-wrap">
              {mode === 'csv' && parseResult ? (
                (() => {
                  const fm = FORMAT_LABEL[parseResult.format] ?? FORMAT_LABEL.generic
                  return (
                    <span
                      className="text-[11px] font-semibold px-2.5 py-1 rounded-full"
                      style={{ color: fm.color, background: `${fm.color}18`, border: `1px solid ${fm.color}30` }}
                    >
                      {fm.label}
                    </span>
                  )
                })()
              ) : (
                <span
                  className="text-[11px] font-semibold px-2.5 py-1 rounded-full"
                  style={{ color: '#0A84FF', background: 'rgba(10,132,255,0.12)', border: '1px solid rgba(10,132,255,0.25)' }}
                >
                  📷 Скриншот
                </span>
              )}
              <span className="text-[13px]" style={{ color: 'rgba(255,255,255,0.40)' }}>
                {rowCount} транзакций
                {mode === 'csv' && parseResult?.errors.length ? `, ${parseResult.errors.length} пропущено` : ''}
              </span>
            </div>

            {mode === 'csv' && parseResult?.format === 'generic' && (
              <p
                className="text-[12px] rounded-xl p-3"
                style={{ color: '#FF9F0A', background: 'rgba(255,159,10,0.08)', border: '1px solid rgba(255,159,10,0.15)' }}
              >
                Формат не распознан. Первые колонки: {parseResult.headers.slice(0, 4).join(', ')}
              </p>
            )}

            {/* Table */}
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
                  {previewRows.slice(0, 8).map((tx, i) => (
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
                  {rowCount > 8 && (
                    <tr style={{ borderTop: '1px solid rgba(255,255,255,0.04)' }}>
                      <td colSpan={3} className="px-3 py-2 text-center text-[11px]" style={{ color: 'rgba(255,255,255,0.20)' }}>
                        … и ещё {rowCount - 8}
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
                <select value={accountId} onChange={e => setAccountId(e.target.value)} className="input">
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
                disabled={rowCount === 0 || !accountId}
                className="btn-primary flex-1 py-3"
              >
                Импортировать {rowCount}
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
