package usecase_test

import (
	"context"
	"testing"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/usecase"
	"github.com/hillscheck/internal/usecase/port"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockUserRepo struct {
	users  map[string]domain.User // keyed by email
	byID   map[string]domain.User
	nextID string
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]domain.User),
		byID:   make(map[string]domain.User),
		nextID: "user-1",
	}
}

func (m *mockUserRepo) Create(_ context.Context, p port.CreateUserParams) (domain.User, error) {
	if _, exists := m.users[p.Email]; exists {
		return domain.User{}, domain.ErrEmailTaken
	}
	u := domain.User{
		ID:           m.nextID,
		Name:         p.Name,
		Email:        p.Email,
		EmailHash:    p.EmailHash,
		PasswordHash: p.PasswordHash,
		Plan:         domain.PlanFree,
		Settings:     p.Settings,
	}
	m.users[p.Email] = u
	m.byID[u.ID] = u
	return u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (domain.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmailHash(_ context.Context, _ string) (domain.User, error) {
	return domain.User{}, domain.ErrNotFound
}

func (m *mockUserRepo) Update(_ context.Context, u domain.User) (domain.User, error) {
	m.users[u.Email] = u
	return u, nil
}

type mockGameRepo struct{}

func (m *mockGameRepo) CreateProfile(_ context.Context, _ string) (domain.GameProfile, error) {
	return domain.GameProfile{Level: 1, HP: 100}, nil
}

func (m *mockGameRepo) GetProfile(_ context.Context, _ string) (domain.GameProfile, error) {
	return domain.GameProfile{}, nil
}

func (m *mockGameRepo) ApplyEvent(_ context.Context, _ domain.GameEvent) (domain.GameProfile, error) {
	return domain.GameProfile{}, nil
}

func (m *mockGameRepo) ListEvents(_ context.Context, _ string, _ int) ([]domain.GameEvent, error) {
	return nil, nil
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestRegisterUser_HappyPath(t *testing.T) {
	uc := usecase.NewRegisterUser(newMockUserRepo(), &mockGameRepo{})

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
	// Make sure the raw password is NOT stored
	if out.User.PasswordHash == "supersecret123" {
		t.Error("password must be hashed, not stored in plaintext")
	}
}

func TestRegisterUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewRegisterUser(repo, &mockGameRepo{})

	input := usecase.RegisterInput{
		Name:     "Grisha",
		Email:    "grisha@example.com",
		Password: "supersecret123",
	}

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
	uc := usecase.NewRegisterUser(repo, &mockGameRepo{})

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
