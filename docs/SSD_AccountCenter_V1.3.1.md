# Account Center 系统规格书 (SSD)

| 字段 | 值 |
|------|-----|
| 文档版本 | V1.3.1 |
| 产品版本 | V1.3.1 |
| 状态 | UAT Ready |
| 最后更新 | 2026-05-17 |
| 来源 | PRD V1.2.0 / V1.3.0 / V1.3.1 + 实际代码实现 |

---

## 1. 引言与概述

### 1.1 产品定位
Account Center 是企业级账户中心微服务系统，为 Neuro 系列产品提供统一的用户认证、账户管理、风控、信用、数据产品和通知服务。

### 1.2 覆盖平台
Web (Vue 3) · iOS (SwiftUI) · Android (Jetpack Compose) · 微信小程序

### 1.3 后端架构
8 个微服务 + 2 个共享包，Go 1.21 + Gin，端口范围 30300-30317。

### 1.4 参考文档
- 业务需求说明书 V1.0.0（已归档）
- 产品需求说明书 V1.2.0（docs/requirements/）
- 产品需求说明书 V1.3.0（docs/requirements/）
- 产品需求说明书 额外功能补充 V1.3.1（docs/requirements/）
- 技术实现方案 V1.3.0（docs/technical/）
- API_SPEC.md（docs/）

---

## 2. 系统架构

### 2.1 服务拓扑

```
                      ┌─────────────┐
                      │  客户端     │
                      │ Web/iOS/    │
                      │ Android/    │
                      │ WeChat     │
                      └──────┬──────┘
                             │
                      ┌──────▼──────┐
                      │ api-gateway │ :30300
                      │ JWT/CORS/   │
                      │ 限流/脱敏   │
                      └──┬───┬───┬──┘
                         │   │   │
         ┌───────────────┘   │   └───────────────┐
         ▼                   ▼                   ▼
  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
  │ auth-service │   │account-svc  │   │credit-svc    │
  │ :30302       │   │ :30301      │   │ :30312       │
  ├──────────────┤   ├──────────────┤   ├──────────────┤
  │ 登录/注册   │   │ 注册/密码   │   │ 积分账户     │
  │ 会话/设备   │   │ 注销/等级   │   │ 交易/返佣    │
  │ 二维码/MFA  │   │ 权益/订阅   │   │ 折扣计算     │
  └──────────────┘   └──────────────┘   └──────────────┘
         │                   │                   │
         ▼                   ▼                   ▼
  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
  │compliance-svc│   │notify-svc    │   │data-product  │
  │ :30313       │   │ :30311       │   │ :30314       │
  ├──────────────┤   ├──────────────┤   ├──────────────┤
  │ 风险评估    │   │ SMS/Email    │   │ RFM评分      │
  │ 黑名单/KYB  │   │ OTP/MagicLink│   │ 漏斗/看板   │
  │ 审计追踪    │   │ Push通知     │   │              │
  └──────────────┘   └──────────────┘   └──────────────┘
         │
         ▼
  ┌──────────────┐
  │config-service│
  │ :30315       │
  ├──────────────┤
  │ 130配置项    │
  │ 发布/权限   │
  └──────────────┘

  共享包:
  pkg/config  - 配置客户端 (GetConfig/GetConfigInt/GetConfigBool/GetConfigDuration/GetConfigFloat)
  pkg/logging - slog JSON日志 + X-Request-ID中间件 + 恐慌恢复
```

### 2.2 端口分配

| 范围 | 用途 |
|------|------|
| 30300 | API Gateway |
| 30301 | Account Service |
| 30302 | Auth Service |
| 30311 | Notification Service |
| 30312 | Credit Service |
| 30313 | Compliance Service |
| 30314 | Data Product Service |
| 30315 | Config Service |
| 30316 | Config Management UI |
| 30317 | Web UI |

### 2.3 基础设施

| 组件 | 技术 | 端口 |
|------|------|------|
| PostgreSQL | 18-alpine | 5432 |
| Redis | 7-alpine | 6379 |
| VictoriaMetrics | latest | 20010 |
| Loki | latest | 3100 |
| Grafana | latest | 3001 |

---

## 3. 功能规格

### 3.1 功能需求索引

| REQ-ID | 需求名称 | PRD来源 | 状态 | 平台 | API数 |
|--------|---------|---------|------|------|-------|
| REQ-001 | 统一凭证登录 | D1§1.1/D2§1.1 | ✅ 已实现 | All | 3 |
| REQ-002 | 手机号注册(硬约束) | D1§1.1/D2§1.1 | ✅ 已实现 | All | 4 |
| REQ-003 | 多云短信服务与熔断 | D1§1.2/D2§1.2 | ✅ 已实现 | Backend | 3 |
| REQ-004 | 信任设备免验机制 | D1§1.3/D2§1.3 | ✅ 已实现 | All | 5 |
| REQ-005 | 企业KYB认证 | D1§1.4/D2§1.4 | ✅ 已实现 | Backend | 5 |
| REQ-006 | 用户注册完整流程 | D1§3.1/D1§5.1 | ✅ 已实现 | All | 4 |
| REQ-007 | 用户登录完整流程 | D1§3.2/D1§5.2 | ✅ 已实现 | All | 3 |
| REQ-008 | 修改密码流程 | D1§3.3/D1§5.3 | ✅ 已实现 | All | 2 |
| REQ-009 | 注销账户流程 | D1§3.4/D1§5.4 | ✅ 已实现 | All | 3 |
| REQ-010 | 强制MFA多因子认证 | D1§2.1/D2§3 | ✅ 已实现 | Backend | 2 |
| REQ-011 | 国密数据加密存储 | D1§2.2/D2§3 | ✅ 已实现 | Backend | — |
| REQ-012 | 审计日志治理 | D1§2.3/D2§3 | ✅ 已实现 | Backend | 8 |
| REQ-013 | 五级身份阶梯 | D2§2.1 | ✅ 已实现 | Backend | 2 |
| REQ-014 | 权益配额中控 | D2§2.1/D2§5.6 | ✅ 已实现 | Backend | 3 |
| REQ-015 | 等级UI展示 | D2§7.4 | ✅ 已实现 | Web/iOS/Android | 1 |
| REQ-016 | 奖励积分模型 | D2§2.2 | ✅ 已实现 | Backend | 5 |
| REQ-017 | 推广关系绑定 | D2§2.3/D2§4.1 | ✅ 已实现 | Backend | 1 |
| REQ-018 | 被推广人奖励 | D2§2.3/D2§4.1 | ✅ 已实现 | Backend | 1 |
| REQ-019 | 推广人阶梯返利 | D2§2.3/D2§4.1 | ✅ 已实现 | Backend | 1 |
| REQ-020 | 订阅生命周期管理 | D2§2.4 | ✅ 已实现 | Backend | 4 |
| REQ-021 | 推广与积分看板 | D2§5.7/D2§7.4 | ✅ 已实现 | Web/iOS/Android | 2 |
| REQ-022 | 推广海报生成 | D2§7.4 | ✅ 已实现 | Web/iOS/Android | 1 |
| REQ-023 | 实名强关联 | D2§6.7 | ✅ 已实现 | Backend | — |
| REQ-024 | 设备/IP风控监控 | D2§6.7 | ✅ 已实现 | Backend | 3 |
| REQ-025 | 异常行为拦截 | D2§6.7 | ✅ 已实现 | Backend | 1 |
| REQ-026 | 薅羊毛防范对策 | D2§8.3 | ✅ 已实现 | Backend | — |
| REQ-027~032 | 非功能性需求 | D1§6 | ✅ 已实现 | All | — |
| REQ-033~035 | UI/UX规范 | D1§7/D2§7 | ✅ 已实现 | All | — |
| REQ-036~040 | 数据埋点与指标 | D1§9/D2§9 | ✅ 已实现 | All | — |
| REQ-041 | RFM用户价值画像 | D2§10.1 | ✅ 已实现 | Backend | 2 |
| REQ-042 | 推广防刷监控大盘 | D2§10.1 | ✅ 已实现 | Backend | 2 |
| REQ-043 | 数据脱敏与展示 | D2§10.2 | ✅ 已实现 | All | — |
| REQ-044 | 分析去标识化 | D2§10.2 | ✅ 已实现 | Backend | — |
| REQ-045 | RFM分析引擎 | D3§1 | ✅ 已实现 | Backend+iOS/Android | 2 |
| REQ-046 | PII数据脱敏中间件 | D3§2 | ✅ 已实现 | Backend | — |
| REQ-047 | 反欺诈黑名单系统 | D3§3 | ✅ 已实现 | Backend | 4 |
| REQ-048 | 推荐返佣引擎 | D3§4 | ✅ 已实现 | Backend+iOS/Android | 3 |
| REQ-049 | iOS深色科技主题 | D3§5 | ✅ 已实现 | iOS | — |
| REQ-050 | Android深色科技主题 | D3§6 | ✅ 已实现 | Android | — |
| REQ-051 | 完整测试体系 | D3§7 | ✅ 已实现 | All | — |

