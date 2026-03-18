package domain

import "time"

type UserSettings struct {
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
}

// Plan represents the user's subscription tier.
type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
)

type User struct {
	ID           string
	Name         string
	Email        string
	EmailHash    string // SHA-256 of normalised email (kept for privacy features)
	PasswordHash string
	Plan         Plan
	Settings     UserSettings
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
