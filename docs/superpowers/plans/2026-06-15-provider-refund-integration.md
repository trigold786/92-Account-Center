# Provider Refund Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect refund approval to WeChat Pay and Alipay provider refund calls, persist provider refund evidence, and prevent internal reversal when provider refund fails.

**Architecture:** Keep `RefundService.ApproveRefund()` as the orchestration point. Extend `refunds` and `model.Refund` with provider evidence, add provider-result repository methods, inject the existing provider registry into the refund service, then call the provider before marking an order refunded or notifying account-service.

**Tech Stack:** Go, Gin, PostgreSQL, goose SQL migration, existing `payment-service/internal/provider` abstraction, Vue finance UI, PowerShell verification with `CGO_ENABLED=1`.

---

## File Structure

- Modify `db-migrations/006_payment_schema.sql`: add refund provider evidence columns and indexes.
- Modify `payment-service/internal/model/refund.go`: expose provider refund fields in JSON.
- Modify `payment-service/internal/repository/refund_repository.go`: scan new fields and update provider success/failure.
- Modify `payment-service/internal/service/refund_service.go`: add provider registry dependency and provider refund orchestration.
- Modify `payment-service/internal/service/refund_test.go`: TDD coverage for success, failure, and unknown provider.
- Modify `payment-service/cmd/main.go`: instantiate providers before refund service and pass registry to refund service.
- Modify `web-ui/src/api/finance.ts` and `web-ui/src/views/FinanceAdmin.vue`: show provider refund evidence.
- Modify `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`: record evidence.

Do not commit unless the user explicitly asks for a git commit.

---

### Task 1: Write failing refund provider orchestration tests

**Files:**
- Modify: `payment-service/internal/service/refund_test.go`

- [ ] **Step 1: Update imports**

Add imports for `errors` and `payment-service/internal/provider`.

- [ ] **Step 2: Add provider mocks**

Add a `mockPaymentProvider` implementing `provider.PaymentProvider` with exact signatures from `payment-service/internal/provider/provider.go`. Its `Refund` method must set `called=true`, store the request, return the configured error if present, otherwise return `RefundResponse{RefundNo: req.RefundNo, Status: "SUCCESS", RefundID: "WXREFUND123"}`.

Add `mockProviderRegistry` with `Get(name string) (provider.PaymentProvider, bool)`.

- [ ] **Step 3: Extend mock refund repo**

Replace `mockRefundRepo` so it records `providerRefundID`, `providerStatus`, `providerError`, `failedStatus`, and `approvedStatus`. Add methods `UpdateProviderResult` and `MarkProviderFailure` matching the new repository interface.

- [ ] **Step 4: Add success test**

Add `TestRefundService_ApproveRefund_CallsProviderBeforeInternalReversal`. It should create a paid subscription order with `PaymentMethod: "wechat_native"`, a registry containing `wechat`, and assert provider refund was called with order number, total amount, refund amount, refund result was persisted, refund became approved, order became refunded, and subscription cancellation ran.

- [ ] **Step 5: Add provider failure test**

Add `TestRefundService_ApproveRefund_ProviderFailureBlocksInternalReversal`. It should create a paid subscription order with `PaymentMethod: "alipay_wap"`, provider returns `errors.New("provider down")`, and assert refund is failed with provider error, order is not refunded, subscription cancellation is not called.

- [ ] **Step 6: Add unknown method test**

Add `TestRefundService_ApproveRefund_UnknownPaymentMethodFailsBeforeReversal`. It should use `PaymentMethod: "bank_transfer"`, empty registry, and assert refund is failed with provider error, order is not refunded, subscription cancellation is not called.

