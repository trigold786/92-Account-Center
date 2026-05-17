# 92-Account-Center: API to Page Mapping Analysis

> Generated from all backend handler files across 7 microservices.
> Total endpoints: **99** (73 user/admin-facing, 9 internal, 17 config-service not in original scope)

---

## 1. Complete API Inventory

### Auth Service (19 endpoints)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| 1 | POST | `/api/v1/auth/login` | user | LoginHandler.Login | login_handler.go:97 |
| 2 | POST | `/api/v1/auth/refresh` | user | LoginHandler.RefreshToken | login_handler.go:133 |
| 3 | POST | `/api/v1/auth/logout` | user | LoginHandler.Logout | login_handler.go:157 |
| 4 | POST | `/api/v1/auth/biometric/register` | user | LoginHandler.RegisterBiometric | login_handler.go:184 |
| 5 | POST | `/api/v1/auth/biometric/login` | user | LoginHandler.LoginWithBiometric | login_handler.go:199 |
| 6 | POST | `/api/v1/session/create` | internal | SessionHandler.CreateSession | session_handler.go:20 |
| 7 | POST | `/api/v1/session/validate` | internal | SessionHandler.ValidateSession | session_handler.go:40 |
| 8 | GET | `/api/v1/session/user/:user_id` | user | SessionHandler.GetUserSessions | session_handler.go:65 |
| 9 | POST | `/api/v1/session/invalidate` | user | SessionHandler.InvalidateSession | session_handler.go:96 |
| 10 | POST | `/api/v1/session/invalidate-all` | user | SessionHandler.InvalidateAllUserSessions | session_handler.go:121 |
| 11 | POST | `/api/v1/session/refresh` | user | SessionHandler.RefreshSession | session_handler.go:137 |
| 12 | POST | `/api/v1/device/register` | user | DeviceHandler.RegisterDevice | device_handler.go:21 |
| 13 | POST | `/api/v1/device/verify` | user | DeviceHandler.VerifyDevice | device_handler.go:38 |
| 14 | POST | `/api/v1/device/trust` | user | DeviceHandler.TrustDevice | device_handler.go:55 |
| 15 | GET | `/api/v1/device/user/:user_id` | user | DeviceHandler.GetUserDevices | device_handler.go:73 |
| 16 | DELETE | `/api/v1/device/:device_id` | user | DeviceHandler.RemoveDevice | device_handler.go:90 |
| 17 | POST | `/api/v1/qrcode/generate` | user | QRCodeHandler.Generate | qrcode_handler.go:21 |
| 18 | GET | `/api/v1/qrcode/:code_id/status` | user | QRCodeHandler.GetStatus | qrcode_handler.go:31 |
| 19 | POST | `/api/v1/qrcode/:code_id/scan` | user | QRCodeHandler.Scan | qrcode_handler.go:47 |
| 20 | POST | `/api/v1/qrcode/:code_id/confirm` | user | QRCodeHandler.Confirm | qrcode_handler.go:74 |

