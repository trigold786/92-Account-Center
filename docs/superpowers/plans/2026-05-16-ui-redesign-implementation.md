# UI Redesign — Dark Tech Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply dark tech visual identity (#0D1117 base, violet-cyan gradient, Inter/Space Grotesk) to all iOS and Android screens.

**Architecture:** Platform-native theme layer (Color extension + ViewModifier for iOS; Color.kt + Theme.kt for Android) → reusable components → screen rewrites. No backend changes.

**Tech Stack:** iOS Swift 5.9+ SwiftUI, Android Kotlin 1.9+ Jetpack Compose, Inter + Space Grotesk (SIL OFL)

**Spec:** `docs/superpowers/specs/2026-05-16-ui-redesign-tech-dark.md`

---

## File Structure

### iOS New Files
| File | Responsibility |
|------|---------------|
| `ios/AccountCenter/Extensions/Color+Theme.swift` | All 15 palette colors, gradient helper |
| `ios/AccountCenter/Extensions/View+Style.swift` | `.cardStyle()`, `.gradientButton()`, `.glowingInput()`, `.sectionTitle()` |

### iOS Modified Files
| File | Change |
|------|--------|
| `ios/AccountCenter/Features/Login/LoginView.swift` | Full dark theme rewrite |
| `ios/AccountCenter/Features/Register/RegisterView.swift` | Full dark theme rewrite |
| `ios/AccountCenter/Features/Home/HomeView.swift` | Card-based layout rewrite |
| `ios/AccountCenter/Features/Subscription/SubscriptionView.swift` | Glass card rewrite |
| `ios/AccountCenter/Features/Credits/CreditsView.swift` | Gradient balance + dark list |
| `ios/AccountCenter/Features/Security/SecurityView.swift` | Dark list + level colors |
| `ios/AccountCenter/Features/About/AboutView.swift` | Centered logo layout |

### Android New Files
| File | Responsibility |
|------|---------------|
| `android/.../ui/theme/Shape.kt` | 16dp rounded shape for all cards |
| `android/.../ui/components/AppCard.kt` | Reusable dark card composable |
| `android/.../ui/components/GradientButton.kt` | Reusable gradient button |

### Android Modified Files
| File | Change |
|------|--------|
| `android/.../ui/theme/Color.kt` | Replace palette with dark tokens |
| `android/.../ui/theme/Theme.kt` | darkColorScheme with new tokens, add fonts |
| `android/.../ui/theme/Type.kt` | Add Inter + Space Grotesk typography |
| `android/.../ui/login/LoginScreen.kt` | Dark theme rewrite |
| `android/.../ui/register/RegisterScreen.kt` | Dark theme rewrite |
| `android/.../ui/home/HomeScreen.kt` | Card-based layout rewrite |
| `android/.../ui/subscription/SubscriptionScreen.kt` | Dark card rewrite |
| `android/.../ui/credits/CreditsScreen.kt` | Gradient balance rewrite |
| `android/.../ui/security/SecurityScreen.kt` | Dark list rewrite |
| `android/.../ui/about/AboutScreen.kt` | Centered logo layout |

---

### Task 1: iOS Design Tokens

**Files:**
- Create: `ios/AccountCenter/Extensions/Color+Theme.swift`
- Create: `ios/AccountCenter/Extensions/View+Style.swift`

- [ ] **Step 1: Create Color+Theme.swift**

```swift
import SwiftUI

extension Color {
    static let bgPrimary = Color(hex: "#0D1117")
    static let bgCard = Color(hex: "#161B22")
    static let bgInput = Color(hex: "#161B22")
    static let brandPrimary = Color(hex: "#6C63FF")
    static let brandSecondary = Color(hex: "#00D4FF")
    static let textPrimary = Color.white.opacity(0.87)
    static let textSecondary = Color(hex: "#8B949E")
    static let divider = Color(hex: "#21262D")
    static let danger = Color(hex: "#FF4757")
    static let success = Color(hex: "#2ED573")
    static let tierFree = Color(hex: "#8B949E")
    static let tierBasic = Color(hex: "#6C63FF")
    static let tierPremium = Color(hex: "#FF9800")
    static let tierEnterprise = Color(hex: "#7B1FA2")

    static let brandGradient = LinearGradient(
        colors: [.brandPrimary, .brandSecondary],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        let scanner = Scanner(string: hex)
        var int: UInt64 = 0
        scanner.scanHexInt64(&int)
        let r, g, b: UInt64
        (r, g, b) = ((int >> 16) & 0xFF, (int >> 8) & 0xFF, int & 0xFF)
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: 1
        )
    }
}
```

- [ ] **Step 2: Create View+Style.swift**

```swift
import SwiftUI

extension View {
    func cardStyle() -> some View {
        self
            .padding(16)
            .background(Color.bgCard)
            .cornerRadius(16)
    }

    func gradientButton() -> some View {
        self
            .font(.system(size: 16, weight: .semibold))
            .foregroundColor(.white)
            .frame(maxWidth: .infinity)
            .frame(height: 52)
            .background(
                LinearGradient(
                    colors: [.brandPrimary, .brandSecondary],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
            )
            .cornerRadius(12)
    }

    func glowingInput() -> some View {
        self
            .padding()
            .background(Color.bgInput)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(Color.brandSecondary.opacity(0.0), lineWidth: 1)
            )
    }

    func sectionTitle() -> some View {
        self
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(.textSecondary)
            .textCase(.uppercase)
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add ios/AccountCenter/Extensions/
git commit -m "feat(ui): add iOS design tokens and style modifiers"
```

---

### Task 2: iOS Login + Register Dark Theme

**Files:**
- Modify: `ios/AccountCenter/Features/Login/LoginView.swift`
- Modify: `ios/AccountCenter/Features/Register/RegisterView.swift`

- [ ] **Step 1: Rewrite LoginView.swift**

