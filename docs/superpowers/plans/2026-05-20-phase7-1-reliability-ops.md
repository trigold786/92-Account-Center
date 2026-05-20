# Phase 7.1 — Reliability & Ops Implementation Plan

> **For agentic workers:** Step-by-step implementation. Each task is self-contained with tests before code.

**Goal:** Improve production reliability by extracting shared circuit breaker, adding real dependency health checks, enforcing lint standards, and polishing CI/CD.

**Architecture:** Create `pkg/circuitbreaker` and `pkg/health` as workspace-level shared Go modules. Wire health checks into all 9 services via a composable checker pattern. Introduce `.golangci.yml` with strict linters and fix existing violations. Enhance CI pipeline for parallel builds and staging/prod deployment.

**Tech Stack:** Go 1.24, Gin, golangci-lint, GitHub Actions

**Dependencies:** P1-3 (CI/CD) depends on P1-4 (lint). Execute P1-1 → P1-2 → P1-4 → P1-3 in order.

---

## File Structure

### New files to create:
```
pkg/circuitbreaker/
├── circuitbreaker.go        # Enhanced shared circuit breaker
├── circuitbreaker_test.go   # State machine + concurrency + callback tests
└── go.mod                   # Module: github.com/trigold786/92-Account-Center/pkg/circuitbreaker

pkg/health/
├── checker.go               # HealthChecker interface + Result types
├── checker_test.go          # Test interface + composite
├── postgres.go              # PostgresChecker: SELECT 1
├── postgres_test.go
├── redis.go                 # RedisChecker: PING
├── redis_test.go
├── response.go              # JSON response builder
├── response_test.go
└── go.mod                   # Module: github.com/trigold786/92-Account-Center/pkg/health

.golangci.yml                # Strict linter configuration
```

### Files to modify:
```
go.work                           # Add pkg/circuitbreaker, pkg/health
notification-service/go.mod      # Replace local path with shared module
notification-service/pkg/circuitbreaker/circuitbreaker.go  # Remove (moved to shared)
notification-service/internal/service/sms_service.go       # Import shared
{all 9 services}/cmd/main.go     # Upgrade /health endpoint
.github/workflows/ci.yml         # Remove || true, add staging/prod
```

### Files to delete:
```
notification-service/pkg/circuitbreaker/circuitbreaker.go  # Replaced by shared
```

---

## Task P1-1: NF-03 — Extract Circuit Breaker to Shared Package

**Files:**
- Create: `pkg/circuitbreaker/circuitbreaker.go`
- Create: `pkg/circuitbreaker/circuitbreaker_test.go`
- Create: `pkg/circuitbreaker/go.mod`
- Modify: `go.work`
- Modify: `notification-service/go.mod`
- Modify: `notification-service/internal/service/sms_service.go`
- Delete: `notification-service/pkg/circuitbreaker/circuitbreaker.go`

- [ ] **Step 1: Create shared circuit breaker module**

`pkg/circuitbreaker/go.mod`:
```
module github.com/trigold786/92-Account-Center/pkg/circuitbreaker

go 1.24
```

