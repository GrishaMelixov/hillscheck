package usecase

import (
	"context"
	"fmt"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// Quest is a generated challenge based on spending patterns.
type Quest struct {
	ID          string
	Title       string
	Description string
	Attribute   string // target RPG attribute
	Reward      int    // XP reward on completion
	Progress    int    // 0-100
}

type GetQuests struct {
	gameRepo port.GameRepository
	txRepo   port.TransactionRepository
	accounts port.AccountRepository
}

func NewGetQuests(
	gameRepo port.GameRepository,
	txRepo port.TransactionRepository,
	accounts port.AccountRepository,
) *GetQuests {
	return &GetQuests{
		gameRepo: gameRepo,
		txRepo:   txRepo,
		accounts: accounts,
	}
}

// Execute generates quests from the user's current game profile and spending history.
// Quests are derived from the profile's weakest attributes.
func (u *GetQuests) Execute(ctx context.Context, userID string) ([]Quest, error) {
	profile, err := u.gameRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile for quests: %w", err)
	}

	events, err := u.gameRepo.ListEvents(ctx, userID, 50)
	if err != nil {
		return nil, fmt.Errorf("list events for quests: %w", err)
	}

	return generateQuests(profile, events), nil
}

// generateQuests builds quests targeting the weakest RPG attributes.
func generateQuests(profile domain.GameProfile, events []domain.GameEvent) []Quest {
	// Count how often each attribute was trained recently.
	counts := map[string]int{}
	for _, e := range events {
		if e.Delta > 0 {
			counts[e.Attribute]++
		}
	}

	var quests []Quest

	if profile.Intellect < 20 && counts["intellect"] < 3 {
		quests = append(quests, Quest{
			ID:          "quest-read-books",
			Title:       "Scholar",
			Description: "Buy 3 books or courses this month",
			Attribute:   port.AttrIntellect,
			Reward:      50,
			Progress:    clamp(counts["intellect"]*33, 0, 100),
		})
	}

	if profile.Strength < 20 && counts["strength"] < 3 {
		quests = append(quests, Quest{
			ID:          "quest-gym",
			Title:       "Warrior",
			Description: "Visit the gym or buy sports gear 3 times",
			Attribute:   port.AttrStrength,
			Reward:      40,
			Progress:    clamp(counts["strength"]*33, 0, 100),
		})
	}

	if profile.HP < 50 {
		quests = append(quests, Quest{
			ID:          "quest-eat-healthy",
			Title:       "Survivor",
			Description: "Avoid fast food for a week",
			Attribute:   port.AttrHP,
			Reward:      30,
			Progress:    clamp(100-profile.HP, 0, 100),
		})
	}

	// Default quest: always give something to do.
	if len(quests) == 0 {
		nextXP := profile.XPForNextLevel()
		progress := int(float64(profile.XP) / float64(nextXP) * 100)
		quests = append(quests, Quest{
			ID:          fmt.Sprintf("quest-level-%d", profile.Level),
			Title:       "Adventurer",
			Description: fmt.Sprintf("Reach level %d", profile.Level+1),
			Attribute:   port.AttrXP,
			Reward:      100,
			Progress:    clamp(progress, 0, 100),
		})
	}

	return quests
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
