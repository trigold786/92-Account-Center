# Phase 6 P0 Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Clear the first P0 foundation blockers for Account Center V2.0 full 100% completion.

**Architecture:** Start with low-risk, locally verifiable foundations: repository hygiene, a truth-based requirements matrix, registration password hashing, payment-service deployment reachability, and baseline verification. Keep changes minimal and evidence-driven; do not claim external production readiness without credentials.

**Tech Stack:** Go 1.x services, Vue 3 web-ui, Docker Compose, PostgreSQL, Redis, PowerShell on Windows, Markdown documentation.

---

## File Structure

- Modify: `.gitignore` — prevent binary/runtime artifacts from re-entering the repo.
- Delete from git/worktree: tracked or untracked generated binaries such as `*.exe`, `cmd.exe`, `nul` when present.
- Create: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md` — truth matrix for 86 PRD requirements.
- Modify: `account-service/internal/service/user_service.go` — registration password hashing must use argon2id-compatible format or delegate to shared hashing.
- Test: `account-service/internal/service/user_service_test.go` or adjacent existing tests — verify new registration no longer stores raw SM3-only hashes.
- Modify: `docker-compose.yml` — include payment-service and expose env vars to api-gateway if missing.
- Modify: `api-gateway/cmd/main.go` — proxy payment routes if missing.
- Test: Go tests for gateway route/proxy config where practical; otherwise a deterministic config smoke script.

---

### Task 1: Repository Hygiene and Artifact Guard

**Files:**
- Modify: `.gitignore`
- Delete if present: `generate_down.exe`, `*/main.exe`, `*/cmd.exe`, `nul`

- [ ] **Step 1: Inspect tracked binary artifacts**

Run:
```powershell
git ls-files | rg "(^|/)(main|cmd|generate_down)\.exe$|(^|/)nul$"
```
Expected before fix: outputs any tracked binary/runtime artifact paths or no output if already untracked.

- [ ] **Step 2: Inspect untracked binary artifacts**

Run:
```powershell
git status --short | rg "\.exe$|\bnul$"
```
Expected before fix: outputs any untracked binary/runtime artifacts.

- [ ] **Step 3: Update `.gitignore` artifact rules**

Ensure `.gitignore` contains:
```gitignore
# Generated binaries and Windows runtime artifacts
*.exe
cmd.exe
main.exe
nul

# Local isolated worktrees
.worktrees/
worktrees/
```

- [ ] **Step 4: Remove artifacts from working tree/index**

Run after confirming paths from Step 1/2:
```powershell
git rm -f --ignore-unmatch generate_down.exe nul
Get-ChildItem -Recurse -File -Include main.exe,cmd.exe | ForEach-Object { git rm -f --ignore-unmatch $_.FullName; if (Test-Path -LiteralPath $_.FullName) { Remove-Item -LiteralPath $_.FullName -Force } }
```
Expected: artifact files no longer appear in `git status` as present files.

- [ ] **Step 5: Verify artifact guard**

Run:
```powershell
git ls-files | rg "(^|/)(main|cmd|generate_down)\.exe$|(^|/)nul$"
git status --short | rg "\.exe$|\bnul$"
```
Expected: no output from both commands.

---

### Task 2: Create Reality Traceability Matrix

**Files:**
- Create: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Create matrix document**

Create the document with columns:
```markdown
# Account Center V2.0 需求真实性追溯矩阵

> 依据：PRD V2.0.0 / SSD V2.0.0 / 当前源码静态审计
> 状态枚举：完成 / 部分完成 / Stub-Mock / 缺失 / 外部依赖阻塞 / 不适用

