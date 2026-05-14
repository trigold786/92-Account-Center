package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	rdb *redis.Client
}

func NewSlidingWindowLimiter(rdb *redis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{rdb: rdb}
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string, window time.Duration, maxCount int64) (bool, int64, error) {
	if l.rdb == nil {
		return true, 0, nil
	}
	now := time.Now()
	pipe := l.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now.Add(-window).UnixNano()))
	pipe.ZCard(ctx, key)
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixNano()), Member: now.UnixNano()})
	pipe.Expire(ctx, key, window+time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, 0, err
	}
	count := countCmd.Val()
	if count >= maxCount {
		return false, count, nil
	}
	return true, count, nil
}

func (l *SlidingWindowLimiter) CheckRegistrationLimit(ctx context.Context, ip string) (bool, int64, error) {
	key := fmt.Sprintf("ratelimit:register:ip:%s", ip)
	return l.Allow(ctx, key, time.Hour, 3)
}

func (l *SlidingWindowLimiter) CheckReferralAbuse(ctx context.Context, referrerCode string) (bool, int64, error) {
	key := fmt.Sprintf("ratelimit:referral:code:%s", referrerCode)
	return l.Allow(ctx, key, time.Hour, 50)
}
