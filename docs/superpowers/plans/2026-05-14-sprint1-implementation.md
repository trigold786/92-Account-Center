# 商业化迭代 Sprint 1 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 完成服务合并（11→7）+ DB 迁移 + 商业化核心功能（五级身份阶梯、权益中控、订阅管理、积分账务），确保所有服务编译通过、Docker Compose 一键启动。

**Architecture:** 7 个容器（api-gateway + auth-service + notification-service + account-service + credit-service + compliance-service + data-product-service），共享自建 PostgreSQL/Redis/MinIO 基础设施。Go/Gin + PostgreSQL + Redis + Asynq。

**Tech Stack:** Go 1.21+, Gin, PostgreSQL, Redis 7, Asynq, SM3/SM4 国密, Docker Compose

**Design Spec:** `docs/superpowers/specs/2026-05-14-commercialization-design.md`

---

## Task 1: DB Migration — 新增商业化表

**Files:**
- Create: `db-migrations/002_commercialization_schema.sql`
- Copy: `migrations/002_commercialization_schema.sql`

**Context:** 现有 5 张表（users, enterprises, sub_accounts, audit_logs, risk_events）。需新增 5 张商业化表 + 修改 users 表。所有 SQL 使用 Goose 格式（`-- +goose Up` / `-- +goose Down`）。

- [ ] **Step 1:** 创建迁移脚本 `db-migrations/002_commercialization_schema.sql`

内容包含（完整 SQL）：
- ALTER TABLE users ADD identity_tier INT DEFAULT 0, ADD status VARCHAR(20) DEFAULT 'ACTIVE'
- CREATE TABLE subscriptions (id BIGSERIAL PK, user_id BIGINT FK, tier_level INT CHECK(2,3,4), start_time TIMESTAMPTZ, end_time TIMESTAMPTZ, status VARCHAR(20), price DECIMAL(10,2), payment_method VARCHAR(50), order_id VARCHAR(100) UNIQUE, created_at/updated_at)
- CREATE TABLE entitlements (id BIGSERIAL PK, user_id BIGINT FK, feature_code VARCHAR(100), total_quota INT, used_quota INT, reset_time TIMESTAMPTZ, created_at/updated_at, UNIQUE(user_id,feature_code))
- CREATE TABLE credit_accounts (id BIGSERIAL PK, user_id BIGINT FK UNIQUE, balance DECIMAL(12,2) CHECK>=0, status VARCHAR(20), created_at/updated_at)
- CREATE TABLE credit_transactions (id BIGSERIAL PK, credit_account_id BIGINT FK, type VARCHAR(50), amount DECIMAL(12,2), reference_id VARCHAR(100), details JSONB, sm3_hash VARCHAR(128), status VARCHAR(20), created_at)
- CREATE TABLE referral_relations (id BIGSERIAL PK, referrer_id BIGINT FK, referee_id BIGINT FK UNIQUE, referee_subscription_count INT DEFAULT 0, status VARCHAR(20), created_at/updated_at)
- CREATE TABLE device_fingerprints (id BIGSERIAL PK, user_id BIGINT FK, device_id VARCHAR(100), fingerprint_hash VARCHAR(128) UNIQUE, device_info TEXT, ip_address VARCHAR(45), last_login_at TIMESTAMPTZ, trusted_until TIMESTAMPTZ, is_trusted BOOLEAN DEFAULT FALSE, created_at/updated_at)
- 所有表包含合理索引
- Down 部分逆序 DROP + ALTER DROP

- [ ] **Step 2:** 复制到 `migrations/` 目录

- [ ] **Step 3:** 提交 `git commit -m "feat: add commercialization DB migration (5 new tables + users alter)"`

---

## Task 2: 创建 notification-service（合并 sms+email+push）

**Files:**
- Create: `notification-service/cmd/main.go`
- Create: `notification-service/internal/handler/sms_handler.go`
- Create: `notification-service/internal/handler/email_handler.go`
- Create: `notification-service/internal/handler/push_handler.go`
- Create: `notification-service/internal/service/sms_service.go`
- Create: `notification-service/internal/service/email_service.go`
- Create: `notification-service/internal/service/push_service.go`
- Create: `notification-service/internal/provider/sms.go`
- Create: `notification-service/internal/provider/email.go`
- Create: `notification-service/internal/provider/push.go`
- Create: `notification-service/internal/model/sms.go`
- Create: `notification-service/internal/model/email.go`
- Create: `notification-service/internal/model/push.go`
- Create: `notification-service/pkg/circuitbreaker/circuitbreaker.go`
- Create: `notification-service/Dockerfile`
- Create: `notification-service/go.mod`

