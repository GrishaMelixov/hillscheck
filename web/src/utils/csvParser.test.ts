import { describe, it, expect } from 'vitest'
import { parseCSV } from './csvParser'

// Minimal Tinkoff-format CSV (semicolon, Russian headers with MCC column)
const TINKOFF_CSV = `Дата операции;Дата платежа;Номер карты;Статус;Сумма операции;Валюта операции;Сумма платежа;Валюта платежа;Кэшбэк;Категория;MCC;Описание;Бонусы (начислено)
12.03.2025 14:35:00;13.03.2025;*1234;OK;-250,50;RUB;-250,50;RUB;0;Рестораны;5812;Кафе Уют;0
05.03.2025 09:00:00;06.03.2025;*1234;OK;5000,00;RUB;5000,00;RUB;0;Переводы;0;Пополнение;0`

// Minimal Sber-format CSV
const SBER_CSV = `Дата операции;Вид операции;Счёт;Сумма;Валюта;Описание
10.03.2025;Покупка;40817810000000000001;-1500,00;RUB;Магазин Пятёрочка
08.03.2025;Зачисление;40817810000000000001;20000,00;RUB;Зарплата`

const EMPTY_CSV = `Col1;Col2`

describe('csvParser — detectFormat', () => {
  it('detects Tinkoff format', () => {
    const { format } = parseCSV(TINKOFF_CSV)
    expect(format).toBe('tinkoff')
  })

  it('detects Sber format', () => {
    const { format } = parseCSV(SBER_CSV)
    expect(format).toBe('sber')
  })

  it('falls back to generic for unknown headers', () => {
    const { format } = parseCSV(EMPTY_CSV)
    expect(format).toBe('generic')
  })
})

describe('csvParser — Tinkoff parsing', () => {
  it('parses correct number of rows', () => {
    const { transactions } = parseCSV(TINKOFF_CSV)
    expect(transactions).toHaveLength(2)
  })

  it('converts amount to cents', () => {
    const { transactions } = parseCSV(TINKOFF_CSV)
    expect(transactions[0].amountCents).toBe(-25050)
    expect(transactions[1].amountCents).toBe(500000)
  })

  it('extracts MCC', () => {
    const { transactions } = parseCSV(TINKOFF_CSV)
    expect(transactions[0].mcc).toBe(5812)
  })

  it('extracts description', () => {
    const { transactions } = parseCSV(TINKOFF_CSV)
    expect(transactions[0].description).toBe('Кафе Уют')
  })

  it('builds RFC3339-compatible occurredAt', () => {
    const { transactions } = parseCSV(TINKOFF_CSV)
    expect(transactions[0].occurredAt).toMatch(/^2025-03-12T/)
  })

  it('generates unique externalIds', () => {
    const { transactions } = parseCSV(TINKOFF_CSV)
    const ids = new Set(transactions.map(t => t.externalId))
    expect(ids.size).toBe(transactions.length)
  })
})

describe('csvParser — Sber parsing', () => {
  it('parses correct number of rows', () => {
    const { transactions } = parseCSV(SBER_CSV)
    expect(transactions).toHaveLength(2)
  })

  it('converts amount to cents', () => {
    const { transactions } = parseCSV(SBER_CSV)
    expect(transactions[0].amountCents).toBe(-150000)
    expect(transactions[1].amountCents).toBe(2000000)
  })

  it('extracts description', () => {
    const { transactions } = parseCSV(SBER_CSV)
    expect(transactions[0].description).toBe('Магазин Пятёрочка')
  })
})

describe('csvParser — edge cases', () => {
  it('returns empty transactions for generic format', () => {
    const { transactions } = parseCSV(EMPTY_CSV)
    expect(transactions).toHaveLength(0)
  })

  it('returns error when file has fewer than 2 lines', () => {
    const { errors } = parseCSV('')
    expect(errors.length).toBeGreaterThan(0)
  })

  it('strips BOM from UTF-8 files', () => {
    const withBom = '\uFEFF' + TINKOFF_CSV
    const { format } = parseCSV(withBom)
    expect(format).toBe('tinkoff')
  })
})
