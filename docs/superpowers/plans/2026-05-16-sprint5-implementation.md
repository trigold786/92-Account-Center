# Sprint 5 Implementation Plan — Mobile Feature Pages Expansion (v2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Add 4 new mobile feature pages + RFM profile card to both iOS and Android, building on Sprint 4's mobile foundation.

**Architecture:** All data via api-gateway (`localhost:30300`) to existing backend services (account, credit, compliance, notification, data-product). No new backend work.

**Platform constraints:**
- iOS: Swift 5.9+, SwiftUI, ObservableObject, URLSession async/await, Keychain, iOS 16+
- Android: Kotlin 1.9+, Jetpack Compose Material3, Hilt, Retrofit/OkHttp, API 24+

**Spec:** `docs/superpowers/specs/2026-05-16-sprint4-design.md`

---

## API Response Wrapping Convention

Many credit/referral/risk endpoints return wrapped responses:
```json
{"code": 200, "message": "success", "data": { ...actual model... }}
```

### iOS: Add generic response wrapper
```swift
struct ApiDataResponse<T: Codable>: Codable {
    let code: Int
    let message: String?
    let data: T?
}
```

### Android: Add generic response wrapper
```kotlin
data class ApiDataResponse<T>(
    val code: Int,
    val message: String? = null,
    val data: T? = null
)
```

Both platforms use these only for wrapped endpoints. Login/auth endpoints return flat JSON (no wrapper) and continue using direct decode.

---

## Task 1: Models and Network Layer

### Step 1.1 — Both platforms: Add ApiDataResponse wrapper + iOS AnyCodable → [String: String]

### Step 1.2 — iOS: Add models

Create `ios/AccountCenter/Models/Subscription.swift`:
- `Subscription`, `TierInfo`

Create `ios/AccountCenter/Models/Credit.swift`:
- `CreditAccount`, `Transaction`, `TransactionList`, `ReferralSummary`, `CalculateDiscountRequest`, `DiscountInfo`

Create `ios/AccountCenter/Models/Risk.swift`:
- `RiskEvent` (details: `[String: String]`, NOT AnyCodable), `RiskHistoryResponse` (wrapped: `{code, message, data: {events, limit}}`), `RiskHistoryData`, `PushDevice`, `DeviceList`

### Step 1.3 — iOS: Add endpoint constants

Append to `ios/AccountCenter/Core/Network/Endpoints.swift` — add endpoint helpers for: subscriptions, tier, credits, transactions, referral, discount, risk history, devices, referral link generate

### Step 1.4 — Android: Add models

Create `android/.../model/Subscription.kt` — same models as iOS + `ApiDataResponse<T>` generic wrapper
- `ApiDataResponse`, `Subscription`, `TierInfo`, `CreditAccount`, `Transaction`, `TransactionList`, `ReferralSummary`, `RiskEvent`, `RiskHistoryData`, `PushDevice`, `DeviceList`, `CalculateDiscountRequest`, `DiscountInfo`

### Step 1.5 — Android: Add API endpoints to ApiClient.kt

Add Retrofit methods for: getUserSubscriptions, getUserTier, getCreditAccount, getTransactions, getReferralSummary, calculateDiscount, getRiskHistory, getUserDevices

---

## Task 2: iOS — Subscription Management Page

**Create:** `SubscriptionView.swift`, `SubscriptionViewModel.swift`
**Modify:** `HomeView.swift` (link feature item)

- View: tier badge (colored by level 0-4), current subscription card (plan name, status, dates), fallback "无活跃订阅" if empty
- No purchase/upgrade POST — read-only display only
- ViewModel: loads subscriptions + tier from API

---

## Task 3: iOS — Credits Center Page

**Create:** `CreditsView.swift`, `CreditsViewModel.swift`
**Modify:** `HomeView.swift`

- View: balance card (large number), transaction history list (type icon, +/- amount, date grouped), referral summary card (total referees, earned coins, share referral link button)
- Share button copies referral link to clipboard (uses `UIPasteboard`)
- Pagination: pull-to-refresh loads page 1; scroll-to-bottom loads next page
- ViewModel: manages pagination state, shares `AuthManager` via init

---

## Task 4: iOS — Security & About Page

**Create:** `SecurityView.swift`, `SecurityViewModel.swift`, `AboutView.swift`
**Modify:** `HomeView.swift`

- Security: risk events list (level badge red/orange/green, timestamp), linked devices (platform icon, name, last active, swipe-to-delete calls device logout)
- About: app version (from Bundle), build number, "服务条款" / "隐私政策" links (plain Text, no webview needed for now)
- ViewModel: loads risk history + devices, provides device removal action

---

## Task 5: iOS — HomeView RFM Card Enhancement

**Modify:** `HomeView.swift`

