package port

import (
	"context"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
)

type CreateAccountParams struct {
	UserID   string
	Name     string
	Type     domain.AccountType
	Currency string
}

type AccountRepository interface {
	Create(ctx context.Context, p CreateAccountParams) (domain.Account, error)
	GetByID(ctx context.Context, id string) (domain.Account, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Account, error)
	UpdateBalance(ctx context.Context, id string, delta int64) (domain.Account, error)
}
