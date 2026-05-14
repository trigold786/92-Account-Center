# 账户管理微服务商业化迭代方案与实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 基于 BRD V1.0.0、PRD V1.3.0、TIP V1.3.0，将现有基础账户系统升级为支持五级身份阶梯、奖励积分账务、阶梯退坡返利、全链路防刷风控的商业化账户中台。

**Architecture:** 在现有 11 个微服务基础上，新增 4 个微服务（entitlement-service、credit-service、referral-service、data-product-service），扩展 auth-service/account-service/api-gateway 的功能，引入 PostgreSQL 新表（SUBSCRIPTION、ENTITLEMENT、CREDIT_ACCOUNT、CREDIT_TRANSACTION、REFERRAL_RELATION），并强化风控与审计能力。全部采用 Go/Gin + PostgreSQL + Redis 技术栈，消息队列使用 Redis Streams（开发测试）/ Kafka（生产）。

**Tech Stack:** Go 1.21+, Gin, PostgreSQL 18, Redis 7, Asynq, 国密 SM2/SM3/SM4, VictoriaMetrics, Vector, OpenTelemetry, Docker Compose

---

## 一、当前状态与差距分析

### 1.1 已实现功能（阶段一前期成果）

| 模块 | 服务 | 端口 | 已实现功能 |
|------|------|------|-----------|
| API 网关 | api-gateway | 30300 | 反向代理、JWT 认证、限流、CORS、缓存控制 |
| 用户管理 | account-service | 30301 | 注册、密码修改/重置、账户删除、用户 CRUD |
| 认证服务 | auth-service | 30302 | 登录、JWT 生成/刷新、登出、Redis 黑名单、生物识别 |
| 短信邮件 | sms-email-service | 30303 | 阿里云 SMS、邮件 OTP、熔断器 |
| 企业认证 | kyb-service | 30304 | KYB 提交、小额打款、活体检测、子账号管理 |
| 审计日志 | audit-log-service | 30305 | 日志记录、批量写入、SM3 完整性校验、180 天留存 |
| 风险检测 | risk-detection-service | 30306 | 风险评估、地理异常、设备异常、速率检测 |
| 会话管理 | session-service | 30307 | Redis 会话、并发控制、20 分钟静默超时 |
| 邮件服务 | email-service | 30308 | OTP、Magic Link、多 Provider |
| 设备指纹 | device-fingerprint-service | 30309 | 设备注册/验证/信任（无持久化） |
| 推送通知 | push-notification-service | 30310 | APNs/FCM 等 7 平台 Stub、设备注册（Redis） |

### 1.2 BRD/PRD/TIP 要求但尚未实现的功能

| 需求分类 | 具体需求 | PRD 章节 | TIP 章节 | 当前状态 |
|---------|---------|---------|---------|---------|
| **五级身份阶梯** | L0-L4 身份等级管理、等级自动升降 | PRD §2.1 | TIP §3.3 | ❌ 不存在 |
| **权益中控** | 权益配额管理、高频核销 (<10ms) | PRD §5.6 | TIP §3.3 | ❌ 不存在 |
| **订阅管理** | 购买/续费/升级/降级/到期回退 | PRD §2.4 | TIP §3.3 | ❌ 不存在 |
| **奖励积分账户** | 复式记账、SM3 防篡改摘要链 | PRD §2.2 | TIP §3.4 | ❌ 不存在 |
| **积分交易流水** | 发放/扣减/过期/退款 | PRD §2.2 | TIP §3.4 | ❌ 不存在 |
| **推广关系** | 推广绑定、唯一性、不可篡改 | PRD §2.3 | TIP §3.5 | ❌ 不存在 |
| **阶梯退坡返利** | 50%-30%-20%-10% 算法 | PRD §2.3 | TIP §3.5 | ❌ 不存在 |
| **全链路防刷** | IP/设备限流、延迟发放 T+7、黑名单 | PRD §6.7 | TIP §3.6 | ⚠️ 部分存在（risk-detection 有基础评估） |
| **实名与推广关联** | 实名认证触发积分赠送 | PRD §4.1 | TIP §3.5 | ❌ 不存在 |
| **权益核销 API** | 内部高并发扣减（Lua 脚本） | PRD §5.6 | TIP §3.3 | ❌ 不存在 |
| **数据产品服务** | RFM 画像、推广防刷监控大盘 | PRD §10 | TIP §3 | ❌ 不存在 |
| **动态脱敏网关** | API Gateway 脱敏拦截器 | PRD §10.2 | TIP §5.3 | ❌ 不存在 |
| **DB 迁移** | 新增 5 张商业化表 | — | TIP §2 | ⚠️ 迁移框架存在，缺新表 |
| **go.work 同步** | push-notification-service 未加入 | — | — | ❌ 需修复 |
| **设备指纹持久化** | device-fingerprint-service 无 DB | — | TIP §3.2 | ❌ 需修复 |

### 1.3 数据库差距

**现有表 (5 张):** `users`, `enterprises`, `sub_accounts`, `audit_logs`, `risk_events`

**需新增表 (5 张):** `subscriptions`, `entitlements`, `credit_accounts`, `credit_transactions`, `referral_relations`

**需修改表 (1 张):** `users` — 新增 `identity_tier INT DEFAULT 0` 字段

---

## 二、迭代规划总览

将 TIP 原规划的 4 个阶段调整为 3 个冲刺（Sprint），适配现有代码基础和开发资源：

