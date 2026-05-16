# UI Redesign: Dark Tech Theme — Account Center Mobile App

> **Goal:** Redesign all iOS and Android screens with a dark tech-forward visual identity inspired by Apple's precision, using deep backgrounds, aurora purple-blue gradients, and open-source typography.

---

## Design System

### Color Palette

| Token | Hex | Usage |
|-------|-----|-------|
| `bgPrimary` | `#0D1117` | Root background |
| `bgCard` | `#161B22` | Card surface |
| `bgInput` | `#161B22` | Input field background |
| `brandPrimary` | `#6C63FF` | Primary purple |
| `brandSecondary` | `#00D4FF` | Secondary cyan |
| `brandGradient` | `linear-gradient(135deg, #6C63FF, #00D4FF)` | Buttons, highlights |
| `textPrimary` | `#FFFFFF` (87%) | Primary text |
| `textSecondary` | `#8B949E` (60%) | Secondary text |
| `divider` | `#21262D` | Hairline separators |
| `danger` | `#FF4757` | Risk/consumption alerts |
| `success` | `#2ED573` | Earnings, active status |
| `tierFree` | `#8B949E` | Tier 0 (gray) |
| `tierBasic` | `#6C63FF` | Tier 1 (purple-blue) |
| `tierPremium` | `#FF9800` | Tier 2 (orange) |
| `tierEnterprise` | `#7B1FA2` | Tier 3 (purple) |

### Typography

| Role | Font | Weight | Size |
|------|------|--------|------|
| Hero title | Space Grotesk | Bold | 28pt |
| Page title | Space Grotesk | Bold | 22pt |
| Section header | Inter | Semibold | 15pt |
| Body | Inter | Regular | 15pt |
| Caption | Inter | Regular | 12pt |
| Balance amount | Space Grotesk | Bold | 36pt |

Both fonts are SIL Open Font License — free for commercial use.

### Card Style

- Background: `bgCard` (`#161B22`)
- Corner radius: 16pt
- No hard borders — use `shadow` (iOS) / `elevation` (Android) or a 0.5pt `#21262D` inner stroke
- Inner padding: 16-20pt
- Optional glass morphism: `background(.ultraThinMaterial)` (iOS) or semi-transparent Surface (Android)

### Icon Style

- iOS: SF Symbols — prefer `.fill` variants, tinted with `brandGradient`
- Android: Material Icons — filled style, tinted with `brandPrimary`
- Feature row icons: consistently colored purple-blue, 24pt

### Button Style

| Type | Style |
|------|-------|
| Primary | `brandGradient` fill, cornerRadius(12), white text, 56pt height |
| Outline | 1pt `brandPrimary` border, transparent fill, `brandPrimary` text |
| Text | Plain text in `brandSecondary`, no background |

---

## Screen Designs

### 1. Login Screen

**Layout:**
- Full-screen `#0D1117` background + subtle radial gradient vignette (purple-blue glow at top/bottom)
- Centered layout with top padding for visual breathing room
- Logo area: wallet/V-shaped icon + "账户中心" in Space Grotesk 24pt Bold
- Subtitle: "登录您的账户" in Inter 14pt `textSecondary`
- Input fields: `bgInput` fill, no outer border, inner focused glow (1pt `brandSecondary` bottom line)
- Segmented control: "密码登录" / "验证码登录" — thin-line style, selected segment has gradient underline
- Password field shows/hides with SF Symbol eye icon
- Verification code mode: code input + "发送验证码" button (countdown state shows seconds, disabled gray)
- Primary button: `brandGradient`, cornerRadius(12), subtle scale animation on press
- Bottom row: "还没有账号？" + "立即注册" in `brandSecondary`
- Error messages: `danger` color, Inter 12pt, auto-hide after 3s

**States:**
- Idle, focused, error, loading (ProgressView replaces button text)

### 2. Register Screen