**Context:** 合并 sms-email-service、email-service、push-notification-service 三个服务为一个 notification-service。保留所有现有功能代码和 API 路由不变。Import path: `github.com/trigold786/92-Account-Center/notification-service`。端口 30311。

**Routes (全部保留原路由):**
- `POST /api/v1/sms/send` → sms_handler
- `POST /api/v1/sms/verify` → sms_handler
- `GET /api/v1/sms/providers/status` → sms_handler
- `POST /api/v1/email/otp/send` → email_handler
- `POST /api/v1/email/otp/verify` → email_handler
- `POST /api/v1/email/magic-link/send` → email_handler
- `GET /api/v1/email/magic-link/verify` → email_handler
- `POST /api/v1/email/send` → email_handler
- `POST /api/v1/push/send` → push_handler
- `POST /api/v1/push/device/register` → push_handler
- `GET /api/v1/push/user/:user_id/devices` → push_handler
- `ANY /health`

**Implementation:**
1. 从 sms-email-service 复制 SMS 相关代码（handler, service, provider, model）
2. 从 email-service 复制 Email 相关代码（handler, service, provider, model）
3. 从 push-notification-service 复制 Push 相关代码（handler, service, model）
4. 复制 circuitbreaker 包
5. 在 cmd/main.go 中统一注册所有路由、初始化 Redis 连接
6. 修复所有 import path 为 `github.com/trigold786/92-Account-Center/notification-service/...`
7. Dockerfile 遵循 SOP L5 Appendix H 模板，端口 30311
8. go.mod 依赖: gin, go-redis/v9, 其他必要的

**Verification:** `$env:GOWORK="off"; cd notification-service; go build ./...; go vet ./...`

- [ ] **Step 1-8:** 实现并验证
- [ ] **Step 9:** 提交 `git commit -m "feat: create notification-service merging sms+email+push (port 30311)"`

---

## Task 3: 创建 compliance-service（合并 risk+audit+kyb）

**Files:**
- Create: `compliance-service/cmd/main.go`
- Create: `compliance-service/internal/handler/risk_handler.go`
- Create: `compliance-service/internal/handler/audit_handler.go`
- Create: `compliance-service/internal/handler/kyb_handler.go`
- Create: `compliance-service/internal/service/risk_service.go`
- Create: `compliance-service/internal/service/audit_service.go`
- Create: `compliance-service/internal/service/kyb_service.go`
- Create: `compliance-service/internal/service/sub_account_service.go`
- Create: `compliance-service/internal/repository/risk_repository.go`
- Create: `compliance-service/internal/repository/audit_repository.go`
- Create: `compliance-service/internal/repository/enterprise_repository.go`
- Create: `compliance-service/internal/model/risk.go`
- Create: `compliance-service/internal/model/risk_request.go`
- Create: `compliance-service/internal/model/audit.go`
- Create: `compliance-service/internal/model/log_request.go`
- Create: `compliance-service/internal/model/enterprise.go`
- Create: `compliance-service/internal/model/verification.go`
- Create: `compliance-service/internal/model/sub_account.go`
- Create: `compliance-service/pkg/mq/mq.go`
- Create: `compliance-service/pkg/mq/redis_streams.go`
- Create: `compliance-service/pkg/mq/kafka.go`
- Create: `compliance-service/pkg/crypto/encryptor.go`
- Create: `compliance-service/pkg/crypto/sm4.go`
- Create: `compliance-service/pkg/crypto/sm3.go`
- Create: `compliance-service/Dockerfile`
- Create: `compliance-service/go.mod`

**Context:** 合并 risk-detection-service、audit-log-service、kyb-service。Import path: `github.com/trigold786/92-Account-Center/compliance-service`。端口 30313。

**Routes:**
- `POST /api/v1/risk/assess` → risk_handler
- `GET /api/v1/risk/history/:user_id` → risk_handler
- `GET /api/v1/risk/event/:event_id` → risk_handler
- `POST /api/v1/audit/logs` → audit_handler
- `POST /api/v1/audit/logs/batch` → audit_handler
- `GET /api/v1/audit/logs/user/:user_id` → audit_handler
- `GET /api/v1/audit/logs` → audit_handler
- `GET /api/v1/audit/logs/:log_id/verify` → audit_handler
- `POST /api/v1/audit/logs/cleanup` → audit_handler
- `POST /api/v1/kyb/submit` → kyb_handler
- `POST /api/v1/kyb/micro-payment/initiate` → kyb_handler
- `POST /api/v1/kyb/micro-payment/verify` → kyb_handler
- `POST /api/v1/kyb/face-verify` → kyb_handler
- `GET /api/v1/kyb/status/:enterprise_id` → kyb_handler
- `ANY /health`

