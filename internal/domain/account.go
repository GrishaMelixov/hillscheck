package domain

import "time"

type AccountType string

const (
	AccountTypeDebit  AccountType = "debit"
	AccountTypeCredit AccountType = "credit"
	AccountTypeCash   AccountType = "cash"
)

type Account struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	Balance   int64       `json:"balance"`
	Currency  string      `json:"currency"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
