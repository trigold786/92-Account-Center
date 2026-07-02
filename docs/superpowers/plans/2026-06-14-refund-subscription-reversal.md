# Refund Subscription Reversal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure approved refunds for subscription orders cancel paid subscriptions and revoke entitlements, with idempotent account-service behavior and payment-service rollback on failure.

**Architecture:** account-service remains the authority for subscriptions and entitlements. payment-service approves refunds, updates the order to `refunded`, and calls account-service internal cancellation for subscription orders. Both sides use narrow interfaces and TDD tests; duplicate refund notifications are idempotent.

**Tech Stack:** Go, Gin, PostgreSQL repositories, Docker Compose service-to-service HTTP, PowerShell verification.

---

## File Structure

- Modify: `account-service/internal/model/subscription.go` — add `CancelRefundedOrderRequest`.
- Modify: `account-service/internal/service/subscription_service.go` — add `CancelRefundedOrderSubscription`.
- Modify: `account-service/internal/service/subscription_service_test.go` — TDD tests for cancellation and idempotency.
- Modify: `account-service/internal/handler/subscription_handler.go` — add internal cancellation handler.
- Modify: `account-service/cmd/main.go` — register `/internal/v1/subscriptions/cancel-refunded-order`.
- Modify: `payment-service/internal/service/refund_service.go` — add subscription cancellation notifier and update order status to refunded.
- Modify: `payment-service/internal/service/refund_test.go` — TDD tests for subscription cancellation and rollback.
- Modify: `payment-service/cmd/main.go` — implement HTTP cancellation notifier against account-service.
- Modify: `docker-compose.yml` — reuse `ACCOUNT_SERVICE_URL` for payment-service.
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md` — update FN-02 refund reversal evidence.

---

### Task 1: Account-Service Refunded Subscription Cancellation

**Files:**
- Modify: `account-service/internal/model/subscription.go`
- Modify: `account-service/internal/service/subscription_service.go`
- Test: `account-service/internal/service/subscription_service_test.go`

- [ ] **Step 1: Write failing cancellation test**

Add `TestSubscriptionService_CancelRefundedOrder_CancelsSubscriptionAndDeletesEntitlements` to `account-service/internal/service/subscription_service_test.go`. It should arrange `GetByOrderID("102")` returning an active subscription, expect `UpdateStatus(id, "REFUNDED")`, `UpdateIdentityTier(userID, 0)`, and `DeleteUserEntitlements(userID)` or equivalent service method, then call `CancelRefundedOrderSubscription`.

- [ ] **Step 2: Verify RED**

Run from `account-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestSubscriptionService_CancelRefundedOrder" -count=1 -v
```

Expected: FAIL because request type/service method/entitlement revoke method do not exist yet.

- [ ] **Step 3: Implement request model and service contract**

Add to `account-service/internal/model/subscription.go`:

```go
type CancelRefundedOrderRequest struct {
    UserID  int64  `json:"user_id" binding:"required"`
    OrderID string `json:"order_id" binding:"required"`
    Reason  string `json:"reason"`
}
```

Add `CancelRefundedOrderSubscription(ctx context.Context, req *model.CancelRefundedOrderRequest) error` to `SubscriptionService`.

- [ ] **Step 4: Add entitlement revocation seam**

Extend the service-level `EntitlementService` interface with:

```go
DeleteUserEntitlements(ctx context.Context, userID int64) error
```

Implement it in `account-service/internal/service/entitlement_service.go` by delegating to `EntitlementRepository.DeleteByUserID` and clearing cache if supported by existing cache methods.

- [ ] **Step 5: Implement cancellation behavior**

In `CancelRefundedOrderSubscription`:

1. Reject empty `OrderID` with `ErrPaidOrderRequired`.
2. `GetByOrderID`; if nil, return nil for idempotency.
3. If status is `REFUNDED` or `CANCELLED`, return nil.
4. If `UserID` mismatches, return `ErrPaidOrderRequired`.
5. `UpdateStatus(subscription.ID, "REFUNDED")`.
6. `UpdateIdentityTier(userID, 0)`.
7. `DeleteUserEntitlements(userID)`.

- [ ] **Step 6: Verify GREEN**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestSubscriptionService_CancelRefundedOrder" -count=1 -v
```