`pkg/circuitbreaker/circuitbreaker.go`:
```go
package circuitbreaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type Options struct {
	MaxFailures   int
	Timeout       time.Duration
	HalfOpenMax   int
	OnStateChange func(from, to State)
}

type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	maxFailures     int
	timeout         time.Duration
	halfOpenMax     int
	halfOpenSuccess int
	lastStateChange time.Time
	totalSuccess    atomic.Int64
	totalFailure    atomic.Int64
	onStateChange   func(from, to State)
}

func New(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return NewWithOptions(Options{
		MaxFailures: maxFailures,
		Timeout:     timeout,
		HalfOpenMax: 1,
	})
}

func NewWithOptions(opts Options) *CircuitBreaker {
	halfOpenMax := opts.HalfOpenMax
	if halfOpenMax <= 0 {
		halfOpenMax = 1
	}
	return &CircuitBreaker{
		state:         StateClosed,
		maxFailures:   opts.MaxFailures,
		timeout:       opts.Timeout,
		halfOpenMax:   halfOpenMax,
		onStateChange: opts.OnStateChange,
	}
}

func (cb *CircuitBreaker) Execute(f func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastStateChange) < cb.timeout {
			cb.mu.Unlock()
			return ErrOpen
		}
		cb.setState(StateHalfOpen)
		cb.halfOpenSuccess = 0
	case StateHalfOpen:
		if cb.halfOpenSuccess >= cb.halfOpenMax {
			cb.mu.Unlock()
			return ErrOpen
		}
	}

	err := f()
	if err != nil {
		cb.failureCount++
		cb.totalFailure.Add(1)
		if cb.failureCount >= cb.maxFailures {
			cb.setState(StateOpen)
		}
		cb.mu.Unlock()
		return err
	}

	cb.halfOpenSuccess++
	cb.totalSuccess.Add(1)
	cb.failureCount = 0
	cb.setState(StateClosed)
	cb.mu.Unlock()
	return nil
}

func (cb *CircuitBreaker) setState(newState State) {
	old := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()
	if cb.onStateChange != nil {
		cb.onStateChange(old, newState)
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) SuccessCount() int64 {
	return cb.totalSuccess.Load()
}

func (cb *CircuitBreaker) FailureCount() int64 {
	return cb.totalFailure.Load()
}

func (cb *CircuitBreaker) FailureRate() float64 {
	total := cb.totalSuccess.Load() + cb.totalFailure.Load()
	if total == 0 {
		return 0
	}
	return float64(cb.totalFailure.Load()) / float64(total)
}
```

- [ ] **Step 2: Write comprehensive tests**

`pkg/circuitbreaker/circuitbreaker_test.go`:
```go
package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewDefault(t *testing.T) {
	cb := New(3, time.Second)
	if cb.State() != StateClosed {
		t.Fatalf("expected closed, got %v", cb.State())
	}
	if cb.SuccessCount() != 0 || cb.FailureCount() != 0 {
		t.Fatalf("expected zero counts")
	}
}

func TestClosedToOpenOnFailures(t *testing.T) {
	cb := New(2, time.Minute)
	err1 := cb.Execute(func() error { return errors.New("fail") })
	if err1 == nil {
		t.Fatal("expected error")
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after 1 failure, got %v", cb.State())
	}
	err2 := cb.Execute(func() error { return errors.New("fail") })
	if err2 == nil {
		t.Fatal("expected error")
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected open after 2 failures, got %v", cb.State())
	}
}

func TestOpenRejectsRequests(t *testing.T) {
	cb := New(1, time.Minute)
	cb.Execute(func() error { return errors.New("fail") })
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestOpenToHalfOpenAfterTimeout(t *testing.T) {
	cb := New(1, 50*time.Millisecond)
	cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(80 * time.Millisecond)
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil after timeout, got %v", err)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after success in half-open, got %v", cb.State())
	}
}

func TestHalfOpenLimitedProbes(t *testing.T) {
	cb := New(3, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errors.New("fail") })
	}
	time.Sleep(80 * time.Millisecond)

	err1 := cb.Execute(func() error { return nil })
	if err1 != nil {
		t.Fatalf("expected first half-open probe to pass, got %v", err1)
	}

	err2 := cb.Execute(func() error { return nil })
	if !errors.Is(err2, ErrOpen) {
		t.Fatalf("expected ErrOpen for second half-open probe, got %v", err2)
	}
}

func TestConcurrency(t *testing.T) {
	cb := New(5, time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Execute(func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}()
	}
	wg.Wait()
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after all success, got %v", cb.State())
	}
}

func TestOnStateChangeCallback(t *testing.T) {
	var transitions []string
	cb := NewWithOptions(Options{
		MaxFailures: 1,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 1,
		OnStateChange: func(from, to State) {
			transitions = append(transitions, from.String()+"->"+to.String())
		},
	})

	cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(80 * time.Millisecond)
	cb.Execute(func() error { return nil })

	if len(transitions) < 2 {
		t.Fatalf("expected at least 2 transitions, got %v", transitions)
	}
}

func TestMetrics(t *testing.T) {
	cb := New(3, time.Minute)
	cb.Execute(func() error { return nil })
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return nil })

	if cb.SuccessCount() != 2 {
		t.Fatalf("expected 2 successes, got %d", cb.SuccessCount())
	}
	if cb.FailureCount() != 1 {
		t.Fatalf("expected 1 failure, got %d", cb.FailureCount())
	}
}
```