```swift
import SwiftUI

struct LoginView: View {
    @StateObject private var viewModel = LoginViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                VStack(spacing: 8) {
                    Image(systemName: "person.circle.fill")
                        .font(.system(size: 48))
                        .foregroundStyle(Color.brandGradient)
                    Text("账户中心")
                        .font(.custom("SpaceGrotesk-Bold", size: 28))
                        .foregroundColor(.textPrimary)
                    Text("登录您的账户")
                        .font(.custom("Inter-Regular", size: 14))
                        .foregroundColor(.textSecondary)
                }
                .padding(.bottom, 48)

                VStack(spacing: 16) {
                    TextField("手机号", text: $viewModel.phoneNumber)
                        .keyboardType(.numberPad)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    Picker("登录方式", selection: $viewModel.loginMode) {
                        Text("密码登录").tag(LoginViewModel.LoginMode.password)
                        Text("验证码登录").tag(LoginViewModel.LoginMode.verificationCode)
                    }
                    .pickerStyle(.segmented)

                    if viewModel.loginMode == .password {
                        SecureField("密码", text: $viewModel.password)
                            .font(.custom("Inter-Regular", size: 15))
                            .glowingInput()
                    } else {
                        HStack(spacing: 8) {
                            TextField("验证码", text: $viewModel.verificationCode)
                                .keyboardType(.numberPad)
                                .font(.custom("Inter-Regular", size: 15))
                                .glowingInput()

                            Button(viewModel.countdownSeconds > 0 ? "\(viewModel.countdownSeconds)s" : "发送验证码") {
                                Task { await viewModel.sendVerificationCode() }
                            }
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.brandSecondary)
                            .disabled(viewModel.countdownSeconds > 0 || viewModel.phoneNumber.isEmpty)
                        }
                    }

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage)
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.danger)
                    }

                    Button(action: { Task { await viewModel.login() } }) {
                        if viewModel.isLoading {
                            ProgressView()
                        } else {
                            Text("登录")
                                .font(.custom("Inter-Semibold", size: 16))
                        }
                    }
                    .gradientButton()
                    .disabled(viewModel.isLoading)
                }
                .padding(.horizontal, 32)

                HStack(spacing: 4) {
                    Text("还没有账号？")
                        .font(.custom("Inter-Regular", size: 13))
                        .foregroundColor(.textSecondary)
                    NavigationLink(destination: RegisterView()) {
                        Text("立即注册")
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.brandSecondary)
                    }
                }
                .padding(.top, 24)

                Spacer()
            }
        }
        .navigationBarHidden(true)
    }
}
```

- [ ] **Step 2: Rewrite RegisterView.swift**

```swift
import SwiftUI

struct RegisterView: View {
    @StateObject private var viewModel = RegisterViewModel()
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 16) {
                    Spacer().frame(height: 24)

                    VStack(spacing: 8) {
                        Text("创建账户")
                            .font(.custom("SpaceGrotesk-Bold", size: 28))
                            .foregroundColor(.textPrimary)
                        Text("注册新账户以使用完整功能")
                            .font(.custom("Inter-Regular", size: 14))
                            .foregroundColor(.textSecondary)
                    }
                    .padding(.bottom, 32)

                    TextField("手机号", text: $viewModel.phoneNumber)
                        .keyboardType(.numberPad)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    HStack(spacing: 8) {
                        TextField("验证码", text: $viewModel.verificationCode)
                            .keyboardType(.numberPad)
                            .font(.custom("Inter-Regular", size: 15))
                            .glowingInput()

                        Button(viewModel.countdownSeconds > 0 ? "\(viewModel.countdownSeconds)s" : "发送验证码") {
                            Task { await viewModel.sendVerificationCode() }
                        }
                        .font(.custom("Inter-Regular", size: 13))
                        .foregroundColor(.brandSecondary)
                        .disabled(viewModel.countdownSeconds > 0 || viewModel.phoneNumber.isEmpty)
                    }

                    TextField("账户ID", text: $viewModel.accountId)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    SecureField("密码", text: $viewModel.password)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    SecureField("确认密码", text: $viewModel.confirmPassword)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    TextField("推荐码（可选）", text: $viewModel.referralCode)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    Toggle(isOn: $viewModel.agreeToTerms) {
                        Text("我已阅读并同意服务条款和隐私政策")
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.textSecondary)
                    }
                    .tint(.brandPrimary)

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage)
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.danger)
                    }

                    Button(action: { Task { await viewModel.register() } }) {
                        if viewModel.isLoading {
                            ProgressView()
                        } else {
                            Text("注册")
                                .font(.custom("Inter-Semibold", size: 16))
                        }
                    }
                    .gradientButton()
                    .disabled(viewModel.isLoading)

                    HStack(spacing: 4) {
                        Text("已有账号？")
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                        Button(action: { dismiss() }) {
                            Text("立即登录")
                                .font(.custom("Inter-Semibold", size: 13))
                                .foregroundColor(.brandSecondary)
                        }
                    }
                    .padding(.top, 16)
                }
                .padding(.horizontal, 32)
            }
        }
        .navigationBarTitleDisplayMode(.inline)
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add ios/AccountCenter/Features/Login/ ios/AccountCenter/Features/Register/
git commit -m "feat(ui): iOS Login + Register dark theme"
```

---

### Task 3: iOS Home Dark Theme

**Files:**
- Modify: `ios/AccountCenter/Features/Home/HomeView.swift`

- [ ] **Step 1: Rewrite HomeView.swift with card-based layout**