**Implementation:**
1. 从 risk-detection-service 复制全部代码
2. 从 audit-log-service 复制全部代码（含 mq 包）
3. 从 kyb-service 复制全部代码（含 crypto 包）
4. cmd/main.go 统一注册路由，初始化 PG + Redis
5. 修复所有 import path
6. Dockerfile + go.mod

**Verification:** `$env:GOWORK="off"; cd compliance-service; go build ./...; go vet ./...`

- [ ] **Implement and verify**
- [ ] **Commit:** `git commit -m "feat: create compliance-service merging risk+audit+kyb (port 30313)"`

---

## Task 4: 扩展 auth-service（合并 session+device+新增 qrcode）

**Files:**
- Modify: `auth-service/cmd/main.go` — 添加 session/device/qrcode 路由
- Create: `auth-service/internal/handler/session_handler.go`
- Create: `auth-service/internal/handler/device_handler.go`
- Create: `auth-service/internal/handler/qrcode_handler.go`
- Create: `auth-service/internal/service/session_service.go`
- Create: `auth-service/internal/service/device_service.go`
- Create: `auth-service/internal/service/qrcode_service.go`
- Create: `auth-service/internal/repository/session_repository.go`
- Create: `auth-service/internal/repository/device_repository.go`
- Create: `auth-service/internal/model/session.go`
- Create: `auth-service/internal/model/session_request.go`
- Create: `auth-service/internal/model/device.go`
- Create: `auth-service/internal/model/qrcode.go`

**Context:** 合并 session-service 和 device-fingerprint-service 到 auth-service，新增扫码登录功能。auth-service 已有登录/JWT/生物识别功能。端口保持 30302。

**新增 Routes:**
- `POST /api/v1/session/create` → session_handler
- `POST /api/v1/session/validate` → session_handler
- `GET /api/v1/session/user/:user_id` → session_handler
- `POST /api/v1/session/invalidate` → session_handler
- `POST /api/v1/session/invalidate-all` → session_handler
- `POST /api/v1/session/refresh` → session_handler
- `POST /api/v1/device/register` → device_handler
- `POST /api/v1/device/verify` → device_handler
- `POST /api/v1/device/trust` → device_handler
- `GET /api/v1/device/user/:user_id` → device_handler
- `DELETE /api/v1/device/:device_id` → device_handler
- `POST /api/v1/qrcode/generate` → qrcode_handler (NEW)
- `GET /api/v1/qrcode/:code_id/status` → qrcode_handler (NEW)
- `POST /api/v1/qrcode/:code_id/scan` → qrcode_handler (NEW)
- `POST /api/v1/qrcode/:code_id/confirm` → qrcode_handler (NEW)

**QR Code 逻辑:**
- generate: 生成 UUID code_id, Redis `qrcode:{code_id}` → JSON `{status:pending, user_id:null, created_at}`, TTL 5min, 返回 code_id
- status: 查 Redis 返回当前状态 (pending/scanned/confirmed/expired)
- scan: 更新 status=scanned, 存扫码用户 user_id
- confirm: 更新 status=confirmed, 调用 JWT 生成, 存 token 到 `qrcode:{code_id}:token`

**Implementation:**
1. 从 session-service 复制 model/handler/service/repository 代码
2. 从 device-fingerprint-service 复制 model/handler/service 代码，新增 device_repository（连接 PG 写入 device_fingerprints 表）
3. 新增 qrcode handler/service/model
4. 更新 cmd/main.go 注册所有新路由
5. go.mod 添加必要依赖

**Verification:** `$env:GOWORK="off"; cd auth-service; go build ./...; go vet ./...`

- [ ] **Implement and verify**
- [ ] **Commit:** `git commit -m "feat: extend auth-service with session+device+qrcode (merged from 3 services)"`

---

## Task 5: 扩展 account-service（身份等级+权益+订阅）

**Files:**
- Modify: `account-service/cmd/main.go` — 添加新路由
- Modify: `account-service/internal/model/user.go` — 添加 IdentityTier, Status 字段
- Modify: `account-service/internal/repository/user_repository.go` — 更新 CRUD 含新字段
- Modify: `account-service/internal/service/user_service.go` — 注册支持 referral_code
- Create: `account-service/internal/handler/entitlement_handler.go`
- Create: `account-service/internal/handler/subscription_handler.go`
- Create: `account-service/internal/handler/tier_handler.go`
- Create: `account-service/internal/model/entitlement.go`
- Create: `account-service/internal/model/subscription.go`
- Create: `account-service/internal/repository/entitlement_repository.go`
- Create: `account-service/internal/repository/subscription_repository.go`
- Create: `account-service/internal/service/entitlement_service.go`
- Create: `account-service/internal/service/subscription_service.go`
- Create: `account-service/internal/cache/entitlement_cache.go`

