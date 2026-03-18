package port

import (
	"context"
	"time"
)

// CategoryStat — расходы по одной категории за период.
type CategoryStat struct {
	Category string  `json:"category"`
	Amount   int64   `json:"amount"`   // в копейках/центах
	Count    int     `json:"count"`    // количество транзакций
	Percent  float64 `json:"percent"`  // доля от общих расходов, 0-100
}

// DailyStat — сумма трат за один день (для тренд-графика).
type DailyStat struct {
	Date   time.Time `json:"date"`
	Amount int64     `json:"amount"`
}

// AnalyticsSummary — полный аналитический отчёт за период.
type AnalyticsSummary struct {
	AccountID     string         `json:"account_id"`
	PeriodDays    int            `json:"period_days"`
	TotalExpenses int64          `json:"total_expenses"`
	TotalIncome   int64          `json:"total_income"`
	ByCategory    []CategoryStat `json:"by_category"`
	DailyTrend    []DailyStat    `json:"daily_trend"`
	GeneratedAt   time.Time      `json:"generated_at"`
}

// AnalyticsRepo — порт для аналитических запросов к хранилищу.
// Реализуется в adapter/postgres/analytics_repo.go
type AnalyticsRepo interface {
	// Summary возвращает агрегированную аналитику по аккаунту за последние periodDays дней.
	Summary(ctx context.Context, accountID string, periodDays int) (AnalyticsSummary, error)
}
