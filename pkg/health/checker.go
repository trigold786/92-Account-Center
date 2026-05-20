package health

import "context"

type Status int

const (
	StatusUp       Status = 0
	StatusDegraded Status = 1
	StatusDown     Status = 2
)

func (s Status) String() string {
	switch s {
	case StatusUp:
		return "up"
	case StatusDegraded:
		return "degraded"
	case StatusDown:
		return "down"
	default:
		return "unknown"
	}
}

type ComponentHealth struct {
	Name      string                     `json:"name"`
	Status    Status                     `json:"status"`
	LatencyMs int64                      `json:"latency_ms,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Checks map[string]ComponentHealth `json:"checks,omitempty"`
}

type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

type CheckFunc func(ctx context.Context) ComponentHealth

func (f CheckFunc) Name() string { return "custom" }

func (f CheckFunc) Check(ctx context.Context) ComponentHealth {
	return f(ctx)
}

type CompositeChecker struct {
	Checkers []Checker
}

func (c CompositeChecker) Name() string { return "composite" }

func (c CompositeChecker) Check(ctx context.Context) ComponentHealth {
	agg := ComponentHealth{Name: "composite", Status: StatusUp}
	for _, checker := range c.Checkers {
		ch := checker.Check(ctx)
		if agg.Checks == nil {
			agg.Checks = make(map[string]ComponentHealth)
		}
		agg.Checks[checker.Name()] = ch
		if ch.Status == StatusDown {
			agg.Status = StatusDown
		} else if ch.Status == StatusDegraded && agg.Status != StatusDown {
			agg.Status = StatusDegraded
		}
	}
	return agg
}
