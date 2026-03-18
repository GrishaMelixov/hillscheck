package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// GameEngine processes a single transaction through the enrichment pipeline:
// classify → persist category → apply game event → push real-time update.
type GameEngine struct {
	txRepo       port.TransactionRepository
	accountRepo  port.AccountRepository
	gameRepo     port.GameRepository
	classifier   port.CategoryProvider
	notifier     port.Notifier
	log          *zap.Logger
}

func NewGameEngine(
	txRepo port.TransactionRepository,
	accountRepo port.AccountRepository,
	gameRepo port.GameRepository,
	classifier port.CategoryProvider,
	notifier port.Notifier,
	log *zap.Logger,
) *GameEngine {
	return &GameEngine{
		txRepo:      txRepo,
		accountRepo: accountRepo,
		gameRepo:    gameRepo,
		classifier:  classifier,
		notifier:    notifier,
		log:         log,
	}
}

// ProcessTransaction is meant to be called inside a worker pool job.
// It enriches the transaction and applies the resulting RPG effect.
func (e *GameEngine) ProcessTransaction(ctx context.Context, tx domain.Transaction) error {
	// 1. Resolve user ID via the account.
	account, err := e.accountRepo.GetByID(ctx, tx.AccountID)
	if err != nil {
		return fmt.Errorf("get account for tx %s: %w", tx.ID, err)
	}

	// 2. Classify the transaction.
	result, err := e.classifier.Classify(ctx, tx.OriginalDescription, tx.MCC)
	if err != nil {
		_ = e.txRepo.UpdateStatus(ctx, tx.ID, domain.TxStatusFailed, "")
		return fmt.Errorf("classify tx %s: %w", tx.ID, err)
	}

	// 3. Persist the clean category and mark processed.
	if err := e.txRepo.UpdateStatus(ctx, tx.ID, domain.TxStatusProcessed, result.Category); err != nil {
		return fmt.Errorf("update tx status %s: %w", tx.ID, err)
	}
	tx.CleanCategory = result.Category
	tx.Status = domain.TxStatusProcessed

	// 4. Build and apply the game event.
	event := domain.GameEvent{
		TransactionID: tx.ID,
		UserID:        account.UserID,
		Attribute:     result.Impact.Attribute,
		Delta:         result.Impact.Value,
		Reason:        result.Category,
	}

	updatedProfile, err := e.gameRepo.ApplyEvent(ctx, event)
	if err != nil {
		e.log.Error("apply game event failed", zap.String("tx_id", tx.ID), zap.Error(err))
		return fmt.Errorf("apply game event for tx %s: %w", tx.ID, err)
	}

	// 5. Push real-time updates to connected clients.
	e.notifier.PushProfileUpdate(account.UserID, updatedProfile)
	e.notifier.PushTransactionProcessed(account.UserID, tx)

	e.log.Info("transaction processed",
		zap.String("tx_id", tx.ID),
		zap.String("category", result.Category),
		zap.String("attribute", result.Impact.Attribute),
		zap.Int("delta", result.Impact.Value),
	)
	return nil
}
