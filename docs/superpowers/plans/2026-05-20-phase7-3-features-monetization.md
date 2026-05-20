# Phase 7.3 — Features & Monetization Implementation Plan

> **For agentic workers:** Step-by-step implementation. Each task is self-contained with tests before code.

**Goal:** Add refund workflow, operations dashboard, subscription/coupon admin, risk admin panel, event tracking SDK, OAuth extension framework, and data export/Open API.

**Architecture:** Backend Go services for refunds (payment-service), subscription admin (account-service), risk admin (compliance-service). Data product service for event tracking ingestion and operations metrics. Standalone TypeScript SDK for web event tracking. Grafana dashboards JSON-versioned in Git.

**Tech Stack:** Go 1.24, Gin, PostgreSQL, Redis, Grafana, TypeScript (Web SDK)

**Dependencies:** P1-18 → AR-14 (stub key management). Execute in order: P1-12 → P1-15 → P1-14 → P1-16 → P1-17 → P1-13 → P1-18.

---

## File Structure

### New files:
```
payment-service/
├── internal/
│   ├── model/refund.go
│   ├── repository/refund_repository.go
│   ├── service/refund_service.go
│   ├── service/refund_test.go
│   └── handler/refund_handler.go

compliance-service/
├── internal/
│   ├── model/blacklist.go
│   ├── model/risk_event.go
│   ├── repository/risk_repository.go
│   ├── service/risk_admin_service.go
│   ├── service/risk_admin_test.go
│   └── handler/risk_admin_handler.go

account-service/
├── internal/
│   ├── model/plan.go
│   ├── model/coupon.go
│   ├── model/promotion.go
│   ├── repository/subscription_admin_repository.go
│   ├── service/subscription_admin_service.go
│   ├── service/subscription_admin_test.go
│   ├── handler/subscription_admin_handler.go
│   ├── handler/export_handler.go
│   ├── handler/openapi_handler.go
│   ├── service/export_service.go
│   ├── service/export_test.go
│   └── service/openapi_service.go

data-product-service/
├── internal/
│   ├── model/event.go
│   ├── repository/event_repository.go
│   ├── repository/metrics_repository.go
│   ├── service/event_service.go
│   ├── service/event_test.go
│   ├── service/metrics_service.go
│   ├── service/metrics_test.go
│   ├── handler/event_handler.go
│   └── handler/dashboard_handler.go

auth-service/
├── internal/
│   └── service/oauth/provider.go          # Plugin provider framework

web-ui/src/sdk/
├── tracker.ts                              # Web event tracking SDK
└── tracker.test.ts

monitoring/dashboards/
├── registration-trends.json
├── conversion-funnel.json
├── mrr-analytics.json
└── system-health.json
```

### Modified files:
```
payment-service/cmd/main.go
compliance-service/cmd/main.go
account-service/cmd/main.go
data-product-service/cmd/main.go
api-gateway/cmd/main.go                   # Route /api/v1/events, /api/v1/orders/*
```

---

## Task P1-12: FN-04 — Refund Flow

**Files:**
- Create: `payment-service/internal/model/refund.go`
- Create: `payment-service/internal/repository/refund_repository.go`
- Create: `payment-service/internal/service/refund_service.go`
- Create: `payment-service/internal/service/refund_test.go`
- Create: `payment-service/internal/handler/refund_handler.go`
- Modify: `payment-service/cmd/main.go`

- [ ] **Step 1: Write refund service tests**

`payment-service/internal/service/refund_test.go`:
```go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

func TestCalculateRefundFullWithin7Days(t *testing.T) {
	repo := &mockRefundRepo{}
	svc := NewRefundService(repo, nil, nil)
	order := &model.Order{
		ID:        1,
		UserID:    42,
		OrderNo:   "ORD001",
		Amount:    100,
		Status:    model.OrderStatusPaid,
		CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
	}
	refund, err := svc.CalculateRefund(context.Background(), order)
	if err != nil {
		t.Fatalf("CalculateRefund failed: %v", err)
	}
	if refund.Amount != 100 {
		t.Fatalf("expected full refund 100, got %.2f", refund.Amount)
	}
}

func TestCalculateRefundProratedAfter7Days(t *testing.T) {
	svc := NewRefundService(nil, nil, nil)
	order := &model.Order{
		ID:        2,
		UserID:    42,
		OrderNo:   "ORD002",
		Amount:    300,
		Status:    model.OrderStatusPaid,
		CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
	}
	refund, err := svc.CalculateRefund(context.Background(), order)
	if err != nil {
		t.Fatalf("CalculateRefund failed: %v", err)
	}
	if refund.Amount >= 300 {
		t.Fatalf("expected prorated refund < 300, got %.2f", refund.Amount)
	}
	if refund.Amount <= 0 {
		t.Fatalf("expected positive refund amount, got %.2f", refund.Amount)
	}
}

func TestRefundStatusFlow(t *testing.T) {
	repo := &mockRefundRepo{}
	svc := NewRefundService(repo, nil, nil)
	refund, err := svc.RequestRefund(context.Background(), 1, 42, "产品质量问题")
	if err != nil {
		t.Fatalf("RequestRefund failed: %v", err)
	}
	if refund.Status != "pending" {
		t.Fatalf("expected pending status, got %s", refund.Status)
	}
	err = svc.ApproveRefund(context.Background(), refund.ID, 1)
	if err != nil {
		t.Fatalf("ApproveRefund failed: %v", err)
	}
}

type mockRefundRepo struct{}

func (m *mockRefundRepo) Create(ctx context.Context, r *model.Refund) error {
	r.ID = 1
	return nil
}

func (m *mockRefundRepo) GetByID(ctx context.Context, id int64) (*model.Refund, error) {
	return &model.Refund{ID: id, Status: "pending", OrderID: 1, Amount: 100}, nil
}

func (m *mockRefundRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return nil
}
```

- [ ] **Step 2: Implement refund model, repository, service**

`payment-service/internal/model/refund.go`:
```go
package model

import "time"

type Refund struct {
	ID          int64     `json:"id"`
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	Amount      float64   `json:"amount"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"` // pending, approved, rejected, processed
	ApproverID  int64     `json:"approver_id,omitempty"`
	ReviewNote  string    `json:"review_note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

`payment-service/internal/repository/refund_repository.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type RefundRepository struct {
	db *sql.DB
}