```swift
import SwiftUI

struct HomeView: View {
    @StateObject private var viewModel = HomeViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    // User Card
                    HStack(spacing: 16) {
                        ZStack {
                            Circle()
                                .fill(Color.brandGradient)
                                .frame(width: 64, height: 64)
                            Text(viewModel.currentUser?.accountId.prefix(1).uppercased() ?? "U")
                                .font(.custom("SpaceGrotesk-Bold", size: 24))
                                .foregroundColor(.white)
                        }
                        VStack(alignment: .leading, spacing: 4) {
                            Text(viewModel.currentUser?.accountId ?? "用户")
                                .font(.custom("SpaceGrotesk-Bold", size: 18))
                                .foregroundColor(.textPrimary)
                            Text("未绑定手机号")
                                .font(.custom("Inter-Regular", size: 13))
                                .foregroundColor(.textSecondary)
                            HStack(spacing: 4) {
                                Circle()
                                    .fill(Color.brandSecondary)
                                    .frame(width: 6, height: 6)
                                Text("Lv.2")
                                    .font(.custom("SpaceGrotesk-Semibold", size: 11))
                                    .foregroundColor(.brandSecondary)
                            }
                        }
                        Spacer()
                    }
                    .cardStyle()

                    // RFM Card
                    if let rfm = viewModel.rfmScore {
                        HStack(spacing: 12) {
                            Text("\u{1F3AF}")
                                .font(.title2)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(rfm.rfmSegmentCn)
                                    .font(.custom("SpaceGrotesk-Semibold", size: 14))
                                    .foregroundColor(.textPrimary)
                                Text("RFM \(rfm.rfmSegment)")
                                    .font(.custom("Inter-Regular", size: 12))
                                    .foregroundColor(.textSecondary)
                            }
                            Spacer()
                        }
                        .cardStyle()
                        .overlay(
                            RoundedRectangle(cornerRadius: 16)
                                .stroke(Color.brandGradient, lineWidth: 1)
                                .opacity(0.3)
                        )
                    }

                    // Features
                    VStack(alignment: .leading, spacing: 0) {
                        Text("功能")
                            .sectionTitle()
                            .padding(.horizontal, 4)
                            .padding(.bottom, 8)

                        VStack(spacing: 0) {
                            FeatureRow(icon: "cart", label: "订阅管理", destination: SubscriptionView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "creditcard", label: "积分中心", destination: CreditsView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "lock", label: "安全设置", destination: SecurityView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "info.circle", label: "关于", destination: AboutView())
                        }
                        .background(Color.bgCard)
                        .cornerRadius(16)
                    }

                    // Logout
                    Button(action: { Task { await viewModel.logout() } }) {
                        Text("退出登录")
                            .font(.custom("Inter-Regular", size: 15))
                            .foregroundColor(.danger)
                            .frame(maxWidth: .infinity)
                            .frame(height: 52)
                            .background(Color.bgCard)
                            .cornerRadius(12)
                    }
                    .padding(.top, 8)
                }
                .padding(16)
            }
        }
        .navigationTitle("用户中心")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .task { await viewModel.loadRFM() }
    }
}

private struct FeatureRow<Destination: View>: View {
    let icon: String
    let label: String
    let destination: Destination

    var body: some View {
        NavigationLink(destination: destination) {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.system(size: 18))
                    .foregroundColor(.brandPrimary)
                    .frame(width: 24)
                Text(label)
                    .font(.custom("Inter-Regular", size: 15))
                    .foregroundColor(.textPrimary)
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.textSecondary)
            }
            .frame(height: 56)
            .padding(.horizontal, 16)
        }
    }
}
```

- [ ] **Step 2: Commit**

```bash
git add ios/AccountCenter/Features/Home/HomeView.swift
git commit -m "feat(ui): iOS Home dark theme card-based layout"
```

---

### Task 4: iOS Feature Pages Dark Theme (Sub/Credits/Security/About)

**Files:**
- Modify: `ios/AccountCenter/Features/Subscription/SubscriptionView.swift`
- Modify: `ios/AccountCenter/Features/Credits/CreditsView.swift`
- Modify: `ios/AccountCenter/Features/Security/SecurityView.swift`
- Modify: `ios/AccountCenter/Features/About/AboutView.swift`

- [ ] **Step 1: Rewrite SubscriptionView.swift**

```swift
import SwiftUI

struct SubscriptionView: View {
    @StateObject private var viewModel = SubscriptionViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    if viewModel.isLoading && viewModel.subscriptions.isEmpty {
                        ProgressView().padding()
                    }

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage)
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.danger)
                    }

                    // Tier badge
                    if let tier = viewModel.currentTier {
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                Circle()
                                    .fill(viewModel.tierColor(for: tier.identityTier))
                                    .frame(width: 12, height: 12)
                                Text(viewModel.tierName(for: tier.identityTier))
                                    .font(.custom("SpaceGrotesk-Semibold", size: 16))
                                    .foregroundColor(.textPrimary)
                                Spacer()
                                Text("Lv.\(tier.identityTier)")
                                    .font(.custom("SpaceGrotesk-Semibold", size: 12))
                                    .foregroundColor(.brandSecondary)
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 4)
                                    .background(Color.brandSecondary.opacity(0.15))
                                    .cornerRadius(6)
                            }
                        }
                        .cardStyle()
                    }

                    // Active subscription
                    VStack(alignment: .leading, spacing: 8) {
                        Text("当前订阅")
                            .sectionTitle()

                        if let sub = viewModel.activeSubscription {
                            VStack(alignment: .leading, spacing: 12) {
                                HStack(spacing: 6) {
                                    Circle().fill(Color.success).frame(width: 8, height: 8)
                                    Text("状态").font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
                                    Spacer()
                                    Text("生效中").font(.custom("Inter-Semibold", size: 14)).foregroundColor(.success)
                                }
                                Divider().background(Color.divider)
                                labeledRow("开始时间", sub.startTime)
                                Divider().background(Color.divider)
                                labeledRow("结束时间", sub.endTime)
                                Divider().background(Color.divider)
                                labeledRow("金额", "\u{00A5}\(String(format: "%.2f", sub.price))")
                                if let method = sub.paymentMethod {
                                    Divider().background(Color.divider)
                                    labeledRow("支付方式", method)
                                }
                            }
                            .cardStyle()
                            .overlay(
                                RoundedRectangle(cornerRadius: 16)
                                    .fill(Color.brandPrimary.opacity(0))
                                    .overlay(
                                        Rectangle().fill(Color.brandPrimary).frame(width: 3),
                                        alignment: .leading
                                    )
                                    .clipShape(RoundedRectangle(cornerRadius: 16))
                            )
                        } else {
                            Text("暂无活跃订阅")
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.textSecondary)
                                .cardStyle()
                                .frame(maxWidth: .infinity)
                        }
                    }

                    // History
                    VStack(alignment: .leading, spacing: 0) {
                        Text("订阅历史")
                            .sectionTitle()
                            .padding(.bottom, 8)

                        VStack(spacing: 0) {
                            ForEach(viewModel.subscriptions.filter { $0.status != "active" }) { sub in
                                HStack {
                                    VStack(alignment: .leading, spacing: 4) {
                                        Text("\(sub.startTime) - \(sub.endTime)")
                                            .font(.custom("Inter-Regular", size: 13))
                                            .foregroundColor(.textSecondary)
                                    }
                                    Spacer()
                                    Text("\u{00A5}\(String(format: "%.2f", sub.price))")
                                        .font(.custom("Inter-Semibold", size: 14))
                                        .foregroundColor(.textPrimary)
                                }
                                .padding(.horizontal, 16)
                                .frame(height: 48)
                                if sub.id != viewModel.subscriptions.last?.id {
                                    Divider().background(Color.divider).padding(.leading, 16)
                                }
                            }
                        }
                        .background(Color.bgCard)
                        .cornerRadius(16)
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle("订阅管理")
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
    }

    private func labeledRow(_ label: String, _ value: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
            Spacer()
            Text(value).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textPrimary)
        }
    }
}
```

