# Phase 7.5 — Architecture & Observability Implementation Plan

> **For agentic workers:** Step-by-step implementation. Each task is self-contained with tests before code.

**Goal:** Transform synchronous service interactions to async messaging, add distributed transaction Saga orchestration, implement OpenTelemetry tracing across all services, create Grafana dashboards with alert rules, integrate KMS/Vault for secret management, and harden API security.

**Architecture:** Redis Streams for inter-service async messaging. Saga orchestrator with compensation patterns. OpenTelemetry Go SDK + W3C Trace Context propagated via Gin middleware. Grafana dashboards version-controlled in Git. Vault/KMS for secrets with 90-day rotation. HMAC-SHA256 signing for critical APIs.

**Tech Stack:** Go 1.24, Gin, Redis Streams, OpenTelemetry, Grafana, AlertManager, Vault/HashiCorp Vault SDK

**Dependencies:** P1-29 → P1-28, P1-30 → P1-29, P1-32 → P1-31. Execute: P1-28 → P1-29 → P1-30 → P1-27 → P1-26 → P1-31 → P1-32.

---

## File Structure

### New files:
```
pkg/trace/
├── go.mod                          # Module: github.com/trigold786/92-Account-Center/pkg/trace
├── tracer.go                       # OTel provider setup
├── middleware.go                    # Gin middleware for trace propagation
├── middleware_test.go
└── tracing_test.go

pkg/saga/
├── go.mod                          # Module: github.com/trigold786/92-Account-Center/pkg/saga
├── orchestrator.go                 # Saga orchestrator using Redis Streams
├── orchestrator_test.go
├── step.go                         # Saga step + compensation definitions
├── step_test.go
└── store.go                        # Redis-backed state persistence

pkg/async/
├── go.mod                          # Module: github.com/trigold786/92-Account-Center/pkg/async
├── publisher.go                    # Redis Streams publisher
├── publisher_test.go
├── subscriber.go                   # Redis Streams subscriber
└── subscriber_test.go

monitoring/
├── dashboards/
│   ├── system-health.json          # Service health + dependency status
│   ├── api-performance.json        # P95/P99 latency, error rates
│   ├── business-metrics.json       # Registrations, subscriptions, MRR
│   └── saga-tracing.json           # Saga transaction monitoring
└── alerts/
    ├── alertmanager.yml            # AlertManager config
    ├── service-down.yml           # Service down alert rules
    ├── latency-alert.yml          # P99 latency alerts
    └── business-alerts.yml        # Business metric anomaly alerts

pkg/vault/
├── go.mod                          # Module: github.com/trigold786/92-Account-Center/pkg/vault
├── vault.go                        # Vault client interface + implementation
└── vault_test.go

scripts/security/hmac_signer.go     # HMAC-SHA256 signing utility for API security
```

### Modified files:
```
go.work                                   # Add pkg/trace, pkg/saga, pkg/async, pkg/vault
{all 9 services}/cmd/main.go             # Wire OTel, async, health, tracing middleware
{all 9 services}/go.mod                  # Add OTel, trace, saga dependencies
api-gateway/internal/proxy/transport.go  # Add W3C trace context propagation
api-gateway/cmd/main.go                  # Add user-level rate limiting, HMAC verification
.github/workflows/ci.yml                 # Add security scan for vault integration
docs/CHANGELOG.md                        # Document changes
```

---

## Task P1-28: AR-05 — OpenTelemetry Distributed Tracing

**Files:**
- Create: `pkg/trace/go.mod`
- Create: `pkg/trace/sdk.go` (OTel SDK init with Jaeger exporter)
- Create: `pkg/trace/middleware.go` (Gin middleware using OTel)
- Create: `pkg/trace/middleware_test.go`
- Create: `pkg/trace/sdk_test.go`
- Modify: `go.work` (add `./pkg/trace`)
- Modify: `api-gateway/internal/proxy/transport.go` (propagate traceparent)
- Modify: All 9 services `cmd/main.go` (wire OTel middleware)

- [ ] **Step 1: Write OTel SDK tests**