- [ ] **Step 7: Run RED test**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestRefundService_ApproveRefund_(CallsProviderBeforeInternalReversal|ProviderFailureBlocksInternalReversal|UnknownPaymentMethodFailsBeforeReversal)" -count=1
```

Expected: FAIL with compile errors because `NewRefundService` does not yet accept provider registry and `RefundRepository` lacks provider result methods.

---

### Task 2: Extend refund schema, model, and repository

**Files:**
- Modify: `db-migrations/006_payment_schema.sql`
- Modify: `payment-service/internal/model/refund.go`
- Modify: `payment-service/internal/repository/refund_repository.go`

- [ ] **Step 1: Update refunds table SQL**

In `db-migrations/006_payment_schema.sql`, add these columns to `refunds`: `refund_no VARCHAR(64) NOT NULL DEFAULT ''`, `provider VARCHAR(50) NOT NULL DEFAULT ''`, `provider_refund_id VARCHAR(128) NOT NULL DEFAULT ''`, `provider_status VARCHAR(50) NOT NULL DEFAULT ''`, `provider_error TEXT NOT NULL DEFAULT ''`, `approved_at TIMESTAMP WITH TIME ZONE`, `failed_at TIMESTAMP WITH TIME ZONE`.

Add indexes: `idx_refunds_provider` on `provider`, and `idx_refunds_refund_no` on `refund_no` where `refund_no <> ''`.

- [ ] **Step 2: Replace refund model**

Replace `payment-service/internal/model/refund.go` with fields: `ID`, `OrderID`, `UserID`, `Amount`, `Reason`, `Status`, `RefundNo`, `Provider`, `ProviderRefundID`, `ProviderStatus`, `ProviderError`, `ApproverID`, `ReviewNote`, `ApprovedAt *time.Time`, `FailedAt *time.Time`, `CreatedAt`, `UpdatedAt`. Use snake_case JSON tags exactly matching the field names in the SQL columns.

- [ ] **Step 3: Update repository selects**

In `payment-service/internal/repository/refund_repository.go`, update `GetByID`, `FindByOrderID`, and `ListByUserID` SELECT lists to scan all new fields. Use `COALESCE(approver_id,0)` for `ApproverID`; scan nullable timestamps into `*time.Time` fields.

- [ ] **Step 4: Add provider update methods**

Add `UpdateProviderResult(ctx, id, refundNo, providerName, providerRefundID, providerStatus string) error` that updates `refund_no`, `provider`, `provider_refund_id`, `provider_status`, clears `provider_error`, and sets `updated_at=NOW()`.

Add `MarkProviderFailure(ctx, id, refundNo, providerName, providerError string) error` that sets `status='failed'`, `refund_no`, `provider`, `provider_error`, `failed_at=NOW()`, and `updated_at=NOW()`.

- [ ] **Step 5: Update status timestamps**

Update `UpdateStatus` so status `approved` sets `approved_at=NOW()` and all status updates set `updated_at=NOW()`.

- [ ] **Step 6: Run repository compile check**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/repository ./internal/model -count=1
```

Expected: PASS or no test files, no compile errors.

---

### Task 3: Implement provider refund orchestration

**Files:**
- Modify: `payment-service/internal/service/refund_service.go`

- [ ] **Step 1: Extend interfaces and struct**

Add `UpdateProviderResult` and `MarkProviderFailure` to the service-level `RefundRepository` interface.

Add `PaymentProviderRegistry` interface with `Get(name string) (provider.PaymentProvider, bool)`.

Add `providerRegistry PaymentProviderRegistry` field to `RefundService`.

- [ ] **Step 2: Update constructor without breaking old calls**

Keep the existing varargs shape but support both optional subscription notifier and optional provider registry. Existing calls with only notifier must keep compiling. Tests can call `NewRefundService(repo, orderRepo, creditSvc, notifier, registry)`.

- [ ] **Step 3: Add provider-name resolver**

Add helper `providerNameFromPaymentMethod(method string) (string, bool)` returning `wechat,true` for methods starting `wechat`, `alipay,true` for methods starting `alipay`, and `false` otherwise. Use `strings.ToLower(strings.TrimSpace(method))`.

- [ ] **Step 4: Add refund number helper**

