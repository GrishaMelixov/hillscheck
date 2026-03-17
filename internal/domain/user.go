package domain

import "time"

type UserSettings struct {
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
}

type User struct {
	ID        string
	Name      string
	EmailHash string // SHA-256 of normalised email
	Settings  UserSettings
	CreatedAt time.Time
	UpdatedAt time.Time
}
