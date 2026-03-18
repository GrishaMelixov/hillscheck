package usecase_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

func newTestImporter(txRepo *mockTxRepo, poolFull bool) *usecase.TransactionImport {
	log, _ := zap.NewDevelopment()
	accountRepo := newMockAccountRepo(domain.Account{ID: "acc-1", UserID: "user-1"})
	engine := usecase.NewGameEngine(
		txRepo, accountRepo, newMockGameRepo(),
		&mockClassifier{result: port.ClassificationResult{
			Category: "General Purchase",
			Impact:   port.Impact{Attribute: port.AttrXP, Value: 1},
		}},
		&mockNotifier{}, log,
	)
	return usecase.NewTransactionImport(txRepo, &mockWorkerPool{full: poolFull}, engine, log)
}

// TestTransactionImport_CreatesNewTransactions — два новых external_id → два created.
func TestTransactionImport_CreatesNewTransactions(t *testing.T) {
	importer := newTestImporter(newMockTxRepo(), false)

	res, err := importer.Import(context.Background(), usecase.ImportRequest{
		AccountID: "acc-1",
		Transactions: []port.CreateTransactionParams{
			{ExternalID: "tink-001", Amount: -50000, MCC: 5411, OccurredAt: time.Now()},
			{ExternalID: "tink-002", Amount: -30000, MCC: 5812, OccurredAt: time.Now()},
		},
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(res.Created) != 2 {
		t.Errorf("created: want 2, got %d", len(res.Created))
	}
	if len(res.Duplicates) != 0 {
		t.Errorf("duplicates: want 0, got %d", len(res.Duplicates))
	}
}

// TestTransactionImport_Idempotency — ключевой инвариант:
// повторный импорт того же external_id → дубликат, новой записи нет.
//
// "0% потерь данных" означает в том числе отсутствие дублирования.
func TestTransactionImport_Idempotency(t *testing.T) {
	txRepo := newMockTxRepo()
	importer := newTestImporter(txRepo, false)

	req := usecase.ImportRequest{
		AccountID: "acc-1",
		Transactions: []port.CreateTransactionParams{
			{ExternalID: "tink-001", Amount: -50000, MCC: 5411, OccurredAt: time.Now()},
		},
	}

	// Первый импорт
	res1, _ := importer.Import(context.Background(), req)
	if len(res1.Created) != 1 {
		t.Fatalf("first import: want 1 created, got %d", len(res1.Created))
	}

	// Повторный импорт того же CSV
	res2, _ := importer.Import(context.Background(), req)
	if len(res2.Created) != 0 {
		t.Errorf("second import: want 0 created (idempotent), got %d", len(res2.Created))
	}
	if len(res2.Duplicates) != 1 {
		t.Errorf("second import: want 1 duplicate, got %d", len(res2.Duplicates))
	}
}

// TestTransactionImport_ZeroAmountRejected — нулевая сумма невалидна.
func TestTransactionImport_ZeroAmountRejected(t *testing.T) {
	_, err := newTestImporter(newMockTxRepo(), false).Import(context.Background(), usecase.ImportRequest{
		AccountID: "acc-1",
		Transactions: []port.CreateTransactionParams{
			{ExternalID: "bad", Amount: 0, OccurredAt: time.Now()},
		},
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

// TestTransactionImport_EmptyAccountID — пустой account_id отклоняется немедленно.
func TestTransactionImport_EmptyAccountID(t *testing.T) {
	_, err := newTestImporter(newMockTxRepo(), false).Import(context.Background(), usecase.ImportRequest{
		AccountID:    "",
		Transactions: []port.CreateTransactionParams{{ExternalID: "x", Amount: -100}},
	})
	if err == nil {
		t.Fatal("expected error for empty account_id")
	}
}

// TestTransactionImport_PoolFull — когда пул переполнен, транзакция создаётся
// но помечается failed для последующего retry. Импорт не падает.
func TestTransactionImport_PoolFull(t *testing.T) {
	txRepo := newMockTxRepo()
	importer := newTestImporter(txRepo, true) // pool full

	res, err := importer.Import(context.Background(), usecase.ImportRequest{
		AccountID: "acc-1",
		Transactions: []port.CreateTransactionParams{
			{ExternalID: "pool-test", Amount: -100, OccurredAt: time.Now()},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Транзакция создана (не потеряна), но обработка отложена
	if len(res.Created) != 1 {
		t.Errorf("want 1 created even with full pool, got %d", len(res.Created))
	}
}
