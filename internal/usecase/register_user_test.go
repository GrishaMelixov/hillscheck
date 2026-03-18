package usecase_test

import (
	"context"
	"testing"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
)

func TestRegisterUser_HappyPath(t *testing.T) {
	uc := usecase.NewRegisterUser(newMockUserRepo(), &mockGameRepo{}, &mockAccountRepo{})

	out, err := uc.Execute(context.Background(), usecase.RegisterInput{
		Name:     "Grisha",
		Email:    "grisha@example.com",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.User.Email != "grisha@example.com" {
		t.Errorf("expected email grisha@example.com, got %q", out.User.Email)
	}
	if out.User.Plan != domain.PlanFree {
		t.Errorf("new user should be on free plan, got %q", out.User.Plan)
	}
	if out.User.PasswordHash == "" {
		t.Error("password hash must not be empty")
	}
	if out.User.PasswordHash == "supersecret123" {
		t.Error("password must be hashed, not stored in plaintext")
	}
}

func TestRegisterUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewRegisterUser(repo, &mockGameRepo{}, &mockAccountRepo{})

	input := usecase.RegisterInput{Name: "Grisha", Email: "grisha@example.com", Password: "supersecret123"}

	if _, err := uc.Execute(context.Background(), input); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err := uc.Execute(context.Background(), input)
	if err != domain.ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got: %v", err)
	}
}

func TestRegisterUser_EmailNormalized(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewRegisterUser(repo, &mockGameRepo{}, &mockAccountRepo{})

	out, err := uc.Execute(context.Background(), usecase.RegisterInput{
		Name:     "Grisha",
		Email:    "  GRISHA@Example.COM  ",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.Email != "grisha@example.com" {
		t.Errorf("email should be lowercased and trimmed, got %q", out.User.Email)
	}
}
