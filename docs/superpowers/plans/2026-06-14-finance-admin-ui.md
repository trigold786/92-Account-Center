# Finance Admin UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a real finance admin UI entry for orders, refunds, invoices, and reconciliation actions backed by existing payment-service gateway APIs.

**Architecture:** Add a focused `finance.ts` API module and a `FinanceAdmin.vue` view. Route and sidebar expose it behind existing finance/admin permissions. Keep the first iteration operational: list/search orders/refunds/invoices, approve/reject refunds, and trigger reconciliation.

**Tech Stack:** Vue 3, TypeScript, Element Plus, Axios API client, Vite/vue-tsc build verification.

---

## File Structure

- Create: `web-ui/src/api/finance.ts` — payment/order/refund/invoice/reconcile API wrappers.
- Create: `web-ui/src/views/FinanceAdmin.vue` — finance operations page.
- Modify: `web-ui/src/router/index.ts` — add `/finance` route.
- Modify: `web-ui/src/layouts/DefaultLayout.vue` — add finance nav item gated by `nav.finance`.
- Modify: `web-ui/src/views/Dashboard.vue` — add finance quick link for finance/admin roles.
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md` — update FN-02 finance GUI evidence.

---

### Task 1: Finance API Module

**Files:**
- Create: `web-ui/src/api/finance.ts`

- [ ] **Step 1: Create API wrappers**

Implement functions:

```ts
listOrders(params)
listRefunds(params)
approveRefund(id)
rejectRefund(id, note)
listInvoices(params)
createInvoice(payload)
reconcile(provider, date)
getReconciliationReport(reportId)
```

Use the existing `client` from `@/api/client` and endpoints `/orders`, `/refunds`, `/invoices`, `/payment/reconcile`.

- [ ] **Step 2: Verify TypeScript imports**

Run later via `npm run build`.

---

### Task 2: Finance Admin View

**Files:**
- Create: `web-ui/src/views/FinanceAdmin.vue`

- [ ] **Step 1: Implement view**

Create tabs:

1. 订单：filter by status/user, table order_no/user/product/amount/status/payment method.
2. 退款：table pending refunds with approve/reject buttons gated by permissions.
3. 发票：list invoices and create invoice dialog.
4. 对账：provider/date form, trigger reconcile, show report summary.

Use Element Plus and error messages.

- [ ] **Step 2: Verify build later**

Run later via `npm run build`.

---

### Task 3: Routing and Navigation

**Files:**
- Modify: `web-ui/src/router/index.ts`
- Modify: `web-ui/src/layouts/DefaultLayout.vue`
- Modify: `web-ui/src/views/Dashboard.vue`

- [ ] **Step 1: Add route**

Add:

```ts
{ path: '/finance', name: 'FinanceAdmin', component: () => import('@/views/FinanceAdmin.vue'), meta: { requiresAuth: true } }
```

- [ ] **Step 2: Add sidebar item**

Add finance menu item gated by `hasPermission('nav.finance')`.

- [ ] **Step 3: Add dashboard quick link**

Finance/admin/operator roles should see `/finance`.

---

### Task 4: Evidence and Verification

**Files:**
- Modify: `docs/requirements-traceability/AccountCenter_V2.0_Reality_Matrix.md`

- [ ] **Step 1: Update FN-02**

Mention finance UI files and build verification.

- [ ] **Step 2: Run verification**

Run from `web-ui`:

```powershell
npm run build
```

Run backend smoke again if needed:

```powershell
$env:CGO_ENABLED="1"; go test ./... -count=1
```

for `payment-service`, `account-service`, and `api-gateway`.

---

## Self-Review Notes

- This plan adds an operational finance UI. It does not implement external tax-platform invoice issuance or advanced accounting exports.
- API field normalization is defensive because payment-service endpoints return mixed `data` shapes.
- No git commit is included without explicit authorization.