```
Sprint 1 (T0 → T+3周): 商业化核心 — 身份阶梯、权益中控、订阅管理、积分账务
Sprint 2 (T+4周 → T+6周): 推广返利与风控 — 推广关系、阶梯退坡、防刷强化
Sprint 3 (T+7周 → T+9周): 数据产品与合规 — RFM 画像、脱敏网关、运维完善
```

每个 Sprint 结束时要求：
1. 所有新增/修改服务编译通过 (`go build`, `go vet`)
2. 单元测试通过 (`go test ./...`)
3. 集成测试脚本更新并通过
4. Docker Compose 本地启动验证
5. 代码推送到 GitHub main 分支

---

## 三、Sprint 1：商业化核心（T0 → T+3周）

### 里程碑 1.1：数据库迁移（2天）

**目标：** 新增 5 张商业化表，修改 users 表

**Files:**
- Create: `db-migrations/002_commercialization_schema.sql`
- Modify: `push-notification-service/go.mod`（添加到 go.work）

- [ ] **Step 1: 编写数据库迁移脚本**

创建 `db-migrations/002_commercialization_schema.sql`:

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS identity_tier INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE';

CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    tier_level INT NOT NULL CHECK (tier_level IN (2, 3, 4)),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'EXPIRED', 'CANCELED')),
    price DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(50),
    order_id VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_end_time ON subscriptions(end_time);

CREATE TABLE IF NOT EXISTS entitlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    feature_code VARCHAR(100) NOT NULL,
    total_quota INT NOT NULL DEFAULT 0,
    used_quota INT NOT NULL DEFAULT 0,
    reset_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, feature_code)
);

CREATE INDEX idx_entitlements_user_id ON entitlements(user_id);

CREATE TABLE IF NOT EXISTS credit_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT UNIQUE,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_accounts_user_id ON credit_accounts(user_id);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id BIGSERIAL PRIMARY KEY,
    credit_account_id BIGINT NOT NULL REFERENCES credit_accounts(id) ON DELETE RESTRICT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('EARN_REFERRAL', 'EARN_VERIFY', 'CONSUME_SUB', 'REFUND_SUB', 'EXPIRED')),
    amount DECIMAL(12, 2) NOT NULL,
    reference_id VARCHAR(100),
    details JSONB,
    sm3_hash VARCHAR(128) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE' CHECK (status IN ('AVAILABLE', 'PENDING', 'FROZEN', 'CONSUMED', 'EXPIRED', 'REJECTED')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_transactions_account_id ON credit_transactions(credit_account_id);
CREATE INDEX idx_credit_transactions_reference_id ON credit_transactions(reference_id);
CREATE INDEX idx_credit_transactions_type ON credit_transactions(type);

CREATE TABLE IF NOT EXISTS referral_relations (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    referee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    referee_subscription_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(referee_id)
);

CREATE INDEX idx_referral_relations_referrer_id ON referral_relations(referrer_id);
CREATE INDEX idx_referral_relations_referee_id ON referral_relations(referee_id);

-- +goose Down
DROP TABLE IF EXISTS referral_relations;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS credit_accounts;
DROP TABLE IF EXISTS entitlements;
DROP TABLE IF EXISTS subscriptions;
ALTER TABLE users DROP COLUMN IF EXISTS identity_tier;
ALTER TABLE users DROP COLUMN IF EXISTS status;
```

- [ ] **Step 2: 更新 go.work 加入 push-notification-service**

```
go 1.21