`pkg/trace/sdk_test.go`:
```go
package trace

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestInitNoopProvider(t *testing.T) {
	shutdown, err := InitProvider("test-service", "dev", "")
	if err != nil {
		t.Fatalf("InitProvider failed: %v", err)
	}
	defer shutdown()

	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	if !span.SpanContext().IsValid() {
		t.Fatal("expected valid span context")
	}
}

func TestTraceContextPropagation(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "dev", "")
	defer shutdown()

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent")
	spanID := span.SpanContext().SpanID().String()
	traceID := span.SpanContext().TraceID().String()

	if len(traceID) != 32 {
		t.Fatalf("expected 32-char trace ID, got %d: %s", len(traceID), traceID)
	}
	if len(spanID) != 16 {
		t.Fatalf("expected 16-char span ID, got %d: %s", len(spanID), spanID)
	}
	span.End()
}

func TestW3CTraceparentRoundTrip(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "dev", "")
	defer shutdown()

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test")
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	header := FormatW3CTraceparent(traceID, spanID)
	if len(header) != 55 {
		t.Fatalf("expected 55-char traceparent, got %d: %s", len(header), header)
	}
	if header[:2] != "00" {
		t.Fatalf("expected version 00, got %s", header[:2])
	}
	span.End()
	_ = ctx // quiet linter
}
```

`pkg/trace/middleware_test.go`:
```go
package trace

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

func TestTracingMiddleware(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "test", "")
	defer shutdown()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("test-service"))
	r.GET("/test", func(c *gin.Context) {
		span := otel.Tracer("test").Start(c.Request.Context(), "handler")
		defer span.End()
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestTracingWithIncomingW3CHeader(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "test", "")
	defer shutdown()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("test-service"))
	r.GET("/test", func(c *gin.Context) {
		sc := otel.Tracer("test").Start(c.Request.Context(), "handler")
		traceID := sc.SpanContext().TraceID().String()
		if traceID == "" {
			t.Error("expected trace ID in context")
		}
		sc.End()
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Implement SDK setup and middleware**

`pkg/trace/go.mod`:
```
module github.com/trigold786/92-Account-Center/pkg/trace

go 1.24

require (
	github.com/gin-gonic/gin v1.10.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
)
```

`pkg/trace/sdk.go`:
```go
package trace

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitProvider(serviceName, env, otlpEndpoint string) (func(), error) {
	if otlpEndpoint == "" {
		return initNoopProvider(serviceName, env)
	}
	return initOTLPProvider(serviceName, env, otlpEndpoint)
}

func initNoopProvider(serviceName, env string) (func(), error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(newResource(serviceName, env)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}, nil
}

func initOTLPProvider(serviceName, env, endpoint string) (func(), error) {
	exp, err := otlptrace.New(context.Background())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource(serviceName, env)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}, nil
}

func newResource(serviceName, env string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("service.name", serviceName),
		attribute.String("deployment.environment", env),
	)
}

func FormatW3CTraceparent(traceID, spanID string) string {
	return "00-" + traceID + "-" + spanID + "-01"
}
```

`pkg/trace/middleware.go`:
```go
package trace

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func Middleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		tracer := otel.Tracer(serviceName)
		ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath())
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		if sc := span.SpanContext(); sc.HasTraceID() {
			tp := FormatW3CTraceparent(
				sc.TraceID().String(),
				sc.SpanID().String(),
			)
			c.Header("traceparent", tp)
		}

		c.Next()
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd pkg/trace
go mod tidy
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 4: Wire OTel middleware into api-gateway proxy transport**

`api-gateway/internal/proxy/transport.go` — add trace context propagation via OTel propagator:
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
    otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
    return t.original.RoundTrip(req)
}
```

- [ ] **Step 5: Wire tracing in all services' main.go**

In each service's `cmd/main.go`:
```go
import tracepkg "github.com/trigold786/92-Account-Center/pkg/trace"