**Context:** account-service 已有注册/密码/删除功能，需扩展五级身份阶梯、权益配额管理和订阅生命周期管理。端口保持 30301。

**新增 Routes:**
- `GET /api/v1/account/:user_id/tier` — 查询用户等级
- `PUT /internal/v1/account/:user_id/tier` — 内部更新等级
- `GET /api/v1/entitlements/:user_id` — 查询权益配额
- `POST /internal/v1/entitlements/consume` — 原子扣减（Redis Lua）
- `POST /internal/v1/entitlements/grant` — 授予权益
- `POST /api/v1/subscriptions/purchase` — 购买订阅
- `POST /api/v1/subscriptions/upgrade` — 升级
- `POST /api/v1/subscriptions/renew` — 续费
- `GET /api/v1/subscriptions/:user_id` — 查询订阅

**权益中控核心:**
- Redis Hash: `entitlement:{user_id}` → field=feature_code, value=JSON{total,used,reset}
- Lua 原子扣减脚本
- 订阅变更 → 更新 users.identity_tier → 刷新权益缓存 → 发 SubscriptionPaidEvent

**Verification:** `$env:GOWORK="off"; cd account-service; go build ./...; go vet ./...`

- [ ] **Implement and verify**
- [ ] **Commit:** `git commit -m "feat: extend account-service with tiered identity, entitlement engine and subscription lifecycle"`

---

## Task 6: 创建 credit-service（积分账务+推广返利）

**Files:**
- Create: `credit-service/cmd/main.go`
- Create: `credit-service/internal/handler/credit_handler.go`
- Create: `credit-service/internal/handler/referral_handler.go`
- Create: `credit-service/internal/service/credit_service.go`
- Create: `credit-service/internal/service/referral_service.go`
- Create: `credit-service/internal/repository/credit_repository.go`
- Create: `credit-service/internal/repository/referral_repository.go`
- Create: `credit-service/internal/model/credit.go`
- Create: `credit-service/internal/model/referral.go`
- Create: `credit-service/pkg/crypto/sm3.go`
- Create: `credit-service/Dockerfile`
- Create: `credit-service/go.mod`

**Context:** 全新服务，实现奖励积分复式记账、SM3 防篡改摘要链、推广关系管理。Sprint 2 将添加阶梯退坡返利引擎。端口 30312。

**Routes:**
- `GET /api/v1/credits/:user_id/account` — 查询积分余额
- `GET /api/v1/credits/:user_id/transactions` — 查询流水
- `POST /internal/v1/credits/earn` — 发放积分
- `POST /internal/v1/credits/consume` — 扣减积分
- `POST /internal/v1/credits/refund` — 退回积分（Saga）
- `POST /api/v1/credits/calculate-discount` — 计算可抵扣
- `POST /api/v1/referral/bind` — 绑定推广关系
- `POST /api/v1/referral/generate-link` — 生成推广链接
- `GET /api/v1/referral/:user_id/summary` — 推广收益汇总
- `ANY /health`

**SM3 摘要链算法:**
```
hash = SM3(fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s", id, accountID, type, amount, refID, createdAt, prevHash))
```

**幂等:** reference_id 唯一索引 + 查重检查

**乐观锁:** `UPDATE credit_accounts SET balance = balance + ?, updated_at = NOW() WHERE id = ? AND balance + ? >= 0`

**Verification:** `$env:GOWORK="off"; cd credit-service; go build ./...; go vet ./...`

- [ ] **Implement and verify**
- [ ] **Commit:** `git commit -m "feat: create credit-service with double-entry ledger, SM3 hash chain and referral binding"`

---

## Task 7: 创建 data-product-service（骨架）

**Files:**
- Create: `data-product-service/cmd/main.go`
- Create: `data-product-service/internal/handler/data_handler.go`
- Create: `data-product-service/internal/service/data_service.go`
- Create: `data-product-service/internal/repository/data_repository.go`
- Create: `data-product-service/Dockerfile`
- Create: `data-product-service/go.mod`

**Context:** Sprint 3 将实现完整 RFM/防刷功能，Sprint 1 只创建骨架和基础 health 端点。端口 30314。

