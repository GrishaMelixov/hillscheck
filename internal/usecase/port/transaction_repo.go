package port

import (
	"context"
	"time"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
)

type CreateTransactionParams struct {
	AccountID           string
	ExternalID          string // idempotency key from the client
	Amount              int64  // cents
	MCC                 int
	OriginalDescription string
	OccurredAt          time.Time
}

type TransactionRepository interface {
	// CreateIfNotExists performs INSERT ... ON CONFLICT DO NOTHING.
	// Returns (transaction, alreadyExisted, error).
	CreateIfNotExists(ctx context.Context, p CreateTransactionParams) (domain.Transaction, bool, error)

	GetByID(ctx context.Context, id string) (domain.Transaction, error)
	UpdateStatus(ctx context.Context, id string, status domain.TxStatus, category string) error
	ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]domain.Transaction, error)
}
