package usecase

import (
	"context"

	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

type Logout struct {
	tokenStore port.TokenStore
}

func NewLogout(ts port.TokenStore) *Logout {
	return &Logout{tokenStore: ts}
}

func (uc *Logout) Execute(ctx context.Context, refreshToken string) error {
	return uc.tokenStore.Delete(ctx, refreshToken)
}
