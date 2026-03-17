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

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, p port.CreateUserParams) (domain.User, error) {
	const q = `
		INSERT INTO users (name, email_hash, settings)
		VALUES ($1, $2, $3)
		RETURNING id, name, email_hash, settings, created_at, updated_at`

	var u domain.User
	err := r.pool.QueryRow(ctx, q, p.Name, p.EmailHash, p.Settings).
		Scan(&u.ID, &u.Name, &u.EmailHash, &u.Settings, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (domain.User, error) {
	const q = `
		SELECT id, name, email_hash, settings, created_at, updated_at
		FROM users WHERE id = $1`

	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Name, &u.EmailHash, &u.Settings, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmailHash(ctx context.Context, emailHash string) (domain.User, error) {
	const q = `
		SELECT id, name, email_hash, settings, created_at, updated_at
		FROM users WHERE email_hash = $1`

	var u domain.User
	err := r.pool.QueryRow(ctx, q, emailHash).
		Scan(&u.ID, &u.Name, &u.EmailHash, &u.Settings, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email hash: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u domain.User) (domain.User, error) {
	const q = `
		UPDATE users SET name = $2, settings = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, email_hash, settings, created_at, updated_at`

	var updated domain.User
	err := r.pool.QueryRow(ctx, q, u.ID, u.Name, u.Settings).
		Scan(&updated.ID, &updated.Name, &updated.EmailHash, &updated.Settings, &updated.CreatedAt, &updated.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return updated, nil
}