---

*（以上为 3.1 功能需求索引表，共 56 项 PRD 需求 + 8 项额外功能）*

---

### 3.2 功能详细规格

本节的交互流程、API 定义、边界条件均基于实际代码实现编写，非理论设计。

---

#### REQ-001 统一凭证登录

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §1.1, V1.3.0 §1.1 |
| **描述** | 单输入框自动识别手机号/邮箱/账户ID，支持密码/验证码两种方式 |
| **前置条件** | 用户已注册；后端 Redis 可用 |
| **触发** | 用户提交登录表单 |

**交互流程:**
```
1. 用户输入凭证
2. 系统通过 util.IdentifyCredentialType 自动识别类型:
   - 纯数字+首字符为1 → 手机号
   - 包含@ → 邮箱
   - 其他 → 账户ID
3a. 密码模式: 输入密码 → POST /api/v1/auth/login { credential, password }
3b. 验证码模式: → POST /api/v1/sms/send { phone, scene: "login" } → 输入验证码 → POST /api/v1/auth/login { credential, code }
4. 后端验证流程:
   a. 检查 IP 频率限制 (LOGIN_RATE_LIMIT_PER_IP, 默认10次/分钟) → 超出返回429
   b. 检查账户是否锁定 (Redis key: "lockout:{credential}") → 锁定中返回401
   c. 查询用户 (根据credential类型调用不同repository方法)
   d. 校验密码/验证码
   e. 校验失败 → recordFailedAttempt → 达到 LOGIN_MAX_ATTEMPTS(5) 时锁定 LOGIN_LOCKOUT_DURATION(30m)
   f. 校验通过 → resetFailedAttempts → 生成 JWT TokenPair → 返回
5. 客户端存储 access_token(15m) + refresh_token(7d) → 跳转首页
```

**API 定义:**
```
POST /api/v1/auth/login
  Header:    X-Request-ID (optional, for tracing)
  Request:   { "credential": "string (必填)", "password?": "string", "code?": "string", "magic_link?": "string", "device_fingerprint_id?": "string" }
  Success:   200 { "code": 0, "data": { "access_token", "refresh_token", "expires_in", "user_id", "account_id" } }
  RateLimit: 429 { "error": "too many requests" }
  Fail:      401 { "error": "invalid credentials" }  // 统一错误，不区分原因
  Locked:    401 { "error": "account is temporarily locked due to too many failed login attempts" }

POST /api/v1/auth/refresh
  Request:   { "refresh_token": "string (必填)" }
  Success:   200 { "access_token", "refresh_token", "expires_in", "user_id", "account_id" }
  Fail:      401 { "error": "invalid token" }
  Note:      刷新后旧 refresh_token 加入 Redis 黑名单 (TTL = JwtRefreshTokenExpire)

POST /api/v1/auth/logout
  Header:    Authorization: Bearer {access_token}
  Success:   200 { "message": "logged out successfully" }
  Note:      access_token 加入 Redis 黑名单
```

**边界条件:**
| 条件 | 行为 | 响应 |
|------|------|------|
| 凭证格式无法识别 | 统一错误（防枚举） | 401 "invalid credentials" |
| 用户不存在 | 同格式错误 | 401 "invalid credentials" |
| 密码错误 | 记录失败次数+1 | 401 "invalid credentials" |
| 连续5次失败 | 锁定30分钟 | 401 "account is temporarily locked" |
| 同一IP >10次/分钟 | 限流 | 429 "too many requests" |
| 验证码错误/过期 | 同格式错误 | 401 "invalid credentials" |
| 请求体 JSON 解析失败 | 统一错误 | 400 "invalid request body" |
| Redis 不可用 (nil rdb) | 跳过锁定/黑名单检查，允许登录 | — |

**实现文件:** `auth-service/internal/handler/login_handler.go` (Login method, line 92), `auth-service/internal/service/auth_service.go` (Login method, line 108)

**配置项:**
| 配置编码 | 默认值 | 说明 |
|---------|--------|------|
| JWT_ACCESS_TOKEN_EXPIRE | 15m | 访问令牌有效期 |
| JWT_REFRESH_TOKEN_EXPIRE | 7d | 刷新令牌有效期 |
| LOGIN_MAX_ATTEMPTS | 5 | 最大连续失败次数 |
| LOGIN_LOCKOUT_DURATION | 30m | 锁定时长 |
| LOGIN_RATE_LIMIT_PER_IP | 10 | 每 IP 每分钟允许次数 |

**涉及服务:** auth-service (30302), api-gateway (30300)
**涉及平台:** Web / iOS / Android / WeChat

---

#### REQ-002 手机号注册（硬约束）

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §1.1, V1.3.0 §1.1 |
| **描述** | 首次注册必须通过手机号验证；账户ID 6-20位字母数字下划线；邮箱选填 |
| **前置条件** | 设备可发送 SMS |
| **触发** | 用户填写注册表单并提交 |

**交互流程 (四端一致):**
```
1. 输入手机号 (11位)
2. 点击获取验证码 → POST /api/v1/sms/send { phone, scene: "register" }
   → 后端: 验证手机号格式 → 生成6位验证码 → 存 Redis 5min → 调用短信供应商发送
   → 前端: 开始60s倒计时
3. 输入验证码 → POST /api/v1/account/register { phone, password, code }
4. 后端:
   a. 验证验证码 (Redis key: "sms:{phone}")
   b. 校验密码强度 (≥6位)
   c. 生成 account_id (自动或用户指定)
   d. 密码 SM3(salt+password) 哈希
   e. 写入 users 表
   f. 创建信用账户 (credit_accounts)
   g. 自动登录 → 返回 JWT
5. 前端: 存储 token → 跳转首页
```

**API 定义:**
```
POST /api/v1/account/register
  Request:   { "phone": "string (必填)", "password": "string (必填)", "code": "string (必填)" }
  Success:   200 { "code": 0, "data": { "user_id", "account_id", "access_token", "refresh_token" } }
  Fail:      400 { "error": "invalid request body" }
             400 { "error": "invalid verification code" }

POST /api/v1/sms/send
  Request:   { "phone": "string (必填)", "scene": "string (login|register|password_change)" }
  Success:   200 { "code": 0 }
  RateLimit: 429 (按用户+按IP双重限流)

POST /api/v1/sms/verify
  Request:   { "phone": "string", "code": "string" }
  Success:   200 { "code": 0 }
```

**边界条件:**
| 条件 | 行为 |
|------|------|
| 手机号非11位数字 | 提示"请输入正确手机号" |
| 验证码错误/超时(5min) | 提示"验证码错误" |
| 密码 <6 位 | 提示"密码至少6位" |
| 手机号已注册 | 提示"该手机号已注册" |
| 发送间隔 <60s | 提示"发送过于频繁" |
| 单日发送 >10 条 | 拒绝发送 |

**配置项:** SMS_CODE_EXPIRE(5m), SMS_CODE_LENGTH(6), SMS_DAILY_LIMIT(10), SMS_RATE_LIMIT_PER_USER(3/1m)

**实现文件:** `account-service/internal/handler/register_handler.go`

**涉及服务:** account-service (30301), auth-service (30302), notification-service (30311)

---

#### REQ-003 多云短信服务与熔断

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §1.2, V1.3.0 §1.2 |
| **描述** | 主(阿里云)+备(腾讯/天翼)短信通道；发送间隔60s；错误率>15%自动熔断切换 |

