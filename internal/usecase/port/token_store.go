package port

import (
	"context"
	"time"
)

// TokenStore persists refresh tokens (UUID → userID) with a TTL.
// Backed by Redis — allows instant revocation on logout.
type TokenStore interface {
	Save(ctx context.Context, token, userID string, ttl time.Duration) error
	Get(ctx context.Context, token string) (userID string, err error)
	Delete(ctx context.Context, token string) error
}
