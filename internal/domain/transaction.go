package domain

import "time"

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusProcessed TxStatus = "processed"
	TxStatusFailed    TxStatus = "failed"
)

type Transaction struct {
	ID                  string
	ExternalID          string   // idempotency key from the client / bank
	AccountID           string
	Amount              int64    // cents; negative = spend, positive = income
	MCC                 int      // ISO 18245 merchant category code
	OriginalDescription string
	CleanCategory       string   // assigned after classification
	Status              TxStatus
	OccurredAt          time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