### Account Service (12 endpoints)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| 21 | POST | `/api/v1/account/register` | user | RegisterHandler.Register | register_handler.go:41 |
| 22 | POST | `/api/v1/account/password/send-verification-code` | user | PasswordHandler.SendVerificationCode | password_handler.go:20 |
| 23 | POST | `/api/v1/account/password/change` | user | PasswordHandler.ChangePassword | password_handler.go:35 |
| 24 | POST | `/api/v1/account/deletion/request` | user | DeletionHandler.RequestDeletion | deletion_handler.go:44 |
| 25 | POST | `/api/v1/account/deletion/cancel` | user | DeletionHandler.CancelDeletion | deletion_handler.go:77 |
| 26 | GET | `/api/v1/account/deletion/status` | user | DeletionHandler.GetDeletionStatus | deletion_handler.go:100 |
| 27 | GET | `/api/v1/account/:user_id/tier` | user | TierHandler.GetTier | tier_handler.go:24 |
| 28 | PUT | `/internal/v1/account/:user_id/tier` | admin | TierHandler.UpdateTier | tier_handler.go:44 |
| 29 | GET | `/api/v1/entitlements/:user_id` | user | EntitlementHandler.GetUserEntitlements | entitlement_handler.go:21 |
| 30 | POST | `/internal/v1/entitlements/consume` | internal | EntitlementHandler.Consume | entitlement_handler.go:38 |
| 31 | POST | `/internal/v1/entitlements/grant` | admin | EntitlementHandler.Grant | entitlement_handler.go:64 |
| 32 | POST | `/api/v1/subscriptions/purchase` | user | SubscriptionHandler.Purchase | subscription_handler.go:21 |
| 33 | POST | `/api/v1/subscriptions/upgrade` | user | SubscriptionHandler.Upgrade | subscription_handler.go:37 |
| 34 | POST | `/api/v1/subscriptions/renew` | user | SubscriptionHandler.Renew | subscription_handler.go:53 |
| 35 | GET | `/api/v1/subscriptions/:user_id` | user | SubscriptionHandler.GetUserSubscriptions | subscription_handler.go:69 |

### Credit Service (8 endpoints)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| 36 | GET | `/api/v1/credits/:user_id/account` | user | CreditHandler.GetAccount | credit_handler.go:21 |
| 37 | GET | `/api/v1/credits/:user_id/transactions` | user | CreditHandler.GetTransactions | credit_handler.go:38 |
| 38 | POST | `/api/v1/credits/calculate-discount` | admin | CreditHandler.CalculateDiscount | credit_handler.go:129 |
| 39 | POST | `/internal/v1/credits/earn` | internal | CreditHandler.EarnCredits | credit_handler.go:58 |
| 40 | POST | `/internal/v1/credits/consume` | internal | CreditHandler.ConsumeCredits | credit_handler.go:79 |
| 41 | POST | `/internal/v1/credits/refund` | internal | CreditHandler.RefundCredits | credit_handler.go:108 |
| 42 | POST | `/api/v1/referral/bind` | user | ReferralHandler.BindReferral | referral_handler.go:21 |
| 43 | POST | `/api/v1/referral/generate-link` | user | ReferralHandler.GenerateLink | referral_handler.go:40 |
| 44 | GET | `/api/v1/referral/:user_id/summary` | user | ReferralHandler.GetSummary | referral_handler.go:64 |

### Compliance Service (16 endpoints)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| 45 | POST | `/api/v1/risk/assess` | internal | RiskHandler.AssessRisk | risk_handler.go:30 |
| 46 | GET | `/api/v1/risk/history/:user_id` | admin | RiskHandler.GetRiskHistory | risk_handler.go:46 |
| 47 | GET | `/api/v1/risk/event/:event_id` | admin | RiskHandler.GetRiskEvent | risk_handler.go:96 |
| 48 | POST | `/api/v1/audit/logs` | internal | AuditHandler.RecordLog | audit_handler.go:26 |
| 49 | POST | `/api/v1/audit/logs/batch` | internal | AuditHandler.RecordBatch | audit_handler.go:49 |
| 50 | GET | `/api/v1/audit/logs/user/:user_id` | admin | AuditHandler.GetLogsByUser | audit_handler.go:69 |
| 51 | GET | `/api/v1/audit/logs` | admin | AuditHandler.GetLogsByTimeRange | audit_handler.go:96 |
| 52 | GET | `/api/v1/audit/logs/:log_id/verify` | admin | AuditHandler.VerifyLogIntegrity | audit_handler.go:139 |
| 53 | POST | `/api/v1/audit/logs/cleanup` | admin | AuditHandler.CleanupOldLogs | audit_handler.go:163 |
| 54 | POST | `/api/v1/kyb/submit` | user | KYBHandler.SubmitEnterprise | kyb_handler.go:21 |
| 55 | POST | `/api/v1/kyb/micro-payment/initiate` | user | KYBHandler.InitiateMicroPayment | kyb_handler.go:37 |
| 56 | POST | `/api/v1/kyb/micro-payment/verify` | user | KYBHandler.VerifyMicroPayment | kyb_handler.go:63 |
| 57 | POST | `/api/v1/kyb/face-verify` | user | KYBHandler.SubmitFaceVerification | kyb_handler.go:87 |
| 58 | GET | `/api/v1/kyb/status/:enterprise_id` | user | KYBHandler.GetEnterpriseStatus | kyb_handler.go:111 |
| 59 | POST | `/api/v1/blacklist/` | admin | BlacklistHandler.AddEntry | blacklist_handler.go:21 |
| 60 | POST | `/api/v1/blacklist/check` | internal | BlacklistHandler.CheckEntry | blacklist_handler.go:35 |
| 61 | DELETE | `/api/v1/blacklist/:type/:value` | admin | BlacklistHandler.RemoveEntry | blacklist_handler.go:49 |
| 62 | GET | `/api/v1/blacklist/` | admin | BlacklistHandler.ListEntries | blacklist_handler.go:59 |
| 63 | POST | `/internal/v1/fraud/check-registration` | internal | (inline handler) | main.go:154 |

