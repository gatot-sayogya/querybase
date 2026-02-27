package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklistService handles blacklisting of JWT tokens using Redis
type TokenBlacklistService struct {
	redisClient *redis.Client
}

// NewTokenBlacklistService creates a new token blacklist service
func NewTokenBlacklistService(redisClient *redis.Client) *TokenBlacklistService {
	return &TokenBlacklistService{
		redisClient: redisClient,
	}
}

// BlacklistToken adds a token JTI to the blacklist with an expiration time
func (s *TokenBlacklistService) BlacklistToken(ctx context.Context, jti string, expiration time.Duration) error {
	if s.redisClient == nil {
		return nil // Fallback if Redis is not available
	}

	key := fmt.Sprintf("blacklist:%s", jti)
	return s.redisClient.Set(ctx, key, "1", expiration).Err()
}

// IsBlacklisted checks if a token JTI is in the blacklist
func (s *TokenBlacklistService) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if s.redisClient == nil {
		return false, nil // Fallback if Redis is not available
	}

	key := fmt.Sprintf("blacklist:%s", jti)
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// StoreRefreshToken stores a refresh token in Redis mapped to a user ID
func (s *TokenBlacklistService) StoreRefreshToken(ctx context.Context, refreshToken string, userID string, expiration time.Duration) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	return s.redisClient.Set(ctx, key, userID, expiration).Err()
}

// GetUserByRefreshToken retrieves the user ID associated with a refresh token
func (s *TokenBlacklistService) GetUserByRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	if s.redisClient == nil {
		return "", fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	return s.redisClient.Get(ctx, key).Result()
}

// DeleteRefreshToken removes a refresh token from Redis
func (s *TokenBlacklistService) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	if s.redisClient == nil {
		return nil
	}

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	return s.redisClient.Del(ctx, key).Err()
}
