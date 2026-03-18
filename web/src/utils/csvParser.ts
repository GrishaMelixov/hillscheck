// CSV parser — supports Tinkoff, Sber, and generic formats.
// Handles semicolon and comma delimiters, quoted fields.

export interface ParsedTransaction {
  externalId: string
  amountCents: number   // negative = expense
  mcc: number
  description: string
  occurredAt: string    // ISO 8601 / RFC3339
}

export type BankFormat = 'tinkoff' | 'sber' | 'generic'

export interface ParseResult {
  format: BankFormat
  headers: string[]
  rows: string[][]
  transactions: ParsedTransaction[]
  errors: string[]
}

function detectDelimiter(line: string): string {
  const sc = (line.match(/;/g) || []).length
  const co = (line.match(/,/g) || []).length
  return sc >= co ? ';' : ','
}

function parseCSVLine(line: string, delimiter: string): string[] {
  const result: string[] = []
  let current = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (ch === '"') {
      inQuotes = !inQuotes
    } else if (ch === delimiter && !inQuotes) {
      result.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  result.push(current.trim())
  return result
}

function col(headers: string[], row: string[], name: string): string {
  const i = headers.findIndex(h => h.toLowerCase().includes(name.toLowerCase()))
  return i >= 0 ? (row[i] ?? '').trim() : ''
}

function parseAmountCents(raw: string): number {
  const cleaned = raw.replace(/\s/g, '').replace(',', '.')
  const n = parseFloat(cleaned)
  return isNaN(n) ? 0 : Math.round(n * 100)
}

// Parses "DD.MM.YYYY[ HH:MM:SS]" → RFC3339
function parseDMY(dateStr: string): string {
  const [datePart, timePart = '00:00:00'] = dateStr.split(' ')
  const [d, m, y] = datePart.split('.')
  if (!d || !m || !y) return new Date().toISOString()
  return `${y}-${m.padStart(2, '0')}-${d.padStart(2, '0')}T${timePart}+03:00`
}

function detectFormat(headers: string[]): BankFormat {
  const h = headers.map(x => x.toLowerCase())
  if (h.some(x => x.includes('дата операции')) && h.some(x => x.includes('mcc'))) return 'tinkoff'
  if (h.some(x => x.includes('вид операции'))) return 'sber'
  return 'generic'
}

// Tinkoff: Дата операции;Дата платежа;Номер карты;Статус;Сумма операции;Валюта операции;...;MCC;Описание
function parseTinkoffRow(headers: string[], row: string[], idx: number): ParsedTransaction | null {
  const amountCents = parseAmountCents(col(headers, row, 'сумма операции'))
  if (amountCents === 0) return null
  const dateStr = col(headers, row, 'дата операции')
  const mcc = parseInt(col(headers, row, 'mcc')) || 0
  const description = col(headers, row, 'описание') || col(headers, row, 'категория')
  return {
    externalId: `tinkoff-${idx}-${dateStr.replace(/[^0-9]/g, '')}-${Math.abs(amountCents)}`,
    amountCents,
    mcc,
    description,
    occurredAt: parseDMY(dateStr),
  }
}

// Sber: Дата операции;Вид операции;Счёт;Сумма;Валюта;Описание
function parseSberRow(headers: string[], row: string[], idx: number): ParsedTransaction | null {
  const amountCents = parseAmountCents(col(headers, row, 'сумма'))
  if (amountCents === 0) return null
  const dateStr = col(headers, row, 'дата операции')
  const description = col(headers, row, 'описание') || col(headers, row, 'вид операции')
  return {
    externalId: `sber-${idx}-${dateStr.replace(/[^0-9]/g, '')}-${Math.abs(amountCents)}`,
    amountCents,
    mcc: 0,
    description,
    occurredAt: parseDMY(dateStr),
  }
}

export function parseCSV(text: string): ParseResult {
  // Strip BOM if present
  const clean = text.replace(/^\uFEFF/, '')
  const lines = clean.split('\n').map(l => l.replace(/\r$/, '')).filter(l => l.trim())

  if (lines.length < 2) {
    return {
      format: 'generic',
      headers: [],
      rows: [],
      transactions: [],
      errors: ['Файл пуст или содержит только заголовок'],
    }
  }

  const delimiter = detectDelimiter(lines[0])
  const headers = parseCSVLine(lines[0], delimiter)
  const format = detectFormat(headers)
  const rows: string[][] = []
  const transactions: ParsedTransaction[] = []
  const errors: string[] = []

  for (let i = 1; i < lines.length; i++) {
    const row = parseCSVLine(lines[i], delimiter)
    rows.push(row)

    if (format === 'generic') continue // user must map columns manually

    const tx = format === 'tinkoff'
      ? parseTinkoffRow(headers, row, i)
      : parseSberRow(headers, row, i)

    if (tx) {
      transactions.push(tx)
    } else {
      errors.push(`Строка ${i + 1}: пропущена (нулевая сумма или неверный формат)`)
    }
  }

  return { format, headers, rows, transactions, errors }
}
