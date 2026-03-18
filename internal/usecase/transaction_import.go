package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

type ImportRequest struct {
	AccountID    string
	Transactions []port.CreateTransactionParams
}

type ImportResult struct {
	Created    []domain.Transaction
	Duplicates []domain.Transaction
}

type TransactionImport struct {
	txRepo port.TransactionRepository
	pool   port.WorkerPool
	engine *GameEngine
	log    *zap.Logger
}

func NewTransactionImport(
	txRepo port.TransactionRepository,
	pool port.WorkerPool,
	engine *GameEngine,
	log *zap.Logger,
) *TransactionImport {
	return &TransactionImport{
		txRepo: txRepo,
		pool:   pool,
		engine: engine,
		log:    log,
	}
}

// Import validates and persists a batch of transactions.
// New transactions are enqueued for async enrichment.
// Duplicate external IDs are returned without re-processing (idempotency).
func (u *TransactionImport) Import(ctx context.Context, req ImportRequest) (ImportResult, error) {
	if req.AccountID == "" {
		return ImportResult{}, fmt.Errorf("account_id is required")
	}

	var res ImportResult

	for _, p := range req.Transactions {
		if p.Amount == 0 {
			return ImportResult{}, domain.ErrInvalidAmount
		}
		p.AccountID = req.AccountID

		tx, existed, err := u.txRepo.CreateIfNotExists(ctx, p)
		if err != nil {
			return ImportResult{}, fmt.Errorf("persist transaction %q: %w", p.ExternalID, err)
		}

		if existed {
			res.Duplicates = append(res.Duplicates, tx)
			continue
		}

		res.Created = append(res.Created, tx)

		// Capture for closure — avoids loop variable aliasing.
		captured := tx
		job := func(jCtx context.Context) error {
			return u.engine.ProcessTransaction(jCtx, captured)
		}

		if err := u.pool.Submit(job); err != nil {
			// Pool is full: mark as failed so it can be retried later.
			_ = u.txRepo.UpdateStatus(ctx, captured.ID, domain.TxStatusFailed, "")
			u.log.Warn("worker pool full, transaction marked failed",
				zap.String("tx_id", captured.ID),
				zap.Error(err),
			)
		}
	}

	u.log.Info("import complete",
		zap.Int("created", len(res.Created)),
		zap.Int("duplicates", len(res.Duplicates)),
	)
	return res, nil
}