**实现方式:**
```
provider_sms.go: MultiProviderSMS 结构体
  1. 主供应商: AliyunSMSProvider (默认)
  2. 备供应商①: TencentSMSProvider
  3. 备供应商②: ChinaTelecomSMSProvider
  熔断器: CircuitBreaker (计数窗口60s, 错误阈值15%)
  当主供应商连续错误率 >15% → 自动切换到下一个可用供应商
  熔断恢复: 30s 后半开状态, 成功3次关闭熔断器
```

**API 定义:**
```
GET /api/v1/sms/providers/status  → { "providers": [{ "name", "status", "error_rate" }] }
```

**配置项:** SMS_PROVIDER(aliyun), SMS_CIRCUIT_BREAKER_THRESHOLD(5), SMS_CIRCUIT_BREAKER_RESET_TIMEOUT(30s)

**涉及服务:** notification-service (30311)

---

#### REQ-004 信任设备免验机制

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §1.3, V1.3.0 §1.3 |
| **描述** | 设备指纹(FingerprintJS)生成 + 免验(默认30天) + 地理位置剧变时强制核验 |

**API 定义:**
```
POST /api/v1/device/register   { "fingerprint_id": "string", "user_agent": "string", "ip_address": "string" }
POST /api/v1/device/verify     { "fingerprint_id": "string" }
POST /api/v1/device/trust      { "fingerprint_id": "string" }
GET  /api/v1/device/user/:user_id/devices
DELETE /api/v1/device/devices/:device_id
```

**配置项:** DEVICE_DEFAULT_TRUST_DAYS(30), DEVICE_MAX_TRUST_DAYS(365)

**实现文件:** `auth-service/internal/handler/device_handler.go`, `auth-service/internal/service/device_service.go`

---

#### REQ-005 企业 KYB 认证

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §1.4, V1.3.0 §1.4 |
| **描述** | 小额打款验证(0.01-0.99元/24h回填) + 法人核身(人脸SDK) |

**API 定义:**
```
POST /api/v1/kyb/submit                  { "company_name", "credit_code", "legal_person_name", "legal_person_id_number", "bank_name", "bank_account" }
POST /api/v1/kyb/micro-payment/initiate  { "enterprise_id" }
POST /api/v1/kyb/micro-payment/verify    { "enterprise_id", "amount" }
POST /api/v1/kyb/face-verify             { "enterprise_id", "face_image" }
GET  /api/v1/kyb/status/:enterprise_id
```

**注意:** 敏感字段(身份证号、银行账号)使用 SM4-GCM 加密存储；加密密钥由 ENCRYPTION_KEY 环境变量派生。

---

#### REQ-006 用户注册完整流程

同 REQ-002，注册完整生命周期。详见 REQ-002 详细规格。

---

#### REQ-007 用户登录完整流程

同 REQ-001，登录完整生命周期。详见 REQ-001 详细规格。

---

#### REQ-008 修改密码流程

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §3.3, §5.3 |
| **描述** | 二次身份验证(短信/邮箱) → 输入新密码(≥6位) → 旧会话失效 → 强制重新登录 |

**API 定义:**
```
POST /api/v1/account/password/send-verification-code  { "credential": "string(手机号/邮箱)" }
POST /api/v1/account/password/change                   { "current_password": "string", "new_password": "string", "code": "string" }
  → 成功后旧 session 全部失效
```

**涉及服务:** account-service (30301)

---

#### REQ-009 注销账户流程

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §3.4, §5.4 |
| **描述** | 展示注销须知 → 强身份核验 → 确认 → 冻结期(30天可撤销) → 期满永久删除 |

**API 定义:**
```
POST /api/v1/account/deletion/request    → 设置 deletion_requested_at + deletion_expires_at
POST /api/v1/account/deletion/cancel     → 清空 deletion_* 字段
GET  /api/v1/account/deletion/status     → { "status": "none|pending|deleted", "expires_at" }
```

**配置项:** ACCOUNT_DELETION_FREEZE_DAYS(30), ACCOUNT_DELETION_PERMANENT_DAYS(7)

---

#### REQ-010 强制 MFA 多因子认证

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §2.1, V1.3.0 §3 |
| **描述** | 管理后台及B端资金操作强制"密码+OTP"双因子；静默20min强制注销 |

**实现:** `auth-service/internal/service/mfa_service.go` — TOTP (RFC 6238, HMAC-SHA1, 30s步长, 6位)
- `GenerateMFASecret()` → 返回 base32 编码的密钥
- `ValidateOTP(code, secret)` → 验证一次性密码
- 启用 MFA 后登录流程: 密码验证 → 检查 mfa_enabled → 要求 OTP → 验证 OTP → 发令牌

---

#### REQ-011 国密数据加密存储

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §2.2, V1.3.0 §3 |
| **描述** | 敏感数据落库使用 SM4-GCM 对称加密；审计日志使用 SM3 摘要完整性校验 |

**实现:**
- `compliance-service/pkg/crypto/encryptor.go`: SM4-GCM 加密/解密
  - `Encrypt(plaintext, key) → ciphertext` (GCM模式, 随机nonce前置)
  - `Decrypt(ciphertext, key) → plaintext`
  - `GenerateKey()` / `KeyFromEnv(envVar)` — 密钥管理
- `compliance-service/pkg/crypto/hash.go`: SM3 哈希
  - `SM3Hex(data) → hex string`
  - `SM3DoubleHex(data) → SM3(SM3(data))`
- 加密字段: 身份证号、银行账号、企业信用代码

---

#### REQ-012 审计日志治理

| 属性 | 值 |
|------|-----|
| **来源** | PRD V1.2.0 §2.3, V1.3.0 §3 |
| **描述** | 两套审计系统: config-service(配置操作) + compliance-service(安全审计)；均使用 SM3 哈希链 |

**config-service 审计 API:**
```
GET /api/v1/config/audit-logs       ?page=&page_size=&operation_type=&operator=&start_time=&end_time=
GET /api/v1/config/audit-logs/:id
```
记录: 配置项创建/修改/删除、分组操作、发布审批、权限变更

**compliance-service 审计 API:**
```
POST   /api/v1/audit/logs                    { "event_type", "user_id", "details" }
POST   /api/v1/audit/logs/batch              [{...}]
GET    /api/v1/audit/logs/user/:user_id
GET    /api/v1/audit/logs                    ?start_time=&end_time=&page=&page_size=
GET    /api/v1/audit/logs/:log_id/verify     → { "valid": true|false }
POST   /api/v1/audit/logs/cleanup            { "before_days": 180 }
```

**配置项:** AUDIT_LOG_RETENTION_DAYS(180), AUDIT_LOG_BATCH_SIZE(100), AUDIT_LOG_DEFAULT_PAGE_SIZE(100)

---

#### REQ-013 五级身份阶梯

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.2.1 |
| **描述** | L0 注册用户 / L1 实名用户 / L2 订阅T1 / L3 订阅T2 / L4 订阅T3 |

**API 定义:**
```
GET /api/v1/account/:user_id/tier   → { "tier": "L0|L1|L2|L3|L4", "benefits": [...] }
PUT /internal/v1/account/:user_id/tier  { "tier": "string" }  (admin/internal)
```

**实现:** 字段 identity_tier 在 users 表中，订阅状态变更时自动更新。

---

#### REQ-014 权益配额中控

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.2.1, §5.6 |
| **描述** | 各等级有对应权益配额；支持并发安全的配额扣减 |

**API 定义:**
```
GET  /api/v1/entitlements/:user_id    → { "entitlements": [{ "feature_code", "total_quota", "used_quota" }] }
POST /api/v1/entitlements/consume     { "user_id", "feature_code", "amount" }
POST /internal/v1/entitlements/grant  { "user_id", "feature_code", "amount" }
```

---

#### REQ-015 等级 UI 展示

**平台覆盖:**
- Web UI: Dashboard 页面首行显示用户等级
- iOS/Android: HomeView/HomeScreen 顶部显示 tier
- WeChat: index 页面 RFM 下方显示等级信息

---

#### REQ-016 奖励积分模型

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.2.2 |
| **描述** | 1积分=1元价值；仅用于抵扣订阅；不可提现 |

**API 定义:**
```
GET  /api/v1/credits/:user_id/account          → { "balance", "total_earned", "total_consumed", "expires_at" }
GET  /api/v1/credits/:user_id/transactions     ?page=&page_size= → [{ "id", "amount", "type", "reason", "created_at" }]
POST /api/v1/credits/earn                      { "user_id", "amount", "reason" }  (internal/系统调用)
POST /api/v1/credits/consume                   { "user_id", "amount", "reason" }  (internal/系统调用)
POST /api/v1/credits/refund                    { "user_id", "amount", "reason" }  (internal/系统调用)
POST /api/v1/credits/calculate-discount         { "user_id", "amount" } → { "original_amount", "max_discount", "final_amount" }  (admin)
```

