package health

import (
	"context"
	"time"
)

type RedisChecker struct {
	RedisNop bool
	Ping     func(context.Context) error
}

func (rc *RedisChecker) Name() string { return "redis" }

func (rc *RedisChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	if rc.RedisNop {
		return ComponentHealth{
			Name:   "redis",
			Status: StatusDown,
			Error:  "no redis configured",
		}
	}
	if rc.Ping == nil {
		return ComponentHealth{
			Name:   "redis",
			Status: StatusUp,
		}
	}
	err := rc.Ping(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return ComponentHealth{
			Name:      "redis",
			Status:    StatusDown,
			LatencyMs: latency,
			Error:     err.Error(),
		}
	}
	return ComponentHealth{
		Name:      "redis",
		Status:    StatusUp,
		LatencyMs: latency,
	}
}