Add helper `refundNoFor(refundID int64) string` returning `RF` plus the refund ID, e.g. refund ID `123` becomes `RF123`.

- [ ] **Step 5: Call provider before internal reversal**

In `ApproveRefund`, after loading refund and order and calculating amount, resolve provider. If missing or unsupported, call `MarkProviderFailure` and return an error. If present, call `paymentProvider.Refund` with order number, refund number, total amount, calculated refund amount, and reason. On provider error, call `MarkProviderFailure` and return an error. On success, call `UpdateProviderResult`, then `UpdateStatus(..., "approved", ...)`, then continue existing order-refunded/subscription-cancel/credit-reversal flow.

- [ ] **Step 6: Run GREEN refund service tests**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestRefundService_ApproveRefund_(CallsProviderBeforeInternalReversal|ProviderFailureBlocksInternalReversal|UnknownPaymentMethodFailsBeforeReversal|CancelsSubscriptionOrder)" -count=1
```

Expected: PASS.

---

### Task 4: Wire provider registry into main

**Files:**
- Modify: `payment-service/cmd/main.go`

- [ ] **Step 1: Move provider registry creation before refund service**

In `main.go`, create and register WeChat/Alipay providers before creating `refundSvc`.

- [ ] **Step 2: Pass registry to refund service**

Change refund service construction to:

```go
refundSvc := service.NewRefundService(refundRepo, orderRepoRefund, creditSvc, accountNotifier, providerRegistry)
```

- [ ] **Step 3: Avoid duplicate provider registry declaration**

Remove the later duplicate `providerRegistry := provider.NewProviderRegistry()` declaration. The existing callback/create-payment/reconciliation handlers must use the same registry.

- [ ] **Step 4: Run main package compile check**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./cmd -count=1
```

Expected: PASS or no test files, no compile errors.

---

### Task 5: Surface provider refund evidence in finance UI

**Files:**
- Modify: `web-ui/src/api/finance.ts`
- Modify: `web-ui/src/views/FinanceAdmin.vue`

- [ ] **Step 1: Extend finance refund type**

In `web-ui/src/api/finance.ts`, add optional fields to the refund type/interface: `refund_no`, `provider`, `provider_refund_id`, `provider_status`, `provider_error`, `approved_at`, `failed_at`.

- [ ] **Step 2: Add table columns**

In the refunds tab of `FinanceAdmin.vue`, add display columns for refund number, provider, provider refund ID, provider status, and provider error. Keep existing columns and use empty fallback text for missing fields.

- [ ] **Step 3: Run web build**

Run from `web-ui`:

```powershell
npm run build
```

Expected: build succeeds. Chunk size warnings are acceptable.

---

### Task 6: Update traceability matrix and full verification

**Files:**
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Update matrix evidence**

In FN-02 or the finance/refund requirement row, add evidence for provider refund API orchestration: `payment-service/internal/service/refund_service.go`, `payment-service/internal/repository/refund_repository.go`, `db-migrations/006_payment_schema.sql`, and refund service tests.

- [ ] **Step 2: Run payment-service tests**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Run account-service tests**

Run from `account-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 4: Run api-gateway tests**

Run from `api-gateway`:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Run web-ui build**

Run from `web-ui`:

```powershell
npm run build
```

Expected: PASS. Chunk size warnings are acceptable.

- [ ] **Step 6: Validate compose config**

Run from repository root:

```powershell
docker compose config --quiet
```

Expected: exit code 0.

---

## Self-Review

Spec coverage: provider success, provider failure, unknown payment method, persistence, main wiring, UI evidence, and verification are covered by Tasks 1-6.

Placeholder scan: no TBD/TODO/fill-in placeholders remain. Each task has exact files and commands.

Type consistency: provider registry uses the existing `provider.PaymentProvider` and `ProviderRegistry.Get(name) (PaymentProvider, bool)` signature. Refund repository method names match the service plan.