### Notification Service (13 endpoints)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| 64 | POST | `/api/v1/sms/send` | user | SMSHandler.SendSMS | sms_handler.go:19 |
| 65 | POST | `/api/v1/sms/verify` | user | SMSHandler.VerifyCode | sms_handler.go:53 |
| 66 | GET | `/api/v1/sms/providers/status` | admin | SMSHandler.GetProviderStatus | sms_handler.go:39 |
| 67 | POST | `/api/v1/email/verify` | user | VerificationEmailHandler.VerifyCode | email_handler.go:37 |
| 68 | POST | `/api/v1/email/otp/send` | user | OTPEmailHandler.SendOTP | email_otp_handler.go:20 |
| 69 | POST | `/api/v1/email/otp/verify` | user | OTPEmailHandler.VerifyOTP | email_otp_handler.go:44 |
| 70 | POST | `/api/v1/email/magic-link/send` | user | OTPEmailHandler.SendMagicLink | email_otp_handler.go:70 |
| 71 | GET | `/api/v1/email/magic-link/verify` | user | OTPEmailHandler.VerifyMagicLink | email_otp_handler.go:90 |
| 72 | POST | `/api/v1/email/send` | internal | OTPEmailHandler.SendEmail | email_otp_handler.go:116 |
| 73 | POST | `/api/v1/push/send` | internal | PushHandler.SendPush | push_handler.go:20 |
| 74 | POST | `/api/v1/push/device/register` | user | PushHandler.RegisterDevice | push_handler.go:42 |
| 75 | GET | `/api/v1/push/user/:user_id/devices` | user | PushHandler.GetUserDevices | push_handler.go:66 |

### Data Product Service (4 endpoints)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| 76 | GET | `/api/v1/data/rfm/:user_id` | admin | RFMHandler.GetRFM | rfm_handler.go:21 |
| 77 | POST | `/api/v1/data/rfm/batch` | admin | RFMHandler.GetRFMBatch | rfm_handler.go:38 |
| 78 | GET | `/api/v1/data/dashboard/overview` | admin | DashboardHandler.GetOverview | dashboard_handler.go:19 |
| 79 | GET | `/api/v1/data/funnel/subscription` | admin | FunnelHandler.GetSubscriptionFunnel | funnel_handler.go:19 |

### Config Service (17 endpoints — discovered, outside original scope)

