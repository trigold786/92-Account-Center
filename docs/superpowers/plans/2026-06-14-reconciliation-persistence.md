# Reconciliation Persistence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist payment reconciliation reports to PostgreSQL and allow finance admins to retrieve historical reports after service restart.

**Architecture:** Add a reconciliation repository backed by the existing `reconciliation_reports` table. The reconciliation service writes each generated report to the repository and reads reports from the repository before falling back to in-memory cache. Main wires the repository into the service.

**Tech Stack:** Go, PostgreSQL `database/sql`, JSON encoding for mismatch orders, existing payment-service tests.

---

## File Structure

- Create: `payment-service/internal/repository/reconciliation_report.go` — save/get reconciliation reports.
- Modify: `payment-service/internal/service/reconciliation.go` — repository interface, persistence on reconcile, DB-backed get.
- Modify: `payment-service/internal/service/payment_test.go` — TDD test for Save call and DB-backed Get.
- Modify: `payment-service/cmd/main.go` — instantiate repository and pass to service.
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md` — update FN-02 evidence.

---

### Task 1: Reconciliation Repository and Service Persistence

**Files:**
- Create: `payment-service/internal/repository/reconciliation_report.go`
- Modify: `payment-service/internal/service/reconciliation.go`
- Test: `payment-service/internal/service/payment_test.go`

- [ ] **Step 1: Write failing persistence test**

Add a mock reconciliation report repository and a test `TestReconciliation_PersistsReport` in `payment_test.go`. It should assert `Save(ctx, report)` is called after `ReconcileOrders`.

- [ ] **Step 2: Verify RED**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "TestReconciliation_PersistsReport" -count=1 -v
```

Expected: FAIL because the service has no repository dependency.

- [ ] **Step 3: Implement repository interface and constructor**

Add service-level interface:

```go
type ReconciliationReportRepository interface {
    Save(ctx context.Context, report *ReconciliationReport) error
    GetByID(ctx context.Context, id string) (*ReconciliationReport, error)
}
```

Add constructor `NewReconciliationServiceWithRepository(...)` and have existing constructor delegate with nil repository.

- [ ] **Step 4: Implement PostgreSQL repository**

`Save` inserts/upserts into `reconciliation_reports`. `GetByID` reads the row and unmarshals `mismatch_orders` JSON.

- [ ] **Step 5: Verify GREEN**

Run targeted test. Expected: PASS.

---

### Task 2: Wire Main and Matrix

**Files:**
- Modify: `payment-service/cmd/main.go`
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Wire repository**

Use:

```go
reconciliationRepo := repository.NewReconciliationReportRepository(db)
reconciliationSvc := service.NewReconciliationServiceWithRepository(providerRegistry, orderRepo, reconciliationRepo, logger)
```

- [ ] **Step 2: Update matrix**

FN-02 should mention reconciliation report DB persistence and retrieval.

- [ ] **Step 3: Verify**

Run payment-service `go test ./... -count=1`, web-ui `npm run build`, account-service/api-gateway tests, and compose config.

---

## Self-Review Notes

- This implements persistence for generated reports only. It does not implement remote downloadable settlement file ingestion.
- JSON mismatch order persistence uses the existing `JSONB` column.
- No git commit is included without explicit authorization.