- [ ] **Step 2: Rewrite CreditsView.swift**

```swift
import SwiftUI

struct CreditsView: View {
    @StateObject private var viewModel = CreditsViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    if viewModel.isLoading {
                        ProgressView().padding()
                    }

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage).font(.custom("Inter-Regular", size: 12)).foregroundColor(.danger)
                    }

                    if let account = viewModel.account {
                        VStack(spacing: 4) {
                            Text("当前积分")
                                .font(.custom("Inter-Regular", size: 13))
                                .foregroundColor(.textSecondary)
                            Text("\u{00A5}\(String(format: "%.2f", account.balance))")
                                .font(.custom("SpaceGrotesk-Bold", size: 36))
                                .foregroundStyle(Color.brandGradient)
                            Text(account.status == "active" ? "账户正常" : account.status)
                                .font(.custom("Inter-Regular", size: 12))
                                .foregroundColor(.success)
                        }
                        .cardStyle()
                        .frame(maxWidth: .infinity)
                    }

                    if let referral = viewModel.referral {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("邀请推广").font(.custom("Inter-Semibold", size: 15)).foregroundColor(.textPrimary)
                            labeledRow("邀请人数", "\(referral.totalReferees)")
                            labeledRow("活跃好友", "\(referral.activeReferees)")
                            labeledRow("累计收益", "\u{00A5}\(String(format: "%.2f", referral.totalEarned))")
                            Button(action: { Task { await viewModel.generateLink() } }) {
                                Label("复制推荐链接", systemImage: "link")
                                    .font(.custom("Inter-Regular", size: 14))
                                    .foregroundColor(.brandPrimary)
                            }
                            .buttonStyle(.bordered)
                            .tint(.brandPrimary)
                        }
                        .cardStyle()
                    }

                    VStack(alignment: .leading, spacing: 0) {
                        Text("交易记录")
                            .sectionTitle()
                            .padding(.bottom, 8)

                        if viewModel.transactions.isEmpty {
                            Text("暂无交易记录")
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.textSecondary)
                                .padding(24)
                                .frame(maxWidth: .infinity)
                                .background(Color.bgCard)
                                .cornerRadius(16)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(viewModel.transactions) { txn in
                                    HStack(spacing: 12) {
                                        Image(systemName: viewModel.transactionTypeIcon(txn.type))
                                            .font(.system(size: 16))
                                            .foregroundColor(viewModel.transactionTypeColor(txn.type))
                                            .frame(width: 24)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(txn.details ?? txn.type)
                                                .font(.custom("Inter-Regular", size: 14))
                                                .foregroundColor(.textPrimary)
                                            Text(txn.createdAt)
                                                .font(.custom("Inter-Regular", size: 11))
                                                .foregroundColor(.textSecondary)
                                        }
                                        Spacer()
                                        Text("\(txn.amount >= 0 ? "+" : "")\u{00A5}\(String(format: "%.2f", txn.amount))")
                                            .font(.custom("Inter-Semibold", size: 14))
                                            .foregroundColor(viewModel.transactionTypeColor(txn.type))
                                    }
                                    .padding(.horizontal, 16)
                                    .frame(minHeight: 52)
                                    if txn.id != viewModel.transactions.last?.id {
                                        Divider().background(Color.divider).padding(.leading, 52)
                                    }
                                }
                            }
                            .background(Color.bgCard)
                            .cornerRadius(16)
                        }

                        if viewModel.hasMore {
                            Button("加载更多...") { Task { await viewModel.loadMore() } }
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.brandSecondary)
                                .padding(.top, 8)
                        }
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle("积分中心")
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
    }

    private func labeledRow(_ label: String, _ value: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
            Spacer()
            Text(value).font(.custom("Inter-Semibold", size: 14)).foregroundColor(.textPrimary)
        }
    }
}
```

- [ ] **Step 3: Rewrite SecurityView.swift**

```swift
import SwiftUI

struct SecurityView: View {
    @StateObject private var viewModel = SecurityViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    if viewModel.isLoading {
                        ProgressView().padding()
                    }
                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage).font(.custom("Inter-Regular", size: 12)).foregroundColor(.danger)
                    }

                    VStack(alignment: .leading, spacing: 0) {
                        Text("风险事件").sectionTitle().padding(.bottom, 8)

                        if viewModel.riskEvents.isEmpty {
                            Text("暂无风险事件")
                                .font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
                                .padding(24).frame(maxWidth: .infinity).background(Color.bgCard).cornerRadius(16)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(viewModel.riskEvents) { event in
                                    HStack(spacing: 12) {
                                        Circle().fill(viewModel.riskLevelColor(event.riskLevel)).frame(width: 10, height: 10)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(event.eventType).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textPrimary)
                                            Text(event.createdAt).font(.custom("Inter-Regular", size: 11)).foregroundColor(.textSecondary)
                                        }
                                        Spacer()
                                        Text(event.riskLevel.uppercased())
                                            .font(.custom("Inter-Semibold", size: 10))
                                            .foregroundColor(viewModel.riskLevelColor(event.riskLevel))
                                            .padding(.horizontal, 6).padding(.vertical, 3)
                                            .background(viewModel.riskLevelColor(event.riskLevel).opacity(0.15))
                                            .cornerRadius(4)
                                    }
                                    .padding(.horizontal, 16).frame(minHeight: 52)
                                    if event.id != viewModel.riskEvents.last?.id {
                                        Divider().background(Color.divider).padding(.leading, 38)
                                    }
                                }
                            }
                            .background(Color.bgCard).cornerRadius(16)
                        }
                    }

                    VStack(alignment: .leading, spacing: 0) {
                        Text("登录设备").sectionTitle().padding(.bottom, 8)

                        if viewModel.devices.isEmpty {
                            Text("暂无设备记录")
                                .font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
                                .padding(24).frame(maxWidth: .infinity).background(Color.bgCard).cornerRadius(16)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(viewModel.devices) { device in
                                    HStack(spacing: 12) {
                                        Image(systemName: device.platform == "ios" ? "iphone" : "desktopcomputer")
                                            .font(.system(size: 18)).foregroundColor(.brandPrimary).frame(width: 24)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(device.deviceName ?? device.platform)
                                                .font(.custom("Inter-Regular", size: 14)).foregroundColor(.textPrimary)
                                            if let lastActive = device.lastActiveAt {
                                                Text("最近活跃: \(lastActive)")
                                                    .font(.custom("Inter-Regular", size: 11)).foregroundColor(.textSecondary)
                                            }
                                        }
                                        Spacer()
                                        Text(device.isActive ? "活跃中" : "离线")
                                            .font(.custom("Inter-Semibold", size: 11))
                                            .foregroundColor(device.isActive ? .success : .textSecondary)
                                    }
                                    .padding(.horizontal, 16).frame(minHeight: 52)
                                    if device.id != viewModel.devices.last?.id {
                                        Divider().background(Color.divider).padding(.leading, 52)
                                    }
                                }
                            }
                            .background(Color.bgCard).cornerRadius(16)
                        }
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle("安全设置")
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
    }
}
```