- [ ] **Step 3: Run tests to verify they pass**

```bash
cd pkg/circuitbreaker
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 4: Register in go.work**

`go.work` — add after `./pkg/cache`:
```
./pkg/circuitbreaker
```

- [ ] **Step 5: Update notification-service to use shared package**

`notification-service/go.mod` — add:
```
require github.com/trigold786/92-Account-Center/pkg/circuitbreaker v0.0.0
replace github.com/trigold786/92-Account-Center/pkg/circuitbreaker => ../pkg/circuitbreaker
```

Update import in `notification-service/internal/service/sms_service.go`:
- Remove: `"github.com/trigold786/92-Account-Center/notification-service/pkg/circuitbreaker"`
- Add: `circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"`

Delete: `notification-service/pkg/circuitbreaker/circuitbreaker.go`

- [ ] **Step 6: Run notification-service tests**

```bash
cd notification-service
go mod tidy
go test -v -race ./...
Expected: All tests PASS (SMS circuit breaker tests still work with shared package)
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: extract circuitbreaker to shared pkg/circuitbreaker"
```

---

## Task P1-2: NF-04 — Health Checks with Real Dependency Detection

**Files:**
- Create: `pkg/health/go.mod`
- Create: `pkg/health/checker.go`
- Create: `pkg/health/checker_test.go`
- Create: `pkg/health/postgres.go`
- Create: `pkg/health/postgres_test.go`
- Create: `pkg/health/redis.go`
- Create: `pkg/health/redis_test.go`
- Create: `pkg/health/response.go`
- Create: `pkg/health/response_test.go`
- Modify: `go.work` (add `./pkg/health`)
- Modify: All 9 services' `cmd/main.go` (upgrade `/health` endpoint)

- [ ] **Step 1: Write tests for health checker**

`pkg/health/checker_test.go`:
```go
package health

import (
	"context"
	"errors"
	"testing"
)

func TestComponentHealthStatus(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		wantOK bool
	}{
		{"up", StatusUp, true},
		{"degraded", StatusDegraded, true},
		{"down", StatusDown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := ComponentHealth{Name: "test", Status: tt.status}
			if (ch.Status == StatusUp || ch.Status == StatusDegraded) != tt.wantOK {
				t.Errorf("unexpected ok state for status %v", tt.status)
			}
		})
	}
}

func TestCompositeChecker(t *testing.T) {
	ok := &mockChecker{name: "ok", status: StatusUp}
	fail := &mockChecker{name: "fail", status: StatusDown, err: errors.New("fail")}

	allOK := CompositeChecker{Checkers: []Checker{ok}}
	result := allOK.Check(context.Background())
	if result.Status != StatusUp {
		t.Fatalf("expected up, got %v", result.Status)
	}

	withFail := CompositeChecker{Checkers: []Checker{ok, fail}}
	result2 := withFail.Check(context.Background())
	if result2.Status != StatusDown {
		t.Fatalf("expected down, got %v", result2.Status)
	}
	if len(result2.Checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(result2.Checks))
	}
}

type mockChecker struct {
	name   string
	status Status
	err    error
}

func (m *mockChecker) Name() string { return m.name }

func (m *mockChecker) Check(ctx context.Context) ComponentHealth {
	ch := ComponentHealth{Name: m.name, Status: m.status}
	if m.err != nil {
		ch.Error = m.err.Error()
	}
	return ch
}
```