- Under user card, add RFM segment badge (if data available)
- Background fetch on HomeViewModel init or pull-to-refresh
- Display: e.g. "🎯 高价值用户 · RFM S1" (segment_cn string)
- Uses API: `GET /api/v1/data/rfm/:user_id`

**Create/Modify:** `HomeViewModel.swift` — add RFM loading

---

## Task 6: Android — Subscription Management Page

**Create:** `SubscriptionScreen.kt`, `SubscriptionViewModel.kt`
**Modify:** `HomeScreen.kt` (link), `NavGraph.kt` (add route)

Mirrors Task 2. Material3 Cards, tier badge colors, empty state.

---

## Task 7: Android — Credits Center Page

**Create:** `CreditsScreen.kt`, `CreditsViewModel.kt`
**Modify:** `HomeScreen.kt`, `NavGraph.kt`

Mirrors Task 3. Material3 styling, LazyColumn with pagination, share intent for referral link.

---

## Task 8: Android — Security & About Page

**Create:** `SecurityScreen.kt`, `SecurityViewModel.kt`, `AboutScreen.kt`
**Modify:** `HomeScreen.kt`, `NavGraph.kt`

Mirrors Task 4. Device removal via AlertDialog, LazyColumn for risk events.

---

## Task 9: Android — HomeScreen RFM Card Enhancement

**Modify:** `HomeViewModel.kt` — add RFM state + API call
**Modify:** `HomeScreen.kt` — display RFM segment below user card

Alternatively, add RFM data to `UserRepository` and expose via `UserDisplay`.

---

## Task 10: Navigation Updates (Consolidated)

### iOS: `HomeView.swift`
Single edit: replace all 4 `NavigationLink(destination: EmptyView())` with real destinations:
- 订阅管理 → SubscriptionView
- 积分中心 → CreditsView
- 安全设置 → SecurityView
- 关于 → AboutView

### Android: `NavGraph.kt`
Add 4 new routes + composable blocks:
- `Screen.Subscription → SubscriptionScreen`
- `Screen.Credits → CreditsScreen`
- `Screen.Security → SecurityScreen`
- `Screen.About → AboutScreen`

### Android: `HomeScreen.kt`
Replace `FeatureItem(...) { }` empty onClick closures with navigation lambdas.

---

## Task 11: Unit Tests

### iOS (`ios/AccountCenterTests/`)
- `SubscriptionViewModelTests.swift` — mock APIClient, test load states
- `CreditsViewModelTests.swift` — test pagination, balance display

### Android (`android/app/src/test/java/com/accountcenter/`)
- `SubscriptionViewModelTest.kt`
- `CreditsViewModelTest.kt`

---

## API Reference (All Verified Existing)

| Feature | Method | Path | Wrapped | Notes |
|---------|--------|------|---------|-------|
| Get subscriptions | GET | `/api/v1/subscriptions/:user_id` | No | Returns `[Subscription]` |
| Get tier | GET | `/api/v1/account/:user_id/tier` | No | Flat `{user_id, identity_tier}` |
| Get credit account | GET | `/api/v1/credits/:user_id/account` | Yes | `{code, message, data: CreditAccount}` |
| Get transactions | GET | `/api/v1/credits/:user_id/transactions` | Yes | `{code, message, data: TransactionList}` |
| Calculate discount | POST | `/api/v1/credits/calculate-discount` | Yes | `{code, message, data: DiscountInfo}` |
| Referral summary | GET | `/api/v1/referral/:user_id/summary` | Yes | `{code, message, data: ReferralSummary}` |
| Generate link | POST | `/api/v1/referral/generate-link` | Yes | `{code, message, data: {referral_code, link}}` |
| Risk history | GET | `/api/v1/risk/history/:user_id` | Yes | `{code, message, data: {events, limit}}` |
| User devices | GET | `/api/v1/push/user/:user_id/devices` | Yes | `{code, message, data: {devices}}` |
| RFM score | GET | `/api/v1/data/rfm/:user_id` | No | Flat `{user_id, rfm_segment, ...}` |

---

## File Change Summary

| Task | iOS (+/modified) | Android (+/modified) |
|------|------------------|---------------------|
| 1 — Models + Network | +3 models, +1 endpoint edit | +1 model file, +1 ApiClient edit |
| 2 — Subscription | +2, mod HomeView | +2, mod HomeScreen + NavGraph |
| 3 — Credits | +2, mod HomeView | +2, mod HomeScreen + NavGraph |
| 4 — Security & About | +3, mod HomeView | +3, mod HomeScreen + NavGraph |
| 5 — RFM Card | mod HomeViewModel | mod HomeViewModel + HomeScreen |
| 10 — Navigation | mod HomeView (consolidated) | mod NavGraph + HomeScreen |
| 11 — Tests | +2 | +2 |

**Total: ~12 new iOS files, ~12 new Android files, 4 modified shared files per platform**