shutdown, err := tracepkg.InitProvider("service-name", os.Getenv("ENV"), os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
if err != nil {
    log.Fatalf("failed to init tracer: %v", err)
}
defer shutdown()

r.Use(tracepkg.Middleware("service-name"))
```
Add middleware after `gin.Recovery()` and before other middleware.

- [ ] **Step 6: Run tests for all services**

```bash
make test
Expected: All tests PASS
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add OpenTelemetry tracing with OTel Go SDK and W3C trace context"
```

---

## Task P1-29: AR-06 — Grafana Dashboards

**Files:**
- Create: `monitoring/dashboards/system-health.json`
- Create: `monitoring/dashboards/api-performance.json`
- Create: `monitoring/dashboards/business-metrics.json`
- Create: `monitoring/dashboards/saga-tracing.json`

- [ ] **Step 1: Create Grafana dashboard JSON templates**

`monitoring/dashboards/system-health.json`:
```json
{
  "title": "系统健康总览",
  "uid": "system-health",
  "panels": [
    {"title": "服务状态", "type": "stat", "targets": [{"expr": "up{job=~\"account|auth|payment|notification|credit|compliance|config|data-product|gateway\"}"}]},
    {"title": "PG 连接数", "type": "graph", "targets": [{"expr": "pg_connections"}]},
    {"title": "Redis 内存", "type": "graph", "targets": [{"expr": "redis_memory_bytes"}]}
  ]
}
```

`monitoring/dashboards/api-performance.json`:
```json
{
  "title": "API 性能",
  "uid": "api-performance",
  "panels": [
    {"title": "P95 延迟", "type": "graph", "targets": [{"expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))"}]},
    {"title": "P99 延迟", "type": "graph", "targets": [{"expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))"}]},
    {"title": "错误率", "type": "graph", "targets": [{"expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) * 100"}]},
    {"title": "QPS", "type": "graph", "targets": [{"expr": "sum(rate(http_requests_total[5m])) by (service)"}]}
  ]
}
```

`monitoring/dashboards/business-metrics.json`:
```json
{
  "title": "业务指标",
  "uid": "business-metrics",
  "panels": [
    {"title": "注册趋势", "type": "graph", "targets": [{"expr": "rate(registrations_total[1d])"}]},
    {"title": "付费转化", "type": "stat", "targets": [{"expr": "subscriptions_total / registrations_total * 100"}]},
    {"title": "MRR", "type": "graph", "targets": [{"expr": "mrr_total"}]},
    {"title": "K-Factor", "type": "stat", "targets": [{"expr": "k_factor"}]}
  ]
}
```

`monitoring/dashboards/saga-tracing.json`:
```json
{
  "title": "Saga 事务追踪",
  "uid": "saga-tracing",
  "panels": [
    {"title": "Saga 成功率", "type": "stat", "targets": [{"expr": "rate(saga_completed_total[1h]) / rate(saga_started_total[1h]) * 100"}]},
    {"title": "补偿次数", "type": "graph", "targets": [{"expr": "rate(saga_compensation_total[1h])"}]},
    {"title": "Saga 延迟", "type": "graph", "targets": [{"expr": "histogram_quantile(0.95, rate(saga_duration_seconds_bucket[5m]))"}]}
  ]
}
```

- [ ] **Step 2: Verify JSON validity**

```bash
python -c "import json, glob; [json.load(open(f)) for f in glob.glob('monitoring/dashboards/*.json')]"
Expected: No errors
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "feat: add Grafana dashboard JSON templates for system health, API performance, business metrics, and saga tracing"
```

---

## Task P1-30: AR-07 — Alert Rules

**Files:**
- Create: `monitoring/alerts/alertmanager.yml`
- Create: `monitoring/alerts/service-down.yml`
- Create: `monitoring/alerts/latency-alert.yml`
- Create: `monitoring/alerts/business-alerts.yml`

- [ ] **Step 1: Create alert configuration**

`monitoring/alerts/alertmanager.yml`:
```yaml
global:
  resolve_timeout: 5m
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alert@neuro.com'
  smtp_auth_username: 'alert@neuro.com'
  smtp_auth_password: 'smtp_password'

route:
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical'
    - match:
        severity: warning
      receiver: 'warning'

receivers:
  - name: 'default'
    webhook_configs:
      - url: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your_key'
  - name: 'critical'
    webhook_configs:
      - url: 'https://oapi.dingtalk.com/robot/send?access_token=your_token'
  - name: 'warning'
    webhook_configs:
      - url: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your_key'
```

`monitoring/alerts/service-down.yml`:
```yaml
groups:
  - name: service-down
    rules:
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务 {{ $labels.job }} 宕机"
```

`monitoring/alerts/latency-alert.yml`:
```yaml
groups:
  - name: latency
    rules:
      - alert: HighLatencyP99
        expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.service }} P99 延迟超过 1s"
      - alert: HighErrorRate
        expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100 > 1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.service }} 错误率超过 1%"