- [ ] **Step 4: Rewrite AboutView.swift**

```swift
import SwiftUI

struct AboutView: View {
    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer().frame(height: 48)

                VStack(spacing: 12) {
                    Image(systemName: "person.circle.fill")
                        .font(.system(size: 56))
                        .foregroundStyle(Color.brandGradient)
                    Text("账户中心")
                        .font(.custom("SpaceGrotesk-Bold", size: 22))
                        .foregroundColor(.textPrimary)
                    HStack(spacing: 4) {
                        Text("Version")
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                        Text(Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0")
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.textPrimary)
                        Text("Build")
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                            .padding(.leading, 8)
                        Text(Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "1")
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.textPrimary)
                    }
                }
                .padding(.bottom, 48)

                VStack(spacing: 0) {
                    aboutRow("版本号", Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0")
                    Divider().background(Color.divider).padding(.leading, 16)
                    aboutRow("构建号", Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "1")
                }
                .background(Color.bgCard).cornerRadius(16)
                .padding(.horizontal, 16)

                VStack(spacing: 0) {
                    legalRow("服务条款")
                    Divider().background(Color.divider).padding(.leading, 16)
                    legalRow("隐私政策")
                }
                .background(Color.bgCard).cornerRadius(16)
                .padding(.horizontal, 16)
                .padding(.top, 12)

                Spacer()

                Text("\u{00A9} 2026 Account Center. All rights reserved.")
                    .font(.custom("Inter-Regular", size: 11))
                    .foregroundColor(.textSecondary)
                    .padding(.bottom, 24)
            }
        }
        .navigationTitle("关于")
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
    }

    private func aboutRow(_ label: String, _ value: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
            Spacer()
            Text(value).font(.custom("Inter-Semibold", size: 14)).foregroundColor(.textPrimary)
        }
        .padding(.horizontal, 16).frame(height: 48)
    }

    private func legalRow(_ label: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.brandSecondary)
            Spacer()
            Image(systemName: "chevron.right").font(.system(size: 12)).foregroundColor(.textSecondary)
        }
        .padding(.horizontal, 16).frame(height: 48)
    }
}
```

- [ ] **Step 5: Commit**

```bash
git add ios/AccountCenter/Features/Subscription/ ios/AccountCenter/Features/Credits/ ios/AccountCenter/Features/Security/ ios/AccountCenter/Features/About/
git commit -m "feat(ui): iOS feature pages dark theme"
```

---

### Task 5: Android Theme + Design Tokens

**Files:**
- Modify: `android/app/src/main/java/com/accountcenter/ui/theme/Color.kt`
- Modify: `android/app/src/main/java/com/accountcenter/ui/theme/Theme.kt`
- Modify: `android/app/src/main/java/com/accountcenter/ui/theme/Type.kt`
- Create: `android/app/src/main/java/com/accountcenter/ui/theme/Shape.kt`
- Create: `android/app/src/main/java/com/accountcenter/ui/components/AppCard.kt`
- Create: `android/app/src/main/java/com/accountcenter/ui/components/GradientButton.kt`

- [ ] **Step 1: Rewrite Color.kt**

```kotlin
package com.accountcenter.ui.theme

import androidx.compose.ui.graphics.Color

val BgPrimary = Color(0xFF0D1117)
val BgCard = Color(0xFF161B22)
val BgInput = Color(0xFF161B22)
val BrandPrimary = Color(0xFF6C63FF)
val BrandSecondary = Color(0xFF00D4FF)
val TextPrimary = Color(0xDEFFFFFF) // 87%
val TextSecondary = Color(0xFF8B949E)
val Divider = Color(0xFF21262D)
val Danger = Color(0xFFFF4757)
val Success = Color(0xFF2ED573)
val TierFree = Color(0xFF8B949E)
val TierBasic = Color(0xFF6C63FF)
val TierPremium = Color(0xFFFF9800)
val TierEnterprise = Color(0xFF7B1FA2)
```

- [ ] **Step 2: Create Shape.kt**

```kotlin
package com.accountcenter.ui.theme

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Shapes
import androidx.compose.ui.unit.dp

val AppShapes = Shapes(
    small = RoundedCornerShape(8.dp),
    medium = RoundedCornerShape(12.dp),
    large = RoundedCornerShape(16.dp)
)
```

- [ ] **Step 3: Update Theme.kt** — replace `DarkColorScheme` with new tokens mapping `primary = BrandPrimary`, `secondary = BrandSecondary`, `background = BgPrimary`, `surface = BgCard`, `onPrimary = White`, `onBackground = TextPrimary`, `onSurface = TextPrimary` etc. Add `shapes = AppShapes`. Disable dynamic color (the whole point is custom palette).

- [ ] **Step 4: Update Type.kt** — reference Inter and Space Grotesk font families via `FontFamily`. Add Space Grotesk for headline styles.

