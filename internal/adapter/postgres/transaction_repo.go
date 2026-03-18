package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/usecase/port"
)

type TransactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

// CreateIfNotExists performs INSERT ... ON CONFLICT DO NOTHING.
// Returns (transaction, alreadyExisted, error).
func (r *TransactionRepo) CreateIfNotExists(
	ctx context.Context,
	p port.CreateTransactionParams,
) (domain.Transaction, bool, error) {
	const q = `
		INSERT INTO transactions
			(external_id, account_id, amount, mcc, original_description, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (account_id, external_id) DO NOTHING
		RETURNING id, external_id, account_id, amount, mcc,
		          original_description, COALESCE(clean_category, ''), status, occurred_at, created_at, updated_at`

	var tx domain.Transaction
	err := r.pool.QueryRow(ctx, q,
		p.ExternalID, p.AccountID, p.Amount, p.MCC, p.OriginalDescription, p.OccurredAt,
	).Scan(
		&tx.ID, &tx.ExternalID, &tx.AccountID, &tx.Amount, &tx.MCC,
		&tx.OriginalDescription, &tx.CleanCategory, &tx.Status,
		&tx.OccurredAt, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		existing, fetchErr := r.getByExternalID(ctx, p.AccountID, p.ExternalID)
		return existing, true, fetchErr
	}
	if err != nil {
		return domain.Transaction{}, false, fmt.Errorf("create transaction: %w", err)
	}
	return tx, false, nil
}

func (r *TransactionRepo) getByExternalID(ctx context.Context, accountID, externalID string) (domain.Transaction, error) {
	const q = `
		SELECT id, external_id, account_id, amount, mcc,
		       original_description, COALESCE(clean_category, ''), status, occurred_at, created_at, updated_at
		FROM transactions WHERE account_id = $1 AND external_id = $2`

	var tx domain.Transaction
	err := r.pool.QueryRow(ctx, q, accountID, externalID).Scan(
		&tx.ID, &tx.ExternalID, &tx.AccountID, &tx.Amount, &tx.MCC,
		&tx.OriginalDescription, &tx.CleanCategory, &tx.Status,
		&tx.OccurredAt, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transaction{}, domain.ErrNotFound
	}
	return tx, err
}

func (r *TransactionRepo) GetByID(ctx context.Context, id string) (domain.Transaction, error) {
	const q = `
		SELECT id, external_id, account_id, amount, mcc,
		       original_description, COALESCE(clean_category, ''), status, occurred_at, created_at, updated_at
		FROM transactions WHERE id = $1`

	var tx domain.Transaction
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&tx.ID, &tx.ExternalID, &tx.AccountID, &tx.Amount, &tx.MCC,
		&tx.OriginalDescription, &tx.CleanCategory, &tx.Status,
		&tx.OccurredAt, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transaction{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("get transaction: %w", err)
	}
	return tx, nil
}

func (r *TransactionRepo) UpdateStatus(ctx context.Context, id string, status domain.TxStatus, category string) error {
	const q = `
		UPDATE transactions
		SET status = $2, clean_category = $3, updated_at = now()
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id, status, category)
	if err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepo) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]domain.Transaction, error) {
	const q = `
		SELECT id, external_id, account_id, amount, mcc,
		       original_description, COALESCE(clean_category, ''), status, occurred_at, created_at, updated_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, q, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.ExternalID, &tx.AccountID, &tx.Amount, &tx.MCC,
			&tx.OriginalDescription, &tx.CleanCategory, &tx.Status,
			&tx.OccurredAt, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}
