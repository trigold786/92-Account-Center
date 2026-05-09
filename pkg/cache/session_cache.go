package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionCache defines the interface for session cache operations.
type SessionCache interface {
	InvalidateUserSessions(ctx context.Context, userID string) error
	AddSession(ctx context.Context, userID, token string, ttl time.Duration) error
	GetUserSessions(ctx context.Context, userID string) ([]string, error)
}

// sessionCache implements SessionCache using Redis.
type sessionCache struct {
	redis *redis.Client
}

// NewSessionCache creates a new SessionCache.
func NewSessionCache(redis *redis.Client) SessionCache {
	return &sessionCache{redis: redis}
}

// AddSession adds a session token for a user.
func (c *sessionCache) AddSession(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := "session:" + token
	return c.redis.Set(ctx, key, userID, ttl).Err()
}

// InvalidateUserSessions removes all sessions for a user.
func (c *sessionCache) InvalidateUserSessions(ctx context.Context, userID string) error {
	// Pattern to match all session keys for this user
	pattern := "user_sessions:" + userID
	sessions, err := c.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	// Delete all session keys
	if len(sessions) > 0 {
		pipe := c.redis.Pipeline()
		for _, session := range sessions {
			pipe.Del(ctx, session)
		}
		_, err = pipe.Exec(ctx)
	}
	return err
}

// GetUserSessions retrieves all session tokens for a user.
func (c *sessionCache) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	key := "user_sessions:" + userID
	return c.redis.SMembers(ctx, key).Result()
}