use (
    ./account-service
    ./auth-service
    ./device-fingerprint-service
    ./sms-email-service
    ./email-service
    ./session-service
    ./audit-log-service
    ./kyb-service
    ./risk-detection-service
    ./api-gateway
    ./push-notification-service
)
```

- [ ] **Step 3: 复制迁移文件到 migrations/ 目录**

将 `db-migrations/002_commercialization_schema.sql` 同步复制到 `migrations/002_commercialization_schema.sql`。

- [ ] **Step 4: 提交**

```bash
git add db-migrations/002_commercialization_schema.sql migrations/002_commercialization_schema.sql go.work
git commit -m "feat: add commercialization database schema and update go.work"
```

---

### 里程碑 1.2：entitlement-service（权益服务）（4天）

**目标：** 实现五级身份阶梯管理、权益配额缓存（Redis Hash）、Lua 脚本原子扣减、订阅生命周期管理

**Files:**
- Create: `entitlement-service/cmd/main.go`
- Create: `entitlement-service/internal/model/entitlement.go`
- Create: `entitlement-service/internal/model/subscription.go`
- Create: `entitlement-service/internal/repository/entitlement_repository.go`
- Create: `entitlement-service/internal/repository/subscription_repository.go`
- Create: `entitlement-service/internal/service/entitlement_service.go`
- Create: `entitlement-service/internal/service/subscription_service.go`
- Create: `entitlement-service/internal/handler/entitlement_handler.go`
- Create: `entitlement-service/internal/handler/subscription_handler.go`
- Create: `entitlement-service/internal/cache/entitlement_cache.go`
- Create: `entitlement-service/Dockerfile`
- Create: `entitlement-service/go.mod`

**核心 API:**
- `GET /api/v1/entitlements/:user_id` — 查询用户所有权益及配额
- `POST /internal/v1/entitlements/consume` — 内部扣减用户配额（高并发，Lua 脚本）
- `POST /internal/v1/entitlements/grant` — 内部授予用户权益/配额
- `POST /api/v1/subscriptions/purchase` — 购买订阅
- `POST /api/v1/subscriptions/upgrade` — 升级订阅
- `POST /api/v1/subscriptions/renew` — 续费订阅
- `GET /api/v1/subscriptions/:user_id` — 查询用户订阅状态
- `POST /api/v1/subscriptions/check-expiry` — 检查到期订阅（定时任务调用）
- `ANY /health`

**核心逻辑:**
1. **权益预热**：用户登录或订阅变更时，异步将 `entitlements` 表数据加载到 Redis Hash `entitlement:{user_id}`
2. **Lua 原子扣减**：接收其他微服务的核销请求，执行 Redis Lua 脚本校验余额→扣减→返回结果
3. **异步落库**：扣减成功后通过 Redis Streams 发送 `QuotaConsumedEvent`，消费端批量更新 PostgreSQL
4. **配额重置**：Asynq 定时任务根据 `reset_time` 重置配额
5. **订阅到期**：Asynq 定时扫描到期订阅，自动将用户等级回退至 L1/L0
6. **等级联动**：订阅状态变化时，更新 `users.identity_tier`

**依赖:** PostgreSQL, Redis, Asynq

- [ ] **Step 1: 创建 entitlement-service 骨架和 go.mod**

```go
module github.com/trigold786/92-Account-Center/entitlement-service

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.5.1
    github.com/redis/go-redis/v9 v9.0.0
    github.com/hibiken/asynq v0.24.1
)
```

- [ ] **Step 2: 编写 model — entitlement.go 和 subscription.go**

定义 `Entitlement`, `EntitlementQuota`, `Subscription`, `SubscriptionPurchaseRequest`, `SubscriptionUpgradeRequest`, `SubscriptionRenewRequest`, `ConsumeRequest`, `ConsumeResponse`, `GrantRequest` 等结构体。

- [ ] **Step 3: 编写 repository — entitlement_repository.go 和 subscription_repository.go**

实现 PostgreSQL CRUD：`CreateEntitlement`, `GetByUserID`, `GetByUserAndFeature`, `UpdateQuota`, `BatchUpdateUsed` 等方法。以及 `CreateSubscription`, `GetActiveByUserID`, `UpdateStatus`, `GetExpiredSubscriptions` 等方法。

- [ ] **Step 4: 编写 cache — entitlement_cache.go**

实现 Redis 缓存层：
- `WarmCache(userID)` — 从 DB 加载到 Redis Hash
- `GetQuota(userID, featureCode)` — 从 Redis 读取
- `ConsumeQuota(userID, featureCode, amount)` — Lua 脚本原子扣减
- `GrantQuota(userID, featureCode, total)` — 设置配额
- `InvalidateCache(userID)` — 失效缓存

Lua 脚本：
```lua
local key = KEYS[1]
local field = ARGV[1]
local amount = tonumber(ARGV[2])
local data = redis.call('HGET', key, field)
if data then
    local obj = cjson.decode(data)
    if obj.total - obj.used >= amount then
        obj.used = obj.used + amount
        redis.call('HSET', key, field, cjson.encode(obj))
        return 1
    else
        return 0
    end
else
    return -1
end
```

- [ ] **Step 5: 编写 service — entitlement_service.go**

实现 `EntitlementService` 接口：
- `GetUserEntitlements(userID)` — 优先读 Redis，miss 时从 DB 加载并预热
- `ConsumeQuota(userID, featureCode, amount)` — 调用 cache 层 Lua 扣减，成功后发 MQ 事件
- `GrantEntitlements(userID, tierLevel)` — 根据等级模板创建权益记录，预热缓存

- [ ] **Step 6: 编写 service — subscription_service.go**

实现 `SubscriptionService` 接口：
- `PurchaseSubscription(userID, tierLevel, price, paymentMethod)` — 创建订阅、更新用户等级、发放权益
- `UpgradeSubscription(userID, newTier, priceDiff)` — 升级（补差价）
- `RenewSubscription(userID)` — 续费
- `CheckExpired()` — 定时检查到期订阅，回退等级
- `GetUserSubscription(userID)` — 查询当前活跃订阅

每次订阅变更后：
1. 更新 `users.identity_tier`
2. 调用 `GrantEntitlements` 更新权益
3. 通过 Redis Streams 发送 `SubscriptionPaidEvent`

- [ ] **Step 7: 编写 handler — entitlement_handler.go 和 subscription_handler.go**

注册 Gin 路由，实现 HTTP 处理逻辑，统一响应格式 `{code, message, data}`。

- [ ] **Step 8: 编写 cmd/main.go**

初始化 PostgreSQL、Redis、Asynq 连接，注册路由，优雅关闭。

- [ ] **Step 9: 编写 Dockerfile**

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /entitlement-service cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /entitlement-service /usr/local/bin/
EXPOSE 30311
CMD ["entitlement-service"]
```

- [ ] **Step 10: 更新 docker-compose.yml 和 api-gateway**

在 docker-compose.yml 添加 entitlement-service（端口 30311），更新 api-gateway 添加反向代理路由 `ANY /api/v1/entitlements/*path` 和 `ANY /api/v1/subscriptions/*path`，更新 api-gateway 的 `ENTITLEMENT_SERVICE_URL` 环境变量。

- [ ] **Step 11: 更新 go.work**

添加 `./entitlement-service` 到 use 列表。

- [ ] **Step 12: 编译验证**

```powershell
$env:GOWORK="off"
cd entitlement-service
go build ./...
go vet ./...
```

- [ ] **Step 13: 提交**

