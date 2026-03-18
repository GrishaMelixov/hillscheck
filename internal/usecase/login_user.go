package usecase

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/infrastructure/auth"
	"github.com/hillscheck/internal/usecase/port"
)

type LoginUser struct {
	users      port.UserRepository
	jwt        *auth.JWTService
	tokenStore port.TokenStore
}

func NewLoginUser(users port.UserRepository, jwt *auth.JWTService, ts port.TokenStore) *LoginUser {
	return &LoginUser{users: users, jwt: jwt, tokenStore: ts}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	User         domain.User
}

func (uc *LoginUser) Execute(ctx context.Context, in LoginInput) (LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		return LoginOutput{}, domain.ErrWrongPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return LoginOutput{}, domain.ErrWrongPassword
	}

	access, err := uc.jwt.IssueAccess(user.ID, user.Plan)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("issue access token: %w", err)
	}

	refresh, err := newRefreshToken()
	if err != nil {
		return LoginOutput{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := uc.tokenStore.Save(ctx, refresh, user.ID, auth.RefreshTokenTTL); err != nil {
		return LoginOutput{}, fmt.Errorf("save refresh token: %w", err)
	}

	return LoginOutput{AccessToken: access, RefreshToken: refresh, User: user}, nil
}