Expected: PASS.

---

### Task 2: Account-Service Internal Cancellation Endpoint

**Files:**
- Modify: `account-service/internal/handler/subscription_handler.go`
- Modify: `account-service/cmd/main.go`

- [ ] **Step 1: Implement handler**

Add to `SubscriptionHandler`:

```go
func (h *SubscriptionHandler) CancelRefundedOrder(c *gin.Context) {
    var req model.CancelRefundedOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    if err := h.svc.CancelRefundedOrderSubscription(c.Request.Context(), &req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "subscription cancelled"})
}
```

Register route in existing internal subscriptions group:

```go
internalSubscriptionGroup.POST("/cancel-refunded-order", subscriptionHandler.CancelRefundedOrder)
```

- [ ] **Step 2: Verify compile**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

Expected: PASS.

---

### Task 3: Payment-Service Refund Approval Notifier

**Files:**
- Modify: `payment-service/internal/service/refund_service.go`
- Modify: `payment-service/internal/service/refund_test.go`
- Modify: `payment-service/cmd/main.go`

- [ ] **Step 1: Write failing notifier test**

Add `TestRefundService_ApproveRefund_CancelsSubscriptionOrder` in `payment-service/internal/service/refund_test.go`. It should arrange a paid subscription order and pending refund; expect refund status approved, order status refunded, and subscription cancel notifier called with user/order/reason.

- [ ] **Step 2: Verify RED**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestRefundService_ApproveRefund_CancelsSubscriptionOrder" -count=1 -v
```

Expected: FAIL because RefundService has no subscription cancellation notifier and does not update order refunded.

- [ ] **Step 3: Implement notifier interface**

Add to `refund_service.go`:

```go
type SubscriptionCancellationNotifier interface {
    CancelRefundedOrderSubscription(ctx context.Context, userID int64, orderID int64, reason string) error
}
```

Extend `RefundService` and constructor to accept the notifier.

- [ ] **Step 4: Update approval flow**

In `ApproveRefund` after calculating refund and before credit reversal:

1. Update order status to `refunded` using `orderRepo.UpdateStatus(ctx, order.ID, "refunded")` or equivalent adapter.
2. If `order.ProductType == "subscription"`, call cancellation notifier.
3. If order update/notifier/credit reversal fails, set refund back to `pending` and return error.

- [ ] **Step 5: Implement HTTP notifier in `cmd/main.go`**

Add method on existing `httpNotifier`:

```go
func (n *httpNotifier) CancelRefundedOrderSubscription(ctx context.Context, userID int64, orderID int64, reason string) error
```

POST to:

```text
{ACCOUNT_SERVICE_URL}/internal/v1/subscriptions/cancel-refunded-order
```

Payload:

```json
{"user_id":1,"order_id":"102","reason":"refund approved"}
```

- [ ] **Step 6: Verify GREEN**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestRefundService_ApproveRefund" -count=1 -v
```

Expected: PASS.

---

### Task 4: Evidence and Final Verification

**Files:**
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Update matrix**

Update FN-02 to mention refund approval updates orders to refunded and subscription refunds trigger account-service cancellation/revocation.

- [ ] **Step 2: Run full verification**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

from `account-service`, `payment-service`, and `api-gateway`.

Run:

```powershell
docker compose config --quiet
git status --short
git diff --stat
```

Expected: all tests pass; compose config has no output; changed files are intended only.

---

## Self-Review Notes

- This plan focuses only on refund → subscription reversal. It does not implement external provider refund APIs or financial GUI.
- The cancellation is intentionally idempotent: missing/already-refunded subscriptions return success.
- The implementation must not commit because git mutations require explicit user authorization.
