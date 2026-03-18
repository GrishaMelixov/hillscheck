package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/GrishaMelixov/wealthcheck/internal/infrastructure/auth"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

type RefreshToken struct {
	users      port.UserRepository
	jwt        *auth.JWTService
	tokenStore port.TokenStore
}

func NewRefreshToken(users port.UserRepository, jwt *auth.JWTService, ts port.TokenStore) *RefreshToken {
	return &RefreshToken{users: users, jwt: jwt, tokenStore: ts}
}

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

func (uc *RefreshToken) Execute(ctx context.Context, oldRefresh string) (RefreshOutput, error) {
	userID, err := uc.tokenStore.Get(ctx, oldRefresh)
	if err != nil {
		return RefreshOutput{}, err // ErrTokenInvalid
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return RefreshOutput{}, err
	}

	// Rotate: delete old, issue new
	_ = uc.tokenStore.Delete(ctx, oldRefresh)

	access, err := uc.jwt.IssueAccess(user.ID, user.Plan)
	if err != nil {
		return RefreshOutput{}, fmt.Errorf("issue access token: %w", err)
	}

	newRefresh, err := newRefreshToken()
	if err != nil {
		return RefreshOutput{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := uc.tokenStore.Save(ctx, newRefresh, user.ID, auth.RefreshTokenTTL); err != nil {
		return RefreshOutput{}, fmt.Errorf("save refresh token: %w", err)
	}

	return RefreshOutput{AccessToken: access, RefreshToken: newRefresh}, nil
}

// newRefreshToken generates a cryptographically secure random token.
func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
