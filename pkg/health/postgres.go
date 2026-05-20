package health

import (
	"context"
	"time"
)

type PostgresChecker struct {
	DBNop bool
	Ping  func(context.Context) error
}

func (pc *PostgresChecker) Name() string { return "postgres" }

func (pc *PostgresChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	if pc.DBNop {
		return ComponentHealth{
			Name:   "postgres",
			Status: StatusDown,
			Error:  "no database configured",
		}
	}
	if pc.Ping == nil {
		return ComponentHealth{
			Name:   "postgres",
			Status: StatusUp,
		}
	}
	err := pc.Ping(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return ComponentHealth{
			Name:      "postgres",
			Status:    StatusDown,
			LatencyMs: latency,
			Error:     err.Error(),
		}
	}
	return ComponentHealth{
		Name:      "postgres",
		Status:    StatusUp,
		LatencyMs: latency,
	}
}