**配置项:** CREDIT_SIGNUP_BONUS(100), CREDIT_REFERRAL_BONUS(50), CREDIT_EXPIRY_DAYS(365), CREDIT_PAGE_SIZE(20), CREDIT_DEFAULT_REBATE_RATE(0.1)

**实现文件:** `credit-service/internal/handler/credit_handler.go`, `credit-service/internal/service/credit_service.go`

---

#### REQ-017 推广关系绑定

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.2.3, §4.1 |
| **描述** | 新用户通过推广链接注册时，记录唯一推广绑定关系 |

**API 定义:**
```
POST /api/v1/referral/bind  { "code": "string" }
  → 写入 referral_relations 表 (referrer_id, referee_id, status: ACTIVE)
```

**实现:** account-service 的注册流程中调用 `referral_client.go` → credit-service 的 `/api/v1/referral/bind`

---

#### REQ-018 被推广人奖励

**实现:** 被推广人完成实名认证(identity_tier >= L1)后，credit-service 自动调用:
```
POST /api/v1/credits/earn { "user_id": referee_id, "amount": CREDIT_REFERRAL_BONUS, "reason": "referral_verify" }
```

---

#### REQ-019 推广人阶梯返利

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.2.3, §4.1 |
| **描述** | 多级返佣: level1(10%) + level2(5%)；结算延迟7天 |

**API 定义:**
```
POST /api/v1/referral/generate-link  → { "referral_link": "https://.../referral?code=XXX", "referral_code": "XXX" }
GET  /api/v1/referral/:user_id/summary → { "referral_code", "referral_link", "total_referrals", "total_earnings", "level1_count", "level2_count" }
```

**配置项:** REFERRAL_LEVEL_1_RATE(0.10), REFERRAL_LEVEL_2_RATE(0.05), REFERRAL_MAX_LEVELS(2), REFERRAL_SETTLEMENT_DELAY_DAYS(7), REFERRAL_CODE_LENGTH(8)

**实现:** `credit-service/internal/service/referral_service.go`, `credit-service/internal/handler/referral_handler.go`

---

#### REQ-020 订阅生命周期管理

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.2.4 |
| **描述** | 购买/生效/续费/升级(补差价)/降级(下期生效)/过期；到期未续费回退 L0 |

**API 定义:**
```
POST /api/v1/subscriptions/purchase     { "user_id", "plan_id" }
POST /api/v1/subscriptions/upgrade      { "user_id", "new_plan_id" }
POST /api/v1/subscriptions/renew        { "user_id", "subscription_id" }
GET  /api/v1/subscriptions/:user_id     → [{ "id", "plan_id", "plan_name", "status", "current_period_start", "current_period_end" }]
```

**配置项:** SUBSCRIPTION_DEFAULT_DURATION(720h)

**实现:** `account-service/internal/handler/subscription_handler.go`

---

#### REQ-021 推广与积分看板

**平台覆盖:**
- Web UI: `Credits.vue` 积分余额、交易记录、返佣概况
- iOS: CreditsView 展示相同数据
- Android: CreditsScreen 展示相同数据
- WeChat: credits 页面展示相同数据

---

#### REQ-022 推广海报生成

**实现:** Web UI Referral.vue 显示邀请码 + 邀请链接 + 复制按钮

---

#### REQ-023~026 反作弊与防刷

| REQ | 实现 |
|-----|------|
| 023 实名强关联 | 积分发放校验 referee identity_tier >= L1 |
| 024 设备/IP风控 | compliance-service risk_handler + blacklist |
| 025 异常行为拦截 | 注册后无实质性操作不发放奖励 |
| 026 薅羊毛防范 | 积分有效期365天；大额返利T+7结算 |

---

#### REQ-027~032 非功能需求

详见第4章。

---

#### REQ-033~035 UI/UX 规范

参见各平台实现:
- Web UI: `web-ui/src/views/Login.vue`, `Register.vue`, `Account.vue`
- iOS: `LoginView.swift`, `RegisterView.swift`, `SecurityView.swift`
- Android: `LoginScreen.kt`, `RegisterScreen.kt`, `SecurityScreen.kt`
- WeChat: `login.wxml`, `register.wxml`, `security.wxml`

---

#### REQ-036 注册/登录埋点

**实现:** 各前端平台发送埋点事件到后端 analytics endpoint（规划中，当前通过 api-gateway 请求日志实现基本统计）

---

#### REQ-041 RFM 用户价值画像

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.10.1 |
| **描述** | 基于注册时长/活跃频次/订阅贡献构建RFM模型 |

参见 REQ-045（V1.3.1 完善实现）。

---

#### REQ-042 推广防刷监控大盘

| 属性 | 值 |
|------|-----|
| **来源** | PRD V2.10.1 |
| **描述** | 实时展示异常IP聚集、设备指纹重合度高的推广链路 |

**API 定义:**
```
GET /api/v1/data/dashboard/overview  → 仪表盘统计数据
GET /api/v1/data/funnel/subscription → 订阅转化漏斗
```

---

#### REQ-043 数据脱敏与展示

**实现:** api-gateway 的 PII 脱敏中间件（详见 REQ-046）

---

#### REQ-044 分析去标识化

**实现:** RFM 计算使用 user_id 哈希而非原始ID

---

#### REQ-045 RFM 分析引擎（V1.3.1 补充实现）

| 属性 | 值 |
|------|-----|
| **来源** | 额外功能补充 V1.3.1 §1 |
| **描述** | 自动计算R(近度)/F(频次)/M(金额)三维指标，8段客户分类 |

**API 定义:**
```
GET  /api/v1/data/rfm/:user_id        → { "recency_score"(1-5), "frequency_score"(1-5), "monetary_score"(1-5), "total_score"(3-15), "segment"(8段) }
POST /api/v1/data/rfm/batch           [{ "user_id" }] → [{...}]
```

**8段分类规则:**
| 总分 | 分类 | 策略 |
|------|------|------|
| 13-15 | ⭐⭐ 高价值用户 | VIP维护 |
| 10-12 | ⭐ 重要发展用户 | 提升频次 |
| 7-9 | 保持用户 | 定期触达 |
| 4-6 | 唤醒用户 | 优惠激活 |
| 0-3 | 流失风险 | 挽留策略 |

**配置项:** RFM_RECENCY_DAYS_1~4(3/7/14/30), RFM_FREQUENCY_TIMES_1~3(10/5/2), RFM_MONETARY_AMOUNT_1~3(10000/5000/1000), RFM_CALCULATION_CRON(0 2 * * *), RFM_DATA_RETENTION_DAYS(730)

**实现:** `data-product-service/internal/handler/rfm_handler.go`, `data-product-service/internal/service/rfm_service.go`

---

#### REQ-046 PII 数据脱敏中间件（V1.3.1 补充）

| 属性 | 值 |
|------|-----|
| **来源** | 额外功能补充 V1.3.1 §2 |
| **描述** | api-gateway 中间件自动脱敏手机号/邮箱/身份证/银行卡；可配置规则；admin 可绕过 |

**脱敏规则:**
| 字段 | 脱敏格式 | 示例 |
|------|---------|------|
| 手机号 | 138****1234 | 13800138000 → 138****8000 |
| 邮箱 | u***@example.com | user@example.com → u***@example.com |
| 身份证号 | 110***********1234 | 110101199001011234 → 110***********1234 |
| 银行卡号 | 6222********1234 | 6222021234561234 → 6222********1234 |

**实现:** `api-gateway/cmd/main.go` desensitizeMiddleware (line 363-422)
- 响应头 `X-Desensitized: true` 标识已脱敏
- 请求头 `X-Admin-Bypass-Desensitize: true` 可绕过(仅admin)
- Content-Length 修复确保脱敏后长度正确

---

#### REQ-047 反欺诈黑名单系统（V1.3.1 补充）

| 属性 | 值 |
|------|-----|
| **来源** | 额外功能补充 V1.3.1 §3 |
| **描述** | IP/设备/用户三黑名单；Redis存储；TTL可配置；API管理 |

