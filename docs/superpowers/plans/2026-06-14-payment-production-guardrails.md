# Payment Production Guardrails Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent sandbox/insecure payment provider configuration from being used in production mode and make callback verifier mode explicit and testable.

**Architecture:** Add a small payment runtime mode guard in payment-service. Providers keep sandbox/HMAC compatibility for tests and UAT, but production mode fails fast unless required real credentials are present. The main process validates config before registering providers.

**Tech Stack:** Go, provider config structs, unit tests, Docker Compose environment variables.

---

## File Structure

- Modify: `payment-service/internal/service/wechat_pay.go` — add `Mode`, `ValidateProduction`, and secure production requirements.
- Modify: `payment-service/internal/service/alipay.go` — add `Mode`, `ValidateProduction`, and secure production requirements.
- Modify: `payment-service/internal/service/payment_test.go` — TDD tests for production guardrails.
- Modify: `payment-service/cmd/main.go` — read `PAYMENT_MODE`, validate providers before serving.
- Modify: `docker-compose.yml` — set `PAYMENT_MODE` default to `sandbox`.
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md` — update FN-01 evidence.

---

### Task 1: Provider Production Validation

**Files:**
- Modify: `payment-service/internal/service/wechat_pay.go`
- Modify: `payment-service/internal/service/alipay.go`
- Test: `payment-service/internal/service/payment_test.go`

- [ ] **Step 1: Write failing tests**

Add tests:

```go
func TestWeChatProvider_ProductionModeRequiresRealCredentials(t *testing.T)
func TestAlipayProvider_ProductionModeRequiresRealCredentials(t *testing.T)
```

They should assert production mode rejects sandbox/default app IDs, missing API keys/private keys/public keys, and missing WeChat certificate serial/private key path.

- [ ] **Step 2: Verify RED**

Run from `payment-service`:

```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "Test.*ProductionModeRequiresRealCredentials" -count=1 -v
```

Expected: FAIL because validation methods do not exist.

- [ ] **Step 3: Implement minimal validation**

Add `Mode string` to `WeChatPayConfig` and `AlipayConfig`. Add methods:

```go
func (c WeChatPayConfig) ValidateProduction() error
func (c AlipayConfig) ValidateProduction() error
```

When `Mode != "production"`, return nil. In production, reject empty fields and values containing `sandbox`, `test`, or default placeholders.

- [ ] **Step 4: Verify GREEN**

Run the same targeted test command. Expected: PASS.

---

### Task 2: Main Startup Guard

**Files:**
- Modify: `payment-service/cmd/main.go`
- Modify: `docker-compose.yml`

- [ ] **Step 1: Wire mode into provider configs**

Read:

```go
paymentMode := getEnv("PAYMENT_MODE", "sandbox")
```

Pass it into both provider configs. If `ValidateProduction()` fails, log error and `os.Exit(1)`.

- [ ] **Step 2: Compose default**

Set payment-service environment:

```yaml
PAYMENT_MODE: ${PAYMENT_MODE:-sandbox}
```

- [ ] **Step 3: Verify**

Run:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

from `payment-service`.

---

### Task 3: Matrix and Final Verification

**Files:**
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Update FN-01**

Mention production startup guardrails and tests.

- [ ] **Step 2: Final verification**

Run account-service/payment-service/api-gateway `go test ./... -count=1`, `docker compose config --quiet`, `git status --short`, and `git diff --stat`.

---

## Self-Review Notes

- This does not replace real WeChat platform certificate/RSA verification implementation. It prevents accidental production startup with sandbox/HMAC defaults.
- Sandbox/UAT remains usable because `PAYMENT_MODE` defaults to `sandbox`.
- No git commit is included without explicit user authorization.