```bash
git add entitlement-service/ docker-compose.yml go.work api-gateway/
git commit -m "feat: add entitlement-service with tiered identity, quota management and subscription lifecycle"
```

---

### 里程碑 1.3：credit-service（积分服务）（4天）

**目标：** 实现奖励积分复式记账、SM3 防篡改摘要链、积分发放/扣减/过期 API

**Files:**
- Create: `credit-service/cmd/main.go`
- Create: `credit-service/internal/model/credit.go`
- Create: `credit-service/internal/repository/credit_repository.go`
- Create: `credit-service/internal/service/credit_service.go`
- Create: `credit-service/internal/handler/credit_handler.go`
- Create: `credit-service/Dockerfile`
- Create: `credit-service/go.mod`

**核心 API:**
- `GET /api/v1/credits/:user_id/account` — 查询积分账户余额
- `GET /api/v1/credits/:user_id/transactions` — 查询积分交易流水
- `POST /internal/v1/credits/earn` — 内部发放奖励积分
- `POST /internal/v1/credits/consume` — 内部扣减积分（订阅抵扣）
- `POST /internal/v1/credits/refund` — 内部退回积分（Saga 补偿）
- `POST /api/v1/credits/calculate-discount` — 计算订阅可抵扣积分
- `ANY /health`

**核心逻辑:**
1. **复式记账**：任何积分变动必须生成 `credit_transactions` 流水
2. **SM3 摘要链**：每条流水落库前，将 `(id, credit_account_id, type, amount, reference_id, created_at, prev_hash)` 拼接后 SM3 哈希，存入 `sm3_hash` 字段
3. **乐观锁**：`credit_accounts` 余额更新使用 CAS（`UPDATE ... SET balance = balance + ?, updated_at = NOW() WHERE id = ? AND balance + ? >= 0`）
4. **幂等性**：使用 `reference_id` 唯一索引防重
5. **Saga 补偿**：订阅失败时调用 `/internal/v1/credits/refund` 退回已扣积分
6. **定时过期**：Asynq 任务扫描过期积分（`status=PENDING` 超过 T+7 天且无异常则转 `AVAILABLE`）

- [ ] **Step 1: 创建 credit-service 骨架和 go.mod**

- [ ] **Step 2: 编写 model — credit.go**

定义 `CreditAccount`, `CreditTransaction`, `EarnRequest`, `ConsumeRequest`, `RefundRequest`, `CalculateDiscountRequest`, `CalculateDiscountResponse`, `AccountResponse`, `TransactionListResponse` 等结构体。

- [ ] **Step 3: 编写 repository — credit_repository.go**

实现：
- `CreateAccount(userID)` — 创建积分账户
- `GetAccountByUserID(userID)` — 查询账户
- `GetAccountByID(id)` — 内部查询
- `UpdateBalance(id, delta)` — CAS 乐观锁更新余额
- `CreateTransaction(txn)` — 插入流水（含 SM3 哈希计算）
- `GetTransactionsByAccountID(accountID, offset, limit)` — 分页查询流水
- `GetLastTransaction(accountID)` — 获取最后一条流水（用于计算 SM3 链）
- `UpdateTransactionStatus(id, status)` — 更新流水状态
- `GetExpiredPendingTransactions()` — 查询待过期流水

- [ ] **Step 4: 编写 service — credit_service.go**

实现 `CreditService` 接口：
- `GetAccount(userID)` — 查询余额
- `GetTransactions(userID, page, pageSize)` — 分页查询流水
- `EarnCredits(userID, amount, txnType, referenceID, details)` — 发放积分（事务：更新余额 + 插入流水）
- `ConsumeCredits(userID, amount, referenceID, details)` — 扣减积分（校验余额→扣减→插入流水）
- `RefundCredits(userID, amount, referenceID, details)` — 退回积分（Saga 补偿）
- `CalculateDiscount(userID, subscriptionPrice)` — 计算可抵扣积分
- `VerifyIntegrity()` — 校验 SM3 摘要链完整性
- `ProcessExpiredCredits()` — 处理过期积分

SM3 摘要链算法：
```go
func (s *creditService) computeSM3Hash(txn *CreditTransaction, prevHash string) string {
    data := fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s",
        txn.ID, txn.CreditAccountID, txn.Type,
        txn.Amount.String(), txn.ReferenceID,
        txn.CreatedAt.Format(time.RFC3339Nano), prevHash)
    return crypto.SM3Hash(data)
}
```

- [ ] **Step 5: 编写 handler — credit_handler.go**

注册 Gin 路由，实现 HTTP 处理。

- [ ] **Step 6: 编写 cmd/main.go, Dockerfile**

初始化 PostgreSQL、Redis、Asynq，注册路由，优雅关闭。

- [ ] **Step 7: 更新 docker-compose.yml 和 api-gateway**

添加 credit-service（端口 30312），更新 api-gateway 路由 `ANY /api/v1/credits/*path`。

- [ ] **Step 8: 编译验证**

```powershell
$env:GOWORK="off"
cd credit-service
go build ./...
go vet ./...
```

- [ ] **Step 9: 提交**

```bash
git add credit-service/ docker-compose.yml go.work api-gateway/
git commit -m "feat: add credit-service with double-entry ledger, SM3 hash chain and quota management"
```

---

### 里程碑 1.4：account-service 扩展 — 身份等级联动（2天）

**目标：** 扩展 account-service 支持身份等级查询与更新，注册时建立推广关系