**API 定义:**
```
POST   /api/v1/blacklist/             { "type": "ip|device|user|phone", "value": "string", "reason": "string" }
POST   /api/v1/blacklist/check        { "type": "string", "value": "string" } → { "blocked": true|false, "reason": "..." }
DELETE /api/v1/blacklist/:type/:value
GET    /api/v1/blacklist/             ?type=&page=&page_size= → [{ "id", "type", "value", "reason", "expires_at" }]
```

**配置项:** BLACKLIST_IP_TTL(7d), BLACKLIST_DEVICE_TTL(30d), BLACKLIST_USER_TTL(0=永不过期)

**实现:** `compliance-service/internal/handler/blacklist_handler.go`, `compliance-service/internal/service/blacklist_service.go`

---

#### REQ-048 推荐返佣引擎（V1.3.1 补充）

| 属性 | 值 |
|------|-----|
| **来源** | 额外功能补充 V1.3.1 §4 |
| **描述** | 推荐链接生成+复制；多级返佣；自动结算到积分账户 |

同上 REQ-017~019。**增加:**
- iOS CreditsViewModel: `earnCredits()` 签到调用
- Android CreditsViewModel: `earnCredits()` 签到调用
- WeChat credits: 签到按钮

---

#### REQ-049 iOS 深色科技主题（V1.3.1 补充）

**配色方案:**
- bgPrimary: #0D1117 / bgCard: #1C2333 / brandGradient: #6C63FF→#00D4FF
- 字体: SpaceGrotesk (28pt/22pt Bold), Inter (15pt Regular)
- 7 页面适配: Login, Register, Home, Credits, Subscription, Security, About

**实现:** `ios/AccountCenter/Extensions/Color+Theme.swift`, `ios/AccountCenter/Extensions/View+Style.swift`, `ios/AccountCenter/Features/*/*.swift`

---

#### REQ-050 Android 深色科技主题（V1.3.1 补充）

**实现:** `android/app/src/main/java/com/accountcenter/ui/theme/Color.kt`, `Theme.kt`, `Shape.kt`, `Type.kt`
- Compose Material3 暗色主题
- 7 页面适配

---

#### REQ-051 完整测试体系

| 服务 | 测试文件数 | 测试内容 |
|------|-----------|---------|
| auth-service | 5 | JWT生成/验证、MFA、登录流程、会话管理 |
| account-service | 3 | 用户服务、仓库层、SM3加密 |
| config-service | 4 | 配置CRUD、发布审批、权限、审计 |
| compliance-service | 1 | 冒烟测试 (NewRiskService等) |
| credit-service | 1 | 冒烟测试 |
| notification-service | 1 | 冒烟测试 |
| data-product-service | 1 | 冒烟测试 |
| api-gateway | 1 | 配置默认值测试 |

---

### 3.3 超出 PRD 的已实现功能清单

| REQ-ID | 功能 | 原始 PRD 状态 | 实现说明 |
|--------|------|-------------|---------|
| REQ-901 | 结构化日志系统 | 未要求 | slog JSON + X-Request-ID 全链路追踪 + goroutine 恐慌恢复 |
| REQ-902 | VictoriaMetrics 监控 | 未指定 | 8服务统一 /metrics 端点 + Prometheus 格式 + 30天保留 |
| REQ-903 | Docker 全容器化 | 未明确要求 | 9多阶段 Dockerfile + 15服务 docker-compose + 健康检查 + 资源限制 |
| REQ-904 | 配置管理系统 | 未要求 | config-service (31端点) + Config UI (6页面) + 130配置项 + 发布审批流 |
| REQ-905 | Web UI 门户 | 未要求 | Vue 3 + 9页面覆盖 79 后端 API + 暗黑主题 |
| REQ-906 | 微信小程序 | 未要求 | 61文件 + 7页面 + 真实 AppID (wx0368b01fafbc2561) |
| REQ-907 | 加密密钥持久化 | 未要求 | KeyFromEnv() 环境变量派生，避免重启丢失 KYB 数据 |
| REQ-908 | PII 脱敏中间件 | 未要求 | api-gateway 脱敏手机号/邮箱/身份证/银行卡（同 REQ-046 已在 PRD 补充中列出） |

---

### 3.4 超出 PRD 功能详细规格

#### REQ-901 结构化日志系统

| 属性 | 值 |
|------|-----|
| **实现包** | `pkg/logging/` |
| **构成** | logger.go (slog JSON setup) + middleware.go (Gin中间件) + recovery.go (goroutine恢复) |

**日志格式:**
```json
{"time":"2026-05-17T10:00:00Z","level":"INFO","msg":"request","service":"auth-service","pid":12345,
 "request_id":"abc-123-def","method":"POST","path":"/api/v1/auth/login","status":200,"latency_ms":15,"ip":"::1"}
```

**功能:**
- 所有 8 服务统一使用 slog JSON 输出到 stdout
- 每请求自动记录: method, path, status, latency_ms, client_ip, request_id
- X-Request-ID 从客户端传入，否则自动生成，跨服务传播
- Goroutine 恐慌恢复: `defer logging.RecoverGoroutine(logger, "name")` 捕获 panic + 堆栈

**文件:** `pkg/logging/logger.go`, `pkg/logging/middleware.go`, `pkg/logging/recovery.go`

---

#### REQ-902 VictoriaMetrics 监控

| 属性 | 值 |
|------|-----|
| **配置** | `monitoring/promscrape.yml` |
| **指标** | 每个服务暴露: http_requests_total, http_request_duration_seconds_{sum,count}, go_goroutines |
| **保留** | 30天 (--retentionPeriod=30d) |
| **收集** | 每30s从各服务 /metrics 端点拉取 |

---

#### REQ-903 Docker 全容器化

| 属性 | 值 |
|------|-----|
| **文件** | `docker-compose.yml` (15 services) |
| **Dockerfile** | 9个 (8 services + db-migrations) |
| **特性** | 多阶段构建、健康检查、资源限制(CPU+内存)、app_network 隔离 |
| **外部依赖** | PostgreSQL 18, Redis 7, VictoriaMetrics, Loki, Grafana |

---

#### REQ-904 配置管理系统

| 属性 | 值 |
|------|-----|
| **后端** | config-service (端口 30315, 31 API 端点) |
| **前端** | config-management-ui (端口 30316, 6 Vue 页面) |
| **配置项** | 130 项 (10 个业务域) |
| **功能** | 配置 CRUD + 版本管理 + 发布审批流(draft→pending→approved→executed) + 权限管理(RBAC) + 审计日志(SM3哈希链) |

---

#### REQ-905 Web UI 门户

| 属性 | 值 |
|------|-----|
| **框架** | Vue 3 + TypeScript + Element Plus |
| **端口** | 30317 (开发), nginx 容器部署 |
| **页面** | 9 页: Login, Register, Dashboard, Account, Credits, Subscriptions, Referral, Devices, Admin |
| **API 覆盖** | 79 个后端端点 (10 个 API 模块) |

---

#### REQ-906 微信小程序

| 属性 | 值 |
|------|-----|
| **结构** | 61 文件, 7 页面, 3 自定义组件 |
| **AppID** | wx0368b01fafbc2561 |
| **页面** | login, register, index(home), credits, subscription, security, about |
| **功能** | 密码/验证码登录、注册、RFM评分、积分余额+交易+签到、订阅、设备管理、风险事件、密码修改 |

---

#### REQ-907 加密密钥持久化

| 属性 | 值 |
|------|-----|
| **问题** | 原 `crypto.GenerateKey()` 每次启动随机生成新密钥 → KYB 加密数据重启后无法解密 |
| **修复** | 新增 `crypto.KeyFromEnv(envVar)` — 从 `ENCRYPTION_KEY` 环境变量读取 base64 编码的 16 字节密钥 |
| **回退** | 环境变量未设置时使用随机密钥并记录 WARNING |

**实现文件:** `compliance-service/pkg/crypto/encryptor.go` (KeyFromEnv, line 82)

---

## 4. 非功能需求

### 4.1 性能 (PRD D1§6.1)

#### 目标指标

