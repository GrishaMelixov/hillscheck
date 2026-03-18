package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

type RegisterUser struct {
	users    port.UserRepository
	game     port.GameRepository
	accounts port.AccountRepository
}

func NewRegisterUser(users port.UserRepository, game port.GameRepository, accounts port.AccountRepository) *RegisterUser {
	return &RegisterUser{users: users, game: game, accounts: accounts}
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

	// Create a default debit account so the user can import transactions immediately
	if _, err := uc.accounts.Create(ctx, port.CreateAccountParams{
		UserID:   user.ID,
		Name:     "Основной счёт",
		Type:     domain.AccountTypeDebit,
		Currency: "RUB",
	}); err != nil {
		return RegisterOutput{}, fmt.Errorf("create default account: %w", err)
	}

	return RegisterOutput{User: user}, nil
}