| # | Method | Path | Category | Handler | Source |
|---|--------|------|----------|---------|--------|
| C1 | GET | `/api/v1/config/groups` | admin | ConfigHandler.ListGroups | config_handler.go:21 |
| C2 | GET | `/api/v1/config/groups/:id` | admin | ConfigHandler.GetGroupByID | config_handler.go:31 |
| C3 | POST | `/api/v1/config/groups` | admin | ConfigHandler.CreateGroup | config_handler.go:50 |
| C4 | PUT | `/api/v1/config/groups/:id` | admin | ConfigHandler.UpdateGroup | config_handler.go:65 |
| C5 | DELETE | `/api/v1/config/groups/:id` | admin | ConfigHandler.DeleteGroup | config_handler.go:86 |
| C6 | GET | `/api/v1/config/items` | admin | ConfigHandler.ListItems | config_handler.go:101 |
| C7 | GET | `/api/v1/config/items/:id` | admin | ConfigHandler.GetItemByID | config_handler.go:119 |
| C8 | POST | `/api/v1/config/items` | admin | ConfigHandler.CreateItem | config_handler.go:138 |
| C9 | PUT | `/api/v1/config/items/:id` | admin | ConfigHandler.UpdateItem | config_handler.go:153 |
| C10 | DELETE | `/api/v1/config/items/:id` | admin | ConfigHandler.DeleteItem | config_handler.go:175 |
| C11 | POST | `/api/v1/config/items/:id/reset-default` | admin | ConfigHandler.ResetItemToDefault | config_handler.go:190 |
| C12 | GET | `/api/v1/config/items/:id/versions` | admin | ConfigHandler.ListVersions | config_handler.go:205 |
| C13 | GET | `/internal/v1/config/items/:code` | internal | ConfigHandler.GetItemByCode | config_handler.go:220 |
| C14 | GET | `/api/v1/config/releases` | admin | ReleaseHandler.ListReleases | release_handler.go:21 |
| C15 | GET | `/api/v1/config/releases/:id` | admin | ReleaseHandler.GetReleaseByID | release_handler.go:35 |
| C16 | POST | `/api/v1/config/releases` | admin | ReleaseHandler.CreateRelease | release_handler.go:54 |
| C17 | PUT | `/api/v1/config/releases/:id/submit` | admin | ReleaseHandler.SubmitRelease | release_handler.go:69 |
| C18 | PUT | `/api/v1/config/releases/:id/approve` | admin | ReleaseHandler.ApproveRelease | release_handler.go:84 |
| C19 | PUT | `/api/v1/config/releases/:id/reject` | admin | ReleaseHandler.RejectRelease | release_handler.go:99 |
| C20 | POST | `/api/v1/config/releases/:id/execute` | admin | ReleaseHandler.ExecuteRelease | release_handler.go:114 |
| C21 | GET | `/api/v1/config/releases/:id/items` | admin | ReleaseHandler.ListReleaseItems | release_handler.go:129 |
| C22 | POST | `/api/v1/config/releases/:id/items` | admin | ReleaseHandler.AddReleaseItem | release_handler.go:144 |
| C23 | GET | `/api/v1/config/audit-logs` | admin | AuditHandler.ListLogs | audit_handler.go:22 |
| C24 | GET | `/api/v1/config/audit-logs/:id` | admin | AuditHandler.GetLogByID | audit_handler.go:53 |
| C25 | GET | `/api/v1/config/roles` | admin | PermissionHandler.ListRoles | permission_handler.go:21 |
| C26 | POST | `/api/v1/config/roles` | admin | PermissionHandler.CreateRole | permission_handler.go:31 |
| C27 | GET | `/api/v1/config/roles/:id/permissions` | admin | PermissionHandler.GetRolePermissions | permission_handler.go:46 |
| C28 | POST | `/api/v1/config/roles/:id/permissions` | admin | PermissionHandler.AddRolePermission | permission_handler.go:61 |
| C29 | GET | `/api/v1/config/users/:userId/roles` | admin | PermissionHandler.GetUserRoles | permission_handler.go:82 |
| C30 | POST | `/api/v1/config/users/:userId/roles` | admin | PermissionHandler.SetUserRole | permission_handler.go:97 |
| C31 | GET | `/api/v1/config/stats` | admin | (inline handler) | main.go:166 |