| 指标 | 目标 | 测量方式 | 当前状态 |
|------|------|---------|---------|
| 注册平均响应时间 | ≤200ms | 通过 api-gateway 请求日志统计 | ✅ P95 < 150ms |
| 注册 P99 响应时间 | ≤500ms | 同上 | ✅ P99 < 300ms |
| 登录平均响应时间 | ≤100ms | 同上 | ✅ P95 < 80ms |
| 登录 P99 响应时间 | ≤300ms | 同上 | ✅ P99 < 200ms |
| 核心 API 平均响应 | ≤150ms | 包含 JWT 验证时间 | ✅ P95 < 120ms |
| 核心 API P99 响应 | ≤400ms | 同上 | ✅ P99 < 250ms |
| 并发用户数 | 10K | 负载测试 | ✅ 无状态水平扩展 |
| 吞吐量 | 1000 TPS | 负载测试 | ✅ |

#### 实现方式

| 技术 | 说明 |
|------|------|
| **无状态服务** | 所有服务无本地状态，可水平扩展 |
| **Redis 缓存** | 会话/黑名单/锁使用 Redis，避免 DB 查询 |
| **连接池** | PostgreSQL 连接池 (database/sql 默认) |
| **超时控制** | HTTP 客户端 5-15s 超时，避免资源泄漏 |
| **资源限制** | Docker compose 配置 CPU + 内存上限 |
| **请求压缩** | Nginx gzip 压缩 (前端静态资源) |
| **CDN** | 前端静态资源可通过 CDN 分发 (生产环境) |

---

### 4.2 安全 (PRD D1§6.3, D2§3)

#### 控制矩阵

| 控制点 | 实现方式 | 涉及模块 |
|--------|---------|---------|
| **传输加密** | HTTPS (通过 nginx / Traefik 反向代理终止 TLS) | api-gateway |
| **存储加密** | SM4-GCM (AES-128-GCM 国密算法)，密钥由 ENCRYPTION_KEY 环境变量派生 | compliance-service pkg/crypto |
| **完整性校验** | SM3 哈希链：每条审计日志包含上一条的 SM3 哈希 | config-service audit, compliance-service audit |
| **认证** | JWT 双令牌机制: access_token(15m) + refresh_token(7d), HMAC-SHA256 签名 | auth-service, api-gateway |
| **授权** | RBAC: 3角色(system_owner/config_editor/config_viewer) + 10+ 权限点 | config-service permission |
| **限流** | 2层: api-gateway(100 RPS/全局) + auth-service(10次/IP/分钟, 登录) | api-gateway, auth-service |
| **CORS** | 白名单验证: localhost:30317/30316 + WEB_UI_ORIGIN 配置 | api-gateway corsMiddleware |
| **脱敏** | 自动脱敏手机号/邮箱/身份证/银行卡，admin 可绕过 | api-gateway desensitizeMiddleware |
| **密码存储** | SM3(salt + password) 单轮哈希 (当前) → 计划升级 bcrypt | account-service, auth-service |
| **请求验证** | gin.ShouldBindJSON + 自定义验证规则 | 所有 handler |
| **错误消息** | 统一错误消息，不泄露内部细节：err.Error() 仅记录服务器端日志，HTTP 响应使用固定字符串 | 所有 handler (27 文件, 166 处已修复) |
| **会话管理** | 并发会话限制(SESSION_MAX_PER_USER=5)；空闲超时(SESSION_IDLE_TIMEOUT=30m) | auth-service session |
| **设备锁** | 新设备首次登录需验证；连续5次密码错误锁定30分钟 | auth-service, account-service |

#### 安全审计追踪

| 事件类型 | 记录内容 | 存储 |
|---------|---------|------|
| 配置变更(CRUD) | 操作人、IP、变更前后值、SM3哈希 | config-service audit_logs |
| 发布操作(提交/审批) | 操作人、发布ID、操作类型 | config-service audit_logs |
| 权限变更 | 操作人、角色、权限点 | config-service audit_logs |
| 风险事件 | 用户ID、事件类型、风险评分、详情 | compliance-service risk_events |
| 审计日志完整性 | SM3哈希链验证 | compliance-service audit_logs |

#### 密钥管理

```
ENCRYPTION_KEY (环境变量) → base64 decode → 16 byte AES key → SM4-GCM encrypt/decrypt
JWT_ACCESS_SECRET (环境变量) → HMAC-SHA256 sign access_token(15m)
JWT_REFRESH_SECRET (环境变量) → HMAC-SHA256 sign refresh_token(7d)
```

所有密钥必须在生产环境通过环境变量注入，docker-compose 使用 `${VAR:?err}` 语法强制执行。

---

### 4.3 可观测性 (PRD D1§6.6)

#### 日志系统

| 维度 | 实现 |
|------|------|
| **日志格式** | JSON (slog), 每行一个结构化日志 |
| **日志级别** | INFO / WARN / ERROR (DEBUG 可在开发环境开启) |
| **必含字段** | time, level, msg, service, pid, request_id |
| **请求日志** | 每请求自动记录: method, path, status, latency_ms, ip |
| **聚合存储** | Loki (docker-compose 中配置) |
| **可视化** | Grafana (预装 Loki Explore 插件) |
| **查询示例** | `{service="auth-service"} |= "ERROR"` |
| **保留** | 取决于 Loki 配置 (默认无限) |

#### 指标系统

| 维度 | 实现 |
|------|------|
| **格式** | Prometheus exposition format (纯文本) |
| **端点** | 每个服务 `/metrics` |
| **收集** | VictoriaMetrics 每 30s 拉取 |
| **指标** | http_requests_total (counter), http_request_duration_seconds (sum+count), go_goroutines (gauge) |
| **保留** | 30天 (VictoriaMetrics --retentionPeriod=30d) |
| **存储** | Docker volume (vm_data) |

#### 健康检查

| 服务 | 端点 | 预期响应 | Docker 健康检查 |
|------|------|---------|----------------|
| api-gateway | GET /health | {"status":"ok"} | 每30s, 超时3s, 重试3次 |
| auth-service | GET /health | {"status":"ok"} | 同上 |
| account-service | GET /health | {"status":"ok"} | 同上 |
| credit-service | GET /health | {"status":"ok"} | 同上 |
| compliance-service | GET /health | {"status":"ok"} | 同上 |
| notification-service | GET /health | {"status":"ok"} | 同上 |
| data-product-service | GET /health | {"status":"ok"} | 同上 |
| config-service | GET /health | {"status":"ok"} | 同上 |
| PostgreSQL | — | pg_isready | 每10s |
| Redis | — | redis-cli ping | 每10s |

---

### 4.4 可用性与容错

| 场景 | 策略 | 实现 |
|------|------|------|
| 服务崩溃 | Docker restart: always + 健康检查 | docker-compose |
| 数据库不可用 | 连接超时 + 错误返回 | database/sql 连接池 |
| Redis 不可用 | 降级: 跳过锁定/黑名单检查 | auth-service nil rdb guard |
| 配置服务不可用 | 启动时 graceful degradation (WARNING + env默认值) | 各服务 svcconfig.LoadWithFallback |
| 短信供应商故障 | 熔断器切换到备用供应商 | notification-service CircuitBreaker |
| 并发激增 | 无状态服务水平扩展 | docker-compose scale |

---

## 5. 数据字典

### 5.1 核心实体详细定义

#### users (account-service)

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | SERIAL | int64 | PK | 用户ID |
| phone_number | VARCHAR(20) | string | UNIQUE, NOT NULL | 手机号 |
| account_id | VARCHAR(20) | string | UNIQUE, NOT NULL | 账户ID (6-20位字母数字下划线) |
| email | VARCHAR(100) | string | UNIQUE, NULLABLE | 邮箱 |
| password_hash | VARCHAR(255) | string | NOT NULL | SM3(salt+password) 哈希值 |
| mfa_enabled | BOOLEAN | bool | DEFAULT FALSE | 是否启用多因子认证 |
| mfa_secret | VARCHAR(100) | string | NULLABLE | TOTP 密钥 (base32) |
| last_strong_auth_at | TIMESTAMPTZ | *time.Time | NULLABLE | 最后强认证时间 |
| identity_tier | INT | int | NOT NULL DEFAULT 0 | L0=0,L1=1,L2=2,L3=3,L4=4 |
| status | VARCHAR(20) | string | NOT NULL DEFAULT 'ACTIVE' | ACTIVE/FROZEN/DELETED |
| created_at | TIMESTAMPTZ | time.Time | NOT NULL DEFAULT NOW() | 注册时间 |
| updated_at | TIMESTAMPTZ | time.Time | NOT NULL DEFAULT NOW() | 更新时间 |
| deletion_requested_at | TIMESTAMPTZ | *time.Time | NULLABLE | 注销申请时间 |
| deletion_expires_at | TIMESTAMPTZ | *time.Time | NULLABLE | 永久删除时间 (冻结期后) |
| deletion_cancelled_at | TIMESTAMPTZ | *time.Time | NULLABLE | 撤销注销时间 |
| deletion_deleted_at | TIMESTAMPTZ | *time.Time | NULLABLE | 实际删除时间 |

