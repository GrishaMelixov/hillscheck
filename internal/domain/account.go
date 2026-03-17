package domain

import "time"

type AccountType string

const (
	AccountTypeDebit  AccountType = "debit"
	AccountTypeCredit AccountType = "credit"
	AccountTypeCash   AccountType = "cash"
)

type Account struct {
	ID        string
	UserID    string
	Name      string
	Type      AccountType
	Balance   int64  // stored in cents; int64, never float
	Currency  string // ISO 4217, e.g. "USD"
	CreatedAt time.Time
	UpdatedAt time.Time
}