**Routes (Sprint 1 骨架):**
- `ANY /health`
- Sprint 3 将添加: `/api/v1/data/users/rfm`, `/api/v1/data/referral/fraud-monitor` 等

**Verification:** `$env:GOWORK="off"; cd data-product-service; go build ./...; go vet ./...`

- [ ] **Implement and verify**
- [ ] **Commit:** `git commit -m "feat: create data-product-service skeleton (port 30314)"`

---

## Task 8: 更新 api-gateway 路由

**Files:**
- Modify: `api-gateway/cmd/main.go`

**Context:** 合并服务后路由变化：
- `/api/v1/sms/*`, `/api/v1/email/*`, `/api/v1/push/*` → notification-service:30311
- `/api/v1/risk/*`, `/api/v1/audit/*`, `/api/v1/kyb/*` → compliance-service:30313
- `/api/v1/device/*`, `/api/v1/session/*`, `/api/v1/qrcode/*` → auth-service:30302
- `/api/v1/credits/*`, `/api/v1/referral/*` → credit-service:30312
- `/api/v1/entitlements/*`, `/api/v1/subscriptions/*` → account-service:30301
- `/api/v1/data/*` → data-product-service:30314
- 删除旧的独立服务路由配置
- 添加动态脱敏中间件骨架（Sprint 3 完善）

**Verification:** `$env:GOWORK="off"; cd api-gateway; go build ./...; go vet ./...`

- [ ] **Implement and verify**
- [ ] **Commit:** `git commit -m "feat: update api-gateway routes for merged services architecture"`

---

## Task 9: 更新 docker-compose.yml + go.work + 清理旧目录

**Files:**
- Modify: `docker-compose.yml` — 替换旧服务为合并后的 7 个服务
- Modify: `go.work` — 更新 use 列表
- Delete: `session-service/`, `device-fingerprint-service/`, `sms-email-service/`, `email-service/`, `push-notification-service/`, `risk-detection-service/`, `audit-log-service/`, `kyb-service/`

**docker-compose.yml 变更:**
- 保留: postgres, redis, minio, db-migrate
- 保留: api-gateway (30300), account-service (30301), auth-service (30302)
- 新增: notification-service (30311), credit-service (30312), compliance-service (30313), data-product-service (30314)
- 删除: sms-email-service (30303), kyb-service (30304), audit-log-service (30305), risk-detection-service (30306), session-service (30307), email-service (30308), device-fingerprint-service (30309), push-notification-service (30310)
- 所有容器添加 resource limits (CPU 0.5, memory 512M)
- 所有容器使用 alpine:3.19 基础镜像

**go.work 变更:**
```
go 1.21
use (
    ./account-service
    ./auth-service
    ./api-gateway
    ./notification-service
    ./credit-service
    ./compliance-service
    ./data-product-service
)
```

- [ ] **Implement**
- [ ] **Commit:** `git commit -m "feat: update docker-compose and go.work for 7-service architecture, remove merged services"`

---

## Task 10: 全量编译验证

**验证所有服务编译通过:**
```powershell
$env:GOWORK="off"; $env:GOPROXY="https://goproxy.cn,direct"
# 逐一验证每个服务
cd account-service; go build ./...; go vet ./...
cd ../auth-service; go build ./...; go vet ./...
cd ../notification-service; go build ./...; go vet ./...
cd ../credit-service; go build ./...; go vet ./...
cd ../compliance-service; go build ./...; go vet ./...
cd ../data-product-service; go build ./...; go vet ./...
cd ../api-gateway; go build ./...; go vet ./...
```

- [ ] **Verify all 7 services compile and pass vet**

---

## Task 11: Docker Compose 启动验证

```bash
docker compose up -d
docker compose ps  # 所有 10 个容器 (3 infra + 7 services) healthy
curl http://localhost:30300/health
curl http://localhost:30301/health
curl http://localhost:30302/health
curl http://localhost:30311/health
curl http://localhost:30312/health
curl http://localhost:30313/health
curl http://localhost:30314/health
```

- [ ] **Verify all services start and health check passes**

---

## Task 12: 集成测试更新

**Files:**
- Modify: `docs/integration_test.sh` — 更新端口和服务映射

新增测试场景：
- 用户注册 → 查询等级 L0
- 购买订阅 → 查询等级 L2
- 查询权益配额
- 积分发放 → 查询余额
- 积分扣减 → 查询流水
- 推广关系绑定
- 所有合并服务的原有 API 验证

- [ ] **Update and run integration tests**
- [ ] **Commit:** `git commit -m "test: update integration tests for merged service architecture"`