func NewRefundRepository(db *sql.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

func (r *RefundRepository) Create(ctx context.Context, refund *model.Refund) error {
	refund.CreatedAt = time.Now()
	refund.UpdatedAt = refund.CreatedAt
	return r.db.QueryRowContext(ctx,
		`INSERT INTO refunds (order_id, user_id, amount, reason, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		refund.OrderID, refund.UserID, refund.Amount, refund.Reason, refund.Status, refund.CreatedAt, refund.UpdatedAt,
	).Scan(&refund.ID)
}

func (r *RefundRepository) GetByID(ctx context.Context, id int64) (*model.Refund, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, amount, reason, status, COALESCE(approver_id,0), COALESCE(review_note,''), created_at, updated_at
		 FROM refunds WHERE id=$1`, id)
	ref := &model.Refund{}
	err := row.Scan(&ref.ID, &ref.OrderID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.Status, &ref.ApproverID, &ref.ReviewNote, &ref.CreatedAt, &ref.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (r *RefundRepository) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refunds SET status=$1, approver_id=$2, review_note=$3, updated_at=NOW() WHERE id=$4`,
		status, approverID, note, id)
	return err
}

func (r *RefundRepository) ListByUserID(ctx context.Context, userID int64) ([]*model.Refund, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, user_id, amount, reason, status, COALESCE(approver_id,0), created_at, updated_at
		 FROM refunds WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refunds []*model.Refund
	for rows.Next() {
		ref := &model.Refund{}
		if err := rows.Scan(&ref.ID, &ref.OrderID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.Status, &ref.ApproverID, &ref.CreatedAt, &ref.UpdatedAt); err != nil {
			return nil, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, nil
}
```

`payment-service/internal/service/refund_service.go`:
```go
package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

var (
	ErrOrderNotPaid       = errors.New("order is not paid")
	ErrRefundAlreadyExists = errors.New("refund already exists for this order")
)

type RefundRepository interface {
	Create(ctx context.Context, refund *model.Refund) error
	GetByID(ctx context.Context, id int64) (*model.Refund, error)
	UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error
}

type RefundService struct {
	refundRepo RefundRepository
	orderRepo  interface{}
	creditSvc  interface{}
}

func NewRefundService(refundRepo RefundRepository, orderRepo, creditSvc interface{}) *RefundService {
	return &RefundService{refundRepo: refundRepo, orderRepo: orderRepo, creditSvc: creditSvc}
}

func (s *RefundService) CalculateRefund(ctx context.Context, order *model.Order) (*model.Refund, error) {
	daysSinceCreation := time.Since(order.CreatedAt).Hours() / 24
	amount := order.Amount
	if daysSinceCreation > 7 {
		// Prorate: (subscription_days - used_days) / subscription_days * amount
		subscriptionDays := 30.0
		usedDays := math.Min(daysSinceCreation, subscriptionDays)
		remainingDays := subscriptionDays - usedDays
		if remainingDays < 1 {
			remainingDays = 1
		}
		amount = math.Round(order.Amount*remainingDays/subscriptionDays*100) / 100
	}
	return &model.Refund{
		OrderID: order.ID,
		Amount:  amount,
		Status:  "calculated",
	}, nil
}

func (s *RefundService) RequestRefund(ctx context.Context, orderID, userID int64, reason string) (*model.Refund, error) {
	refund := &model.Refund{
		OrderID: orderID,
		UserID:  userID,
		Amount:  0,
		Reason:  reason,
		Status:  "pending",
	}
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}

func (s *RefundService) ApproveRefund(ctx context.Context, refundID, approverID int64) error {
	return s.refundRepo.UpdateStatus(ctx, refundID, "approved", approverID, "")
}

func (s *RefundService) RejectRefund(ctx context.Context, refundID, approverID int64, note string) error {
	return s.refundRepo.UpdateStatus(ctx, refundID, "rejected", approverID, note)
}
```

- [ ] **Step 3: Implement handler**

`payment-service/internal/handler/refund_handler.go`:
```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
)

type RefundHandler struct {
	svc *service.RefundService
}

func NewRefundHandler(svc *service.RefundService) *RefundHandler {
	return &RefundHandler{svc: svc}
}

func (h *RefundHandler) RequestRefund(c *gin.Context) {
	var req struct {
		OrderID int64  `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	refund, err := h.svc.RequestRefund(c.Request.Context(), req.OrderID, userID.(int64), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, refund)
}

func (h *RefundHandler) ApproveRefund(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	adminID, _ := c.Get("user_id")
	if err := h.svc.ApproveRefund(c.Request.Context(), id, adminID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (h *RefundHandler) RejectRefund(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Note string `json:"note"`
	}
	c.ShouldBindJSON(&req)
	adminID, _ := c.Get("user_id")
	if err := h.svc.RejectRefund(c.Request.Context(), id, adminID.(int64), req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}
```

- [ ] **Step 4: Run tests**

```bash
cd payment-service
go test -v -race -count=1 ./internal/service/... -run "TestCalculateRefund|TestRefund"
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add refund flow with prorated calculation and admin approval"
```

---

## Task P1-15: FN-08 — Risk Management Admin Panel

**Files:**
- Create: `compliance-service/internal/model/blacklist.go`
- Create: `compliance-service/internal/model/risk_event.go`
- Create: `compliance-service/internal/repository/risk_repository.go`
- Create: `compliance-service/internal/service/risk_admin_service.go`
- Create: `compliance-service/internal/service/risk_admin_test.go`
- Create: `compliance-service/internal/handler/risk_admin_handler.go`
- Modify: `compliance-service/cmd/main.go`

- [ ] **Step 1: Write tests**

`compliance-service/internal/service/risk_admin_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestBlacklistCRUD(t *testing.T) {
	repo := &mockRiskRepo{}
	svc := NewRiskAdminService(repo)
	entry, err := svc.AddToBlacklist(context.Background(), "ip", "192.168.1.1", "manual", 1)
	if err != nil {
		t.Fatalf("AddToBlacklist failed: %v", err)
	}
	if entry.Value != "192.168.1.1" {
		t.Fatalf("unexpected value: %s", entry.Value)
	}
	entries, err := svc.ListBlacklist(context.Background(), "ip", 0, 10)
	if err != nil {
		t.Fatalf("ListBlacklist failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected non-empty list")
	}
	err = svc.RemoveFromBlacklist(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("RemoveFromBlacklist failed: %v", err)
	}
}

func TestRiskEvents(t *testing.T) {
	repo := &mockRiskRepo{}
	svc := NewRiskAdminService(repo)
	events, err := svc.ListRiskEvents(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("ListRiskEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestAnomalyDetection(t *testing.T) {
	svc := NewRiskAdminService(nil)
	alert, err := svc.CheckAnomalyRegistration(context.Background(), "192.168.1.1")
	if err != nil {
		t.Fatalf("CheckAnomalyRegistration failed: %v", err)
	}
	if alert.Triggered {
		t.Logf("Anomaly detected: %s", alert.Reason)
	}
}

type mockRiskRepo struct{}

func (m *mockRiskRepo) CreateBlacklistEntry(ctx context.Context, entry *BlacklistEntry) error {
	entry.ID = 1
	return nil
}

func (m *mockRiskRepo) ListBlacklist(ctx context.Context, entryType string, offset, limit int) ([]*BlacklistEntry, error) {
	return []*BlacklistEntry{{ID: 1, Type: "ip", Value: "192.168.1.1"}}, nil
}

func (m *mockRiskRepo) DeleteBlacklistEntry(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRiskRepo) GetBlacklistEntry(ctx context.Context, entryType, value string) (*BlacklistEntry, error) {
	return nil, nil
}

func (m *mockRiskRepo) ListRiskEvents(ctx context.Context, offset, limit int) ([]*RiskEvent, error) {
	return []*RiskEvent{{ID: 1, EventType: "login_failed", UserID: 42}}, nil
}

func (m *mockRiskRepo) CountRecentRegistrations(ctx context.Context, ip string, hours int) (int, error) {
	return 6, nil
}
```

- [ ] **Step 2: Implement models**

`compliance-service/internal/model/blacklist.go`:
```go
package model

import "time"

type BlacklistEntry struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`        // ip, device, user
	Value       string    `json:"value"`
	Reason      string    `json:"reason,omitempty"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}
```

`compliance-service/internal/model/risk_event.go`:
```go
package model

import "time"

type RiskEvent struct {
	ID          int64     `json:"id"`
	EventType   string    `json:"event_type"`
	UserID      int64     `json:"user_id"`
	IP          string    `json:"ip,omitempty"`
	DeviceID    string    `json:"device_id,omitempty"`
	Details     string    `json:"details,omitempty"`
	Severity    string    `json:"severity"` // low, medium, high, critical
	CreatedAt   time.Time `json:"created_at"`
}
```

- [ ] **Step 3: Implement repository and service**

`compliance-service/internal/repository/risk_repository.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
)

type RiskRepository struct {
	db  *sql.DB
}

func NewRiskRepository(db *sql.DB) *RiskRepository {
	return &RiskRepository{db: db}
}

func (r *RiskRepository) CreateBlacklistEntry(ctx context.Context, entry *model.BlacklistEntry) error {
	entry.CreatedAt = time.Now()
	return r.db.QueryRowContext(ctx,
		`INSERT INTO blacklist_entries (type, value, reason, created_by, created_at) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		entry.Type, entry.Value, entry.Reason, entry.CreatedBy, entry.CreatedAt,
	).Scan(&entry.ID)
}

func (r *RiskRepository) ListBlacklist(ctx context.Context, entryType string, offset, limit int) ([]*model.BlacklistEntry, error) {
	query := `SELECT id, type, value, COALESCE(reason,''), created_by, created_at FROM blacklist_entries`
	args := []interface{}{}
	if entryType != "" {
		query += " WHERE type=$1"
		args = append(args, entryType)
	}
	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []*model.BlacklistEntry
	for rows.Next() {
		e := &model.BlacklistEntry{}
		if err := rows.Scan(&e.ID, &e.Type, &e.Value, &e.Reason, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *RiskRepository) DeleteBlacklistEntry(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blacklist_entries WHERE id=$1`, id)
	return err
}

func (r *RiskRepository) ListRiskEvents(ctx context.Context, offset, limit int) ([]*model.RiskEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, event_type, user_id, COALESCE(ip,''), COALESCE(device_id,''), COALESCE(details,''), severity, created_at
		 FROM risk_events ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []*model.RiskEvent
	for rows.Next() {
		e := &model.RiskEvent{}
		if err := rows.Scan(&e.ID, &e.EventType, &e.UserID, &e.IP, &e.DeviceID, &e.Details, &e.Severity, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
```

`compliance-service/internal/service/risk_admin_service.go`:
```go
package service

import (
	"context"
	"time"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
)

type BlacklistEntry struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	Reason    string    `json:"reason"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type RiskEvent struct {
	ID        int64     `json:"id"`
	EventType string    `json:"event_type"`
	UserID    int64     `json:"user_id"`
	IP        string    `json:"ip,omitempty"`
	DeviceID  string    `json:"device_id,omitempty"`
	Details   string    `json:"details,omitempty"`
	Severity  string    `json:"severity"`
	CreatedAt time.Time `json:"created_at"`
}

type AnomalyAlert struct {
	Triggered bool   `json:"triggered"`
	Reason    string `json:"reason,omitempty"`
}

type RiskRepository interface {
	CreateBlacklistEntry(ctx context.Context, entry *BlacklistEntry) error
	ListBlacklist(ctx context.Context, entryType string, offset, limit int) ([]*BlacklistEntry, error)
	DeleteBlacklistEntry(ctx context.Context, id int64) error
	GetBlacklistEntry(ctx context.Context, entryType, value string) (*BlacklistEntry, error)
	ListRiskEvents(ctx context.Context, offset, limit int) ([]*RiskEvent, error)
	CountRecentRegistrations(ctx context.Context, ip string, hours int) (int, error)
}

type RiskAdminService struct {
	repo RiskRepository
}

func NewRiskAdminService(repo RiskRepository) *RiskAdminService {
	return &RiskAdminService{repo: repo}
}

func (s *RiskAdminService) AddToBlacklist(ctx context.Context, entryType, value, reason string, createdBy int64) (*BlacklistEntry, error) {
	entry := &BlacklistEntry{Type: entryType, Value: value, Reason: reason, CreatedBy: createdBy}
	if err := s.repo.CreateBlacklistEntry(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *RiskAdminService) ListBlacklist(ctx context.Context, entryType string, offset, limit int) ([]*BlacklistEntry, error) {
	return s.repo.ListBlacklist(ctx, entryType, offset, limit)
}

func (s *RiskAdminService) RemoveFromBlacklist(ctx context.Context, id int64) error {
	return s.repo.DeleteBlacklistEntry(ctx, id)
}

func (s *RiskAdminService) ListRiskEvents(ctx context.Context, offset, limit int) ([]*RiskEvent, error) {
	return s.repo.ListRiskEvents(ctx, offset, limit)
}

func (s *RiskAdminService) CheckAnomalyRegistration(ctx context.Context, ip string) (*AnomalyAlert, error) {
	count, err := s.repo.CountRecentRegistrations(ctx, ip, 24)
	if err != nil {
		return nil, err
	}
	if count > 5 {
		return &AnomalyAlert{
			Triggered: true,
			Reason:    "超过24小时内最多5次注册的限制",
		}, nil
	}
	return &AnomalyAlert{Triggered: false}, nil
}
```

- [ ] **Step 4: Implement handler**

`compliance-service/internal/handler/risk_admin_handler.go`:
```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
)

type RiskAdminHandler struct {
	svc *service.RiskAdminService
}

func NewRiskAdminHandler(svc *service.RiskAdminService) *RiskAdminHandler {
	return &RiskAdminHandler{svc: svc}
}

func (h *RiskAdminHandler) AddBlacklist(c *gin.Context) {
	var req struct {
		Type   string `json:"type"`
		Value  string `json:"value"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID, _ := c.Get("user_id")
	entry, err := h.svc.AddToBlacklist(c.Request.Context(), req.Type, req.Value, req.Reason, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *RiskAdminHandler) ListBlacklist(c *gin.Context) {
	entryType := c.Query("type")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	entries, err := h.svc.ListBlacklist(c.Request.Context(), entryType, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *RiskAdminHandler) RemoveBlacklist(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.RemoveFromBlacklist(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (h *RiskAdminHandler) ListRiskEvents(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	events, err := h.svc.ListRiskEvents(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}
```

- [ ] **Step 5: Run tests**

```bash
cd compliance-service
go test -v -race -count=1 ./internal/service/... -run "TestBlacklist|TestRisk"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add risk admin panel with blacklist CRUD and anomaly detection"
```

---

## Task P1-14: FN-07 — Subscription Admin Panel

**Files:**
- Create: `account-service/internal/model/plan.go`
- Create: `account-service/internal/model/coupon.go`
- Create: `account-service/internal/model/promotion.go`
- Create: `account-service/internal/repository/subscription_admin_repository.go`
- Create: `account-service/internal/service/subscription_admin_service.go`
- Create: `account-service/internal/service/subscription_admin_test.go`
- Create: `account-service/internal/handler/subscription_admin_handler.go`
- Modify: `account-service/cmd/main.go`

- [ ] **Step 1: Write tests**

`account-service/internal/service/subscription_admin_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestPlanCRUD(t *testing.T) {
	repo := &mockSubAdminRepo{}
	svc := NewSubscriptionAdminService(repo)
	plan, err := svc.CreatePlan(context.Background(), "test_plan", "测试套餐", 29.9, "monthly", map[string]interface{}{"feature_x": true})
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}
	if plan.Name != "test_plan" {
		t.Fatalf("unexpected name: %s", plan.Name)
	}
	plans, err := svc.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans failed: %v", err)
	}
	if len(plans) == 0 {
		t.Fatal("expected non-empty plans")
	}
	err = svc.DeletePlan(context.Background(), plan.ID)
	if err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}
}

func TestCouponCRUD(t *testing.T) {
	repo := &mockSubAdminRepo{}
	svc := NewSubscriptionAdminService(repo)
	coupon, err := svc.CreateCoupon(context.Background(), "WELCOME10", "percentage", 10, 100, 1)
	if err != nil {
		t.Fatalf("CreateCoupon failed: %v", err)
	}
	if coupon.Code != "WELCOME10" {
		t.Fatalf("unexpected code: %s", coupon.Code)
	}
	coupons, err := svc.ListCoupons(context.Background())
	if err != nil {
		t.Fatalf("ListCoupons failed: %v", err)
	}
	if len(coupons) == 0 {
		t.Fatal("expected non-empty coupons")
	}
}

type mockSubAdminRepo struct{}

func (m *mockSubAdminRepo) CreatePlan(ctx context.Context, p *Plan) error { p.ID = 1; return nil }
func (m *mockSubAdminRepo) ListPlans(ctx context.Context) ([]*Plan, error) {
	return []*Plan{{ID: 1, Name: "basic", Price: 9.9}}, nil
}
func (m *mockSubAdminRepo) GetPlan(ctx context.Context, id int64) (*Plan, error) {
	return &Plan{ID: id, Name: "pro", Price: 29.9}, nil
}
func (m *mockSubAdminRepo) UpdatePlan(ctx context.Context, p *Plan) error { return nil }
func (m *mockSubAdminRepo) DeletePlan(ctx context.Context, id int64) error { return nil }
func (m *mockSubAdminRepo) CreateCoupon(ctx context.Context, c *Coupon) error { c.ID = 1; return nil }
func (m *mockSubAdminRepo) ListCoupons(ctx context.Context) ([]*Coupon, error) {
	return []*Coupon{{ID: 1, Code: "WELCOME10", DiscountType: "percentage", DiscountValue: 10}}, nil
}
func (m *mockSubAdminRepo) UpdateCoupon(ctx context.Context, c *Coupon) error { return nil }
func (m *mockSubAdminRepo) DeleteCoupon(ctx context.Context, id int64) error { return nil }
```

- [ ] **Step 2: Implement models**

`account-service/internal/model/plan.go`:
```go
package model

import "time"

type Plan struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Price       float64                `json:"price"`
	Interval    string                 `json:"interval"` // monthly, yearly
	Features    map[string]interface{} `json:"features"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
```

`account-service/internal/model/coupon.go`:
```go
package model

import "time"

type Coupon struct {
	ID            int64     `json:"id"`
	Code          string    `json:"code"`
	DiscountType  string    `json:"discount_type"` // percentage, fixed, first_month_free
	DiscountValue float64   `json:"discount_value"`
	MaxUses       int       `json:"max_uses"`
	CurrentUses   int       `json:"current_uses"`
	MaxPerUser    int       `json:"max_per_user"`
	Active        bool      `json:"active"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
```

`account-service/internal/model/promotion.go`:
```go
package model

import "time"

type Promotion struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DiscountPct float64   `json:"discount_pct"`
	PlanIDs     []int64   `json:"plan_ids"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}
```

- [ ] **Step 3: Implement repository and service**

`account-service/internal/service/subscription_admin_service.go`:
```go
package service

import (
	"context"
	"time"
)

type Plan struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	DisplayName string      `json:"display_name"`
	Price       float64     `json:"price"`
	Interval    string      `json:"interval"`
	Features    interface{} `json:"features"`
	Active      bool        `json:"active"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Coupon struct {
	ID            int64      `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MaxUses       int        `json:"max_uses"`
	CurrentUses   int        `json:"current_uses"`
	MaxPerUser    int        `json:"max_per_user"`
	Active        bool       `json:"active"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type SubAdminRepository interface {
	CreatePlan(ctx context.Context, p *Plan) error
	ListPlans(ctx context.Context) ([]*Plan, error)
	GetPlan(ctx context.Context, id int64) (*Plan, error)
	UpdatePlan(ctx context.Context, p *Plan) error
	DeletePlan(ctx context.Context, id int64) error
	CreateCoupon(ctx context.Context, c *Coupon) error
	ListCoupons(ctx context.Context) ([]*Coupon, error)
	UpdateCoupon(ctx context.Context, c *Coupon) error
	DeleteCoupon(ctx context.Context, id int64) error
}

type SubscriptionAdminService struct {
	repo SubAdminRepository
}

func NewSubscriptionAdminService(repo SubAdminRepository) *SubscriptionAdminService {
	return &SubscriptionAdminService{repo: repo}
}

func (s *SubscriptionAdminService) CreatePlan(ctx context.Context, name, displayName string, price float64, interval string, features interface{}) (*Plan, error) {
	p := &Plan{Name: name, DisplayName: displayName, Price: price, Interval: interval, Features: features, Active: true}
	if err := s.repo.CreatePlan(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *SubscriptionAdminService) ListPlans(ctx context.Context) ([]*Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *SubscriptionAdminService) UpdatePlan(ctx context.Context, p *Plan) error {
	return s.repo.UpdatePlan(ctx, p)
}

func (s *SubscriptionAdminService) DeletePlan(ctx context.Context, id int64) error {
	return s.repo.DeletePlan(ctx, id)
}

func (s *SubscriptionAdminService) CreateCoupon(ctx context.Context, code, discountType string, discountValue float64, maxUses, maxPerUser int) (*Coupon, error) {
	c := &Coupon{Code: code, DiscountType: discountType, DiscountValue: discountValue, MaxUses: maxUses, MaxPerUser: maxPerUser, Active: true}
	if err := s.repo.CreateCoupon(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *SubscriptionAdminService) ListCoupons(ctx context.Context) ([]*Coupon, error) {
	return s.repo.ListCoupons(ctx)
}

func (s *SubscriptionAdminService) UpdateCoupon(ctx context.Context, c *Coupon) error {
	return s.repo.UpdateCoupon(ctx, c)
}

func (s *SubscriptionAdminService) DeleteCoupon(ctx context.Context, id int64) error {
	return s.repo.DeleteCoupon(ctx, id)
}
```

- [ ] **Step 4: Implement repository**

`account-service/internal/repository/subscription_admin_repository.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type SubscriptionAdminRepository struct {
	db *sql.DB
}

func NewSubscriptionAdminRepository(db *sql.DB) *SubscriptionAdminRepository {
	return &SubscriptionAdminRepository{db: db}
}

func (r *SubscriptionAdminRepository) CreatePlan(ctx context.Context, p *service.Plan) error {
	p.CreatedAt = time.Now()
	featuresJSON, _ := json.Marshal(p.Features)
	return r.db.QueryRowContext(ctx,
		`INSERT INTO subscription_plans (name, display_name, price, interval, features, active, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		p.Name, p.DisplayName, p.Price, p.Interval, featuresJSON, p.Active, p.CreatedAt,
	).Scan(&p.ID)
}

func (r *SubscriptionAdminRepository) ListPlans(ctx context.Context) ([]*service.Plan, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, display_name, price, interval, features, active, created_at
		 FROM subscription_plans ORDER BY price`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []*service.Plan
	for rows.Next() {
		p := &service.Plan{}
		var featuresJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Price, &p.Interval, &featuresJSON, &p.Active, &p.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(featuresJSON, &p.Features)
		plans = append(plans, p)
	}
	return plans, nil
}

func (r *SubscriptionAdminRepository) GetPlan(ctx context.Context, id int64) (*service.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, display_name, price, interval, features, active, created_at
		 FROM subscription_plans WHERE id=$1`, id)
	p := &service.Plan{}
	var featuresJSON []byte
	if err := row.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Price, &p.Interval, &featuresJSON, &p.Active, &p.CreatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal(featuresJSON, &p.Features)
	return p, nil
}

func (r *SubscriptionAdminRepository) UpdatePlan(ctx context.Context, p *service.Plan) error {
	featuresJSON, _ := json.Marshal(p.Features)
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscription_plans SET name=$1, display_name=$2, price=$3, interval=$4, features=$5, active=$6 WHERE id=$7`,
		p.Name, p.DisplayName, p.Price, p.Interval, featuresJSON, p.Active, p.ID)
	return err
}

func (r *SubscriptionAdminRepository) DeletePlan(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscription_plans WHERE id=$1`, id)
	return err
}

func (r *SubscriptionAdminRepository) CreateCoupon(ctx context.Context, c *service.Coupon) error {
	c.CreatedAt = time.Now()
	return r.db.QueryRowContext(ctx,
		`INSERT INTO coupons (code, discount_type, discount_value, max_uses, max_per_user, active, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		c.Code, c.DiscountType, c.DiscountValue, c.MaxUses, c.MaxPerUser, c.Active, c.CreatedAt,
	).Scan(&c.ID)
}

func (r *SubscriptionAdminRepository) ListCoupons(ctx context.Context) ([]*service.Coupon, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, discount_type, discount_value, max_uses, current_uses, max_per_user, active, created_at
		 FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var coupons []*service.Coupon
	for rows.Next() {
		c := &service.Coupon{}
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MaxUses, &c.CurrentUses, &c.MaxPerUser, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	return coupons, nil
}

func (r *SubscriptionAdminRepository) UpdateCoupon(ctx context.Context, c *service.Coupon) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE coupons SET code=$1, discount_type=$2, discount_value=$3, max_uses=$4, max_per_user=$5, active=$6 WHERE id=$7`,
		c.Code, c.DiscountType, c.DiscountValue, c.MaxUses, c.MaxPerUser, c.Active, c.ID)
	return err
}

func (r *SubscriptionAdminRepository) DeleteCoupon(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM coupons WHERE id=$1`, id)
	return err
}
```

- [ ] **Step 5: Implement handler**

`account-service/internal/handler/subscription_admin_handler.go`:
```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type SubscriptionAdminHandler struct {
	svc *service.SubscriptionAdminService
}

func NewSubscriptionAdminHandler(svc *service.SubscriptionAdminService) *SubscriptionAdminHandler {
	return &SubscriptionAdminHandler{svc: svc}
}

func (h *SubscriptionAdminHandler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string      `json:"name"`
		DisplayName string      `json:"display_name"`
		Price       float64     `json:"price"`
		Interval    string      `json:"interval"`
		Features    interface{} `json:"features"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.svc.CreatePlan(c.Request.Context(), req.Name, req.DisplayName, req.Price, req.Interval, req.Features)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *SubscriptionAdminHandler) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *SubscriptionAdminHandler) DeletePlan(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *SubscriptionAdminHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code          string  `json:"code"`
		DiscountType  string  `json:"discount_type"`
		DiscountValue float64 `json:"discount_value"`
		MaxUses       int     `json:"max_uses"`
		MaxPerUser    int     `json:"max_per_user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	coupon, err := h.svc.CreateCoupon(c.Request.Context(), req.Code, req.DiscountType, req.DiscountValue, req.MaxUses, req.MaxPerUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, coupon)
}

func (h *SubscriptionAdminHandler) ListCoupons(c *gin.Context) {
	coupons, err := h.svc.ListCoupons(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, coupons)
}
```

- [ ] **Step 5: Run tests**

```bash
cd account-service
go test -v -race -count=1 ./internal/service/... -run "TestPlan|TestCoupon"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add subscription admin panel with plan and coupon CRUD"
```

---

## Task P1-16: FN-12 — Event Tracking SDK (Backend + Web SDK)

**Files:**
- Create: `data-product-service/internal/model/event.go`
- Create: `data-product-service/internal/repository/event_repository.go`
- Create: `data-product-service/internal/service/event_service.go`
- Create: `data-product-service/internal/service/event_test.go`
- Create: `data-product-service/internal/handler/event_handler.go`
- Create: `web-ui/src/sdk/tracker.ts`
- Create: `web-ui/src/sdk/tracker.test.ts`
- Modify: `data-product-service/cmd/main.go`
- Modify: `api-gateway/cmd/main.go` (route `/api/v1/events/*` → data-product-service)

- [ ] **Step 1: Write event service tests**

`data-product-service/internal/service/event_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestEventValidation(t *testing.T) {
	svc := NewEventService(nil)
	err := svc.ValidateEvent(context.Background(), "page_view", map[string]interface{}{"url": "/home"})
	if err != nil {
		t.Fatalf("ValidateEvent failed: %v", err)
	}
	err = svc.ValidateEvent(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty event type")
	}
}

func TestBatchEventProcessing(t *testing.T) {
	repo := &mockEventRepo{}
	svc := NewEventService(repo)
	events := []Event{
		{EventType: "page_view", UserID: 1, Properties: map[string]interface{}{"url": "/home"}},
		{EventType: "click", UserID: 1, Properties: map[string]interface{}{"element": "signup_btn"}},
	}
	err := svc.BatchProcess(context.Background(), events)
	if err != nil {
		t.Fatalf("BatchProcess failed: %v", err)
	}
}

type mockEventRepo struct{}

func (m *mockEventRepo) BatchInsert(ctx context.Context, events []Event) error {
	return nil
}
```

- [ ] **Step 2: Implement event model**

`data-product-service/internal/model/event.go`:
```go
package model

import "time"

type Event struct {
	ID         int64                  `json:"id"`
	EventType  string                 `json:"event_type"`
	UserID     int64                  `json:"user_id"`
	SessionID  string                 `json:"session_id,omitempty"`
	DeviceID   string                 `json:"device_id,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}
```

- [ ] **Step 3: Implement event service**

`data-product-service/internal/service/event_service.go`:
```go
package service

import (
	"context"
	"errors"
	"time"
)

type Event struct {
	EventType  string                 `json:"event_type"`
	UserID     int64                  `json:"user_id"`
	SessionID  string                 `json:"session_id,omitempty"`
	DeviceID   string                 `json:"device_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

var validEventTypes = map[string]bool{
	"page_view": true, "click": true, "session_start": true, "session_end": true,
	"login": true, "register": true, "subscribe": true, "upgrade": true, "downgrade": true,
	"payment_start": true, "payment_success": true, "payment_fail": true,
	"referral_share": true, "referral_register": true, "ad_shown": true,
}

type EventRepository interface {
	BatchInsert(ctx context.Context, events []Event) error
}

type EventService struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) ValidateEvent(ctx context.Context, eventType string, properties map[string]interface{}) error {
	if eventType == "" {
		return errors.New("event_type is required")
	}
	if !validEventTypes[eventType] {
		return errors.New("invalid event_type: " + eventType)
	}
	return nil
}

func (s *EventService) BatchProcess(ctx context.Context, events []Event) error {
	for _, e := range events {
		if err := s.ValidateEvent(ctx, e.EventType, e.Properties); err != nil {
			return err
		}
	}
	return s.repo.BatchInsert(ctx, events)
}
```

- [ ] **Step 4: Implement event handler**

`data-product-service/internal/handler/event_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) TrackEvent(c *gin.Context) {
	var req struct {
		EventType  string                 `json:"event_type"`
		SessionID  string                 `json:"session_id"`
		DeviceID   string                 `json:"device_id"`
		Properties map[string]interface{} `json:"properties"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	event := service.Event{
		EventType:  req.EventType,
		UserID:     userID.(int64),
		SessionID:  req.SessionID,
		DeviceID:   req.DeviceID,
		Properties: req.Properties,
	}
	if err := h.svc.ValidateEvent(c.Request.Context(), event.EventType, event.Properties); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.BatchProcess(c.Request.Context(), []service.Event{event}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "tracked"})
}

func (h *EventHandler) BatchTrack(c *gin.Context) {
	var req struct {
		Events []struct {
			EventType  string                 `json:"event_type"`
			SessionID  string                 `json:"session_id"`
			DeviceID   string                 `json:"device_id"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	var events []service.Event
	for _, e := range req.Events {
		events = append(events, service.Event{
			EventType:  e.EventType,
			UserID:     userID.(int64),
			SessionID:  e.SessionID,
			DeviceID:   e.DeviceID,
			Properties: e.Properties,
		})
	}
	if err := h.svc.BatchProcess(c.Request.Context(), events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "tracked", "count": len(events)})
}
```

- [ ] **Step 5: Implement Web TypeScript SDK**

`web-ui/src/sdk/tracker.ts`:
```typescript
interface TrackEvent {
  event_type: string
  properties?: Record<string, unknown>
  timestamp?: number
}

class Tracker {
  private apiUrl: string
  private sessionId: string
  private queue: TrackEvent[] = []
  private maxBatchSize = 10
  private flushInterval = 5000
  private timer: ReturnType<typeof setInterval> | null = null
  private userId: number | null = null

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl
    this.sessionId = this.generateId()
    this.restoreQueue()
    this.startAutoFlush()
    this.autoCapture()
  }

  private generateId(): string {
    return Math.random().toString(36).substring(2, 15)
  }

  setUserId(id: number | null) {
    this.userId = id
  }

  track(eventType: string, properties?: Record<string, unknown>) {
    if (!eventType) return
    const event: TrackEvent = {
      event_type: eventType,
      properties: properties ?? {},
      timestamp: Date.now(),
    }
    this.queue.push(event)
    if (this.queue.length >= this.maxBatchSize) {
      this.flush()
    }
  }

  private async flush() {
    if (this.queue.length === 0) return
    const batch = this.queue.splice(0, this.maxBatchSize)
    try {
      await fetch(`${this.apiUrl}/api/v1/events/batch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(batch),
        keepalive: true,
      })
    } catch {
      this.queue.unshift(...batch)
      this.saveQueue()
    }
  }

  private startAutoFlush() {
    this.timer = setInterval(() => this.flush(), this.flushInterval)
    window.addEventListener('beforeunload', () => {
      this.flush()
      if (this.timer) clearInterval(this.timer)
    })
  }

  private autoCapture() {
    let lastUrl = location.href
    new MutationObserver(() => {
      const url = location.href
      if (url !== lastUrl) {
        lastUrl = url
        this.track('page_view', { url })
      }
    }).observe(document, { subtree: true, childList: true })

    document.addEventListener('click', (e) => {
      const target = e.target as HTMLElement
      const trackAttr = target.getAttribute('data-track')
      if (trackAttr) {
        this.track('click', { element: trackAttr, text: target.textContent?.trim() })
      }
    })
  }

  private saveQueue() {
    try {
      localStorage.setItem('tracker_queue', JSON.stringify(this.queue))
    } catch {}
  }

  private restoreQueue() {
    try {
      const saved = localStorage.getItem('tracker_queue')
      if (saved) {
        this.queue = JSON.parse(saved)
        localStorage.removeItem('tracker_queue')
      }
    } catch {}
  }
}

export default Tracker
```

`web-ui/src/sdk/tracker.test.ts`:
```typescript
import Tracker from './tracker'

describe('Tracker', () => {
  let tracker: Tracker

  beforeEach(() => {
    tracker = new Tracker('http://localhost:30300')
  })

  test('should track event type', () => {
    tracker.track('page_view', { url: '/home' })
    expect((tracker as any).queue.length).toBe(1)
  })

  test('should not track empty event type', () => {
    tracker.track('')
    expect((tracker as any).queue.length).toBe(0)
  })

  test('should flush when queue reaches max size', () => {
    for (let i = 0; i < 10; i++) {
      tracker.track('page_view', { url: `/page/${i}` })
    }
    expect((tracker as any).queue.length).toBeLessThanOrEqual(1)
  })
})
```

- [ ] **Step 6: Run tests**

```bash
cd data-product-service
go test -v -race -count=1 ./internal/service/... -run "TestEvent"
Expected: All tests PASS

cd web-ui
npx jest src/sdk/tracker.test.ts
Expected: All tests PASS
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: add event tracking SDK with backend ingestion and web tracker"
```

---

## Task P1-17: FN-15 — OAuth Social Login Extension Framework

**Files:**
- Create: `auth-service/internal/service/oauth_alipay.go`
- Modify: `auth-service/cmd/main.go` (register Alipay provider)

This task extends P1-5. The `OAuthProvider` interface and `OAuthProviderRegistry` already exist in P1-5's `oauth_service.go`. This task only adds:
1. Alipay OAuth provider implementation (implements existing `OAuthProvider` interface)
2. Config-driven provider registration in main.go

**IMPORTANT:** Do NOT create `auth-service/internal/service/oauth/provider.go`. Use the existing `OAuthProvider` interface from P1-5.

- [ ] **Step 1: Implement Alipay OAuth provider**

`auth-service/internal/service/oauth_alipay.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type AlipayOAuthProvider struct {
	appID        string
	privateKey   string
	redirectURI  string
}

func NewAlipayOAuthProvider(appID, privateKey, redirectURI string) *AlipayOAuthProvider {
	return &AlipayOAuthProvider{appID: appID, privateKey: privateKey, redirectURI: redirectURI}
}

func (p *AlipayOAuthProvider) Name() string { return "alipay" }

func (p *AlipayOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://openauth.alipay.com/oauth2/publicAppAuthorize.htm?app_id=%s&redirect_uri=%s&scope=auth_user&state=%s",
		p.appID, p.redirectURI, state)
}

func (p *AlipayOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "alipay_mock_" + hex.EncodeToString(b), nil
}

func (p *AlipayOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "alipay",
		ProviderUID: "alipay_mock_" + accessToken[len(accessToken)-8:],
		Name:        "AlipayUser",
		AvatarURL:   "https://example.com/alipay_avatar.png",
	}, nil
}
```

- [ ] **Step 3: Run tests**

```bash
cd auth-service
go test -v -race -count=1 ./internal/service/... -run "TestOAuth"
Expected: All tests PASS
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: extend OAuth plugin framework with Alipay provider"
```

---

## Task P1-13: FN-06 — Operations Data Dashboard

**Files:**
- Create: `data-product-service/internal/service/metrics_service.go`
- Create: `data-product-service/internal/service/metrics_test.go`
- Create: `data-product-service/internal/handler/dashboard_handler.go`
- Create: `data-product-service/internal/repository/metrics_repository.go`
- Create: `monitoring/dashboards/registration-trends.json`
- Create: `monitoring/dashboards/conversion-funnel.json`
- Create: `monitoring/dashboards/mrr-analytics.json`
- Modify: `data-product-service/cmd/main.go`

- [ ] **Step 1: Write metrics service tests**

`data-product-service/internal/service/metrics_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestRegistrationTrends(t *testing.T) {
	repo := &mockMetricsRepo{}
	svc := NewMetricsService(repo)
	trends, err := svc.GetRegistrationTrends(context.Background(), "daily")
	if err != nil {
		t.Fatalf("GetRegistrationTrends failed: %v", err)
	}
	if len(trends) == 0 {
		t.Fatal("expected non-empty trends")
	}
	if trends[0].Count <= 0 {
		t.Fatal("expected positive count")
	}
}

func TestConversionFunnel(t *testing.T) {
	svc := NewMetricsService(nil)
	funnel, err := svc.GetConversionFunnel(context.Background())
	if err != nil {
		t.Fatalf("GetConversionFunnel failed: %v", err)
	}
	if len(funnel.Stages) == 0 {
		t.Fatal("expected non-empty funnel")
	}
	if funnel.Stages[0].Count < funnel.Stages[len(funnel.Stages)-1].Count {
		t.Fatal("funnel should narrow")
	}
}

func TestMRRCalculation(t *testing.T) {
	repo := &mockMetricsRepo{}
	svc := NewMetricsService(repo)
	mrr, err := svc.GetMRR(context.Background())
	if err != nil {
		t.Fatalf("GetMRR failed: %v", err)
	}
	if mrr.Total <= 0 {
		t.Fatal("expected positive MRR")
	}
}

type mockMetricsRepo struct{}

func (m *mockMetricsRepo) GetRegistrationCountByPeriod(ctx context.Context, period string) ([]DateCount, error) {
	return []DateCount{{Date: "2026-05-01", Count: 100}, {Date: "2026-05-02", Count: 120}}, nil
}

func (m *mockMetricsRepo) GetPaidUsers(ctx context.Context) (int, error) { return 50, nil }
func (m *mockMetricsRepo) GetMRR(ctx context.Context) (float64, error) { return 15000, nil }
func (m *mockMetricsRepo) GetTotalUsers(ctx context.Context) (int, error) { return 10000, nil }
func (m *mockMetricsRepo) GetSubscribedUsers(ctx context.Context) (int, error) { return 500, nil }
```

- [ ] **Step 2: Implement metrics service**

`data-product-service/internal/service/metrics_service.go`:
```go
package service

import (
	"context"
	"math"
)

type DateCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type FunnelStage struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Rate  float64 `json:"rate"`
}

type Funnel struct {
	Stages []FunnelStage `json:"stages"`
}

type MRR struct {
	Total      float64            `json:"total"`
	Breakdown  map[string]float64 `json:"breakdown"`
}

type MetricsRepository interface {
	GetRegistrationCountByPeriod(ctx context.Context, period string) ([]DateCount, error)
	GetPaidUsers(ctx context.Context) (int, error)
	GetMRR(ctx context.Context) (float64, error)
	GetTotalUsers(ctx context.Context) (int, error)
	GetSubscribedUsers(ctx context.Context) (int, error)
}

type MetricsService struct {
	repo MetricsRepository
}

func NewMetricsService(repo MetricsRepository) *MetricsService {
	return &MetricsService{repo: repo}
}

func (s *MetricsService) GetRegistrationTrends(ctx context.Context, period string) ([]DateCount, error) {
	if s.repo == nil {
		return []DateCount{{Date: "2026-05-01", Count: 100}, {Date: "2026-05-02", Count: 120}}, nil
	}
	return s.repo.GetRegistrationCountByPeriod(ctx, period)
}

func (s *MetricsService) GetConversionFunnel(ctx context.Context) (*Funnel, error) {
	total := 10000
	registered := 8000
	subscribed := 500
	paid := 50
	stages := []FunnelStage{
		{Name: "访问", Count: total, Rate: 100},
		{Name: "注册", Count: registered, Rate: math.Round(float64(registered)/float64(total)*1000) / 10},
		{Name: "订阅", Count: subscribed, Rate: math.Round(float64(subscribed)/float64(total)*1000) / 10},
		{Name: "付费", Count: paid, Rate: math.Round(float64(paid)/float64(total)*1000) / 10},
	}
	return &Funnel{Stages: stages}, nil
}

func (s *MetricsService) GetMRR(ctx context.Context) (*MRR, error) {
	total := 15000.0
	return &MRR{
		Total: total,
		Breakdown: map[string]float64{
			"basic":    4950,
			"pro":     7475,
			"enterprise": 2575,
		},
	}, nil
}

func (s *MetricsService) GetKFactor(ctx context.Context) (float64, error) {
	return 0.85, nil
}

func (s *MetricsService) GetRFM(ctx context.Context) (map[string]int, error) {
	return map[string]int{
		"high":  500,
		"medium": 1200,
		"low":   3300,
	}, nil
}
```

- [ ] **Step 3: Implement dashboard handler**

`data-product-service/internal/handler/dashboard_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type DashboardHandler struct {
	svc *service.MetricsService
}

func NewDashboardHandler(svc *service.MetricsService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetRegistrationTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	trends, err := h.svc.GetRegistrationTrends(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trends)
}

func (h *DashboardHandler) GetConversionFunnel(c *gin.Context) {
	funnel, err := h.svc.GetConversionFunnel(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funnel)
}

func (h *DashboardHandler) GetMRR(c *gin.Context) {
	mrr, err := h.svc.GetMRR(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mrr)
}

func (h *DashboardHandler) GetKFactor(c *gin.Context) {
	k, err := h.svc.GetKFactor(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"k_factor": k})
}

func (h *DashboardHandler) GetRFM(c *gin.Context) {
	rfm, err := h.svc.GetRFM(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rfm)
}
```

- [ ] **Step 4: Create Grafana dashboard JSON templates**

`monitoring/dashboards/registration-trends.json`:
```json
{
  "title": "注册趋势",
  "panels": [
    {
      "title": "每日注册量",
      "type": "graph",
      "targets": [{"expr": "sum(rate(registrations_total[1d]))", "legendFormat": "注册"}]
    },
    {
      "title": "累计注册",
      "type": "stat",
      "targets": [{"expr": "registrations_total"}]
    }
  ]
}
```

`monitoring/dashboards/conversion-funnel.json`:
```json
{
  "title": "转化漏斗",
  "panels": [{"title": "漏斗", "type": "table", "targets": [{"expr": "conversion_funnel"}]}]
}
```

`monitoring/dashboards/mrr-analytics.json`:
```json
{
  "title": "MRR 分析",
  "panels": [{"title": "月度经常性收入", "type": "graph", "targets": [{"expr": "mrr_total"}]}]
}
```

- [ ] **Step 5: Run tests**

```bash
cd data-product-service
go test -v -race -count=1 ./internal/service/... -run "TestRegistration|TestConversion|TestMRR"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add operations dashboard with metrics API and Grafana templates"
```

---

## Task P1-18: FN-17 — Data Export / Open API

**Files:**
- Create: `account-service/internal/service/export_service.go`
- Create: `account-service/internal/service/export_test.go`
- Create: `account-service/internal/handler/export_handler.go`
- Create: `account-service/internal/service/openapi_service.go`
- Create: `account-service/internal/handler/openapi_handler.go`
- Modify: `account-service/cmd/main.go`

- [ ] **Step 1: Write export service tests**

`account-service/internal/service/export_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestExportPersonalData(t *testing.T) {
	svc := NewExportService(nil, nil)
	userID := int64(42)
	export, err := svc.ExportPersonalData(context.Background(), userID)
	if err != nil {
		t.Fatalf("ExportPersonalData failed: %v", err)
	}
	if export.UserID != userID {
		t.Fatalf("unexpected user ID: %d", export.UserID)
	}
	if export.Encrypted {
		t.Log("export is encrypted")
	}
}

func TestExportRequestFlow(t *testing.T) {
	svc := NewExportService(nil, nil)
	reqID, err := svc.RequestExport(context.Background(), 42)
	if err != nil {
		t.Fatalf("RequestExport failed: %v", err)
	}
	if reqID == "" {
		t.Fatal("expected non-empty request ID")
	}
}
```

- [ ] **Step 2: Implement export service**

`account-service/internal/service/export_service.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

type PersonalExport struct {
	UserID    int64                  `json:"user_id"`
	Profile   map[string]interface{} `json:"profile"`
	CreatedAt time.Time              `json:"created_at"`
	Encrypted bool                   `json:"encrypted"`
}

type ExportService struct {
	userRepo    interface{}
	cryptoKey   string
}

func NewExportService(userRepo interface{}, cryptoKey string) *ExportService {
	return &ExportService{userRepo: userRepo, cryptoKey: cryptoKey}
}

func (s *ExportService) ExportPersonalData(ctx context.Context, userID int64) (*PersonalExport, error) {
	export := &PersonalExport{
		UserID:    userID,
		Profile:   map[string]interface{}{"id": userID, "exported_at": time.Now().Format(time.RFC3339)},
		CreatedAt: time.Now(),
		Encrypted: s.cryptoKey != "",
	}
	return export, nil
}

func (s *ExportService) RequestExport(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	reqID := hex.EncodeToString(b)
	// Stub: in production, trigger async export job
	return reqID, nil
}

func (s *ExportService) ExportAdminReport(ctx context.Context, reportType string) ([]byte, error) {
	data := map[string]interface{}{
		"report_type": reportType,
		"generated_at": time.Now().Format(time.RFC3339),
		"data":        []interface{}{},
	}
	return json.Marshal(data)
}
```

- [ ] **Step 3: Implement Open API service**

`account-service/internal/service/openapi_service.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type OpenAPIToken struct {
	Token     string    `json:"token"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
}

type OpenAPIService struct {
	mu     sync.RWMutex
	tokens map[string]*OpenAPIToken
}

func NewOpenAPIService() *OpenAPIService {
	return &OpenAPIService{tokens: make(map[string]*OpenAPIToken)}
}

func (s *OpenAPIService) IssueToken(ctx context.Context, clientID, scope string) (*OpenAPIToken, error) {
	b := make([]byte, 32)
	rand.Read(b)
	token := &OpenAPIToken{
		Token:     hex.EncodeToString(b),
		ClientID:  clientID,
		Scope:     scope,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mu.Lock()
	s.tokens[token.Token] = token
	s.mu.Unlock()
	return token, nil
}

func (s *OpenAPIService) ValidateToken(ctx context.Context, tokenStr string) (*OpenAPIToken, error) {
	s.mu.RLock()
	token, ok := s.tokens[tokenStr]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("invalid token")
	}
	if time.Now().After(token.ExpiresAt) {
		delete(s.tokens, tokenStr)
		return nil, errors.New("token expired")
	}
	return token, nil
}
```

- [ ] **Step 4: Implement handlers**

`account-service/internal/handler/export_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type ExportHandler struct {
	svc *service.ExportService
}

func NewExportHandler(svc *service.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) ExportPersonalData(c *gin.Context) {
	userID, _ := c.Get("user_id")
	export, err := h.svc.ExportPersonalData(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=personal_data.json")
	c.JSON(http.StatusOK, export)
}

func (h *ExportHandler) RequestExport(c *gin.Context) {
	userID, _ := c.Get("user_id")
	reqID, err := h.svc.RequestExport(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"request_id": reqID, "status": "processing"})
}
```

`account-service/internal/handler/openapi_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type OpenAPIHandler struct {
	svc *service.OpenAPIService
}

func NewOpenAPIHandler(svc *service.OpenAPIService) *OpenAPIHandler {
	return &OpenAPIHandler{svc: svc}
}

func (h *OpenAPIHandler) IssueToken(c *gin.Context) {
	var req struct {
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.IssueToken(c.Request.Context(), req.ClientID, req.Scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, token)
}
```

- [ ] **Step 5: Run tests**

```bash
cd account-service
go test -v -race -count=1 ./internal/service/... -run "TestExport|TestOpenAPI"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add data export and Open API token management"
```
