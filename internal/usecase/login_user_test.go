package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/infrastructure/auth"
	"github.com/hillscheck/internal/usecase"
)

// ── Mock TokenStore ────────────────────────────────────────────────────────────

type mockTokenStore struct {
	tokens map[string]string // token → userID
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{tokens: make(map[string]string)}
}

func (m *mockTokenStore) Save(_ context.Context, token, userID string, _ time.Duration) error {
	m.tokens[token] = userID
	return nil
}

func (m *mockTokenStore) Get(_ context.Context, token string) (string, error) {
	id, ok := m.tokens[token]
	if !ok {
		return "", domain.ErrTokenInvalid
	}
	return id, nil
}

func (m *mockTokenStore) Delete(_ context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func registerTestUser(t *testing.T, repo *mockUserRepo, email, password string) {
	t.Helper()
	uc := usecase.NewRegisterUser(repo, &mockGameRepo{})
	if _, err := uc.Execute(context.Background(), usecase.RegisterInput{
		Name: "Test User", Email: email, Password: password,
	}); err != nil {
		t.Fatalf("setup: register user: %v", err)
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestLoginUser_HappyPath(t *testing.T) {
	repo := newMockUserRepo()
	registerTestUser(t, repo, "grisha@example.com", "supersecret123")

	jwt := auth.NewJWTService("test-secret-key-32-bytes-minimum!")
	ts := newMockTokenStore()
	uc := usecase.NewLoginUser(repo, jwt, ts)

	out, err := uc.Execute(context.Background(), usecase.LoginInput{
		Email:    "grisha@example.com",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.AccessToken == "" {
		t.Error("access token must not be empty")
	}
	if out.RefreshToken == "" {
		t.Error("refresh token must not be empty")
	}
	if out.User.Email != "grisha@example.com" {
		t.Errorf("unexpected user email: %q", out.User.Email)
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	registerTestUser(t, repo, "grisha@example.com", "supersecret123")

	jwt := auth.NewJWTService("test-secret-key-32-bytes-minimum!")
	uc := usecase.NewLoginUser(repo, jwt, newMockTokenStore())

	_, err := uc.Execute(context.Background(), usecase.LoginInput{
		Email:    "grisha@example.com",
		Password: "wrongpassword",
	})
	if err != domain.ErrWrongPassword {
		t.Errorf("expected ErrWrongPassword, got: %v", err)
	}
}

func TestLoginUser_UnknownEmail(t *testing.T) {
	jwt := auth.NewJWTService("test-secret-key-32-bytes-minimum!")
	uc := usecase.NewLoginUser(newMockUserRepo(), jwt, newMockTokenStore())

	_, err := uc.Execute(context.Background(), usecase.LoginInput{
		Email:    "nobody@example.com",
		Password: "whatever",
	})
	if err != domain.ErrWrongPassword {
		// Must return the same generic error — no user enumeration
		t.Errorf("expected ErrWrongPassword (no user enumeration), got: %v", err)
	}
}
