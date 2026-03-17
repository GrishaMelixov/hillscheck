package port

import (
	"context"

	"github.com/hillscheck/internal/domain"
)

type CreateUserParams struct {
	Name      string
	EmailHash string
	Settings  domain.UserSettings
}

type UserRepository interface {
	Create(ctx context.Context, p CreateUserParams) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmailHash(ctx context.Context, emailHash string) (domain.User, error)
	Update(ctx context.Context, user domain.User) (domain.User, error)
}
