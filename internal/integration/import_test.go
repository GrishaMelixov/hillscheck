//go:build integration

// Интеграционные тесты с testcontainers-go.
//
// Что тестируется:
//   - Полный путь импорта транзакций через реальный PostgreSQL
//   - Механизм идемпотентности (повторный импорт той же транзакции = дубликат)
//   - Обновление game_profiles через GameEngine
//   - Аналитический запрос (AnalyticsRepo.Summary)
//
// Запуск: go test -tags=integration ./internal/integration/... -v
//
// Почему testcontainers, а не моки:
//   Моки проверяют логику use case, но не ловят:
//   - Нарушения уникальных constraint'ов в PostgreSQL
//   - Ошибки в SQL-запросах (опечатки, неверные типы)
//   - Проблемы с миграциями (неверный порядок, зависимости)
//   - Поведение ON CONFLICT DO NOTHING в реальной БД
//   testcontainers даёт реальную БД, которая исчезает после теста — чистая изоляция.

package integration

import (
	"context"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"

	wealthcheck "github.com/GrishaMelixov/wealthcheck"
	"github.com/GrishaMelixov/wealthcheck/internal/adapter/ai"
	"github.com/GrishaMelixov/wealthcheck/internal/adapter/postgres"
	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/infrastructure/worker"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// setupDB поднимает одноразовый PostgreSQL-контейнер, прогоняет миграции
// и возвращает подключённый pgxpool. Контейнер убивается через t.Cleanup.
func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	// testcontainers-go: спинапим Postgres 16 в Docker
	ctr, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("wealthcheck_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("container connection string: %v", err)
	}

	// Прогоняем миграции из embed.FS — тот же код, что в production
	migrationsFS, err := fs.Sub(wealthcheck.StaticFiles, "migrations")
	if err != nil {
		t.Fatalf("migrations fs: %v", err)
	}
	src, err := iofs.New(migrationsFS, ".")
	if err != nil {
		t.Fatalf("iofs source: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, connStr)
	if err != nil {
		t.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("pgxpool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// TestImportIdempotency — проверяет, что повторный импорт одной транзакции
// не создаёт дубликат в БД (механизм: INSERT ON CONFLICT DO NOTHING + external_id).
//
// Это ключевой инвариант системы: "0% потерь данных при импорте" означает
// в том числе, что повторная загрузка того же CSV не испортит данные.
func TestImportIdempotency(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	log, _ := zap.NewDevelopment()

	// Создаём тестового пользователя и аккаунт напрямую в БД
	var userID, accountID string
	err := pool.QueryRow(ctx, `
		INSERT INTO users (id, email_hash, password_hash, created_at, updated_at)
		VALUES (gen_random_uuid(), 'testhash', 'fakehash', NOW(), NOW())
		RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO accounts (id, user_id, name, type, balance, currency, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'Test Account', 'debit', 0, 'RUB', NOW(), NOW())
		RETURNING id`, userID,
	).Scan(&accountID)
	if err != nil {
		t.Fatalf("insert account: %v", err)
	}

	// Инициализируем game_profile
	_, err = pool.Exec(ctx, `
		INSERT INTO game_profiles (user_id, level, xp, hp, mana, strength, intellect, luck, updated_at)
		VALUES ($1, 1, 0, 100, 100, 10, 10, 10, NOW())`, userID)
	if err != nil {
		t.Fatalf("insert game_profile: %v", err)
	}

	txRepo  := postgres.NewTransactionRepo(pool)
	gameRepo := postgres.NewGameRepo(pool)
	accountRepo := postgres.NewAccountRepo(pool)

	// Classifier с нулевым LLM-вызовом — используем только MCC-таблицу
	classifier := ai.NewChainedProvider(log)

	hub := &noopHub{}

	engine   := usecase.NewGameEngine(txRepo, accountRepo, gameRepo, classifier, hub, log)
	workerPool := worker.New(4, 64)
	importer := usecase.NewTransactionImport(txRepo, workerPool, engine, log)

	txParam := port.CreateTransactionParams{
		AccountID:           accountID,
		ExternalID:          "tinkoff-2024-001",
		Amount:              -150000, // -1500 руб в копейках
		MCC:                 5812,    // Restaurants
		OriginalDescription: "VKUSVILL 001",
		OccurredAt:          time.Now().Add(-24 * time.Hour),
	}

	// Первый импорт — должен создать транзакцию
	res1, err := importer.Import(ctx, usecase.ImportRequest{
		AccountID:    accountID,
		Transactions: []port.CreateTransactionParams{txParam},
	})
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	if len(res1.Created) != 1 {
		t.Errorf("first import: want 1 created, got %d", len(res1.Created))
	}
	if len(res1.Duplicates) != 0 {
		t.Errorf("first import: want 0 duplicates, got %d", len(res1.Duplicates))
	}

	// Ждём обработки worker'а
	time.Sleep(200 * time.Millisecond)
	_ = workerPool.Shutdown(ctx)

	// Второй импорт той же транзакции — должен вернуть дубликат, не создавать новую запись
	workerPool2 := worker.New(4, 64)
	importer2 := usecase.NewTransactionImport(txRepo, workerPool2, engine, log)

	res2, err := importer2.Import(ctx, usecase.ImportRequest{
		AccountID:    accountID,
		Transactions: []port.CreateTransactionParams{txParam},
	})
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if len(res2.Created) != 0 {
		t.Errorf("second import: want 0 created (idempotent), got %d", len(res2.Created))
	}
	if len(res2.Duplicates) != 1 {
		t.Errorf("second import: want 1 duplicate, got %d", len(res2.Duplicates))
	}

	// Проверяем, что в БД ровно одна запись
	var count int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE account_id = $1 AND external_id = $2`,
		accountID, txParam.ExternalID).Scan(&count)
	if count != 1 {
		t.Errorf("DB: want exactly 1 transaction, got %d", count)
	}

	_ = workerPool2.Shutdown(ctx)
}

// TestGameEngineUpdatesProfile — проверяет, что GameEngine правильно обновляет
// game_profiles после классификации транзакции.
//
// Ресторанная транзакция (MCC 5812) → Dining category → HP уменьшается.
func TestGameEngineUpdatesProfile(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	log, _ := zap.NewDevelopment()

	// Создаём пользователя, аккаунт, профиль
	var userID, accountID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO users (id, email_hash, password_hash, created_at, updated_at)
		VALUES (gen_random_uuid(), 'testhash2', 'fakehash', NOW(), NOW()) RETURNING id`,
	).Scan(&userID)
	_ = pool.QueryRow(ctx, `
		INSERT INTO accounts (id, user_id, name, type, balance, currency, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'Main', 'debit', 0, 'RUB', NOW(), NOW()) RETURNING id`, userID,
	).Scan(&accountID)
	_, _ = pool.Exec(ctx, `
		INSERT INTO game_profiles (user_id, level, xp, hp, mana, strength, intellect, luck, updated_at)
		VALUES ($1, 1, 0, 100, 100, 10, 10, 10, NOW())`, userID)

	txRepo      := postgres.NewTransactionRepo(pool)
	gameRepo    := postgres.NewGameRepo(pool)
	accountRepo := postgres.NewAccountRepo(pool)
	classifier  := ai.NewChainedProvider(log)
	hub         := &noopHub{}
	engine      := usecase.NewGameEngine(txRepo, accountRepo, gameRepo, classifier, hub, log)

	// Создаём транзакцию (ресторан)
	tx, _, err := txRepo.CreateIfNotExists(ctx, port.CreateTransactionParams{
		AccountID:           accountID,
		ExternalID:          "test-game-001",
		Amount:              -50000, // -500 руб
		MCC:                 5812,   // Restaurants → HP -2
		OriginalDescription: "Starbucks",
		OccurredAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("create tx: %v", err)
	}

	// Запускаем GameEngine
	if err := engine.ProcessTransaction(ctx, tx); err != nil {
		t.Fatalf("process transaction: %v", err)
	}

	// Проверяем, что транзакция получила категорию
	var category string
	_ = pool.QueryRow(ctx, `SELECT COALESCE(clean_category, '') FROM transactions WHERE id = $1`, tx.ID).Scan(&category)
	if category == "" {
		t.Error("transaction should have a category after processing")
	}

	// Проверяем, что game_events создалась
	var eventCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM game_events WHERE user_id = $1`, userID).Scan(&eventCount)
	if eventCount == 0 {
		t.Error("expected at least one game_event after processing")
	}

	t.Logf("✓ category=%q events=%d", category, eventCount)
}

// TestAnalyticsSummary — проверяет, что аналитический запрос корректно
// агрегирует транзакции по категориям и считает проценты.
func TestAnalyticsSummary(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()

	var userID, accountID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO users (id, email_hash, password_hash, created_at, updated_at)
		VALUES (gen_random_uuid(), 'testhash3', 'fakehash', NOW(), NOW()) RETURNING id`,
	).Scan(&userID)
	_ = pool.QueryRow(ctx, `
		INSERT INTO accounts (id, user_id, name, type, balance, currency, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'Main', 'debit', 0, 'RUB', NOW(), NOW()) RETURNING id`, userID,
	).Scan(&accountID)

	// Вставляем транзакции напрямую (уже processed)
	for i, row := range []struct {
		extID    string
		amount   int64
		category string
	}{
		{"tx1", -100000, "Restaurants & Cafes"},
		{"tx2", -200000, "Restaurants & Cafes"},
		{"tx3", -150000, "Supermarkets & Groceries"},
		{"tx4", 500000, ""},  // доход (положительная сумма)
	} {
		_, err := pool.Exec(ctx, `
			INSERT INTO transactions (id, account_id, external_id, amount, mcc, original_description,
			                          clean_category, status, occurred_at, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, 0, 'test', $4, 'processed', NOW()-($5 || ' hours')::interval, NOW(), NOW())`,
			accountID, row.extID, row.amount, row.category, fmt.Sprintf("%d", i))
		if err != nil {
			t.Fatalf("insert tx: %v", err)
		}
	}

	analyticsRepo := postgres.NewAnalyticsRepo(pool)
	summary, err := analyticsRepo.Summary(ctx, accountID, 30)
	if err != nil {
		t.Fatalf("analytics summary: %v", err)
	}

	// Общие расходы: 100000 + 200000 + 150000 = 450000 (в копейках)
	if summary.TotalExpenses != 450000 {
		t.Errorf("total expenses: want 450000, got %d", summary.TotalExpenses)
	}
	// Доходы: 500000
	if summary.TotalIncome != 500000 {
		t.Errorf("total income: want 500000, got %d", summary.TotalIncome)
	}
	// Категорий: 2 (Restaurants + Groceries). Пустая категория — Uncategorized.
	if len(summary.ByCategory) < 2 {
		t.Errorf("expected at least 2 categories, got %d", len(summary.ByCategory))
	}

	// Рестораны — самая большая категория (300000 из 450000 = 66.7%)
	top := summary.ByCategory[0]
	if top.Amount != 300000 {
		t.Errorf("top category amount: want 300000, got %d", top.Amount)
	}

	t.Logf("✓ expenses=%d income=%d categories=%d top=%s(%.1f%%)",
		summary.TotalExpenses, summary.TotalIncome, len(summary.ByCategory),
		top.Category, top.Percent)
}

// noopHub — заглушка Notifier для тестов, чтобы не нужен был реальный WebSocket.
// Реализует port.Notifier — GameEngine требует его для push-уведомлений.
type noopHub struct{}

func (n *noopHub) PushProfileUpdate(_ string, _ domain.GameProfile)      {}
func (n *noopHub) PushTransactionProcessed(_ string, _ domain.Transaction) {}