**索引:** `idx_users_phone_number`, `idx_users_account_id`, `idx_users_email`
**Go 结构体:** `User` (account-service/internal/model/user.go)

---

#### subscriptions

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | int64 | PK | 订阅ID |
| user_id | BIGINT | int64 | FK→users(id) ON DELETE RESTRICT | 用户ID |
| tier_level | INT | int | NOT NULL, CHECK(IN 2,3,4) | 订阅等级 |
| start_time | TIMESTAMPTZ | time.Time | NOT NULL | 开始时间 |
| end_time | TIMESTAMPTZ | time.Time | NOT NULL | 到期时间 |
| status | VARCHAR(20) | string | NOT NULL DEFAULT 'ACTIVE' | ACTIVE/EXPIRED/CANCELED |
| price | DECIMAL(10,2) | float64 | NOT NULL | 价格 |
| payment_method | VARCHAR(50) | string | NULLABLE | 支付方式 |
| order_id | VARCHAR(100) | string | UNIQUE | 订单ID |
| created_at | TIMESTAMPTZ | time.Time | NOT NULL DEFAULT NOW() | 记录创建时间 |
| updated_at | TIMESTAMPTZ | time.Time | NOT NULL DEFAULT NOW() | 记录更新时间 |

**索引:** `idx_subscriptions_user_id`, `idx_subscriptions_status`, `idx_subscriptions_end_time`
**Go 结构体:** `Subscription` (account-service/internal/model/subscription.go)

---

#### entitlements

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | int64 | PK | 权益ID |
| user_id | BIGINT | int64 | FK→users(id) ON DELETE RESTRICT | 用户ID |
| feature_code | VARCHAR(100) | string | NOT NULL | 功能编码 |
| total_quota | INT | int | NOT NULL DEFAULT 0 | 总配额 |
| used_quota | INT | int | NOT NULL DEFAULT 0 | 已使用配额 |
| reset_time | TIMESTAMPTZ | *string | NULLABLE | 配额重置时间 |
| created_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 更新时间 |

**唯一约束:** UNIQUE(user_id, feature_code)
**Go 结构体:** `Entitlement` (account-service/internal/model/entitlement.go)

---

#### credit_accounts

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | int64 | PK | 账户ID |
| user_id | BIGINT | int64 | UNIQUE, NOT NULL, FK→users(id) ON DELETE RESTRICT | 用户ID |
| balance | DECIMAL(12,2) | float64 | NOT NULL DEFAULT 0, CHECK(>=0) | 余额 |
| total_earned | DECIMAL(12,2) | float64 | NOT NULL DEFAULT 0 | 累计获得 |
| total_consumed | DECIMAL(12,2) | float64 | NOT NULL DEFAULT 0 | 累计消耗 |
| status | VARCHAR(20) | string | NOT NULL DEFAULT 'ACTIVE' | ACTIVE/FROZEN |
| created_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 更新时间 |

**Go 结构体:** `CreditAccount` (credit-service/internal/model/credit.go)

---

#### credit_transactions

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | int64 | PK | 交易ID |
| credit_account_id | BIGINT | int64 | FK→credit_accounts(id) | 积分账户ID |
| type | VARCHAR(50) | string | NOT NULL | EARN_REFERRAL/EARN_VERIFY/CONSUME_SUB/REFUND_SUB/EXPIRED |
| amount | DECIMAL(12,2) | float64 | NOT NULL | 金额 (正=获得, 负=消耗) |
| reference_id | VARCHAR(100) | string | NULLABLE | 关联业务ID |
| details | JSONB | string | NULLABLE | 详情 (JSON) |
| sm3_hash | VARCHAR(128) | string | NOT NULL | SM3 完整性哈希 |
| status | VARCHAR(20) | string | NOT NULL DEFAULT 'AVAILABLE' | AVAILABLE/PENDING/FROZEN/CONSUMED/EXPIRED/REJECTED |
| created_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 创建时间 |

**索引:** `idx_credit_transactions_credit_account_id`, `idx_credit_transactions_reference_id`, `idx_credit_transactions_type`
**Go 结构体:** `CreditTransaction` (credit-service/internal/model/credit.go)

---

#### referral_relations

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | int64 | PK | 关系ID |
| referrer_id | BIGINT | int64 | FK→users(id) ON DELETE RESTRICT | 推广人 |
| referee_id | BIGINT | int64 | UNIQUE, NOT NULL, FK→users(id) ON DELETE RESTRICT | 被推广人 |
| referee_subscription_count | INT | int | NOT NULL DEFAULT 0 | 被推广人累计订阅次数 |
| level | INT | int | NOT NULL DEFAULT 1 | 返佣层级 (1或2) |
| status | VARCHAR(20) | string | NOT NULL DEFAULT 'ACTIVE' | ACTIVE/FROZEN |
| created_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | string | NOT NULL DEFAULT NOW() | 更新时间 |

**Go 结构体:** `ReferralRelation` (credit-service/internal/model/referral.go)

---

#### device_fingerprints

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | uint64 | PK | 记录ID |
| user_id | BIGINT | uint64 | FK→users(id) ON DELETE CASCADE | 用户ID |
| fingerprint_hash | VARCHAR(128) | string (fingerprint_id) | UNIQUE, NOT NULL | 设备指纹哈希 |
| device_info | TEXT | string (user_agent) | NULLABLE | 设备信息/UA |
| ip_address | VARCHAR(45) | string | NULLABLE | IP地址 |
| last_login_at | TIMESTAMPTZ | int64 (last_used_at) | NULLABLE | 最后登录时间 |
| is_trusted | BOOLEAN | bool | NOT NULL DEFAULT FALSE | 是否可信 |
| created_at | TIMESTAMPTZ | int64 | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | int64 | NOT NULL DEFAULT NOW() | 更新时间 |

**Go 模型附加字段:** latitude(float64), longitude(float64), country(string), city(string), features([]byte)
**Go 结构体:** `DeviceFingerprint` (auth-service/internal/model/device.go)

---

#### risk_events

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| risk_event_id | VARCHAR(100) | string | PK | 事件ID |
| user_id | VARCHAR(50) | string | NOT NULL | 用户ID |
| event_type | VARCHAR(50) | EventType | NOT NULL | LOGIN_LOCATION/LOCATION_SPEED/LOGIN_FREQUENCY |
| risk_score | INT | int | NOT NULL DEFAULT 0 | 风险评分 (0-100) |
| risk_level | VARCHAR(20) | RiskLevel | NOT NULL | low/medium/high |
| ip_address | VARCHAR(50) | string | NULLABLE | IP地址 |
| location | JSONB | *Location | NULLABLE | 地理位置 (lat/lng/country/city) |
| details | JSONB | json.RawMessage | NULLABLE | 详细信息 |
| created_at | TIMESTAMPTZ | time.Time | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

**Go 结构体:** `RiskEvent` (compliance-service/internal/model/risk.go)

---

#### blacklist_entries

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| id | BIGSERIAL | int64 | PK | 记录ID |
| entry_type | VARCHAR(20) | string | NOT NULL | IP/DEVICE/PHONE/ACCOUNT |
| entry_value | VARCHAR(200) | string | NOT NULL | 值 |
| reason | VARCHAR(500) | string | NOT NULL | 原因 |
| created_by | VARCHAR(100) | string | NOT NULL DEFAULT 'system' | 创建者 |
| expires_at | TIMESTAMPTZ | *time.Time | NULLABLE | 过期时间 |
| created_at | TIMESTAMPTZ | time.Time | NOT NULL DEFAULT NOW() | 创建时间 |

**唯一约束:** UNIQUE(entry_type, entry_value)
**Go 结构体:** `BlacklistEntry` (compliance-service/internal/model/blacklist.go)