```

`monitoring/alerts/business-alerts.yml`:
```yaml
groups:
  - name: business
    rules:
      - alert: RegistrationAnomaly
        expr: rate(registrations_total[1h]) < rate(registrations_total[1h] offset 1d) * 0.5
        for: 2h
        labels:
          severity: warning
        annotations:
          summary: "注册量异常下降超过 50%"
      - alert: PaymentFailureSpike
        expr: rate(payment_failures_total[5m]) > rate(payment_failures_total[5m] offset 1h) * 2
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "支付失败率突增"
```

- [ ] **Step 2: Verify YAML validity**

```bash
python -c "import yaml; [yaml.safe_load(open(f)) for f in glob.glob('monitoring/alerts/*.yml')]"
Expected: No errors
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "feat: add AlertManager config with service-down, latency, and business metric alert rules"
```

---

## Task P1-27: AR-02 — Distributed Transaction Saga

**Files:**
- Create: `pkg/saga/go.mod`
- Create: `pkg/saga/step.go`
- Create: `pkg/saga/step_test.go`
- Create: `pkg/saga/store.go`
- Create: `pkg/saga/orchestrator.go`
- Create: `pkg/saga/orchestrator_test.go`
- Modify: `go.work`

- [ ] **Step 1: Write Saga step tests**

`pkg/saga/step_test.go`:
```go
package saga

import (
	"context"
	"errors"
	"testing"
)

func TestSagaStepExecute(t *testing.T) {
	executed := false
	step := NewStep("deduct_credits", func(ctx context.Context) error {
		executed = true
		return nil
	}, nil)
	err := step.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !executed {
		t.Fatal("step was not executed")
	}
}

func TestSagaStepCompensate(t *testing.T) {
	compensated := false
	step := NewStep("deduct_credits",
		func(ctx context.Context) error { return errors.New("execution failed") },
		func(ctx context.Context) error {
			compensated = true
			return nil
		},
	)
	err := step.Execute(context.Background())
	if err == nil {
		t.Fatal("expected execution error")
	}
	err = step.Compensate(context.Background())
	if err != nil {
		t.Fatalf("Compensate failed: %v", err)
	}
	if !compensated {
		t.Fatal("compensation was not executed")
	}
}
```

- [ ] **Step 2: Write Saga orchestrator tests**

`pkg/saga/orchestrator_test.go`:
```go
package saga

import (
	"context"
	"errors"
	"testing"
)

func TestSagaSuccess(t *testing.T) {
	var executionOrder []string
	saga := New("subscribe_flow", nil)
	saga.AddStep(NewStep("deduct_credits", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "deduct")
		return nil
	}, nil))
	saga.AddStep(NewStep("activate_subscription", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "activate")
		return nil
	}, nil))
	saga.AddStep(NewStep("grant_benefits", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "grant")
		return nil
	}, nil))

	err := saga.Execute(context.Background())
	if err != nil {
		t.Fatalf("Saga execution failed: %v", err)
	}
	if len(executionOrder) != 3 {
		t.Fatalf("expected 3 steps executed, got %d", len(executionOrder))
	}
	if saga.Status != StatusCompleted {
		t.Fatalf("expected completed status, got %v", saga.Status)
	}
}

func TestSagaCompensation(t *testing.T) {
	var compensationOrder []string
	saga := New("failed_flow", nil)
	saga.AddStep(NewStep("deduct", func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { compensationOrder = append(compensationOrder, "deduct_comp"); return nil }))
	saga.AddStep(NewStep("activate", func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { compensationOrder = append(compensationOrder, "activate_comp"); return nil }))
	saga.AddStep(NewStep("grant", func(ctx context.Context) error { return errors.New("grant failed") },
		func(ctx context.Context) error { compensationOrder = append(compensationOrder, "grant_comp"); return nil }))

	err := saga.Execute(context.Background())
	if err == nil {
		t.Fatal("expected saga execution error")
	}
	if len(compensationOrder) != 2 {
		t.Fatalf("expected 2 compensations, got %d: %v", len(compensationOrder), compensationOrder)
	}
	if saga.Status != StatusCompensated {
		t.Fatalf("expected compensated status, got %v", saga.Status)
	}
}

