package usecase_test

// Файл содержит все тест-моки для пакета usecase_test.
// Централизованные моки позволяют избежать дублирования между test-файлами.

import (
	"context"
	"errors"
	"time"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// ─── mockUserRepo ─────────────────────────────────────────────────────────────

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

// ─── mockAccountRepo ──────────────────────────────────────────────────────────
// Поддерживает хранение аккаунтов в памяти для настройки в тестах game engine.

type mockAccountRepo struct {
	accounts map[string]domain.Account // keyed by account ID
}

func newMockAccountRepo(accounts ...domain.Account) *mockAccountRepo {
	m := &mockAccountRepo{accounts: make(map[string]domain.Account)}
	for _, a := range accounts {
		m.accounts[a.ID] = a
	}
	return m
}

func (m *mockAccountRepo) Create(_ context.Context, _ port.CreateAccountParams) (domain.Account, error) {
	return domain.Account{ID: "acc-1"}, nil
}

func (m *mockAccountRepo) GetByID(_ context.Context, id string) (domain.Account, error) {
	a, ok := m.accounts[id]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

func (m *mockAccountRepo) ListByUser(_ context.Context, userID string) ([]domain.Account, error) {
	var res []domain.Account
	for _, a := range m.accounts {
		if a.UserID == userID {
			res = append(res, a)
		}
	}
	if len(res) == 0 {
		return []domain.Account{{ID: "acc-1"}}, nil
	}
	return res, nil
}

func (m *mockAccountRepo) UpdateBalance(_ context.Context, _ string, _ int64) (domain.Account, error) {
	return domain.Account{}, nil
}

// ─── mockGameRepo ─────────────────────────────────────────────────────────────
// Хранит профили и события в памяти.

type mockGameRepo struct {
	profiles map[string]*domain.GameProfile
	events   []domain.GameEvent
}

func newMockGameRepo() *mockGameRepo {
	return &mockGameRepo{profiles: make(map[string]*domain.GameProfile)}
}

func (m *mockGameRepo) CreateProfile(_ context.Context, userID string) (domain.GameProfile, error) {
	if m.profiles == nil {
		m.profiles = make(map[string]*domain.GameProfile)
	}
	p := domain.GameProfile{UserID: userID, Level: 1, HP: 100, Mana: 100, Strength: 10, Intellect: 10, Luck: 10}
	m.profiles[userID] = &p
	return p, nil
}

func (m *mockGameRepo) GetProfile(_ context.Context, userID string) (domain.GameProfile, error) {
	if p, ok := m.profiles[userID]; ok {
		return *p, nil
	}
	return domain.GameProfile{}, domain.ErrNotFound
}

func (m *mockGameRepo) ApplyEvent(_ context.Context, e domain.GameEvent) (domain.GameProfile, error) {
	p, ok := m.profiles[e.UserID]
	if !ok {
		p = &domain.GameProfile{UserID: e.UserID, HP: 100, Mana: 100, Strength: 10, Intellect: 10, Luck: 10}
		m.profiles[e.UserID] = p
	}
	m.events = append(m.events, e)
	switch e.Attribute {
	case port.AttrHP:
		p.HP += e.Delta
	case port.AttrMana:
		p.Mana += e.Delta
	case port.AttrStrength:
		p.Strength += e.Delta
	case port.AttrIntellect:
		p.Intellect += e.Delta
	case port.AttrLuck:
		p.Luck += e.Delta
	case port.AttrXP:
		p.XP += int64(e.Delta)
	}
	return *p, nil
}

func (m *mockGameRepo) ListEvents(_ context.Context, _ string, _ int) ([]domain.GameEvent, error) {
	return m.events, nil
}

// ─── mockTxRepo ───────────────────────────────────────────────────────────────

type mockTxRepo struct {
	txs    map[string]*domain.Transaction // keyed by external_id
	status map[string]domain.TxStatus     // keyed by tx ID
}

func newMockTxRepo() *mockTxRepo {
	return &mockTxRepo{
		txs:    make(map[string]*domain.Transaction),
		status: make(map[string]domain.TxStatus),
	}
}

func (m *mockTxRepo) CreateIfNotExists(_ context.Context, p port.CreateTransactionParams) (domain.Transaction, bool, error) {
	if t, ok := m.txs[p.ExternalID]; ok {
		return *t, true, nil
	}
	tx := domain.Transaction{
		ID:                  "tx-" + p.ExternalID,
		AccountID:           p.AccountID,
		ExternalID:          p.ExternalID,
		Amount:              p.Amount,
		MCC:                 p.MCC,
		OriginalDescription: p.OriginalDescription,
		Status:              domain.TxStatusPending,
		OccurredAt:          p.OccurredAt,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	m.txs[p.ExternalID] = &tx
	return tx, false, nil
}

func (m *mockTxRepo) UpdateStatus(_ context.Context, id string, s domain.TxStatus, category string) error {
	m.status[id] = s
	// Update category in stored tx
	for _, t := range m.txs {
		if t.ID == id {
			t.CleanCategory = category
			t.Status = s
		}
	}
	return nil
}

func (m *mockTxRepo) GetByID(_ context.Context, id string) (domain.Transaction, error) {
	for _, t := range m.txs {
		if t.ID == id {
			return *t, nil
		}
	}
	return domain.Transaction{}, errors.New("not found")
}

func (m *mockTxRepo) ListByAccount(_ context.Context, _ string, _, _ int) ([]domain.Transaction, error) {
	return nil, nil
}

// ─── mockClassifier ───────────────────────────────────────────────────────────

type mockClassifier struct {
	result port.ClassificationResult
	err    error
}

func (m *mockClassifier) Classify(_ context.Context, _ string, _ int) (port.ClassificationResult, error) {
	return m.result, m.err
}

// ─── mockNotifier ─────────────────────────────────────────────────────────────

type mockNotifier struct {
	profileUpdates []domain.GameProfile
	txUpdates      []domain.Transaction
}

func (m *mockNotifier) PushProfileUpdate(_ string, p domain.GameProfile)        { m.profileUpdates = append(m.profileUpdates, p) }
func (m *mockNotifier) PushTransactionProcessed(_ string, t domain.Transaction) { m.txUpdates = append(m.txUpdates, t) }

// ─── mockWorkerPool ───────────────────────────────────────────────────────────

type mockWorkerPool struct{ full bool }

func (m *mockWorkerPool) Submit(job port.Job) error {
	if m.full {
		return domain.ErrPoolFull
	}
	_ = job(context.Background())
	return nil
}

func (m *mockWorkerPool) Shutdown(_ context.Context) error { return nil }

// ─── mockTokenStore ───────────────────────────────────────────────────────────

type mockTokenStore struct{ tokens map[string]string }

func newMockTokenStore() *mockTokenStore { return &mockTokenStore{tokens: make(map[string]string)} }

// Save(ctx, token, userID, ttl) — порядок аргументов как в реальном RedisTokenStore
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