`pkg/health/response_test.go`:
```go
package health

import (
	"encoding/json"
	"testing"
)

func TestHealthResponseJSON(t *testing.T) {
	checks := map[string]ComponentHealth{
		"postgres": {Name: "postgres", Status: StatusUp, LatencyMs: 2},
	}
	resp := BuildResponse(checks)
	if resp.Status != "ok" {
		t.Fatalf("expected ok, got %s", resp.Status)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if parsed["checks"] == nil {
		t.Fatal("expected checks field in JSON")
	}
}
```

`pkg/health/postgres_test.go`:
```go
package health

import (
	"context"
	"testing"
)

func TestPostgresCheckerMissingDB(t *testing.T) {
	pc := &PostgresChecker{DBNop: true}
	result := pc.Check(context.Background())
	if result.Status != StatusDown {
		t.Fatalf("expected down without DB, got %v", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestPostgresCheckerName(t *testing.T) {
	pc := &PostgresChecker{}
	if pc.Name() != "postgres" {
		t.Fatalf("unexpected name: %s", pc.Name())
	}
}
```

`pkg/health/redis_test.go`:
```go
package health

import (
	"context"
	"testing"
)

func TestRedisCheckerMissingRedis(t *testing.T) {
	rc := &RedisChecker{RedisNop: true}
	result := rc.Check(context.Background())
	if result.Status != StatusDown {
		t.Fatalf("expected down without Redis, got %v", result.Status)
	}
}

func TestRedisCheckerName(t *testing.T) {
	rc := &RedisChecker{}
	if rc.Name() != "redis" {
		t.Fatalf("unexpected name: %s", rc.Name())
	}
}
```

- [ ] **Step 2: Implement the health checker module**

`pkg/health/go.mod`:
```
module github.com/trigold786/92-Account-Center/pkg/health

go 1.24
```

`pkg/health/checker.go`:
```go
package health

import (
	"context"
)

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
	Name      string `json:"name"`
	Status    Status `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
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
```

`pkg/health/response.go`:
```go
package health

type HealthResponse struct {
	Status   string                       `json:"status"`
	Checks   map[string]ComponentHealth   `json:"checks,omitempty"`
}

func BuildResponse(checks map[string]ComponentHealth) HealthResponse {
	overall := "ok"
	for _, ch := range checks {
		if ch.Status == StatusDown {
			overall = "down"
			break
		}
		if ch.Status == StatusDegraded {
			overall = "degraded"
		}
	}
	return HealthResponse{
		Status: overall,
		Checks: checks,
	}
}
```

`pkg/health/postgres.go`:
```go
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
```

`pkg/health/redis.go`:
```go
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
```

- [ ] **Step 3: Run tests**

```bash
cd pkg/health
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 4: Register in go.work**

Add `./pkg/health` to `go.work`.

- [ ] **Step 5: Upgrade /health endpoint in all 9 services**

Example for `account-service/cmd/main.go`:
```go
import healthpkg "github.com/trigold786/92-Account-Center/pkg/health"

// In setup, after DB and Redis connections:
var healthCheckers []healthpkg.Checker
if db != nil {
    healthCheckers = append(healthCheckers, &healthpkg.PostgresChecker{
        Ping: func(ctx context.Context) error {
            _, err := db.ExecContext(ctx, "SELECT 1")
            return err
        },
    })
}
if rdb != nil {
    healthCheckers = append(healthCheckers, &healthpkg.RedisChecker{
        Ping: func(ctx context.Context) error {
            return rdb.Ping(ctx).Err()
        },
    })
}
composite := healthpkg.CompositeChecker{Checkers: healthCheckers}

// Replace existing /health handler:
r.Any("/health", func(c *gin.Context) {
    result := composite.Check(c.Request.Context())
    resp := healthpkg.BuildResponse(result.Checks)
    statusCode := 200
    if result.Status == healthpkg.StatusDown {
        statusCode = 503
    }
    c.JSON(statusCode, resp)
})
```

Apply this pattern to all 9 services. Services that use both PG and Redis get both checkers; services that only use one get just that one (e.g., notification-service is Redis-only).