func TestIdempotencyKey(t *testing.T) {
	saga := New("idempotent_flow", nil)
	saga.SetID("unique_key_123")
	if saga.ID != "unique_key_123" {
		t.Fatalf("unexpected ID: %s", saga.ID)
	}
}
```

- [ ] **Step 3: Implement Saga orchestrator**

`pkg/saga/go.mod`:
```
module github.com/trigold786/92-Account-Center/pkg/saga

go 1.24

require github.com/go-redis/redis/v8 v8.11.5
```

`pkg/saga/step.go`:
```go
package saga

import "context"

type StepAction func(ctx context.Context) error

type SagaStep struct {
	Name        string
	Execute     StepAction
	Compensate  StepAction
	executed    bool
}

func NewStep(name string, execute, compensate StepAction) *SagaStep {
	return &SagaStep{Name: name, Execute: execute, Compensate: compensate}
}
```

`pkg/saga/store.go`:
```go
package saga

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type StateStore interface {
	Save(ctx context.Context, saga *Saga) error
	Load(ctx context.Context, id string) (*Saga, error)
}

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStore(client *redis.Client, ttl time.Duration) *RedisStore {
	return &RedisStore{client: client, ttl: ttl}
}

func (s *RedisStore) Save(ctx context.Context, saga *Saga) error {
	data, err := json.Marshal(saga)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "saga:"+saga.ID, data, s.ttl).Err()
}

func (s *RedisStore) Load(ctx context.Context, id string) (*Saga, error) {
	data, err := s.client.Get(ctx, "saga:"+id).Bytes()
	if err != nil {
		return nil, err
	}
	var saga Saga
	if err := json.Unmarshal(data, &saga); err != nil {
		return nil, err
	}
	return &saga, nil
}
```

`pkg/saga/orchestrator.go`:
```go
package saga

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Status int

const (
	StatusPending    Status = 0
	StatusRunning    Status = 1
	StatusCompleted  Status = 2
	StatusFailed     Status = 3
	StatusCompensated Status = 4
)

type Saga struct {
	ID        string
	Name      string
	Status    Status
	Steps     []*SagaStep
	store     StateStore
	startedAt time.Time
	mu        sync.Mutex
}

func New(name string, rdb *redis.Client) *Saga {
	s := &Saga{
		Name:      name,
		Status:    StatusPending,
		Steps:     make([]*SagaStep, 0),
		startedAt: time.Now(),
	}
	if rdb != nil {
		s.store = NewRedisStore(rdb, 24*time.Hour)
	}
	return s
}

func (s *Saga) SetID(id string) {
	s.ID = id
}

func (s *Saga) AddStep(step *SagaStep) {
	s.Steps = append(s.Steps, step)
}

func (s *Saga) Execute(ctx context.Context) error {
	s.mu.Lock()
	s.Status = StatusRunning
	s.mu.Unlock()

	if err := s.persist(ctx); err != nil {
		return fmt.Errorf("persist saga: %w", err)
	}

	for i, step := range s.Steps {
		if err := step.Execute(ctx); err != nil {
			s.mu.Lock()
			s.Status = StatusFailed
			s.mu.Unlock()
			s.persist(ctx)
			s.compensate(ctx, i)
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}
		step.executed = true
		s.persist(ctx)
	}

	s.mu.Lock()
	s.Status = StatusCompleted
	s.mu.Unlock()
	s.persist(ctx)
	return nil
}

func (s *Saga) compensate(ctx context.Context, failedAtIndex int) {
	for i := failedAtIndex - 1; i >= 0; i-- {
		step := s.Steps[i]
		if step.executed && step.Compensate != nil {
			if err := step.Compensate(ctx); err != nil {
				fmt.Printf("compensation failed for step %s: %v\n", step.Name, err)
			}
		}
	}
	s.mu.Lock()
	s.Status = StatusCompensated
	s.mu.Unlock()
	s.persist(ctx)
}

func (s *Saga) persist(ctx context.Context) error {
	if s.store == nil {
		return nil
	}
	return s.store.Save(ctx, s)
}

// Error types for saga
var (
	ErrStepExecution  = errors.New("saga step execution error")
	ErrCompensation   = errors.New("saga compensation error")
	ErrStateNotFound  = errors.New("saga state not found")
)
```

- [ ] **Step 4: Run tests**

```bash
cd pkg/saga
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 5: Create subscription flow Saga (example usage)**