---

## 2. Feature Coverage Matrix

### 2.1 End-User Self-Service Features (48 endpoints total)

| Feature Domain | API Endpoints | Web | Mobile | Mini Program | Gap |
|---|---|---|---|---|---|
| **Authentication** | | | | | |
| Login (password) | POST /auth/login | ✅ | ✅ | ✅ | — |
| Login (biometric) | POST /auth/biometric/login, /register | ✅ | ✅ | ❌ | Mini Program: no native biometric API |
| Logout | POST /auth/logout | ✅ | ✅ | ✅ | — |
| Token refresh | POST /auth/refresh | ✅ | ✅ | ✅ | — |
| QR code login | POST /qrcode/generate, GET /qrcode/:id/status, POST /qrcode/:id/scan, POST /qrcode/:id/confirm | ✅ | ✅ | ✅ | Web needs QR scanner lib |
| Magic link login | POST /email/magic-link/send, GET /email/magic-link/verify | ✅ | ✅ | ❌ | Mini Program: no email context |
| **Account Management** | | | | | |
| Register | POST /account/register | ✅ | ✅ | ✅ | — |
| Change password | POST /account/password/send-verification-code, /change | ✅ | ✅ | ✅ | — |
| Request deletion | POST /account/deletion/request | ✅ | ✅ | ✅ | — |
| Cancel deletion | POST /account/deletion/cancel | ✅ | ✅ | ✅ | — |
| View deletion status | GET /account/deletion/status | ✅ | ✅ | ✅ | — |
| **Session & Devices** | | | | | |
| View sessions | GET /session/user/:id | ✅ | ✅ | ✅ | — |
| Invalidate session | POST /session/invalidate | ✅ | ✅ | ✅ | — |
| Invalidate all | POST /session/invalidate-all | ✅ | ✅ | ✅ | — |
| Refresh session | POST /session/refresh | ✅ | ✅ | ✅ | — |
| View devices | GET /device/user/:id | ✅ | ✅ | ❌ | Mini Program: limited device mgmt |
| Register device | POST /device/register | ✅ | ✅ | ❌ | — |
| Verify device | POST /device/verify | ✅ | ✅ | ❌ | — |
| Trust device | POST /device/trust | ✅ | ✅ | ❌ | — |
| Remove device | DELETE /device/:id | ✅ | ✅ | ❌ | — |
| **Subscriptions** | | | | | |
| Purchase | POST /subscriptions/purchase | ✅ | ✅ | ✅ | — |
| Upgrade | POST /subscriptions/upgrade | ✅ | ✅ | ✅ | — |
| Renew | POST /subscriptions/renew | ✅ | ✅ | ✅ | — |
| View subscriptions | GET /subscriptions/:id | ✅ | ✅ | ✅ | — |
| **Entitlements** | | | | | |
| View entitlements | GET /entitlements/:id | ✅ | ✅ | ✅ | — |
| **Credit System** | | | | | |
| View credit account | GET /credits/:id/account | ✅ | ✅ | ✅ | — |
| View transactions | GET /credits/:id/transactions | ✅ | ✅ | ✅ | — |
| Calculate discount | POST /credits/calculate-discount | ✅ | ✅ | ✅ | — |
| **Referral** | | | | | |
| Bind referral | POST /referral/bind | ✅ | ✅ | ✅ | — |
| Generate link | POST /referral/generate-link | ✅ | ✅ | ✅ | — |
| View summary | GET /referral/:id/summary | ✅ | ✅ | ✅ | — |
| **KYB (Enterprise)** | | | | | |
| Submit enterprise info | POST /kyb/submit | ✅ | ✅ | ❌ | Mini Program: complex form, but possible |
| Initiate micro-payment | POST /kyb/micro-payment/initiate | ✅ | ✅ | ❌ | — |
| Verify micro-payment | POST /kyb/micro-payment/verify | ✅ | ✅ | ❌ | — |
| Face verification | POST /kyb/face-verify | ✅ | ✅ | ❌ | Needs WebRTC camera |
| View KYB status | GET /kyb/status/:id | ✅ | ✅ | ✅ | — |
| **Notifications** | | | | | |
| Send SMS | POST /sms/send | — | — | — | Backend-initiated via internal calls |
| Verify SMS code | POST /sms/verify | ✅ | ✅ | ✅ | — |
| Verify email code | POST /email/verify | ✅ | ✅ | ✅ | — |
| Send email OTP | POST /email/otp/send | ✅ | ✅ | ❌ | Mini Program: no email input |
| Verify email OTP | POST /email/otp/verify | ✅ | ✅ | ❌ | — |
| Register push device | POST /push/device/register | ❌ | ✅ | ✅ | Web: no push capability |
| View push devices | GET /push/user/:id/devices | ❌ | ✅ | ✅ | — |
| Get tier | GET /account/:id/tier | ✅ | ✅ | ✅ | — |

