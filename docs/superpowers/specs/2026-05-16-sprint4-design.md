# Sprint 4 设计规格 — 双平台原生移动端架构 + 核心流程

> **版本**: v1.0.0
> **日期**: 2026-05-16
> **状态**: Approved
> **前置**: Sprint 3 (RFM + 脱敏 + VictoriaMetrics) 已完成并推送

---

## 1. 范围

Sprint 4 搭建 iOS (Swift) + Android (Kotlin) 双平台原生移动端，完成架构基础设施和 3 个核心用户流程。

| # | 功能 | 平台 | 优先级 |
|---|------|------|--------|
| 1 | 项目架构搭建（网络层、存储、DI、导航） | iOS + Android | P0 |
| 2 | 登录（密码 + 验证码） | iOS + Android | P0 |
| 3 | 注册（手机号 + 验证码 + 推荐码） | iOS + Android | P0 |
| 4 | 用户中心（信息展示 + 退出） | iOS + Android | P0 |

**不在本 Sprint 范围**:
- 订阅管理、积分、合规、通知、数据产品页面（Sprint 5+）
- 推送通知 (APNs / FCM)
- 生物识别（Face ID / Touch ID / 指纹）
- 离线缓存 / 数据库
- 深色模式
- 国际化 (i18n)
- ECS 部署（推迟至后续 Sprint）

---

## 2. 项目结构

Monorepo，在当前仓库根目录创建 `ios/` 和 `android/`：

```
92-Account-Center/
├── ios/
│   ├── AccountCenter.xcodeproj
│   ├── AccountCenter/
│   │   ├── App/
│   │   │   ├── AccountCenterApp.swift     # @main entry
│   │   │   └── RootView.swift             # NavigationStack root
│   │   ├── Core/
│   │   │   ├── Network/
│   │   │   │   ├── APIClient.swift        # URLSession singleton
│   │   │   │   ├── APIError.swift         # Error enum
│   │   │   │   └── Endpoints.swift        # API path constants
│   │   │   ├── Storage/
│   │   │   │   └── TokenManager.swift     # Keychain token CRUD
│   │   │   └── Auth/
│   │   │       └── AuthManager.swift      # @Observable auth state
│   │   ├── Features/
│   │   │   ├── Login/
│   │   │   │   ├── LoginView.swift
│   │   │   │   └── LoginViewModel.swift
│   │   │   ├── Register/
│   │   │   │   ├── RegisterView.swift
│   │   │   │   └── RegisterViewModel.swift
│   │   │   └── Home/
│   │   │       ├── HomeView.swift
│   │   │       └── HomeViewModel.swift
│   │   └── Models/
│   │       ├── User.swift
│   │       ├── Token.swift
│   │       └── AuthRequests.swift
│   └── AccountCenterTests/
│       ├── APIClientTests.swift
│       └── TokenManagerTests.swift
│
├── android/
│   ├── app/
│   │   ├── build.gradle.kts
│   │   └── src/main/
│   │       ├── AndroidManifest.xml
│   │       ├── java/com/accountcenter/
│   │       │   ├── AccountCenterApp.kt        # Application class
│   │       │   ├── MainActivity.kt
│   │       │   ├── di/
│   │       │   │   ├── AppModule.kt           # Hilt module (network, storage)
│   │       │   │   └── RepositoryModule.kt
│   │       │   ├── network/
│   │       │   │   ├── ApiClient.kt           # Retrofit singleton
│   │       │   │   ├── ApiError.kt
│   │       │   │   ├── AuthInterceptor.kt     # OkHttp token injector
│   │       │   │   └── TokenAuthenticator.kt  # 401 auto-refresh
│   │       │   ├── storage/
│   │       │   │   └── TokenManager.kt        # EncryptedSharedPreferences
│   │       │   ├── repository/
│   │       │   │   ├── AuthRepository.kt
│   │       │   │   └── UserRepository.kt
│   │       │   ├── model/
│   │       │   │   ├── User.kt
│   │       │   │   ├── Token.kt
│   │       │   │   └── AuthRequests.kt
│   │       │   └── ui/
│   │       │       ├── theme/
│   │       │       │   ├── Theme.kt
│   │       │       │   ├── Color.kt
│   │       │       │   └── Type.kt
│   │       │       ├── navigation/
│   │       │       │   └── NavGraph.kt
│   │       │       ├── login/
│   │       │       │   ├── LoginScreen.kt
│   │       │       │   └── LoginViewModel.kt
│   │       │       ├── register/
│   │       │       │   ├── RegisterScreen.kt
│   │       │       │   └── RegisterViewModel.kt
│   │       │       └── home/
│   │       │           ├── HomeScreen.kt
│   │       │           └── HomeViewModel.kt
│   │       └── res/
│   │           ├── values/
│   │           │   ├── strings.xml
│   │           │   └── themes.xml
│   │           └── ...
│   ├── build.gradle.kts          # root build
│   ├── settings.gradle.kts
│   └── gradle.properties
│
├── docker-compose.yml            # (existing, unchanged)
└── ... (existing services)
```