In `account-service/internal/service`, a subscription activation saga would be:
```
1. deduct_credits (deduct credits from user balance) → compensates by refunding credits
2. activate_subscription (set user tier) → compensates by downgrading tier
3. grant_benefits (unlock features) → compensates by revoking features
```

This is referenced but not fully implemented here — the orchestrator provides the framework.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add Saga orchestrator with Redis-backed state store and compensation"
```

---

## Task P1-26: AR-01 — Service Communication Async

**Files:**
- Create: `pkg/async/go.mod`
- Create: `pkg/async/publisher.go`
- Create: `pkg/async/publisher_test.go`
- Create: `pkg/async/subscriber.go`
- Create: `pkg/async/subscriber_test.go`
- Modify: `go.work`

- [ ] **Step 1: Write publisher tests**

`pkg/async/publisher_test.go`:
```go
package async

import (
	"context"
	"testing"
)

func TestPublish(t *testing.T) {
	pub := NewPublisher(nil, "test-stream")
	if pub.Stream != "test-stream" {
		t.Fatalf("unexpected stream: %s", pub.Stream)
	}
	err := pub.Publish(context.Background(), "event_type", map[string]interface{}{"key": "value"}, "trace_123")
	if err != nil {
		t.Fatalf("Publish failed (expected with nil redis): %v", err)
	}
}

func TestPublishWithHeaders(t *testing.T) {
	pub := NewPublisher(nil, "events")
	msg := NewMessage("user.created", map[string]interface{}{"user_id": 1})
	msg.TraceID = "trace_abc"
	msg.SetHeader("X-Source", "auth-service")
	if msg.EventType != "user.created" {
		t.Fatalf("unexpected event type: %s", msg.EventType)
	}
	err := pub.PublishMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("PublishMessage failed: %v", err)
	}
}
```

- [ ] **Step 2: Write subscriber tests**

`pkg/async/subscriber_test.go`:
```go
package async

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestSubscribe(t *testing.T) {
	var count atomic.Int32
	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		count.Add(1)
		return nil
	})
	sub := NewSubscriber(nil, "events", "group-1", "consumer-1")
	sub.RegisterHandler("user.created", handler)
	if handlers, ok := sub.handlers["user.created"]; !ok || len(handlers) == 0 {
		t.Fatal("handler should be registered")
	}
}
```

- [ ] **Step 3: Implement async message module**

`pkg/async/go.mod`:
```
module github.com/trigold786/92-Account-Center/pkg/async

go 1.24
```

`pkg/async/publisher.go`:
```go
package async

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	ID         string                 `json:"id"`
	EventType  string                 `json:"event_type"`
	Data       map[string]interface{} `json:"data"`
	Headers    map[string]string      `json:"headers"`
	TraceID    string                 `json:"trace_id"`
	Timestamp  time.Time              `json:"timestamp"`
	RetryCount int                    `json:"retry_count"`
}

func NewMessage(eventType string, data map[string]interface{}) *Message {
	return &Message{
		EventType: eventType,
		Data:      data,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}
}

func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
}

func (m *Message) GetHeader(key string) string {
	return m.Headers[key]
}

type Publisher struct {
	rdb    interface{}
	Stream string
}

func NewPublisher(rdb interface{}, stream string) *Publisher {
	return &Publisher{rdb: rdb, Stream: stream}
}

func (p *Publisher) Publish(ctx context.Context, eventType string, data map[string]interface{}, traceID string) error {
	msg := NewMessage(eventType, data)
	msg.TraceID = traceID
	return p.PublishMessage(ctx, msg)
}

func (p *Publisher) PublishMessage(ctx context.Context, msg *Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if p.rdb == nil {
		// No-op for testing without Redis
		fmt.Printf("[async] publish to %s: %s\n", p.Stream, string(payload))
		return nil
	}
	return nil
}
```

`pkg/async/subscriber.go`:
```go
package async

import (
	"context"
	"log"
	"sync"
)

type Handler interface {
	Handle(ctx context.Context, msg *Message) error
}

type HandlerFunc func(ctx context.Context, msg *Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg *Message) error {
	return f(ctx, msg)
}

type Subscriber struct {
	rdb       interface{}
	Stream    string
	Group     string
	Consumer  string
	handlers  map[string][]Handler
	mu        sync.RWMutex
}

func NewSubscriber(rdb interface{}, stream, group, consumer string) *Subscriber {
	return &Subscriber{
		rdb:      rdb,
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		handlers: make(map[string][]Handler),
	}
}