**Files:**
- Modify: `account-service/cmd/main.go`
- Modify: `account-service/internal/model/user.go`
- Modify: `account-service/internal/service/user_service.go`
- Modify: `account-service/internal/repository/user_repository.go`
- Modify: `account-service/internal/handler/register_handler.go`
- Create: `account-service/internal/handler/tier_handler.go`

**新增/修改 API:**
- `GET /api/v1/account/:user_id/tier` — 查询用户等级
- `PUT /internal/v1/account/:user_id/tier` — 内部更新用户等级（供 entitlement-service 调用）
- `GET /api/v1/account/:user_id/profile` — 查询用户完整资料（含等级）

**核心逻辑:**
1. User model 新增 `IdentityTier int` 和 `Status string` 字段
2. 注册时支持 `referral_code` 参数，建立推广关系
3. 实名认证完成后触发积分赠送（调用 credit-service 内部 API）

- [ ] **Step 1: 修改 model/user.go**

在 `User` 结构体中添加 `IdentityTier int` 和 `Status string` 字段。

- [ ] **Step 2: 修改 repository/user_repository.go**

在 Create/Get 查询中包含 `identity_tier` 和 `status` 字段，新增 `UpdateIdentityTier(userID, tier)` 方法。

- [ ] **Step 3: 修改 service/user_service.go**

注册流程扩展：支持 `referral_code` 参数，注册成功后若有推广码，调用 referral 绑定（Sprint 2 实现，此处预留接口）。

- [ ] **Step 4: 创建 handler/tier_handler.go**

实现 `GetUserTier` 和 `UpdateUserTier` 处理函数。

- [ ] **Step 5: 更新 cmd/main.go 注册新路由**

- [ ] **Step 6: 编译验证**

```powershell
$env:GOWORK="off"
cd account-service
go build ./...
go vet ./...
```

- [ ] **Step 7: 提交**

```bash
git add account-service/
git commit -m "feat: extend account-service with identity tier management and profile endpoints"
```

---

### 里程碑 1.5：Sprint 1 集成测试与部署（1天）

**目标：** 验证所有新增服务的端到端功能

- [ ] **Step 1: 更新 docs/integration_test.sh**

新增测试场景：
- 创建用户 → 查询身份等级（应为 L0）
- 购买订阅 → 查询身份等级（应为 L2）
- 查询权益配额
- 扣减配额 → 验证剩余
- 订阅到期 → 等级回退
- 发放积分 → 查询余额
- 扣减积分 → 查询流水
- SM3 完整性校验

- [ ] **Step 2: 本地 Docker Compose 启动验证**

```bash
docker compose up -d
docker compose ps  # 所有服务 healthy
```

- [ ] **Step 3: 运行集成测试**

```bash
bash docs/integration_test.sh
```

