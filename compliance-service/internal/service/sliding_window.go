package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/svcconfig"
)

type SlidingWindowLimiter struct {
	rdb *redis.Client
	cfg *svcconfig.ComplianceConfig
}

func NewSlidingWindowLimiter(rdb *redis.Client, cfg *svcconfig.ComplianceConfig) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{rdb: rdb, cfg: cfg}
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
	return l.Allow(ctx, key, l.cfg.SlidingWindowRegWindow, int64(l.cfg.SlidingWindowRegLimit))
}

func (l *SlidingWindowLimiter) CheckReferralAbuse(ctx context.Context, referrerCode string) (bool, int64, error) {
	key := fmt.Sprintf("ratelimit:referral:code:%s", referrerCode)
	return l.Allow(ctx, key, l.cfg.SlidingWindowRefAbuseWindow, int64(l.cfg.SlidingWindowRefAbuseLimit))
}
