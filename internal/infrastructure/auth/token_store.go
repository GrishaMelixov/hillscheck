package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/hillscheck/internal/domain"
)

const redisPrefix = "refresh:"

// RedisTokenStore implements port.TokenStore using Redis.
type RedisTokenStore struct {
	rdb *redis.Client
}

func NewRedisTokenStore(rdb *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{rdb: rdb}
}

func (s *RedisTokenStore) Save(ctx context.Context, token, userID string, ttl time.Duration) error {
	if err := s.rdb.Set(ctx, redisPrefix+token, userID, ttl).Err(); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (s *RedisTokenStore) Get(ctx context.Context, token string) (string, error) {
	userID, err := s.rdb.Get(ctx, redisPrefix+token).Result()
	if errors.Is(err, redis.Nil) {
		return "", domain.ErrTokenInvalid
	}
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return userID, nil
}

func (s *RedisTokenStore) Delete(ctx context.Context, token string) error {
	if err := s.rdb.Del(ctx, redisPrefix+token).Err(); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}