- [ ] **Step 4: 交叉编译并部署到 ECS**

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED=0; $env:GOWORK="off"
# 编译 entitlement-service, credit-service
# SCP 二进制和 docker-compose.yml 到 ECS
# 在 ECS 上重启服务
```

- [ ] **Step 5: 推送到 GitHub**

```bash
git push origin main
```

---

## 四、Sprint 2：推广返利与风控强化（T+4周 → T+6周）

### 里程碑 2.1：referral-service（推广服务）（3天）

**目标：** 实现推广关系绑定、推广链接生成、推广收益汇总查询

**Files:**
- Create: `referral-service/cmd/main.go`
- Create: `referral-service/internal/model/referral.go`
- Create: `referral-service/internal/repository/referral_repository.go`
- Create: `referral-service/internal/service/referral_service.go`
- Create: `referral-service/internal/handler/referral_handler.go`
- Create: `referral-service/Dockerfile`
- Create: `referral-service/go.mod`

**核心 API:**
- `POST /api/v1/referral/bind` — 绑定推广关系（注册时调用）
- `POST /api/v1/referral/generate-link` — 生成专属推广链接
- `GET /api/v1/referral/:user_id/summary` — 查询推广收益汇总
- `GET /api/v1/referral/:user_id/referees` — 查询被推广人列表
- `ANY /health`

**核心逻辑:**
1. 推广关系唯一性：`referee_id` 唯一索引保证一个用户只能被一个推广者邀请
2. 推广码生成：`SM3(user_id + salt)` 截取前 8 位作为推广码
3. 防自推：推广人和被推广人不能为同一实名主体
4. 实名触发：被推广用户完成实名后，调用 credit-service 发放 10 奖励积分

- [ ] **Step 1-7: 同 Sprint 1 模式创建 referral-service（端口 30313）**

- [ ] **Step 8: 更新 docker-compose.yml 和 api-gateway 路由**

- [ ] **Step 9: 编译验证并提交**

```bash
git add referral-service/ docker-compose.yml go.work api-gateway/
git commit -m "feat: add referral-service with referral binding, link generation and summary"
```

---

### 里程碑 2.2：阶梯退坡返利引擎（3天）

**目标：** 在 credit-service 中实现事件驱动的阶梯退坡返利计算

**Files:**
- Modify: `credit-service/internal/service/credit_service.go`
- Modify: `credit-service/internal/handler/credit_handler.go`
- Modify: `credit-service/cmd/main.go`
- Create: `credit-service/internal/consumer/subscription_event_consumer.go`

**核心逻辑:**
1. 监听 Redis Streams 的 `SubscriptionPaidEvent`（由 entitlement-service 发送）
2. 查询 `referral_relations` 获取该被推广用户的推广人和历史订阅次数
3. 匹配退坡比例：
   - N=0（首次订阅）：50%
   - 1≤N≤4：30%
   - 5≤N≤9：20%
   - N≥10：10%
4. 计算奖励积分 = 实付金额 × 比例
5. 调用 risk-detection-service 实时风控评估
6. 低风险：立即发放；中/高风险：置为 `PENDING` 待审核
7. 使用订单 ID 作为 `reference_id` 幂等防重
8. 在同一事务中：更新推广用户积分余额 + 插入流水 + 更新 `referee_subscription_count + 1`

- [ ] **Step 1: 创建 consumer/subscription_event_consumer.go**

实现 Redis Streams 消费者，监听 `subscription:paid` stream，解析事件，调用 credit-service 的返利逻辑。

- [ ] **Step 2: 在 credit_service.go 中添加 ProcessReferralReward 方法**

```go
func (s *creditService) ProcessReferralReward(ctx context.Context, refereeUserID int64, orderID string, paidAmount decimal.Decimal) error {
    relation, err := s.referralRepo.GetByRefereeID(ctx, refereeUserID)
    if err != nil {
        return nil // 无推广关系，不执行返利
    }
    
    count := relation.RefereeSubscriptionCount
    var rate decimal.Decimal
    switch {
    case count == 0:
        rate = decimal.NewFromFloat(0.50)
    case count >= 1 && count <= 4:
        rate = decimal.NewFromFloat(0.30)
    case count >= 5 && count <= 9:
        rate = decimal.NewFromFloat(0.20)
    default:
        rate = decimal.NewFromFloat(0.10)
    }
    
    rewardAmount := paidAmount.Mul(rate)
    
    // 调用风控评估
    riskLevel, err := s.riskService.AssessReferralRisk(ctx, relation.ReferrerID, refereeUserID)
    if err == nil && (riskLevel == "high" || riskLevel == "medium") {
        // 高风险：延迟发放
        return s.EarnCredits(ctx, relation.ReferrerID, rewardAmount, "EARN_REFERRAL", orderID, map[string]interface{}{"status": "PENDING"})
    }
    
    // 低风险：立即发放
    return s.EarnCredits(ctx, relation.ReferrerID, rewardAmount, "EARN_REFERRAL", orderID, nil)
}
```

- [ ] **Step 3: 更新 cmd/main.go 启动消费者**

- [ ] **Step 4: 编译验证并提交**

---

### 里程碑 2.3：风控强化（3天）

**目标：** 增强 risk-detection-service 支持推广防刷、延迟发放 T+7、黑名单

**Files:**
- Modify: `risk-detection-service/internal/service/risk_service.go`
- Modify: `risk-detection-service/internal/handler/risk_handler.go`
- Create: `risk-detection-service/internal/service/anti_fraud.go`
- Create: `risk-detection-service/internal/service/blacklist.go`
- Create: `risk-detection-service/internal/repository/blacklist_repository.go`

**新增 API:**
- `POST /internal/v1/risk/evaluate` — 综合风险评估（注册、登录、推广）
- `POST /internal/v1/risk/blacklist/add` — 添加黑名单
- `DELETE /internal/v1/risk/blacklist/:user_id` — 移除黑名单
- `GET /internal/v1/risk/blacklist/:user_id` — 查询黑名单状态
- `POST /internal/v1/risk/event` — 上报风险事件

**核心逻辑:**
1. **滑动窗口限流**：Redis INCRBY + EXPIRE，限制单 IP/单设备 1 小时内注册/实名上限 3 次
2. **推广防刷**：限制单推广链接 1 小时内被使用上限 50 次
3. **黑名单**：Redis Set `blacklist:users` + PostgreSQL 持久化，风控命中自动加入
4. **T+7 延迟发放**：Asynq 任务在大额积分发放后 T+7 天检查订单状态和用户行为
5. **规则引擎**：轻量级规则引擎（基于 Go `expr` 库），支持运营配置动态规则

- [ ] **Step 1-5: 实现并提交**

```bash
git add risk-detection-service/
git commit -m "feat: enhance risk-detection with anti-fraud, blacklist and T+7 delay"
```

---

### 里程碑 2.4：account-service 推广注册联动（2天）

**目标：** 注册时绑定推广关系，实名后触发积分赠送

**Files:**
- Modify: `account-service/internal/service/user_service.go`
- Modify: `account-service/internal/handler/register_handler.go`

**核心逻辑:**
1. 注册请求支持 `referral_code` 参数
2. 注册成功后，调用 referral-service `/api/v1/referral/bind` 绑定关系
3. 实名认证完成（identity_tier 从 L0 升为 L1）后，调用 credit-service `/internal/v1/credits/earn` 向被推广用户发放 10 奖励积分

- [ ] **Step 1-2: 实现并提交**

```bash
git add account-service/
git commit -m "feat: integrate referral binding in registration and credit reward on verification"
```

---

### 里程碑 2.5：Sprint 2 集成测试与部署（1天）

- [ ] **Step 1: 更新集成测试脚本**

新增场景：
- 注册时带推广码 → 推广关系建立
- 被推广用户实名 → 获赠 10 积分
- 被推广用户首次订阅 → 推广用户获 50% 返利
- 被推广用户第 5 次订阅 → 推广用户获 20% 返利
- 同 IP 批量注册 → 风控拦截
- 黑名单用户 → 推广资格取消

- [ ] **Step 2: 本地验证 → ECS 部署 → GitHub 推送**

---

## 五、Sprint 3：数据产品与合规完善（T+7周 → T+9周）

### 里程碑 3.1：data-product-service（数据产品服务）（4天）

**目标：** 实现 RFM 用户画像、推广防刷监控大盘 API

**Files:**
- Create: `data-product-service/cmd/main.go`
- Create: `data-product-service/internal/model/rfm.go`
- Create: `data-product-service/internal/model/fraud_monitor.go`
- Create: `data-product-service/internal/repository/data_repository.go`
- Create: `data-product-service/internal/service/rfm_service.go`
- Create: `data-product-service/internal/service/fraud_monitor_service.go`
- Create: `data-product-service/internal/handler/data_handler.go`
- Create: `data-product-service/Dockerfile`
- Create: `data-product-service/go.mod`

**核心 API:**
- `GET /internal/v1/data/users/rfm` — RFM 用户画像（去标识化）
- `GET /internal/v1/data/referral/fraud-monitor` — 推广防刷监控大盘（去标识化）
- `GET /internal/v1/data/dashboard/subscription-funnel` — 订阅转化漏斗
- `GET /internal/v1/data/dashboard/referral-stats` — 推广统计
- `ANY /health`

**RFM 模型:**
- **R (Recency)**：距上次活跃天数
- **F (Frequency)**：30 天内活跃天数
- **M (Monetary)**：累计订阅金额 + 积分贡献

**去标识化:** `user_id` → `SHA256(user_id + salt)` 转为匿名 ID

- [ ] **Step 1-7: 同 Sprint 模式创建 data-product-service（端口 30314）**

---

### 里程碑 3.2：API Gateway 动态脱敏（2天）

**目标：** 在 api-gateway 实现脱敏拦截器，根据请求者角色动态掩码敏感字段

**Files:**
- Modify: `api-gateway/cmd/main.go`
- Create: `api-gateway/internal/middleware/desensitization.go`

**核心逻辑:**
1. 响应拦截：在反向代理后，对响应 JSON 进行字段级脱敏
2. 手机号：`138****1234`
3. 邮箱：`t***@example.com`
4. 真实姓名：`张*`
5. 身份证号：`310***********1234`
6. 根据请求 JWT 中的角色决定是否脱敏（管理员不脱敏，普通用户脱敏）

- [ ] **Step 1-2: 实现并提交**

---

### 里程碑 3.3：device-fingerprint-service 持久化（1天）

**目标：** 将设备指纹数据从内存迁移到 PostgreSQL

**Files:**
- Modify: `device-fingerprint-service/cmd/main.go`
- Create: `device-fingerprint-service/internal/repository/device_repository.go`
- Create: `db-migrations/003_device_fingerprint_table.sql`

- [ ] **Step 1: 创建迁移脚本**

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS device_fingerprints (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(100) NOT NULL,
    fingerprint_hash VARCHAR(128) NOT NULL,
    device_info TEXT,
    ip_address VARCHAR(45),
    last_login_at TIMESTAMP WITH TIME ZONE,
    trusted_until TIMESTAMP WITH TIME ZONE,
    is_trusted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(fingerprint_hash)
);

CREATE INDEX idx_device_fingerprints_user_id ON device_fingerprints(user_id);

-- +goose Down
DROP TABLE IF EXISTS device_fingerprints;
```