---

## 3. 技术栈

### 3.1 iOS

| 组件 | 选择 | 说明 |
|------|------|------|
| 语言 | Swift 5.9+ | |
| UI 框架 | SwiftUI | iOS 16+ |
| 最低版本 | iOS 16.0 | 覆盖 ~95% 设备 |
| 网络 | URLSession + async/await | 无第三方依赖 |
| Token 存储 | Keychain Services | 安全存储 |
| 状态管理 | @Observable + @Environment | iOS 17 pattern, backward compat via ObservableObject |
| 导航 | NavigationStack | |
| DI | 手动 DI | 通过 @Environment 注入 |
| 构建 | Xcode 15+ / SPM | 无 CocoaPods |

### 3.2 Android

| 组件 | 选择 | 说明 |
|------|------|------|
| 语言 | Kotlin 1.9+ | |
| UI 框架 | Jetpack Compose | BOM 2024.x |
| 最低 SDK | API 24 (Android 7.0) | 覆盖 ~95% 设备 |
| 目标 SDK | API 34 (Android 14) | |
| 网络 | Retrofit 2 + OkHttp 4 | |
| Token 存储 | EncryptedSharedPreferences + DataStore | 安全 + 结构化 |
| 状态管理 | ViewModel + StateFlow | |
| 导航 | Compose Navigation | |
| DI | Hilt | |
| 构建 | Gradle 8+ / AGP 8+ / KSP | |

---

## 4. 网络层设计

所有请求通过 api-gateway (`BASE_URL:30300`) 统一入口。

### 4.1 API 端点（对接现有后端）

| 功能 | Method | Path | 请求体 | 响应 |
|------|--------|------|--------|------|
| 密码登录 | POST | `/api/v1/auth/login` | `{phone, password}` | `{access_token, refresh_token, user}` |
| 验证码登录 | POST | `/api/v1/auth/biometric/login` | `{phone, code}` | `{access_token, refresh_token, user}` |
| 发送验证码 | POST | `/api/v1/sms/send` | `{phone}` | `{message}` |
| 注册 | POST | `/api/v1/account/register` | `{phone, password, code, referral_code?}` | `{user_id, ...}` |
| 刷新 Token | POST | `/api/v1/auth/refresh` | `{refresh_token}` | `{access_token, refresh_token}` |
| 获取用户信息 | GET | `/api/v1/account/:id` | - | `{user_id, phone, email, tier, ...}` |
| 退出登录 | POST | `/api/v1/auth/logout` | - | `{message}` |

### 4.2 Token 管理

```
┌──────────┐    401 Response     ┌──────────────┐
│ API Call │ ──────────────────> │ Refresh Token │
│ (interceptor) │                 │ (auto-retry)  │
└──────────┘ <────────────────── └──────────────┘
                  New Access Token
```

- Access token: 短期（15-30 min），存 Keychain / EncryptedSP
- Refresh token: 长期（7-30 days），同上
- 401 时自动刷新 + 重试原请求（最多 1 次）
- 刷新失败则清除 token，跳转登录页

### 4.3 错误处理

统一 `APIError` 枚举：
- `.networkError(Error)` — 无网络 / 超时
- `.httpError(statusCode, message)` — 4xx / 5xx
- `.unauthorized` — token 过期且刷新失败
- `.decodingError(Error)` — JSON 解析失败
- `.unknown(Error)` — 其他

