package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/usecase/port"
)

type RegisterUser struct {
	users port.UserRepository
	game  port.GameRepository
}

func NewRegisterUser(users port.UserRepository, game port.GameRepository) *RegisterUser {
	return &RegisterUser{users: users, game: game}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterOutput struct {
	User domain.User
}

func (uc *RegisterUser) Execute(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	// Check duplicate
	_, err := uc.users.GetByEmail(ctx, email)
	if err == nil {
		return RegisterOutput{}, domain.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("hash password: %w", err)
	}

	emailHash := fmt.Sprintf("%x", sha256.Sum256([]byte(email)))

	user, err := uc.users.Create(ctx, port.CreateUserParams{
		Name:         strings.TrimSpace(in.Name),
		Email:        email,
		EmailHash:    emailHash,
		PasswordHash: string(hash),
		Settings:     domain.UserSettings{Currency: "RUB", Timezone: "Europe/Moscow"},
	})
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("create user: %w", err)
	}

	// Bootstrap game profile
	if _, err := uc.game.CreateProfile(ctx, user.ID); err != nil {
		return RegisterOutput{}, fmt.Errorf("create game profile: %w", err)
	}

	return RegisterOutput{User: user}, nil
}