- [ ] **Step 2: 实现 device_repository.go**

- [ ] **Step 3: 更新 main.go 注入 repository（替换 nil）**

- [ ] **Step 4: 编译验证并提交**

---

### 里程碑 3.4：qrcode-service（扫码登录服务）（3天）

**目标：** 实现扫码登录功能（移动端适配中优先级需求）

**Files:**
- Create: `qrcode-service/cmd/main.go`
- Create: `qrcode-service/internal/model/qrcode.go`
- Create: `qrcode-service/internal/service/qrcode_service.go`
- Create: `qrcode-service/internal/handler/qrcode_handler.go`
- Create: `qrcode-service/Dockerfile`
- Create: `qrcode-service/go.mod`

**核心 API:**
- `POST /api/v1/qrcode/generate` — 生成登录二维码（返回 QR code ID + 图片/URL）
- `GET /api/v1/qrcode/:code_id/status` — 轮询扫码状态（pending/scanned/confirmed/expired）
- `POST /api/v1/qrcode/:code_id/scan` — 移动端扫码确认
- `POST /api/v1/qrcode/:code_id/confirm` — 移动端确认登录
- `ANY /health`

**核心逻辑:**
1. Redis 存储二维码状态：`qrcode:{code_id}` → `{status, user_id, created_at}`，TTL 5 分钟
2. Web 端轮询状态，移动端扫码确认后，Web 端获取 JWT Token
3. 集成 auth-service 的 JWT 生成

- [ ] **Step 1-5: 实现并提交**

---

### 里程碑 3.5：监控与告警完善（2天）

**目标：** 接入 VictoriaMetrics 指标采集，配置 Grafana 仪表盘和告警规则

**Files:**
- Modify: 所有服务的 `cmd/main.go`（添加 `/metrics` 端点暴露给 VictoriaMetrics）
- Create: `docs/monitoring/grafana-dashboard.json`
- Create: `docs/monitoring/alert-rules.yml`

**核心指标:**
- 系统指标：CPU、内存、Goroutine 数量
- 应用指标：QPS、P99 延迟、错误率
- 业务指标：注册成功率、订阅转化率、积分发放量、风控拦截次数

- [ ] **Step 1-3: 实现并提交**

---