**Layout:**
- ScrollView with same input styling as Login
- Fields: 手机号, 验证码, 账户ID, 密码, 确认密码, 推荐码(可选)
- SMS code button: same pattern as Login
- Terms checkbox: custom purple-blue checkbox, "我已阅读并同意服务条款和隐私政策"
- Primary "注册" button with gradient
- Bottom: "已有账号？" + "立即登录"
- Same error display as Login

**Validation:**
- Inline field validation on focus loss (red border + message)
- Password match check
- Phone format check (11 digits)

### 3. Home Screen (Dashboard)

**Layout:**
- NOT a native List — custom ScrollView composition with independent card sections
- TopAppBar: "用户中心" with semi-transparent background
- **User Card** (glass-style):
  - Avatar: 64pt circle with `brandGradient` fill, centered initial letter
  - Account ID: Space Grotesk Bold
  - Phone status: Inter caption `textSecondary`
  - Tier badge: small pill "Lv.2" with tier color
- **RFM Card** (spacer 12pt):
  - Subtle gradient border (1pt, `brandGradient`, opacity 0.3)
  - Left: "🎯" emoji
  - Content: `rfmSegmentCn` (Space Grotesk 14pt) + "RFM `rfmSegment`" (Inter caption, `textSecondary`)
- **Feature List** (spacer 24pt):
  - Section title "功能" in Inter 13pt `textSecondary`, uppercase tracking
  - 4 feature rows: 56pt fixed height, left icon (SF Symbol/24pt), label, right chevron `>`
  - Icons consistently tinted `brandPrimary`
- **Logout Button**:
  - bgCard background, `danger` text, cornerRadius(12), full width
  - No icon — clean text-only approach

**Loading:**
- Skeleton shimmer on initial load (optional enhancement)

### 4. Subscription Management Screen

**Layout:**
- TopAppBar with back button + "订阅管理"
- **Tier Badge Card**:
  - Left: 12pt colored circle (tier color)
  - Label: tier name (e.g. "高级版"), right: "Lv.2" small pill badge
- **Active Subscription Card** (if exists):
  - bgCard with subtle `brandPrimary` left border (3pt)
  - Key-value rows: 状态 (with green dot), 开始时间, 结束时间, 金额, 支付方式
  - Status "生效中" in `success` green
- **Empty State** (if no active subscription):
  - "暂无活跃订阅" centered text + 3 plan reference rows (基础版 ¥99/月, etc.)
- **Subscription History**:
  - Section title "订阅历史"
  - Compact rows: status left, price right, date range below
  - No borders between items — just vertical padding 8pt

### 5. Credits Center Screen

**Layout:**
- TopAppBar with back button + "积分中心"
- **Balance Card**:
  - Centered layout
  - Label: "当前积分" (Inter 13pt, `textSecondary`)
  - Amount: Space Grotesk 36pt Bold with `brandGradient` text (iOS: `foregroundStyle(.linearGradient...)`)
  - Status: "账户正常" in `success` green
- **Referral Card**:
  - "邀请推广" section header
  - Stats: 邀请人数, 活跃邀请, 累计收益
  - "复制推荐链接" button — outline style with `brandGradient` border
- **Transaction List**:
  - "交易记录" section header
  - Each row: left icon (⊕ green / ⊖ red / ⟲ green), description, date, amount (± with color)
  - "+100" = `success` green, "-50" = `danger` red
  - `hasMore` → "加载更多..." text button at bottom
  - Auto-load on scroll-to-bottom (iOS) / manual button (Android acceptable)

**Transaction Type Mapping:**
- Positive: `earn`, `referral_bonus`, `refund` → green `+`
- Negative: `consume`, `subscription_payment` → red `-`

### 6. Security Settings Screen

**Layout:**
- TopAppBar with back button + "安全设置"
- **Risk Events Section**:
  - Each event: left colored dot (critical=red, high=orange, medium=yellow, low=green), event type, timestamp, right risk level badge
  - Empty: "暂无风险事件"
