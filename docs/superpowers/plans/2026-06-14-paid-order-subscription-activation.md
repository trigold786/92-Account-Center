# Paid Order Subscription Activation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make payment-service verified payment callbacks automatically activate subscription entitlements in account-service without allowing unpaid or duplicate activations.

**Architecture:** The order is the source of payment truth. payment-service updates an order to `paid` after a verified provider callback, then calls account-service internal activation endpoint with the paid order details. account-service performs idempotent validation, checks the order via payment-service, and creates the subscription, tier update, and entitlement grant exactly once.

**Tech Stack:** Go services, Gin HTTP handlers, PostgreSQL repositories, Docker Compose service-to-service HTTP, PowerShell verification.

---

## File Structure

- Modify: `account-service/internal/model/subscription.go` — add `ActivatePaidOrderRequest` for internal activation.
- Modify: `account-service/internal/repository/subscription_repository.go` — add `GetByOrderID` for idempotency.
- Modify: `account-service/internal/service/subscription_service.go` — add `ActivatePaidOrderSubscription` behavior that rejects duplicate or unpaid orders.
- Modify: `account-service/internal/service/subscription_service_test.go` — TDD tests for idempotent paid-order activation.
- Modify: `account-service/internal/handler/subscription_handler.go` — add internal activation handler.
- Modify: `account-service/cmd/main.go` — register `/internal/v1/subscriptions/activate-paid-order`.
- Modify: `payment-service/cmd/main.go` — replace empty notifier with HTTP POST to account-service and inject it into callback handling.
- Modify: `payment-service/internal/handler/callback.go` — call notifier after successful paid order update for subscription orders.
- Modify: `payment-service/internal/handler/callback_test.go` — TDD test that callback notifies subscription activation after marking order paid.
- Modify: `docker-compose.yml` — add `ACCOUNT_SERVICE_URL=http://account-service:30301` to payment-service.
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md` — update FN-01/FN-02 evidence.

---

### Task 1: Account-Service Idempotent Activation Service

**Files:**
- Modify: `account-service/internal/model/subscription.go`
- Modify: `account-service/internal/repository/subscription_repository.go`
- Modify: `account-service/internal/service/subscription_service.go`
- Test: `account-service/internal/service/subscription_service_test.go`

- [ ] **Step 1: Write failing duplicate-order activation test**

Add a test named `TestSubscriptionService_ActivatePaidOrder_IsIdempotentByOrderID` in `account-service/internal/service/subscription_service_test.go`. It should arrange an existing subscription with `OrderID: "1001"`, call `ActivatePaidOrderSubscription` with the same order, and assert that it returns the existing subscription without calling `Create`, `UpdateIdentityTier`, or `GrantEntitlements`.

- [ ] **Step 2: Verify RED**

Run from `account-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestSubscriptionService_ActivatePaidOrder_IsIdempotentByOrderID" -count=1 -v
```

Expected: FAIL because `ActivatePaidOrderSubscription` and repository `GetByOrderID` do not exist yet.

- [ ] **Step 3: Implement model and repository interface support**

Add to `account-service/internal/model/subscription.go`:

```go
type ActivatePaidOrderRequest struct {
    UserID        int64   `json:"user_id" binding:"required"`
    OrderID       string  `json:"order_id" binding:"required"`
    TierLevel     int     `json:"tier_level" binding:"required,oneof=2 3 4"`
    Price         float64 `json:"price" binding:"required"`
    PaymentMethod string  `json:"payment_method"`
}
```

Add `GetByOrderID(ctx context.Context, orderID string) (*model.Subscription, error)` to `SubscriptionRepository` and implement it with `SELECT ... FROM subscriptions WHERE order_id = $1 LIMIT 1`.

- [ ] **Step 4: Implement minimal idempotent activation service**

Add `ActivatePaidOrderSubscription(ctx context.Context, req *model.ActivatePaidOrderRequest) (*model.Subscription, error)` to `SubscriptionService`. It must:

1. Reject empty `OrderID`.
2. Return existing subscription if `GetByOrderID` finds one.
3. Verify the paid order through the existing `PaymentOrderVerifier`.
4. Reject if order status is not `paid`, user mismatch, or amount mismatch.
5. Create an `ACTIVE` subscription with the same `OrderID`.
6. Update user tier.
7. Grant entitlements.

- [ ] **Step 5: Verify GREEN**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestSubscriptionService_ActivatePaidOrder" -count=1 -v
```

Expected: PASS.

---

### Task 2: Account-Service Internal Activation HTTP Endpoint