### 里程碑 3.6：Sprint 3 集成测试与最终部署（1天）

- [ ] **Step 1: 更新集成测试脚本**

新增场景：
- RFM 画像查询（验证去标识化）
- 推广防刷监控查询
- 动态脱敏验证（普通用户看到掩码，管理员看到明文）
- 扫码登录完整流程
- VictoriaMetrics 指标采集验证

- [ ] **Step 2: 全量回归测试**

- [ ] **Step 3: ECS 全量部署（15 个服务）**

- [ ] **Step 4: 更新项目报告**

- [ ] **Step 5: GitHub 最终推送**

---

## 六、服务端口总览（Sprint 完成后）

| 服务 | 端口 | Sprint |
|------|------|--------|
| api-gateway | 30300 | 已有 |
| account-service | 30301 | 已有（Sprint 1 扩展） |
| auth-service | 30302 | 已有 |
| sms-email-service | 30303 | 已有 |
| kyb-service | 30304 | 已有 |
| audit-log-service | 30305 | 已有 |
| risk-detection-service | 30306 | 已有（Sprint 2 强化） |
| session-service | 30307 | 已有 |
| email-service | 30308 | 已有 |
| device-fingerprint-service | 30309 | 已有（Sprint 3 持久化） |
| push-notification-service | 30310 | 已有 |
| **entitlement-service** | **30311** | **Sprint 1 新增** |
| **credit-service** | **30312** | **Sprint 1 新增** |
| **referral-service** | **30313** | **Sprint 2 新增** |
| **data-product-service** | **30314** | **Sprint 3 新增** |
| **qrcode-service** | **30315** | **Sprint 3 新增** |

---

## 七、数据库表总览（Sprint 完成后）

| 表名 | Sprint | 说明 |
|------|--------|------|
| users | 已有 | 新增 `identity_tier`, `status` 字段 |
| enterprises | 已有 | — |
| sub_accounts | 已有 | — |
| audit_logs | 已有 | — |
| risk_events | 已有 | — |
| **subscriptions** | **Sprint 1** | 订阅生命周期管理 |
| **entitlements** | **Sprint 1** | 权益配额管理 |
| **credit_accounts** | **Sprint 1** | 奖励积分账户 |
| **credit_transactions** | **Sprint 1** | 积分交易流水（SM3 防篡改） |
| **referral_relations** | **Sprint 1** | 推广关系（Sprint 2 使用） |
| **device_fingerprints** | **Sprint 3** | 设备指纹持久化 |

---

## 八、风险与应对

| 风险 | 影响 | 应对策略 |
|------|------|---------|
| 积分账务一致性 | 高 | 复式记账 + SM3 摘要链 + 乐观锁 + 定时校验 |
| 权益超卖 | 高 | Redis Lua 原子扣减 |
| 推广薅羊毛 | 高 | 多维反作弊（IP/设备/行为） + T+7 延迟发放 + 黑名单 |
| 跨服务事务一致性 | 中 | Saga 模式（订阅失败退回积分） |
| 新增服务资源占用 | 中 | Go 低内存占用，每服务 < 50MB |
| ECS 磁盘空间（59GB） | 中 | 定期清理日志和未使用镜像 |

---

## 九、验收标准（对应 TIP §12）

1. ✅ 所有服务 Go/Gin 技术栈，Docker Compose 一键启动
2. ✅ 权益核销接口 5000 TPS 下 P99 < 15ms（Redis Lua 扣减）
3. ✅ 10 万次并发积分操作后，余额与流水总和 100% 一致
4. ✅ SM3 摘要链校验无异常
5. ✅ 15 次连续订阅准确按 50%-30%-20%-10% 返利
6. ✅ 同 IP 连续注册 10 账号，第 4 个触发拦截
7. ✅ 数据产品接口敏感字段正确脱敏
8. ✅ VictoriaMetrics 采集所有服务指标
9. ✅ 通过日志 trace_id 追踪跨服务请求
10. ✅ Docker Compose 30 秒内启动所有服务

---

## 十、依赖关系图

```
Sprint 1:
  1.1 DB Migration
    └→ 1.2 entitlement-service (依赖新表)
    └→ 1.3 credit-service (依赖新表)
    └→ 1.4 account-service 扩展 (依赖新字段)
  1.5 集成测试 (依赖 1.2, 1.3, 1.4)

Sprint 2:
  2.1 referral-service (依赖 credit-service)
  2.2 阶梯退坡引擎 (依赖 referral-service + credit-service + risk-detection)
  2.3 风控强化 (独立)
  2.4 推广注册联动 (依赖 referral-service + credit-service)
  2.5 集成测试

Sprint 3:
  3.1 data-product-service (依赖所有 Sprint 1+2 数据)
  3.2 动态脱敏 (依赖 api-gateway)
  3.3 设备指纹持久化 (独立)
  3.4 qrcode-service (依赖 auth-service)
  3.5 监控告警 (独立)
  3.6 最终集成测试与部署
```

---

## 十一、技术约束提醒

- 所有 go 命令前设置 `$env:GOWORK="off"`
- 所有 go mod 操作前设置 `$env:GOPROXY="https://goproxy.cn,direct"`
- 交叉编译: `$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED=0`
- Import 路径统一使用 `github.com/trigold786/92-Account-Center/<service>`
- ECS Docker 网络名: `w004_w004_network`
- 代码推送使用临时 deploy key (GitHub API → push → 删除)
- 监控方案: VictoriaMetrics (端口 20010)，NOT Prometheus