```kotlin
package com.accountcenter.ui.theme

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp
import com.accountcenter.R

val Inter = FontFamily(
    Font(R.font.inter_regular, FontWeight.Normal),
    Font(R.font.inter_semibold, FontWeight.SemiBold),
    Font(R.font.inter_bold, FontWeight.Bold)
)

val SpaceGrotesk = FontFamily(
    Font(R.font.spacegrotesk_bold, FontWeight.Bold),
    Font(R.font.spacegrotesk_semibold, FontWeight.SemiBold)
)

val Typography = Typography(
    displayLarge = TextStyle(fontFamily = SpaceGrotesk, fontWeight = FontWeight.Bold, fontSize = 28.sp),
    headlineLarge = TextStyle(fontFamily = SpaceGrotesk, fontWeight = FontWeight.Bold, fontSize = 24.sp),
    headlineMedium = TextStyle(fontFamily = SpaceGrotesk, fontWeight = FontWeight.Bold, fontSize = 20.sp),
    titleLarge = TextStyle(fontFamily = SpaceGrotesk, fontWeight = FontWeight.SemiBold, fontSize = 18.sp),
    titleMedium = TextStyle(fontFamily = Inter, fontWeight = FontWeight.SemiBold, fontSize = 15.sp),
    bodyLarge = TextStyle(fontFamily = Inter, fontWeight = FontWeight.Normal, fontSize = 15.sp),
    bodyMedium = TextStyle(fontFamily = Inter, fontWeight = FontWeight.Normal, fontSize = 14.sp),
    bodySmall = TextStyle(fontFamily = Inter, fontWeight = FontWeight.Normal, fontSize = 12.sp),
    labelSmall = TextStyle(fontFamily = Inter, fontWeight = FontWeight.Medium, fontSize = 11.sp)
)
```

- [ ] **Step 5: Create AppCard.kt**

```kotlin
package com.accountcenter.ui.components

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.accountcenter.ui.theme.BgCard

@Composable
fun AppCard(
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(containerColor = BgCard),
        shape = androidx.compose.foundation.shape.RoundedCornerShape(16.dp)
    ) {
        androidx.compose.foundation.layout.Box(modifier = Modifier.padding(16.dp)) {
            content()
        }
    }
}
```

- [ ] **Step 6: Create GradientButton.kt**

```kotlin
package com.accountcenter.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.accountcenter.ui.theme.BrandPrimary
import com.accountcenter.ui.theme.BrandSecondary

@Composable
fun GradientButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    val gradient = Brush.horizontalGradient(listOf(BrandPrimary, BrandSecondary))
    Button(
        onClick = onClick,
        modifier = modifier
            .fillMaxWidth()
            .height(52.dp)
            .clip(RoundedCornerShape(12.dp)),
        colors = ButtonDefaults.buttonColors(containerColor = Color.Transparent),
        enabled = enabled
    ) {
        Text(text, color = Color.White)
    }
}
```

- [ ] **Step 7: Commit**

```bash
git add android/app/src/main/java/com/accountcenter/ui/theme/Color.kt android/app/src/main/java/com/accountcenter/ui/theme/Theme.kt android/app/src/main/java/com/accountcenter/ui/theme/Type.kt android/app/src/main/java/com/accountcenter/ui/theme/Shape.kt android/app/src/main/java/com/accountcenter/ui/components/
git commit -m "feat(ui): Android design tokens and reusable components"
```

---

### Task 6: Android Login + Register Dark Theme

**Files:**
- Modify: `android/app/src/main/java/com/accountcenter/ui/login/LoginScreen.kt`
- Modify: `android/app/src/main/java/com/accountcenter/ui/register/RegisterScreen.kt`

- [ ] **Step 1: Rewrite LoginScreen.kt**

```kotlin
package com.accountcenter.ui.login

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.accountcenter.ui.components.GradientButton
import com.accountcenter.ui.theme.*

@Composable
fun LoginScreen(
    viewModel: LoginViewModel = hiltViewModel(),
    onNavigateToRegister: () -> Unit,
    onLoginSuccess: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()

    Box(
        modifier = Modifier.fillMaxSize().background(BgPrimary)
    ) {
        Column(
            modifier = Modifier.fillMaxSize().padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = Icons.Default.Person,
                contentDescription = null,
                modifier = Modifier.size(56.dp),
                tint = Brush.horizontalGradient(listOf(BrandPrimary, BrandSecondary))
            )
            Spacer().height(12.dp)
            Text(
                "账户中心",
                style = MaterialTheme.typography.headlineLarge,
                color = TextPrimary
            )
            Text(
                "登录您的账户",
                style = MaterialTheme.typography.bodyMedium,
                color = TextSecondary
            )
            Spacer().height(48.dp)

            OutlinedTextField(
                value = uiState.phoneNumber,
                onValueChange = viewModel::onPhoneNumberChange,
                label = { Text("手机号") },
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = BrandSecondary,
                    unfocusedBorderColor = Divider,
                    focusedLabelColor = BrandSecondary,
                    unfocusedLabelColor = TextSecondary
                )
            )
            Spacer().height(16.dp)

            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                SegmentedButton(
                    selected = uiState.loginMode == LoginMode.PASSWORD,
                    onClick = { viewModel.onLoginModeChange(LoginMode.PASSWORD) },
                    shape = SegmentedButtonDefaults.itemShape(index = 0, count = 2)
                ) { Text("密码登录") }
                SegmentedButton(
                    selected = uiState.loginMode == LoginMode.VERIFICATION_CODE,
                    onClick = { viewModel.onLoginModeChange(LoginMode.VERIFICATION_CODE) },
                    shape = SegmentedButtonDefaults.itemShape(index = 1, count = 2)
                ) { Text("验证码登录") }
            }

            Spacer().height(16.dp)

            if (uiState.loginMode == LoginMode.PASSWORD) {
                OutlinedTextField(
                    value = uiState.password,
                    onValueChange = viewModel::onPasswordChange,
                    label = { Text("密码") },
                    singleLine = true,
                    visualTransformation = PasswordVisualTransformation(),
                    modifier = Modifier.fillMaxWidth(),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = BrandSecondary,
                        unfocusedBorderColor = Divider,
                        focusedLabelColor = BrandSecondary,
                        unfocusedLabelColor = TextSecondary
                    )
                )
            } else {
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(
                        value = uiState.verificationCode,
                        onValueChange = viewModel::onVerificationCodeChange,
                        label = { Text("验证码") },
                        singleLine = true,
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                        modifier = Modifier.weight(1f),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = BrandSecondary,
                            unfocusedBorderColor = Divider,
                            focusedLabelColor = BrandSecondary,
                            unfocusedLabelColor = TextSecondary
                        )
                    )
                    val btnText = if (uiState.countdownSeconds > 0) "${uiState.countdownSeconds}s" else "发送验证码"
                    OutlinedButton(
                        onClick = viewModel::sendVerificationCode,
                        enabled = !uiState.isLoading && uiState.countdownSeconds == 0 && uiState.phoneNumber.isNotEmpty()
                    ) {
                        Text(btnText, color = BrandSecondary)
                    }
                }
            }

            uiState.errorMessage?.let {
                Text(it, color = Danger, style = MaterialTheme.typography.bodySmall)
            }
            Spacer().height(16.dp)

            GradientButton(
                text = "登录",
                onClick = { viewModel.login(onLoginSuccess) },
                enabled = !uiState.isLoading
            )

            Spacer().height(24.dp)
            Row(horizontalArrangement = Arrangement.Center, modifier = Modifier.fillMaxWidth()) {
                Text("还没有账号？", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                TextButton(onClick = onNavigateToRegister) {
                    Text("立即注册", color = BrandSecondary)
                }
            }
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(
                modifier = Modifier.align(Alignment.Center),
                color = BrandSecondary
            )
        }
    }
}
```

