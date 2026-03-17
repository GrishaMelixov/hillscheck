package domain

import "time"

type GameProfile struct {
	UserID    string
	Level     int
	XP        int64
	HP        int
	Mana      int
	Strength  int
	Intellect int
	Luck      int
	UpdatedAt time.Time
}

// XPForNextLevel returns the XP threshold needed to reach the next level.
// Uses a quadratic progression: level² × 100.
func (g GameProfile) XPForNextLevel() int64 {
	return int64(g.Level) * int64(g.Level) * 100
}

type GameEvent struct {
	ID            string
	TransactionID string
	UserID        string
	Attribute     string // "xp" | "hp" | "mana" | "strength" | "intellect" | "luck"
	Delta         int    // signed delta applied to the attribute
	Reason        string // human-readable label, e.g. "Food purchase"
	OccurredAt    time.Time
}