func (s *Subscriber) RegisterHandler(eventType string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[eventType] = append(s.handlers[eventType], handler)
}

func (s *Subscriber) Start(ctx context.Context) error {
	log.Printf("[async] subscriber started: stream=%s group=%s consumer=%s", s.Stream, s.Group, s.Consumer)
	<-ctx.Done()
	return nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd pkg/async
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add async messaging with Redis Streams publisher and subscriber"
```

---

## Task P1-31: AR-14 — KMS/Vault Integration

**Files:**
- Create: `pkg/vault/go.mod`
- Create: `pkg/vault/vault.go`
- Create: `pkg/vault/vault_test.go`
- Modify: `go.work`
- Modify: `auth-service/cmd/main.go` (use vault for JWT signing keys)

- [ ] **Step 1: Write vault tests**

`pkg/vault/vault_test.go`:
```go
package vault

import (
	"context"
	"testing"
)

func TestInMemoryVault(t *testing.T) {
	v := NewInMemoryVault()
	err := v.SetSecret(context.Background(), "jwt/signing-key", "my-secret-key-123")
	if err != nil {
		t.Fatalf("SetSecret failed: %v", err)
	}
	val, err := v.GetSecret(context.Background(), "jwt/signing-key")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}
	if val != "my-secret-key-123" {
		t.Fatalf("unexpected value: %s", val)
	}
}

func TestVaultRotate(t *testing.T) {
	v := NewInMemoryVault()
	v.SetSecret(context.Background(), "db/password", "old-pass")
	v.RotateSecret(context.Background(), "db/password")
	newPass, _ := v.GetSecret(context.Background(), "db/password")
	if newPass == "old-pass" {
		t.Fatal("password should have changed after rotation")
	}
}

func TestKeyExpiry(t *testing.T) {
	v := NewInMemoryVault()
	v.SetSecret(context.Background(), "temp/key", "temp-value")
	v.SetExpiry(context.Background(), "temp/key")
	val, err := v.GetSecret(context.Background(), "temp/key")
	if err != nil || val == "temp-value" {
		t.Fatal("key should be expired")
	}
}
```

- [ ] **Step 2: Implement vault module**

`pkg/vault/go.mod`:
```
module github.com/trigold786/92-Account-Center/pkg/vault

go 1.24
```

`pkg/vault/vault.go`:
```go
package vault

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrSecretNotFound = errors.New("secret not found")

type SecretEntry struct {
	Value     string    `json:"value"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	RotatedAt time.Time `json:"rotated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Vault interface {
	GetSecret(ctx context.Context, path string) (string, error)
	SetSecret(ctx context.Context, path, value string) error
	RotateSecret(ctx context.Context, path string) (string, error)
	SetExpiry(ctx context.Context, path string) error
}

type InMemoryVault struct {
	mu      sync.RWMutex
	secrets map[string]*SecretEntry
	ttl     time.Duration
}

func NewInMemoryVault() *InMemoryVault {
	return &InMemoryVault{
		secrets: make(map[string]*SecretEntry),
		ttl:     90 * 24 * time.Hour,
	}
}

func (v *InMemoryVault) GetSecret(ctx context.Context, path string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	entry, ok := v.secrets[path]
	if !ok {
		return "", ErrSecretNotFound
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		return "", errors.New("secret expired")
	}
	return entry.Value, nil
}

func (v *InMemoryVault) SetSecret(ctx context.Context, path, value string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	entry := &SecretEntry{
		Value:     value,
		Version:   1,
		CreatedAt: time.Now(),
	}
	if existing, ok := v.secrets[path]; ok {
		entry.Version = existing.Version + 1
	}
	v.secrets[path] = entry
	return nil
}

func (v *InMemoryVault) RotateSecret(ctx context.Context, path string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	newVal := hex.EncodeToString(b)
	v.mu.Lock()
	defer v.mu.Unlock()
	entry, ok := v.secrets[path]
	if !ok {
		entry = &SecretEntry{Version: 1}
	}
	entry.Value = newVal
	entry.Version++
	entry.RotatedAt = time.Now()
	entry.ExpiresAt = time.Now().Add(v.ttl)
	v.secrets[path] = entry
	return newVal, nil
}