- [ ] **Step 2: Rewrite RegisterScreen.kt** — same dark theme pattern: `Box(modifier=Modifier.fillMaxSize().background(BgPrimary))` root, `Icon` with brandGradient tint, Space Grotesk headline, all OutlinedTextField with custom colors (focused=BrandSecondary, unfocused=Divider), GradientButton for register, error text in Danger, "已有账号？" in TextSecondary / BrandSecondary.

- [ ] **Step 3: Commit**

```bash
git add android/app/src/main/java/com/accountcenter/ui/login/ android/app/src/main/java/com/accountcenter/ui/register/
git commit -m "feat(ui): Android Login + Register dark theme"
```

---

### Task 7: Android Home Dark Theme

**Files:**
- Modify: `android/app/src/main/java/com/accountcenter/ui/home/HomeScreen.kt`

- [ ] **Step 1: Rewrite HomeScreen.kt**

```kotlin
package com.accountcenter.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.accountcenter.ui.components.AppCard
import com.accountcenter.ui.theme.*

@Composable
fun HomeScreen(
    homeViewModel: HomeViewModel = hiltViewModel(),
    onLogout: () -> Unit,
    onNavigateToSubscription: () -> Unit = {},
    onNavigateToCredits: () -> Unit = {},
    onNavigateToSecurity: () -> Unit = {},
    onNavigateToAbout: () -> Unit = {}
) {
    val userDisplay by homeViewModel.userDisplay.collectAsStateWithLifecycle()
    val rfmScore by homeViewModel.rfmScore.collectAsStateWithLifecycle()

    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.Top
        ) {
            CenterAlignedTopAppBar(
                title = { Text("用户中心", color = TextPrimary) },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = BgPrimary.copy(alpha = 0.9f)
                )
            )

            Column(
                modifier = Modifier.fillMaxSize().padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                // User card
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .size(64.dp)
                            .clip(CircleShape)
                            .background(
                                Brush.horizontalGradient(listOf(BrandPrimary, BrandSecondary))
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            userDisplay.accountId.firstOrNull()?.uppercase() ?: "U",
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                    Spacer().width(16.dp)
                    Column {
                        Text(
                            userDisplay.accountId.ifEmpty { "用户" },
                            style = MaterialTheme.typography.titleMedium,
                            color = TextPrimary
                        )
                        Text(
                            userDisplay.phoneNumber.ifEmpty { "未绑定手机号" },
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary
                        )
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier.size(6.dp).clip(CircleShape).background(BrandSecondary)
                            )
                            Spacer().width(4.dp)
                            Text("Lv.2", style = MaterialTheme.typography.labelSmall, color = BrandSecondary)
                        }
                    }
                }
                .let { AppCard { it } }

                // RFM card
                rfmScore?.let { score ->
                    AppCard(
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text("\uD83C\uDFAF", style = MaterialTheme.typography.titleMedium)
                            Spacer().width(12.dp)
                            Column {
                                Text(score.rfmSegmentCn, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold, color = TextPrimary)
                                Text("RFM ${score.rfmSegment}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                            }
                        }
                    }
                }

                // Feature list
                Text("功能", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = BgCard),
                    shape = RoundedCornerShape(16.dp)
                ) {
                    Column {
                        FeatureRow(Icons.Default.ShoppingCart, "订阅管理", onNavigateToSubscription)
                        HorizontalDivider(color = Divider)
                        FeatureRow(Icons.Default.CreditCard, "积分中心", onNavigateToCredits)
                        HorizontalDivider(color = Divider)
                        FeatureRow(Icons.Default.Lock, "安全设置", onNavigateToSecurity)
                        HorizontalDivider(color = Divider)
                        FeatureRow(Icons.Default.Info, "关于", onNavigateToAbout)
                    }
                }

                // Logout
                Button(
                    onClick = { homeViewModel.logout(onLogout) },
                    modifier = Modifier.fillMaxWidth().height(52.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = BgCard),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("退出登录", color = Danger)
                }
            }
        }
    }
}

@Composable
private fun FeatureRow(icon: ImageVector, label: String, onClick: () -> Unit) {
    Surface(
        onClick = onClick,
        color = Color.Transparent,
        modifier = Modifier.fillMaxWidth().height(56.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(icon, contentDescription = label, tint = BrandPrimary, modifier = Modifier.size(24.dp))
            Spacer().width(12.dp)
            Text(label, style = MaterialTheme.typography.bodyLarge, color = TextPrimary, modifier = Modifier.weight(1f))
            Icon(Icons.Default.ChevronRight, contentDescription = null, tint = TextSecondary, modifier = Modifier.size(16.dp))
        }
    }
}
```

- [ ] **Step 2: Commit**

```bash
git add android/app/src/main/java/com/accountcenter/ui/home/HomeScreen.kt
git commit -m "feat(ui): Android Home dark theme card-based layout"
```

