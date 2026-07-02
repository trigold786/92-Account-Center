package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb         *redis.Client
	maxAttempts int
	window      time.Duration
}

func NewRateLimiter(rdb *redis.Client, maxAttempts int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, maxAttempts: maxAttempts, window: window}
}

func (rl *RateLimiter) IsLocked(ctx context.Context, key string) bool {
	if rl.rdb == nil {
		return false
	}
	val, err := rl.rdb.Get(ctx, "lockout:"+key).Result()
	if err != nil {
		return false
	}
	return val == "1"
}

func (rl *RateLimiter) RecordAttempt(ctx context.Context, key string) {
	if rl.rdb == nil {
		return
	}
	lockKey := "lockout:" + key
	countKey := "lockout:count:" + key

	count, err := rl.rdb.Incr(ctx, countKey).Result()
	if err != nil {
		return
	}
	if count == 1 {
		rl.rdb.Expire(ctx, countKey, rl.window)
	}
	if count >= int64(rl.maxAttempts) {
		rl.rdb.Set(ctx, lockKey, "1", rl.window)
	}
}

func (rl *RateLimiter) Reset(ctx context.Context, key string) {
	if rl.rdb == nil {
		return
	}
	rl.rdb.Del(ctx, "lockout:"+key, "lockout:count:"+key)
}