**Files:**
- Modify: `account-service/internal/handler/subscription_handler.go`
- Modify: `account-service/cmd/main.go`
- Test: add or update handler tests if existing pattern allows; otherwise compile and service tests verify wiring.

- [ ] **Step 1: Write failing handler test or compile smoke**

Add test `TestSubscriptionHandler_ActivatePaidOrder_ReturnsCreatedSubscription` if handler tests exist; otherwise run after adding route and rely on compile failure before service method exists.

- [ ] **Step 2: Implement handler**

Add method:

```go
func (h *SubscriptionHandler) ActivatePaidOrder(c *gin.Context) {
    var req model.ActivatePaidOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    sub, err := h.svc.ActivatePaidOrderSubscription(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, sub)
}
```

Register in `account-service/cmd/main.go`:

```go
internalSubscriptionGroup := r.Group("/internal/v1/subscriptions")
{
    internalSubscriptionGroup.POST("/activate-paid-order", subscriptionHandler.ActivatePaidOrder)
}
```

- [ ] **Step 3: Verify account-service compiles**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

Expected: PASS.

---

### Task 3: Payment-Service Subscription Activation Notifier

**Files:**
- Modify: `payment-service/cmd/main.go`
- Modify: `payment-service/internal/handler/callback.go`
- Test: `payment-service/internal/handler/callback_test.go`

- [ ] **Step 1: Write failing callback notification test**

Add test `TestCallbackHandler_NotifiesSubscriptionActivationAfterPaidOrderUpdate`. It should create a callback for an order with `ProductType: "subscription"` and `Metadata` containing `{"tier_level":2}`, then assert the notifier receives order id, user id, amount, payment method, and tier level after successful callback processing.

- [ ] **Step 2: Verify RED**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/handler -run "TestCallbackHandler_NotifiesSubscriptionActivationAfterPaidOrderUpdate" -count=1 -v
```

Expected: FAIL because callback handler has no notifier dependency.

- [ ] **Step 3: Add notifier interface to callback handler**

Move or define a narrow handler dependency:

```go
type SubscriptionActivationNotifier interface {
    NotifySubscriptionActivation(ctx context.Context, req SubscriptionActivationRequest) error
}

type SubscriptionActivationRequest struct {
    UserID        int64   `json:"user_id"`
    OrderID       string  `json:"order_id"`
    TierLevel     int     `json:"tier_level"`
    Price         float64 `json:"price"`
    PaymentMethod string  `json:"payment_method"`
}
```

Call it only after `orderSvc.UpdateStatus` succeeds and only when `order.ProductType == "subscription"`.

- [ ] **Step 4: Implement HTTP notifier in `payment-service/cmd/main.go`**

Replace the current no-op `httpNotifier` with a real POST to:

```text
{ACCOUNT_SERVICE_URL}/internal/v1/subscriptions/activate-paid-order
```

JSON payload:

```json
{
  "user_id": 1,
  "order_id": "101",
  "tier_level": 2,
  "price": 99.9,
  "payment_method": "wechat"
}
```

Use `order.ID` as the `order_id` string because account-service verifies against payment-service `/api/v1/orders/{id}`.

- [ ] **Step 5: Verify GREEN**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/handler -run "TestCallbackHandler_NotifiesSubscriptionActivationAfterPaidOrderUpdate" -count=1 -v
```

Expected: PASS.

---

### Task 4: Compose Wiring and Evidence Update

**Files:**
- Modify: `docker-compose.yml`
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Add account-service URL to payment-service**

In `docker-compose.yml`, add to `payment-service.environment`:

```yaml
ACCOUNT_SERVICE_URL: "http://account-service:30301"
```

- [ ] **Step 2: Update matrix**

Update FN-01 and FN-02 rows to mention:

- payment callback persistence
- callback-driven subscription activation notifier
- account-service internal paid-order activation endpoint
- tests that passed

- [ ] **Step 3: Final verification**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

from `account-service`, `payment-service`, and `api-gateway`.

Run from repo root:

```powershell
docker compose config --quiet
git status --short
git diff --stat
```

Expected: tests pass, compose config has no output, status shows only intended changed files.

---

## Self-Review Notes

- This plan implements only verified callback → paid order → subscription activation. It does not implement refunds/downgrades or production provider certificate verification.
- The account-service remains the subscription authority; payment-service only triggers internal activation after a verified payment event.
- Duplicate provider callbacks are handled by account-service `OrderID` idempotency.
- No git commit is included because this workspace requires explicit user authorization before git mutations.
