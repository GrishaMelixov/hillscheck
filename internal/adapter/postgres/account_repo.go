package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

type AccountRepo struct {
	pool *pgxpool.Pool
}

func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{pool: pool}
}

func (r *AccountRepo) Create(ctx context.Context, p port.CreateAccountParams) (domain.Account, error) {
	const q = `
		INSERT INTO accounts (user_id, name, type, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, type, balance, currency, created_at, updated_at`

	var a domain.Account
	err := r.pool.QueryRow(ctx, q, p.UserID, p.Name, p.Type, p.Currency).
		Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return domain.Account{}, fmt.Errorf("create account: %w", err)
	}
	return a, nil
}

func (r *AccountRepo) GetByID(ctx context.Context, id string) (domain.Account, error) {
	const q = `
		SELECT id, user_id, name, type, balance, currency, created_at, updated_at
		FROM accounts WHERE id = $1`

	var a domain.Account
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("get account: %w", err)
	}
	return a, nil
}

func (r *AccountRepo) ListByUser(ctx context.Context, userID string) ([]domain.Account, error) {
	const q = `
		SELECT id, user_id, name, type, balance, currency, created_at, updated_at
		FROM accounts WHERE user_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (r *AccountRepo) UpdateBalance(ctx context.Context, id string, delta int64) (domain.Account, error) {
	const q = `
		UPDATE accounts SET balance = balance + $2, updated_at = now()
		WHERE id = $1
		RETURNING id, user_id, name, type, balance, currency, created_at, updated_at`

	var a domain.Account
	err := r.pool.QueryRow(ctx, q, id, delta).
		Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("update account balance: %w", err)
	}
	return a, nil
}
