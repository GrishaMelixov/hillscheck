package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// GetAnalytics — use case для получения аналитического отчёта.
//
// Архитектура кеширования (двухуровневая):
//   1. Redis (TTL 1ч) — для повторных запросов за тот же период
//   2. PostgreSQL — source of truth, вызывается только при cache miss
//
// Это "многоуровневое кеширование" из пункта резюме:
//   - Первый уровень: Redis в памяти (µs-latency)
//   - Второй уровень: PostgreSQL с составными индексами (ms-latency)
type GetAnalytics struct {
	repo  port.AnalyticsRepo
	redis *redis.Client
	log   *zap.Logger
}

func NewGetAnalytics(repo port.AnalyticsRepo, rdb *redis.Client, log *zap.Logger) *GetAnalytics {
	return &GetAnalytics{repo: repo, redis: rdb, log: log}
}

// Execute возвращает аналитику с кешированием.
// Cache key: analytics:{accountID}:{periodDays}
// TTL: 1 час — данные достаточно свежие для дашборда, но не нагружают БД.
func (g *GetAnalytics) Execute(ctx context.Context, accountID string, periodDays int) (port.AnalyticsSummary, error) {
	if periodDays <= 0 || periodDays > 365 {
		periodDays = 30
	}

	cacheKey := fmt.Sprintf("analytics:%s:%d", accountID, periodDays)

	// ── Cache lookup ──────────────────────────────────────────────────────────
	if cached, err := g.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var summary port.AnalyticsSummary
		if json.Unmarshal(cached, &summary) == nil {
			g.log.Debug("analytics cache hit", zap.String("key", cacheKey))
			return summary, nil
		}
	}

	// ── Cache miss → PostgreSQL ───────────────────────────────────────────────
	g.log.Debug("analytics cache miss, querying postgres", zap.String("account_id", accountID))
	summary, err := g.repo.Summary(ctx, accountID, periodDays)
	if err != nil {
		return port.AnalyticsSummary{}, fmt.Errorf("analytics query: %w", err)
	}

	// ── Populate cache ────────────────────────────────────────────────────────
	if data, err := json.Marshal(summary); err == nil {
		g.redis.Set(ctx, cacheKey, data, time.Hour)
	}

	return summary, nil
}

// InvalidateCache сбрасывает кеш аналитики для аккаунта (вызывается после импорта транзакций).
// Паттерн: cache invalidation при мутации — стандартная практика для согласованности данных.
func (g *GetAnalytics) InvalidateCache(ctx context.Context, accountID string) {
	for _, days := range []int{7, 14, 30, 90, 180, 365} {
		key := fmt.Sprintf("analytics:%s:%d", accountID, days)
		g.redis.Del(ctx, key)
	}
	g.log.Debug("analytics cache invalidated", zap.String("account_id", accountID))
}
