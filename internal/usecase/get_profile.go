package usecase

import (
	"context"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/usecase/port"
)

type GetProfile struct {
	gameRepo port.GameRepository
}

func NewGetProfile(gameRepo port.GameRepository) *GetProfile {
	return &GetProfile{gameRepo: gameRepo}
}

func (u *GetProfile) Execute(ctx context.Context, userID string) (domain.GameProfile, error) {
	return u.gameRepo.GetProfile(ctx, userID)
}
