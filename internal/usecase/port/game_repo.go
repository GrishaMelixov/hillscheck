package port

import (
	"context"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
)

type GameRepository interface {
	CreateProfile(ctx context.Context, userID string) (domain.GameProfile, error)
	GetProfile(ctx context.Context, userID string) (domain.GameProfile, error)

	// ApplyEvent atomically writes the GameEvent and updates game_profiles
	// inside a single DB transaction using FOR UPDATE on the profile row.
	ApplyEvent(ctx context.Context, event domain.GameEvent) (domain.GameProfile, error)

	ListEvents(ctx context.Context, userID string, limit int) ([]domain.GameEvent, error)
}
