package port

import "context"

// RPG attributes that can be affected by a transaction.
const (
	AttrXP        = "xp"
	AttrHP        = "hp"
	AttrMana      = "mana"
	AttrStrength  = "strength"
	AttrIntellect = "intellect"
	AttrLuck      = "luck"
)

// Impact describes which RPG attribute is affected and by how much.
type Impact struct {
	Attribute string // one of the Attr* constants
	Value     int    // signed delta
}

// ClassificationResult is the structured output from a CategoryProvider.
type ClassificationResult struct {
	Category string
	Impact   Impact
}

// CategoryProvider classifies a raw transaction description and MCC into
// a human-readable category and an RPG attribute impact.
// Implementations: GeminiProvider, OllamaProvider, ChainedProvider.
type CategoryProvider interface {
	Classify(ctx context.Context, description string, mcc int) (ClassificationResult, error)
}
