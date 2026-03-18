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

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

const userCols = `id, name, email, email_hash, password_hash, plan, settings, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.EmailHash,
		&u.PasswordHash, &u.Plan, &u.Settings,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func (r *UserRepo) Create(ctx context.Context, p port.CreateUserParams) (domain.User, error) {
	const q = `
		INSERT INTO users (name, email, email_hash, password_hash, settings)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + userCols

	u, err := scanUser(r.pool.QueryRow(ctx, q,
		p.Name, p.Email, p.EmailHash, p.PasswordHash, p.Settings,
	))
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (domain.User, error) {
	const q = `SELECT ` + userCols + ` FROM users WHERE id = $1`

	u, err := scanUser(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `SELECT ` + userCols + ` FROM users WHERE email = $1`

	u, err := scanUser(r.pool.QueryRow(ctx, q, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmailHash(ctx context.Context, emailHash string) (domain.User, error) {
	const q = `SELECT ` + userCols + ` FROM users WHERE email_hash = $1`

	u, err := scanUser(r.pool.QueryRow(ctx, q, emailHash))
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
		RETURNING ` + userCols

	updated, err := scanUser(r.pool.QueryRow(ctx, q, u.ID, u.Name, u.Settings))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return updated, nil
}