- [ ] **Step 6: Run all service tests**

```bash
make test
Expected: All tests PASS
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add pkg/health with real dependency checks for all services"
```

---

## Task P1-4: AR-28 — Lint Strictness

**Files:**
- Create: `.golangci.yml`
- Modify: `.github/workflows/ci.yml` (remove `|| true` from lint)

- [ ] **Step 1: Create .golangci.yml**

`.golangci.yml`:
```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - gosec
    - revive
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - misspell
  disable:
    - exhaustive

linters-settings:
  revive:
    rules:
      - name: exported
        severity: warning
      - name: blank-imports
      - name: context-as-argument
      - name: error-return
      - name: error-strings
      - name: unused-parameter
        severity: warning
  staticcheck:
    checks: ["all"]
  gosec:
    excludes:
      - G101  # Look for hardcoded credentials (false positives in test mocks)
      - G204  # Subprocess launched with variable
  misspell:
    locale: US

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - gosec
        - unused
    - path: cmd/main\.go
      text: "hugeParam"
      linters:
        - revive
  max-issues-per-linter: 0
  max-same-issues: 0
```

- [ ] **Step 2: Update CI to remove `|| true` from lint**

In `.github/workflows/ci.yml`, change:
```yaml
- name: Lint
  run: |
    cd ${{ matrix.service }}
    golangci-lint run ./... || true  # Remove the || true
```
to:
```yaml
- name: Lint
  run: |
    cd ${{ matrix.service }}
    golangci-lint run ./...
```

- [ ] **Step 3: Run lint and fix all errors**

```bash
golangci-lint run ./...
```
Fix all reported issues (unused exports, missing error checks, context propagation, etc.). Common fixes:
- Remove unused functions/variables
- Add missing error handling
- Fix context-as-argument ordering
- Fix comment style for exported symbols

- [ ] **Step 4: Run tests to ensure no regressions**

```bash
make test
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: add .golangci.yml and fix all lint errors"
```

---

## Task P1-3: AR-22 — CI/CD Pipeline Improvement

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Enhance CI workflow**

`.github/workflows/ci.yml` — key improvements:
1. Remove `|| true` from lint and test (done in P1-4)
2. Add parallel build matrix (already done)
3. Add staging deploy job (workflow_dispatch only)
4. Add integration test job
5. Add security scan job

```yaml
name: CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deploy environment'
        required: true
        default: 'dev'
        type: choice
        options:
          - dev
          - staging
          - prod

# ... existing lint, test, build jobs unchanged ...

  integration-test:
    runs-on: ubuntu-latest
    needs: [build]
    services:
      postgres:
        image: postgres:18-alpine
        env:
          POSTGRES_DB: account_center
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run integration tests
        run: |
          cd tests/integration
          go test -v -count=1 -timeout 300s ./...
        env:
          TEST_PG_DSN: postgres://test:test@localhost:5432/account_center?sslmode=disable
          TEST_REDIS_ADDR: localhost:6379

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: -no-fail -fmt sarif -out results.sarif ./...
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif

  deploy-staging:
    runs-on: ubuntu-latest
    needs: [integration-test, security-scan]
    if: github.event_name == 'workflow_dispatch' && github.event.inputs.environment == 'staging'
    # ... existing deploy logic adapted for staging ...

  deploy-prod:
    runs-on: ubuntu-latest
    needs: [integration-test, security-scan]
    if: github.event_name == 'workflow_dispatch' && github.event.inputs.environment == 'prod'
    environment: production
    # ... existing deploy logic adapted for prod with approval gate ...
```

- [ ] **Step 2: Verify CI file is valid YAML**

```bash
python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"
Expected: No errors
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "ci: enhance CI/CD with integration tests, security scan, staging/prod deploy"
```

---

## Task: Update PRD and CHANGELOG

After all tasks complete, update documentation per phase gate requirements:

- [ ] Mark PRD entries for NF-03, NF-04, AR-22, AR-28 as "已实现"
- [ ] Update CHANGELOG.md with Phase 7.1 entries