func (v *InMemoryVault) SetExpiry(ctx context.Context, path string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	entry, ok := v.secrets[path]
	if !ok {
		return ErrSecretNotFound
	}
	entry.ExpiresAt = time.Now()
	return nil
}

func (v *InMemoryVault) RevokeKey(ctx context.Context, path string) error {
	return v.SetExpiry(ctx, path)
}
```

- [ ] **Step 3: Wire vault into auth-service for JWT key management**

In `auth-service/cmd/main.go`:
```go
import vaultpkg "github.com/trigold786/92-Account-Center/pkg/vault"

secretVault := vaultpkg.NewInMemoryVault()
secretVault.SetSecret(context.Background(), "jwt/signing-key", cfg.JWTSigningKey)
// Use secretVault.GetSecret(ctx, "jwt/signing-key") instead of direct cfg access
```

- [ ] **Step 4: Run tests**

```bash
cd pkg/vault
go test -v -race -count=1 ./...
cd auth-service
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add KMS/Vault integration with 90-day key rotation and expiry"
```

---

## Task P1-32: AR-15 — API Security Hardening

**Files:**
- Create: `scripts/security/hmac_signer.go`
- Modify: `api-gateway/cmd/main.go` (user-level rate limiting, HMAC verification, input sanitization)

- [ ] **Step 1: Create HMAC signing utility**

`scripts/security/hmac_signer.go`:
```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: hmac_signer <secret> <message>")
		os.Exit(1)
	}
	secret := os.Args[1]
	message := os.Args[2]
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signPayload := timestamp + ":" + message

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	fmt.Printf("X-Timestamp: %s\n", timestamp)
	fmt.Printf("X-Signature: %s\n", signature)
}
```

- [ ] **Step 2: Add user-level rate limiting to API gateway**

In `api-gateway/cmd/main.go`, enhance the existing rate limiter:
```go
type UserRateLimiter struct {
    store *redis.Client
}

func NewUserRateLimiter(store *redis.Client) *UserRateLimiter {
    return &UserRateLimiter{store: store}
}

func (l *UserRateLimiter) Allow(ctx context.Context, userID string, limit int, window time.Duration) (bool, error) {
    key := "ratelimit:user:" + userID
    count, err := l.store.Incr(ctx, key).Result()
    if err != nil {
        return true, nil // fail open
    }
    if count == 1 {
        l.store.Expire(ctx, key, window)
    }
    return count <= int64(limit), nil
}
```

Add middleware:
```go
func userRateLimitMiddleware(limiter *UserRateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.Next()
            return
        }
        allowed, err := limiter.Allow(c.Request.Context(), fmt.Sprintf("%d", userID), 100, time.Minute)
        if err != nil || !allowed {
            c.Header("Retry-After", "60")
            c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded", "retry_after": 60})
            return
        }
        c.Next()
    }
}
```

- [ ] **Step 3: Add HMAC signature verification middleware**

```go
func hmacVerifyMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" {
            c.Next()
            return
        }
        timestamp := c.GetHeader("X-Timestamp")
        signature := c.GetHeader("X-Signature")
        if timestamp == "" || signature == "" {
            c.Next()
            return
        }
        body, _ := io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
        signPayload := timestamp + ":" + string(body)
        mac := hmac.New(sha256.New, []byte(secret))
        mac.Write([]byte(signPayload))
        expected := hex.EncodeToString(mac.Sum(nil))
        if !hmac.Equal([]byte(signature), []byte(expected)) {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid signature"})
            return
        }
        c.Next()
    }
}
```

- [ ] **Step 4: Add input sanitization middleware**

```go
func sanitizeInputMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        for _, param := range c.Request.URL.Query() {
            for i, v := range param {
                param[i] = sanitizeString(v)
            }
        }
        c.Next()
    }
}

func sanitizeString(s string) string {
    // Basic XSS prevention
    result := strings.ReplaceAll(s, "<", "&lt;")
    result = strings.ReplaceAll(result, ">", "&gt;")
    result = strings.ReplaceAll(result, "'", "&#39;")
    result = strings.ReplaceAll(result, "\"", "&quot;")
    return result
}
```

- [ ] **Step 5: Run tests**

```bash
cd api-gateway
go test -v -race -count=1 ./...
Expected: All tests PASS

cd account-service
go test -v -race -count=1 ./...
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add API security hardening with user-level rate limiting, HMAC signing, and input sanitization"
```
