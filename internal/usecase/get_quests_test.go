package usecase_test

import (
	"context"
	"testing"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// TestGetQuests_WeakIntellect — пользователь с низким Intellect получает квест Scholar.
func TestGetQuests_WeakIntellect(t *testing.T) {
	gameRepo := newMockGameRepo()
	gameRepo.profiles["user-1"] = &domain.GameProfile{
		UserID:    "user-1",
		Level:     2,
		XP:        100,
		HP:        80,
		Mana:      80,
		Strength:  25, // выше порога
		Intellect: 10, // ниже 20 → должен быть квест
		Luck:      15,
	}

	uc := usecase.NewGetQuests(gameRepo, newMockTxRepo(), newMockAccountRepo())
	quests, err := uc.Execute(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, q := range quests {
		if q.Attribute == port.AttrIntellect {
			found = true
			if q.Reward != 50 {
				t.Errorf("Scholar reward: want 50, got %d", q.Reward)
			}
		}
	}
	if !found {
		t.Error("expected Scholar quest for low intellect user")
	}
}

// TestGetQuests_WeakStrength — пользователь с низким Strength получает квест Warrior.
func TestGetQuests_WeakStrength(t *testing.T) {
	gameRepo := newMockGameRepo()
	gameRepo.profiles["user-2"] = &domain.GameProfile{
		UserID:    "user-2",
		Level:     1,
		HP:        90,
		Mana:      90,
		Strength:  5, // ниже 20 → квест Warrior
		Intellect: 25,
		Luck:      10,
	}

	uc := usecase.NewGetQuests(gameRepo, newMockTxRepo(), newMockAccountRepo())
	quests, _ := uc.Execute(context.Background(), "user-2")

	found := false
	for _, q := range quests {
		if q.ID == "quest-gym" {
			found = true
		}
	}
	if !found {
		t.Error("expected Warrior quest for low strength user")
	}
}

// TestGetQuests_LowHP — пользователь с HP < 50 получает квест Survivor.
func TestGetQuests_LowHP(t *testing.T) {
	gameRepo := newMockGameRepo()
	gameRepo.profiles["user-3"] = &domain.GameProfile{
		UserID: "user-3", Level: 3, HP: 30, // ниже 50 → квест Survivor
		Strength: 25, Intellect: 25, Luck: 15, Mana: 80,
	}

	uc := usecase.NewGetQuests(gameRepo, newMockTxRepo(), newMockAccountRepo())
	quests, _ := uc.Execute(context.Background(), "user-3")

	found := false
	for _, q := range quests {
		if q.ID == "quest-eat-healthy" {
			found = true
			if q.Progress <= 0 {
				t.Error("Survivor quest should have positive progress for low HP")
			}
		}
	}
	if !found {
		t.Error("expected Survivor quest for HP < 50 user")
	}
}

// TestGetQuests_HealthyUser — прокачанный пользователь получает дефолтный квест Adventurer.
func TestGetQuests_HealthyUser(t *testing.T) {
	gameRepo := newMockGameRepo()
	gameRepo.profiles["user-4"] = &domain.GameProfile{
		UserID:    "user-4",
		Level:     5,
		XP:        500,
		HP:        90,
		Mana:      90,
		Strength:  30,
		Intellect: 30,
		Luck:      20,
	}

	uc := usecase.NewGetQuests(gameRepo, newMockTxRepo(), newMockAccountRepo())
	quests, _ := uc.Execute(context.Background(), "user-4")

	if len(quests) == 0 {
		t.Fatal("expected at least one quest (default Adventurer)")
	}
	// Здоровый юзер → только дефолтный quest
	if quests[0].Attribute != port.AttrXP {
		t.Errorf("healthy user should get XP quest, got attribute=%s", quests[0].Attribute)
	}
}

// TestGetQuests_ProfileNotFound — если профиль не существует, возвращается ошибка.
func TestGetQuests_ProfileNotFound(t *testing.T) {
	uc := usecase.NewGetQuests(newMockGameRepo(), newMockTxRepo(), newMockAccountRepo())
	_, err := uc.Execute(context.Background(), "ghost-user")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}