| PRD ID | Phase | 优先级 | 需求名称 | 业务闭环 | 当前状态 | 代码证据 | 端侧证据 | DB/迁移证据 | 测试证据 | 部署证据 | 主要缺口 | 下一步 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
```

- [ ] **Step 2: Seed P0 rows from audit**

Add at minimum these rows: NF-01, NF-02, AR-13, AR-16, AR-17, AR-18, AR-21, AR-23, AR-25, FN-01, FN-02, FN-05, FN-10, UX-08.

- [ ] **Step 3: Verify no false completed rows**

Each row marked `完成` must include code, test, and deployment evidence. If evidence is incomplete, mark `部分完成` or `Stub-Mock`.

---

### Task 3: Registration Password Hashing TDD

**Files:**
- Modify: `account-service/internal/service/user_service.go`
- Test: `account-service/internal/service/user_service_test.go` or nearest existing registration service test

- [ ] **Step 1: Write failing test**

Add a registration test asserting stored password hash starts with `$argon2id$` or the repository receives an argon2id hash, not `salt$sm3`.

Example expected assertion:
```go
if !strings.HasPrefix(saved.PasswordHash, "$argon2id$") {
    t.Fatalf("expected argon2id password hash, got %q", saved.PasswordHash)
}
if strings.Contains(saved.PasswordHash, "$") && strings.Count(saved.PasswordHash, "$") < 4 {
    t.Fatalf("argon2id hash format is incomplete: %q", saved.PasswordHash)
}
```

- [ ] **Step 2: Verify RED**

Run:
```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "Test.*Register.*Argon2" -count=1 -v
```
from `account-service`.
Expected: FAIL because registration currently stores SM3-style hash or test function is missing target behavior.

- [ ] **Step 3: Implement minimal argon2id registration hash**

Use the existing auth-service argon2id implementation as the source of truth, or add a focused local helper if cross-module import is not practical. Do not change login compatibility in this task beyond what is needed for registration.

- [ ] **Step 4: Verify GREEN**

Run:
```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -run "Test.*Register.*Argon2" -count=1 -v
```
Expected: PASS.

- [ ] **Step 5: Run account-service service tests**

Run:
```powershell
$env:CGO_ENABLED="1"; go test ./internal/service ./internal/auth ./internal/repository -count=1
```
Expected: PASS or document pre-existing unrelated failures.

---

### Task 4: Payment-Service Deployment Reachability

**Files:**
- Modify: `docker-compose.yml`
- Modify: `api-gateway/cmd/main.go`
- Modify: `api-gateway/internal/svcconfig/config.go` if service URL config is missing

- [ ] **Step 1: Write failing route/config test or smoke check**

Add or run a deterministic check that fails if `payment-service` is not in Compose or if gateway has no `/api/v1/payment`, `/api/v1/orders`, `/api/v1/refunds`, `/api/v1/invoices` proxy path.

PowerShell smoke check:
```powershell
$r1 = Select-String -LiteralPath "docker-compose.yml" -Pattern "payment-service" -Quiet
$r2 = Select-String -LiteralPath "api-gateway/cmd/main.go" -Pattern "payment|orders|refunds|invoices" -Quiet
if (-not ($r1 -and $r2)) { throw "payment-service deployment or gateway routes missing" }
```
Expected before fix: FAIL if missing.

- [ ] **Step 2: Add payment-service to Compose**

Add service using the existing local-build/scratch pattern and env vars for Postgres, Redis, JWT, payment sandbox config, and port `30316`.

- [ ] **Step 3: Add gateway service URL and proxy routes**

Set `PAYMENT_SERVICE_URL=http://payment-service:30316` and add protected route groups for payment/order/refund/invoice endpoints consistent with RBAC roles.

- [ ] **Step 4: Verify smoke check passes**

Run the Step 1 PowerShell smoke check again.
Expected: PASS.

---

### Task 5: Baseline Verification and Evidence Update

**Files:**
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Run targeted Go tests**

Run:
```powershell
$env:CGO_ENABLED="1"; go test ./internal/service -count=1
```
from `account-service`.

Run:
```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```
from `api-gateway`.

- [ ] **Step 2: Run frontend type/build smoke if touched**

If web-ui changed, run:
```powershell
npm run type-check
npm run build
```
from `web-ui`.

- [ ] **Step 3: Update matrix evidence**

For AR-25, AR-13, and payment-service deployment rows, update status and evidence based on actual verification results.

- [ ] **Step 4: Final status check**

Run:
```powershell
git status --short
git diff --stat
```
Expected: only intended files changed.

---

## Self-Review Notes

Spec coverage in this first plan is intentionally limited to Phase 6/P0 foundation blockers. It does not claim to complete all 86 requirements. Follow-up plans cover payment/subscription, account compliance, growth/ops, data/notification/mobile/ads, and final hardening/UAT.

No production code should be changed without a failing test or deterministic failing smoke check first. Configuration-only changes may use smoke checks instead of unit tests.