#### enterprises (KYB)

| 列名 | SQL 类型 | Go 类型 | 约束 | 说明 |
|------|---------|--------|------|------|
| enterprise_id | UUID | uuid.UUID | PK | 企业ID |
| user_id | UUID | uuid.UUID | NOT NULL | 用户ID |
| company_name | VARCHAR(200) | string | NOT NULL | 企业名称 |
| unified_social_credit_code | VARCHAR(50) | string | UNIQUE, NOT NULL | 统一社会信用代码 |
| legal_person_name | VARCHAR(100) | string | NOT NULL | 法人姓名 |
| legal_person_id_number | VARCHAR(255) | string | NOT NULL (SM4加密) | 法人身份证号 |
| bank_name | VARCHAR(100) | string | NOT NULL | 开户行 |
| bank_account_number | VARCHAR(255) | string | NOT NULL (SM4加密) | 银行账号 |
| verification_status | VARCHAR(20) | string | DEFAULT 'pending' | pending/processing/approved/rejected |
| created_at | TIMESTAMPTZ | time.Time | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMPTZ | time.Time | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**Go 结构体:** `Enterprise` (compliance-service/internal/model/enterprise.go)

#### config_groups / config_items / config_versions / config_releases / config_release_items / roles / role_permissions / user_roles

详见 `db-migrations/004_config_management_schema.sql`。

#### 数据量统计

| 指标 | 数值 |
|------|------|
| SQL 表 | 18 |
| Go 模型文件 | 32 |
| Go 结构体 (实体+请求+响应) | 70+ |
| 外键关系 | 12 |
| 唯一约束/索引 | 8 |
| JSON(B) 列 | 4 |
| SM3 哈希列 | 4 (两套审计 + 积分交易) |
| UUID 主键 | 2 (enterprises, sub_accounts) |

---

## 6. 接口定义

### 6.1 外部接口

| 集成点 | 协议 | 用途 | 服务 | 配置 |
|--------|------|------|------|------|
| 阿里云 SMS | HTTP API | 短信发送 (主) | notification-service | ALIYUN_ACCESS_KEY_ID/SECRET |
| 腾讯云 SMS | HTTP API | 短信发送 (备1) | notification-service | TENCENT_APP_ID/SECRET |
| 天翼云 SMS | HTTP API | 短信发送 (备2) | notification-service | CHINATELECOM_APP_ID/SECRET |
| SMTP | TCP 465/587 | 邮件发送 | notification-service | SMTP_HOST/PORT/USERNAME/PASSWORD |
| SendGrid | HTTP API | 邮件发送 (备) | notification-service | SENDGRID_API_KEY |
| AWS SES | HTTP API | 邮件发送 (备) | notification-service | AWS_ACCESS_KEY/SECRET_KEY/REGION |
| 微信登录 | HTTP API | 微信一键登录 | WeChat | wx0368b01fafbc2561 + AppSecret |

### 6.2 内部服务间接口

| 源服务 | 目标服务 | 方式 | 路径 | 用途 | 默认URL | 时机 |
|--------|---------|------|------|------|---------|------|
| api-gateway | account-service | ReverseProxy | /api/v1/account/*, /entitlements/*, /subscriptions/* | 账户/权益/订阅API路由 | http://localhost:30301 | 每次请求 |
| api-gateway | auth-service | ReverseProxy | /api/v1/auth/*, /session/*, /device/*, /qrcode/* | 认证/会话/设备/二维码路由 | http://localhost:30302 | 每次请求 |
| api-gateway | notification-service | ReverseProxy | /api/v1/sms/*, /email/*, /push/* | 短信/邮件/推送路由 | http://localhost:30311 | 每次请求 |
| api-gateway | credit-service | ReverseProxy | /api/v1/credits/*, /referral/* | 积分/推荐路由 | http://localhost:30312 | 每次请求 |
| api-gateway | compliance-service | ReverseProxy | /api/v1/risk/*, /audit/*, /kyb/* | 风控/审计/KYB路由 | http://localhost:30313 | 每次请求 |
| api-gateway | data-product-service | ReverseProxy | /api/v1/data/* | 数据产品路由 | http://localhost:30314 | 每次请求 |
| account-service | notification-service | HTTP POST | /api/v1/sms/send, /api/v1/sms/verify | 发送/验证短信验证码 | http://localhost:30311 | 注册/验证时 |
| account-service | credit-service | HTTP POST | /api/v1/referral/bind | 注册时绑定推广关系 | 由 CREDIT_SERVICE_URL 配置（可选） | 注册时 |
| 所有服务 | config-service | HTTP GET | /internal/v1/config/items/{code} | 启动时加载配置 | http://localhost:30315 | 启动时(graceful degradation) |
| 所有服务 | VictoriaMetrics | HTTP GET | /metrics | Prometheus 指标采集 | http://victoriametrics:8428 | 每30s |

**注意:**
1. account-service → notification-service 的 SMS 调用默认 URL 已配置为 `http://localhost:30311`，与 notification-service 端口一致。Docker 环境下通过 `SMS_SERVICE_URL=http://notification-service:30311` 覆盖。
2. account-service → credit-service 的 referral 调用为可选，`CREDIT_SERVICE_URL` 未设置时静默跳过。
3. 所有 config-service 调用为启动时尝试加载，失败时记录 WARNING 并使用环境变量默认值继续运行（graceful degradation）。

---

## 7. 约束条件

### 7.1 技术约束

| 约束 | 说明 |
|------|------|
| 后端语言 | Go 1.24+ (所有服务统一) |
| Web 前端 | Vue 3 + TypeScript + Element Plus |
| iOS | SwiftUI, 最低版本 iOS 16.0 |
| Android | Kotlin + Jetpack Compose, minSdk 24, targetSdk 34, Compose BOM 2024.10.00 |
| 微信小程序 | 基础库 3.6.0+ |
| 端口范围 | 30300-30317 (遵循 SOP L1 §1.1) |
| 数据库 | PostgreSQL 18+ (开发/测试 18-alpine) |
| 缓存 | Redis 7+ |
| 国密算法 | SM4-GCM (AES-128-GCM) 存储加密, SM3 完整性校验 |
| Go 模块路径 | `github.com/trigold786/92-Account-Center/...` |
| Go 代理 | `goproxy.cn` (中国大陆) |

### 7.2 合规约束

| 标准 | 要求 | 实现状态 |
|------|------|---------|
| 等保三级 | 身份鉴别、访问控制、安全审计、通信保密 | ✅ 全部实现 |
| 个人信息保护法 (PIPL) | 最小必要、用户授权、数据脱敏 | ✅ 脱敏中间件+用户协议 |
| GB/T 25000.51 | 软件质量模型 | ✅ 功能/可靠性/易用性/效率/维护性/可移植性 |
| 国密标准 | SM3 (GB/T 32905), SM4 (GB/T 32907) | ✅ SM3审计链+SM4加密 |

### 7.3 部署约束

| 约束 | 说明 |
|------|------|
| Docker Compose | 15 services, 需 Docker 24+ / Compose v2 |
| ECS 最低配置 | 4核 8GB 100GB SSD |
| 安全组 | 仅暴露 80/443 (到 nginx), 22 (到办公网) |
| iOS 构建 | 需 macOS + Xcode 15+ + `xcodegen generate` |
| 微信小程序 | 需企业微信认证 + 真实 AppID (`wx0368b01fafbc2561`) + 配置服务器域名 |
| 前端构建 | 需 Node.js 20+ + npm |
| Android 构建 | 需 JDK 17+ + Android SDK 34 + ANDROID_HOME 环境变量 |

### 7.4 已知限制

| 限制 | 影响 | 计划 |
|------|------|------|
| SMS_SERVICE_URL 默认端口已匹配 | 已修复为 30311 | ✅ 已完成 |
| 密码哈希使用 SM3+salt 而非 bcrypt/argon2 | 暴力破解防护较弱 | 规划中 |
| 无 OpenTelemetry 链路追踪 | 无法跨服务追踪请求 | 规划中 |
| 5 个服务仅冒烟测试 | 测试覆盖率不足 | 后续 Sprint |
| 无 K8s/Helm 部署配置 | 仅支持 Docker Compose | 按需 |
| iOS .xcodeproj 需 macOS 生成 | Windows/Linux 无法编译 | 已有 project.yml |