---

## 5. Feature 设计

### 5.1 登录 (Login)

**UI 流程**:
1. 手机号输入（11 位校验）
2. 选择登录方式：密码 / 验证码
3. 密码登录：输入密码 → 调用 login API
4. 验证码登录：发送验证码 → 60s 倒计时 → 输入验证码 → 调用 biometric/login API
5. 成功：存储 token，跳转 Home
6. 失败：显示错误提示
7. 底部 "注册" 链接 → Register 页

**ViewModel 状态**:
```
phone: String
password: String
verificationCode: String
loginMode: .password | .smsCode
isLoading: Bool
countdownSeconds: Int
errorMessage: String?
```

### 5.2 注册 (Register)

**UI 流程**:
1. 手机号输入（11 位校验）
2. 发送验证码 → 60s 倒计时
3. 输入验证码
4. 设置密码（8-20 位，必须包含字母+数字）
5. 确认密码
6. 可选：推荐码
7. 提交注册 → 调用 register API
8. 成功：自动用密码登录 → 跳转 Home
9. 失败：显示错误提示
10. 底部 "已有账号？登录" 链接 → Login 页

**ViewModel 状态**:
```
phone: String
verificationCode: String
password: String
confirmPassword: String
referralCode: String
countdownSeconds: Int
isLoading: Bool
errorMessage: String?
```

### 5.3 用户中心 (Home)

**UI 流程**:
1. 顶部用户头像占位 + 用户 ID
2. 手机号（脱敏显示：138****1234）
3. 身份等级（tier）
4. 功能入口列表（预留，Sprint 5+ 启用）：
   - 订阅管理
   - 积分中心
   - 安全设置
   - 关于
5. 退出登录按钮

**ViewModel 状态**:
```
user: User?
isLoading: Bool
errorMessage: String?
```

---

## 6. 安全考量

- Token 存储使用平台安全存储（Keychain / EncryptedSharedPreferences）
- 密码不在客户端缓存
- 所有请求通过 HTTPS（生产环境）/ HTTP（开发环境）
- Token 刷新使用 synchronized lock 防并发问题（Android Authenticator）
- 验证码 60s 倒计时防重发

---

## 7. 测试策略

### iOS
- Unit test: APIClient, TokenManager, ViewModels (XCTest)
- UI test: 登录 → 注册 → Home 完整流程

### Android
- Unit test: ApiClient, TokenManager, ViewModels (JUnit + MockK)
- UI test: 登录 → 注册 → Home (Compose TestRule)

---

## 8. 文件清单

### iOS 新增文件（约 20 个）
- App: AccountCenterApp.swift, RootView.swift
- Core/Network: APIClient.swift, APIError.swift, Endpoints.swift
- Core/Storage: TokenManager.swift
- Core/Auth: AuthManager.swift
- Features/Login: LoginView.swift, LoginViewModel.swift
- Features/Register: RegisterView.swift, RegisterViewModel.swift
- Features/Home: HomeView.swift, HomeViewModel.swift
- Models: User.swift, Token.swift, AuthRequests.swift
- Tests: APIClientTests.swift, TokenManagerTests.swift
- Xcode project / SPM config

### Android 新增文件（约 25 个）
- Gradle config: build.gradle.kts (root + app), settings, gradle.properties
- App: AccountCenterApp.kt, MainActivity.kt
- di: AppModule.kt, RepositoryModule.kt
- network: ApiClient.kt, ApiError.kt, AuthInterceptor.kt, TokenAuthenticator.kt
- storage: TokenManager.kt
- repository: AuthRepository.kt, UserRepository.kt
- model: User.kt, Token.kt, AuthRequests.kt
- ui/theme: Theme.kt, Color.kt, Type.kt
- ui/navigation: NavGraph.kt
- ui/login: LoginScreen.kt, LoginViewModel.kt
- ui/register: RegisterScreen.kt, RegisterViewModel.kt
- ui/home: HomeScreen.kt, HomeViewModel.kt
- AndroidManifest.xml, res/values/strings.xml, res/values/themes.xml
