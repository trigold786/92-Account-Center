# UX-08 Pricing Transparency Design

Date: 2026-06-15
Status: Approved for planning
Scope: Phase 6 / P0 — pricing transparency and purchase entry

## Goal

Create a publicly accessible pricing page that shows all subscription tiers with transparent pricing, feature comparison, and a credit discount calculator. Logged-in users can proceed to purchase; anonymous users are redirected to login with a return redirect.

## Current State

The backend already provides:
- `GET /api/v1/pricing` returning 3 tiers (基础版/专业版/企业版) with monthly/yearly prices, feature lists, and entitlement quotas.
- `POST /api/v1/pricing/calculate-discount` returning credit-discounted final price.

The frontend has:
- `Subscriptions.vue` showing the current user's subscription status with purchase/upgrade/renew buttons.
- No pricing page, no `/pricing` route, no pricing API client.
- The sidebar and dashboard have no link to pricing.

## Design

### Route

Add `/pricing` as a public route using `AuthLayout` (no sidebar, centered). It does NOT require authentication. Both anonymous and authenticated users can view it.

### Page Structure

The pricing page has three sections:

1. **Tier cards** — Three cards side by side. Each card shows:
   - Tier name (基础版/专业版/企业版)
   - Monthly/yearly toggle affecting the displayed price
   - Feature list with checkmark (included) or cross (not included)
   - Entitlement quotas (AI calls/month, credit multiplier)
   - A "选择此方案" button

2. **Credit discount calculator** — A widget below the cards:
   - User inputs a target price and credit amount
   - Calls `POST /api/v1/pricing/calculate-discount`
   - Displays original price, credit value, discount percent, final price

3. **Purchase entry** — When a user clicks "选择此方案":
   - If logged in: redirect to `/subscriptions` to complete purchase
   - If not logged in: redirect to `/login?redirect=/pricing`

### Components

| File | Responsibility |
|------|---------------|
| `web-ui/src/views/Pricing.vue` | Main pricing page: tier cards, toggle, calculator, purchase redirect |
| `web-ui/src/api/pricing.ts` | API client: `getPricing()` and `calculateDiscount(price, credits)` |
| `web-ui/src/router/index.ts` | Add `/pricing` route with `meta: { layout: 'auth' }` |
| `web-ui/src/layouts/DefaultLayout.vue` | Add "定价" sidebar entry |
| `web-ui/src/views/Dashboard.vue` | Add "查看定价" quick link |
| `web-ui/src/views/Login.vue` | After login success, check `redirect` query param |

### API Client

`web-ui/src/api/pricing.ts` will use the existing axios `client` from `web-ui/src/api/client.ts`:

```typescript
export function getPricing() {
  return client.get('/pricing')
}
export function calculateDiscount(price: number, creditsUsed: number) {
  return client.post('/pricing/calculate-discount', { price, creditsUsed })
}
```

Note: The existing client prepends `/api/v1` so the full path is `/api/v1/pricing`.

### Monthly/Yearly Toggle

A simple `el-switch` or `el-radio-group` bound to a reactive `billingCycle` ref. When `yearly`, each card shows the yearly price and a "约 XX/月" hint.

### Credit Discount Calculator

An `el-input-number` for price, an `el-input-number` for credits, and a "计算" button. On submit, calls `calculateDiscount` and displays the result in an `el-card`.

### Login Redirect

`Login.vue` already has a router push to `/` after login. Add a check: if `route.query.redirect` exists, push to that path instead.

## Testing

- `npm run build` (vue-tsc + vite build) must pass.
- Manual verification: visit `/pricing` without login, verify cards render, calculator works, purchase button redirects to login.
- Visit `/pricing` with login, verify purchase button redirects to `/subscriptions`.

## Out of Scope

- Actual payment provider redirect from the pricing page (existing payment-flow handles this).
- Coupon display.
- Multi-currency.
- A/B testing pricing variants.