- **Devices Section**:
  - Each device: platform icon (iOS: SF Symbol `iphone`/`desktopcomputer`; Android: `PhoneAndroid`/`Computer`), device name, last active, status pill ("活跃中" green / "离线" gray)
  - Read-only display — no remove action (no backend endpoint)
  - Empty: "暂无设备记录"

### 7. About Screen

**Layout:**
- TopAppBar with back button + "关于"
- **Logo Section** (centered):
  - Brand icon (V-shape or wallet symbol) with `brandGradient`
  - "账户中心" Space Grotesk 20pt
  - Version line: "Version 1.0 Build 1"
- **Info Section**:
  - Key-value rows: 版本号, 构建号
- **Legal Section**:
  - "服务条款", "隐私政策" — tappable rows with chevron
  - Currently non-interactive (placeholder for future WebView)
- **Footer**:
  - "© 2026 Account Center. All rights reserved." — Inter 11pt, `textSecondary`

---

## Animation Guidelines

| Element | Animation |
|---------|-----------|
| Page transition | Default platform push (no custom) |
| Button press | Scale 0.97 → 1.0, 100ms ease-out |
| Balance number | Count-up animation on appear |
| Card appear | Fade-in + slight vertical slide (20pt, 300ms) |
| Error message | Fade in/out, 300ms |
| Refresh | Platform standard pull-to-refresh |

---

## Implementation Notes

### iOS Specifics
- Custom `Color` extension for all design tokens
- `ViewModifier` for card style, gradient button, glowing input
- `ScrollView` + `VStack` with sections instead of `List` for Home (others can stay `List` if `listStyle(.insetGrouped)` is adapted)
- `UIViewRepresentable` for custom fonts if needed
- SF Symbols preferred over custom icon assets
- `AnyLayout` / `LabeledContent` for key-value rows

### Android Specifics
- Custom `darkColorScheme` in `Theme.kt` with new tokens
- `Card` composable wrapper with `bgCard` surface color + 16.dp shape
- `TextField` styling via `TextFieldDefaults` with custom colors
- `gradientBrush()` extension for `Brush.horizontalGradient`
- Downloadable Fonts API for Inter + Space Grotesk (or bundle TTF in `res/font/`)
- Custom `Modifier.glowBorder()` extension (optional)

### File Scope
- **iOS**: ~12 files modified (Color extensions, reusable modifiers, 7 screen revisions)
- **Android**: ~10 files modified (Color.kt, Theme.kt, 7 screen revisions, reusable composables)

---

## File Change Summary

| File | Change |
|------|--------|
| `ios/AccountCenter/Extensions/Color+Theme.swift` | New — all palette colors |
| `ios/AccountCenter/Extensions/View+Style.swift` | New — card/button/input modifiers |
| `ios/AccountCenter/Features/Login/LoginView.swift` | Rewrite — dark theme layout |
| `ios/AccountCenter/Features/Register/RegisterView.swift` | Rewrite — dark theme layout |
| `ios/AccountCenter/Features/Home/HomeView.swift` | Rewrite — card-based layout |
| `ios/AccountCenter/Features/Subscription/SubscriptionView.swift` | Rewrite — glass cards |
| `ios/AccountCenter/Features/Credits/CreditsView.swift` | Rewrite — gradient balance |
| `ios/AccountCenter/Features/Security/SecurityView.swift` | Rewrite — dark list |
| `ios/AccountCenter/Features/About/AboutView.swift` | Rewrite — centered logo |
| `android/.../ui/theme/Color.kt` | Rewrite — dark palette |
| `android/.../ui/theme/Theme.kt` | Rewrite — dark scheme + fonts |
| `android/.../ui/theme/Shape.kt` | New — 16dp shapes |
| `android/.../ui/components/AppCard.kt` | New — reusable card composable |
| `android/.../ui/.../LoginScreen.kt` | Rewrite — dark theme |
| (+6 more screen files) | Rewrite — apply theme |

---

## Out of Scope

- Server-driven UI changes
- Backend API modifications
- New feature pages (all already exist)
- Icon/assets replacement (system icons retained)
- Accessibility audit (deferred)