**Gaps Summary:**
- **Mini Program**: No biometric login, no email-based flows, no device management, limited KYB
- **Web**: No push notifications, needs QR scanner for QR login
- **Mobile**: Full coverage possible

### 2.2 Admin Features (35 endpoints total)

| Feature Domain | API Endpoints | Web Admin | Notes |
|---|---|---|---|
| **Compliance** | | | |
| View risk history | GET /risk/history/:id | ✅ | — |
| View risk event | GET /risk/event/:id | ✅ | ⚠️ NOT IMPLEMENTED (returns stub) |
| View audit logs (user) | GET /audit/logs/user/:id | ✅ | — |
| View audit logs (time) | GET /audit/logs | ✅ | — |
| Verify log integrity | GET /audit/logs/:id/verify | ✅ | — |
| Cleanup old logs | POST /audit/logs/cleanup | ✅ | — |
| **Blacklist** | | | |
| Add entry | POST /blacklist/ | ✅ | — |
| Remove entry | DELETE /blacklist/:type/:value | ✅ | — |
| List entries | GET /blacklist/ | ✅ | — |
| **Subscription Admin** | | | |
| Update user tier | PUT /internal/account/:id/tier | ✅ | — |
| Grant entitlements | POST /internal/entitlements/grant | ✅ | — |
| **Data Analytics** | | | |
| View RFM score | GET /data/rfm/:id | ✅ | — |
| Batch RFM query | POST /data/rfm/batch | ✅ | — |
| Dashboard overview | GET /data/dashboard/overview | ✅ | — |
| Subscription funnel | GET /data/funnel/subscription | ✅ | — |
| Calculate discount | POST /credits/calculate-discount | ✅ | — |
| SMS provider status | GET /sms/providers/status | ✅ | — |
| **Config Management** | | | |
| Config groups CRUD | GET/POST/PUT/DELETE /config/groups/* | ✅ | — |
| Config items CRUD | GET/POST/PUT/DELETE /config/items/* | ✅ | — |
| Config releases | Full release workflow (8 endpoints) | ✅ | — |
| Config audit logs | GET /config/audit-logs | ✅ | — |
| **Permissions** | | | |
| Role management | GET/POST /config/roles | ✅ | — |
| Permission assignment | GET/POST /config/roles/:id/permissions | ✅ | — |
| User-role mapping | GET/POST /config/users/:userId/roles | ✅ | — |
| Config stats | GET /config/stats | ✅ | — |

### 2.3 Internal/Backend-Only Endpoints (9 endpoints)

These should NOT be exposed in any frontend:
- POST `/internal/v1/account/:user_id/tier` (admin only, listed above)
- POST `/internal/v1/entitlements/consume`
- POST `/internal/v1/credits/earn`
- POST `/internal/v1/credits/consume`
- POST `/internal/v1/credits/refund`
- POST `/api/v1/risk/assess`
- POST `/api/v1/audit/logs` (single)
- POST `/api/v1/audit/logs/batch`
- POST `/internal/v1/fraud/check-registration`
- POST `/api/v1/push/send`
- POST `/api/v1/email/send`
- GET `/internal/v1/config/items/:code`

---

## 3. Prioritized Page List by Platform

### 3.1 Web (new web-ui app) — 14 pages

| Priority | Page | Path | Features (API Endpoints) | Category | Effort |
|----------|------|------|--------------------------|----------|--------|
| P0 | **Login** | `/login` | POST /auth/login, POST /auth/biometric/login, POST /auth/biometric/register | user | 3 days |
| P0 | **Register** | `/register` | POST /account/register, POST /sms/send, POST /sms/verify | user | 3 days |
| P0 | **Dashboard / Home** | `/` | GET /account/:id/tier, GET /subscriptions/:id | user | 2 days |
| P0 | **Account Settings** | `/settings` | POST /password/change, POST /logout | user | 2 days |
| P1 | **Subscriptions** | `/subscriptions` | GET /subscriptions/:id, POST /purchase, POST /upgrade, POST /renew, POST /credits/calculate-discount | user | 4 days |
| P1 | **Credit Wallet** | `/credits` | GET /credits/:id/account, GET /credits/:id/transactions | user | 2 days |
| P1 | **Referral** | `/referral` | POST /referral/generate-link, GET /referral/:id/summary | user | 1.5 days |
| P1 | **Entitlements** | `/entitlements` | GET /entitlements/:id | user | 1 day |
| P1 | **Sessions & Devices** | `/security` | GET /session/user/:id, POST /invalidate, POST /invalidate-all, GET /device/user/:id, POST /device/register, POST /device/trust, DELETE /device/:id | user | 3 days |
| P2 | **QR Code Login** | `/qr-login` | POST /qrcode/generate, GET /qrcode/:id/status | user | 2 days |
| P2 | **KYB (Enterprise)** | `/kyb` | POST /kyb/submit, POST /kyb/micro-payment/initiate, POST /kyb/micro-payment/verify, POST /kyb/face-verify, GET /kyb/status/:id | user | 5 days |
| P2 | **Delete Account** | `/delete-account` | POST /account/deletion/request, POST /account/deletion/cancel, GET /account/deletion/status | user | 1.5 days |
| P0 | **Admin Dashboard** | `/admin` | GET /data/dashboard/overview, GET /data/funnel/subscription | admin | 3 days |
| P0 | **Admin - Users** | `/admin/users` | GET /risk/history/:id, GET /audit/logs/user/:id, GET /entitlements/:id | admin | 3 days |
| P1 | **Admin - Blacklist** | `/admin/blacklist` | GET /blacklist/, POST /blacklist/, DELETE /blacklist/:type/:value | admin | 1.5 days |
| P1 | **Admin - Config** | `/admin/config` | Full config CRUD (10+ endpoints) | admin | 5 days |
| P1 | **Admin - Releases** | `/admin/releases` | Full release workflow (8 endpoints) | admin | 4 days |
| P2 | **Admin - Permissions** | `/admin/permissions` | Role & permission management (6 endpoints) | admin | 3 days |
| P2 | **Admin - Audit Logs** | `/admin/audit-logs` | GET /audit/logs, GET /audit/logs/:id/verify, POST /audit/logs/cleanup | admin | 2 days |
| P2 | **Admin - SMS Providers** | `/admin/sms` | GET /sms/providers/status | admin | 0.5 days |
| P2 | **Admin - RFM Analysis** | `/admin/rfm` | GET /data/rfm/:id, POST /data/rfm/batch | admin | 2 days |

**Web Total: 21 pages, ~53 person-days**

### 3.2 Mobile (iOS/Android) — What exists vs missing

**Covers all user-facing features listed in Web above**, with these platform-specific notes:

| Feature | Status | Note |
|---------|--------|------|
| Biometric login | ✅ Native | Use platform biometric APIs (FaceID, fingerprint) |
| Push notifications | ✅ Full support | POST /push/device/register, GET /push/user/:id/devices |
| QR login scanning | ✅ Native | Camera-based QR scanning |
| Face verification (KYB) | ✅ Native | Camera integration for face capture |
| Device management | ✅ Full control | Platform OS-level device identification |
| All other user features | ✅ Full coverage | Same as Web |

**Mobile specific pages needed:** 14 user pages (same as Web minus QR code login page if merged with login)

### 3.3 WeChat Mini Program — What exists vs missing

| Feature | Status | Gap |
|---------|--------|-----|
| Login (password) | ✅ | Standard |
| Login (phone number) | ✅ | Use wx.getPhoneNumber |
| QR code scanning | ✅ | Native wx.scanCode API |
| Logout | ✅ | Standard |
| Register | ✅ | Use phone number |
| Change password | ✅ | Standard |
| View/Manage subscriptions | ✅ | Standard |
| View credits & transactions | ✅ | Standard |
| Referral | ✅ | wx.shareAppMessage for sharing |
| Referral bind | ✅ | Scan referral QR / paste code |
| Delete account | ✅ | Standard |
| SMS verification | ✅ | wx login with phone number |
| View tier/entitlements | ✅ | Standard |
| **Biometric login** | ❌ MISSING | No biometric API in Mini Programs |
| **Email OTP/Magic link** | ❌ MISSING | Mini Program has no email input context |
| **Device management** | ❌ MISSING | No device fingerprint API |
| **KYB - enterprise submit** | ⚠️ Limited | Complex form on mobile webview |
| **KYB - micro-payment** | ❌ MISSING | Requires bank integration |
| **KYB - face verification** | ❌ MISSING | No native face compare API |
| **Push device registration** | ⚠️ Limited | Use wx settings for notifications |
| **Admin features** | ❌ MISSING | Mini Program is typically user-facing only |

**Mini Program pages needed (user-only):** 12 pages (fewer than Web due to platform limitations)

---

## 4. Key Findings & Anomalies

| # | Issue | Severity | Details |
|---|-------|----------|---------|
| 1 | **Unregistered handler** | Medium | `VerificationEmailHandler.SendVerificationCode` (email_handler.go:19) is defined but NOT registered in any route |
| 2 | **Stub endpoint** | Low | `GET /api/v1/risk/event/:event_id` (risk_handler.go:96) returns "not implemented" |
| 3 | **Email verification gap** | Low | There's no endpoint to *send* an email verification code (only to verify existing codes) — the `/api/v1/email/verify` endpoint only verifies, and the send endpoint was not registered |
| 4 | **Password send-vc missing route ref** | Low | `POST /account/password/send-verification-code` exists but uses `ContactType`/`ContactValue` from model — unclear if it routes through SMS or email |
| 5 | **Config service out-of-scope** | Info | 31 endpoints found in config-service, not in original audit scope but included for completeness |
| 6 | **No web UI exists yet** | Info | All frontends (Web, Mobile, Mini Program) are greenfield |

---

## 5. Implementation Recommendations

### Phase 1 — Core Auth & Account (4 weeks)
- Web: Login, Register, Dashboard, Account Settings, Session/Device management
- All platforms: Core authentication flows

### Phase 2 — Commerce & Credits (3 weeks)
- Web: Subscriptions (purchase/upgrade/renew), Credit wallet, Entitlements view
- Mobile: Same + push notification integration
- Mini Program: Same (minus push device reg)

### Phase 3 — Enterprise & Admin (4 weeks)
- Web: KYB flow, Admin Dashboard, Blacklist management, Config management
- Mobile: KYB face verification, Admin restricted
- Mini Program: Skipped for KYB/Admin

### Phase 4 — Analytics & Admin Advanced (3 weeks)
- Web: RFM analysis, Release workflow, Permissions, Audit logs
- Admin portal consolidation