---

### Task 8: Android Feature Pages Dark Theme

**Files:**
- Modify: `android/app/src/main/java/com/accountcenter/ui/subscription/SubscriptionScreen.kt`
- Modify: `android/app/src/main/java/com/accountcenter/ui/credits/CreditsScreen.kt`
- Modify: `android/app/src/main/java/com/accountcenter/ui/security/SecurityScreen.kt`
- Modify: `android/app/src/main/java/com/accountcenter/ui/about/AboutScreen.kt`

- [ ] **Step 1: Rewrite SubscriptionScreen.kt** — apply dark theme: `Box(modifier=Modifier.fillMaxSize().background(BgPrimary))`, `CenterAlignedTopAppBar` with BgPrimary background. Tier badge: `AppCard` with colored `Box` dot + "Lv.N" pill. Active subscription: `AppCard` with `BrandPrimary` left border overlay, labeled rows. History: `Card(BgCard, RoundedCornerShape(16.dp))` with horizontal dividers. Empty state: "暂无活跃订阅" in TextSecondary.

- [ ] **Step 2: Rewrite CreditsScreen.kt** — apply dark theme: BgPrimary root, CenterAlignedTopAppBar. Balance: AppCard, centered, balance text with `Brush.horizontalGradient(BrandPrimary, BrandSecondary)`, 36sp Space Grotesk weight. Referral: AppCard with labeled rows + `GradientButton("复制推荐链接")`. Transactions: Card(BgCard) list with earn (green=Success) / consume (red=Danger) coloring, horizontal dividers. `hasMore` → TextButton "加载更多..." in BrandSecondary.

- [ ] **Step 3: Rewrite SecurityScreen.kt** — apply dark theme: BgPrimary root, CenterAlignedTopAppBar. Risk events: Card(BgCard) list, each row with colored `Box(12.dp, CircleShape)` dot (red=Danger `0xFFF44336`, orange= `0xFFFF9800`, yellow=`0xFFFFEB3B`, green=Success), event type + timestamp, right risk level badge (colored bg 15% opacity). Devices: Card(BgCard) list, each row with `Icons.Default.PhoneAndroid` / `Icons.Default.Computer` platform icon, device name + "最近活跃", status pill ("活跃中" green / "离线" gray). Read-only.

- [ ] **Step 4: Rewrite AboutScreen.kt** — apply dark theme: BgPrimary root, CenterAlignedTopAppBar. Centered logo section: `Icon(Person, 56dp, brush=horizontalGradient(BrandPrimary,BrandSecondary))` + "账户中心" headlineLarge TextPrimary + "Version X Build Y". Info: Card(BgCard, 16dp) with labeledRow for 版本号、构建号. Legal: Card(BgCard, 16dp) with "服务条款" / "隐私政策" in BrandSecondary + chevron. Footer: "(C) 2026 Account Center" in TextSecondary 11sp.

- [ ] **Step 5: Commit**

```bash
git add android/app/src/main/java/com/accountcenter/ui/subscription/ android/app/src/main/java/com/accountcenter/ui/credits/ android/app/src/main/java/com/accountcenter/ui/security/ android/app/src/main/java/com/accountcenter/ui/about/
git commit -m "feat(ui): Android feature pages dark theme"
```

---

### Task 9: Font Integration (Both Platforms)

**Files:**
- Create: ios Font files (`Inter-Regular`, `Inter-Semibold`, `Inter-Bold`, `SpaceGrotesk-Semibold`, `SpaceGrotesk-Bold`)
- Create: android font resources (`res/font/inter_regular.ttf`, etc.)

- [ ] **Step 1: Download Inter font files**

Download from https://github.com/rsms/inter/releases (Inter v4.0, SIL OFL): `Inter-Regular.otf`, `Inter-SemiBold.otf`, `Inter-Bold.otf`

- [ ] **Step 2: Download Space Grotesk font files**

Download from https://github.com/floriankarsten/space-grotesk/releases (SIL OFL): `SpaceGrotesk-SemiBold.otf`, `SpaceGrotesk-Bold.otf`

- [ ] **Step 3: Add to iOS project**

- Create `ios/AccountCenter/Fonts/` directory
- Add all 5 `.otf` files
- Update `Info.plist` with `UIAppFonts` array:
```xml
<key>UIAppFonts</key>
<array>
    <string>Inter-Regular.otf</string>
    <string>Inter-SemiBold.otf</string>
    <string>Inter-Bold.otf</string>
    <string>SpaceGrotesk-SemiBold.otf</string>
    <string>SpaceGrotesk-Bold.otf</string>
</array>
```

- [ ] **Step 4: Add to Android project**

- Create `android/app/src/main/res/font/` directory
- Add all 5 `.ttf` (convert or download TTF variants) font files
- Reference in `Type.kt` as shown in Task 5 Step 4

- [ ] **Step 5: Commit**

```bash
git add ios/AccountCenter/Fonts/ ios/AccountCenter/Info.plist android/app/src/main/res/font/
git commit -m "feat(ui): add Inter and Space Grotesk fonts"
```

---

### Task 10: Visual Verification

- [ ] **Step 1: Check iOS font rendering**

Run iOS previews (in Xcode on macOS) for each screen. Verify: Space Grotesk renders in titles, Inter in body, brandGradient renders on buttons.

- [ ] **Step 2: Check Android compilation**

Run `cd android && gradlew assembleDebug` (on machine with Gradle/Android SDK). Verify no missing font resource errors.

- [ ] **Step 3: Verify dark theme consistency checklist**

- All screens have `BgPrimary` (`#0D1117`) root background
- All cards use `BgCard` (`#161B22`) with cornerRadius(16)
- All primary buttons use violet-cyan gradient
- All input fields use `BgInput` with focus glow
- All section titles are 13pt semibold uppercase secondary color
- All positive amounts are green (`#2ED573`), negative red (`#FF4757`)
- Tier labels use correct tier colors (gray/blue/orange/purple)
- Risk levels use correct colors (red/orange/yellow/green)
- Feature icons are tinted brandPrimary
- Error messages are danger red Inter 12pt
- All screens use Inter for body text, Space Grotesk for titles

- [ ] **Step 4: Final commit**

```bash
git add .
git commit -m "chore(ui): visual verification and consistency cleanup"
```
