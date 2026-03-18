package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hillscheck/internal/domain"
)

type GameRepo struct {
	pool *pgxpool.Pool
}

func NewGameRepo(pool *pgxpool.Pool) *GameRepo {
	return &GameRepo{pool: pool}
}

func (r *GameRepo) CreateProfile(ctx context.Context, userID string) (domain.GameProfile, error) {
	const q = `
		INSERT INTO game_profiles (user_id, level, xp, hp, mana, strength, intellect, luck)
		VALUES ($1, 1, 0, 100, 50, 10, 10, 10)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING user_id, level, xp, hp, mana, strength, intellect, luck, updated_at`

	var p domain.GameProfile
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.Level, &p.XP, &p.HP, &p.Mana,
		&p.Strength, &p.Intellect, &p.Luck, &p.UpdatedAt,
	)
	if err != nil {
		return domain.GameProfile{}, fmt.Errorf("create game profile: %w", err)
	}
	return p, nil
}

func (r *GameRepo) GetProfile(ctx context.Context, userID string) (domain.GameProfile, error) {
	const q = `
		SELECT user_id, level, xp, hp, mana, strength, intellect, luck, updated_at
		FROM game_profiles WHERE user_id = $1`

	var p domain.GameProfile
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.Level, &p.XP, &p.HP, &p.Mana,
		&p.Strength, &p.Intellect, &p.Luck, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.GameProfile{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.GameProfile{}, fmt.Errorf("get game profile: %w", err)
	}
	return p, nil
}

// ApplyEvent atomically inserts a game event and updates the corresponding
// attribute in game_profiles inside a single DB transaction.
// Uses SELECT FOR UPDATE to prevent concurrent XP/HP races.
func (r *GameRepo) ApplyEvent(ctx context.Context, event domain.GameEvent) (domain.GameProfile, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.GameProfile{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lock the profile row to serialise concurrent updates for this user.
	const lockQ = `SELECT 1 FROM game_profiles WHERE user_id = $1 FOR UPDATE`
	if _, err := tx.Exec(ctx, lockQ, event.UserID); err != nil {
		return domain.GameProfile{}, fmt.Errorf("lock profile: %w", err)
	}

	// Persist the event record.
	const insertEvent = `
		INSERT INTO game_events (transaction_id, user_id, attribute, delta, reason)
		VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.Exec(ctx, insertEvent,
		event.TransactionID, event.UserID, event.Attribute, event.Delta, event.Reason,
	); err != nil {
		return domain.GameProfile{}, fmt.Errorf("insert game event: %w", err)
	}

	// Apply the delta using a safe, attribute-specific query.
	updateQ, err := buildAttributeUpdate(event.Attribute)
	if err != nil {
		return domain.GameProfile{}, err
	}

	var profile domain.GameProfile
	err = tx.QueryRow(ctx, updateQ, event.Delta, event.UserID).Scan(
		&profile.UserID, &profile.Level, &profile.XP, &profile.HP, &profile.Mana,
		&profile.Strength, &profile.Intellect, &profile.Luck, &profile.UpdatedAt,
	)
	if err != nil {
		return domain.GameProfile{}, fmt.Errorf("apply attribute delta: %w", err)
	}

	// Check and apply level-up if XP crossed the threshold.
	if event.Attribute == "xp" && profile.XP >= profile.XPForNextLevel() {
		profile, err = levelUp(ctx, tx, profile)
		if err != nil {
			return domain.GameProfile{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.GameProfile{}, fmt.Errorf("commit game event tx: %w", err)
	}
	return profile, nil
}

func (r *GameRepo) ListEvents(ctx context.Context, userID string, limit int) ([]domain.GameEvent, error) {
	const q = `
		SELECT id, transaction_id, user_id, attribute, delta, reason, occurred_at
		FROM game_events WHERE user_id = $1
		ORDER BY occurred_at DESC LIMIT $2`

	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list game events: %w", err)
	}
	defer rows.Close()

	var events []domain.GameEvent
	for rows.Next() {
		var e domain.GameEvent
		if err := rows.Scan(&e.ID, &e.TransactionID, &e.UserID, &e.Attribute, &e.Delta, &e.Reason, &e.OccurredAt); err != nil {
			return nil, fmt.Errorf("scan game event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// buildAttributeUpdate returns a safe UPDATE query for the given attribute.
// Dynamic column names are not parameterisable in SQL, so we whitelist them here.
func buildAttributeUpdate(attribute string) (string, error) {
	const tpl = `
		UPDATE game_profiles SET %s = GREATEST(0, %s + $1), updated_at = now()
		WHERE user_id = $2
		RETURNING user_id, level, xp, hp, mana, strength, intellect, luck, updated_at`

	allowed := map[string]string{
		"xp":        "xp",
		"hp":        "hp",
		"mana":      "mana",
		"strength":  "strength",
		"intellect": "intellect",
		"luck":      "luck",
	}
	col, ok := allowed[attribute]
	if !ok {
		return "", fmt.Errorf("unknown game attribute: %q", attribute)
	}
	return fmt.Sprintf(tpl, col, col), nil
}

// levelUp increments the level and resets XP to the carry-over remainder.
func levelUp(ctx context.Context, tx pgx.Tx, p domain.GameProfile) (domain.GameProfile, error) {
	const q = `
		UPDATE game_profiles
		SET level = level + 1, xp = xp - (level * level * 100), updated_at = now()
		WHERE user_id = $1
		RETURNING user_id, level, xp, hp, mana, strength, intellect, luck, updated_at`

	var updated domain.GameProfile
	err := tx.QueryRow(ctx, q, p.UserID).Scan(
		&updated.UserID, &updated.Level, &updated.XP, &updated.HP, &updated.Mana,
		&updated.Strength, &updated.Intellect, &updated.Luck, &updated.UpdatedAt,
	)
	if err != nil {
		return domain.GameProfile{}, fmt.Errorf("level up: %w", err)
	}
	return updated, nil
}
