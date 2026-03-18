package usecase_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

func newEngine(txRepo *mockTxRepo, accountRepo *mockAccountRepo, gameRepo *mockGameRepo,
	classifier *mockClassifier, notifier *mockNotifier) *usecase.GameEngine {
	log, _ := zap.NewDevelopment()
	return usecase.NewGameEngine(txRepo, accountRepo, gameRepo, classifier, notifier, log)
}

// TestGameEngine_ProcessTransaction_HappyPath проверяет полный цикл:
// классификация → обновление статуса → применение RPG-события → WebSocket-пуш.
func TestGameEngine_ProcessTransaction_HappyPath(t *testing.T) {
	txRepo := newMockTxRepo()
	txRepo.txs["ext-1"] = &domain.Transaction{ID: "tx-1", AccountID: "acc-1"}

	accountRepo := newMockAccountRepo(domain.Account{ID: "acc-1", UserID: "user-1"})
	gameRepo := newMockGameRepo()
	classifier := &mockClassifier{result: port.ClassificationResult{
		Category: "Restaurants & Cafes",
		Impact:   port.Impact{Attribute: port.AttrHP, Value: -2},
	}}
	notifier := &mockNotifier{}

	engine := newEngine(txRepo, accountRepo, gameRepo, classifier, notifier)
	tx := domain.Transaction{ID: "tx-1", AccountID: "acc-1", MCC: 5812}

	if err := engine.ProcessTransaction(context.Background(), tx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Транзакция помечена как processed
	if txRepo.status["tx-1"] != domain.TxStatusProcessed {
		t.Errorf("status: want processed, got %s", txRepo.status["tx-1"])
	}

	// WebSocket-пуш отправлен
	if len(notifier.profileUpdates) != 1 {
		t.Errorf("profile updates: want 1, got %d", len(notifier.profileUpdates))
	}

	// HP уменьшился на 2 (ресторан = -2 HP)
	if gameRepo.profiles["user-1"].HP != 98 {
		t.Errorf("HP: want 98, got %d", gameRepo.profiles["user-1"].HP)
	}
}

// TestGameEngine_ClassifierError_MarksTransactionFailed — если LLM падает,
// транзакция помечается failed (не теряется — доступна для повтора).
func TestGameEngine_ClassifierError_MarksTransactionFailed(t *testing.T) {
	txRepo := newMockTxRepo()
	txRepo.txs["ext-2"] = &domain.Transaction{ID: "tx-2", AccountID: "acc-1"}

	accountRepo := newMockAccountRepo(domain.Account{ID: "acc-1", UserID: "user-1"})
	gameRepo := newMockGameRepo()
	classifier := &mockClassifier{err: errors.New("LLM timeout")}
	notifier := &mockNotifier{}

	engine := newEngine(txRepo, accountRepo, gameRepo, classifier, notifier)

	err := engine.ProcessTransaction(context.Background(), domain.Transaction{ID: "tx-2", AccountID: "acc-1"})
	if err == nil {
		t.Fatal("expected error when classifier fails")
	}

	if txRepo.status["tx-2"] != domain.TxStatusFailed {
		t.Errorf("status: want failed, got %s", txRepo.status["tx-2"])
	}
}

// TestGameEngine_StrengthBonus — спортивная транзакция увеличивает Strength.
func TestGameEngine_StrengthBonus(t *testing.T) {
	txRepo := newMockTxRepo()
	txRepo.txs["gym"] = &domain.Transaction{ID: "tx-gym", AccountID: "acc-1"}

	accountRepo := newMockAccountRepo(domain.Account{ID: "acc-1", UserID: "user-1"})
	gameRepo := newMockGameRepo()
	classifier := &mockClassifier{result: port.ClassificationResult{
		Category: "Sports & Fitness",
		Impact:   port.Impact{Attribute: port.AttrStrength, Value: 5},
	}}
	notifier := &mockNotifier{}

	engine := newEngine(txRepo, accountRepo, gameRepo, classifier, notifier)
	if err := engine.ProcessTransaction(context.Background(), domain.Transaction{ID: "tx-gym", AccountID: "acc-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gameRepo.profiles["user-1"].Strength != 15 { // 10 base + 5
		t.Errorf("Strength: want 15, got %d", gameRepo.profiles["user-1"].Strength)
	}
}

// TestGameEngine_AccountNotFound — если аккаунт не существует, engine возвращает ошибку.
func TestGameEngine_AccountNotFound(t *testing.T) {
	log, _ := zap.NewDevelopment()
	engine := usecase.NewGameEngine(
		newMockTxRepo(),
		newMockAccountRepo(), // пустой — аккаунт не найден
		newMockGameRepo(),
		&mockClassifier{},
		&mockNotifier{},
		log,
	)
	if err := engine.ProcessTransaction(context.Background(), domain.Transaction{ID: "tx-x", AccountID: "nonexistent"}); err == nil {
		t.Fatal("expected error for missing account")
	}
}
