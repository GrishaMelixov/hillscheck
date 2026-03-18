package domain

import "time"

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusProcessed TxStatus = "processed"
	TxStatusFailed    TxStatus = "failed"
)

type Transaction struct {
	ID                  string    `json:"id"`
	ExternalID          string    `json:"external_id"`
	AccountID           string    `json:"account_id"`
	Amount              int64     `json:"amount"`
	MCC                 int       `json:"mcc"`
	OriginalDescription string    `json:"original_description"`
	CleanCategory       string    `json:"clean_category"`
	Status              TxStatus  `json:"status"`
	OccurredAt          time.Time `json:"occurred_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
