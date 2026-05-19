# Account Center V2.0 系统规格文档

> **文档类型**: 系统规格文档（正式版）
> **版本**: V2.0.0
> **日期**: 2026-05-19
> **状态**: 已完成
> **基线版本**: SSD V1.3.1
> **评估基准**: ITERATION_AccountCenter_V2.0_V1.0.1（77 项改进建议）
> **变更历史**:
> | 版本 | 日期 | 变更内容 | 作者 |
> |------|------|---------|------|
> | V2.0.0 | 2026-05-19 | 初始编制，覆盖全部技术架构变更 | |

---

## 目录

1. 系统概述
2. 系统架构
3. 服务详细设计
   - 3.1 Phase 6 — P0 技术设计
   - 3.2 Phase 7 — P1 技术设计
   - 3.3 Phase 8 — P2 技术设计
   - 3.4 Phase 9 — P3 技术设计
4. 数据设计
5. API 设计
6. 安全设计
7. 部署设计
8. 可观测性设计
9. 附录

---

## 1. 系统概述

### 1.1 系统范围与边界

Account Center V2.0 包含 9 个 Go 微服务（V1.3.1 的 8 个 + 新增 payment-service）、4 端前端（Web/微信小程序/iOS/Android）、基础设施层（PostgreSQL/Redis/Kafka/监控栈）。

系统边界：
- **包含**：统一认证、五级身份体系、积分经济、推荐返利、订阅生命周期、支付闭环、推送通知、广告变现、管理后台
- **不包含**：第三方支付网关内部逻辑、广告平台内部逻辑、移动端操作系统层

### 1.2 V1.3.1 架构基线

基于 SSD V1.3.1，当前架构：
- 8 个 Go 微服务：api-gateway(30300), account-service(30301), auth-service(30302), notification-service(30311), credit-service(30312), compliance-service(30313), data-product-service(30314), config-service(30315)
- 同步 HTTP 通信（部分 Redis Streams 异步）
- Docker Compose 单机部署
- 基础 Prometheus 指标 + VictoriaMetrics
- 配置中心（config-service，启动时单次加载）

### 1.3 V2.0 架构演进目标

从 V1.3.1 到 V2.0 的关键架构演进：
- **新增 payment-service**：独立支付微服务，订单管理 + 支付网关 + 回调 + 对账
- **可观测性**：OpenTelemetry 分布式追踪 + Grafana Dashboard + AlertManager 告警
- **可靠性**：网关超时 + 服务间熔断器 + 真实健康检查 + Saga 分布式事务
- **安全**：argon2id 密码哈希 + KMS 密钥管理 + API 安全加固
- **部署**：K8s Helm Chart + CI/CD 流水线 + 金丝雀发布
- **移动端**：APNs/FCM 推送 + 广告 SDK + 深度链接

## 2. 系统架构

### 2.1 整体架构图

```
                    ┌─────────────────────────────────────┐
                    │           客户端层                    │
                    │  Web(Vue3) | 小程序 | iOS | Android  │
                    └─────────────┬───────────────────────┘
                                  │ HTTPS
                    ┌─────────────▼───────────────────────┐
                    │         API Gateway (30300)           │
                    │  JWT / 限流 / CORS / 脱敏 / 超时       │
                    └─────────────┬───────────────────────┘
                                  │ HTTP (服务间)
          ┌───────────┬───────────┼───────────┬──────────────┐
          ▼           ▼           ▼           ▼              ▼
   ┌────────────┐ ┌────────┐ ┌────────┐ ┌──────────┐ ┌───────────┐
   │ account    │ │ auth   │ │ credit │ │compliance│ │notification│
   │ (30301)    │ │(30302) │ │(30312) │ │(30313)   │ │(30311)     │
   └─────┬──────┘ └───┬────┘ └───┬────┘ └────┬─────┘ └─────┬─────┘
         │            │          │           │             │
         │       ┌────▼────┐    │      ┌────▼────┐  ┌─────▼─────┐
         │       │data-    │    │      │config   │  │payment    │
         │       │product  │    │      │(30315)  │  │(NEW)      │
         │       │(30314)  │    │      └─────────┘  └───────────┘
         │       └─────────┘    │
         └──────────┬───────────┘
                    │
          ┌─────────▼──────────┐
          │    基础设施层        │
          │ PG 18 │ Redis 8.2  │
          │ Kafka 4.2 │ MinIO  │
          │ VM+Loki+Grafana    │
          └────────────────────┘
```

### 2.2 服务清单与职责

| 服务 | 端口 | V1.3.1 职责 | V2.0 变更 |
|------|------|-------------|---------|
| api-gateway | 30300 | JWT/限流/CORS/脱敏/代理 | 新增超时中间件、熔断器、拆分代码 |
| account-service | 30301 | 用户 CRUD、身份等级、注销 | 新增 deletion-worker、Admin API |
| auth-service | 30302 | 认证、Token、生物识别 | argon2id 迁移、OAuth 扩展 |
| notification-service | 30311 | SMS/Email、设备注册 | 新增 APNs/FCM provider |
| credit-service | 30312 | 积分 CRUD、消费、RFM | 无重大变更 |
| compliance-service | 30313 | 审计日志、PII 脱敏 | 无重大变更 |
| data-product-service | 30314 | RFM 引擎、漏斗、概览 | 新增实时行为流 |
| config-service | 30315 | 配置 CRUD、发布审批 | 无重大变更 |
| **payment-service** | **30316** | — | **V2.0 新增**：订单+支付+回调+对账 |

### 2.3 通信模式

| 模式 | 当前 | V2.0 目标 |
|------|------|----------|
| 服务间同步 | HTTP 直连 | HTTP + 熔断器 + 超时 |
| 服务间异步 | 部分使用 Redis Streams | 全面推广 Redis Streams，Prod 切换 Kafka |
| 分布式事务 | 无 | Saga 编排器（积分消费+订阅购买） |
| 事件溯源 | 无 | Redis Streams/Kafka 事件日志 |

### 2.4 技术选型版本矩阵

引用 ITERATION V1.0.1 第 5.8.2 节统一版本基准表（完整版本矩阵见 PRD V2.0.0 附录）。

| 技术 | 版本 |
|------|------|
| Go | 1.26.x |
| Gin | v1.12.x |
| PostgreSQL | 18.x |
| Redis | 8.2.x (LTS) |
| Kafka | 4.2.x |
| VictoriaMetrics | v1.143.x |
| Loki | 3.7.x |
| Grafana | 13.x |
| Docker Engine | 29.x |
| Docker Compose | V5.x |

## 3. 服务详细设计

### 3.0 概述

本节按迭代 Phase 分组，为 PRD V2.0.0 中定义的每项需求提供完整技术设计。每项设计包含：需求 ID 关联、技术方案（架构/流程）、关键代码路径、数据库变更、API 变更、测试策略。

---

### 3.1 Phase 6 — P0 技术设计（14 项）

> 以下 14 项为系统上线前必须完成的阻塞项的技术设计方案。

---

#### 3.1.1 NF-01: 账号注销 Worker

**需求关联**: PRD 3.1 NF-01

**技术方案**:

采用 Asynq 定时任务框架实现 deletion-worker，作为 account-service 的后台任务处理器。

```
┌──────────────┐    提交注销     ┌───────────────────┐
│  用户请求     │ ──────────────→ │  account-service  │
│  DELETE /me   │                │  写入 deletion     │
└──────────────┘                │  request 记录      │
                                │  scheduled_at =     │
                                │  now + 7 days       │
                                └────────┬───────────┘
                                         │ Asynq Schedule
                                         ▼
                                ┌────────────────────┐
                                │  deletion-worker   │
                                │  (Asynq Task)      │
                                │                    │
                                │  1. 检查冻结期     │
                                │  2. 检查是否撤回   │
                                │  3. 匿名化 PII     │
                                │  4. 清理 Redis     │
                                │  5. 写审计日志     │
                                │  6. 标记完成       │
                                └────────────────────┘
```

**关键代码路径**:
- `account-service/internal/worker/deletion.go` — Asynq task handler，实现 `HandleDeletionTask` 函数
- `account-service/internal/repository/user.go` — 新增 `AnonymizeUser(ctx, userID)` 方法
- `account-service/internal/service/deletion.go` — 注销业务逻辑编排

**数据匿名化规则**:
- `phone` → `deleted_{userID}_{timestamp}`
- `email` → `deleted_{userID}_{timestamp}@anon.invalid`
- `real_name` → `deleted`
- `id_card_number` → 置空（PIPL 要求彻底删除）
- `avatar_url` → 置空
- `DeletionDeletedAt` → 写入 `time.Now().UTC()`
- 保留字段：`id`, `created_at`（审计需要）、`deletion_requested_at`, `deletion_deleted_at`

**Redis 清理策略**:
- 删除 `session:{userID}` — 用户会话
- 删除 `cache:user:{userID}` — 用户缓存
- 删除 `cache:user_level:{userID}` — 等级缓存
- 删除 `cache:subscription:{userID}` — 订阅缓存

**数据库变更**:
- 无新表，使用现有 `users` 表的 `deletion_*` 字段
- 新增 Goose migration 添加索引：`CREATE INDEX idx_users_deletion_scheduled ON users(deletion_scheduled_at) WHERE deletion_deleted_at IS NULL`

**测试策略**:
- 单元测试：mock repository，验证匿名化字段替换逻辑、Redis 清理调用
- 集成测试：使用 testcontainers 启动 Redis + PG，完整执行冻结期→匿名化→验证
- 边界测试：冻结期内撤回、重复注销请求、并发注销

---

#### 3.1.2 NF-02: 网关请求超时配置

**需求关联**: PRD 3.1 NF-02

**技术方案**:

为 `api-gateway` 的 `httputil.ReverseProxy` 配置自定义 Transport，添加全局请求超时中间件。

```go
// internal/proxy/transport.go
transport := &http.Transport{
    ResponseHeaderTimeout: 30 * time.Second,
    IdleConnTimeout:       90 * time.Second,
    MaxIdleConns:          100,
    MaxIdleConnsPerHost:   10,
    DisableKeepAlives:     false,
}

// internal/middleware/timeout.go
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()
        c.Request = c.Request.WithContext(ctx)
        done := make(chan struct{})
        go func() {
            c.Next()
            close(done)
        }()
        select {
        case <-done:
            return
        case <-ctx.Done():
            c.JSON(504, gin.H{
                "error":   "request_timeout",
                "message": "request exceeded 60s timeout",
            })
            c.Abort()
        }
    }
}
```

**关键代码路径**:
- `api-gateway/internal/proxy/transport.go` — 自定义 Transport 配置
- `api-gateway/internal/middleware/timeout.go` — 全局超时中间件
- `api-gateway/cmd/main.go` — 注册中间件到 Gin engine

**配置项**（通过 config-service 管理）:
- `gateway.response_header_timeout_sec`: 30（默认）
- `gateway.idle_conn_timeout_sec`: 90（默认）
- `gateway.global_request_timeout_sec`: 60（默认）

**测试策略**:
- 单元测试：模拟慢后端（sleep handler），验证 504 返回
- 压力测试：使用 wrk 模拟后端卡死，确认网关连接池不耗尽

---

#### 3.1.3 AR-13: 密码哈希升级至 argon2id

**需求关联**: PRD 3.1 AR-13

**技术方案**:

采用渐进式迁移策略，新注册/改密直接使用 argon2id，存量 SM3 用户登录时透明 rehash。

```
┌───────────────┐     登录请求      ┌──────────────────────┐
│  用户登录      │ ───────────────→  │  auth-service         │
│  phone+password│                   │                       │
└───────────────┘                   │  1. 读取 password_hash │
                                    │  2. 检查前缀:          │
                                    │     $argon2id$ → argon2id 验证 │
                                    │     $sm3$ → SM3 验证          │
                                    │  3. 验证成功?            │
                                    │     YES → 检查是否 SM3  │
                                    │            → rehash 为 argon2id│
                                    │     NO  → 返回认证失败  │
                                    └──────────────────────┘
```

**密码哈希存储格式**:
- argon2id: `$argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>`
- SM3（过渡期保留）: `$sm3$<salt>$<hash>`

**argon2id 参数配置**:
- `memory`: 64 MB (65536 KB)
- `iterations`: 3
- `parallelism`: 2
- `salt length`: 16 bytes
- `key length`: 32 bytes

**关键代码路径**:
- `auth-service/internal/auth/argon2id.go` — argon2id 哈希和验证函数
- `auth-service/internal/auth/sm3.go` — 保留现有 SM3 验证（只读，不再生成新 SM3 哈希）
- `auth-service/internal/auth/hash_factory.go` — 哈希策略工厂，根据前缀选择算法
- `auth-service/internal/service/auth.go` — 登录逻辑中添加 rehash 分支

**数据库变更**:
- 无 schema 变更，`password_hash` 字段长度已足够（varchar 255）
- 新增 Goose migration 验证：`SELECT COUNT(*) FROM users WHERE password_hash LIKE '$argon2id$%'` — 用于监控迁移进度

**审计保留**:
- SM3 哈希算法保留用于审计日志完整性校验（`compliance-service` 中 SM3 哈希链功能不受影响）
- 审计日志中的密码变更事件记录算法类型

**测试策略**:
- 单元测试：argon2id 哈希/验证、SM3 验证、rehash 逻辑、前缀识别
- 集成测试：完整登录流程，验证 SM3→argon2id 迁移，确认数据库更新
- 性能测试：argon2id 哈希计算耗时 < 500ms

---

#### 3.1.4 AR-16: 第三方安全渗透测试

**需求关联**: PRD 3.1 AR-16

**技术方案**:

渗透测试为外部服务采购，本节定义测试范围、前置条件和支持工作。

**测试范围**:
- **Web 应用测试**: api-gateway 暴露的全部 `/api/v1/*` 端点
- **认证安全测试**: JWT token 机制、密码策略、session 管理、OAuth 流程
- **数据安全测试**: PII 脱敏有效性、SQL 注入、XSS/CSRF
- **API 安全测试**: 认证绕过、越权访问（IDOR）、速率限制有效性
- **移动端安全测试**: APK/IPA 逆向分析、证书固定验证、本地存储安全、网络通信加密
- **依赖库漏洞扫描**: 使用 Trivy 扫描 Docker 镜像，Snyk 扫描 Go/CocoaPods/Gradle 依赖

**前置条件**:
- AR-13（argon2id）必须先完成，确保渗透测试基于最新安全基线
- 搭建独立的 UAT 测试环境，配置与生产环境一致
- 准备测试账号集合（不同等级：L0-L4、Admin 角色）

**支持工具配置**:
- `scripts/security/` 目录放置辅助脚本：
  - `scan_dependencies.sh` — Trivy 镜像扫描 + Snyk 依赖扫描
  - `generate_test_accounts.go` — 生成各等级测试账号
  - `zap_scan_config.json` — OWASP ZAP 自动化扫描配置

**交付物**:
- 第三方渗透测试报告（含修复建议）
- 依赖扫描报告（Trivy + Snyk）
- 移动端安全审计报告
- 修复验证报告（所有 Critical/High 漏洞修复后复测通过）

**测试策略**:
- 内部预扫描：在第三方介入前，使用 OWASP ZAP 进行初步自动化扫描
- 漏洞修复后必须通过回归测试（AR-17/AR-18 覆盖）

---

#### 3.1.5 AR-17: 核心服务单元测试补齐

**需求关联**: PRD 3.1 AR-17

**技术方案**:

为 credit-service、subscription-service（account-service 内）、rebate_service（account-service 内）补齐单元测试。

**测试框架选型**:
- `testing`（Go 标准库）+ `github.com/stretchr/testify`（断言/mock）
- Mock: `github.com/stretchr/testify/mock` 接口 mock
- 覆盖率: `go test -coverprofile=coverage.out -covermode=atomic`

**按服务分解的测试计划**:

| 服务 | 目标覆盖率 | 关键测试模块 |
|------|-----------|-------------|
| credit-service | >60% | 积分 CRUD、积分消费、积分过期、RFM 计算、积分抵扣逻辑 |
| account-service（subscription） | >60% | 订阅购买、升降级、续费、过期降级、权益发放 |
| account-service（rebate） | >60% | 推荐关系、阶梯返利计算、被邀请人奖励、返利结算 |

**测试文件结构**:
```
credit-service/
├── internal/service/credit_test.go      — 积分业务逻辑测试
├── internal/service/rfm_test.go         — RFM 计算测试
├── internal/repository/credit_test.go   — 数据层测试（mock DB）
└── internal/handler/credit_test.go      — HTTP handler 测试

account-service/
├── internal/service/subscription_test.go — 订阅业务逻辑测试
├── internal/service/rebate_test.go       — 返利业务逻辑测试
├── internal/service/level_test.go        — 等级升降级测试
└── internal/repository/subscription_test.go
```

**关键代码路径**:
- 各服务 `internal/service/*_test.go` — 核心业务逻辑测试
- `scripts/test/coverage_check.sh` — 覆盖率门禁脚本

**CI 集成**:
- GitHub Actions job 中执行 `go test -coverprofile=coverage.out ./...`
- 使用 `goverage` 检查阈值，低于 60% 构建失败

**测试策略**:
- 正常路径（Happy Path）测试
- 边界条件测试（积分余额为 0 时消费、订阅过期临界点、返利阶梯边界）
- 错误路径测试（数据库错误、并发冲突）
- 不使用真实数据库，全部通过 mock repository 接口

---

#### 3.1.6 AR-18: 集成测试（全链路）

**需求关联**: PRD 3.1 AR-18

**技术方案**:

基于 Docker Compose 搭建完整测试环境，实现跨服务全链路集成测试。

```
┌──────────────────────────────────────────────────────┐
│  Docker Compose Test Environment                     │
│                                                      │
│  api-gateway:30300 ←→ account-service:30301          │
│       ↓                    ←→ auth-service:30302     │
│       ↓                    ←→ credit-service:30312   │
│       ↓                    ←→ notification:30311     │
│       ↓                    ←→ compliance:30313       │
│                                                      │
│  PostgreSQL:5432  Redis:6379  (test instances)       │
└──────────────────────────────────────────────────────┘
```

**测试框架选型**:
- `testing` + `github.com/stretchr/testify`
- `github.com/testcontainers/testcontainers-go`（可选，替代直接 docker-compose）
- `github.com/jarcoal/httpmock` 或真实 HTTP 调用

**全链路测试用例**:

```go
func TestFullJourney(t *testing.T) {
    // Step 1: 用户注册
    user := registerUser(t, phone, code)
    
    // Step 2: 用户登录
    token := loginUser(t, phone, password)
    
    // Step 3: 查看积分余额（应为初始值）
    credits := getCredits(t, token)
    assert.Equal(t, 0, credits.Balance)
    
    // Step 4: 推荐好友
    referralCode := getReferralCode(t, token)
    
    // Step 5: 好友注册+实名
    friend := registerUser(t, friendPhone, code)
    verifyIdentity(t, friendToken)
    
    // Step 6: 验证推荐奖励积分到账
    credits = getCredits(t, token)
    assert.Greater(t, credits.Balance, 0)
    
    // Step 7: 订阅购买（使用积分抵扣）
    order := createSubscription(t, token, "standard_monthly", creditsToDeduct)
    
    // Step 8: 验证等级变更
    profile := getProfile(t, token)
    assert.Equal(t, "L2", profile.Level)
    
    // Step 9: 模拟订阅过期
    expireSubscription(t, order.ID)
    
    // Step 10: 验证降级
    profile = getProfile(t, token)
    assert.Equal(t, "L1", profile.Level)
}
```

**关键代码路径**:
- `tests/integration/full_journey_test.go` — 全链路测试主文件
- `tests/integration/helpers.go` — 测试辅助函数（HTTP client、断言工具）
- `tests/integration/docker-compose.test.yml` — 测试专用 Docker Compose

**CI 集成**:
- GitHub Actions job: `make integration-test`
- 先启动 Docker Compose，等待所有服务就绪（health check），执行测试，最后 `docker compose down`

**测试策略**:
- 全链路主路径测试（注册→登录→订阅→积分→推荐→过期降级）
- 各服务间通信异常场景测试（下游超时、熔断触发）
- 数据一致性验证（积分余额、订阅状态、等级变更在多服务间一致）

---

#### 3.1.7 AR-23: 数据库备份策略

**需求关联**: PRD 3.1 AR-23

**技术方案**:

**PostgreSQL 备份方案**:

```
┌──────────────┐   pg_dump    ┌──────────┐   upload    ┌─────────┐
│  PostgreSQL  │ ────────────→│ 本地临时  │ ──────────→│   OSS   │
│  18.x        │   每日 02:00 │  目录     │   完成后    │  (S3)   │
│              │              │          │   删除本地  │         │
└──────────────┘              └──────────┘             └─────────┘
       │
       │ WAL 归档
       ▼
  pg_archive_mode=on
  archive_command='scp %p backup-server:/wal_archive/%f'
```

- 全量备份：`pg_dump -Fc` 自定义格式，支持并行恢复，每日 02:00 UTC
- WAL 归档：`archive_mode=on`，支持时间点恢复（PITR）
- 保留策略：每日备份保留 7 天，每周备份保留 4 周，每月备份保留 12 个月
- 恢复演练：至少每季度一次，验证备份完整性和恢复时间

**Redis 备份方案**:

```
┌──────────────┐  RDB (默认)   ┌──────────┐
│  Redis 8.2   │ ────────────→│ dump.rdb │
│  LTS         │  每 5 分钟    │ (自动)    │
│              │               └──────────┘
│              │  AOF          ┌──────────┐
│              │ ────────────→│ appendonly│
│              │  everysec     │ .aof      │
└──────────────┘               └──────────┘
       │
       │ 定期备份脚本 (每日 03:00)
       ▼
  redis-cli BGSAVE → copy dump.rdb → upload to OSS
```

- 持久化配置：RDB + AOF 混合模式（`aof-use-rdb-preamble yes`）
- AOF 策略：`appendfsync everysec`（性能与可靠性平衡）
- 定期备份：每日 03:00 UTC 触发 `BGSAVE`，拷贝 dump.rdb 至 OSS
- 保留策略：与 PG 相同（7 天/4 周/12 个月）

**关键代码路径**:
- `infra/backup/pg_backup.sh` — PostgreSQL 备份脚本
- `infra/backup/redis_backup.sh` — Redis 备份脚本
- `infra/backup/restore_test.sh` — 恢复演练脚本
- `infra/backup/crontab` — 定时任务配置

**监控告警**:
- 备份执行结果写入 Prometheus 指标：`backup_last_success_timestamp`
- 备份失败告警：超过 26 小时未成功备份触发 P1 告警

**测试策略**:
- 恢复演练：在独立环境执行 PG 全量恢复 + PITR，验证数据完整性
- Redis 恢复演练：从 RDB 恢复，验证所有 key 完整
- 记录恢复演练报告（RTO/RPO 实测值）

---

#### 3.1.8 AR-25: 清理仓库

**需求关联**: PRD 3.1 AR-25

**技术方案**:

**清理步骤**:

1. **识别二进制残留**:
   ```bash
   git ls-files | grep -E '\.(exe|dll|bin)$'
   git ls-files | grep -E '/(nul|main\.exe|cmd\.exe)$'
   ```

2. **从 Git 索引移除**:
   ```bash
   git rm --cached main.exe cmd.exe nul
   ```

3. **更新 `.gitignore`**:
   ```
   # Binaries
   *.exe
   *.exe~
   *.dll
   *.so
   *.dylib
   nul
   
   # Build output
   /bin/
   /dist/
   
   # IDE
   .idea/
   .vscode/
   *.swp
   
   # OS
   .DS_Store
   Thumbs.db
   
   # Test
   coverage.out
   *.test
   ```

4. **修正文档不一致**:
   - 对比 README.md、ARCHITECTURE.md 与实际代码结构
   - 更新端口描述、服务列表、目录结构
   - 确保所有文档中的端口号与服务实际端口一致

**关键代码路径**:
- `.gitignore` — 更新忽略规则
- `README.md` — 修正项目结构描述
- `ARCHITECTURE.md` — 修正端口和服务描述

**测试策略**:
- `git status` 确认无二进制文件被追踪
- `git ls-files | grep -E '\.exe$'` 返回空
- 文档中端口号与实际 `docker-compose.yml` 一致

---

#### 3.1.9 AR-21: K8s Helm Chart + CI/CD

**需求关联**: PRD 3.1 AR-21

**技术方案**:

**Helm Chart 结构**:

```
helm/account-center/
├── Chart.yaml                    # Chart 元数据
├── values.yaml                   # 默认配置值
├── values-dev.yaml               # Dev 环境覆盖
├── values-uat.yaml               # UAT 环境覆盖
├── values-prod.yaml              # Prod 环境覆盖
└── templates/
    ├── _helpers.tpl               # 通用模板函数
    ├── api-gateway/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   ├── configmap.yaml
    │   ├── hpa.yaml
    │   └── ingress.yaml
    ├── account-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── auth-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── notification-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── credit-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── compliance-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── data-product-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── config-service/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── payment-service/            # V2.0 新增
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   └── configmap.yaml
    ├── secrets.yaml                # 外部 Secrets（KMS 引用）
    └── _infra/
        ├── postgresql.yaml         # PG StatefulSet
        ├── redis.yaml              # Redis StatefulSet
        └── monitoring.yaml         # VM+Loki+Grafana
```

**HPA 配置**（values.yaml 示例）:

```yaml
api-gateway:
  replicaCount: 2
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80

account-service:
  replicaCount: 2
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 8
    targetCPUUtilizationPercentage: 70
```

**滚动更新策略**:

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1
    maxSurge: 1
```

**CI/CD 流水线**（GitHub Actions）:

```
┌─────────────┐   ┌──────────────┐   ┌────────────┐   ┌──────────────┐   ┌───────────┐
│ golangci-   │──→│ go test      │──→│ docker     │──→│ push to      │──→│ deploy to │
│ lint        │   │ -cover       │   │ build      │   │ registry     │   │ UAT (K8s) │
└─────────────┘   └──────────────┘   └────────────┘   └──────────────┘   └───────────┘
    ↓ fail            ↓ fail            ↓ fail            ↓ fail
  Block PR         Block PR         Block PR         Block PR
```

`.github/workflows/ci.yml` 结构:
- **Trigger**: push to `main` / `develop`, PR to `main`
- **Lint**: `golangci-lint run ./...`
- **Test**: `go test -race -coverprofile=coverage.out ./...`
- **Build**: `docker build -t account-center/{service}:${{ github.sha }} .`
- **Push**: 推送到 Container Registry（Docker Hub / 阿里云 ACR）
- **Deploy**: `helm upgrade --install account-center ./helm/account-center -n uat`

**关键代码路径**:
- `helm/account-center/` — Helm Chart 全部文件
- `.github/workflows/ci.yml` — CI/CD 流水线
- `Makefile` — 统一构建入口（lint/test/build/push/deploy）

**测试策略**:
- `helm lint helm/account-center/` — Chart 语法验证
- `helm template` — 渲染验证（确保模板无语法错误）
- `helm upgrade --dry-run` — 干跑验证（确认 K8s 资源正确）
- Dev 环境实际部署验证

---

#### 3.1.10 FN-01: 支付网关集成

**需求关联**: PRD 3.1 FN-01  
**依赖**: FN-02（订单管理系统）

**技术方案**:

在新增的 `payment-service`（端口 30316）中实现支付网关集成，支持微信支付和支付宝。

```
┌──────────────┐   创建订单    ┌─────────────────┐   唤起支付    ┌──────────────┐
│  前端/移动端  │ ───────────→ │  payment-service │ ───────────→ │  微信支付     │
│              │              │  (30316)         │              │  /支付宝      │
└──────────────┘              └────────┬────────┘              └──────┬───────┘
                                       │                              │
                                       │ 异步回调                      │ 异步回调
                                       ▼                              │
                              ┌─────────────────┐                    │
                              │  callback handler│ ←─────────────────┘
                              │  1. 验签         │
                              │  2. 幂等检查     │
                              │  3. 更新订单状态  │
                              │  4. 发放权益      │
                              │  5. 写审计日志    │
                              └─────────────────┘
                                       │
                              ┌────────▼────────┐
                              │  account-service │
                              │  (权益发放)      │
                              └─────────────────┘
```

**payment-service 内部架构**:

```
payment-service/
├── cmd/main.go
├── internal/
│   ├── handler/
│   │   ├── payment.go          # 创建支付、查询支付状态
│   │   └── callback.go         # 支付回调处理
│   ├── service/
│   │   ├── payment.go          # 支付业务逻辑
│   │   ├── wechat_pay.go       # 微信支付 provider
│   │   ├── alipay.go           # 支付宝 provider
│   │   └── reconciliation.go   # 对账逻辑
│   ├── repository/
│   │   ├── order.go            # 订单数据层
│   │   └── payment_record.go   # 支付流水数据层
│   ├── model/
│   │   ├── order.go            # 订单模型
│   │   └── payment_record.go   # 支付流水模型
│   └── provider/               # 支付渠道 provider 接口
│       └── provider.go         # 统一接口定义
└── pkg/
    └── wechat/                 # 微信支付 SDK 封装
```

**Provider 接口设计**:

```go
type PaymentProvider interface {
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error)
    QueryPayment(ctx context.Context, outTradeNo string) (*PaymentStatus, error)
    Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error)
    VerifyCallback(r *http.Request) (*CallbackData, error)
    Name() string
}
```

**微信支付场景**:
- H5 支付：适用于手机浏览器，返回 `mweb_url` 跳转
- 小程序支付：调用 `wx.requestPayment`，需要 `prepay_id`
- Native 扫码支付：生成二维码 URL，用户扫码支付

**支付宝场景**:
- 手机网站支付：H5 页面跳转支付宝收银台
- APP 支付：SDK 调用支付宝客户端

**回调处理流程**:
1. 接收回调 → 2. 验签（平台公钥） → 3. 幂等检查（order_id + status） → 4. 更新订单为 `paid` → 5. 调用 account-service 发放权益 → 6. 写审计日志 → 7. 返回 `SUCCESS`

**对账方案**:
- 每日 06:00 定时任务，下载前一日微信支付/支付宝对账单
- 逐笔比对订单状态与平台状态，差异订单标记 `reconcile_mismatch`
- 差异告警推送到运维（钉钉/企微 webhook）

**关键代码路径**:
- `payment-service/cmd/main.go` — 服务入口
- `payment-service/internal/provider/provider.go` — Provider 统一接口
- `payment-service/internal/service/wechat_pay.go` — 微信支付实现
- `payment-service/internal/service/alipay.go` — 支付宝实现
- `payment-service/internal/handler/callback.go` — 回调处理
- `payment-service/internal/service/reconciliation.go` — 对账逻辑

**数据库变更**:
- 新增 `payment_records` 表（见第 4 章数据设计）
- 新增 `reconciliation_results` 表

**API 变更**:
- `POST /api/v1/payments` — 创建支付
- `GET /api/v1/payments/{id}` — 查询支付状态
- `POST /api/v1/payments/wechat/callback` — 微信支付回调
- `POST /api/v1/payments/alipay/callback` — 支付宝回调
- `GET /api/v1/admin/reconciliation` — 对账结果查询

**测试策略**:
- 单元测试：mock 支付平台 SDK，验证 Provider 接口实现
- 集成测试：使用支付平台沙箱环境（微信支付沙箱、支付宝沙箱）
- 回调测试：模拟回调请求，验证幂等性和状态更新
- 对账测试：构造差异数据，验证标记和告警

---

#### 3.1.11 FN-02: 订单管理系统

**需求关联**: PRD 3.1 FN-02

**技术方案**:

在 `payment-service` 中实现订单管理系统，独立于 account-service 的订阅逻辑。

**订单状态机**:

```
                  ┌──────────┐
    创建订单 ────→ │ pending  │
                  └────┬─────┘
                       │ 支付成功回调
                       ▼
                  ┌──────────┐
                  │   paid   │ ────→ 发放权益
                  └────┬─────┘
                       │ 用户申请退款
                       ▼
                  ┌──────────┐
                  │ refunded │ ────→ 撤回权益
                  └──────────┘

  pending ──(超时/用户取消)──→ cancelled
```

**非法状态跳转校验**:
- `cancelled` → `paid`: 禁止
- `refunded` → `paid`: 禁止
- `paid` → `pending`: 禁止
- 允许: `pending` → `paid`, `pending` → `cancelled`, `paid` → `refunded`

**数据库表结构**（`orders` 表）:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 订单唯一标识 |
| user_id | UUID | NOT NULL, FK → users.id | 用户 ID |
| order_no | VARCHAR(32) | UNIQUE, NOT NULL | 订单编号（业务展示用） |
| product_type | VARCHAR(20) | NOT NULL | 商品类型：subscription/credit_pack |
| product_id | VARCHAR(50) | NOT NULL | 商品 ID（套餐 ID / 积分包 ID） |
| product_name | VARCHAR(100) | NOT NULL | 商品名称（冗余，快照） |
| amount_cents | INTEGER | NOT NULL | 订单金额（分） |
| currency | VARCHAR(3) | DEFAULT 'CNY' | 货币类型 |
| credits_used | INTEGER | DEFAULT 0 | 积分抵扣数量 |
| credits_discount_cents | INTEGER | DEFAULT 0 | 积分抵扣金额（分） |
| actual_amount_cents | INTEGER | NOT NULL | 实付金额（分） |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | 订单状态 |
| payment_method | VARCHAR(20) | | 支付方式：wechat/alipay |
| payment_channel | VARCHAR(20) | | 支付渠道：h5/mini/native/app |
| paid_at | TIMESTAMP | | 支付完成时间 |
| cancelled_at | TIMESTAMP | | 取消时间 |
| refunded_at | TIMESTAMP | | 退款时间 |
| expires_at | TIMESTAMP | NOT NULL | 订单过期时间（待付订单 30 分钟） |
| metadata | JSONB | DEFAULT '{}' | 扩展元数据 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |

**索引**:
- `idx_orders_user_id` ON (user_id)
- `idx_orders_status` ON (status)
- `idx_orders_order_no` UNIQUE ON (order_no)
- `idx_orders_created_at` ON (created_at)
- `idx_orders_expires_at` ON (expires_at) WHERE status = 'pending'

**关键代码路径**:
- `payment-service/internal/model/order.go` — 订单模型
- `payment-service/internal/repository/order.go` — 订单数据层
- `payment-service/internal/service/order.go` — 订单业务逻辑（状态机）
- `payment-service/internal/handler/order.go` — 订单 HTTP handler

**API 变更**:
- `POST /api/v1/orders` — 创建订单
- `GET /api/v1/orders/{id}` — 查询订单详情
- `GET /api/v1/orders` — 订单列表（支持分页、过滤）
- `POST /api/v1/orders/{id}/cancel` — 取消订单
- `GET /api/v1/admin/orders` — 管理端订单查询（多维度）
- `GET /api/v1/admin/orders/export` — 订单导出（CSV/Excel）

**测试策略**:
- 单元测试：状态机所有合法/非法状态跳转、订单过期处理、金额计算（积分抵扣）
- 集成测试：创建订单→支付→权益发放→取消/退款完整流程

---

#### 3.1.12 FN-05: 用户管理后台

**需求关联**: PRD 3.1 FN-05

**技术方案**:

在 account-service 中新增 Admin API 模块，与管理端前端配合实现用户管理功能。

**Admin API 架构**:

```
┌──────────────────┐     JWT (Admin Role)     ┌──────────────────┐
│  Admin Frontend  │ ────────────────────────→ │  api-gateway     │
│  (Vue 3)         │                          │  /api/v1/admin/* │
└──────────────────┘                          └────────┬─────────┘
                                                       │ 角色鉴权
                                                       ▼
                                              ┌──────────────────┐
                                              │  account-service │
                                              │  Admin Module    │
                                              │                  │
                                              │  - 用户列表/搜索  │
                                              │  - 用户详情       │
                                              │  - 等级调整       │
                                              │  - 积分调整       │
                                              │  - 封禁/解封      │
                                              │  - 实名审核       │
                                              └──────────────────┘
```

**鉴权设计**:
- 网关层：JWT 中 `role` 字段为 `admin` 才允许访问 `/api/v1/admin/*`
- 服务层：account-service 中 Admin handler 独立路由组，添加 `RequireAdmin` 中间件
- 审计：所有 Admin 操作写入 `admin_audit_logs` 表

**关键代码路径**:
- `account-service/internal/handler/admin.go` — Admin API handler
- `account-service/internal/service/admin.go` — Admin 业务逻辑
- `account-service/internal/middleware/admin_auth.go` — Admin 鉴权中间件

**API 变更**:
- `GET /api/v1/admin/users` — 用户列表（分页、过滤）
- `GET /api/v1/admin/users/{id}` — 用户详情
- `PUT /api/v1/admin/users/{id}/level` — 等级调整
- `POST /api/v1/admin/users/{id}/credits/adjust` — 积分调整
- `PUT /api/v1/admin/users/{id}/ban` — 封禁用户
- `PUT /api/v1/admin/users/{id}/unban` — 解封用户
- `PUT /api/v1/admin/users/{id}/identity/approve` — 实名审核通过
- `PUT /api/v1/admin/users/{id}/identity/reject` — 实名审核驳回
- `GET /api/v1/admin/audit-logs` — 审计日志查询

**数据库变更**:
- 新增 `admin_users` 表（管理员账户，独立于普通用户表）
- 新增 `admin_audit_logs` 表（管理操作审计日志）

**测试策略**:
- 单元测试：Admin 鉴权中间件（非 Admin 角色拒绝）、各 API handler
- 集成测试：Admin 操作→审计日志记录→权限隔离验证
- 安全测试：验证普通用户 token 无法访问 Admin API

---

#### 3.1.13 FN-10: APNs/FCM 推送集成

**需求关联**: PRD 3.1 FN-10

**技术方案**:

扩展 notification-service 的 provider 架构，新增 APNs 和 FCM 推送 provider。

```
┌──────────────────┐     推送请求     ┌───────────────────────────┐
│  account-service │ ──────────────→ │  notification-service      │
│  (触发事件)       │                │                            │
└──────────────────┘                │  Push Provider Interface   │
                                    │  ┌───────────────────────┐ │
                                    │  │ APNs Provider (iOS)   │ │
                                    │  │  - HTTP/2             │ │
                                    │  │  - JWT Token Auth     │ │
                                    │  └───────────────────────┘ │
                                    │  ┌───────────────────────┐ │
                                    │  │ FCM Provider (Android)│ │
                                    │  │  - HTTP v1 API        │ │
                                    │  └───────────────────────┘ │
                                    │  ┌───────────────────────┐ │
                                    │  │ HMS Provider (华为)   │ │
                                    │  │  - Push Kit API       │ │
                                    │  └───────────────────────┘ │
                                    └───────────────────────────┘
```

**Provider 接口设计**（与 SMS provider 模式一致）:

```go
type PushProvider interface {
    Send(ctx context.Context, notification PushNotification) (*PushResult, error)
    SendBatch(ctx context.Context, notifications []PushNotification) ([]PushResult, error)
    ValidateToken(ctx context.Context, token string) (bool, error)
    Name() string
}
```

**设备 Token 管理**:
- `POST /api/v1/devices/register` — 设备注册/更新 Token
- `DELETE /api/v1/devices/{token}` — 设备注销
- 数据库 `push_tokens` 表：user_id, device_type (ios/android), token, app_version, created_at, updated_at

**APNs 实现方案**:
- 使用 `github.com/sideshow/apns2` 库
- JWT Token 认证（.p8 密钥文件，通过 KMS 管理）
- 支持 `alert`（通知栏）和 `background`（静默推送）两种类型
- 环境区分：Development / Production

**FCM 实现方案**:
- 使用 Google Firebase Admin SDK for Go (`firebase.google.com/go/v4/messaging`)
- Service Account JSON 认证（通过 KMS 管理）
- 支持 `data` 消息和 `notification` 消息

**华为 HMS 实现方案**:
- 使用华为 Push Kit REST API
- OAuth 2.0 客户端凭证认证
- 国内 Android 设备主要推送通道

**关键代码路径**:
- `notification-service/internal/provider/push.go` — Push Provider 接口
- `notification-service/internal/provider/apns.go` — APNs 实现
- `notification-service/internal/provider/fcm.go` — FCM 实现
- `notification-service/internal/provider/hms.go` — HMS 实现
- `notification-service/internal/handler/device.go` — 设备 Token 管理
- `notification-service/internal/model/push_token.go` — Token 数据模型

**数据库变更**:
- 新增 `push_tokens` 表（见第 4 章数据设计）
- 新增 `push_logs` 表（推送发送记录）

**API 变更**:
- `POST /api/v1/devices/register` — 注册/更新设备 Token
- `DELETE /api/v1/devices/{token}` — 注销设备
- `POST /api/v1/notifications/push` — 发送推送（内部 API）
- `GET /api/v1/admin/notifications/push/logs` — 推送日志查询

**测试策略**:
- 单元测试：mock APNs/FCM/HMS HTTP 调用，验证 Provider 接口
- 集成测试：使用 APNs/FCM 沙箱环境发送真实推送
- Token 管理测试：注册、更新、注销、过期 Token 清理

---

#### 3.1.14 UX-08: 定价透明度

**需求关联**: PRD 3.1 UX-08

**技术方案**:

定价信息由 config-service 动态配置管理，前端通过 API 获取并渲染定价页面。

**定价配置数据结构**（config-service）:

```json
{
  "pricing_plans": [
    {
      "id": "free",
      "name": "免费版",
      "level": "L0",
      "monthly_price_cents": 0,
      "yearly_price_cents": 0,
      "features": [
        {"name": "基础功能", "included": true},
        {"name": "AI 调用额度", "value": "10 次/天"},
        {"name": "积分倍率", "value": "1.0x"}
      ]
    },
    {
      "id": "standard_monthly",
      "name": "标准版（月付）",
      "level": "L2",
      "monthly_price_cents": 2900,
      "yearly_price_cents": null,
      "features": [
        {"name": "标准版全部功能", "included": true},
        {"name": "AI 调用额度", "value": "100 次/天"},
        {"name": "积分倍率", "value": "2.0x"}
      ]
    }
  ],
  "credit_exchange_rate": 100,
  "credit_value_cents": 1
}
```

**前端实现**:

```
pricing-page/
├── PricingPage.vue           # 定价页面主组件
├── PlanCard.vue              # 等级卡片组件
├── FeatureComparisonTable.vue # 权益对比矩阵
├── CreditCalculator.vue      # 积分抵扣计算器
└── MobilePricingPage.swift   # iOS 定价页
└── PricingScreen.kt          # Android 定价页
```

**积分抵扣计算器逻辑**:
- 服务端 API：`GET /api/v1/pricing/credit-discount?credits={n}&plan_id={id}`
- 返回：`{ "original_price_cents": 2900, "credit_discount_cents": 1000, "final_price_cents": 1900 }`
- 前端实时计算：输入积分数 → 调用 API → 展示最终金额

**关键代码路径**:
- `account-service/internal/handler/pricing.go` — 定价 API handler
- `account-service/internal/service/pricing.go` — 定价计算逻辑
- `web/src/views/PricingPage.vue` — Web 定价页面
- `ios/AccountCenter/Views/PricingView.swift` — iOS 定价页
- `android/.../PricingScreen.kt` — Android 定价页

**API 变更**:
- `GET /api/v1/pricing/plans` — 获取所有定价方案
- `GET /api/v1/pricing/credit-discount` — 积分抵扣计算

**测试策略**:
- 单元测试：定价计算逻辑（积分抵扣、年度折扣）
- 前端测试：定价页面各组件渲染、计算器交互
- 跨端一致性测试：Web/iOS/Android 定价信息展示一致

---

### 3.1 P0 技术设计小结

| 需求 ID | 技术方案核心 | 涉及服务 | 新增文件估计 |
|---------|------------|---------|-------------|
| NF-01 | Asynq deletion-worker + 数据匿名化 | account-service | ~8 文件 |
| NF-02 | ReverseProxy Transport 超时 + 全局中间件 | api-gateway | ~4 文件 |
| AR-13 | argon2id 渐进式迁移 + 前缀标识 | auth-service | ~5 文件 |
| AR-16 | 外部渗透测试 + 内部预扫描工具 | 全服务 | ~4 脚本 |
| AR-17 | credit/subscription/rebate 单元测试补齐 | 3 服务 | ~12 测试文件 |
| AR-18 | Docker Compose 全链路集成测试 | 全服务 | ~5 测试文件 |
| AR-23 | pg_dump + WAL + Redis RDB/AOF 备份脚本 | 基础设施 | ~4 脚本 |
| AR-25 | git rm + .gitignore + 文档修正 | 全局 | ~3 文件 |
| AR-21 | Helm Chart 9 服务 + HPA + 滚动更新 | 全服务 | ~40+ YAML |
| FN-01 | payment-service 支付网关（微信/支付宝） | payment-service(NEW) | ~20 文件 |
| FN-02 | 订单状态机 + orders 表 | payment-service | ~8 文件 |
| FN-05 | Admin API + 权限隔离 + 审计日志 | account-service | ~6 文件 |
| FN-10 | APNs/FCM/HMS 推送 Provider 架构 | notification-service | ~10 文件 |
| UX-08 | config-service 定价配置 + 四端定价页 | 全端 | ~8 文件 |

### 3.2 Phase 7 — P1 技术设计（32 项）

> 以下 32 项为 V2.1 迭代目标，Phase 6 全部完成后启动，覆盖可靠性、UX、商业化、移动端、架构与可观测性维度。

---

#### 3.2.1 NF-03: 服务间熔断器提升为共享包

**需求关联**: PRD 3.2 NF-03

**技术方案**:

将现有 `notification-service/pkg/circuitbreaker` 提升至顶层共享包 `pkg/circuitbreaker`，增强为可配置的多维度熔断器，所有微服务统一引入。

```
┌───────────────────────────────────────────────────┐
│  pkg/circuitbreaker (共享包)                        │
│                                                    │
│  ┌──────────────┐  ┌──────────────┐               │
│  │ Config       │  │ State Machine│               │
│  │ - maxFailures│  │ Closed       │               │
│  │ - errorRate  │  │ Open         │               │
│  │ - timeoutRate│  │ HalfOpen     │               │
│  │ - halfOpenWait│ └──────┬───────┘               │
│  │ - onStateChange│       │                       │
│  └──────────────┘       │                       │
│                         ▼                       │
│  ┌──────────────────────────────────────────┐     │
│  │  Metrics Integration                      │     │
│  │  - circuitbreaker_state (gauge)           │     │
│  │  - circuitbreaker_failures_total (counter)│     │
│  │  - circuitbreaker_opens_total (counter)   │     │
│  └──────────────────────────────────────────┘     │
└───────────────────────────────────────────────────┘
         │
    ┌────┴────┬──────────┬───────────┬──────────┐
    ▼         ▼          ▼           ▼          ▼
 api-gw   account    auth      credit    data-product
          (替换内联)  (替换内联)  (新增)    (新增)
```

**增强特性**（相比现有实现）:
- 新增 `Config` 结构体：`MaxFailures int`、`ErrorRateThreshold float64`、`HalfOpenWait time.Duration`、`OnStateChange func(old, new State)`
- 新增 `Name() string` 方法，用于标识各服务熔断器实例
- 新增 Prometheus 指标集成（`circuitbreaker_state`、`circuitbreaker_failures_total`）
- 新增 `ExecuteWithFallback(ctx, f func() error, fallback func() error) error` 降级方法

**关键代码路径**:
- `pkg/circuitbreaker/circuitbreaker.go` — 从 `notification-service/pkg/circuitbreaker/circuitbreaker.go` 迁移并增强
- `pkg/circuitbreaker/config.go` — 新增配置结构体
- `pkg/circuitbreaker/metrics.go` — Prometheus 指标集成
- `notification-service/cmd/main.go` — 改为引用 `github.com/trigold786/92-Account-Center/pkg/circuitbreaker`
- `notification-service/internal/service/sms_service.go` — 更新 import 路径
- `api-gateway/internal/middleware/circuitbreaker.go` — 网关侧熔断器中间件
- 各服务 `cmd/main.go` — 初始化共享熔断器实例

**测试策略**:
- 单元测试：Closed→Open→HalfOpen→Closed 完整状态机，错误率阈值触发，半开恢复逻辑
- 单元测试：`ExecuteWithFallback` 降级回调验证
- 并发测试：多 goroutine 并发调用 `Execute`，验证状态一致性
- 覆盖率目标 ≥80%

---

#### 3.2.2 NF-04: 健康检查增加真实依赖检测

**需求关联**: PRD 3.2 NF-04

**技术方案**:

为每个微服务的 `/health` 端点增加真实依赖探测（PostgreSQL、Redis、下游服务），替代当前的进程存活检查。

```
GET /health

┌───────────────────┐
│  /health handler  │
│                   │
│  1. Check PG      │ ──→ SELECT 1          (timeout: 2s)
│  2. Check Redis   │ ──→ PING              (timeout: 1s)
│  3. Check Downstream│ ──→ HTTP HEAD       (timeout: 2s)
│                   │
│  All OK → 200     │
│  Any FAIL → 503   │  + JSON details
└───────────────────┘
```

**响应格式**:

```json
{
  "status": "degraded",
  "timestamp": "2026-05-19T10:00:00Z",
  "checks": {
    "postgresql": { "status": "ok", "latency_ms": 3 },
    "redis": { "status": "ok", "latency_ms": 1 },
    "downstream_account": { "status": "fail", "error": "connection refused" }
  }
}
```

**关键代码路径**:
- `pkg/health/health.go` — 共享健康检查框架，提供 `Checker` 接口和 `HealthHandler`
- `pkg/health/postgres.go` — PostgreSQL 检查器（`SELECT 1`）
- `pkg/health/redis.go` — Redis 检查器（`PING`）
- `pkg/health/http_downstream.go` — 下游 HTTP 服务检查器
- 各服务 `cmd/main.go` — 注册健康检查器

**各服务依赖矩阵**:

| 服务 | PG | Redis | 下游 |
|------|----|-------|------|
| api-gateway | — | — | account, auth |
| account-service | ✅ | ✅ | auth, credit, compliance |
| auth-service | ✅ | ✅ | — |
| notification-service | ✅ | ✅ | — |
| credit-service | ✅ | ✅ | account |
| compliance-service | ✅ | ✅ | — |
| data-product-service | ✅ | ✅ | account, credit |
| config-service | ✅ | — | — |
| payment-service | ✅ | ✅ | account, notification |

**Prometheus 指标**:
- `health_check_status{service,dependency}` — 1=healthy, 0=unhealthy

**测试策略**:
- 单元测试：模拟各依赖故障（mock DB/Redis/HTTP），验证 200/503 响应
- 集成测试：停止 Redis 容器，验证健康检查返回 503 + Redis 失败详情
- 超时测试：模拟慢查询，验证 5s 内返回

---

#### 3.2.3 UX-01: 一键登录（微信/Apple/Google）

**需求关联**: PRD 3.2 UX-01  
**依赖**: FN-15（OAuth 社交登录扩展）

**技术方案**:

基于 FN-15 的 `OAuthProvider` 接口，实现微信/Apple/Google 一键登录，复用 auth-service 已有统一认证框架。

```
┌──────────┐  1. 社交SDK授权  ┌──────────────┐  2. authorization_code  ┌──────────────┐
│  移动端   │ ──────────────→ │  社交平台     │ ─────────────────────→  │  auth-service │
│  iOS/     │                 │  WeChat/Apple │                         │              │
│  Android  │ ←────────────── │  /Google      │ ←────────────────────── │  3. 换取      │
│           │  auth_code      │               │  user_info+token        │  access_token │
└──────────┘                  └──────────────┘                         │  + 用户信息    │
                                                                       └──────┬───────┘
                                                                              │
                                                                     4. 查找/创建本地用户
                                                                     5. 签发 JWT
                                                                              │
                                                                              ▼
                                                                     ┌──────────────┐
                                                                     │  已有账号？   │
                                                                     │  YES→合并绑定 │
                                                                     │  NO →自动注册 │
                                                                     └──────────────┘
```

**关键代码路径**:
- `auth-service/internal/provider/wechat_oauth.go` — 微信 OAuth provider（移动端）
- `auth-service/internal/provider/apple_signin.go` — Apple Sign-In provider
- `auth-service/internal/provider/google_onetap.go` — Google One Tap provider
- `auth-service/internal/handler/social_login_handler.go` — 统一社交登录 handler
- `auth-service/internal/service/social_bind_service.go` — 社交账号绑定/合并逻辑
- `auth-service/internal/repository/social_account.go` — 社交账号数据层
- `android/.../ui/login/LoginViewModel.kt` — Android 一键登录 UI
- `ios/AccountCenter/Features/Login/LoginView.swift` — iOS 一键登录 UI

**数据库变更**:
- 新增 `social_accounts` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | 主键 |
| user_id | UUID | NOT NULL, FK → users.id | 关联本地用户 |
| provider | VARCHAR(20) | NOT NULL | wechat/apple/google |
| provider_user_id | VARCHAR(128) | NOT NULL | 社交平台用户 ID |
| union_id | VARCHAR(128) | | 微信 UnionID |
| access_token_encrypted | TEXT | | 加密存储的社交 token |
| refresh_token_encrypted | TEXT | | 加密存储的刷新 token |
| expires_at | TIMESTAMP | | token 过期时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 绑定时间 |

- 索引: `UNIQUE(provider, provider_user_id)`, `INDEX(user_id)`

**API 变更**:
- `POST /api/v1/auth/social/login` — 社交登录（provider + code/id_token）
- `POST /api/v1/auth/social/bind` — 绑定社交账号（需已登录）
- `DELETE /api/v1/auth/social/unbind/{provider}` — 解绑社交账号
- `GET /api/v1/auth/social/accounts` — 查询已绑定的社交账号列表

**测试策略**:
- 单元测试：各 OAuthProvider 实现（mock 社交平台 HTTP）、账号合并逻辑、冲突处理
- 集成测试：使用社交平台沙箱环境完成完整登录流程
- 安全测试：伪造 callback 验证失败、token 过期、重复绑定

---

#### 3.2.4 UX-02: 生物识别快捷登录

**需求关联**: PRD 3.2 UX-02  
**依赖**: UX-01（一键登录框架提供设备绑定基础）

**技术方案**:

在移动端实现 Face ID/Touch ID（iOS）和指纹识别（Android），利用 auth-service 已有的 `DeviceFingerprintService` 扩展生物识别 Token 验证。

```
┌──────────┐  1. 用户点击     ┌───────────────┐
│  移动端   │  "生物识别登录"  │  本地生物识别   │
│          │ ──────────────→ │  Face ID/指纹  │
│          │                  └───────┬───────┘
│          │                          │ 认证成功
│          │  2. 取出 device_token    │
│          │ ─────────────────────────┘
│          │
│          │  3. POST /auth/biometric/login
│          │ ────────────────→ ┌──────────────────┐
│          │                   │  auth-service     │
│          │ ←──────────────── │                   │
│          │  4. JWT pair      │  验证 device_token │
│          │                   │  检查设备指纹      │
│          │                   │  检查 token 有效期  │
│          │                   └──────────────────┘
└──────────┘
```

**设备 Token 安全机制**:
- 首次登录成功后，服务端生成 `device_token`（随机 256-bit，AES-256-GCM 加密存储）
- `device_token` 绑定 `device_fingerprint_id`，服务端校验一致性
- Token 有效期 90 天，过期后回退至密码/验证码登录
- 客户端使用 iOS Keychain / Android EncryptedSharedPreferences 存储

**关键代码路径**:
- `auth-service/internal/handler/biometric_handler.go` — 生物识别登录 handler
- `auth-service/internal/service/biometric_service.go` — 设备 token 生成/验证/刷新
- `auth-service/internal/repository/biometric_token.go` — 设备 token 数据层
- `ios/AccountCenter/Features/Login/BiometricAuth.swift` — iOS Face ID/Touch ID 调用
- `android/.../ui/login/BiometricLogin.kt` — Android 指纹识别
- `android/.../storage/BiometricTokenManager.kt` — Android 生物识别 token 管理

**数据库变更**:
- 新增 `biometric_device_tokens` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | 主键 |
| user_id | UUID | NOT NULL, FK → users.id | 用户 ID |
| device_fingerprint_id | VARCHAR(128) | NOT NULL | 设备指纹 ID |
| token_hash | VARCHAR(64) | NOT NULL | 设备 token 哈希（SHA-256） |
| expires_at | TIMESTAMP | NOT NULL | Token 过期时间 |
| last_used_at | TIMESTAMP | | 最后使用时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |

- 索引: `UNIQUE(user_id, device_fingerprint_id)`, `INDEX(expires_at)`

**API 变更**:
- `POST /api/v1/auth/biometric/enroll` — 启用生物识别（绑定设备 token）
- `POST /api/v1/auth/biometric/login` — 生物识别登录（device_token + device_fingerprint）
- `DELETE /api/v1/auth/biometric/disable` — 关闭生物识别
- `POST /api/v1/auth/biometric/refresh` — 刷新设备 token

**测试策略**:
- 单元测试：token 生成/验证/过期逻辑、设备指纹匹配、token 刷新
- 移动端测试：iOS Face ID/Touch ID 回调处理、Android BiometricPrompt
- 安全测试：伪造 device_token、跨设备使用、过期 token 重放

---

#### 3.2.5 UX-05: 个性化仪表盘

**需求关联**: PRD 3.2 UX-05

**技术方案**:

基于 config-service 的配置驱动机制，为不同等级用户动态下发仪表盘卡片布局，客户端按配置渲染。

```
┌──────────┐  GET /dashboard/config  ┌──────────────┐  读取配置  ┌──────────────┐
│  客户端   │ ──────────────────────→ │  account     │ ────────→ │  config      │
│  (四端)   │                         │  -service    │           │  -service    │
│          │ ←────────────────────── │              │ ←──────── │              │
│          │  cards[] + order        │  补充用户数据  │           │  dashboard_  │
│          │                         └──────────────┘           │  layout_*    │
└──────────┘                                                     └──────────────┘
```

**配置数据结构**（config-service 中 `dashboard_layout_{level}` 键）:

```json
{
  "level": "L0",
  "cards": [
    {"id": "upgrade_guide", "type": "cta", "priority": 1, "title": "完成实名认证"},
    {"id": "level_benefits", "type": "info", "priority": 2, "title": "等级权益说明"},
    {"id": "quick_actions", "type": "action", "priority": 3}
  ],
  "layout": "scroll"
}
```

**按等级卡片策略**:

| 等级 | 核心卡片 | 目的 |
|------|---------|------|
| L0 | 升级引导、实名入口、等级权益说明 | 引导转化 |
| L1 | 积分概览、推荐入口、订阅推荐 | 付费转化 |
| L2+ | 积分余额/使用记录、权益使用情况 | 价值感知 |
| L4 | 专属服务入口、优先客服、定制推荐 | VIP 体验 |

**关键代码路径**:
- `account-service/internal/handler/dashboard.go` — 仪表盘配置 API handler
- `account-service/internal/service/dashboard.go` — 按等级组装卡片数据
- `config-service` 新增配置键: `dashboard_layout_L0` ~ `dashboard_layout_L4`
- `web-ui/src/views/Dashboard.vue` — 修改为配置驱动渲染
- `ios/AccountCenter/Features/Home/HomeView.swift` — iOS 仪表盘配置化
- `android/.../ui/home/HomeScreen.kt` — Android 仪表盘配置化

**API 变更**:
- `GET /api/v1/dashboard/config` — 获取当前用户仪表盘配置（含卡片数据和布局）
- `GET /api/v1/admin/dashboard/layouts` — 管理端查看/编辑各等级布局配置

**测试策略**:
- 单元测试：各等级卡片排序逻辑、配置缺失时默认兜底
- 前端测试：各等级渲染正确卡片、运营修改配置后客户端即时响应
- 跨端一致性：Web/iOS/Android 三端仪表盘卡片内容和顺序一致

---

#### 3.2.6 UX-09: 支付流程闭环

**需求关联**: PRD 3.2 UX-09

**技术方案**:

在 FN-01（支付网关集成）基础上，补齐支付结果页、电子发票、失败重试和异常订单自动修复的用户体验闭环。

```
┌──────────┐  1. 创建支付    ┌──────────────┐  2. 唤起收银台  ┌──────────────┐
│  前端     │ ─────────────→ │  payment     │ ─────────────→ │  微信/支付宝  │
│          │                │  -service    │                │              │
│          │                └──────┬───────┘                └──────┬───────┘
│          │                       │ 3. 回调                       │
│          │                       ▼                               │
│          │                ┌──────────────┐                       │
│          │                │  callback    │ ←─────────────────────┘
│          │                │  handler     │
│          │                │  验签+幂等   │
│          │                └──────┬───────┘
│          │                       │
│          │  4. 前端轮询结果       │
│          │ ←─────────────────────┘
│          │  5. 支付结果页          │
│          │  (订单号/金额/发票入口)  │
└──────────┘
```

**支付结果页设计**:
- 成功页：订单号、支付金额、支付方式、订阅生效时间、"申请电子发票"按钮
- 失败页：错误原因分类展示（余额不足/网络超时/用户取消/系统异常）、"一键重试"按钮
- 前端轮询：支付创建后每 2s 轮询 `GET /api/v1/payments/{id}`，最多 30 次（60s）

**电子发票**:
- payment-service 新增发票模块，对接第三方电子发票平台
- 用户申请后异步生成，完成后推送通知（APNs/FCM + 邮件）

**异常订单自动修复**:
- Asynq 定时任务每小时扫描：`status=pending` 且 `created_at > 30min` 的订单
- 主动查询支付平台订单状态，若平台显示已支付则更新为 `paid` 并发放权益
- 修复记录写入 `reconciliation_results` 表

**关键代码路径**:
- `payment-service/internal/handler/payment_result.go` — 支付结果查询 API
- `payment-service/internal/service/invoice.go` — 电子发票服务
- `payment-service/internal/worker/payment_timeout.go` — Asynq 定时任务，异常订单自动修复
- `web-ui/src/views/PaymentResult.vue` — Web 支付结果页
- `ios/AccountCenter/Features/Subscription/PaymentResultView.swift` — iOS
- `android/.../ui/subscription/PaymentResultScreen.kt` — Android

**API 变更**:
- `GET /api/v1/payments/{id}/result` — 支付结果详情（含发票状态）
- `POST /api/v1/payments/{id}/retry` — 重新发起支付
- `POST /api/v1/invoices/apply` — 申请电子发票
- `GET /api/v1/invoices/{id}` — 查询发票状态/下载

**测试策略**:
- 单元测试：支付状态轮询逻辑、发票生成流程、异常订单修复条件判断
- 集成测试：支付→回调丢失→自动修复→权益发放完整流程
- UX 测试：各端支付结果页展示、重试流程、发票申请

---

#### 3.2.7 UX-10: 升降级体验优化

**需求关联**: PRD 3.2 UX-10  
**依赖**: UX-09（支付流程闭环）

**技术方案**:

在 account-service 的订阅服务中新增费用预览计算和降级挽留逻辑，客户端以弹窗形式呈现，无页面跳转。

```
升级流程:
┌──────────┐  选择新等级   ┌──────────────────────────┐
│  用户     │ ──────────→ │  费用预览计算器             │
│          │              │                            │
│          │ ←─────────── │  当前等级剩余价值:          │
│          │  展示预览     │    = (剩余天数/总天数)×月费  │
│          │              │  实际应付:                   │
│          │              │    = 新等级月费 - 剩余价值    │
│          │              │  积分可抵扣: xxx            │
│          │              │  "升级后新权益立即生效"      │
│          │              └────────────┬───────────────┘
│          │                           │ 确认升级
│          │ ←──────────────────────── │ ≤5s 内生效
└──────────┘              立即生效 + 成功动画

降级流程:
┌──────────┐  选择降级    ┌──────────────────────────┐
│  用户     │ ──────────→ │  降级确认弹窗              │
│          │              │                            │
│          │ ←─────────── │  "降级将在当前周期结束后    │
│          │  挽留弹窗     │   生效（2026-06-15）"      │
│          │              │                            │
│          │              │  [推荐：升级至 L3 享 8 折]  │
│          │              │  [继续降级] [取消]          │
└──────────┘              └──────────────────────────┘
```

**关键代码路径**:
- `account-service/internal/service/subscription_upgrade.go` — 升级费用计算 + 立即生效逻辑
- `account-service/internal/service/subscription_downgrade.go` — 降级预览 + 挽留方案
- `account-service/internal/handler/subscription_handler.go` — 新增预览/确认 API
- `web-ui/src/components/UpgradePreviewDialog.vue` — Web 升级预览弹窗
- `web-ui/src/components/DowngradeRetentionDialog.vue` — Web 降级挽留弹窗
- `ios/AccountCenter/Features/Subscription/UpgradePreviewView.swift` — iOS
- `android/.../ui/subscription/UpgradePreviewDialog.kt` — Android

**API 变更**:
- `POST /api/v1/subscriptions/upgrade/preview` — 升级费用预览（当前等级、目标等级、积分抵扣数）
- `POST /api/v1/subscriptions/upgrade/confirm` — 确认升级（创建支付订单）
- `POST /api/v1/subscriptions/downgrade/preview` — 降级预览（生效日期、权益变化）
- `POST /api/v1/subscriptions/downgrade/confirm` — 确认降级（下期生效）

**数据库变更**:
- `subscriptions` 表新增 `pending_downgrade_to` VARCHAR(5) — 待降级目标等级
- `subscriptions` 表新增 `pending_downgrade_effective_at` TIMESTAMP — 降级生效时间

**测试策略**:
- 单元测试：费用计算（剩余天数折算、积分抵扣、四舍五入）、降级生效日期计算
- 集成测试：升级→立即生效→权益刷新（≤5s）、降级→到期自动降级
- UX 测试：弹窗交互步骤 ≤3、四端一致性

---

#### 3.2.8 UX-11: 订阅续费提醒

**需求关联**: PRD 3.2 UX-11  
**依赖**: FN-12（事件埋点 SDK）

**技术方案**:

使用 Asynq 定时任务扫描即将到期订阅用户，通过多通道（Push/SMS/Email）发送续费提醒，附带深度链接直达支付页。

```
┌──────────────────────────────────────┐
│  Asynq Scheduler (每日 09:00 执行)    │
│                                       │
│  1. 查询 T-7/T-3/T-1 即将到期用户     │
│  2. 去重检查（Redis SET 防重复推送）   │
│  3. 按 T-N 模板组装消息               │
│  4. 分发到 notification-service       │
│     - Push (APNs/FCM/HMS)            │
│     - SMS                             │
│     - Email                           │
│  5. 记录推送事件到埋点系统             │
└──────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────┐
│  用户收到提醒                          │
│  "您的 Premium 将在 3 天后到期"        │
│  [一键续费] ← 深度链接直达支付页       │
└──────────────────────────────────────┘
```

**去重机制**:
- Redis Key: `renewal_reminder:{userID}:{T-N}:{date}`，TTL 48h
- 同一用户同一天同一 T-N 只发送一次

**消息模板**（config-service 管理）:

| 触发时间 | 模板 ID | 通道优先级 | 内容要点 |
|---------|---------|----------|---------|
| T-7 | `renewal_t7` | Email > Push | 到期日期、续费优惠、深度链接 |
| T-3 | `renewal_t3` | Push > SMS | 紧急提醒、一键续费链接 |
| T-1 | `renewal_t1` | SMS > Push > Email | 最后提醒、即时续费 |

**关键代码路径**:
- `account-service/internal/worker/renewal_reminder.go` — Asynq 定时任务，扫描到期用户
- `account-service/internal/service/renewal_reminder.go` — 续费提醒业务逻辑
- `notification-service/internal/handler/template.go` — 消息模板管理
- `config-service` 新增配置键: `renewal_reminder_templates`

**API 变更**:
- `GET /api/v1/subscriptions/renewal/status` — 查询当前续费提醒偏好
- `PUT /api/v1/subscriptions/renewal/preferences` — 设置提醒通道偏好

**数据库变更**:
- 新增 `renewal_reminder_logs` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | 主键 |
| user_id | UUID | NOT NULL | 用户 ID |
| subscription_id | UUID | NOT NULL | 订阅 ID |
| reminder_type | VARCHAR(5) | NOT NULL | T-7/T-3/T-1 |
| channel | VARCHAR(10) | NOT NULL | push/sms/email |
| sent_at | TIMESTAMP | NOT NULL | 发送时间 |
| status | VARCHAR(10) | NOT NULL | sent/failed |

**埋点事件**:
- `renewal_reminder_sent` — 提醒发送成功
- `renewal_reminder_clicked` — 用户点击续费链接
- `renewal_completed` — 续费完成

**测试策略**:
- 单元测试：T-7/T-3/T-1 日期计算、去重逻辑、通道选择
- 集成测试：完整流程（扫描→去重→推送→点击→续费）
- Asynq 任务测试：验证定时调度和失败重试

---

#### 3.2.9 UX-12: 推荐进度可视化

**需求关联**: PRD 3.2 UX-12  
**依赖**: FN-12（事件埋点 SDK）

**技术方案**:

在 data-product-service 中新增推荐漏斗聚合查询 API，前端渲染漏斗图和收益趋势图。数据从缓存读取，保证页面加载 ≤2s。

```
┌─────────────────────────────────────────────┐
│  推荐漏斗数据聚合                              │
│                                              │
│  分享数 (share_count)                         │
│    ↓ 转化率                                   │
│  注册数 (register_count)                      │
│    ↓ 转化率                                   │
│  实名认证数 (verify_count)                    │
│    ↓ 转化率                                   │
│  付费转化数 (pay_count)                       │
│                                              │
│  收益趋势: 近30天 待结算 / 已结算 (按日聚合)    │
└─────────────────────────────────────────────┘
```

**数据预计算策略**:
- Asynq 定时任务每小时聚合推荐数据，写入 Redis 缓存
- Redis Key: `referral_funnel:{userID}` TTL 2h，`referral_revenue_trend:{userID}` TTL 2h
- 缓存未命中时降级为实时查询（限制 QPS）

**关键代码路径**:
- `account-service/internal/handler/referral.go` — 推荐进度 API handler
- `account-service/internal/service/referral_funnel.go` — 推荐漏斗聚合逻辑
- `data-product-service/internal/service/referral_aggregator.go` — 推荐数据预计算
- `web-ui/src/views/Referral.vue` — 修改为漏斗图 + 收益趋势图
- `ios/AccountCenter/Features/Referral/ReferralFunnelView.swift` — iOS
- `android/.../ui/referral/ReferralFunnelScreen.kt` — Android

**API 变更**:
- `GET /api/v1/referrals/funnel` — 推荐漏斗数据（share→register→verify→pay 各阶段数量和转化率）
- `GET /api/v1/referrals/revenue-trend` — 收益趋势（近 30 天按日/周/月）
- `GET /api/v1/referrals/records` — 推荐记录列表（含被邀请人状态和奖励发放状态）

**测试策略**:
- 单元测试：漏斗数据聚合、转化率计算、收益趋势计算
- 性能测试：页面加载时间 ≤2s（缓存命中场景）
- 前端测试：漏斗图渲染、趋势图交互、空数据状态

---

#### 3.2.10 FN-04: 退款流程

**需求关联**: PRD 3.2 FN-04  
**依赖**: UX-09（支付流程闭环）

**技术方案**:

在 payment-service 中实现退款流程，包含退款策略计算、自动/人工审核、原路退款和积分扣回。

```
┌──────────┐  申请退款    ┌──────────────────────────────────┐
│  用户     │ ──────────→ │  payment-service                  │
│          │              │                                    │
│          │              │  1. 退款策略判定:                   │
│          │              │     ≤7天 → 全额退款               │
│          │              │     >7天 → 按剩余天数比例退款       │
│          │              │                                    │
│          │              │  2. 审核流:                        │
│          │              │     自动通过: ≤7天+未使用高级权益   │
│          │              │     人工审核: 其他情况 (≤48h)       │
│          │              │                                    │
│          │              │  3. 原路退款:                      │
│          │              │     wechat → wechat Refund API     │
│          │              │     alipay → alipay Refund API     │
│          │              │                                    │
│          │              │  4. 后置处理:                      │
│          │              │     - 积分抵扣部分等额扣回          │
│          │              │     - 订阅状态立即降级              │
│          │              │     - 写入审计日志                  │
└──────────┘              └──────────────────────────────────┘
```

**退款金额计算**:
- 月付: `refund = actual_amount × (remaining_days / total_days)`
- 年付: `refund = actual_amount × (remaining_months / total_months)`
- 积分抵扣部分: 不退还积分，从退款金额中扣除

**关键代码路径**:
- `payment-service/internal/handler/refund.go` — 退款 API handler
- `payment-service/internal/service/refund.go` — 退款业务逻辑（策略+审核+退款调用）
- `payment-service/internal/service/refund_calculator.go` — 退款金额计算
- `payment-service/internal/repository/refund.go` — 退款数据层
- `account-service/internal/service/subscription.go` — 退款后降级逻辑

**数据库变更**:
- 新增 `refunds` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 退款 ID |
| order_id | UUID | NOT NULL, FK → orders.id | 关联订单 |
| user_id | UUID | NOT NULL | 用户 ID |
| refund_amount_cents | INTEGER | NOT NULL | 退款金额（分） |
| credits_deducted | INTEGER | NOT NULL | 扣回积分数 |
| reason | VARCHAR(200) | NOT NULL | 退款原因 |
| status | VARCHAR(20) | NOT NULL | pending_auto/pending_manual/approved/rejected/refunded/failed |
| reviewed_by | UUID | | 审核人 ID（人工审核） |
| reviewed_at | TIMESTAMP | | 审核时间 |
| refunded_at | TIMESTAMP | | 退款到账时间 |
| refund_method | VARCHAR(20) | | wechat/alipay |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | |

- 索引: `INDEX(user_id)`, `INDEX(order_id)`, `INDEX(status)`

**API 变更**:
- `POST /api/v1/refunds` — 申请退款
- `GET /api/v1/refunds/{id}` — 查询退款状态
- `POST /api/v1/admin/refunds/{id}/approve` — 审核通过（财务管理员）
- `POST /api/v1/admin/refunds/{id}/reject` — 审核驳回

**测试策略**:
- 单元测试：退款金额计算（全额/比例/积分扣回）、审核策略判定、状态机流转
- 集成测试：申请→自动审核→原路退款→积分扣回→降级完整流程
- 边界测试：7 天临界点、零退款、重复退款请求

---

#### 3.2.11 FN-06: 运营数据大屏

**需求关联**: PRD 3.2 FN-06  
**依赖**: AR-06（Grafana 仪表盘）

**技术方案**:

基于 Grafana 13.x + VictoriaMetrics 构建运营数据大屏，数据源为 data-product-service 聚合查询 + Prometheus 指标。

```
┌─────────────────────────────────────────────────┐
│  Grafana Dashboard: "运营数据大屏"                │
│                                                   │
│  ┌──────────────┐  ┌──────────────┐              │
│  │ 注册趋势      │  │ 付费转化漏斗  │              │
│  │ (日/周/月)    │  │ 访问→注册→    │              │
│  │              │  │ 实名→首次付费  │              │
│  └──────────────┘  └──────────────┘              │
│  ┌──────────────┐  ┌──────────────┐              │
│  │ MRR/ARR      │  │ RFM 分布     │              │
│  │ 收入趋势      │  │ 散点图/热力图 │              │
│  └──────────────┘  └──────────────┘              │
│  ┌──────────────┐  ┌──────────────┐              │
│  │ 推荐 K-factor│  │ 渠道分布      │              │
│  └──────────────┘  └──────────────┘              │
│                                                   │
│  筛选: [日期范围] [渠道] [用户等级]                 │
└─────────────────────────────────────────────────┘
```

**关键代码路径**:
- `monitoring/grafana/dashboards/operations_dashboard.json` — Grafana Dashboard JSON 模板
- `data-product-service/internal/handler/metrics.go` — 运营指标 API
- `data-product-service/internal/service/registration_trend.go` — 注册趋势聚合
- `data-product-service/internal/service/conversion_funnel.go` — 付费转化漏斗
- `data-product-service/internal/service/revenue_metrics.go` — MRR/ARR 计算
- `data-product-service/internal/service/rfm_distribution.go` — RFM 分布查询
- `data-product-service/internal/service/kfactor.go` — 推荐 K-factor 计算

**API 变更**:
- `GET /api/v1/admin/metrics/registration-trend` — 注册趋势
- `GET /api/v1/admin/metrics/conversion-funnel` — 付费转化漏斗
- `GET /api/v1/admin/metrics/revenue` — MRR/ARR
- `GET /api/v1/admin/metrics/rfm-distribution` — RFM 分布
- `GET /api/v1/admin/metrics/kfactor` — 推荐 K-factor

**测试策略**:
- 单元测试：MRR/ARR 计算、K-factor 计算、漏斗聚合
- 集成测试：Grafana Dashboard 自动导入验证
- 性能测试：大屏加载时间 ≤5s（数据刷新间隔 5 分钟）

---

#### 3.2.12 FN-07: 订阅管理后台

**需求关联**: PRD 3.2 FN-07

**技术方案**:

在 account-service 中扩展 Admin API，新增订阅套餐 CRUD、优惠券管理和促销活动管理模块。

```
┌──────────────────┐  Admin JWT   ┌──────────────┐  ┌──────────────────┐
│  Admin Frontend  │ ───────────→ │  api-gateway  │→ │  account-service │
│  (config-mgmt-ui)│              │  /admin/*     │  │  Admin Module    │
└──────────────────┘              └──────────────┘  │                  │
                                                     │  Plan CRUD       │
                                                     │  Coupon CRUD     │
                                                     │  Promotion CRUD  │
                                                     └──────────────────┘
```

**套餐管理**:
- CRUD 操作影响 config-service 中的 `pricing_plans` 配置项
- 变更需审批流（草稿→审核→生效）

**优惠券系统**:

| 类型 | 说明 |
|------|------|
| 百分比折扣 | 如 8 折（20% off） |
| 固定金额 | 如减 10 元 |
| 免首月 | 首月免费 |

- 优惠券 code: 唯一字符串，支持批量生成
- 使用限制: 总量、每人限用次数、有效期

**关键代码路径**:
- `account-service/internal/handler/admin_plan.go` — 套餐管理 Admin handler
- `account-service/internal/handler/admin_coupon.go` — 优惠券 Admin handler
- `account-service/internal/handler/admin_promotion.go` — 促销活动 Admin handler
- `account-service/internal/service/coupon.go` — 优惠券业务逻辑（生成、核销、过期）
- `account-service/internal/service/promotion.go` — 促销活动业务逻辑
- `account-service/internal/repository/coupon.go` — 优惠券数据层
- `account-service/internal/repository/promotion.go` — 促销数据层

**数据库变更**:
- 新增 `coupons` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 优惠券 ID |
| code | VARCHAR(32) | UNIQUE, NOT NULL | 优惠券码 |
| type | VARCHAR(20) | NOT NULL | percentage/fixed/free_first_month |
| value | INTEGER | NOT NULL | 折扣值（百分比/固定金额分） |
| applicable_plans | TEXT[] | | 适用套餐 ID 列表 |
| max_uses | INTEGER | | 总量限制（NULL 无限） |
| max_uses_per_user | INTEGER | DEFAULT 1 | 每人限用次数 |
| used_count | INTEGER | DEFAULT 0 | 已使用次数 |
| valid_from | TIMESTAMP | NOT NULL | 生效时间 |
| valid_until | TIMESTAMP | NOT NULL | 过期时间 |
| created_by | UUID | NOT NULL | 创建人 |
| created_at | TIMESTAMP | NOT NULL | |

- 新增 `coupon_usages` 表: `id, coupon_id, user_id, order_id, used_at`
- 新增 `promotions` 表: `id, name, type, config(JSONB), status, valid_from, valid_until, created_by, created_at`

**API 变更**:
- `POST/GET/PUT/DELETE /api/v1/admin/plans` — 套餐 CRUD
- `POST/GET/PUT/DELETE /api/v1/admin/coupons` — 优惠券 CRUD
- `POST /api/v1/admin/coupons/batch-generate` — 批量生成优惠券
- `POST /api/v1/admin/coupons/{id}/invalidate` — 作废优惠券
- `POST/GET/PUT/DELETE /api/v1/admin/promotions` — 促销活动 CRUD
- `POST /api/v1/coupons/validate` — 前端验证优惠券（用户侧）

**测试策略**:
- 单元测试：优惠券生成/核销/过期、促销活动状态管理、套餐变更审批流
- 集成测试：创建优惠券→用户下单使用→核销→限额验证
- 权限测试：非 Admin 角色无法访问管理 API

---

#### 3.2.13 FN-08: 风控管理后台

**需求关联**: PRD 3.2 FN-08

**技术方案**:

在 compliance-service 中新增风控管理 API，实现黑名单管理、风险事件记录和异常注册预警。

```
┌──────────────────┐  Admin JWT   ┌──────────────┐  ┌───────────────────┐
│  Admin Frontend  │ ───────────→ │  api-gateway  │→ │  compliance       │
│                  │              │  /admin/*     │  │  -service         │
└──────────────────┘              └──────────────┘  │                   │
                                                     │  Blacklist CRUD   │
                                                     │  Risk Events      │
                                                     │  Auto-Alert       │
                                                     └───────────────────┘
                                                              │
                                                     ┌────────▼────────┐
                                                     │  Redis          │
                                                     │  blacklist:ip:* │
                                                     │  blacklist:     │
                                                     │  device:*       │
                                                     │  blacklist:     │
                                                     │  user:*         │
                                                     └─────────────────┘
```

**黑名单存储**:
- 高频查询使用 Redis SET + PostgreSQL 持久化双写
- Redis Key: `blacklist:ip:{ip}`, `blacklist:device:{fingerprint}`, `blacklist:user:{userID}`
- 支持过期时间（临时封禁）

**异常注册检测**:
- 基于 Redis 计数器: `register_count:{ip}:{date}`, `register_count:{device}:{date}`
- 24h 内同一 IP/设备注册超 5 次 → 自动标记 `suspicious`
- 标记后写入 `risk_events` 表并推送告警

**关键代码路径**:
- `compliance-service/internal/handler/admin_risk.go` — 风控管理 Admin handler
- `compliance-service/internal/service/blacklist.go` — 黑名单业务逻辑
- `compliance-service/internal/service/risk_detector.go` — 风险检测引擎
- `compliance-service/internal/repository/blacklist.go` — 黑名单数据层
- `compliance-service/internal/repository/risk_event.go` — 风险事件数据层
- `api-gateway/internal/middleware/blacklist.go` — 网关层黑名单检查中间件

**数据库变更**:
- 新增 `blacklist_entries` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PK | 主键 |
| type | VARCHAR(10) | NOT NULL | ip/device/user |
| value | VARCHAR(128) | NOT NULL | IP/设备指纹/用户 ID |
| reason | VARCHAR(200) | NOT NULL | 加黑原因 |
| expires_at | TIMESTAMP | | 过期时间（NULL 永久） |
| created_by | UUID | NOT NULL | 操作人 |
| created_at | TIMESTAMP | NOT NULL | |

- 索引: `UNIQUE(type, value)`, `INDEX(expires_at)`
- 新增 `risk_events` 表: `id, type, severity(high/medium/low), source_ip, device_fingerprint, user_id, details(JSONB), resolved, resolved_by, created_at`

**API 变更**:
- `POST/GET/DELETE /api/v1/admin/blacklist` — 黑名单 CRUD
- `POST /api/v1/admin/blacklist/batch` — 批量导入黑名单
- `GET /api/v1/admin/risk-events` — 风险事件列表（分页、过滤）
- `PUT /api/v1/admin/risk-events/{id}/resolve` — 处置风险事件
- `GET /api/v1/admin/risk-events/suspicious-registrations` — 可疑注册列表

**测试策略**:
- 单元测试：黑名单增删查、过期清理、Redis+PG 双写一致性
- 集成测试：注册→触发异常检测→标记可疑→管理员封禁→网关拦截
- 性能测试：黑名单检查延迟 ≤1ms（Redis 命中）

---

#### 3.2.14 FN-12: 事件埋点 SDK

**需求关联**: PRD 3.2 FN-12

**技术方案**:

开发轻量级事件采集 SDK，覆盖 Web（TypeScript）、iOS（Swift）和 Android（Kotlin）三端，自动采集基础事件并支持手动上报业务事件。

```
┌──────────────────────────────────────────────────┐
│  SDK Architecture (三端统一设计)                    │
│                                                    │
│  ┌─────────────┐  ┌─────────────┐                │
│  │ Auto Track  │  │ Manual Track│                │
│  │ page_view   │  │ 14 业务事件  │                │
│  │ click       │  │ subscription │                │
│  │ dwell_time  │  │ _purchased   │                │
│  └──────┬──────┘  └──────┬──────┘                │
│         │                │                       │
│         ▼                ▼                       │
│  ┌──────────────────────────────────────────┐     │
│  │  Event Queue (本地缓存)                    │     │
│  │  - batch 累积（最多 20 条或 5s）           │     │
│  │  - 离线缓存（SQLite/UserDefaults）        │     │
│  └──────────────────┬───────────────────────┘     │
│                     │ batch send                   │
│                     ▼                             │
│  ┌──────────────────────────────────────────┐     │
│  │  data-product-service                     │     │
│  │  POST /api/v1/events/batch                │     │
│  └──────────────────────────────────────────┘     │
└──────────────────────────────────────────────────┘
```

**SDK 包结构**:

```
# Web SDK (TypeScript)
sdks/web/
├── src/
│   ├── index.ts            # 入口，初始化
│   ├── tracker.ts          # 核心追踪器
│   ├── auto-track.ts       # 自动采集 (page_view, click, dwell)
│   ├── event-queue.ts      # 事件队列 + batch 发送
│   ├── storage.ts          # localStorage 离线缓存
│   └── types.ts            # 事件类型定义
├── package.json
└── tsconfig.json

# iOS SDK (Swift)
sdks/ios/
├── Sources/
│   ├── NeuroTracker.swift        # 入口
│   ├── AutoTracker.swift         # 自动采集
│   ├── EventQueue.swift          # 队列 + batch
│   └── PersistentStorage.swift   # CoreData 离线缓存
└── Package.swift

# Android SDK (Kotlin)
sdks/android/
├── src/main/java/com/neuro/tracker/
│   ├── NeuroTracker.kt           # 入口
│   ├── AutoTracker.kt            # 自动采集
│   ├── EventQueue.kt             # 队列 + batch
│   └── PersistentStorage.kt      # Room 离线缓存
└── build.gradle.kts
```

**14 个业务事件**: `subscription_purchased`, `subscription_cancelled`, `credit_earned`, `credit_used`, `referral_shared`, `referral_registered`, `level_upgraded`, `level_downgraded`, `renewal_reminder_sent`, `renewal_reminder_clicked`, `renewal_completed`, `wx_subscribe_msg_sent`, `miniprogram_share_chat`, `ad_splash_shown`

**关键代码路径**:
- `sdks/web/src/` — Web SDK 全部源码
- `sdks/ios/Sources/` — iOS SDK 全部源码
- `sdks/android/src/` — Android SDK 全部源码
- `data-product-service/internal/handler/event.go` — 事件 batch 接收 API
- `data-product-service/internal/service/event_processor.go` — 事件处理和分发
- `data-product-service/internal/repository/event.go` — 事件存储

**数据库变更**:
- 新增 `events` 表: `id, event_name, user_id, session_id, properties(JSONB), platform, app_version, timestamp, created_at`
- 按月分区：`events_2026_05`, `events_2026_06`, ...
- 索引: `INDEX(user_id, timestamp)`, `INDEX(event_name, timestamp)`

**API 变更**:
- `POST /api/v1/events/batch` — 批量上报事件（一次最多 50 条）

**测试策略**:
- 单元测试：SDK 事件构造、队列 batch 逻辑、离线缓存读写
- 集成测试：SDK→data-product-service 完整上报链路
- 性能测试：SDK 包体积 ≤50KB（Web gzipped）、初始化时间 ≤100ms
- 离线测试：断网→缓存→恢复网络→自动重传

---

#### 3.2.15 FN-15: OAuth 社交登录扩展

**需求关联**: PRD 3.2 FN-15

**技术方案**:

在 auth-service 中设计插件化 OAuth Provider 架构，新增支付宝、Apple ID、Google OAuth provider。

```
┌────────────────────────────────────────────────┐
│  auth-service: OAuth Provider Plugin System     │
│                                                  │
│  ┌────────────────────────────────────────┐     │
│  │  OAuthProvider Interface               │     │
│  │  + GetAuthURL(state) string            │     │
│  │  + ExchangeCode(code) (*UserInfo, err) │     │
│  │  + RefreshToken(rt) error             │     │
│  │  + RevokeToken(token) error           │     │
│  │  + Name() string                       │     │
│  └────────────┬───────────────────────────┘     │
│               │                                  │
│  ┌────────┬───┴───┬──────────┬──────────┐      │
│  │ WeChat │ Alipay │ Apple ID │ Google   │      │
│  │ (已有) │ (新增) │ (新增)   │ (新增)   │      │
│  └────────┴───────┴──────────┴──────────┘      │
│                                                  │
│  Provider Registry: map[string]OAuthProvider     │
│  config-service → oauth_providers 配置项          │
└────────────────────────────────────────────────┘
```

**各 Provider 技术方案**:

| Provider | 协议 | 关键参数 |
|----------|------|---------|
| Alipay | OAuth 2.0 | `app_id`, `private_key`, `alipay_public_key` |
| Apple ID | OAuth 2.0 + OIDC | `client_id`, `team_id`, `key_id`, `.p8` 私钥 |
| Google | OAuth 2.0 + OIDC | `client_id`, `client_secret`, `redirect_uri` |

**关键代码路径**:
- `auth-service/internal/provider/oauth_provider.go` — OAuthProvider 接口定义
- `auth-service/internal/provider/alipay_oauth.go` — 支付宝 OAuth 实现
- `auth-service/internal/provider/apple_oauth.go` — Apple ID OAuth 实现
- `auth-service/internal/provider/google_oauth.go` — Google OAuth 实现
- `auth-service/internal/handler/oauth_handler.go` — 统一 OAuth 回调 handler
- `auth-service/internal/service/oauth_registry.go` — Provider 注册表

**API 变更**:
- `GET /api/v1/auth/oauth/{provider}/url` — 获取授权跳转 URL
- `POST /api/v1/auth/oauth/{provider}/callback` — OAuth 回调处理
- 与 UX-01 的社交登录 API 复用同一套端点

**测试策略**:
- 单元测试：各 Provider `ExchangeCode`（mock HTTP）、用户信息解析、token 刷新
- 集成测试：使用社交平台沙箱完成完整 OAuth 流程
- 插件化测试：新增 mock Provider 验证注册表动态加载

---

#### 3.2.16 FN-17: 数据导出/开放 API

**需求关联**: PRD 3.2 FN-17  
**依赖**: AR-14（KMS 密钥管理）

**技术方案**:

实现三类数据导出能力：PIPL 个人数据导出、运营报表导出、OAuth2 开放 API。

```
┌──────────────────────────────────────────────────────┐
│  数据导出架构                                          │
│                                                        │
│  1. PIPL 个人数据导出:                                  │
│     用户申请 → Asynq 异步生成 → OSS 存储 → 下载链接     │
│                                                        │
│  2. 运营报表导出:                                       │
│     Admin 请求 → data-product-service 生成 CSV/Excel   │
│     → 流式下载 (SSE)                                   │
│                                                        │
│  3. 开放 API:                                          │
│     合作伙伴 → OAuth2 Token → api-gateway 限流          │
│     → 只读查询接口                                      │
└──────────────────────────────────────────────────────┘
```

**PIPL 导出数据范围**:
- 用户基本信息（手机号/邮箱脱敏）、实名信息（脱敏）、积分记录、订阅记录、订单记录、推荐记录
- 格式: JSON + ZIP（AES-256 加密）
- 链接有效期 7 天

**关键代码路径**:
- `data-product-service/internal/handler/data_export.go` — 数据导出 handler
- `data-product-service/internal/service/pipl_export.go` — PIPL 导出逻辑
- `data-product-service/internal/service/report_generator.go` — 报表生成（CSV/Excel）
- `data-product-service/internal/worker/export_worker.go` — Asynq 异步导出任务
- `api-gateway/internal/middleware/oauth2.go` — OAuth2 开放 API 认证中间件

**数据库变更**:
- 新增 `data_export_requests` 表: `id, user_id, type(pipl/report/open_api), status, file_url, file_key_encrypted, expires_at, created_at`
- 新增 `oauth2_clients` 表: `id, client_id, client_secret_hash, name, redirect_uris, scopes, rate_limit, created_at`

**API 变更**:
- `POST /api/v1/data/export` — 申请个人数据导出
- `GET /api/v1/data/export/{id}/download` — 下载导出文件
- `GET /api/v1/admin/reports/{type}/export` — 运营报表导出（CSV/Excel）
- `POST /api/v1/oauth2/token` — 开放 API Token 获取
- `GET /api/v1/open/user/info` — 开放 API：用户信息查询
- `GET /api/v1/open/credits/balance` — 开放 API：积分查询

**测试策略**:
- 单元测试：导出数据脱敏、CSV/Excel 生成、OAuth2 token 验证
- 集成测试：PIPL 导出完整流程、开放 API 限流（100 次/分钟）
- 安全测试：导出文件加密验证、开放 API 越权访问

---

#### 3.2.17 MB-02: Android 字体集成

**需求关联**: PRD 3.2 MB-02

**技术方案**:

将 Inter 和 Space Grotesk 字体文件集成到 Android 项目资源目录，替换系统默认字体。

```
android/app/src/main/res/font/
├── inter_regular.ttf          # Inter Regular (~200KB, subset)
├── inter_medium.ttf           # Inter Medium (~200KB, subset)
├── inter_semibold.ttf         # Inter Semibold (~200KB, subset)
├── space_grotesk_bold.ttf     # Space Grotesk Bold (~180KB, subset)
└── space_grotesk_semibold.ttf # Space Grotesk Semibold (~180KB, subset)
总计 ≤2MB (通过 pyftsubset 裁剪 CJK 以外字符)
```

**字体子集化**:
- 使用 `pyftsubset` 裁剪：保留 Latin-1 + Latin Extended + 常用符号
- 中文字符回退至系统默认字体（Noto Sans CJK）

**关键代码路径**:
- `android/app/src/main/res/font/` — 字体文件目录
- `android/app/src/main/java/com/accountcenter/ui/theme/Type.kt` — 修改 `FontFamily.Default` 为 Inter
- `android/app/src/main/java/com/accountcenter/ui/theme/Theme.kt` — 全局 Typography 配置
- `scripts/font_subset.sh` — 字体子集化脚本

**测试策略**:
- 视觉回归测试：对比字体替换前后各页面截图
- 兼容性测试：Android 5.0+ (API 21+) 各版本渲染一致
- 包体积测试：字体文件总大小 ≤2MB

---

#### 3.2.18 MB-09: Token 安全存储升级验证

**需求关联**: PRD 3.2 MB-09

**技术方案**:

升级移动端 token 存储策略：access_token 仅内存持有，refresh_token AES-256-GCM 加密存储并绑定设备指纹。

```
iOS:
┌─────────────────────────────────────┐
│  Token Storage (iOS)                │
│                                      │
│  access_token:                       │
│    → 内存变量 (AccessTokenManager)   │
│    → 不写入 Keychain/文件            │
│                                      │
│  refresh_token:                      │
│    → AES-256-GCM 加密               │
│    → Keychain kSecAttrAccessible:    │
│      kSecAttrAccessibleWhenUnlocked  │
│    → Key 由 Secure Enclave 保护      │
│                                      │
│  设备绑定:                           │
│    → DeviceCheck DCDevice.token      │
│    → 服务端校验 token+设备一致性      │
└─────────────────────────────────────┘

Android:
┌─────────────────────────────────────┐
│  Token Storage (Android)            │
│                                      │
│  access_token:                       │
│    → 内存缓存 (TokenManager)         │
│    → 不写入 SharedPreferences        │
│                                      │
│  refresh_token:                      │
│    → AES-256-GCM 加密               │
│    → EncryptedSharedPreferences      │
│    → MasterKey 由 Android Keystore   │
│      硬件安全模块保护                 │
│                                      │
│  设备绑定:                           │
│    → SafetyNet Attestation           │
│    → 服务端校验 token+设备一致性      │
└─────────────────────────────────────┘
```

**关键代码路径**:
- `ios/AccountCenter/Core/TokenManager.swift` — iOS token 安全管理器（重写）
- `ios/AccountCenter/Core/DeviceBinding.swift` — iOS DeviceCheck 集成
- `android/.../storage/TokenManager.kt` — Android token 安全管理器（重写）
- `android/.../storage/DeviceBinding.kt` — Android SafetyNet 集成
- `auth-service/internal/handler/device_handler.go` — 设备指纹校验 API

**测试策略**:
- 安全审计：验证 access_token 不出现在 Keychain/SharedPreferences/日志/剪贴板
- 设备绑定测试：同一 token 在不同设备上使用被拒绝
- 边界测试：Token 过期→自动刷新→新 token 加密存储、设备指纹变更→强制重新认证

---

#### 3.2.19 MB-10: 证书固定

**需求关联**: PRD 3.2 MB-10

**技术方案**:

在移动端实现 SSL/TLS 证书固定（Certificate Pinning），防止中间人攻击。

```
iOS (URLSession Server Trust):
┌───────────────────────────────────────┐
│  URLSessionDelegate                    │
│                                        │
│  urlSession(_:didReceive:completion:)  │
│    1. 提取服务器证书公钥               │
│    2. SHA-256 哈希                     │
│    3. 与预置指纹比对                   │
│       - Primary: <cert1_fingerprint>   │
│       - Backup:  <cert2_fingerprint>   │
│    4. 匹配 → .performDefaultHandling   │
│       不匹配 → .cancelAuthentication   │
│                    + 安全事件上报       │
└───────────────────────────────────────┘

Android (OkHttp CertificatePinner):
┌───────────────────────────────────────┐
│  OkHttpClient.Builder()               │
│    .certificatePinner(                │
│      CertificatePinner.Builder()      │
│        .add(hostname, "sha256/xxx")   │
│        .add(hostname, "sha256/yyy")   │
│        .build()                       │
│    )                                  │
│    .build()                           │
│                                        │
│  固定失败 → onConnectionFailed         │
│    → 上报安全事件                      │
│    → 展示 "网络不安全" 提示             │
└───────────────────────────────────────┘
```

**证书指纹配置**:
- 主证书: 当前 SSL 证书公钥 SHA-256 指纹
- 备份证书: 备用 CA 签发的证书公钥 SHA-256 指纹
- 通过 config-service 远程配置: `ssl_pin_hashes` 配置项
- 客户端启动时拉取最新指纹列表，支持证书轮换时无缝切换

**关键代码路径**:
- `ios/AccountCenter/Core/Network/CertificatePinning.swift` — iOS 证书固定实现
- `ios/AccountCenter/Core/Network/SecureURLSession.swift` — 安全 URLSession 封装
- `android/.../network/CertificatePinner.kt` — Android 证书固定配置
- `android/.../network/ApiClient.kt` — 修改为使用固定证书的 OkHttpClient
- `config-service` 新增配置: `ssl_pin_hashes`

**测试策略**:
- 单元测试：证书指纹匹配逻辑、主备切换逻辑
- 安全测试：使用 Charles Proxy 注入自签证书，验证连接被拒绝
- 证书轮换测试：更新 config-service 指纹后客户端自动切换

---

#### 3.2.20 MB-13: 小程序订阅消息

**需求关联**: PRD 3.2 MB-13  
**依赖**: FN-12（事件埋点 SDK）

**技术方案**:

对接微信小程序订阅消息 API，为 4 类事件配置消息模板，由 Asynq 定时任务驱动发送。

```
┌────────────────────────────────────────────────────┐
│  微信订阅消息流程                                    │
│                                                      │
│  1. 用户首次触发事件 → 弹出订阅授权 (wx.request       │
│     SubscribeMessage)                                │
│                                                      │
│  2. 用户授权 → 记录 openID + 模板授权                │
│                                                      │
│  3. 事件触发 → Asynq Task                            │
│     → notification-service                           │
│     → 调用微信 subscribeMessage.send API             │
│     → 失败重试 (3次: 5min/15min/60min)               │
│                                                      │
│  4. 埋点: wx_subscribe_msg_sent / _failed            │
└────────────────────────────────────────────────────┘
```

**4 类事件模板**:

| 事件 | 模板 ID | 触发条件 | 关键数据 |
|------|---------|---------|---------|
| 积分到账 | `credit_arrival` | 积分增加 | 积分数、来源、余额 |
| 订阅到期提醒 | `subscription_expiry` | T-7/T-3/T-1 | 等级、到期日期 |
| 推荐奖励发放 | `referral_reward` | 被邀请人付费 | 奖励积分、被邀请人 |
| 等级变更 | `level_change` | 等级升降 | 新等级、变更时间 |

**关键代码路径**:
- `notification-service/internal/provider/wx_subscribe.go` — 微信订阅消息 provider
- `notification-service/internal/handler/wx_template.go` — 模板管理 handler
- `account-service/internal/worker/wx_subscribe_sender.go` — Asynq 发送任务
- `weapp/` — 小程序端 `wx.requestSubscribeMessage` 调用

**API 变更**:
- `POST /api/v1/notifications/wx/subscribe` — 记录用户订阅授权
- `GET /api/v1/notifications/wx/templates` — 获取消息模板列表
- `POST /api/v1/notifications/wx/send` — 内部 API，触发发送

**测试策略**:
- 单元测试：模板组装、重试逻辑（5/15/60min 间隔）
- 集成测试：微信沙箱环境发送订阅消息
- 埋点验证：`wx_subscribe_msg_sent` / `wx_subscribe_msg_failed` 事件上报

---

#### 3.2.21 MB-14: 小程序分享能力

**需求关联**: PRD 3.2 MB-14  
**依赖**: FN-12（事件埋点 SDK）

**技术方案**:

在微信小程序中实现 `onShareAppMessage` 和 `onShareTimeline`，携带 `inviter_id` 参数追踪推荐链路。

```
┌────────────────────────────────────────────────┐
│  小程序分享流程                                  │
│                                                  │
│  onShareAppMessage (分享到聊天):                  │
│    title: "xxx 邀请你加入 Neuro"                  │
│    path: "/pages/register?inviter_id=USER_ID"    │
│    imageUrl: 分享卡片图片                         │
│                                                  │
│  onShareTimeline (分享到朋友圈):                  │
│    title: "Neuro - AI 智能助手"                   │
│    query: "inviter_id=USER_ID"                    │
│    imageUrl: 预览图片                             │
│                                                  │
│  好友点击 → 解析 inviter_id                       │
│    → 注册时自动关联推荐关系                        │
│    → 埋点: miniprogram_share_chat/_timeline      │
└────────────────────────────────────────────────┘
```

**关键代码路径**:
- `weapp/pages/` — 各页面 `onShareAppMessage` / `onShareTimeline` 实现
- `account-service/internal/handler/register_handler.go` — 注册时解析 `inviter_id`
- `account-service/internal/service/referral_client.go` — 推荐关系创建

**测试策略**:
- 功能测试：分享→点击→注册→推荐关系建立完整链路
- 埋点验证：`miniprogram_share_chat`、`miniprogram_share_timeline` 事件
- 边界测试：无效 inviter_id、过期分享链接、自推自

---

#### 3.2.22 MB-16~19: 广告变现基础

**需求关联**: PRD 3.2 MB-16~19  
**依赖**: FN-12（事件埋点 SDK）

**技术方案**:

实现广告变现的四项基础能力：远程配置、SDK 主备切换、频次控制、视频时长限制。

```
┌──────────────────────────────────────────────────────────┐
│  广告变现架构                                              │
│                                                            │
│  config-service                                            │
│  ┌──────────────────────────────────────┐                 │
│  │ ad_config:                            │                 │
│  │   ad_splash_enabled: true             │                 │
│  │   ad_splash_provider: "csj"           │                 │
│  │   ad_banner_enabled: true             │                 │
│  │   ad_banner_provider: "admob"         │                 │
│  │   ad_splash_max_duration_sec: 5       │                 │
│  │   ad_frequency:                       │                 │
│  │     L0: "every_launch"                │                 │
│  │     L1: "daily_once"                  │                 │
│  │     L2+: "disabled"                   │                 │
│  └──────────────────┬───────────────────┘                 │
│                     │ 客户端启动时拉取                      │
│                     ▼                                      │
│  ┌──────────────────────────────────────┐                 │
│  │  客户端 AdManager                      │                 │
│  │                                        │                 │
│  │  1. 读取配置 → 选择主 SDK              │                 │
│  │     iOS: 穿山甲(主) + AdMob(备)        │                 │
│  │     Android: 穿山甲(主) + 优量汇(备)    │                 │
│  │  2. 主 SDK 加载失败 → 自动切备 SDK     │                 │
│  │  3. 展示前请求频控 API                  │                 │
│  │  4. 视频广告 ≤5s 硬限制                 │                 │
│  └──────────────────────────────────────┘                 │
│                                                            │
│  频控服务 (api-gateway):                                    │
│    GET /api/v1/ad/frequency-check?level=L0                 │
│    Redis: ad_shown:{userID}:{date} → count                 │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `config-service` 新增广告配置项（ad_config 分组）
- `api-gateway/internal/handler/ad.go` — 频控检查 API
- `api-gateway/internal/middleware/ad_frequency.go` — Redis 频控中间件
- `ios/AccountCenter/Core/AdManager.swift` — iOS 广告管理器（主备 SDK 切换）
- `android/.../ads/AdManager.kt` — Android 广告管理器
- `android/.../ads/CsjAdLoader.kt` — 穿山甲 SDK 封装
- `android/.../ads/GdtAdLoader.kt` — 优量汇 SDK 封装

**API 变更**:
- `GET /api/v1/ad/config` — 获取广告配置（按用户等级）
- `GET /api/v1/ad/frequency-check` — 频控检查（是否允许展示）

**测试策略**:
- 单元测试：频控逻辑（L0 每次冷启动/L1 每日 1 次/L2+ 禁用）、主备 SDK 切换
- 集成测试：config-service 广告配置变更→客户端即时响应
- 时长限制测试：视频超过 5s 自动跳过

---

#### 3.2.23 AR-01: 服务间通信异步化

**需求关联**: PRD 3.2 AR-01  
**依赖**: AR-02（Saga 编排器）

**技术方案**:

将关键跨服务流程从同步 HTTP 改为 Redis Streams/Kafka 异步消息驱动。

```
同步调用 (当前):
account-service ──HTTP──→ credit-service ──HTTP──→ notification-service
   (积分消费)              (积分扣减)             (通知推送)

异步消息 (目标):
account-service ──Redis Stream──→ credit-service ──Redis Stream──→ notification-service
   (发布事件)                    (消费+发布)                (消费)

消息格式:
{
  "message_id": "uuid",
  "trace_id": "xxx",
  "event_type": "credit.deducted",
  "payload": { ... },
  "timestamp": "2026-05-19T10:00:00Z",
  "version": "v1"
}
```

**目标流程清单**:

| 流程 | 发布者 | 消费者 | Stream/Topic |
|------|--------|--------|-------------|
| 积分消费 | account | credit | `stream:credit_deduct` |
| 订阅激活 | payment | account | `stream:subscription_activate` |
| 权益发放 | account | notification | `stream:entitlement_grant` |
| 推荐奖励 | account | credit | `stream:referral_reward` |
| 等级变更 | account | notification | `stream:level_change` |

**关键设计**:
- 消费者幂等：基于 `message_id` + Redis SET 去重（TTL 24h）
- 失败重试：最多 3 次，间隔递增（1s/5s/30s）
- 死信队列：3 次失败后写入 `stream:dlq`，触发告警
- Dev/UAT: Redis Streams，Prod: Kafka（通过适配器抽象）

**关键代码路径**:
- `pkg/messaging/stream.go` — Stream 抽象接口（Redis/Kafka 适配器）
- `pkg/messaging/redis_stream.go` — Redis Streams 实现
- `pkg/messaging/kafka_stream.go` — Kafka 实现
- `pkg/messaging/consumer.go` — 消费者框架（幂等、重试、死信）
- `pkg/messaging/producer.go` — 生产者封装
- 各服务 `internal/messaging/` — 事件定义和 handler

**测试策略**:
- 单元测试：消息序列化/反序列化、幂等去重、重试逻辑
- 集成测试：完整异步流程（发布→消费→确认→失败→重试→死信）
- 性能测试：异步 vs 同步延迟对比（异步增加 ≤500ms）

---

#### 3.2.24 AR-02: 分布式事务 Saga

**需求关联**: PRD 3.2 AR-02  
**依赖**: AR-05（分布式追踪）

**技术方案**:

基于 Redis Streams 实现 Saga 编排器，保证跨服务事务的最终一致性。

```
┌──────────────────────────────────────────────────────────┐
│  Saga 编排器 (Redis Streams 事件溯源)                      │
│                                                            │
│  示例: 订阅购买 Saga                                       │
│                                                            │
│  Step 1: 积分扣减 (credit-service)                         │
│    └→ 补偿: 积分退还                                       │
│                                                            │
│  Step 2: 订阅激活 (account-service)                        │
│    └→ 补偿: 订阅取消                                       │
│                                                            │
│  Step 3: 权益发放 (account-service)                        │
│    └→ 补偿: 权益撤回                                       │
│                                                            │
│  Step 4: 通知推送 (notification-service)                    │
│    └→ 补偿: 无（容忍失败）                                  │
│                                                            │
│  状态持久化:                                                │
│    saga_instances table:                                   │
│    - saga_id, type, status, current_step                   │
│    - steps JSON: [{name, status, started_at, completed_at}]│
│    - trace_id (关联 OpenTelemetry)                         │
└──────────────────────────────────────────────────────────┘
```

**Saga 状态机**:
```
Created → Running → Completed
                ↘ Compensating → Failed
```

**关键代码路径**:
- `pkg/saga/orchestrator.go` — Saga 编排器核心（步骤执行、补偿触发）
- `pkg/saga/definition.go` — Saga 定义 DSL（步骤 + 补偿函数注册）
- `pkg/saga/state_store.go` — Redis + PG 双重状态持久化
- `pkg/saga/stream_transport.go` — Redis Streams 事件传输
- `account-service/internal/saga/subscription_purchase.go` — 订阅购买 Saga 定义
- `payment-service/internal/saga/payment_flow.go` — 支付流程 Saga 定义

**数据库变更**:
- 新增 `saga_instances` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| saga_id | UUID | PK | Saga 实例 ID |
| saga_type | VARCHAR(50) | NOT NULL | Saga 类型（subscription_purchase 等） |
| status | VARCHAR(20) | NOT NULL | running/compensating/completed/failed |
| current_step | INTEGER | | 当前步骤编号 |
| steps | JSONB | NOT NULL | 各步骤状态详情 |
| trace_id | VARCHAR(32) | | OpenTelemetry trace_id |
| payload | JSONB | | 原始请求参数 |
| created_at | TIMESTAMP | NOT NULL | |
| updated_at | TIMESTAMP | NOT NULL | |

- 索引: `INDEX(saga_type, status)`, `INDEX(trace_id)`, `INDEX(created_at)`

**API 变更**:
- `GET /api/v1/admin/saga/{saga_id}` — 查询 Saga 实例执行进度
- `GET /api/v1/admin/saga?type=&status=` — Saga 实例列表

**测试策略**:
- 单元测试：步骤执行顺序、补偿触发逻辑、状态持久化
- 集成测试：正常完成流程、Step 2 失败→Step 1 补偿→验证数据一致
- 压力测试：100 TPS 并发 Saga 实例，验证无数据不一致，最终一致 ≤10s

---

#### 3.2.25 AR-05: 分布式追踪 OpenTelemetry

**需求关联**: PRD 3.2 AR-05

**技术方案**:

所有 Go 微服务集成 OpenTelemetry Go SDK，通过 W3C Trace Context 传播 `trace_id`，接入 Jaeger/Tempo 作为后端。

```
┌──────────┐  HTTP + traceparent  ┌──────────────┐  HTTP + traceparent  ┌──────────────┐
│  客户端   │ ──────────────────→  │  api-gateway  │ ──────────────────→  │  account     │
│          │                      │  (30300)      │                      │  (30301)     │
└──────────┘                      └──────┬───────┘                      └──────┬───────┘
                                         │ inject traceparent                  │ extract
                                         ▼                                     ▼
                                  ┌──────────────────────────────────────────────┐
                                  │  OpenTelemetry Collector (OTLP)               │
                                  │                                              │
                                  │  → Jaeger/Tempo (traces)                     │
                                  │  → VictoriaMetrics (metrics)                  │
                                  │  → Loki (logs, 通过 trace_id 关联)            │
                                  └──────────────────────────────────────────────┘
```

**集成方案**:
- 使用 `go.opentelemetry.io/otel` SDK
- Gin 中间件: `otelgin.Middleware("service-name")`
- HTTP 客户端: `otelhttp.NewTransport(http.DefaultTransport)`
- Redis: `go.opentelemetry.io/otel/trace` 手动 Span
- 自定义业务标签: `trace.SetAttributes(attribute.String("user_id", uid))`

**配置项**（通过 config-service）:
- `otel.enabled`: true/false
- `otel.endpoint`: `otel-collector:4318` (OTLP HTTP)
- `otel.sampling_rate`: 0.1 (10%)
- `otel.error_sampling_rate`: 1.0 (100%)
- `otel.export_timeout_sec`: 5

**关键代码路径**:
- `pkg/otel/setup.go` — OTel SDK 初始化（Provider、Exporter、Sampler）
- `pkg/otel/gin.go` — Gin 中间件封装
- `pkg/otel/http.go` — HTTP 客户端 Transport 封装
- 各服务 `cmd/main.go` — 调用 `otel.Setup()`
- `monitoring/docker-compose.yml` — 新增 Jaeger/Tempo + OTel Collector 服务
- `monitoring/grafana/datasources/` — 新增 Jaeger/Tempo 数据源配置

**测试策略**:
- 单元测试：traceparent 注入/提取、采样率逻辑、业务标签设置
- 集成测试：跨服务调用→验证 Jaeger/Tempo 中存在完整 trace
- 性能测试：OTel 开启后 P99 延迟增加 ≤5%

---

#### 3.2.26 AR-06: 自定义 Grafana 仪表盘

**需求关联**: PRD 3.2 AR-06  
**依赖**: AR-05（OpenTelemetry）

**技术方案**:

以 JSON 模板管理 Grafana Dashboard，纳入 Git 版本控制，部署时自动导入。

```
monitoring/grafana/dashboards/
├── service_health.json         # 服务健康总览
├── api_latency.json            # API 延迟分布
├── error_rate.json             # 错误率监控
├── business_metrics.json       # 业务指标（注册/付费/收入）
├── operations_dashboard.json   # 运营数据大屏 (FN-06)
└── alert_status.json           # 告警规则状态面板
```

**Dashboard 内容**:

| Dashboard | 面板内容 |
|-----------|---------|
| 服务健康总览 | 各服务 uptime、CPU/MEM、Goroutine 数 |
| API 延迟分布 | P50/P95/P99 按 service+path 分组、热力图 |
| 错误率 | HTTP 5xx 率、4xx 率、按 service+path 趋势 |
| 业务指标 | 注册数、付费转化率、活跃用户、MRR/ARR |
| 告警状态 | AlertManager 规则状态、最近告警列表 |

**自动导入方案**:
- Grafana Provisioning: `monitoring/grafana/provisioning/dashboards/dashboard.yml` 指向 JSON 目录
- Docker Compose 挂载: `./monitoring/grafana/dashboards:/var/lib/grafana/dashboards`

**关键代码路径**:
- `monitoring/grafana/dashboards/*.json` — Dashboard JSON 模板
- `monitoring/grafana/provisioning/dashboards/dashboard.yml` — 自动导入配置
- `monitoring/docker-compose.yml` — Grafana + VictoriaMetrics + Loki + Jaeger 编排

**测试策略**:
- `helm template` 渲染验证 Dashboard JSON 语法
- Dev 环境实际部署验证所有面板数据正确展示
- Dashboard 面板加载时间 ≤5s

---

#### 3.2.27 AR-07: 告警规则配置

**需求关联**: PRD 3.2 AR-07  
**依赖**: AR-06（Grafana 仪表盘）

**技术方案**:

配置 Prometheus AlertManager 规则，支持钉钉/企微 Webhook 通知，告警静默防风暴。

```
monitoring/alertmanager/
├── alertmanager.yml            # AlertManager 全局配置
├── rules/
│   ├── service_down.yml        # 服务宕机告警
│   ├── latency_high.yml        # 延迟超阈值告警
│   ├── error_rate_high.yml     # 错误率超标告警
│   └── connection_pool.yml     # 连接池耗尽告警
└── templates/
    └── webhook.tmpl            # 钉钉/企微消息模板
```

**告警规则**:

| 规则 | 条件 | 持续时间 | 级别 |
|------|------|---------|------|
| 服务宕机 | `up == 0` | 1m | P0 |
| P99 延迟 | `histogram_quantile(0.99, ...) > 2s` | 5m | P1 |
| 错误率 | `rate(http_5xx[5m]) / rate(http_total[5m]) > 0.01` | 3m | P1 |
| 连接池耗尽 | `db_pool_idle < 5` | 2m | P1 |

**告警通知流程**:
```
Prometheus → AlertManager → Webhook → 钉钉/企微机器人
                              │
                              ├─ 消息内容: 服务名/指标值/阈值/持续时间/建议处理
                              ├─ 静默: 同一告警 30min 内不重复
                              └─ 恢复通知: 告警解除自动发送
```

**关键代码路径**:
- `monitoring/alertmanager/alertmanager.yml` — AlertManager 配置
- `monitoring/alertmanager/rules/*.yml` — 告警规则
- `monitoring/alertmanager/templates/webhook.tmpl` — 通知模板
- `monitoring/docker-compose.yml` — AlertManager 服务编排

**测试策略**:
- 规则语法验证：`promtool check rules *.yml`
- 模拟触发：手动停止某服务，验证 1m 后收到钉钉告警
- 恢复通知：重启服务后验证恢复通知发送
- 静默测试：短时间内多次触发同一告警，验证只收到一次

---

#### 3.2.28 AR-14: KMS/Vault 密钥管理

**需求关联**: PRD 3.2 AR-14

**技术方案**:

对接 HashiCorp Vault 或阿里云 KMS，所有服务加密密钥从环境变量迁移至 KMS 托管，支持 90 天自动轮换。

```
┌───────────────────────────────────────────────────┐
│  密钥管理架构                                      │
│                                                     │
│  ┌──────────────┐  ┌──────────────┐               │
│  │ HashiCorp    │  │ 阿里云 KMS   │               │
│  │ Vault        │  │ (Prod 备选)  │               │
│  │ (Dev/UAT)    │  │              │               │
│  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                        │
│         └────────┬────────┘                        │
│                  │                                  │
│         ┌────────▼────────┐                        │
│         │ pkg/kms/client  │                        │
│         │ 统一 KMS 接口    │                        │
│         │                  │                        │
│         │ GetKey(id)       │                        │
│         │ GetKeyVersion(id)│                       │
│         │ RotateKey(id)    │                        │
│         │ RevokeKey(id,v)  │                        │
│         └────────┬────────┘                        │
│                  │                                  │
│     ┌────────────┼────────────┐                    │
│     ▼            ▼            ▼                    │
│  auth-service  payment      notification          │
│  (密码哈希key) (支付签名key) (APNs .p8)           │
└───────────────────────────────────────────────────┘
```

**密钥清单**:

| 密钥 ID | 用途 | 轮换周期 |
|---------|------|---------|
| `auth/jwt-signing` | JWT 签名密钥 | 90 天 |
| `auth/argon2id-pepper` | 密码哈希 pepper | 90 天 |
| `payment/wechat-key` | 微信支付 API 密钥 | 365 天 |
| `payment/alipay-key` | 支付宝私钥 | 365 天 |
| `push/apns-key` | APNs .p8 密钥 | 365 天 |
| `encrypt/refresh-token` | Refresh token 加密密钥 | 90 天 |
| `hmac/api-signing` | API 请求签名密钥 | 90 天 |

**轮换策略**:
- 新密钥生成后，旧密钥保留用于解密（解密始终尝试最新版本，失败则回退）
- 加密始终使用最新版本密钥
- 轮换操作写入审计日志

**关键代码路径**:
- `pkg/kms/client.go` — KMS 统一客户端接口
- `pkg/kms/vault.go` — HashiCorp Vault 实现
- `pkg/kms/aliyun_kms.go` — 阿里云 KMS 实现
- `pkg/kms/rotator.go` — 密钥自动轮换调度器
- `monitoring/docker-compose.yml` — 新增 Vault 服务
- `helm/account-center/templates/secrets.yaml` — K8s Secrets 引用 KMS

**数据库变更**:
- 新增 `key_rotation_audit` 表: `id, key_id, old_version, new_version, rotated_by, rotated_at`

**测试策略**:
- 单元测试：密钥获取、版本回退、轮换逻辑
- 集成测试：Vault/KMS 连接、轮换→新旧密钥并存→加密解密
- 审计测试：轮换操作写入审计日志

---

#### 3.2.29 AR-15: API 安全加固

**需求关联**: PRD 3.2 AR-15  
**依赖**: AR-14（KMS 密钥管理）

**技术方案**:

在 api-gateway 实现用户级限流（Redis 计数器）、HMAC-SHA256 请求签名验证和 CI 安全扫描集成。

```
┌──────────────────────────────────────────────────────┐
│  API 安全加固架构                                      │
│                                                        │
│  请求流程:                                             │
│  客户端 → [限流检查] → [签名验证] → [安全扫描] → 后端   │
│                                                        │
│  1. 用户级限流 (Redis 滑动窗口):                        │
│     Key: ratelimit:{userID}:{window}                   │
│     L0: 60/min, L1: 120/min, L2+: 300/min             │
│     超限 → 429 + Retry-After Header                    │
│                                                        │
│  2. HMAC-SHA256 签名 (关键写操作):                      │
│     签名内容: HTTP Method + Path + Body Hash + Timestamp│
│     Header: X-Signature, X-Timestamp                  │
│     密钥: 从 KMS 获取 per-user 签名密钥                │
│     验证失败 → 401 + 审计日志                           │
│                                                        │
│  3. CI 安全扫描:                                       │
│     GitHub Actions → SQLMap + OWASP ZAP 自动扫描       │
│     高危漏洞 → 阻断部署                                │
└──────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `api-gateway/internal/middleware/ratelimit.go` — 用户级限流中间件（Redis 滑动窗口）
- `api-gateway/internal/middleware/request_signing.go` — HMAC-SHA256 签名验证中间件
- `api-gateway/internal/middleware/security_headers.go` — 安全响应头（CSP, X-Frame-Options）
- `pkg/kms/client.go` — 签名密钥获取
- `.github/workflows/security-scan.yml` — CI 安全扫描 workflow
- `scripts/security/zap_scan.sh` — OWASP ZAP 自动化扫描
- `scripts/security/sqlmap_scan.sh` — SQLMap 注入扫描

**测试策略**:
- 单元测试：限流计数逻辑（各等级阈值）、签名生成/验证、时间戳防重放
- 集成测试：超限→429、签名错误→401、正常请求→200
- 安全扫描：CI 中自动执行 ZAP + SQLMap，验证无高危漏洞

---

#### 3.2.30 AR-19: 性能/压力测试

**需求关联**: PRD 3.2 AR-19

**技术方案**:

使用 k6 编写性能测试脚本，对核心 API 建立性能基线，纳入 Git 管理。

```
tests/performance/
├── k6/
│   ├── scenarios/
│   │   ├── login.js            # 登录性能测试
│   │   ├── register.js         # 注册性能测试
│   │   ├── subscription.js     # 订阅购买性能测试
│   │   ├── credits.js          # 积分查询性能测试
│   │   └── referral.js         # 推荐列表性能测试
│   ├── common/
│   │   ├── config.js           # 测试配置（基准URL/阈值）
│   │   └── helpers.js          # 辅助函数
│   └── run.sh                  # 一键执行入口
├── wrk/
│   └── gateway_bench.lua       # 网关基准测试
└── results/                    # HTML 报告输出目录
```

**性能基线目标**（500 并发用户）:

| API | P95 | P99 | 错误率 | QPS |
|-----|-----|-----|--------|-----|
| 登录 | <500ms | <1s | <0.1% | >200 |
| 注册 | <500ms | <1s | <0.1% | >150 |
| 订阅购买 | <1s | <2s | <0.1% | >100 |
| 积分查询 | <200ms | <500ms | <0.1% | >500 |
| 推荐列表 | <300ms | <500ms | <0.1% | >300 |

**关键代码路径**:
- `tests/performance/k6/scenarios/*.js` — k6 测试脚本
- `tests/performance/k6/run.sh` — 一键执行脚本
- `tests/performance/wrk/gateway_bench.lua` — wrk 基准测试
- `.github/workflows/performance.yml` — CI 性能测试 workflow（可选）

**测试策略**:
- 基线建立：首次运行记录各 API 的 P95/P99/QPS 基线值
- 回归测试：每次版本发布前运行，对比基线，P95 回退 >20% 则标记
- 报告输出：k6 `--out json=results/report.json` + HTML 报告生成

---

#### 3.2.31 AR-22: CI/CD 流水线完善

**需求关联**: PRD 3.2 AR-22  
**依赖**: AR-28（Lint 严格化）

**技术方案**:

在 GitHub Actions 中实现完整的 CI/CD 流水线，支持并行构建，总执行时间 ≤15 分钟。

```
.github/workflows/ci.yml

┌─────────────────────────────────────────────────────────┐
│  CI/CD Pipeline (GitHub Actions)                         │
│                                                           │
│  Trigger: push to main/develop, PR to main                │
│                                                           │
│  ┌─────────────┐  (并行)  ┌──────────────┐              │
│  │ golangci-   │          │ go test      │              │
│  │ lint        │          │ -race -cover │              │
│  │ (all svcs)  │          │ (all svcs)   │              │
│  └──────┬──────┘          └──────┬───────┘              │
│         │ (并行通过)              │                      │
│         └──────────┬─────────────┘                      │
│                    ▼                                      │
│         ┌──────────────────────┐                         │
│         │ Docker Build         │                         │
│         │ (多阶段构建, <50MB)   │                         │
│         │ 并行: 9 服务同时构建  │                         │
│         └──────────┬───────────┘                         │
│                    ▼                                      │
│         ┌──────────────────────┐                         │
│         │ Push to Registry     │                         │
│         │ (Docker Hub / ACR)   │                         │
│         └──────────┬───────────┘                         │
│                    ▼                                      │
│         ┌──────────────────────┐                         │
│         │ Deploy to UAT (K8s)  │                         │
│         │ helm upgrade --install│                        │
│         └──────────────────────┘                         │
│                                                           │
│  任意步骤失败 → 通知钉钉/企微 + Block PR                   │
│  总执行时间目标: ≤15 分钟                                  │
└─────────────────────────────────────────────────────────┘
```

**多阶段 Dockerfile 模板**:

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/service /bin/service
EXPOSE 30301
ENTRYPOINT ["/bin/service"]
```

**关键代码路径**:
- `.github/workflows/ci.yml` — 主 CI/CD workflow
- `.github/workflows/security-scan.yml` — 安全扫描 workflow
- `.github/workflows/performance.yml` — 性能测试 workflow（手动触发）
- `Makefile` — 统一构建入口
- 各服务 `Dockerfile` — 多阶段构建

**测试策略**:
- Pipeline 冒烟测试：提交空 PR，验证完整流水线执行
- 并行验证：确认 9 个服务 Docker Build 并行执行
- 时间验证：完整流水线执行时间 ≤15 分钟

---

#### 3.2.32 AR-28: Lint 严格化

**需求关联**: PRD 3.2 AR-28

**技术方案**:

配置 `.golangci.yml` 启用严格 lint 规则，作为 CI 第一道门禁。

```yaml
# .golangci.yml
run:
  timeout: 5m
  go: "1.26"

linters:
  enable:
    - errcheck        # 未检查的 error
    - govet           # go vet 检查
    - staticcheck     # 高级静态分析
    - gosec           # 安全漏洞检查
    - revive          # 代码风格检查（替代 golint）
    - unused          # 未使用代码
    - gosimple        # 代码简化建议
    - ineffassign     # 无效赋值
    - typecheck       # 类型检查
    - misspell        # 拼写检查
    - gofmt           # 格式检查
    - goimports       # import 排序

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
  gosec:
    excludes:
      - G104  # 审计日志错误暂不强制处理
  revive:
    rules:
      - name: exported
        severity: warning
      - name: unused-parameter
        severity: warning

issues:
  max-issues-per-linter: 50
  max-same-issues: 10
  
severity:
  default-severity: error
```

**关键代码路径**:
- `.golangci.yml` — 项目根目录 lint 配置文件
- `Makefile` — 新增 `lint` target: `golangci-lint run ./...`
- `.github/workflows/ci.yml` — lint 作为第一个 job

**修复计划**:
- Phase 1: 配置 `.golangci.yml` 并运行，记录所有 error
- Phase 2: 逐服务修复 error（优先级：api-gateway > account > auth > credit > notification > compliance > data-product > config）
- Phase 3: 修复 warning（渐进式，不阻断 CI）

**测试策略**:
- `golangci-lint run ./...` 在所有 9 个服务上零 error
- CI 中 lint 步骤阻断后续 job
- 新增规则通过 PR 审核，附说明和示例

---

### 3.2 P1 技术设计小结

| 序号 | 需求 ID | 技术方案核心 | 涉及服务/模块 | 新增文件估计 |
|------|---------|------------|-------------|-------------|
| 1 | NF-03 | 共享熔断器包 `pkg/circuitbreaker` + Prometheus 集成 | 全服务 | ~6 文件 |
| 2 | NF-04 | `pkg/health` 共享健康检查框架 + 503 JSON 详情 | 全服务 | ~8 文件 |
| 3 | UX-01 | 微信/Apple/Google OAuth 一键登录 + `social_accounts` 表 | auth-service + 移动端 | ~10 文件 |
| 4 | UX-02 | Face ID/Touch ID/指纹 + 设备 token + `biometric_device_tokens` 表 | auth-service + 移动端 | ~8 文件 |
| 5 | UX-05 | config-service 配置驱动仪表盘卡片布局 | account-service + 全端 | ~6 文件 |
| 6 | UX-09 | 支付结果页 + 电子发票 + 异常订单自动修复 | payment-service + 全端 | ~10 文件 |
| 7 | UX-10 | 费用预览计算器 + 降级挽留弹窗 + 立即生效 | account-service + 全端 | ~8 文件 |
| 8 | UX-11 | Asynq T-7/T-3/T-1 续费提醒 + 多通道 + 深度链接 | account-service + notification | ~6 文件 |
| 9 | UX-12 | 推荐漏斗聚合 + 收益趋势图 + Redis 缓存预计算 | data-product + 全端 | ~8 文件 |
| 10 | FN-04 | 退款策略 + 自动/人工审核 + 原路退款 + `refunds` 表 | payment-service | ~8 文件 |
| 11 | FN-06 | Grafana 运营大屏 (注册趋势/漏斗/MRR/K-factor) | data-product-service | ~6 文件 + JSON |
| 12 | FN-07 | 套餐 CRUD + 优惠券 + 促销 + `coupons`/`promotions` 表 | account-service | ~10 文件 |
| 13 | FN-08 | 黑名单管理 + 风险事件 + 异常注册预警 + Redis 双写 | compliance-service | ~8 文件 |
| 14 | FN-12 | 三端 SDK (Web TS/iOS Swift/Android Kotlin) + batch 上报 | data-product + sdks | ~20 文件 |
| 15 | FN-15 | OAuthProvider 插件化 + Alipay/Apple/Google provider | auth-service | ~8 文件 |
| 16 | FN-17 | PIPL 导出 + 报表导出 + OAuth2 开放 API | data-product + api-gateway | ~10 文件 |
| 17 | MB-02 | Inter + Space Grotesk 字体子集化集成 | Android | ~3 文件 + 字体 |
| 18 | MB-09 | access_token 内存 + refresh_token AES-256-GCM + 设备绑定 | iOS + Android | ~6 文件 |
| 19 | MB-10 | iOS URLSession Server Trust + Android CertificatePinner | iOS + Android | ~4 文件 |
| 20 | MB-13 | 微信订阅消息 4 模板 + Asynq 重试 | notification + weapp | ~5 文件 |
| 21 | MB-14 | onShareAppMessage/onShareTimeline + inviter_id | weapp | ~3 文件 |
| 22 | MB-16~19 | 广告 SDK 主备 + config 远程配置 + Redis 频控 + 5s 限制 | config + api-gateway + 移动端 | ~10 文件 |
| 23 | AR-01 | Redis Streams/Kafka 异步消息 + 幂等消费者 + 死信队列 | 全服务 | ~8 文件 |
| 24 | AR-02 | Saga 编排器 + Redis Streams 事件溯源 + 补偿操作 | pkg/saga + 多服务 | ~8 文件 |
| 25 | AR-05 | OpenTelemetry Go SDK + W3C Trace Context + Jaeger/Tempo | 全服务 | ~6 文件 |
| 26 | AR-06 | Grafana Dashboard JSON 模板 + 自动导入 | monitoring | ~6 JSON |
| 27 | AR-07 | AlertManager YAML 规则 + 钉钉/企微 Webhook | monitoring | ~6 YAML |
| 28 | AR-14 | HashiCorp Vault/阿里云 KMS + 90 天轮换 + 审计 | pkg/kms + 全服务 | ~6 文件 |
| 29 | AR-15 | 用户级限流 Redis + HMAC-SHA256 签名 + SQLMap/ZAP | api-gateway | ~6 文件 |
| 30 | AR-19 | k6 压测脚本 + 性能基线 + HTML 报告 | tests/performance | ~8 文件 |
| 31 | AR-22 | GitHub Actions 完整流水线 + 并行构建 + ≤15min | CI/CD | ~4 YAML |
| 32 | AR-28 | `.golangci.yml` 严格规则 + errcheck/gosec/revive 等 | 全服务 | ~1 文件 + 修复 |

**汇总**: 32 项 P1 需求预计新增约 200+ 文件，涉及全部 9 个微服务、4 端前端、基础设施层和 CI/CD 流水线。

---

### 3.3 Phase 8 — P2 技术设计（32 项）

> 以下 32 项为竞争力提升项的技术设计方案，Phase 7（P1）全部完成后启动。

---

#### 3.3.01 NF-05: 配置热更新（定时轮询）

**需求关联**: PRD 3.3 NF-05

**技术方案**:

扩展现有 `pkg/config/client.go`（当前仅支持启动时单次加载），增加定时轮询和内存覆盖机制。

```
┌───────────────────────┐
│  各微服务 svcconfig    │
│  Load() 一次性加载     │
│         │              │
│         ▼              │
│  Watcher.Start()       │
│  ┌──────────────────┐  │
│  │ goroutine ticker │  │
│  │ 每 30s 轮询      │  │
│  │ config-service   │  │
│  │ /internal/v1/    │  │
│  │ config/items/:code│ │
│  └────────┬─────────┘  │
│           │ 成功        │
│           ▼             │
│  atomic.Value.Store()  │
│  (内存覆盖)             │
│  slog.Info("config     │
│   hot-reloaded")       │
│           │             │
│  失败 → slog.Warn()    │
│  保留当前内存配置       │
└───────────────────────┘
```

**热更新配置项分类**:
- **可热更新**: `rate_limit_rps`、`cache_max_age`、`gateway_timeout_sec`、`db_pool_max_open`、`ad_splash_enabled` 等运行时参数
- **需重启**: 数据库连接串、Redis 地址、Kafka broker 地址等连接类配置
- 配置项 `reloadable` 标记在 config-service `config_items` 表中新增布尔字段

**关键代码路径**:
- `pkg/config/watcher.go` — 新增 `ConfigWatcher` 结构体，含 goroutine ticker + atomic.Value
- `pkg/config/client.go` — 新增 `Watch(code string, callback func(newVal string))` 方法
- `api-gateway/internal/svcconfig/config.go` — 各配置项注册为可热更新
- `config-service/internal/model/config_item.go` — 新增 `Reloadable bool` 字段

**数据库变更**:
- `config_items` 表新增 `reloadable BOOLEAN DEFAULT false` 列
- Goose migration: `ALTER TABLE config_items ADD COLUMN reloadable BOOLEAN DEFAULT false`

**测试策略**:
- 单元测试：模拟 config-service 返回新值，验证 atomic.Value 更新；模拟 config-service 不可达，验证内存配置保留
- 集成测试：启动 config-service + 一个消费服务，修改配置后验证 60s 内生效
- 边界测试：并发读写配置值、轮询超时、配置值格式非法

---

#### 3.3.02 NF-06: 移动端深度链接

**需求关联**: PRD 3.3 NF-06

**技术方案**:

实现 iOS Universal Links、Android App Links 和 `neuro://` 自定义 scheme 三层深度链接。

```
┌──────────────────────────────────────────────────────────────┐
│  iOS 深度链接                                                  │
│  ┌──────────────────────┐    ┌──────────────────────────┐    │
│  │ AASA 文件托管         │    │ Info.plist 配置           │    │
│  │ https://neuro.ai/    │←──│ com.apple.developer.      │    │
│  │ .well-known/         │   │ associated-domains        │    │
│  │ apple-app-site-      │   └──────────────────────────┘    │
│  │ association          │                                    │
│  └──────────────────────┘                                    │
│                                                              │
│  Android 深度链接                                             │
│  ┌──────────────────────┐    ┌──────────────────────────┐    │
│  │ assetlinks.json 托管  │    │ AndroidManifest.xml       │    │
│  │ https://neuro.ai/    │←──│ <intent-filter>            │    │
│  │ .well-known/         │   │ autoVerify=true            │    │
│  │ assetlinks.json      │   └──────────────────────────┘    │
│  └──────────────────────┘                                    │
│                                                              │
│  兜底: neuro://path?inviter_id=xxx → 未安装时跳转应用商店       │
└──────────────────────────────────────────────────────────────┘
```

**深度链接路由表**:

| 路径 | 目标页面 | 参数 |
|------|---------|------|
| `/subscribe` | 订阅页 | `plan_id`, `inviter_id` |
| `/referral/register` | 注册页 | `inviter_id` |
| `/credits` | 积分页 | — |
| `/settings` | 设置页 | `tab` |

**关键代码路径**:
- `ios/AccountCenter/DeepLink/DeepLinkRouter.swift` — Universal Links 路由解析
- `ios/AccountCenter/Info.plist` — `com.apple.developer.associated-domains` 配置
- `android/.../deeplink/DeepLinkRouter.kt` — App Links 路由解析
- `android/.../AndroidManifest.xml` — `<intent-filter autoVerify="true">` 配置
- `web/public/.well-known/apple-app-site-association` — AASA 文件
- `web/public/.well-known/assetlinks.json` — Android 验证文件

**API 变更**:
- `GET /api/v1/deeplink/resolve?url={encoded_url}` — 服务端解析深度链接并返回目标路由信息

**测试策略**:
- iOS 测试：Xcode `xcrun applinks` 验证 Universal Links；模拟冷启动/热启动深度链接
- Android 测试：`adb shell am start -a android.intent.action.VIEW -d` 验证 App Links
- 兜底测试：未安装 App 时 `neuro://` scheme 跳转应用商店
- 推荐参数透传测试：`inviter_id` 从链接→注册→关联推荐关系完整链路

---

#### 3.3.03 NF-07: API Gateway 代码重构

**需求关联**: PRD 3.3 NF-07

**技术方案**:

将当前 `api-gateway/cmd/main.go`（461 行）拆分为模块化结构，保持全部现有功能不变。

```
api-gateway/
├── cmd/
│   └── main.go                     # 仅路由注册和启动（~80 行）
├── internal/
│   ├── middleware/
│   │   ├── jwt.go                  # jwtAuthMiddleware（~60 行）
│   │   ├── ratelimit.go            # tokenBucket + ipRateLimiter + rateLimitMiddleware（~70 行）
│   │   ├── cors.go                 # corsMiddleware（~25 行）
│   │   ├── desensitize.go          # desensitizeMiddleware + 正则（~80 行）
│   │   ├── request_id.go           # requestIDMiddleware（~15 行）
│   │   ├── cache.go                # cacheControlMiddleware（~10 行）
│   │   └── metrics.go              # 请求计数中间件（~15 行）
│   ├── proxy/
│   │   ├── reverse_proxy.go        # proxyHandler + Transport 配置（~40 行）
│   │   └── writer.go               # responseWriter / responseCaptureWriter（~30 行）
│   └── svcconfig/
│       └── config.go               # （现有，保持不变）
```

**关键代码路径**:
- `api-gateway/cmd/main.go` — 仅保留 `main()` 函数：配置加载→Gin engine 创建→中间件注册→路由注册→graceful shutdown
- `api-gateway/internal/middleware/jwt.go` — 从 main.go 提取 `jwtAuthMiddleware`
- `api-gateway/internal/middleware/ratelimit.go` — 从 main.go 提取 `tokenBucket`、`ipRateLimiter`、`rateLimitMiddleware`
- `api-gateway/internal/middleware/cors.go` — 从 main.go 提取 `corsMiddleware`
- `api-gateway/internal/middleware/desensitize.go` — 从 main.go 提取 `desensitizeMiddleware` 及正则
- `api-gateway/internal/proxy/reverse_proxy.go` — 从 main.go 提取 `proxyHandler`

**测试策略**:
- 单元测试：每个中间件独立测试（JWT 验证/过期/缺失、限流触发/恢复、CORS header、脱敏替换）
- 集成测试：重构后全链路代理行为与重构前完全一致（通过镜像流量对比验证）
- 覆盖率目标：各中间件模块 ≥70%

---

#### 3.3.04 UX-03: 短信验证码自动填充

**需求关联**: PRD 3.3 UX-03

**技术方案**:

```
┌─────────────────────────────────────────────────────────┐
│  iOS SMS Code AutoFill                                   │
│  ┌──────────────────────┐                               │
│  │ TextField 设置:       │                               │
│  │ .textContentType =   │                               │
│  │  .oneTimeCode        │                               │
│  │ .keyboardType =      │                               │
│  │  .numberPad          │                               │
│  └──────────────────────┘                               │
│  系统自动识别短信中的 \d{4,6} 格式验证码                  │
│                                                          │
│  Android SMS Retriever API                               │
│  ┌──────────────────────┐                               │
│  │ SmsRetrieverClient    │                               │
│  │ .startSmsRetriever() │                               │
│  │ → SmsRetrieverStatus │                               │
│  │ → BroadcastReceiver  │                               │
│  │ → 解析短信中验证码    │                               │
│  └──────────────────────┘                               │
│  短信格式要求：<#> 你的验证码是 123456。AppHash           │
└─────────────────────────────────────────────────────────┘
```

**短信模板适配**（notification-service 侧）:
- 短信内容末尾添加 App Hash（Android SMS Retriever 要求）
- 格式：`<#> 你的验证码是 123456。A1B2C3`（A1B2C3 为 App Hash）
- iOS 无特殊格式要求，但 `<#>` 前缀可提升识别率

**关键代码路径**:
- `ios/AccountCenter/Views/LoginView.swift` — TextField `.textContentType(.oneTimeCode)` 配置
- `android/.../ui/login/LoginScreen.kt` — SmsRetrieverClient 集成
- `android/.../receiver/SmsReceiver.kt` — BroadcastReceiver 解析验证码
- `notification-service/internal/service/sms_service.go` — 短信模板添加 App Hash 后缀

**测试策略**:
- iOS 测试：发送短信后验证 AutoFill 建议栏出现，点击自动填充
- Android 测试：SmsRetriever 集成测试，验证码解析正确率
- 降级测试：短信格式不匹配时手动输入正常可用

---

#### 3.3.05 UX-04: 渐进式注册

**需求关联**: PRD 3.3 UX-04

**技术方案**:

```
漏斗阶段:
┌──────────────────────────────────────────────────┐
│  阶段 1: 游客浏览                                 │
│  → 分配 guest_token（无 user_id）                  │
│  → 可访问: 定价页、功能介绍、公开内容               │
│                                                   │
│  阶段 2: 邮箱注册（最低门槛）                       │
│  → POST /api/v1/auth/register/email               │
│  → 创建 L0 用户（email only, phone=null）          │
│  → 可使用: 基础功能                                │
│                                                   │
│  阶段 3: 手机验证（触发时引导）                     │
│  → 触发点: 订阅购买、积分操作、推荐分享              │
│  → POST /api/v1/account/bind-phone                │
│  → 升级为完整 L0 用户                              │
└──────────────────────────────────────────────────┘
```

**guest_token 设计**:
- 网关层生成短期 JWT（24h TTL），`sub: "guest:{uuid}"`，`role: "guest"`
- guest_token 不关联 user 记录，仅用于偏好追踪和后续注册关联

**关键代码路径**:
- `auth-service/internal/handler/guest_handler.go` — 游客 token 生成、邮箱注册
- `auth-service/internal/service/guest_service.go` — 渐进注册业务逻辑
- `auth-service/internal/handler/bind_phone_handler.go` — 手机号绑定
- `api-gateway/cmd/main.go` — guest 路径加入 `publicPaths`
- `web/src/views/GuestLanding.vue` — 游客引导页面
- `ios/AccountCenter/Views/GuestBrowseView.swift` — iOS 游客浏览
- `android/.../ui/guest/GuestScreen.kt` — Android 游客浏览

**数据库变更**:
- `users` 表 `phone_number` 列改为 `NULLABLE`（当前可能已是，需验证）
- 新增 `guest_sessions` 表: `guest_id UUID PK`, `fingerprint VARCHAR`, `created_at TIMESTAMP`

**API 变更**:
- `POST /api/v1/auth/guest/token` — 获取游客 token
- `POST /api/v1/auth/register/email` — 邮箱注册（创建 L0 用户）
- `POST /api/v1/account/bind-phone` — 已注册用户绑定手机号

**测试策略**:
- 单元测试：guest_token 生成/验证、邮箱注册验证、手机绑定流程
- 集成测试：游客→邮箱注册→手机绑定完整漏斗
- 埋点验证：`progressive_signup_email`、`progressive_signup_phone_bound` 事件正确上报

---

#### 3.3.06 UX-06: 空状态引导

**需求关联**: PRD 3.3 UX-06

**技术方案**:

为积分页、订阅页、推荐页设计空状态引导卡片组件，数据为空时展示引导内容。

```
┌───────────────────────────────────┐
│  EmptyStateCard 组件               │
│  ┌─────────────────────────────┐  │
│  │  插图/图标（暗色科技风 SVG） │  │
│  │  主标题                     │  │
│  │  副标题（说明文案）          │  │
│  │  ┌─────────────────────┐   │  │
│  │  │  CTA 按钮            │   │  │
│  │  └─────────────────────┘   │  │
│  └─────────────────────────────┘  │
│                                   │
│  按页面配置:                       │
│  积分页: "邀请好友赚积分"          │
│  订阅页: 等级权益对比+"立即开通"    │
│  推荐页: "生成推荐链接"            │
└───────────────────────────────────┘
```

**空状态配置来源**:
- 通过 config-service 管理空状态文案和 CTA 链接
- 配置项: `empty_state_credits_title`、`empty_state_credits_cta_text` 等
- 支持运营动态调整文案，无需发版

**关键代码路径**:
- `web/src/components/EmptyStateCard.vue` — Web 空状态组件
- `ios/AccountCenter/Components/EmptyStateCard.swift` — iOS 空状态组件
- `android/.../ui/components/EmptyStateCard.kt` — Android 空状态组件
- `miniprogram/components/empty-state/` — 小程序空状态组件

**测试策略**:
- 组件测试：空状态组件渲染、CTA 点击事件、配置文案动态加载
- 集成测试：新用户首次进入积分/订阅/推荐页时空状态正确展示
- 视觉回归测试：暗色科技风设计一致性

---

#### 3.3.07 UX-13: 社交分享优化

**需求关联**: PRD 3.3 UX-13

**技术方案**:

```
┌──────────────────────────────────────────────────────────────┐
│  分享流程                                                     │
│                                                              │
│  用户点击"分享"                                               │
│       │                                                      │
│       ▼                                                      │
│  ┌──────────────┐                                            │
│  │ 模板选择面板  │ ← ≥3 模板（科技风/简约风/节日风）          │
│  └──────┬───────┘                                            │
│         │ 选定模板                                            │
│         ▼                                                    │
│  ┌──────────────────────────────┐                            │
│  │ 服务端生成分享图（Puppeteer/  │                            │
│  │ Canvas 渲染 HTML→PNG）        │                            │
│  │ 携带 inviter_id 参数          │                            │
│  └──────┬───────────────────────┘                            │
│         │                                                    │
│    ┌────┼────┐                                               │
│    ▼    ▼    ▼                                               │
│  微信   朋友圈  短链+OG                                       │
│  聊天   (小程序) (Web)                                        │
└──────────────────────────────────────────────────────────────┘
```

**海报模板系统**:
- 服务端存储 HTML 模板，包含用户信息占位符（`{{nickname}}`、`{{qrcode_url}}`）
- 分享时实时渲染为 PNG，CDN 缓存 24h

**Open Graph 元数据**（Web 端）:
```html
<meta property="og:title" content="好友邀请你加入 Neuro">
<meta property="og:description" content="专属推荐链接">
<meta property="og:image" content="https://cdn.neuro.ai/share/preview/{inviter_id}.png">
```

**关键代码路径**:
- `credit-service/internal/handler/share_handler.go` — 分享图生成 API
- `credit-service/internal/service/share_service.go` — 海报模板渲染逻辑
- `web/src/views/SharePreview.vue` — Web 分享预览页（含 OG meta）
- `miniprogram/pages/share/` — 小程序分享模板

**API 变更**:
- `POST /api/v1/referral/share/image` — 生成分享海报图（参数: `template_id`, `inviter_id`）
- `GET /api/v1/referral/share/og-preview?inviter_id={id}` — OG 预览元数据

**测试策略**:
- 单元测试：模板渲染正确性、OG 元数据格式
- 集成测试：生成→分享→链接→注册→推荐关联完整链路
- 埋点验证：`share_template_selected`、`share_channel_used` 事件

---

#### 3.3.08 UX-15: 消息中心

**需求关联**: PRD 3.3 UX-15  
**依赖**: FN-10（APNs/FCM 推送集成）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  notification-service 消息中心模块                         │
│                                                          │
│  ┌──────────┐ 事件触发  ┌───────────────────────┐        │
│  │ 各服务    │ ────────→ │ MessageCenterService  │        │
│  │ (异步消息)│          │ 1. 生成消息记录        │        │
│  └──────────┘          │ 2. 写入 notifications  │        │
│                        │    表                  │        │
│                        │ 3. 推送 Push 通知      │        │
│                        └───────────┬───────────┘        │
│                                    │                     │
│  ┌─────────────────────────────────▼──────────────────┐  │
│  │ 消息中心 API                                        │  │
│  │ GET /notifications          → 消息列表（分页）       │  │
│  │ GET /notifications/unread-count → 未读计数           │  │
│  │ PUT /notifications/{id}/read → 标记已读             │  │
│  │ PUT /notifications/read-all → 全部已读              │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

**消息类型枚举**:

| type | 分类 | 图标 | 跳转 |
|------|------|------|------|
| `subscription_expiry` | 系统 | ⏰ | 订阅页 |
| `credit_change` | 系统 | 💰 | 积分页 |
| `security_alert` | 系统 | 🔒 | 安全设置 |
| `level_change` | 系统 | ⭐ | 等级页 |
| `referral_reward` | 系统 | 🎁 | 推荐页 |
| `promo_activity` | 运营 | 📢 | 活动页 |

**关键代码路径**:
- `notification-service/internal/handler/notification_handler.go` — 消息中心 API
- `notification-service/internal/service/notification_service.go` — 消息业务逻辑
- `notification-service/internal/repository/notification_repository.go` — 消息数据层
- `notification-service/internal/model/notification.go` — 消息模型

**数据库变更**:
- 新增 `notifications` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 消息 ID |
| user_id | UUID | NOT NULL, FK | 用户 ID |
| type | VARCHAR(30) | NOT NULL | 消息类型 |
| category | VARCHAR(10) | NOT NULL | system/promo |
| title | VARCHAR(200) | NOT NULL | 消息标题 |
| body | TEXT | | 消息正文 |
| link | VARCHAR(500) | | 跳转链接 |
| is_read | BOOLEAN | DEFAULT false | 已读状态 |
| read_at | TIMESTAMP | | 已读时间 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| expired_at | TIMESTAMP | | 过期时间（默认 90 天） |

- 索引: `idx_notifications_user_unread` ON (user_id, is_read, created_at DESC)

**API 变更**:
- `GET /api/v1/notifications` — 消息列表（分页、类型筛选）
- `GET /api/v1/notifications/unread-count` — 未读消息计数
- `PUT /api/v1/notifications/{id}/read` — 标记已读
- `PUT /api/v1/notifications/read-all` — 全部已读

**测试策略**:
- 单元测试：消息创建、已读标记、全部已读、过期归档逻辑
- 集成测试：事件触发→消息创建→Push 推送→客户端拉取完整链路
- 性能测试：未读计数查询 <50ms（1000 条消息）

---

#### 3.3.09 UX-16: 帮助/FAQ 系统

**需求关联**: PRD 3.3 UX-16

**技术方案**:

```
┌──────────────────────────────────────────────────────┐
│  FAQ 数据来源: config-service / 独立 CMS               │
│                                                      │
│  ┌──────────────┐  CRUD   ┌────────────────────┐    │
│  │ 管理后台      │ ──────→ │ config-service     │    │
│  │ (运营操作)    │         │ faq_items 表        │    │
│  └──────────────┘         └──────────┬─────────┘    │
│                                      │               │
│                           GET /api/v1/faq            │
│                                      │               │
│            ┌────────────┬────────────┼──────────┐    │
│            ▼            ▼            ▼          ▼    │
│          Web          iOS        Android    小程序    │
│       FAQ页面       FAQ页面     FAQ页面    FAQ页面    │
│                                                      │
│  搜索: PostgreSQL tsvector 全文索引                   │
│  客服入口: 第三方 SDK (美洽/智齿) Web 聊天组件         │
└──────────────────────────────────────────────────────┘
```

**FAQ 分类**:
- `account`: 账号管理
- `payment`: 支付订阅
- `credits`: 积分推荐
- `security`: 安全设置

**关键代码路径**:
- `config-service/internal/handler/faq_handler.go` — FAQ CRUD API
- `config-service/internal/service/faq_service.go` — FAQ 业务逻辑
- `config-service/internal/model/faq.go` — FAQ 数据模型
- `web/src/views/FAQPage.vue` — Web FAQ 页面（搜索+分类+详情）
- `ios/AccountCenter/Views/FAQView.swift` — iOS FAQ 页面
- `android/.../ui/faq/FAQScreen.kt` — Android FAQ 页面

**数据库变更**:
- 新增 `faq_items` 表: `id UUID PK`, `category VARCHAR(20)`, `question TEXT`, `answer TEXT`, `tags TEXT[]`, `sort_order INT`, `is_published BOOLEAN DEFAULT true`, `created_at TIMESTAMP`, `updated_at TIMESTAMP`
- 新增 GIN 索引: `idx_faq_search` ON `question` 和 `answer` 的 `tsvector`

**API 变更**:
- `GET /api/v1/faq` — FAQ 列表（支持分类筛选和关键词搜索）
- `POST /api/v1/admin/faq` — 创建 FAQ 条目
- `PUT /api/v1/admin/faq/{id}` — 编辑 FAQ 条目
- `DELETE /api/v1/admin/faq/{id}` — 删除 FAQ 条目

**测试策略**:
- 单元测试：FAQ CRUD、全文搜索查询
- 前端测试：搜索输入→结果展示、分类切换、客服入口点击
- 埋点验证：`faq_search`、`faq_article_viewed`、`faq_feedback_helpful`

---

#### 3.3.10 FN-03: 电子发票/收据

**需求关联**: PRD 3.3 FN-03  
**依赖**: FN-01（支付网关集成）、FN-02（订单管理系统）

**技术方案**:

```
┌──────────────────┐  支付成功   ┌──────────────────┐  开票请求  ┌──────────────┐
│  payment-service  │ ─────────→ │  invoice-service  │ ────────→ │  第三方发票   │
│  callback handler │            │  (模块)           │          │  平台 API     │
│                   │            │                   │          │ (百望/航信)    │
└──────────────────┘            │  1. 创建开票请求   │          └──────┬───────┘
                                │  2. 调用第三方 API  │                 │
                                │  3. 接收发票 PDF   │◄────────────────┘
                                │  4. 邮件推送       │
                                │  5. 记录发票信息   │
                                └──────────────────┘
```

**开票模式**:
- **自动开票**: 支付成功后触发（仅企业用户，需预先维护发票抬头信息）
- **手动开票**: 用户在订单详情页申请，填写发票抬头和税号

**发票信息结构**:
```json
{
  "invoice_type": "personal|enterprise_vat",
  "title": "发票抬头",
  "tax_number": "税号",
  "email": "接收邮箱",
  "amount_cents": 2900,
  "order_id": "uuid"
}
```

**关键代码路径**:
- `payment-service/internal/handler/invoice_handler.go` — 发票申请 API
- `payment-service/internal/service/invoice_service.go` — 发票业务逻辑
- `payment-service/internal/provider/invoice_provider.go` — 第三方发票平台接口
- `payment-service/internal/service/invoice_notify.go` — 发票邮件推送

**数据库变更**:
- 新增 `invoices` 表: `id UUID PK`, `order_id UUID FK`, `user_id UUID`, `invoice_type VARCHAR(20)`, `title VARCHAR(200)`, `tax_number VARCHAR(50)`, `amount_cents INT`, `status VARCHAR(20) DEFAULT 'pending'`, `pdf_url VARCHAR(500)`, `invoice_number VARCHAR(50)`, `invoice_date TIMESTAMP`, `retry_count INT DEFAULT 0`, `created_at TIMESTAMP`, `updated_at TIMESTAMP`
- 新增 `user_invoice_info` 表: `user_id UUID PK`, `default_title VARCHAR(200)`, `default_tax_number VARCHAR(50)`, `default_email VARCHAR(255)`

**API 变更**:
- `POST /api/v1/invoices` — 申请开票
- `GET /api/v1/invoices/{id}` — 查询发票详情
- `GET /api/v1/invoices` — 发票列表
- `PUT /api/v1/users/me/invoice-info` — 维护默认发票信息
- `GET /api/v1/admin/invoices` — 管理端发票查询

**测试策略**:
- 单元测试：发票信息校验、金额匹配、重试逻辑
- 集成测试：支付成功→自动开票→邮件推送（使用第三方沙箱环境）
- 异常测试：第三方平台超时/失败重试、发票信息格式错误

---

#### 3.3.11 FN-09: 内容/通知管理

**需求关联**: PRD 3.3 FN-09  
**依赖**: FN-10（APNs/FCM 推送集成）

**技术方案**:

```
┌─────────────────────────────────────────────────────────┐
│  管理后台 — 通知模板管理                                  │
│                                                         │
│  ┌──────────────┐  CRUD   ┌──────────────────────────┐ │
│  │ Admin UI     │ ──────→ │ notification-service      │ │
│  │ 模板编辑器   │         │ /api/v1/admin/templates   │ │
│  └──────────────┘         └────────────┬─────────────┘ │
│                                        │                │
│  模板示例:                              │ 发送           │
│  "亲爱的 {{user_name}}，您的            ▼                │
│   订阅将于 {{expire_date}} 到期"   ┌──────────────┐     │
│                                    │ 变量插值引擎  │     │
│  支持渠道:                         │ text/template │     │
│  - SMS (短信)                      └──────┬───────┘     │
│  - Email (邮件)                           │             │
│  - Push (APNs/FCM)             ┌───────────┼────────┐  │
│                                ▼           ▼        ▼  │
│                             SMS Provider  Email    Push │
│                                         Provider Provider│
└─────────────────────────────────────────────────────────┘
```

**模板引擎**: Go `text/template` 标准库，支持变量插值 `{{variable_name}}`

**关键代码路径**:
- `notification-service/internal/handler/template_handler.go` — 模板 CRUD API
- `notification-service/internal/service/template_service.go` — 模板渲染和变量插值
- `notification-service/internal/model/template.go` — 模板数据模型
- `notification-service/internal/repository/template_repository.go` — 模板数据层
- `notification-service/internal/repository/send_record_repository.go` — 发送记录数据层

**数据库变更**:
- 新增 `notification_templates` 表: `id UUID PK`, `code VARCHAR(50) UNIQUE`, `channel VARCHAR(10)` (sms/email/push), `subject TEXT`（邮件主题/推送标题）, `body TEXT NOT NULL`（模板正文）, `variables TEXT[]`（可用变量列表）, `is_active BOOLEAN DEFAULT true`, `created_at TIMESTAMP`, `updated_at TIMESTAMP`
- 新增 `notification_send_records` 表: `id UUID PK`, `template_id UUID FK`, `user_id UUID`, `channel VARCHAR(10)`, `status VARCHAR(10)` (pending/sent/failed), `rendered_content TEXT`, `error_message TEXT`, `scheduled_at TIMESTAMP`, `sent_at TIMESTAMP`, `retry_count INT DEFAULT 0`, `created_at TIMESTAMP`

**API 变更**:
- `POST /api/v1/admin/notifications/templates` — 创建模板
- `PUT /api/v1/admin/notifications/templates/{id}` — 编辑模板
- `GET /api/v1/admin/notifications/templates` — 模板列表
- `POST /api/v1/admin/notifications/send` — 发送通知（支持定时/定向）
- `GET /api/v1/admin/notifications/send-records` — 发送记录查询
- `POST /api/v1/admin/notifications/preview` — 预览渲染结果

**测试策略**:
- 单元测试：变量插值渲染、模板校验、发送状态机
- 集成测试：创建模板→渲染→发送（mock SMS/Email/Push provider）→记录查询
- 异常测试：变量缺失、模板语法错误、发送失败重试

---

#### 3.3.12 FN-11: 推送策略引擎

**需求关联**: PRD 3.3 FN-11  
**依赖**: FN-10（APNs/FCM 推送集成）

**技术方案**:

```
┌─────────────────────────────────────────────────────────┐
│  推送策略引擎                                             │
│                                                         │
│  ┌───────────────┐                                       │
│  │ 策略配置       │  config-service 管理                  │
│  │ push_strategy │                                       │
│  │ 表            │                                       │
│  └───────┬───────┘                                       │
│          │                                               │
│          ▼                                               │
│  ┌───────────────────────────────┐                      │
│  │ PushStrategyEngine            │                      │
│  │                               │                      │
│  │ 发送前检查:                    │                      │
│  │ 1. 频率限制: Redis INCR        │                      │
│  │    key: push:freq:{user_id}:  │                      │
│  │    {date}                     │                      │
│  │    阈值: 3 条/日（可配置）     │                      │
│  │                               │                      │
│  │ 2. 免打扰: 查询用户设置        │                      │
│  │    dnd_start / dnd_end        │                      │
│  │    → 延迟至时段结束            │                      │
│  │                               │                      │
│  │ 3. 标签定向: 用户属性筛选      │                      │
│  │    level IN (L2,L3) AND       │                      │
│  │    registered_at > ?          │                      │
│  │                               │                      │
│  │ 4. 定时发送: Asynq Schedule   │                      │
│  └───────────────────────────────┘                      │
│          │                                               │
│          ▼                                               │
│  APNs / FCM / HMS Provider                              │
└─────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `notification-service/internal/service/push_strategy_engine.go` — 策略引擎核心
- `notification-service/internal/repository/push_preference_repository.go` — 用户推送偏好
- `notification-service/internal/model/push_strategy.go` — 策略数据模型
- `notification-service/internal/handler/push_strategy_handler.go` — Admin 策略配置 API

**数据库变更**:
- 新增 `push_strategies` 表: `id UUID PK`, `name VARCHAR(100)`, `target_tags JSONB`, `schedule_at TIMESTAMP`, `frequency_limit INT DEFAULT 3`, `is_active BOOLEAN DEFAULT true`, `template_id UUID FK`, `created_at TIMESTAMP`
- 新增 `user_push_preferences` 表: `user_id UUID PK`, `dnd_start TIME`, `dnd_end TIME`, `dnd_enabled BOOLEAN DEFAULT false`, `quiet_push_enabled BOOLEAN DEFAULT true`, `updated_at TIMESTAMP`

**Redis Key 设计**:
- `push:freq:{user_id}:{YYYYMMDD}` — 当日推送计数（TTL: 到次日 00:00）
- `push:delayed:{user_id}` — 免打扰延迟队列

**API 变更**:
- `POST /api/v1/admin/push/strategies` — 创建推送策略
- `GET /api/v1/admin/push/strategies` — 策略列表
- `PUT /api/v1/admin/push/strategies/{id}` — 编辑策略
- `PUT /api/v1/users/me/push-preferences` — 用户设置免打扰

**测试策略**:
- 单元测试：频率限制计数、免打扰时段判断、标签筛选逻辑
- 集成测试：创建策略→筛选用户→执行推送→验证到达
- 边界测试：频率限制边界、免打扰跨午夜、静默推送

---

#### 3.3.13 FN-13: 实时用户行为流

**需求关联**: PRD 3.3 FN-13  
**依赖**: FN-12（事件埋点 SDK）、AR-05（分布式追踪）

**技术方案**:

```
┌──────────────┐  埋点事件   ┌────────────────────────────┐
│ 客户端 SDK   │ ──────────→ │ data-product-service       │
│ (FN-12)      │  HTTP batch │ /api/v1/events/collect     │
└──────────────┘             └──────────┬─────────────────┘
                                        │
                                        ▼
                             ┌──────────────────────┐
                             │ Kafka Topic:          │
                             │ user-events           │
                             │ (或 Redis Stream)     │
                             └──────────┬───────────┘
                                        │ 消费
                              ┌─────────▼──────────┐
                              │ StreamProcessor    │
                              │ (实时聚合)          │
                              │                    │
                              │ 1. 在线人数         │
                              │   Redis PFCOUNT    │
                              │   HyperLogLog      │
                              │                    │
                              │ 2. 实时漏斗         │
                              │   Redis INCR       │
                              │   funnel:{step}    │
                              │                    │
                              │ 3. 实时收入         │
                              │   Redis INCRBYFLOAT│
                              │   revenue:{hour}   │
                              └─────────┬──────────┘
                                        │
                              ┌─────────▼──────────┐
                              │ 异常检测引擎        │
                              │ 滑动窗口计数        │
                              │ 超阈值→告警         │
                              └────────────────────┘
```

**关键代码路径**:
- `data-product-service/internal/handler/event_handler.go` — 事件收集 API
- `data-product-service/internal/service/stream_processor.go` — 实时流处理器
- `data-product-service/internal/service/anomaly_detector.go` — 异常检测引擎
- `data-product-service/internal/handler/realtime_handler.go` — 实时数据查询 API

**Redis 数据结构**:
- `realtime:online` — HyperLogLog（用户 UUID，1 分钟 TTL 判断在线）
- `realtime:funnel:{event_type}:{YYYYMMDDHH}` — 漏斗各步骤计数
- `realtime:revenue:{YYYYMMDDHH}` — 实时收入累计
- `realtime:anomaly:{user_id}:{event_type}` — 异常行为计数（1 分钟窗口）

**API 变更**:
- `POST /api/v1/events/collect` — 事件收集（批量）
- `GET /api/v1/realtime/online-count` — 当前在线人数
- `GET /api/v1/realtime/funnel` — 实时转化漏斗
- `GET /api/v1/realtime/revenue` — 实时收入
- `GET /api/v1/admin/realtime/alerts` — 异常行为告警列表

**测试策略**:
- 单元测试：事件解析、滑动窗口计数、异常阈值判断
- 集成测试：埋点 SDK→事件收集→Kafka→流处理→API 查询端到端 ≤5s
- 压力测试：1000 事件/秒吞吐量下聚合精度和延迟

---

#### 3.3.14 MB-01: 设计系统组件库统一

**需求关联**: PRD 3.3 MB-01

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  Design Token 源文件 (JSON/YAML)                          │
│  {                                                       │
│    "color": { "primary": "#6C5CE7", "accent": "#00D2FF" },│
│    "spacing": { "base": "4px", "unit": [4,8,12,16,24] }, │
│    "radius": { "sm": "4px", "md": "8px", "lg": "16px" }, │
│    "font": { "body": "Inter", "heading": "Space Grotesk"} │
│  }                                                       │
│                                                          │
│  ┌─────────────────────── 自动导出 ─────────────────────┐ │
│  │                                                      │ │
│  │  Web (CSS Variables)    ios (Swift)      Android (KT) │ │
│  │  :root {               extension Color  object       │ │
│  │    --color-primary:    { static let     Theme {       │ │
│  │      #6C5CE7;           primary = ...    val primary  │ │
│  │  }                     }               }             │ │
│  │                                                      │ │
│  │  小程序 (WXSS)                                       │ │
│  │  page { --color-primary: #6C5CE7; }                  │ │
│  └──────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

**Design Token 文件结构**:
```
design-tokens/
├── tokens/
│   ├── color.json
│   ├── spacing.json
│   ├── typography.json
│   ├── border-radius.json
│   └── shadow.json
├── build/
│   ├── to-css.js          # Web 导出脚本
│   ├── to-swift.js        # iOS 导出脚本
│   ├── to-kotlin.js       # Android 导出脚本
│   └── to-wxss.js         # 小程序导出脚本
└── package.json
```

**核心组件清单（≥20 个）**:
Button、Input、TextArea、Card、Dialog、Toast、TabBar、NavBar、Avatar、Badge、Skeleton（骨架屏）、EmptyState、LoadingSpinner、Switch、Checkbox、Radio、BottomSheet、SegmentedControl、Progress、Tag

**关键代码路径**:
- `design-tokens/` — Design Token 源文件和导出脚本
- `web/src/components/ui/` — Web 组件库（Vue 3 SFC）
- `ios/AccountCenter/DesignSystem/` — iOS 组件库（SwiftUI View）
- `android/.../designsystem/` — Android 组件库（Compose）
- `miniprogram/components/ui/` — 小程序组件库
- `web/.storybook/` — Storybook 配置

**测试策略**:
- 视觉回归测试：Storybook + Chromatic 截图对比
- 单元测试：各组件 props/events/slots 行为验证
- 跨端一致性：四端相同组件的视觉对比验收

---

#### 3.3.15 MB-03: 响应式布局

**需求关联**: PRD 3.3 MB-03

**技术方案**:

在 Web（Vue 3）端实现响应式布局，使用 CSS Media Query 和 Vue 3 `useBreakpoint` composable。

**断点定义**:
```css
/* mobile: < 768px */
/* tablet: 768px - 1024px */
/* desktop: > 1024px */

@media (max-width: 767px) { /* mobile */ }
@media (min-width: 768px) and (max-width: 1024px) { /* tablet */ }
@media (min-width: 1025px) { /* desktop */ }
```

**关键适配点**:

| 组件 | mobile | tablet | desktop |
|------|--------|--------|---------|
| NavBar | 汉堡菜单 | 汉堡菜单 | 完整导航 |
| 定价卡片 | 纵向堆叠 | 2 列 | 4 列 |
| 权益对比表 | 可横滑 | 完整展示 | 完整展示 |
| 仪表盘 | 单列卡片 | 2 列网格 | 4 列网格 |

**关键代码路径**:
- `web/src/composables/useBreakpoint.ts` — 响应式断点 composable
- `web/src/components/layout/ResponsiveNavBar.vue` — 自适应导航栏
- `web/src/assets/styles/responsive.scss` — 全局响应式样式
- `web/src/views/PricingPage.vue` — 定价页响应式布局

**测试策略**:
- 视觉测试：iPhone SE (375px)、iPad (1024px)、Desktop (1440px) 三种尺寸
- 功能测试：各断点下导航、表格、卡片功能完整
- E2E 测试：Playwright 多 viewport 验证

---

#### 3.3.16 MB-05: 启动性能优化

**需求关联**: PRD 3.3 MB-05

**技术方案**:

```
┌─────────────────────────────────────────────────────┐
│  iOS 启动优化                                        │
│                                                     │
│  App Launch → didBecomeActive                       │
│       │                                             │
│       ├─ [同步] 认证状态检查 (Keychain token)        │
│       ├─ [同步] 用户基本信息 (CoreData 缓存)         │
│       │                                             │
│       ├─ [异步] 积分余额、等级、订阅状态              │
│       ├─ [异步] RFM 数据、推荐数据                   │
│       └─ [异步] 远程配置拉取                         │
│                                                     │
│  优化手段:                                           │
│  - 减少AppDelegate同步任务                          │
│  - SwiftUI .task {} 延迟加载非首屏数据              │
│  - Instruments Time Profiler 分析启动耗时           │
├─────────────────────────────────────────────────────┤
│  Android 启动优化                                    │
│                                                     │
│  App Launch → onCreate → setContent                 │
│       │                                             │
│  优化手段:                                           │
│  - Baseline Profiles (AGP 自动生成)                  │
│  - Compose LazyColumn 延迟加载                      │
│  - DataStore 异步读取 token                         │
│  - 非关键 Module 延迟初始化                          │
└─────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `ios/AccountCenter/AppDelegate.swift` — 减少同步任务，拆分启动流程
- `ios/AccountCenter/Services/StartupManager.swift` — 启动任务编排
- `android/.../MainActivity.kt` — Baseline Profiles 集成
- `android/.../di/AppModule.kt` — 延迟注入非关键依赖

**测试策略**:
- 性能测试：Xcode Metrics / Android Profiler 测量冷/热启动耗时
- 埋点验证：`app_cold_start_duration`、`app_warm_start_duration` 按 OS/设备型号统计
- 回归测试：确保优化后功能无损失

---

#### 3.3.17 MB-06: 离线能力

**需求关联**: PRD 3.3 MB-06  
**依赖**: MB-09（Token 安全存储）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  离线数据层                                               │
│                                                          │
│  ┌─────────────────┐  在线  ┌───────────────────────┐   │
│  │ 远程 API        │ ─────→ │ CoreData (iOS)         │   │
│  │ /api/v1/credits │ ←───── │ Room (Android)         │   │
│  │ /api/v1/profile │  同步  │                        │   │
│  └─────────────────┘       │ 本地缓存:              │   │
│                            │ - credits_balance       │   │
│                            │ - user_level            │   │
│                            │ - subscription_status   │   │
│                            │ - recent_credit_changes │   │
│                            │   (最近 10 条)          │   │
│                            └───────────┬─────────────┘   │
│                                        │                  │
│  网络状态监听                          │                  │
│  NWPathMonitor (iOS)                   │ 自动同步         │
│  ConnectivityManager (Android)         │ (网络恢复时)     │
│                                        ▼                  │
│                            服务端数据覆盖本地              │
│                            UI 自动刷新                    │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `ios/AccountCenter/Persistence/CacheManager.swift` — CoreData 缓存管理
- `ios/AccountCenter/Services/SyncManager.swift` — 网络恢复后自动同步
- `android/.../data/local/AppDatabase.kt` — Room 数据库定义
- `android/.../data/repository/OfflineRepository.kt` — 离线优先 Repository
- `android/.../util/NetworkMonitor.kt` — 网络状态监听

**安全要求**:
- 本地缓存使用 `NSFileProtectionComplete`（iOS）/ `EncryptedSharedPreferences`（Android）
- 登出时 `deleteAll()` 清除本地数据

**测试策略**:
- 单元测试：缓存读写、冲突解决（服务端优先）、离线状态判断
- 集成测试：在线→离线→在线完整切换，数据一致性验证
- 安全测试：验证本地数据加密存储

---

#### 3.3.18 MB-07: 图片/资源优化

**需求关联**: PRD 3.3 MB-07

**技术方案**:

```
┌────────────────────────────────────────────────────────┐
│  图片优化流水线                                         │
│                                                        │
│  设计稿 (PNG/PSD)                                      │
│       │                                                │
│       ▼                                                │
│  构建脚本 (sharp/libvips)                               │
│  ├─ → WebP (质量 80, 用于 <picture> 标签)              │
│  ├─ → AVIF (质量 65, 用于支持的浏览器)                  │
│  └─ → PNG (fallback)                                   │
│       │                                                │
│       ▼                                                │
│  CDN (CloudFront/阿里云 CDN)                            │
│  ├─ Accept: image/avif → AVIF                          │
│  ├─ Accept: image/webp → WebP                          │
│  └─ fallback → PNG                                     │
│                                                        │
│  客户端加载策略:                                        │
│  ├─ 首屏图片: 优先加载, <img loading="eager">          │
│  ├─ 非首屏: 延迟加载, <img loading="lazy">             │
│  └─ 推荐海报: 按需下载（用户触发分享时）                 │
│                                                        │
│  渐进式加载:                                            │
│  blur placeholder (20px 宽) → 低质量 → 全尺寸           │
└────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `web/scripts/optimize-images.js` — 图片格式转换脚本
- `web/src/composables/useLazyImage.ts` — 延迟加载 composable
- `web/src/components/ui/ProgressiveImage.vue` — 渐进式图片组件
- `ios/AccountCenter/Extensions/ImageCache.swift` — iOS LRU 图片缓存
- `android/.../ui/util/ImageLoader.kt` — Android Coil 图片加载配置

**测试策略**:
- 构建测试：WebP/AVIF 体积比 PNG 减少 ≥30%
- 性能测试：图片加载耗时、缓存命中率
- 兼容性测试：各浏览器/平台的格式降级

---

#### 3.3.19 MB-08: 骨架屏加载状态

**需求关联**: PRD 3.3 MB-08  
**依赖**: MB-01（设计系统组件库）

**技术方案**:

为各主要页面创建与内容结构匹配的骨架屏组件，替代现有通用 Loading overlay。

**骨架屏页面清单**:

| 页面 | 骨架屏结构 |
|------|-----------|
| 首页仪表盘 | 3 个卡片占位（用户信息+积分+订阅状态） |
| 积分页 | 余额数字占位 + 变动列表 5 行 |
| 订阅页 | 当前等级卡片 + 权益列表 4 行 |
| 推荐页 | 统计数据 3 格 + 推荐列表 3 行 |
| 订单列表 | 订单卡片 3 行 |

**动画效果**: 脉冲渐变（`@keyframes pulse { 0% { opacity: 0.6 } → 100% { opacity: 1 } }`），与暗色科技风一致

**关键代码路径**:
- `web/src/components/ui/Skeleton.vue` — 通用骨架屏基础组件
- `web/src/components/skeleton/DashboardSkeleton.vue` — 仪表盘骨架屏
- `web/src/components/skeleton/CreditsSkeleton.vue` — 积分页骨架屏
- `ios/AccountCenter/Components/SkeletonView.swift` — iOS 骨架屏组件
- `android/.../ui/components/Skeleton.kt` — Android 骨架屏组件

**测试策略**:
- 组件测试：骨架屏渲染、过渡动画、数据加载完成后平滑切换
- 视觉测试：各骨架屏与暗色科技风设计一致性
- 性能测试：骨架屏渲染时间 ≤16ms（60fps）

---

#### 3.3.20 MB-11: Root/越狱检测

**需求关联**: PRD 3.3 MB-11

**技术方案**:

```
┌─────────────────────────────────────────────────────────┐
│  iOS 越狱检测                                            │
│  ┌───────────────────────────────────────────┐          │
│  │ 1. 检查 Cydia.app 路径存在性               │          │
│  │ 2. 检测沙盒完整性 (fopen /private/test)    │          │
│  │ 3. 检查异常符号链接                        │          │
│  │ 4. 检测 DYLD_INSERT_LIBRARIES 环境变量     │          │
│  │ 5. 代码完整性校验 (NSBundle hash)          │          │
│  └───────────────────────────────────────────┘          │
│                                                         │
│  Android Root 检测                                       │
│  ┌───────────────────────────────────────────┐          │
│  │ Google Play Integrity API:                 │          │
│  │ 1. 请求 IntegrityToken                     │          │
│  │ 2. 服务端验证 token                        │          │
│  │ 3. 检测 MEETS_BASIC_INTEGRITY              │          │
│  │                                            │          │
│  │ 本地辅助检测:                               │          │
│  │ 1. 检查 su / Superuser.apk                 │          │
│  │ 2. 检查 /system/app/Superuser.apk          │          │
│  │ 3. 检测 Build.TAGS = "test-keys"           │          │
│  └───────────────────────────────────────────┘          │
│                                                         │
│  检测到后行为:                                           │
│  → 展示安全警告弹窗                                     │
│  → 限制支付/密码修改/身份信息查看                         │
│  → 上报服务端 POST /api/v1/security/device-integrity     │
└─────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `ios/AccountCenter/Security/JailbreakDetector.swift` — 越狱检测
- `android/.../security/RootDetector.kt` — Root 检测
- `android/.../security/PlayIntegrityChecker.kt` — Play Integrity API 调用
- `auth-service/internal/handler/device_integrity_handler.go` — 服务端完整性验证
- `compliance-service/internal/service/risk_service.go` — 设备完整性风险记录

**API 变更**:
- `POST /api/v1/security/device-integrity` — 上报设备完整性检测结果

**测试策略**:
- 单元测试：检测逻辑（mock 文件系统检查）
- 集成测试：越狱/root 设备上运行，验证检测率和限制行为
- 防绕过测试：代码混淆后验证检测逻辑不可轻易绕过

---

#### 3.3.21 MB-12: 应用截屏防护

**需求关联**: PRD 3.3 MB-12

**技术方案**:

```
┌─────────────────────────────────────────────────────────┐
│  iOS 截屏监听                                            │
│  ┌───────────────────────────────────────────┐          │
│  │ NotificationCenter.default.addObserver    │          │
│  │   forName: UIScreen.capturedDidChange     │          │
│  │                                            │          │
│  │ UIScreen.main.isCaptured == true           │          │
│  │ → 展示"检测到截屏/录屏"提醒                │          │
│  │ → 埋点 screenshot_detected                │          │
│  └───────────────────────────────────────────┘          │
│                                                         │
│  Android 截屏防护                                        │
│  ┌───────────────────────────────────────────┐          │
│  │ window.setFlags(                           │          │
│  │   LayoutParams.FLAG_SECURE,                │          │
│  │   LayoutParams.FLAG_SECURE                 │          │
│  │ )                                          │          │
│  │ → 截屏/录屏内容为黑色                      │          │
│  └───────────────────────────────────────────┘          │
│                                                         │
│  敏感页面自动启用:                                       │
│  - 密码输入页                                            │
│  - 支付确认页                                            │
│  - 身份信息页                                            │
│  离开页面 → 自动解除                                     │
└─────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `ios/AccountCenter/Security/ScreenshotMonitor.swift` — 截屏监听和提醒
- `ios/AccountCenter/Views/SensitiveScreenModifier.swift` — SwiftUI ViewModifier 敏感页面标记
- `android/.../security/ScreenshotProtection.kt` — FLAG_SECURE 管理器
- `android/.../ui/theme/SensitiveActivity.kt` — 敏感页面基类

**测试策略**:
- 手动测试：敏感页面截屏/录屏验证防护效果
- 自动化测试：页面切换时防护启用/解除的生命周期管理
- 埋点验证：`screenshot_detected` 事件包含 `page_type` 和 `user_id`

---

#### 3.3.22 MB-15: 小程序分包加载

**需求关联**: PRD 3.3 MB-15

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  小程序分包结构 (app.json)                                │
│                                                          │
│  {                                                       │
│    "pages": [                                            │
│      "pages/index/index",           # 首页              │
│      "pages/login/login",           # 登录              │
│      "pages/profile/profile"        # 个人中心           │
│    ],                                                    │
│    "subpackages": [                                      │
│      {                                                   │
│        "root": "packageSubscribe",                       │
│        "name": "subscribe",                              │
│        "pages": ["pages/subscribe", "pages/orders"]      │
│      },                                                  │
│      {                                                   │
│        "root": "packageReferral",                        │
│        "name": "referral",                               │
│        "pages": ["pages/referral", "pages/share"]        │
│      },                                                  │
│      {                                                   │
│        "root": "packageHelp",                            │
│        "name": "help",                                   │
│        "pages": ["pages/faq", "pages/contact"]           │
│      }                                                   │
│    ],                                                    │
│    "preloadRule": {                                      │
│      "pages/index/index": {                              │
│        "network": "wifi",                                │
│        "packages": ["subscribe", "referral"]             │
│      }                                                   │
│    }                                                     │
│  }                                                       │
│                                                          │
│  主包 ≤ 2MB, 总包 ≤ 20MB                                │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `miniprogram/app.json` — 分包配置和预加载规则
- `miniprogram/packageSubscribe/` — 订阅分包
- `miniprogram/packageReferral/` — 推荐分包
- `miniprogram/packageHelp/` — 帮助分包
- `scripts/miniprogram-size-check.sh` — 主包体积检查脚本

**测试策略**:
- 构建测试：`ci` 环境构建后验证主包 ≤2MB
- 性能测试：4G 环境首次启动 ≤3s
- 功能测试：分包内页面跳转、预加载验证、分包加载失败降级

---

#### 3.3.23 MB-20~21: 广告数据埋点+无广告升级引导

**需求关联**: PRD 3.3 MB-20~21  
**依赖**: FN-12（事件埋点 SDK）、MB-16~19（广告变现基础）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  广告埋点事件                                             │
│                                                          │
│  ad_splash_shown     → 开屏广告展示                      │
│  ad_splash_skipped   → 开屏广告跳过                      │
│  ad_banner_shown     → Banner 广告展示                   │
│  ad_provider_switched → 主/备 SDK 切换                   │
│  ad_load_failed      → 广告加载失败                      │
│                                                          │
│  维度: level / platform / sdk / ad_position              │
│  → data-product-service 事件收集                         │
│  → Grafana 广告数据大屏                                  │
│                                                          │
│  无广告升级引导:                                          │
│  ┌────────────────────────────┐                          │
│  │ L0/L1 用户关闭/跳过广告时  │                          │
│  │ → 展示轻量提示             │                          │
│  │ "升级 Premium 去除广告"    │                          │
│  │ 展示时间 ≤1.5s             │                          │
│  │ 频率: 每日最多 1 次        │                          │
│  │ Redis: ad:upgrade_prompt:  │                          │
│  │   {user_id}:{date}         │                          │
│  │ → 点击跳转定价页            │                          │
│  └────────────────────────────┘                          │
│                                                          │
│  埋点:                                                   │
│  ad_upgrade_prompt_shown                                 │
│  ad_upgrade_prompt_clicked                               │
│  ad_upgrade_completed                                    │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `android/.../ad/AdEventListener.kt` — 广告事件回调→FN-12 SDK 上报
- `ios/AccountCenter/Ads/AdEventReporter.swift` — 广告事件上报
- `android/.../ui/ad/UpgradePromptDialog.kt` — 升级引导弹窗
- `ios/AccountCenter/Views/AdUpgradePrompt.swift` — iOS 升级引导
- `credit-service/internal/handler/ad_handler.go` — 广告数据查询 API

**API 变更**:
- `GET /api/v1/admin/ads/metrics` — 广告数据指标查询（展示量/跳过率/填充率）
- `GET /api/v1/ads/upgrade-prompt-eligibility` — 查询用户是否可展示升级引导

**测试策略**:
- 埋点测试：验证 5 个广告事件正确上报到 data-product-service
- 频率测试：升级引导每日最多展示 1 次（Redis 计数验证）
- 转化测试：引导→点击→定价页→支付完整链路埋点

---

#### 3.3.24 AR-03: 服务发现与注册

**需求关联**: PRD 3.3 AR-03

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  Consul 服务注册/发现                                     │
│                                                          │
│  ┌──────────────────┐  注册  ┌──────────────────────┐   │
│  │ api-gateway:30300 │ ────→ │ Consul Cluster       │   │
│  │ account:30301     │       │ (≥3 节点)            │   │
│  │ auth:30302        │       │                      │   │
│  │ notification:30311│       │ HTTP API: 8500       │   │
│  │ credit:30312      │       │ DNS: 8600            │   │
│  │ compliance:30313  │       │                      │   │
│  │ data-product:30314│       │ 健康检查间隔: 10s    │   │
│  │ config:30315      │       │                      │   │
│  │ payment:30316     │       │                      │   │
│  └──────────────────┘       └──────────┬───────────┘   │
│                                        │                │
│  查询目标地址                          │                │
│  ┌──────────────────┐  ←──────────────┘                │
│  │ api-gateway       │                                  │
│  │ DNS: account-     │                                  │
│  │ service.service.  │                                  │
│  │ consul:30301      │                                  │
│  │ → 动态路由        │                                  │
│  └──────────────────┘                                   │
│                                                          │
│  降级: Consul 故障时使用本地缓存地址继续通信              │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `pkg/discovery/registry.go` — Consul 注册/注销/发现封装
- `pkg/discovery/resolver.go` — DNS 或 HTTP API 服务发现客户端
- 各服务 `cmd/main.go` — 启动时调用 `registry.Register()`，停止时 `registry.Deregister()`
- `api-gateway/internal/proxy/reverse_proxy.go` — 使用服务发现获取目标地址
- `infra/consul/` — Consul 集群部署配置

**配置项**（通过 config-service 管理）:
- `discovery.consul_addresses`: Consul 集群地址列表
- `discovery.health_check_interval_sec`: 10
- `discovery.deregister_critical_after_sec`: 30

**测试策略**:
- 单元测试：服务注册/注销/发现逻辑、本地缓存降级
- 集成测试：服务上下线后 ≤15s 内其他服务感知变更
- 混沌测试：Consul 节点故障，验证服务通信不受影响

---

#### 3.3.25 AR-08: 日志关联优化

**需求关联**: PRD 3.3 AR-08  
**依赖**: AR-05（分布式追踪提供 trace_id/span_id）

**技术方案**:

扩展现有 `pkg/logging/logger.go`，从 OpenTelemetry context 提取 `trace_id` 和 `span_id` 并注入 slog 日志字段。

```
┌──────────────────────────────────────────────────────────┐
│  当前: slog JSON 日志                                    │
│  { "service":"auth-service", "request_id":"xxx", ... }   │
│                                                          │
│  优化后: 追加 trace_id / span_id                         │
│  {                                                       │
│    "service": "auth-service",                            │
│    "request_id": "xxx",                                  │
│    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",       │
│    "span_id": "00f067aa0ba902b7",                        │
│    "msg": "user login success"                           │
│  }                                                       │
│                                                          │
│  LoggerFromContext 改造:                                  │
│  1. 从 ctx 提取 otel trace.SpanContext                   │
│  2. 添加 trace_id/span_id 到 logger attributes           │
│  3. Loki 通过 {trace_id="xxx"} 查询关联日志              │
│  4. Grafana 提供 Loki→Jaeger 跳转链接                    │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `pkg/logging/logger.go` — 修改 `LoggerFromContext()`，从 `otel trace.SpanFromContext(ctx)` 提取 ID
- `pkg/logging/middleware.go` — Gin 中间件确保 context 传播 trace context
- `infra/grafana/dashboards/` — 日志面板配置 Loki→Jaeger 数据链接
- `infra/loki/loki-config.yaml` — Loki 索引 `trace_id` 字段

**测试策略**:
- 单元测试：验证 `LoggerFromContext` 正确提取 trace_id/span_id
- 集成测试：跨服务调用，验证同一 trace_id 在各服务日志中出现
- Grafana 测试：日志面板一键跳转 Jaeger 追踪页面

---

#### 3.3.26 AR-09: 数据库迁移 Down 脚本

**需求关联**: PRD 3.3 AR-09

**技术方案**:

为所有现有和新增的 Goose migration 文件补充 `Down()` 回滚脚本，并添加 CI 验证步骤。

**Goose migration 文件格式**:
```sql
-- +goose Up
CREATE TABLE new_table (
    id UUID PRIMARY KEY,
    name VARCHAR(100)
);

-- +goose Down
DROP TABLE IF EXISTS new_table;
```

**CI 验证步骤**（新增 GitHub Actions job）:
```yaml
- name: Verify migration up/down
  run: |
    goose -dir migrations postgres "$DATABASE_URL" up
    goose -dir migrations postgres "$DATABASE_URL" down
    goose -dir migrations postgres "$DATABASE_URL" up
```

**关键代码路径**:
- 各服务 `migrations/` 目录 — 所有 `.sql` 文件补充 `-- +goose Down` 段
- `.github/workflows/ci.yml` — 新增 migration 验证步骤
- `docs/migrations/` — migration 变更说明文档

**测试策略**:
- CI 测试：`goose up` → `goose down` → `goose up` 三步验证
- 环境测试：dev/staging/prod 三个环境均可执行 down
- 评审机制：生产环境 `goose down` 需人工审批

---

#### 3.3.27 AR-10: Redis 高可用配置验证

**需求关联**: PRD 3.3 AR-10  
**依赖**: AR-21（K8s Helm Chart）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  Redis 高可用部署 (Sentinel 模式)                         │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Redis Master │  │ Redis Slave1 │  │ Redis Slave2 │  │
│  │ (read/write) │←─│ (read only)  │←─│ (read only)  │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                  │          │
│  ┌──────▼─────────────────▼──────────────────▼───────┐  │
│  │ Sentinel 1    Sentinel 2    Sentinel 3            │  │
│  │ (监控+故障转移)                                      │  │
│  │ 主节点故障 → Sentinel 投票选举新主 → ≤30s           │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  持久化配置:                                              │
│  save 900 1            # 15 分钟 1 次 RDB                │
│  save 300 10           # 5 分钟 10 次                    │
│  appendonly yes        # AOF 开启                       │
│  appendfsync everysec  # 每秒 fsync                     │
│  aof-use-rdb-preamble yes  # RDB+AOF 混合               │
└──────────────────────────────────────────────────────────┘
```

**Go 客户端配置**:
```go
// 使用 go-redis Sentinel 客户端
rdb := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{":26379", ":26380", ":26381"},
    PoolSize:      100,
    MinIdleConns:  10,
    DialTimeout:   5 * time.Second,
    ReadTimeout:   3 * time.Second,
    WriteTimeout:  3 * time.Second,
})
```

**关键代码路径**:
- `helm/account-center/templates/_infra/redis-sentinel.yaml` — Sentinel StatefulSet
- 各服务 Redis 客户端初始化 — 从直连改为 Sentinel 模式
- `infra/redis/failover-drill.sh` — 主从切换演练脚本

**测试策略**:
- 演练测试：手动 kill 主节点，验证 Sentinel ≤30s 内完成切换
- 客户端测试：主从切换期间请求重连，无数据丢失
- 数据一致性：切换前后数据验证
- 演练报告：记录 RTO 实测值

---

#### 3.3.28 AR-11: 数据库连接池调优

**需求关联**: PRD 3.3 AR-11  
**依赖**: AR-19（性能/压力测试）、NF-05（配置热更新）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  连接池调优流程                                           │
│                                                          │
│  1. 基线测量 (AR-19 压测数据)                             │
│     → 当前 max_open/max_idle/conn_max_lifetime           │
│     → Prometheus 指标采集                                 │
│                                                          │
│  2. 调优参数                                             │
│     max_open_conns: 25 (默认) → 调优后值                  │
│     max_idle_conns: 10 (默认) → 调优后值                  │
│     conn_max_lifetime: 30m (默认) → 调优后值              │
│     conn_max_idle_time: 5m → 调优后值                    │
│                                                          │
│  3. Prometheus 指标                                       │
│     db_pool_open_connections                             │
│     db_pool_in_use                                       │
│     db_pool_idle                                         │
│     db_pool_wait_count                                   │
│     db_pool_wait_duration                                │
│     db_pool_max_idle_closed                              │
│     db_pool_max_lifetime_closed                          │
│                                                          │
│  4. 动态调整 (NF-05 热更新)                               │
│     config-service 管理:                                  │
│     db_pool_max_open / db_pool_max_idle                  │
│     → 通过 SetMaxOpenConns/SetMaxIdleConns 热更新        │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `pkg/database/pool.go` — 连接池初始化 + Prometheus 指标采集
- `pkg/database/pool_watcher.go` — 热更新监听（配合 NF-05）
- 各服务 `cmd/main.go` — 使用 `pkg/database` 统一初始化
- `infra/grafana/dashboards/db-pool.json` — 连接池监控面板

**测试策略**:
- 压测：500 并发下连接池利用率 <80%，无等待超时
- 热更新测试：运行时修改 `db_pool_max_open`，验证连接池参数实时生效
- Prometheus 指标验证：Grafana 面板正确展示连接池指标

---

#### 3.3.29 AR-20: E2E 测试

**需求关联**: PRD 3.3 AR-20  
**依赖**: AR-18（集成测试）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  E2E 测试框架                                             │
│                                                          │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐     │
│  │ XCUITest     │ │ Compose      │ │ Playwright   │     │
│  │ (iOS)        │ │ Testing      │ │ (Web)        │     │
│  │              │ │ (Android)    │ │              │     │
│  │ 核心流程:    │ │ 核心流程:    │ │ 核心流程:    │     │
│  │ 注册→登录    │ │ 注册→登录    │ │ 注册→登录    │     │
│  │ →浏览→订阅   │ │ →浏览→订阅   │ │ →浏览→订阅   │     │
│  │ →积分→推荐   │ │ →积分→推荐   │ │ →积分→推荐   │     │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘     │
│         │                │                │              │
│         └────────────────┼────────────────┘              │
│                          │ 三端并行执行                    │
│                          ▼                                │
│                  ┌──────────────┐                         │
│                  │ 测试报告      │                         │
│                  │ HTML + 截图   │                         │
│                  │ + 录屏       │                         │
│                  └──────────────┘                         │
│                                                          │
│  执行环境: 独立测试账号 + staging API                     │
│  触发: 每日定时 / 发版前手动触发                          │
│  超时: ≤30 分钟                                          │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `tests/e2e/ios/AccountCenterUITests/` — XCUITest 用例
- `tests/e2e/android/app/src/androidTest/` — Compose Testing 用例
- `tests/e2e/web/tests/` — Playwright 用例
- `tests/e2e/helpers/TestAccountManager.swift|kt|ts` — 测试账号管理
- `.github/workflows/e2e.yml` — E2E 测试 CI workflow

**测试策略**:
- 核心流程覆盖：注册→登录→浏览→订阅购买→积分查看→推荐分享
- 失败处理：自动截图/录屏附加报告，含失败步骤和预期/实际结果
- 数据隔离：独立测试账号，测试后清理

---

#### 3.3.30 AR-24: 金丝雀发布策略

**需求关联**: PRD 3.3 AR-24  
**依赖**: AR-21（K8s Helm Chart）、AR-07（告警规则）

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  K8s Ingress 金丝雀发布                                   │
│                                                          │
│  nginx.ingress.kubernetes.io/canary: "true"              │
│                                                          │
│  ┌──────────────┐    5%     ┌──────────────────┐        │
│  │              │ ─────────→│ canary Deployment │        │
│  │              │           │ (新版本)           │        │
│  │   Ingress    │           └──────────────────┘        │
│  │   Controller │                                       │
│  │              │   95%    ┌──────────────────┐         │
│  │              │ ────────→│ stable Deployment │         │
│  └──────────────┘          │ (旧版本)           │         │
│                            └──────────────────┘         │
│                                                          │
│  发布阶段:                                                │
│  5% → 监控 10min → 25% → 监控 10min                     │
│  → 50% → 监控 10min → 100%                              │
│                                                          │
│  自动回滚条件:                                            │
│  - 错误率 >0.5% (持续 3 分钟)                            │
│  - P99 延迟超基线 50%                                    │
│  → kubectl rollout undo                                 │
│                                                          │
│  用户标签定向:                                            │
│  canary-by-header: X-User-Level=L4                      │
│  canary-by-cookie: canary_group=beta                     │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- `helm/account-center/templates/api-gateway/canary.yaml` — 金丝雀 Ingress 模板
- `scripts/deploy/canary-promote.sh` — 金丝雀推进脚本（5%→25%→50%→100%）
- `scripts/deploy/canary-rollback.sh` — 自动回滚脚本
- `infra/grafana/dashboards/canary.json` — 金丝雀监控面板（新旧版本对比）

**Helm values 配置**:
```yaml
canary:
  enabled: true
  weight: 5
  maxWeight: 100
  stepWeight: 20
  stepInterval: 10m
  rollbackOn:
    errorRate: 0.005
    p99LatencyMs: 1500
```

**测试策略**:
- 集成测试：部署 canary 版本后验证流量分配比例
- 回滚测试：注入错误，验证自动回滚在 3 分钟内触发
- 监控测试：Grafana 面板正确展示新旧版本指标

---

#### 3.3.31 AR-26: 共享中间件包提取

**需求关联**: PRD 3.3 AR-26

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  提取前 (各服务 main.go 重复代码):                        │
│                                                          │
│  api-gateway/main.go (461行)                             │
│  config-service/main.go (255行)                          │
│  auth-service/main.go                                    │
│  ...                                                     │
│  重复: metrics / health / shutdown / config loading      │
│                                                          │
│  提取后:                                                 │
│  pkg/server/                                             │
│  ├── server.go        — 标准启动模板                     │
│  ├── metrics.go       — /metrics handler (Prometheus)    │
│  ├── health.go        — /health handler (依赖检测)       │
│  ├── shutdown.go      — graceful shutdown                 │
│  ├── config.go        — 统一配置加载模式                  │
│  └── middleware.go    — 通用 Gin 中间件注册               │
│                                                          │
│  各服务 main.go 仅包含:                                   │
│  - 业务路由注册                                           │
│  - 业务特定中间件                                         │
│  - 预期: ≤80 行/文件 (减少 ≥50%)                         │
└──────────────────────────────────────────────────────────┘
```

**pkg/server 标准启动模板**:
```go
type ServerConfig struct {
    ServiceName     string
    Port            string
    ConfigClient    *config.Client
    RegisterRoutes  func(*gin.Engine)
    CustomMiddleware []gin.HandlerFunc
    HealthChecks    map[string]func(context.Context) error
}

func Run(cfg ServerConfig) error { ... }
```

**关键代码路径**:
- `pkg/server/server.go` — 标准启动模板
- `pkg/server/metrics.go` — 通用 `/metrics` handler（从各 main.go 提取）
- `pkg/server/health.go` — 通用 `/health` handler（支持依赖检测）
- `pkg/server/shutdown.go` — graceful shutdown（从各 main.go 提取）
- 各服务 `cmd/main.go` — 重构为使用 `pkg/server.Run()`

**测试策略**:
- 单元测试：`pkg/server` 各模块覆盖率 ≥80%
- 集成测试：每个服务使用 `pkg/server` 启动，功能不变
- 回归测试：确保 metrics/health/shutdown 行为与重构前一致

---

#### 3.3.32 AR-27: API 文档自动生成

**需求关联**: PRD 3.3 AR-27

**技术方案**:

```
┌──────────────────────────────────────────────────────────┐
│  Swaggo 集成流程                                          │
│                                                          │
│  开发代码 + 注解                                          │
│       │                                                  │
│       ▼                                                  │
│  // @Summary 创建订单                                     │
│  // @Description 创建新的订阅订单                          │
│  // @Tags orders                                          │
│  // @Accept json                                          │
│  // @Produce json                                         │
│  // @Param body body CreateOrderRequest true "请求体"     │
│  // @Success 201 {object} OrderResponse                   │
│  // @Failure 400 {object} ErrorResponse                   │
│  // @Router /api/v1/orders [post]                         │
│  func (h *OrderHandler) Create(c *gin.Context) { ... }    │
│       │                                                  │
│       ▼                                                  │
│  CI: swag init → docs/swagger.json (OpenAPI 3.0)          │
│       │                                                  │
│       ▼                                                  │
│  Swagger UI 站点 (Docker 容器, port 30320)                │
│                                                          │
│  同步检查:                                                │
│  CI 对比 swag init 输出与 git tracked docs/               │
│  不一致 → build warning                                  │
└──────────────────────────────────────────────────────────┘
```

**关键代码路径**:
- 各服务 `internal/handler/*.go` — 添加 Swaggo 注解
- 各服务 `docs/` — `swag init` 生成的 `swagger.json` + `swagger.yaml`
- `scripts/swagger-serve.sh` — 本地启动 Swagger UI
- `.github/workflows/ci.yml` — 新增 `swag init` + diff 检查步骤
- `infra/swagger-ui/` — Swagger UI Docker 部署配置

**注解覆盖进度**:
- P0 阶段完成后覆盖 P0 新增 API（payment-service 全部、admin API）
- P2 阶段覆盖全部现有接口（100%）

**测试策略**:
- CI 检查：`swag init` 输出与 `docs/` 一致，否则构建警告
- 完整性检查：对比 Gin 路由注册与 Swagger `@Router` 注解，检测遗漏
- 人工审查：API 文档可读性、请求/响应示例准确性

---

### 3.3 P2 技术设计小结

| 需求 ID | 技术方案核心 | 涉及服务/端 | 新增文件估计 |
|---------|------------|------------|-------------|
| NF-05 | ConfigWatcher 定时轮询 + atomic.Value 热更新 | 全服务 | ~6 文件 |
| NF-06 | iOS Universal Links + Android App Links + neuro:// scheme | 移动端 + Web | ~10 文件 |
| NF-07 | main.go 461 行拆分为 middleware/ + proxy/ + main.go | api-gateway | ~10 文件 |
| UX-03 | iOS SMS AutoFill + Android SMS Retriever API | 移动端 + notification-service | ~5 文件 |
| UX-04 | guest_token + 邮箱注册 + 手机绑定渐进漏斗 | auth-service + 全端 | ~12 文件 |
| UX-06 | EmptyStateCard 组件 + config-service 文案配置 | 全端 | ~6 文件 |
| UX-13 | 海报模板渲染 + Open Graph + 小程序分享 | credit-service + 全端 | ~10 文件 |
| UX-15 | notifications 表 + 消息中心 API + Push 联动 | notification-service + 全端 | ~12 文件 |
| UX-16 | faq_items 表 + 全文搜索 + 智能客服入口 | config-service + 全端 | ~10 文件 |
| FN-03 | 第三方发票平台 API + 邮件推送 PDF + 重试 | payment-service | ~8 文件 |
| FN-09 | 通知模板 CRUD + 变量插值 + 发送记录 | notification-service | ~10 文件 |
| FN-11 | 频率限制 Redis + 免打扰时段 + 标签定向 + Asynq 定时 | notification-service | ~8 文件 |
| FN-13 | Kafka 事件流 + 实时聚合 Redis + 异常检测 | data-product-service | ~10 文件 |
| MB-01 | Design Token JSON → 四端代码导出 + Storybook | 全端 | ~30+ 文件 |
| MB-03 | CSS Media Query 三断点 + Vue composable | Web | ~6 文件 |
| MB-05 | iOS/Android 启动流程优化 + 延迟加载 | 移动端 | ~6 文件 |
| MB-06 | CoreData/Room 本地缓存 + 自动同步 | 移动端 | ~10 文件 |
| MB-07 | CDN + WebP/AVIF + 渐进式加载 + LRU 缓存 | 全端 + CDN | ~8 文件 |
| MB-08 | 页面匹配骨架屏组件 + 脉冲动画 | 全端 | ~12 文件 |
| MB-11 | iOS 越狱检测 + Android Play Integrity API | 移动端 + auth-service | ~6 文件 |
| MB-12 | iOS isCaptured + Android FLAG_SECURE | 移动端 | ~5 文件 |
| MB-15 | 小程序 app.json 分包 + preloadRule 预加载 | 小程序 | ~8 文件 |
| MB-20~21 | 5 个广告事件埋点 + Redis 频控升级引导 | 移动端 + data-product-service | ~8 文件 |
| AR-03 | Consul 服务注册/发现 + 动态路由 + 本地缓存降级 | 全服务 | ~6 文件 |
| AR-08 | slog LoggerFromContext 注入 trace_id/span_id | 全服务 + Grafana | ~4 文件 |
| AR-09 | Goose Down() 脚本 + CI up/down/up 验证 | 全服务 | ~20 脚本修改 |
| AR-10 | Redis Sentinel 3 节点 + RDB/AOF 混合 + 故障转移演练 | 基础设施 + 全服务 | ~6 文件 |
| AR-11 | 连接池参数调优 + Prometheus 指标 + NF-05 热更新 | 全服务 | ~4 文件 |
| AR-20 | XCUITest + Compose Testing + Playwright 三端 E2E | 全端 | ~20 文件 |
| AR-24 | K8s Ingress canary + 自动回滚 + 用户标签定向 | 基础设施 | ~6 文件 |
| AR-26 | pkg/server 共享包 + 各服务 main.go 精简 | 全服务 | ~8 文件 |
| AR-27 | Swaggo 注解 + CI 自动生成 OpenAPI 3.0 | 全服务 | ~50+ 注解 |

---

### 3.4 Phase 9 — P3 技术设计（8 项）

> 以下 8 项为长期规划需求的技术设计方案，Phase 8（P2）全部完成后启动。

---

#### 3.4.1 UX-07: 搜索/快捷操作

**需求关联**: PRD 3.4 UX-07

**技术方案**:

采用"搜索索引 + 快捷操作注册表"双引擎架构。搜索索引基于 Redis 存储操作/页面的关键词倒排表，支持拼音首字母和模糊匹配；快捷操作注册表记录操作元数据（名称、路由、图标、分类），前端按使用频率动态排序。

```
┌──────────────────────────────────────────────────────────────┐
│  前端（Web/iOS/Android）                                      │
│                                                              │
│  ┌──────────────────┐     ┌──────────────────────────────┐   │
│  │  搜索框/快捷键     │     │  快捷操作面板                 │   │
│  │  Cmd+K / 全局入口 │     │  高频操作列表（频率排序）      │   │
│  └────────┬─────────┘     └──────────────┬───────────────┘   │
│           │                              │                    │
│           │ 关键词 / 拼音                │ 点击操作            │
│           ▼                              ▼                    │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Search Orchestrator                                  │    │
│  │  1. 本地索引匹配（页面名称 + 拼音首字母）              │    │
│  │  2. 匹配结果按 category + frequency 排序              │    │
│  │  3. 响应 ≤300ms                                      │    │
│  └──────────────────────┬───────────────────────────────┘    │
└─────────────────────────┼────────────────────────────────────┘
                          │ 埋点上报
                          ▼
                 ┌─────────────────┐
                 │  data-product   │
                 │  -service       │
                 │  quick_action_  │
                 │  triggered      │
                 │  global_search_ │
                 │  query          │
                 └─────────────────┘
```

**搜索索引数据结构**（Redis）:

```
Key: ac:search:items:{category}
  → Sorted Set, score = 使用频率, member = JSON{id, name, pinyin, route, icon, keywords}

Key: ac:search:frequency:{user_id}
  → Hash, field = action_id, value = 使用次数

Key: ac:search:pinyin_map
  → Hash, field = pinyin_prefix, value = action_id 列表
```

**快捷操作注册表**（config-service 管理）:

```json
{
  "quick_actions": [
    {"id": "recharge", "name": "充值", "pinyin": "chongzhi", "keywords": ["充值","充钱","recharge"], "route": "/credits/recharge", "icon": "wallet", "category": "finance"},
    {"id": "check_credits", "name": "查积分", "pinyin": "chajifen", "keywords": ["积分","余额","credits"], "route": "/credits", "icon": "star", "category": "finance"},
    {"id": "change_password", "name": "改密码", "pinyin": "gaimima", "keywords": ["密码","修改密码","password"], "route": "/settings/password", "icon": "lock", "category": "settings"},
    {"id": "contact_support", "name": "联系客服", "pinyin": "liankefu", "keywords": ["客服","帮助","support"], "route": "/support", "icon": "headset", "category": "help"},
    {"id": "share_referral", "name": "分享推荐", "pinyin": "fenxiang", "keywords": ["推荐","分享","invite"], "route": "/referral", "icon": "share", "category": "growth"}
  ]
}
```

**前端搜索匹配逻辑**:
1. 用户输入关键词 → 去除空格，转小写
2. 本地遍历 `quick_actions` 索引：名称匹配 OR 拼音首字母匹配 OR keywords 包含
3. 匹配结果按使用频率（`ac:search:frequency:{user_id}`）降序排列
4. 未登录用户按全局默认频率排序

**关键代码路径**:
- `account-service/internal/handler/search.go` — 搜索 API handler（频率上报、操作列表）
- `account-service/internal/service/search.go` — 搜索索引管理、频率统计
- `api-gateway/internal/proxy/route.go` — 搜索路由注册
- `web-ui/src/components/GlobalSearch.vue` — Web 全局搜索组件
- `web-ui/src/composables/useQuickActions.ts` — 快捷操作 Composable
- `ios/AccountCenter/Views/GlobalSearchView.swift` — iOS 搜索页
- `android/app/src/main/java/com/neuro/accountcenter/ui/search/SearchViewModel.kt` — Android 搜索 ViewModel
- `data-product-service/internal/handler/analytics.go` — 搜索/操作埋点事件接收

**API 变更**:
- `GET /api/v1/search/actions` — 获取快捷操作列表（含用户频率排序）
- `POST /api/v1/search/actions/{id}/trigger` — 上报操作触发（更新频率）
- `GET /api/v1/search/suggest?q={keyword}` — 搜索建议（服务端兜底，前端优先本地匹配）

**测试策略**:
- 单元测试：搜索关键词匹配逻辑（中文/拼音首字母/模糊匹配）、频率排序算法
- 集成测试：搜索索引初始化 → 关键词查询 → 频率上报 → 排序变化完整链路
- 前端测试：全局搜索组件交互、快捷键（Cmd+K）唤起、搜索结果渲染
- 性能测试：搜索响应 P95 < 300ms（1000 条操作索引）

---

#### 3.4.2 UX-14: 排行榜/社交证明

**需求关联**: PRD 3.4 UX-14
**依赖**: FN-12（事件埋点 SDK 提供效果追踪能力）

**技术方案**:

采用 Redis Sorted Set 实现推荐达人排行榜，每日定时从 account-service 的推荐关系数据聚合计算。社交证明基于 account-service 的 rebate 数据实时统计。

```
┌──────────────────────────────────────────────────────────────────┐
│  排行榜计算流水线                                                  │
│                                                                  │
│  ┌────────────────┐  每日 01:00   ┌──────────────────────┐       │
│  │  account-      │ ────────────→ │  聚合 Worker          │       │
│  │  service       │  推荐关系数据  │  (Asynq Scheduled)    │       │
│  │  referrals     │              │                        │       │
│  └────────────────┘              │  1. 按周期聚合推荐数     │       │
│                                  │  2. 计算排名 Top 20      │       │
│                                  │  3. 写入 Redis ZSET      │       │
│                                  │  4. 计算社交证明数据      │       │
│                                  └────────────┬───────────┘       │
│                                               │                   │
│                                               ▼                   │
│                                  ┌──────────────────────┐        │
│                                  │  Redis               │        │
│                                  │  ZSET leaderboard:   │        │
│                                  │    weekly / monthly  │        │
│                                  │  HASH social_proof:  │        │
│                                  │    {user_id} → stats │        │
│                                  └──────────────────────┘        │
│                                                                  │
│  ┌────────────────┐  读取排行榜   ┌──────────────────────┐       │
│  │  前端           │ ←─────────── │  account-service      │       │
│  │  排行榜页       │              │  leaderboard API      │       │
│  └────────────────┘              └──────────────────────┘       │
└──────────────────────────────────────────────────────────────────┘
```

**排行榜数据结构**（Redis）:

```
Key: ac:leaderboard:weekly        → Sorted Set, score = 推荐转化数, member = user_id
Key: ac:leaderboard:monthly       → Sorted Set, score = 推荐转化数, member = user_id
Key: ac:leaderboard:profile:{uid} → Hash {nickname(masked), avatar, count, streak_days}
Key: ac:leaderboard:privacy:{uid} → String "visible" | "hidden"
Key: ac:social_proof:{uid}        → String "你的 {n} 位好友已加入"
Key: ac:social_proof:global       → String "本周已有 {n} 人升级 Premium"
```

**隐私控制**:
- 用户可在推荐设置中切换"是否在排行榜中展示"
- `ac:leaderboard:privacy:{uid}` = `hidden` 时，排行榜中该用户显示为"匿名用户"
- 默认值：`visible`

**社交证明数据来源**:
- "你的 N 位好友已加入"：从 `referrals` 表查询当前用户的社交圈注册数
- "本周已有 N 人升级 Premium"：从 `subscriptions` 表按周聚合

**关键代码路径**:
- `account-service/internal/service/leaderboard.go` — 排行榜聚合与查询逻辑
- `account-service/internal/service/social_proof.go` — 社交证明数据计算
- `account-service/internal/worker/leaderboard.go` — Asynq 定时聚合 Worker
- `account-service/internal/repository/referral.go` — 推荐关系查询（已有，扩展聚合方法）
- `web-ui/src/views/Referral.vue` — 推荐页排行榜组件扩展
- `ios/AccountCenter/Views/LeaderboardView.swift` — iOS 排行榜页
- `android/app/src/main/java/com/neuro/accountcenter/ui/referral/LeaderboardScreen.kt` — Android 排行榜

**数据库变更**:
- 新增 `leaderboard_snapshots` 表（每日排行榜快照，用于历史趋势展示）:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 快照 ID |
| period_type | VARCHAR(10) | NOT NULL | 周期类型：weekly/monthly |
| period_start | DATE | NOT NULL | 周期起始日 |
| period_end | DATE | NOT NULL | 周期结束日 |
| rankings | JSONB | NOT NULL | Top 20 排名数据 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |

- 新增 `user_leaderboard_preferences` 表（隐私设置）:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| user_id | UUID | PK, FK → users.id | 用户 ID |
| visible_on_leaderboard | BOOLEAN | DEFAULT true | 是否在排行榜展示 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

**API 变更**:
- `GET /api/v1/leaderboard?type=weekly|monthly` — 获取排行榜（Top 20）
- `PUT /api/v1/leaderboard/privacy` — 设置隐私开关
- `GET /api/v1/social-proof/referral` — 获取推荐社交证明
- `GET /api/v1/social-proof/upgrade` — 获取升级社交证明

**测试策略**:
- 单元测试：排行榜聚合计算（Top 20 截取、隐私过滤、昵称脱敏）、社交证明文本生成
- 集成测试：推荐关系写入 → Worker 聚合 → Redis 排行榜查询 → API 返回
- 边界测试：推荐数为 0 时的排行榜展示、隐私切换后排行榜实时更新、同分排名处理

---

#### 3.4.3 UX-17: 多语言 i18n 架构

**需求关联**: PRD 3.4 UX-17
**依赖**: MB-01（设计系统组件库提供统一文本渲染规范）

**技术方案**:

为四端建立统一的国际化架构，Web 使用 vue-i18n，iOS 使用 String Catalog，Android 使用资源限定符，小程序使用自定义 i18n 工具。语言列表由 config-service 动态配置管理。

```
┌────────────────────────────────────────────────────────────────────┐
│  i18n 架构总览                                                      │
│                                                                    │
│  ┌────────────────┐      ┌───────────────────────────────────┐    │
│  │  config-service │ ────→│  支持语言列表                      │    │
│  │  i18n config    │      │  ["zh-CN", "en-US", "ja-JP"]      │    │
│  └────────────────┘      │  默认语言: zh-CN                    │    │
│                          └───────────────────────────────────┘    │
│                                                                    │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────┐ │
│  │ Web (Vue 3) │  │ iOS (Swift)  │  │Android(Kotlin)│  │ 小程序  │ │
│  │ vue-i18n    │  │String Catalog│  │Resource Qual. │  │自定义   │ │
│  │             │  │              │  │               │  │i18n     │ │
│  │ locales/    │  │xcstrings     │  │ res/values-   │  │ i18n/   │ │
│  │ zh-CN.json  │  │ catalog      │  │ en/strings.xml│  │zh-CN.js │ │
│  │ en-US.json  │  │              │  │ res/values-   │  │en-US.js │ │
│  └─────────────┘  └──────────────┘  │ zh/strings.xml│  └────────┘ │
│                                     └──────────────┘              │
│                                                                    │
│  共享翻译源：                                                       │
│  translations/                                                     │
│  ├── zh-CN.json    ← 主翻译文件                                    │
│  ├── en-US.json    ← 英文翻译                                      │
│  └── key_format: "module.section.key"                              │
└────────────────────────────────────────────────────────────────────┘
```

**i18n Key 命名规范**:

```
{module}.{section}.{key}

示例：
common.button.save         → "保存"
common.button.cancel       → "取消"
auth.login.title           → "登录"
auth.login.phone_hint      → "请输入手机号"
credits.balance.title      → "积分余额"
subscription.plan.standard → "标准版"
settings.language.title    → "语言设置"
referral.leaderboard.title → "推荐达人榜"
```

**Web（vue-i18n）实现**:

```typescript
// src/i18n/index.ts
import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: { 'zh-CN': zhCN, 'en-US': enUS }
})
```

**iOS（String Catalog）实现**:

```
ios/AccountCenter/
├── AccountCenter.xcstrings          # String Catalog（Xcode 15+）
├── Views/
│   └── 各 View 使用 Text(String(localized: "auth.login.title"))
└── Utils/
    └── LocaleManager.swift          # 语言切换管理
```

**Android（资源限定符）实现**:

```
android/app/src/main/res/
├── values/strings.xml               # 默认（中文）
├── values-en/strings.xml            # 英文
└── values-ja/strings.xml            # 日文
```

**小程序自定义 i18n**:

```
weapp/
├── i18n/
│   ├── index.ts                     # i18n 初始化 + 切换逻辑
│   ├── zh-CN.ts                     # 中文翻译
│   └── en-US.ts                     # 英文翻译
└── utils/
    └── locale.ts                    # 语言存储/读取
```

**日期/数字本地化**:
- Web: `Intl.DateTimeFormat` + `Intl.NumberFormat`（浏览器原生 API）
- iOS: `DateFormatter` + `NumberFormatter`（自动适配 locale）
- Android: `java.text.DateFormat` + `java.text.NumberFormat`
- 小程序: `wx.getSystemInfoSync().language` 判断后自定义格式化

**关键代码路径**:
- `web-ui/src/i18n/index.ts` — vue-i18n 初始化配置
- `web-ui/src/i18n/locales/zh-CN.json` — Web 中文翻译文件
- `web-ui/src/i18n/locales/en-US.json` — Web 英文翻译文件
- `ios/AccountCenter/AccountCenter.xcstrings` — iOS String Catalog
- `ios/AccountCenter/Utils/LocaleManager.swift` — iOS 语言管理
- `android/app/src/main/res/values/strings.xml` — Android 默认字符串
- `android/app/src/main/res/values-en/strings.xml` — Android 英文字符串
- `weapp/i18n/index.ts` — 小程序 i18n 初始化
- `weapp/i18n/zh-CN.ts` — 小程序中文翻译
- `weapp/i18n/en-US.ts` — 小程序英文翻译
- `config-service/internal/handler/i18n.go` — 语言列表配置 API

**API 变更**:
- `GET /api/v1/config/i18n/languages` — 获取支持语言列表
- `PUT /api/v1/users/me/preferences/language` — 设置用户语言偏好

**数据库变更**:
- `users` 表新增 `locale` 字段（VARCHAR(10), DEFAULT 'zh-CN'）
- Goose migration: `ALTER TABLE users ADD COLUMN locale VARCHAR(10) DEFAULT 'zh-CN'`

**测试策略**:
- 单元测试：i18n key 完整性校验（所有 locale 文件 key 一致）、翻译插值变量校验
- 集成测试：语言切换 API → 用户偏好持久化 → 后续请求返回匹配语言内容
- 跨端一致性测试：zh-CN / en-US 下各端关键页面文本对照验证
- 布局测试：英文长文本不截断、不溢出（RTL 语言预留布局兼容）

---

#### 3.4.4 FN-14: A/B 测试框架

**需求关联**: PRD 3.4 FN-14
**依赖**: FN-12（事件埋点 SDK 提供实验指标数据采集）、FN-13（实时用户行为流提供实时实验监控）

**技术方案**:

在 data-product-service 中扩展 A/B 分组引擎，支持用户 ID 哈希分流和标签维度分层。分组结果缓存至 Redis（TTL 24h），前端通过 SDK API 获取实验变体。实验指标通过 FN-12 埋点 SDK 自动采集，data-product-service 定时计算统计显著性。

```
┌─────────────────────────────────────────────────────────────────────┐
│  A/B 测试框架架构                                                    │
│                                                                     │
│  ┌──────────┐  getExperimentVariant  ┌──────────────────────┐      │
│  │  前端 SDK │ ─────────────────────→ │  data-product-service │      │
│  │          │  (experimentId)        │  A/B Engine            │      │
│  └──────────┘                        │                        │      │
│                                      │  1. 查 Redis 缓存      │      │
│                                      │     ac:ab:{uid}:{eid}  │      │
│                                      │  2. 未命中→计算分组     │      │
│                                      │     hash(uid+salt)%100 │      │
│                                      │     → variant_A/B/C    │      │
│                                      │  3. 写入 Redis(TTL 24h)│      │
│                                      │  4. 返回 variant       │      │
│                                      └──────────────────────┘      │
│                                                │                    │
│                                    ┌───────────┼──────────────┐    │
│                                    ▼           ▼              ▼    │
│                              ┌──────────┐ ┌─────────┐ ┌──────────┐│
│                              │ 事件采集  │ │指标计算  │ │统计检验  ││
│                              │(FN-12)   │ │定时任务  │ │贝叶斯/   ││
│                              │exp_id +  │ │转化率    │ │频率学派  ││
│                              │variant   │ │对比     │ │p-value   ││
│                              └──────────┘ └─────────┘ │CI       ││
│                                           │          └──────────┘│
│                                           ▼                       │
│                                    ┌──────────────┐              │
│                                    │  Admin 后台   │              │
│                                    │  实验管理 CRUD │              │
│                                    │  报告查看      │              │
│                                    │  一键全量推送   │              │
│                                    └──────────────┘              │
└─────────────────────────────────────────────────────────────────────┘
```

**分组算法**:

```go
func AssignVariant(userID string, experiment Experiment) string {
    salt := experiment.Salt
    hash := fnv.New32a()
    hash.Write([]byte(userID + salt))
    bucket := hash.Sum32() % 100
    
    cumulative := uint32(0)
    for _, variant := range experiment.Variants {
        cumulative += variant.TrafficPercent
        if bucket < cumulative {
            return variant.Name
        }
    }
    return experiment.Variants[0].Name
}
```

**实验配置数据结构**（PostgreSQL）:

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 实验 ID |
| name | VARCHAR(100) | 实验名称 |
| description | TEXT | 实验描述 |
| hypothesis | TEXT | 实验假设 |
| status | VARCHAR(20) | draft/running/paused/completed |
| salt | VARCHAR(32) | 分组盐值（随机生成） |
| variants | JSONB | 变体配置 [{"name":"A","traffic_percent":50},{"name":"B","traffic_percent":50}] |
| target_audience | JSONB | 目标人群 {"levels":["L2","L3"],"platforms":["web","ios"]} |
| metrics | TEXT[] | 关注指标 ["conversion_rate","revenue_per_user"] |
| start_at | TIMESTAMP | 开始时间 |
| end_at | TIMESTAMP | 预计结束时间 |
| created_by | UUID | 创建人 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**统计显著性计算**:
- 默认使用贝叶斯方法（Beta-Binomial 共轭先验），计算后验概率 P(B > A) > 95% 时判定显著
- 支持切换为频率学派方法（卡方检验 / t 检验），p-value < 0.05 判定显著
- 样本量计算：基于 MDE（最小检测效应）和基础转化率，自动计算所需最小样本量

**关键代码路径**:
- `data-product-service/internal/service/ab_engine.go` — A/B 分组引擎核心逻辑
- `data-product-service/internal/service/ab_statistics.go` — 统计显著性计算
- `data-product-service/internal/repository/experiment.go` — 实验配置数据层
- `data-product-service/internal/handler/experiment.go` — 实验管理 API handler
- `data-product-service/internal/worker/ab_report.go` — Asynq 定时实验报告生成
- `account-service/internal/handler/admin_experiment.go` — Admin 实验管理页面 API
- `web-ui/src/composables/useExperiment.ts` — Web 端 A/B SDK Composable
- `ios/AccountCenter/Services/ABTestService.swift` — iOS A/B SDK
- `android/app/src/main/java/com/neuro/accountcenter/data/ab/ABTestRepository.kt` — Android A/B SDK

**数据库变更**:
- 新增 `experiments` 表（实验配置）
- 新增 `experiment_assignments` 表（用户分组记录）:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 分组记录 ID |
| experiment_id | UUID | FK → experiments.id | 实验 ID |
| user_id | UUID | NOT NULL | 用户 ID |
| variant | VARCHAR(20) | NOT NULL | 分配的变体 |
| assigned_at | TIMESTAMP | NOT NULL | 分配时间 |

- 索引：`UNIQUE(experiment_id, user_id)` 确保每用户每实验仅分配一次

**API 变更**:
- `GET /api/v1/ab/experiments/{id}/variant` — 获取当前用户的实验变体
- `GET /api/v1/ab/experiments` — 实验列表（Admin）
- `POST /api/v1/ab/experiments` — 创建实验（Admin）
- `PUT /api/v1/ab/experiments/{id}/status` — 启停实验（Admin）
- `GET /api/v1/ab/experiments/{id}/report` — 实验报告（Admin）
- `POST /api/v1/ab/experiments/{id}/rollout` — 一键全量推送胜出方案（Admin）

**测试策略**:
- 单元测试：分组算法（确定性、均匀分布、盐值隔离）、统计计算（贝叶斯/频率学派）
- 集成测试：创建实验 → 用户分流 → 事件采集 → 指标计算 → 报告生成完整链路
- 边界测试：100% 流量单变体、多变体流量分配总和验证、实验启停状态切换

---

#### 3.4.5 FN-16: 企业微信/钉钉集成

**需求关联**: PRD 3.4 FN-16
**依赖**: FN-15（OAuth 社交登录扩展提供 OAuth Provider 框架基础）

**技术方案**:

基于 FN-15 建立的 OAuth Provider 框架，在 auth-service 中扩展企业微信和钉钉 OAuth Provider。通讯录同步采用定时拉取 + Webhook 增量更新策略，审批流集成通过回调 URL 接收审批结果。

```
┌──────────────────────────────────────────────────────────────────────┐
│  企业微信/钉钉集成架构                                                │
│                                                                      │
│  ┌─────────────────┐  扫码登录   ┌─────────────────────────────┐    │
│  │  企业用户         │ ─────────→ │  auth-service                │    │
│  │  (企微/钉钉 App) │            │                              │    │
│  └─────────────────┘            │  OAuth Provider Registry      │    │
│                                 │  ┌────────────────────────┐  │    │
│                                 │  │ WorkWechatProvider     │  │    │
│                                 │  │  - OAuth2 授权码模式    │  │    │
│                                 │  │  - snsapi_privateinfo  │  │    │
│                                 │  └────────────────────────┘  │    │
│                                 │  ┌────────────────────────┐  │    │
│                                 │  │ DingTalkProvider       │  │    │
│                                 │  │  - OAuth2 授权码模式    │  │    │
│                                 │  │  - 企业应用授权         │  │    │
│                                 │  └────────────────────────┘  │    │
│                                 └──────────────┬──────────────┘    │
│                                                │                    │
│                                 ┌──────────────▼──────────────┐    │
│                                 │  account-service             │    │
│                                 │  - 企业账号创建/映射          │    │
│                                 │  - 企业部门结构同步           │    │
│                                 │  - 企业权限管理              │    │
│                                 └──────────────┬──────────────┘    │
│                                                │                    │
│  ┌─────────────────────────────────────────────▼─────────────────┐ │
│  │  审批流集成                                                      │ │
│  │                                                                 │ │
│  │  payment-service ──推送审批──→ 企微/钉钉审批 API                │ │
│  │                         ←──回调──  审批结果回调 URL              │ │
│  │                         →──更新──  订单/订阅状态                 │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

**OAuth Provider 实现**（复用 FN-15 Provider 接口）:

```go
type EnterpriseOAuthProvider interface {
    OAuthProvider  // 继承 FN-15 基础 OAuth 接口
    GetUserInfo(ctx context.Context, accessToken string) (*EnterpriseUserInfo, error)
    GetDepartmentList(ctx context.Context, accessToken string) ([]Department, error)
    GetUserListByDepartment(ctx context.Context, accessToken string, deptID int) ([]EnterpriseUser, error)
    CreateApproval(ctx context.Context, req ApprovalRequest) (string, error)
    GetApprovalStatus(ctx context.Context, approvalID string) (*ApprovalStatus, error)
}

type EnterpriseUserInfo struct {
    UserID      string
    Name        string
    Department  []int
    Mobile      string
    Email       string
    EnterpriseID string
}
```

**通讯录同步策略**:
- 全量同步：每日 03:00 定时拉取企微/钉钉全量部门+员工列表
- 增量同步：注册 Webhook 接收企微/钉钉通讯录变更事件
- 冲突处理：以企微/钉钉数据为准，本地标记 `sync_source = "work_wechat" | "dingtalk"`

**审批流集成**:
- 订阅购买（金额 > 阈值）→ payment-service 推送审批至企微/钉钉
- 退款审批 → payment-service 推送审批至企微/钉钉
- 审批回调 URL: `POST /api/v1/enterprise/approval/callback`
- 回调处理：验签 → 查询审批状态 → 更新订单 → 记录审计日志

**关键代码路径**:
- `auth-service/internal/provider/work_wechat.go` — 企业微信 OAuth Provider
- `auth-service/internal/provider/dingtalk.go` — 钉钉 OAuth Provider
- `account-service/internal/service/enterprise_sync.go` — 企业通讯录同步
- `account-service/internal/worker/enterprise_sync.go` — Asynq 定时全量同步 Worker
- `account-service/internal/handler/enterprise.go` — 企业管理 API
- `account-service/internal/handler/approval_callback.go` — 审批流回调 handler
- `account-service/internal/model/enterprise.go` — 企业/部门/员工数据模型
- `account-service/internal/repository/enterprise.go` — 企业数据层
- `payment-service/internal/service/enterprise_approval.go` — 审批推送逻辑

**数据库变更**:
- 新增 `enterprises` 表:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 企业 ID |
| source | VARCHAR(20) | NOT NULL | 来源：work_wechat/dingtalk |
| external_id | VARCHAR(100) | NOT NULL | 外部企业 ID |
| name | VARCHAR(200) | NOT NULL | 企业名称 |
| corp_id | VARCHAR(100) | | 企业 CorpID |
| access_token | VARCHAR(500) | | 加密的 access_token |
| refresh_token | VARCHAR(500) | | 加密的 refresh_token |
| token_expires_at | TIMESTAMP | | token 过期时间 |
| sync_enabled | BOOLEAN | DEFAULT true | 是否启用通讯录同步 |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

- 新增 `enterprise_departments` 表（部门结构）
- 新增 `enterprise_users` 表（企微/钉钉用户与 Account Center 用户映射）
- 新增 `enterprise_approval_records` 表（审批流记录）

**API 变更**:
- `GET /api/v1/auth/enterprise/work-wechat/authorize` — 企微扫码登录入口
- `GET /api/v1/auth/enterprise/work-wechat/callback` — 企微登录回调
- `GET /api/v1/auth/enterprise/dingtalk/authorize` — 钉钉扫码登录入口
- `GET /api/v1/auth/enterprise/dingtalk/callback` — 钉钉登录回调
- `POST /api/v1/enterprise/sync` — 手动触发通讯录同步（Admin）
- `GET /api/v1/enterprise/departments` — 获取企业部门树
- `GET /api/v1/enterprise/users` — 获取企业员工列表
- `POST /api/v1/enterprise/approval/callback` — 审批结果回调（企微/钉钉调用）
- `PUT /api/v1/enterprise/users/{id}/permissions` — 设置企业员工权限（Admin）

**测试策略**:
- 单元测试：OAuth Provider 接口实现（mock 企微/钉钉 API）、通讯录同步逻辑、审批回调处理
- 集成测试：使用企微/钉钉沙箱环境完成 OAuth 流程 + 通讯录拉取 + 审批创建/回调
- 安全测试：回调验签、token 加密存储、企业数据隔离

---

#### 3.4.6 MB-04: 无障碍 Accessibility

**需求关联**: PRD 3.4 MB-04
**依赖**: MB-01（设计系统组件库提供统一组件规范，确保 A11y 标注一致性）

**技术方案**:

为三端（iOS/Android/Web）添加完整的无障碍支持，包括辅助技术标注（VoiceOver/TalkBack/ARIA）、动态字体适配和颜色对比度验证。将 A11y 规范纳入设计系统组件库，确保新增组件默认满足无障碍标准。

```
┌────────────────────────────────────────────────────────────────────┐
│  无障碍架构                                                        │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  设计系统组件库 (MB-01)                                    │     │
│  │  - 每个组件默认包含 A11y 标注                              │     │
│  │  - A11y 审查清单纳入组件开发流程                           │     │
│  └────────────┬──────────────┬──────────────┬───────────────┘     │
│               │              │              │                     │
│  ┌────────────▼──┐  ┌────────▼──────┐  ┌───▼────────────┐        │
│  │ iOS           │  │ Android       │  │ Web            │        │
│  │ VoiceOver     │  │ TalkBack      │  │ ARIA + Screen  │        │
│  │               │  │               │  │ Readers        │        │
│  │ .accessibility │  │ contentDesc   │  │ aria-label     │        │
│  │  Label         │  │ semantics{}   │  │ role           │        │
│  │ .accessibility │  │               │  │ aria-described │        │
│  │  Hint          │  │               │  │  By            │        │
│  │ .accessibility │  │               │  │ tabindex       │        │
│  │  Traits        │  │               │  │                │        │
│  └───────────────┘  └───────────────┘  └────────────────┘        │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  共通规范                                                 │     │
│  │  - 动态字体大小：iOS Dynamic Type / Android SP / Web rem │     │
│  │  - 颜色对比度 ≥ 4.5:1 (WCAG 2.1 AA)                     │     │
│  │  - 最小触控区域 44×44pt (iOS) / 48×48dp (Android)        │     │
│  │  - 焦点顺序：从左到右、从上到下                           │     │
│  └──────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
```

**iOS VoiceOver 标注规范**:

```swift
// 标准标注模式
Button("登录") {
    // action
}
.accessibilityLabel("登录按钮")
.accessibilityHint("双击以提交登录表单")
.accessibilityAddTraits(.isButton)

// 自定义组件标注
struct CreditBalanceView: View {
    let balance: Int
    var body: some View {
        HStack {
            Text("积分余额")
            Text("\(balance)")
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("积分余额 \(balance) 分")
    }
}
```

**Android TalkBack 标注规范**:

```kotlin
// Compose 语义标注
Button(
    onClick = { /* action */ },
    modifier = Modifier.semantics {
        contentDescription = "登录按钮"
        role = Role.Button
    }
) {
    Text("登录")
}

// 自定义组件组合标注
Row(modifier = Modifier.semantics(mergeDescendants = true) {
    contentDescription = "积分余额 $balance 分"
}) {
    Text("积分余额")
    Text("$balance")
}
```

**Web ARIA 标注规范**:

```vue
<template>
  <button
    class="btn-primary"
    :aria-label="t('auth.login.submit')"
    :aria-describedby="error ? 'login-error' : undefined"
    @click="handleLogin"
  >
    {{ t('auth.login.submit') }}
  </button>
  <span v-if="error" id="login-error" role="alert">
    {{ error }}
  </span>
</template>
```

**动态字体适配**:
- iOS: 使用 `.font(.body)` 等 Semantic Font，自动响应 Dynamic Type
- Android: 使用 `sp` 单位，`MaterialTheme.typography` 中的 scalable typography
- Web: 使用 `rem` 单位 + `clamp()` 函数，尊重用户 `prefers-reduced-motion` 和 `prefers-contrast`

**颜色对比度验证**:
- 设计时使用设计系统 Token 中预验证的调色板（所有前景/背景组合对比度 ≥ 4.5:1）
- CI 集成 axe-core 自动化检测（Web 端）
- iOS/Android 通过 Accessibility Inspector / Accessibility Scanner 手动验证

**关键代码路径**:
- `web-ui/src/components/accessibility/AccessibleButton.vue` — A11y 按钮（纳入设计系统）
- `web-ui/src/components/accessibility/AccessibleInput.vue` — A11y 输入框
- `web-ui/src/composables/useA11y.ts` — A11y 工具（焦点管理、屏幕阅读器通知）
- `ios/AccountCenter/Extensions/View+Accessibility.swift` — iOS A11y 扩展
- `ios/AccountCenter/Views/AccessibleComponents/` — iOS A11y 组件目录
- `android/app/src/main/java/com/neuro/accountcenter/ui/components/AccessibleComponents.kt` — Android A11y 组件
- `scripts/a11y/axe_audit.sh` — Web A11y 自动化审计脚本
- `scripts/a11y/contrast_check.py` — 颜色对比度检查工具

**测试策略**:
- 自动化测试：Web 端使用 axe-core + Playwright 扫描全部页面，CI 阻断 A11y 违规
- iOS 测试：XCTest Accessibility 测试 + VoiceOver 手动验收（注册→登录→订阅购买全流程）
- Android 测试：Compose Testing `assertIsDisplayed()` + `assert(hasContentDescription())` + TalkBack 手动验收
- 视觉测试：动态字体最大/最小时 UI 不截断不重叠
- 对比度测试：所有文本/背景组合通过 WCAG 2.1 AA 级验证（≥4.5:1 普通文本，≥3:1 大文本）

---

#### 3.4.7 AR-04: API v2 版本管理

**需求关联**: PRD 3.4 AR-04
**依赖**: AR-27（API 文档自动生成提供版本化文档输出能力）

**技术方案**:

在 API Gateway 中实现 URL Path + Header 双重路由机制，支持 `/api/v1/*` 和 `/api/v2/*` 并存。定义版本生命周期状态机（Active → Deprecated → Retired），每个版本独立路由表和后端 Handler。

```
┌─────────────────────────────────────────────────────────────────────┐
│  API 版本路由架构                                                    │
│                                                                     │
│  客户端请求                                                         │
│       │                                                             │
│       ▼                                                             │
│  ┌──────────────────────────────────────────┐                      │
│  │  API Gateway (30300)                      │                      │
│  │                                           │                      │
│  │  Route Resolution:                        │                      │
│  │  1. URL Path: /api/v2/users/* → v2 Router │                      │
│  │             /api/v1/users/* → v1 Router   │                      │
│  │  2. Header Fallback:                      │                      │
│  │     Accept-Version: v2 → v2 Router        │                      │
│  │     (default: v1)                         │                      │
│  │                                           │                      │
│  │  Version Middleware:                       │                      │
│  │  - Active: 正常代理                       │                      │
│  │  - Deprecated: 代理 + Warning Header      │                      │
│  │    Deprecation: true                      │                      │
│  │    Sunset: "2027-06-01"                   │                      │
│  │  - Retired: 返回 410 Gone                 │                      │
│  └───────────────┬──────────────────────────┘                      │
│                  │                                                  │
│       ┌──────────┴──────────┐                                      │
│       ▼                     ▼                                      │
│  ┌──────────┐        ┌──────────┐                                  │
│  │ v1 Router│        │ v2 Router│                                  │
│  │ (legacy) │        │ (current)│                                  │
│  │          │        │          │                                  │
│  │/api/v1/  │        │/api/v2/  │                                  │
│  │ users/*  │        │ users/*  │                                  │
│  │ orders/* │        │ orders/* │                                  │
│  │ ...      │        │ ...      │                                  │
│  └────┬─────┘        └────┬─────┘                                  │
│       │                   │                                        │
│       ▼                   ▼                                        │
│  各后端服务 Handler（同一服务可同时处理 v1/v2 请求）                  │
└─────────────────────────────────────────────────────────────────────┘
```

**版本配置数据结构**（config-service 管理）:

```json
{
  "api_versions": {
    "v1": {
      "status": "active",
      "base_path": "/api/v1",
      "deprecated_at": null,
      "sunset_at": null,
      "description": "V1 API - 向后兼容"
    },
    "v2": {
      "status": "active",
      "base_path": "/api/v2",
      "deprecated_at": null,
      "sunset_at": null,
      "description": "V2 API - 优化响应格式，新增字段"
    }
  }
}
```

**版本生命周期状态机**:

```
  ┌─────────┐    标记弃用     ┌────────────┐    到达 Sunset 日期   ┌─────────┐
  │ Active  │ ─────────────→ │ Deprecated │ ───────────────────→ │ Retired │
  │         │                │            │                      │         │
  │ 正常服务 │                │ Warning    │                      │ 410 Gone│
  │         │                │ Header     │                      │         │
  └─────────┘                └────────────┘                      └─────────┘
```

**Gateway 版本路由中间件**:

```go
func VersionMiddleware(versions map[string]VersionConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        version := extractVersion(c)
        config, exists := versions[version]
        if !exists {
            c.JSON(404, gin.H{"error": "unsupported_api_version"})
            c.Abort()
            return
        }
        switch config.Status {
        case "active":
            c.Next()
        case "deprecated":
            c.Header("Deprecation", "true")
            c.Header("Sunset", config.SunsetAt.Format(time.RFC1123))
            c.Header("Link", fmt.Sprintf("</api/v2%s>; rel=\"successor-version\"", c.Request.URL.Path))
            c.Next()
        case "retired":
            c.JSON(410, gin.H{
                "error":   "api_version_retired",
                "message": fmt.Sprintf("API %s has been retired. Please migrate to %s.", version, config.SuccessorVersion),
            })
            c.Abort()
        }
    }
}
```

**v2 API Breaking Change 示例**（参考）:
- 响应格式统一：`{ "data": {...}, "meta": {...} }` 结构化封装
- 分页参数：`page/page_size` → `offset/limit`
- 时间格式：Unix timestamp → ISO 8601 (RFC 3339)
- 错误格式：统一 `{ "error": { "code": "xxx", "message": "xxx" } }`

**关键代码路径**:
- `api-gateway/internal/middleware/version.go` — 版本路由中间件
- `api-gateway/internal/proxy/v1_router.go` — v1 路由表
- `api-gateway/internal/proxy/v2_router.go` — v2 路由表
- `api-gateway/cmd/main.go` — 版本路由注册入口
- `config-service/internal/handler/api_version.go` — 版本配置管理 API
- `docs/api/v1/` — v1 API 文档（OpenAPI 3.0，由 AR-27 自动生成）
- `docs/api/v2/` — v2 API 文档（独立输出）

**数据库变更**:
- 新增 `api_version_configs` 表（或通过 config-service 管理，不新增表，使用现有配置机制）

**API 变更**:
- `GET /api/v2/*` — 全部 v2 端点（与 v1 对应，响应格式升级）
- `GET /api/v1/admin/api-versions` — 版本生命周期管理（Admin）
- `PUT /api/v1/admin/api-versions/{version}/status` — 更新版本状态（Admin）

**测试策略**:
- 单元测试：版本提取逻辑（URL Path + Header）、版本状态中间件（Active/Deprecated/Retired 返回）
- 集成测试：v1 和 v2 并发请求 → 路由正确 → 响应格式符合各自规范
- 回归测试：v1 客户端在 v2 上线后功能不受影响
- Header 测试：Deprecated 版本返回正确的 Warning/Sunset/Link Header

---

#### 3.4.8 AR-12: 读写分离

**需求关联**: PRD 3.4 AR-12
**依赖**: AR-21（K8s Helm Chart 提供数据库只读副本的部署编排基础）

**技术方案**:

为 data-product-service 配置双 PostgreSQL 数据源连接：主库（Primary，读写）和只读副本（Replica，仅读）。报表查询和聚合分析自动路由到只读副本，写操作和实时查询走主库。基于 PG Streaming Replication 实现数据同步，通过健康检查监控副本延迟。

```
┌──────────────────────────────────────────────────────────────────────┐
│  读写分离架构                                                         │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐     │
│  │  data-product-service (30314)                               │     │
│  │                                                             │     │
│  │  ┌──────────────────┐      ┌──────────────────────┐       │     │
│  │  │  Primary DSN     │      │  Replica DSN          │       │     │
│  │  │  (读写)           │      │  (只读)               │       │     │
│  │  │  pg-primary:5432 │      │  pg-replica:5432      │       │     │
│  │  └────────┬─────────┘      └───────────┬──────────┘       │     │
│  │           │                            │                    │     │
│  │  ┌────────▼────────────────────────────▼──────────┐       │     │
│  │  │  DualDataSource Router                         │       │     │
│  │  │                                                │       │     │
│  │  │  写操作 (INSERT/UPDATE/DELETE)  → Primary       │       │     │
│  │  │  报表查询 (SELECT 聚合/统计)    → Replica        │       │     │
│  │  │  实时查询 (SELECT 单行/byID)    → Primary       │       │     │
│  │  │                                                │       │     │
│  │  │  Replica 不可用 → 降级至 Primary                │       │     │
│  │  └────────────────────────────────────────────────┘       │     │
│  └────────────────────────────────────────────────────────────┘     │
│                         │                      │                    │
│  ┌──────────────────────▼──┐    ┌──────────────▼─────────────┐    │
│  │  PostgreSQL Primary     │    │  PostgreSQL Replica         │    │
│  │  (读写主库)              │    │  (只读副本)                 │    │
│  │  port: 5432             │    │  port: 5433                 │    │
│  │                         │    │                              │    │
│  │  WAL Sender             │───→│  WAL Receiver + Replay      │    │
│  │  streaming replication  │    │  hot_standby = on           │    │
│  └─────────────────────────┘    └────────────────────────────┘    │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐     │
│  │  健康检查 + 延迟监控                                         │     │
│  │  - 查询 pg_stat_wal_receiver → replica 延迟（字节差）        │     │
│  │  - 延迟 > 1s → 告警 + 查询结果附加时效性提示                │     │
│  │  - Replica 连接失败 → 自动降级至 Primary + P1 告警          │     │
│  └────────────────────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────────────────────┘
```

**双数据源配置**:

```yaml
# values.yaml (Helm)
data-product-service:
  database:
    primary:
      host: pg-primary
      port: 5432
      database: account_center
      max_open_conns: 25
      max_idle_conns: 10
    replica:
      host: pg-replica
      port: 5433
      database: account_center
      max_open_conns: 20
      max_idle_conns: 8
```

**DualDataSource Router 实现**:

```go
type DualDataSource struct {
    primary  *sql.DB
    replica  *sql.DB
    health   *ReplicaHealthChecker
}

func (ds *DualDataSource) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
    if ds.isReadOnlyQuery(query) && ds.health.IsReplicaHealthy() {
        return ds.replica.QueryRowContext(ctx, query, args...)
    }
    return ds.primary.QueryRowContext(ctx, query, args...)
}

func (ds *DualDataSource) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    if ds.isReadOnlyQuery(query) && ds.health.IsReplicaHealthy() {
        rows, err := ds.replica.QueryContext(ctx, query, args...)
        if err == nil {
            return rows, nil
        }
        log.Warn("replica query failed, falling back to primary", "error", err)
    }
    return ds.primary.QueryContext(ctx, query, args...)
}
```

**查询路由规则**:

| 操作类型 | 路由目标 | 说明 |
|---------|---------|------|
| INSERT/UPDATE/DELETE | Primary | 所有写操作走主库 |
| SELECT ... GROUP BY / SUM / COUNT | Replica | 聚合报表查询走副本 |
| SELECT ... FROM rfm_* / funnel_* | Replica | 分析类查询走副本 |
| SELECT ... WHERE id = ? / LIMIT 1 | Primary | 单行实时查询走主库 |
| SELECT ... FOR UPDATE | Primary | 悲观锁查询走主库 |

**PostgreSQL Streaming Replication 配置**:

```ini
# Primary (postgresql.conf)
wal_level = replica
max_wal_senders = 3
wal_keep_size = 256MB

# Replica (postgresql.conf)
hot_standby = on
primary_conninfo = 'host=pg-primary port=5432 user=replication password=xxx'
```

**Helm Chart 副本编排**（AR-21 扩展）:

```yaml
# helm/account-center/templates/_infra/postgresql-replica.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: pg-replica
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: postgres
        env:
        - name: POSTGRES_MODE
          value: "replica"
        - name: POSTGRES_PRIMARY_HOST
          value: "pg-primary"
```

**关键代码路径**:
- `data-product-service/internal/repository/datasource.go` — DualDataSource 实现
- `data-product-service/internal/repository/router.go` — 查询路由逻辑
- `data-product-service/internal/repository/health.go` — Replica 健康检查
- `data-product-service/internal/svcconfig/config.go` — 双数据源配置加载
- `data-product-service/internal/handler/health.go` — 健康检查端点（含副本延迟）
- `helm/account-center/templates/_infra/postgresql-replica.yaml` — PG Replica StatefulSet
- `monitoring/grafana/dashboards/replication.json` — 复制延迟 Dashboard

**数据库变更**:
- 无 schema 变更，Replica 是 Primary 的物理复制
- 新增 Prometheus 自定义指标：
  - `pg_replication_lag_seconds` — 复制延迟（秒）
  - `datasource_replica_healthy` — Replica 健康状态（0/1）
  - `datasource_fallback_total` — 降级至主库次数

**API 变更**:
- `GET /api/v1/health` 扩展响应，包含副本状态：

```json
{
  "status": "healthy",
  "primary_db": "healthy",
  "replica_db": "healthy",
  "replication_lag_ms": 120,
  "data_freshness_hint": "数据可能存在 <1s 延迟"
}
```

- `GET /api/v1/admin/data-product/datasource-status` — 数据源状态（Admin）

**测试策略**:
- 单元测试：DualDataSource 路由逻辑（写→Primary、聚合→Replica、降级回退）
- 集成测试：启动 Primary + Replica → 写入数据 → 验证 Replica 可读 → 验证同步延迟
- 故障测试：停止 Replica → 验证查询自动降级至 Primary → 告警触发 → 重启 Replica → 恢复
- 性能测试：对比读写分离前后报表查询 P95 延迟，确认提升 ≥50%

---

### 3.4 P3 技术设计小结

| 需求 ID | 技术方案核心 | 涉及服务 | 新增文件估计 |
|---------|------------|---------|-------------|
| UX-07 | Redis 搜索索引 + 快捷操作注册表 + 拼音匹配 | account-service, api-gateway, 全端前端 | ~15 文件 |
| UX-14 | Redis Sorted Set 排行榜 + 社交证明聚合 + 隐私控制 | account-service, 全端前端 | ~12 文件 |
| UX-17 | vue-i18n + String Catalog + 资源限定符 + 自定义 i18n | config-service, 全端前端 | ~20 文件 |
| FN-14 | A/B 分组引擎 + Redis 缓存 + 贝叶斯统计 + Admin 管理 | data-product-service, account-service, 全端前端 | ~18 文件 |
| FN-16 | OAuth Provider 扩展 + 通讯录同步 + 审批流回调 | auth-service, account-service, payment-service | ~16 文件 |
| MB-04 | VoiceOver/TalkBack/ARIA 标注 + 动态字体 + 对比度验证 | 全端前端 | ~12 文件 |
| AR-04 | Gateway v1/v2 双路由 + 版本生命周期中间件 + Header 路由 | api-gateway, config-service | ~10 文件 |
| AR-12 | DualDataSource 路由 + PG Streaming Replication + 自动降级 | data-product-service, 基础设施 | ~8 文件 |


## 4. 数据设计

### 4.1 ER 关系图

以下 ASCII ER 图展示 Account Center V2.0 全部实体关系。**[NEW]** 标记 V2.0 新增实体。

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Account Center V2.0 ER 关系图                         │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────┐ 1    N ┌──────────────────┐ 1    1 ┌──────────────────┐      │
│  │    users     │───────→│  subscriptions   │───────→│  entitlements    │      │
│  │ (account-svc)│        │  (account-svc)   │        │  (account-svc)   │      │
│  └──────┬───────┘        └──────┬───────────┘        └──────────────────┘      │
│         │                       │                                              │
│         │ 1                     │ 1                                            │
│         │                       │                                              │
│         │ N                     │ N                                            │
│  ┌──────▼──────────┐   ┌───────▼───────────┐                                  │
│  │  orders **[NEW]**│   │ renewal_reminder  │                                  │
│  │ (payment-svc)   │   │ _logs **[NEW]**   │                                  │
│  └──────┬──────────┘   │ (account-svc)     │                                  │
│         │              └───────────────────┘                                  │
│         │ 1                                                                    │
│         │                                                                      │
│         │ N                                                                    │
│  ┌──────▼───────────┐  1   N ┌───────────────────┐                            │
│  │ payment_records  │←───────│  refunds **[NEW]**│                            │
│  │  **[NEW]**       │        │ (payment-svc)     │                            │
│  │ (payment-svc)    │        └───────────────────┘                            │
│  └──────────────────┘                                                          │
│         │                                                                      │
│         │ (order_id)                                                           │
│  ┌──────▼───────────┐  1    1 ┌───────────────────┐                           │
│  │ invoices **[NEW]**│←───────│ user_invoice_info │                           │
│  │ (payment-svc)    │        │  **[NEW]**         │                           │
│  └──────────────────┘        │ (payment-svc)     │                           │
│                              └───────────────────┘                           │
│                                                                                │
│  ┌──────────────┐ N    1 ┌──────────────┐                                     │
│  │    users     │───────→│credit_accounts│                                     │
│  │              │        │ (credit-svc)  │                                     │
│  │              │        └──────┬───────┘                                     │
│  │              │               │ 1                                            │
│  │              │               │                                              │
│  │              │               │ N                                            │
│  │              │       ┌───────▼───────────┐                                  │
│  │              │       │credit_transactions│                                  │
│  │              │       │ (credit-svc)      │                                  │
│  │              │       └───────────────────┘                                  │
│  └──────┬───────┘                                                               │
│         │                                                                       │
│         ├────────────── 1:N ──→ referral_relations (credit-svc)                │
│         │                                                                       │
│         ├────────────── 1:N ──→ device_fingerprints (auth-svc)                 │
│         │                                                                       │
│         ├────────────── 1:N ──→ social_accounts **[NEW]** (auth-svc)           │
│         │                                                                       │
│         ├────────────── 1:N ──→ biometric_device_tokens **[NEW]** (auth-svc)   │
│         │                                                                       │
│         ├────────────── 1:N ──→ guest_sessions **[NEW]** (auth-svc)            │
│         │                                                                       │
│         ├────────────── 1:N ──→ push_tokens **[NEW]** (notification-svc)       │
│         │                                                                       │
│         ├────────────── 1:N ──→ notifications **[NEW]** (notification-svc)     │
│         │                                                                       │
│         ├────────────── 1:N ──→ coupon_usages **[NEW]** (account-svc)          │
│         │                                                                       │
│         ├────────────── 1:1 ──→ user_push_preferences **[NEW]** (notification) │
│         │                                                                       │
│         ├────────────── 1:1 ──→ user_leaderboard_preferences **[NEW]** (acct)  │
│         │                                                                       │
│         └────────────── 1:N ──→ experiment_assignments **[NEW]** (data-product)│
│                                                                                   │
│  ┌──────────────────────┐                                                        │
│  │ admin_users **[NEW]**│ 1:N → admin_audit_logs **[NEW]** (account-svc)        │
│  │ (account-svc)        │ 1:N → coupons **[NEW]** (account-svc)                │
│  └──────────────────────┘ 1:N → promotions **[NEW]** (account-svc)             │
│                                                                                   │
│  ┌──────────────────────┐                                                         │
│  │ blacklist_entries    │  compliance-service (V1.3.1 已有, V2.0 扩展)           │
│  │ risk_events          │  compliance-service (V1.3.1 已有, V2.0 扩展)           │
│  └──────────────────────┘                                                         │
│                                                                                   │
│  ┌──────────────────────┐                                                         │
│  │ events **[NEW]**     │  data-product-service (按月分区)                       │
│  │ experiments **[NEW]**│  data-product-service                                  │
│  └──────────────────────┘                                                         │
│                                                                                   │
│  ┌──────────────────────┐                                                         │
│  │ notification_        │  notification-svc                                      │
│  │ templates **[NEW]**  │  1:N → notification_send_records **[NEW]**            │
│  ├──────────────────────┤                                                         │
│  │ push_strategies      │  notification-svc                                      │
│  │  **[NEW]**           │                                                        │
│  ├──────────────────────┤                                                         │
│  │ push_logs **[NEW]**  │  notification-svc                                      │
│  └──────────────────────┘                                                         │
│                                                                                   │
│  ┌──────────────────────┐                                                         │
│  │ faq_items **[NEW]**  │  config-service                                        │
│  └──────────────────────┘                                                         │
│                                                                                   │
│  ┌──────────────────────┐                                                         │
│  │ enterprises **[NEW]**│  account-svc                                           │
│  │  1:N → enterprise_   │                                                        │
│  │    departments       │  1:N → enterprise_users                                │
│  └──────────────────────┘                                                         │
│                                                                                   │
│  ┌──────────────────────┐                                                         │
│  │ leaderboard_         │  account-svc                                           │
│  │ snapshots **[NEW]**  │                                                        │
│  └──────────────────────┘                                                         │
│                                                                                   │
│  V1.3.1 已有实体: users, subscriptions, entitlements, credit_accounts,            │
│  credit_transactions, referral_relations, device_fingerprints, risk_events,       │
│  blacklist_entries, enterprises(KYB), config_groups, config_items,                │
│  config_versions, config_releases, roles, role_permissions, user_roles            │
│                                                                                   │
│  V2.0 新增实体: orders, payment_records, refunds, admin_users,                    │
│  admin_audit_logs, push_tokens, push_logs, social_accounts,                       │
│  biometric_device_tokens, coupons, coupon_usages, promotions, events,             │
│  notifications, notification_templates, notification_send_records,                 │
│  faq_items, experiments, experiment_assignments, enterprises(account),             │
│  enterprise_departments, enterprise_users, renewal_reminder_logs,                  │
│  invoices, user_invoice_info, push_strategies, user_push_preferences,              │
│  user_leaderboard_preferences, leaderboard_snapshots, guest_sessions               │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 新增表结构

> 以下为 V2.0 全部新增表的完整 DDL 定义。命名规范：`ac:{service}:{entity}` 数据库为 `{service}` 所属 schema。所有时间字段使用 `TIMESTAMPTZ`，主键采用 `UUID` 或 `BIGSERIAL` 按业务体量选择。

---

#### 4.2.1 payment-service 新增表

##### orders

```sql
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    order_no        VARCHAR(32) NOT NULL,
    product_type    VARCHAR(20) NOT NULL,
    product_id      VARCHAR(50) NOT NULL,
    product_name    VARCHAR(100) NOT NULL,
    amount_cents    INTEGER NOT NULL CHECK (amount_cents >= 0),
    currency        VARCHAR(3) NOT NULL DEFAULT 'CNY',
    credits_used    INTEGER NOT NULL DEFAULT 0,
    credits_discount_cents INTEGER NOT NULL DEFAULT 0,
    actual_amount_cents    INTEGER NOT NULL CHECK (actual_amount_cents >= 0),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','paid','cancelled','refunded')),
    payment_method  VARCHAR(20),
    payment_channel VARCHAR(20),
    paid_at         TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,
    refunded_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_expires_pending ON orders(expires_at) WHERE status = 'pending';
```

| 字段 | 说明 |
|------|------|
| order_no | 业务订单号，格式 `ORD{YYYYMMDD}{6位随机}` |
| product_type | `subscription` / `credit_pack` |
| payment_method | `wechat` / `alipay` |
| payment_channel | `h5` / `mini` / `native` / `app` |
| status 状态机 | pending → paid / cancelled; paid → refunded |
| expires_at | 待付订单 30 分钟过期 |

##### payment_records

```sql
CREATE TABLE payment_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders(id),
    user_id         UUID NOT NULL,
    provider        VARCHAR(20) NOT NULL,
    transaction_id  VARCHAR(100),
    amount_cents    INTEGER NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'CNY',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','success','failed')),
    callback_raw    JSONB,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_records_order_id ON payment_records(order_id);
CREATE INDEX idx_payment_records_user_id ON payment_records(user_id);
CREATE INDEX idx_payment_records_status ON payment_records(status);
CREATE INDEX idx_payment_records_transaction_id ON payment_records(transaction_id);
```

##### refunds

```sql
CREATE TABLE refunds (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID NOT NULL REFERENCES orders(id),
    user_id             UUID NOT NULL,
    refund_amount_cents INTEGER NOT NULL CHECK (refund_amount_cents >= 0),
    credits_deducted    INTEGER NOT NULL DEFAULT 0,
    reason              VARCHAR(200) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending_auto'
                        CHECK (status IN ('pending_auto','pending_manual',
                                          'approved','rejected','refunded','failed')),
    reviewed_by         UUID,
    reviewed_at         TIMESTAMPTZ,
    refunded_at         TIMESTAMPTZ,
    refund_method       VARCHAR(20),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refunds_order_id ON refunds(order_id);
CREATE INDEX idx_refunds_user_id ON refunds(user_id);
CREATE INDEX idx_refunds_status ON refunds(status);
```

##### invoices

```sql
CREATE TABLE invoices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders(id),
    user_id         UUID NOT NULL,
    invoice_type    VARCHAR(20) NOT NULL CHECK (invoice_type IN ('personal','enterprise_vat')),
    title           VARCHAR(200) NOT NULL,
    tax_number      VARCHAR(50),
    amount_cents    INTEGER NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','issued','failed')),
    pdf_url         VARCHAR(500),
    invoice_number  VARCHAR(50),
    invoice_date    TIMESTAMPTZ,
    retry_count     INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_order_id ON invoices(order_id);
CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_status ON invoices(status);
```

##### user_invoice_info

```sql
CREATE TABLE user_invoice_info (
    user_id             UUID PRIMARY KEY,
    default_title       VARCHAR(200),
    default_tax_number  VARCHAR(50),
    default_email       VARCHAR(255),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

#### 4.2.2 account-service 新增表

##### admin_users

```sql
CREATE TABLE admin_users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    email       VARCHAR(255),
    role        VARCHAR(20) NOT NULL DEFAULT 'operator'
                CHECK (role IN ('super_admin','admin','operator','finance')),
    status      VARCHAR(10) NOT NULL DEFAULT 'active'
                CHECK (status IN ('active','disabled')),
    last_login_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_admin_users_username ON admin_users(username);
CREATE INDEX idx_admin_users_role ON admin_users(role);
```

| 字段 | 说明 |
|------|------|
| password_hash | argon2id 哈希（与用户密码同算法） |
| role | super_admin（全权限）、admin（管理）、operator（运营）、finance（财务） |

##### admin_audit_logs

```sql
CREATE TABLE admin_audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    admin_id    UUID NOT NULL,
    admin_name  VARCHAR(100) NOT NULL,
    action      VARCHAR(50) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id   VARCHAR(100),
    detail      JSONB,
    ip_address  VARCHAR(45),
    user_agent  VARCHAR(500),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_logs_admin_id ON admin_audit_logs(admin_id);
CREATE INDEX idx_admin_audit_logs_action ON admin_audit_logs(action);
CREATE INDEX idx_admin_audit_logs_created_at ON admin_audit_logs(created_at);
CREATE INDEX idx_admin_audit_logs_target ON admin_audit_logs(target_type, target_id);
```

##### coupons

```sql
CREATE TABLE coupons (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(32) NOT NULL,
    type            VARCHAR(20) NOT NULL CHECK (type IN ('percentage','fixed','free_first_month')),
    value           INTEGER NOT NULL,
    applicable_plans TEXT[],
    max_uses        INTEGER,
    max_uses_per_user INTEGER NOT NULL DEFAULT 1,
    used_count      INTEGER NOT NULL DEFAULT 0,
    valid_from      TIMESTAMPTZ NOT NULL,
    valid_until     TIMESTAMPTZ NOT NULL,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_coupons_valid_period ON coupons(valid_from, valid_until);
```

##### coupon_usages

```sql
CREATE TABLE coupon_usages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_id   UUID NOT NULL REFERENCES coupons(id),
    user_id     UUID NOT NULL,
    order_id    UUID NOT NULL,
    used_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coupon_usages_coupon_id ON coupon_usages(coupon_id);
CREATE INDEX idx_coupon_usages_user_id ON coupon_usages(user_id);
CREATE UNIQUE INDEX idx_coupon_usages_unique ON coupon_usages(coupon_id, user_id, order_id);
```

##### promotions

```sql
CREATE TABLE promotions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    type        VARCHAR(30) NOT NULL,
    config      JSONB NOT NULL DEFAULT '{}',
    status      VARCHAR(10) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft','active','paused','expired')),
    valid_from  TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_promotions_status ON promotions(status);
CREATE INDEX idx_promotions_valid_period ON promotions(valid_from, valid_until);
```

##### enterprises（account-service — 企业微信/钉钉集成）

```sql
CREATE TABLE account_enterprises (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source          VARCHAR(20) NOT NULL CHECK (source IN ('work_wechat','dingtalk')),
    external_id     VARCHAR(100) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    corp_id         VARCHAR(100),
    access_token    VARCHAR(500),
    refresh_token   VARCHAR(500),
    token_expires_at TIMESTAMPTZ,
    sync_enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_acct_enterprises_source_ext ON account_enterprises(source, external_id);
```

##### enterprise_departments

```sql
CREATE TABLE enterprise_departments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id   UUID NOT NULL REFERENCES account_enterprises(id),
    external_dept_id VARCHAR(100) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    parent_id       UUID REFERENCES enterprise_departments(id),
    sync_source     VARCHAR(20) NOT NULL DEFAULT 'work_wechat',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ent_depts_enterprise ON enterprise_departments(enterprise_id);
CREATE UNIQUE INDEX idx_ent_depts_ext ON enterprise_departments(enterprise_id, external_dept_id);
```

##### enterprise_users

```sql
CREATE TABLE enterprise_users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id   UUID NOT NULL REFERENCES account_enterprises(id),
    department_id   UUID REFERENCES enterprise_departments(id),
    user_id         UUID,
    external_user_id VARCHAR(100) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    mobile          VARCHAR(20),
    email           VARCHAR(255),
    sync_source     VARCHAR(20) NOT NULL DEFAULT 'work_wechat',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ent_users_enterprise ON enterprise_users(enterprise_id);
CREATE UNIQUE INDEX idx_ent_users_ext ON enterprise_users(enterprise_id, external_user_id);
CREATE INDEX idx_ent_users_user_id ON enterprise_users(user_id);
```

##### renewal_reminder_logs

```sql
CREATE TABLE renewal_reminder_logs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL,
    subscription_id UUID NOT NULL,
    reminder_type   VARCHAR(5) NOT NULL CHECK (reminder_type IN ('T-7','T-3','T-1')),
    channel         VARCHAR(10) NOT NULL CHECK (channel IN ('push','sms','email')),
    sent_at         TIMESTAMPTZ NOT NULL,
    status          VARCHAR(10) NOT NULL CHECK (status IN ('sent','failed'))
);

CREATE INDEX idx_renewal_reminder_user ON renewal_reminder_logs(user_id);
CREATE INDEX idx_renewal_reminder_dedup ON renewal_reminder_logs(user_id, reminder_type, channel, sent_at);
```

##### leaderboard_snapshots

```sql
CREATE TABLE leaderboard_snapshots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_type VARCHAR(10) NOT NULL CHECK (period_type IN ('weekly','monthly')),
    period_start DATE NOT NULL,
    period_end   DATE NOT NULL,
    rankings    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leaderboard_snap_period ON leaderboard_snapshots(period_type, period_start);
```

##### user_leaderboard_preferences

```sql
CREATE TABLE user_leaderboard_preferences (
    user_id                 UUID PRIMARY KEY,
    visible_on_leaderboard  BOOLEAN NOT NULL DEFAULT true,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

#### 4.2.3 notification-service 新增表

##### push_tokens

```sql
CREATE TABLE push_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL,
    device_type VARCHAR(10) NOT NULL CHECK (device_type IN ('ios','android')),
    token       VARCHAR(500) NOT NULL,
    bundle_id   VARCHAR(100),
    app_version VARCHAR(20),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_push_tokens_token ON push_tokens(token);
CREATE INDEX idx_push_tokens_user ON push_tokens(user_id, device_type);
```

##### push_logs

```sql
CREATE TABLE push_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL,
    device_type VARCHAR(10) NOT NULL,
    token       VARCHAR(500) NOT NULL,
    title       VARCHAR(200),
    body        TEXT,
    payload     JSONB,
    provider    VARCHAR(20) NOT NULL,
    provider_id VARCHAR(100),
    status      VARCHAR(10) NOT NULL CHECK (status IN ('sent','failed')),
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_logs_user ON push_logs(user_id);
CREATE INDEX idx_push_logs_status ON push_logs(status);
CREATE INDEX idx_push_logs_created ON push_logs(created_at);
```

##### notifications

```sql
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    type        VARCHAR(30) NOT NULL,
    category    VARCHAR(10) NOT NULL CHECK (category IN ('system','promo')),
    title       VARCHAR(200) NOT NULL,
    body        TEXT,
    link        VARCHAR(500),
    is_read     BOOLEAN NOT NULL DEFAULT false,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at  TIMESTAMPTZ
);

CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);
```

##### notification_templates

```sql
CREATE TABLE notification_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50) NOT NULL,
    channel     VARCHAR(10) NOT NULL CHECK (channel IN ('sms','email','push')),
    subject     TEXT,
    body        TEXT NOT NULL,
    variables   TEXT[],
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ntemplates_code ON notification_templates(code);
CREATE INDEX idx_ntemplates_channel ON notification_templates(channel);
```

##### notification_send_records

```sql
CREATE TABLE notification_send_records (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id         UUID REFERENCES notification_templates(id),
    user_id             UUID NOT NULL,
    channel             VARCHAR(10) NOT NULL,
    status              VARCHAR(10) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','sent','failed')),
    rendered_content    TEXT,
    error_message       TEXT,
    scheduled_at        TIMESTAMPTZ,
    sent_at             TIMESTAMPTZ,
    retry_count         INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nsend_records_user ON notification_send_records(user_id);
CREATE INDEX idx_nsend_records_status ON notification_send_records(status);
CREATE INDEX idx_nsend_records_scheduled ON notification_send_records(scheduled_at);
```

##### push_strategies

```sql
CREATE TABLE push_strategies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    target_tags     JSONB DEFAULT '{}',
    schedule_at     TIMESTAMPTZ,
    frequency_limit INTEGER NOT NULL DEFAULT 3,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    template_id     UUID REFERENCES notification_templates(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_strategies_active ON push_strategies(is_active);
```

##### user_push_preferences

```sql
CREATE TABLE user_push_preferences (
    user_id             UUID PRIMARY KEY,
    dnd_start           TIME,
    dnd_end             TIME,
    dnd_enabled         BOOLEAN NOT NULL DEFAULT false,
    quiet_push_enabled  BOOLEAN NOT NULL DEFAULT true,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

#### 4.2.4 auth-service 新增表

##### social_accounts

```sql
CREATE TABLE social_accounts (
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 UUID NOT NULL,
    provider                VARCHAR(20) NOT NULL CHECK (provider IN ('wechat','apple','google','alipay','work_wechat','dingtalk')),
    provider_user_id        VARCHAR(128) NOT NULL,
    union_id                VARCHAR(128),
    access_token_encrypted  TEXT,
    refresh_token_encrypted TEXT,
    expires_at              TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_social_accounts_provider ON social_accounts(provider, provider_user_id);
CREATE INDEX idx_social_accounts_user ON social_accounts(user_id);
CREATE INDEX idx_social_accounts_union ON social_accounts(union_id) WHERE union_id IS NOT NULL;
```

##### biometric_device_tokens

```sql
CREATE TABLE biometric_device_tokens (
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 UUID NOT NULL,
    device_fingerprint_id   VARCHAR(128) NOT NULL,
    token_hash              VARCHAR(64) NOT NULL,
    expires_at              TIMESTAMPTZ NOT NULL,
    last_used_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_biometric_user_device ON biometric_device_tokens(user_id, device_fingerprint_id);
CREATE INDEX idx_biometric_expires ON biometric_device_tokens(expires_at);
```

##### guest_sessions

```sql
CREATE TABLE guest_sessions (
    guest_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fingerprint VARCHAR(128),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guest_sessions_created ON guest_sessions(created_at);
```

---

#### 4.2.5 compliance-service 表扩展

> blacklist_entries 和 risk_events 在 V1.3.1 已存在（见 SSD V1.3.1 §5.1）。V2.0 新增以下字段和索引。

##### blacklist_entries V2.0 扩展

```sql
ALTER TABLE blacklist_entries ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
ALTER TABLE blacklist_entries ADD COLUMN IF NOT EXISTS created_by VARCHAR(100) NOT NULL DEFAULT 'system';

CREATE INDEX IF NOT EXISTS idx_blacklist_expires ON blacklist_entries(expires_at) WHERE expires_at IS NOT NULL;
```

##### risk_events V2.0 扩展

```sql
ALTER TABLE risk_events ADD COLUMN IF NOT EXISTS resolved BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE risk_events ADD COLUMN IF NOT EXISTS resolved_by VARCHAR(100);
ALTER TABLE risk_events ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
ALTER TABLE risk_events ADD COLUMN IF NOT EXISTS source_ip VARCHAR(50);
ALTER TABLE risk_events ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(128);
ALTER TABLE risk_events ADD COLUMN IF NOT EXISTS details JSONB;

CREATE INDEX IF NOT EXISTS idx_risk_events_resolved ON risk_events(resolved) WHERE resolved = false;
CREATE INDEX IF NOT EXISTS idx_risk_events_created ON risk_events(created_at);
```

---

#### 4.2.6 data-product-service 新增表

##### events（按月分区）

```sql
CREATE TABLE events (
    id          BIGSERIAL,
    event_name  VARCHAR(50) NOT NULL,
    user_id     UUID,
    session_id  VARCHAR(100),
    properties  JSONB DEFAULT '{}',
    platform    VARCHAR(10),
    app_version VARCHAR(20),
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_events_user_ts ON events(user_id, timestamp);
CREATE INDEX idx_events_name_ts ON events(event_name, timestamp);
```

**分区管理**（每月自动创建下月分区）:

```sql
CREATE TABLE events_2026_05 PARTITION OF events
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE events_2026_06 PARTITION OF events
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
```

##### experiments

```sql
CREATE TABLE experiments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    hypothesis      TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','running','paused','completed')),
    salt            VARCHAR(32) NOT NULL,
    variants        JSONB NOT NULL,
    target_audience JSONB DEFAULT '{}',
    metrics         TEXT[],
    start_at        TIMESTAMPTZ,
    end_at          TIMESTAMPTZ,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experiments_status ON experiments(status);
```

##### experiment_assignments

```sql
CREATE TABLE experiment_assignments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id   UUID NOT NULL REFERENCES experiments(id),
    user_id         UUID NOT NULL,
    variant         VARCHAR(20) NOT NULL,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_exp_assign_unique ON experiment_assignments(experiment_id, user_id);
CREATE INDEX idx_exp_assign_user ON experiment_assignments(user_id);
```

---

#### 4.2.7 config-service 新增表

##### faq_items

```sql
CREATE TABLE faq_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category    VARCHAR(20) NOT NULL CHECK (category IN ('account','payment','credits','security')),
    question    TEXT NOT NULL,
    answer      TEXT NOT NULL,
    tags        TEXT[],
    sort_order  INTEGER NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_faq_category ON faq_items(category);
CREATE INDEX idx_faq_published ON faq_items(is_published) WHERE is_published = true;
CREATE INDEX idx_faq_search ON faq_items USING GIN (
    setweight(to_tsvector('simple', question), 'A') ||
    setweight(to_tsvector('simple', answer), 'B')
);
```

---

### 4.3 现有表变更

> 以下 ALTER TABLE 语句用于 V1.3.1 → V2.0 数据库迁移。

#### 4.3.1 users 表（account-service）

```sql
ALTER TABLE users
    ALTER COLUMN phone_number DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS locale VARCHAR(10) DEFAULT 'zh-CN',
    ADD COLUMN IF NOT EXISTS pending_downgrade_to VARCHAR(5),
    ADD COLUMN IF NOT EXISTS pending_downgrade_effective_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_locale ON users(locale);
```

| 变更项 | 说明 | 关联需求 |
|--------|------|---------|
| phone_number NULLABLE | 渐进式注册（UX-04），邮箱注册用户可无手机号 | UX-04 |
| locale | 多语言 i18n 架构（UX-17），存储用户语言偏好 | UX-17 |
| pending_downgrade_to | 降级目标等级，下期生效（UX-10） | UX-10 |
| pending_downgrade_effective_at | 降级生效时间（UX-10） | UX-10 |

#### 4.3.2 subscriptions 表（account-service）

```sql
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS pending_downgrade_to VARCHAR(5),
    ADD COLUMN IF NOT EXISTS pending_downgrade_effective_at TIMESTAMPTZ;
```

#### 4.3.3 config_items 表（config-service）

```sql
ALTER TABLE config_items
    ADD COLUMN IF NOT EXISTS reloadable BOOLEAN NOT NULL DEFAULT false;
```

| 变更项 | 说明 | 关联需求 |
|--------|------|---------|
| reloadable | 配置热更新（NF-05），标记该配置项是否支持运行时刷新 | NF-05 |

#### 4.3.4 password_hash 前缀标识（auth-service）

```sql
-- 无 schema 变更。password_hash 字段已为 VARCHAR(255)，足以容纳 argon2id 格式。
-- 新注册和改密使用 argon2id 格式：
--   $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
-- 存量 SM3 格式（过渡期）：
--   $sm3$<salt>$<hash>
-- 登录时根据前缀自动识别算法，SM3 验证成功后自动 rehash 为 argon2id。
-- 监控迁移进度：
--   SELECT COUNT(*) FROM users WHERE password_hash LIKE '$argon2id$%';
--   SELECT COUNT(*) FROM users WHERE password_hash LIKE '$sm3$%';
```

---

### 4.4 数据库迁移方案

> 使用 [Goose](https://github.com/pressly/goose) 管理数据库版本迁移。

#### 4.4.1 迁移文件命名规范

```
db-migrations/
├── 001_init_schema.sql                          # V1.0.0 初始 schema
├── 002_add_subscription_indexes.sql             # V1.1.0
├── 003_add_deletion_fields.sql                  # V1.2.0
├── 004_config_management_schema.sql             # V1.2.0
├── 005_add_extra_features.sql                   # V1.3.0
├── 006_add_identity_tier_to_users.sql           # V1.3.1
├── 007_add_deletion_indexes.sql                 # V1.3.1
│
├── 100_v2_create_payment_tables.sql             # V2.0 Phase 6 — payment-service
├── 101_v2_create_admin_tables.sql               # V2.0 Phase 6 — admin_users, admin_audit_logs
├── 102_v2_create_push_tables.sql                # V2.0 Phase 6 — push_tokens, push_logs
├── 103_v2_alter_users_nullable_phone.sql        # V2.0 Phase 6 — users.phone_number NULLABLE
├── 104_v2_create_orders_table.sql               # V2.0 Phase 6 — orders
├── 105_v2_create_payment_records_table.sql      # V2.0 Phase 6 — payment_records
├── 106_v2_create_refunds_table.sql              # V2.0 Phase 7 — refunds
├── 107_v2_create_social_accounts_table.sql      # V2.0 Phase 7 — social_accounts
├── 108_v2_create_biometric_tokens_table.sql     # V2.0 Phase 7 — biometric_device_tokens
├── 109_v2_alter_subscriptions_downgrade.sql     # V2.0 Phase 7 — pending_downgrade 字段
├── 110_v2_create_coupons_tables.sql             # V2.0 Phase 7 — coupons, coupon_usages
├── 111_v2_create_promotions_table.sql           # V2.0 Phase 7 — promotions
├── 112_v2_create_risk_events_extensions.sql     # V2.0 Phase 7 — risk_events 扩展
├── 113_v2_create_events_table.sql               # V2.0 Phase 7 — events（含初始分区）
├── 114_v2_create_notifications_table.sql        # V2.0 Phase 8 — notifications
├── 115_v2_create_notification_templates.sql     # V2.0 Phase 8 — templates, send_records
├── 116_v2_create_faq_items_table.sql            # V2.0 Phase 8 — faq_items
├── 117_v2_create_invoices_tables.sql            # V2.0 Phase 8 — invoices, user_invoice_info
├── 118_v2_create_push_strategies_tables.sql     # V2.0 Phase 8 — push_strategies, user_push_preferences
├── 119_v2_create_renewal_reminder_logs.sql      # V2.0 Phase 7 — renewal_reminder_logs
├── 120_v2_create_guest_sessions_table.sql       # V2.0 Phase 8 — guest_sessions
├── 121_v2_create_experiments_tables.sql         # V2.0 Phase 9 — experiments, experiment_assignments
├── 122_v2_create_enterprise_tables.sql          # V2.0 Phase 9 — enterprises, departments, users
├── 123_v2_create_leaderboard_tables.sql         # V2.0 Phase 9 — leaderboard_snapshots, user_leaderboard_preferences
├── 124_v2_alter_config_items_reloadable.sql     # V2.0 Phase 8 — config_items.reloadable
├── 125_v2_alter_users_add_locale.sql            # V2.0 Phase 9 — users.locale
└── 126_v2_alter_blacklist_extensions.sql        # V2.0 Phase 7 — blacklist_entries 扩展
```

#### 4.4.2 迁移执行规则

1. **每个迁移文件必须包含 `-- +goose Up` 和 `-- +goose Down`**
2. **CI 验证**：每次 PR 触发 `goose up && goose down && goose up` 三步验证
3. **Phase 标签**：文件名中含 Phase 编号，可按 Phase 分批执行
4. **零停机**：所有 ALTER TABLE 使用 `ADD COLUMN IF NOT EXISTS`，新增表不影响现有逻辑
5. **回滚安全**：Down 脚本只做 `DROP TABLE IF EXISTS` / `DROP COLUMN IF EXISTS`，不删除有数据的表

#### 4.4.3 events 分区维护

```sql
-- 由 Asynq 定时任务每月 25 日自动创建下月分区
CREATE TABLE events_YYYY_MM PARTITION OF events
    FOR VALUES FROM ('YYYY-MM-01') TO ('YYYY-MM+1-01');
```

---

### 4.5 Redis 数据结构变更

> 命名规范：`ac:{service}:{entity}:{identifier}`。V2.0 新增 Key 以 `ac:` 前缀，V1.3.1 存量 Key 保持兼容。

#### 4.5.1 V2.0 新增 Redis Key 清单

| Key 模式 | 类型 | 服务 | TTL | 说明 |
|----------|------|------|-----|------|
| `ac:payment:order:{order_id}` | Hash | payment-service | 24h | 订单缓存 |
| `ac:payment:lock:{order_id}` | String | payment-service | 5min | 支付防重复锁 |
| `ac:payment:reconcile:{date}` | Set | payment-service | 30d | 对账差异订单集合 |
| `ac:auth:session:{user_id}` | Hash | auth-service | 7d | 用户会话（V1.3.1 已有，格式不变） |
| `ac:auth:biometric:{user_id}:{device_fp}` | String | auth-service | 90d | 生物识别 token hash |
| `ac:auth:guest:{guest_id}` | String | auth-service | 24h | 游客 token 映射 |
| `ac:auth:oauth:state:{state}` | String | auth-service | 10min | OAuth state 防 CSRF |
| `ac:notify:push:freq:{user_id}:{YYYYMMDD}` | String | notification-service | 到次日 00:00 | 推送日频计数 |
| `ac:notify:push:delayed:{user_id}` | List | notification-service | 24h | 免打扰延迟队列 |
| `ac:notify:unread:{user_id}` | String | notification-service | 无 | 未读消息计数 |
| `ac:notify:token:{device_type}:{token_hash}` | Hash | notification-service | 无 | 设备 token 缓存 |
| `ac:account:user:{user_id}` | Hash | account-service | 5min | 用户信息缓存 |
| `ac:account:subscription:{user_id}` | Hash | account-service | 5min | 订阅信息缓存 |
| `ac:account:user_level:{user_id}` | String | account-service | 10min | 用户等级缓存 |
| `ac:account:renewal:{user_id}:{T-N}:{date}` | String | account-service | 48h | 续费提醒去重 |
| `ac:account:leaderboard:weekly` | Sorted Set | account-service | 8d | 周排行榜 |
| `ac:account:leaderboard:monthly` | Sorted Set | account-service | 32d | 月排行榜 |
| `ac:account:leaderboard:profile:{uid}` | Hash | account-service | 8d | 排行榜用户资料 |
| `ac:account:leaderboard:privacy:{uid}` | String | account-service | 无 | 排行榜隐私设置 |
| `ac:account:social_proof:{uid}` | String | account-service | 1h | 社交证明数据 |
| `ac:account:social_proof:global` | String | account-service | 1h | 全局社交证明 |
| `ac:compliance:blacklist:ip:{ip}` | String | compliance-service | 与 PG 同步 | IP 黑名单缓存 |
| `ac:compliance:blacklist:device:{fp}` | String | compliance-service | 与 PG 同步 | 设备黑名单缓存 |
| `ac:compliance:blacklist:user:{uid}` | String | compliance-service | 与 PG 同步 | 用户黑名单缓存 |
| `ac:compliance:register_count:{ip}:{date}` | String | compliance-service | 25h | IP 注册计数 |
| `ac:compliance:register_count:device:{fp}:{date}` | String | compliance-service | 25h | 设备注册计数 |
| `ac:dataproduct:online` | HyperLogLog | data-product-service | 2min | 在线用户集合 |
| `ac:dataproduct:funnel:{event}:{YYYYMMDDHH}` | String | data-product-service | 48h | 漏斗步骤计数 |
| `ac:dataproduct:revenue:{YYYYMMDDHH}` | String | data-product-service | 48h | 实时收入累计 |
| `ac:dataproduct:anomaly:{uid}:{event}` | String | data-product-service | 1min | 异常行为计数 |
| `ac:ab:{uid}:{experiment_id}` | String | data-product-service | 24h | A/B 分组缓存 |
| `ac:config:hot:{item_code}` | String | config-service | 30s | 热更新配置缓存 |

#### 4.5.2 V1.3.1 存量 Key 兼容

V1.3.1 已有的 Key（如 `session:{userID}`、`cache:user:{userID}`）在账号注销 Worker 中清理。V2.0 新增 Key 遵循 `ac:` 前缀规范，存量 Key 在后续版本逐步迁移。

---

## 5. API 设计

### 5.1 新增 API 清单

> 以下为 V2.0 全部新增 API，按服务分组。所有 API 通过 api-gateway(30300) 统一入口，路径前缀 `/api/v1`。

#### 5.1.1 payment-service API（端口 30316）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/orders` | 创建订单 | JWT | 6 |
| GET | `/api/v1/orders/{id}` | 查询订单详情 | JWT | 6 |
| GET | `/api/v1/orders` | 订单列表（分页、过滤） | JWT | 6 |
| POST | `/api/v1/orders/{id}/cancel` | 取消订单 | JWT | 6 |
| POST | `/api/v1/payments` | 创建支付 | JWT | 6 |
| GET | `/api/v1/payments/{id}` | 查询支付状态 | JWT | 6 |
| POST | `/api/v1/payments/wechat/callback` | 微信支付回调 | 签名验证 | 6 |
| POST | `/api/v1/payments/alipay/callback` | 支付宝回调 | 签名验证 | 6 |
| POST | `/api/v1/refunds` | 申请退款 | JWT | 7 |
| GET | `/api/v1/refunds/{id}` | 查询退款状态 | JWT | 7 |
| POST | `/api/v1/invoices` | 申请开票 | JWT | 8 |
| GET | `/api/v1/invoices/{id}` | 查询发票详情 | JWT | 8 |
| GET | `/api/v1/invoices` | 发票列表 | JWT | 8 |
| PUT | `/api/v1/users/me/invoice-info` | 维护默认发票信息 | JWT | 8 |
| GET | `/api/v1/admin/orders` | 管理端订单查询 | Admin JWT | 6 |
| GET | `/api/v1/admin/orders/export` | 订单导出（CSV/Excel） | Admin JWT | 6 |
| GET | `/api/v1/admin/reconciliation` | 对账结果查询 | Admin JWT | 6 |
| POST | `/api/v1/admin/refunds/{id}/approve` | 审核通过退款 | Admin JWT | 7 |
| POST | `/api/v1/admin/refunds/{id}/reject` | 审核驳回退款 | Admin JWT | 7 |
| GET | `/api/v1/admin/invoices` | 管理端发票查询 | Admin JWT | 8 |

**关键 Request/Response 示例**:

```
POST /api/v1/orders
Request:
  { "product_type": "subscription", "product_id": "plan_L2_monthly",
    "credits_used": 100, "payment_method": "wechat", "payment_channel": "h5" }
Response 201:
  { "id": "uuid", "order_no": "ORD20260519123456", "amount_cents": 2900,
    "actual_amount_cents": 1900, "status": "pending", "expires_at": "..." }
```

#### 5.1.2 account-service Admin API（端口 30301）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| GET | `/api/v1/admin/users` | 用户列表（分页、过滤） | Admin JWT | 6 |
| GET | `/api/v1/admin/users/{id}` | 用户详情 | Admin JWT | 6 |
| PUT | `/api/v1/admin/users/{id}/level` | 等级调整 | Admin JWT | 6 |
| POST | `/api/v1/admin/users/{id}/credits/adjust` | 积分调整 | Admin JWT | 6 |
| PUT | `/api/v1/admin/users/{id}/ban` | 封禁用户 | Admin JWT | 6 |
| PUT | `/api/v1/admin/users/{id}/unban` | 解封用户 | Admin JWT | 6 |
| PUT | `/api/v1/admin/users/{id}/identity/approve` | 实名审核通过 | Admin JWT | 6 |
| PUT | `/api/v1/admin/users/{id}/identity/reject` | 实名审核驳回 | Admin JWT | 6 |
| GET | `/api/v1/admin/audit-logs` | 审计日志查询 | Admin JWT | 6 |
| POST | `/api/v1/admin/coupons` | 创建优惠券 | Admin JWT | 7 |
| GET | `/api/v1/admin/coupons` | 优惠券列表 | Admin JWT | 7 |
| PUT | `/api/v1/admin/coupons/{id}` | 编辑优惠券 | Admin JWT | 7 |
| DELETE | `/api/v1/admin/coupons/{id}` | 删除优惠券 | Admin JWT | 7 |
| POST | `/api/v1/admin/coupons/batch-generate` | 批量生成优惠券 | Admin JWT | 7 |
| POST | `/api/v1/admin/coupons/{id}/invalidate` | 作废优惠券 | Admin JWT | 7 |
| POST | `/api/v1/admin/promotions` | 创建促销活动 | Admin JWT | 7 |
| GET | `/api/v1/admin/promotions` | 促销活动列表 | Admin JWT | 7 |
| PUT | `/api/v1/admin/promotions/{id}` | 编辑促销活动 | Admin JWT | 7 |
| DELETE | `/api/v1/admin/promotions/{id}` | 删除促销活动 | Admin JWT | 7 |
| POST | `/api/v1/admin/plans` | 创建套餐 | Admin JWT | 7 |
| GET | `/api/v1/admin/plans` | 套餐列表 | Admin JWT | 7 |
| PUT | `/api/v1/admin/plans/{id}` | 编辑套餐 | Admin JWT | 7 |
| DELETE | `/api/v1/admin/plans/{id}` | 删除套餐 | Admin JWT | 7 |
| POST | `/api/v1/admin/enterprise/sync` | 手动触发通讯录同步 | Admin JWT | 9 |
| GET | `/api/v1/admin/enterprise/departments` | 企业部门树 | Admin JWT | 9 |
| GET | `/api/v1/admin/enterprise/users` | 企业员工列表 | Admin JWT | 9 |
| PUT | `/api/v1/admin/enterprise/users/{id}/permissions` | 设置企业员工权限 | Admin JWT | 9 |

**account-service 用户侧新增 API**:

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/coupons/validate` | 验证优惠券 | JWT | 7 |
| POST | `/api/v1/subscriptions/upgrade/preview` | 升级费用预览 | JWT | 7 |
| POST | `/api/v1/subscriptions/upgrade/confirm` | 确认升级 | JWT | 7 |
| POST | `/api/v1/subscriptions/downgrade/preview` | 降级预览 | JWT | 7 |
| POST | `/api/v1/subscriptions/downgrade/confirm` | 确认降级 | JWT | 7 |
| GET | `/api/v1/subscriptions/renewal/status` | 续费提醒状态 | JWT | 7 |
| PUT | `/api/v1/subscriptions/renewal/preferences` | 设置提醒偏好 | JWT | 7 |
| GET | `/api/v1/dashboard/config` | 仪表盘配置 | JWT | 7 |
| GET | `/api/v1/leaderboard` | 排行榜（Top 20） | JWT | 9 |
| PUT | `/api/v1/leaderboard/privacy` | 设置排行榜隐私 | JWT | 9 |
| GET | `/api/v1/social-proof/referral` | 推荐社交证明 | JWT | 9 |
| GET | `/api/v1/social-proof/upgrade` | 升级社交证明 | JWT | 9 |

#### 5.1.3 notification-service API（端口 30311）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/devices/register` | 注册/更新设备 Token | JWT | 6 |
| DELETE | `/api/v1/devices/{token}` | 注销设备 | JWT | 6 |
| POST | `/api/v1/notifications/push` | 发送推送（内部 API） | Service | 6 |
| GET | `/api/v1/notifications` | 消息列表（分页、筛选） | JWT | 8 |
| GET | `/api/v1/notifications/unread-count` | 未读消息计数 | JWT | 8 |
| PUT | `/api/v1/notifications/{id}/read` | 标记已读 | JWT | 8 |
| PUT | `/api/v1/notifications/read-all` | 全部已读 | JWT | 8 |
| PUT | `/api/v1/users/me/push-preferences` | 设置免打扰 | JWT | 8 |
| POST | `/api/v1/admin/notifications/templates` | 创建通知模板 | Admin JWT | 8 |
| PUT | `/api/v1/admin/notifications/templates/{id}` | 编辑通知模板 | Admin JWT | 8 |
| GET | `/api/v1/admin/notifications/templates` | 模板列表 | Admin JWT | 8 |
| POST | `/api/v1/admin/notifications/send` | 发送通知（定时/定向） | Admin JWT | 8 |
| GET | `/api/v1/admin/notifications/send-records` | 发送记录查询 | Admin JWT | 8 |
| POST | `/api/v1/admin/notifications/preview` | 预览渲染结果 | Admin JWT | 8 |
| GET | `/api/v1/admin/notifications/push/logs` | 推送日志查询 | Admin JWT | 6 |
| POST | `/api/v1/admin/push/strategies` | 创建推送策略 | Admin JWT | 8 |
| GET | `/api/v1/admin/push/strategies` | 策略列表 | Admin JWT | 8 |
| PUT | `/api/v1/admin/push/strategies/{id}` | 编辑策略 | Admin JWT | 8 |

#### 5.1.4 auth-service API（端口 30302）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/auth/social/login` | 社交登录 | Public | 7 |
| POST | `/api/v1/auth/social/bind` | 绑定社交账号 | JWT | 7 |
| DELETE | `/api/v1/auth/social/unbind/{provider}` | 解绑社交账号 | JWT | 7 |
| GET | `/api/v1/auth/social/accounts` | 已绑定社交账号列表 | JWT | 7 |
| GET | `/api/v1/auth/oauth/{provider}/url` | 获取 OAuth 授权 URL | Public | 7 |
| POST | `/api/v1/auth/oauth/{provider}/callback` | OAuth 回调处理 | Public | 7 |
| POST | `/api/v1/auth/biometric/enroll` | 启用生物识别 | JWT | 7 |
| POST | `/api/v1/auth/biometric/login` | 生物识别登录 | Public | 7 |
| DELETE | `/api/v1/auth/biometric/disable` | 关闭生物识别 | JWT | 7 |
| POST | `/api/v1/auth/biometric/refresh` | 刷新设备 token | JWT | 7 |
| POST | `/api/v1/auth/guest/token` | 获取游客 token | Public | 8 |
| POST | `/api/v1/auth/register/email` | 邮箱注册（L0） | Public | 8 |
| POST | `/api/v1/account/bind-phone` | 绑定手机号 | JWT | 8 |
| GET | `/api/v1/auth/enterprise/work-wechat/authorize` | 企微扫码登录 | Public | 9 |
| GET | `/api/v1/auth/enterprise/work-wechat/callback` | 企微登录回调 | Public | 9 |
| GET | `/api/v1/auth/enterprise/dingtalk/authorize` | 钉钉扫码登录 | Public | 9 |
| GET | `/api/v1/auth/enterprise/dingtalk/callback` | 钉钉登录回调 | Public | 9 |
| POST | `/api/v1/enterprise/approval/callback` | 审批结果回调 | 签名验证 | 9 |

#### 5.1.5 data-product-service API（端口 30314）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/events/batch` | 批量上报事件（≤50 条） | JWT | 7 |
| GET | `/api/v1/admin/metrics/registration-trend` | 注册趋势 | Admin JWT | 7 |
| GET | `/api/v1/admin/metrics/conversion-funnel` | 付费转化漏斗 | Admin JWT | 7 |
| GET | `/api/v1/admin/metrics/revenue` | MRR/ARR | Admin JWT | 7 |
| GET | `/api/v1/admin/metrics/rfm-distribution` | RFM 分布 | Admin JWT | 7 |
| GET | `/api/v1/admin/metrics/kfactor` | 推荐 K-factor | Admin JWT | 7 |
| POST | `/api/v1/events/collect` | 事件收集（实时流） | JWT | 8 |
| GET | `/api/v1/realtime/online-count` | 当前在线人数 | Admin JWT | 8 |
| GET | `/api/v1/realtime/funnel` | 实时转化漏斗 | Admin JWT | 8 |
| GET | `/api/v1/realtime/revenue` | 实时收入 | Admin JWT | 8 |
| GET | `/api/v1/admin/realtime/alerts` | 异常行为告警 | Admin JWT | 8 |
| POST | `/api/v1/data/export` | 申请个人数据导出 | JWT | 7 |
| GET | `/api/v1/data/export/{id}/download` | 下载导出文件 | JWT | 7 |
| GET | `/api/v1/admin/reports/{type}/export` | 运营报表导出 | Admin JWT | 7 |
| GET | `/api/v1/ab/experiments/{id}/variant` | 获取实验变体 | JWT | 9 |
| GET | `/api/v1/ab/experiments` | 实验列表 | Admin JWT | 9 |
| POST | `/api/v1/ab/experiments` | 创建实验 | Admin JWT | 9 |
| PUT | `/api/v1/ab/experiments/{id}/status` | 启停实验 | Admin JWT | 9 |
| GET | `/api/v1/ab/experiments/{id}/report` | 实验报告 | Admin JWT | 9 |
| POST | `/api/v1/ab/experiments/{id}/rollout` | 全量推送胜出方案 | Admin JWT | 9 |

#### 5.1.6 compliance-service API（端口 30313）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/admin/blacklist` | 新增黑名单 | Admin JWT | 7 |
| GET | `/api/v1/admin/blacklist` | 黑名单列表 | Admin JWT | 7 |
| DELETE | `/api/v1/admin/blacklist/{id}` | 删除黑名单 | Admin JWT | 7 |
| POST | `/api/v1/admin/blacklist/batch` | 批量导入黑名单 | Admin JWT | 7 |
| GET | `/api/v1/admin/risk-events` | 风险事件列表 | Admin JWT | 7 |
| PUT | `/api/v1/admin/risk-events/{id}/resolve` | 处置风险事件 | Admin JWT | 7 |
| GET | `/api/v1/admin/risk-events/suspicious-registrations` | 可疑注册列表 | Admin JWT | 7 |

#### 5.1.7 config-service API（端口 30315）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| GET | `/api/v1/faq` | FAQ 列表（搜索/分类） | Public | 8 |
| POST | `/api/v1/admin/faq` | 创建 FAQ | Admin JWT | 8 |
| PUT | `/api/v1/admin/faq/{id}` | 编辑 FAQ | Admin JWT | 8 |
| DELETE | `/api/v1/admin/faq/{id}` | 删除 FAQ | Admin JWT | 8 |

#### 5.1.8 开放 API（api-gateway OAuth2）

| Method | Path | 说明 | Auth | Phase |
|--------|------|------|------|-------|
| POST | `/api/v1/oauth2/token` | 获取开放 API Token | Client Credentials | 7 |
| GET | `/api/v1/open/user/info` | 用户信息查询 | OAuth2 Token | 7 |
| GET | `/api/v1/open/credits/balance` | 积分查询 | OAuth2 Token | 7 |

---

### 5.2 修改 API 清单

> 以下 V1.3.1 已有 API 在 V2.0 中存在变更。

#### 5.2.1 Breaking Changes（需客户端适配）

| API | 变更类型 | 变更内容 | 影响范围 | Phase |
|-----|---------|---------|---------|-------|
| `POST /api/v1/auth/login` | Response 变更 | password_hash 字段新增 `$argon2id$` / `$sm3$` 前缀标识（服务端内部，API 无变化） | 仅服务端 | 6 |
| `POST /api/v1/account/register` | 参数变更 | `phone_number` 改为可选（UX-04 渐进式注册），`email` 可作为注册凭证 | 全端 | 8 |
| Gateway 所有 API | 行为变更 | 新增 30s ResponseHeaderTimeout / 60s 全局超时 / 504 超时响应 | 全端 | 6 |
| Gateway 所有 API | 行为变更 | 新增黑名单检查中间件（IP/设备/用户级拦截，返回 403） | 全端 | 7 |

#### 5.2.2 Non-Breaking Changes（向后兼容）

| API | 变更类型 | 变更内容 | Phase |
|-----|---------|---------|-------|
| `GET /api/v1/account/me` | Response 扩展 | 新增 `locale`、`pending_downgrade_to`、`pending_downgrade_effective_at` 字段 | 9 |
| `GET /api/v1/subscriptions/current` | Response 扩展 | 新增 `pending_downgrade_to`、`pending_downgrade_effective_at` 字段 | 7 |
| `GET /internal/v1/config/items/{code}` | 行为变更 | 支持热更新轮询（NF-05），每 30s 检查 `reloadable=true` 的配置项 | 8 |
| `POST /api/v1/sms/send` | 行为变更 | 新增模板引擎变量插值（FN-09），通过 `template_code` 指定模板 | 8 |
| `POST /api/v1/email/send` | 行为变更 | 同上，支持模板变量插值 | 8 |
| `GET /api/v1/referral/dashboard` | Response 扩展 | 新增漏斗图数据（share→register→verify→paid 转化率） | 7 |
| `POST /api/v1/referral/share` | Response 扩展 | 新增 `share_image_url`（海报图）和 `og_preview_url`（OG 预览） | 8 |
| `/health` | 行为变更 | 新增 PG/Redis/下游依赖检测，不可用时返回 503 + JSON 详情 | 7 |

---

### 5.3 OpenAPI 规范

#### 5.3.1 自动生成方案

采用 [Swaggo](https://github.com/swaggo/swag) 从 Go 源码注解自动生成 OpenAPI 3.0 规范文档。

**工具链**:

```
Go 源码 (swag 注解)
    → swag init (生成 docs/)
    → swag fmt (格式化注解)
    → Swagger UI (在线文档)
```

**注解示例**:

```go
// @Summary      创建订单
// @Description  创建支付订单，支持积分抵扣
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateOrderRequest  true  "订单请求"
// @Success      201   {object}  OrderResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Router       /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) { ... }
```

#### 5.3.2 执行策略

| 步骤 | 工具 | 时机 | 说明 |
|------|------|------|------|
| 注解编写 | 开发者 | 每个新 API 实现时 | 在 handler 方法上添加 swag 注解 |
| 规范生成 | `swag init` | CI 构建 | 遍历所有服务 handler，生成 `docs/swagger.json` |
| 格式校验 | `swag fmt` | CI lint 阶段 | 统一注解格式 |
| 一致性检查 | 自定义脚本 | CI | 对比注解数量与实际路由数量，覆盖率 100% |
| 文档发布 | Swagger UI | 每次部署 | 挂载在 `/docs` 路径（仅 staging/internal 环境） |

#### 5.3.3 覆盖率目标

- **Phase 6**：payment-service 全部 API 注解完成（~20 个端点）
- **Phase 7**：account-service Admin API + auth-service OAuth API 注解完成（~30 个端点）
- **Phase 8**：notification-service + config-service FAQ API 注解完成（~20 个端点）
- **Phase 9**：data-product-service A/B + enterprise API 注解完成（~15 个端点）
- **最终目标**：100% API 端点覆盖 swag 注解，CI 警告未覆盖的新增端点

#### 5.3.4 多服务文档聚合

```
api-gateway/
└── docs/
    ├── swagger.json              # 聚合全部服务 API
    ├── payment-service.json      # payment-service API
    ├── account-service.json      # account-service API
    ├── auth-service.json         # auth-service API
    ├── notification-service.json # notification-service API
    ├── compliance-service.json   # compliance-service API
    ├── data-product-service.json # data-product-service API
    └── config-service.json       # config-service API
```

每个服务 CI 生成各自的 `swagger.json`，api-gateway CI 聚合为统一文档。Swagger UI 按服务 tag 分组展示。

## 6. 安全设计

### 6.1 密码哈希迁移方案

**目标**：将现有 SM3+salt 哈希升级至 argon2id，同时保持向后兼容与审计链完整性。

**argon2id 参数配置**：

| 参数 | 值 | 说明 |
|------|----|------|
| memory | 64 MB (65536 KB) | 抗 GPU 暴力破解 |
| iterations | 3 | 通行迭代次数 |
| parallelism | 2 | 并行线程数 |
| salt length | 16 bytes | 随机盐值长度 |
| key length | 32 bytes | 输出哈希长度 |

**迁移策略**：

```
┌─────────────┐     登录请求      ┌──────────────────┐
│  新注册用户   │ ──────────────→ │ 直接 argon2id    │
└─────────────┘                  │ 前缀: $argon2id$ │
                                  └──────────────────┘

┌─────────────┐     登录请求      ┌──────────────────┐     验证通过     ┌──────────────────┐
│  存量 SM3    │ ──────────────→ │ SM3 验证         │ ─────────────→ │ Rehash → argon2id│
│  用户        │                  │ 前缀: $sm3$      │                │ 透明写入新哈希    │
└─────────────┘                  └──────────────────┘                 └──────────────────┘
```

**password_hash 字段前缀规范**：

| 前缀 | 算法 | 用途 |
|------|------|------|
| `$argon2id$` | argon2id | 新注册 / 已迁移用户密码验证 |
| `$sm3$` | SM3+salt | 存量用户密码验证（仅登录时使用） |

**SM3 保留场景**：
- 审计日志哈希链完整性校验（audit_log 表 hash_chain 字段仍使用 SM3）
- 历史数据签名验证
- 不再用于新密码哈希

**迁移监控指标**：

| 指标 | 采集方式 | 告警阈值 |
|------|---------|---------|
| `auth_password_sm3_remaining` | 统计 `password_hash LIKE '$sm3$%'` 行数 | > 总用户 50% 持续 7 天 |
| `auth_password_rehash_total` | 登录成功后 rehash 计数 | 低于日活 10% 异常 |
| `auth_password_rehash_errors` | rehash 失败计数 | > 0 触发 P1 告警 |

**数据变更**：

```sql
ALTER TABLE users ALTER COLUMN password_hash TYPE TEXT;
COMMENT ON COLUMN users.password_hash IS '格式: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash> 或 $sm3$<hash>';
```

**测试策略**：
- 单元测试：argon2id 哈希生成与验证、SM3 验证后自动 rehash
- 集成测试：模拟存量 SM3 用户登录 → 验证密码正确 → 验证 password_hash 更新为 argon2id 前缀
- 性能测试：argon2id 哈希计算耗时 < 200ms（目标机器）
- 回归测试：已迁移 argon2id 用户二次登录仍可正常验证

---

### 6.2 KMS/Vault 集成方案

**密钥管理方案选型对比**：

| 维度 | HashiCorp Vault | 阿里云 KMS |
|------|----------------|------------|
| 部署方式 | 自托管（Dev 模式 / HA 集群） | 全托管 SaaS |
| 密钥类型 | 对称加密 / 非对称 / SSH / TLS | 对称 / 非对称 / 信封加密 |
| 自动轮换 | 支持（需配置） | 原生支持 |
| 审计日志 | Syslog / File / 云存储 | ActionTrail 集成 |
| 紧急吊销 | `vault revoke` 即时生效 | API 调用即时生效 |
| 网络要求 | 内网可达即可 | 需公网/VPC Endpoint |
| 成本 | 开源免费 + 运维成本 | 按调用计费 |
| 适用阶段 | Dev / UAT | Prod |

**推荐策略**：Dev/UAT 使用 Vault（Dev 模式），Prod 迁移至阿里云 KMS，通过统一抽象层屏蔽差异。

**密钥轮换策略**：

```
┌─────────────┐    Day 0     ┌─────────────┐
│ Key v1 (当前) │ ──────────→ │ 加密新数据   │
└─────────────┘              └─────────────┘
       │ Day 90
       ▼
┌─────────────┐    Day 90    ┌─────────────┐
│ Key v2 (新)  │ ──────────→ │ 加密新数据   │
└─────────────┘              └─────────────┘
       │
       ▼
┌─────────────┐   Day 90-180  ┌─────────────────┐
│ Key v1 (旧)  │ ──────────→  │ 仅解密，不再加密 │
└─────────────┘               └─────────────────┘
       │ Day 180
       ▼
   Key v1 销毁
```

| 参数 | 值 |
|------|----|
| 轮换周期 | 90 天自动轮换 |
| 重叠期 | 90 天（旧密钥仅解密） |
| 通知 | 轮换前 7 天企业微信通知运维组 |

**审计追踪**：
- 所有密钥操作（创建/轮换/吊销/访问）写入 `key_audit_log` 表
- 记录字段：`operation`, `key_id`, `key_version`, `operator`, `source_ip`, `timestamp`
- 审计日志不可篡改（SM3 哈希链保护）

**紧急吊销流程**：

```
发现密钥泄露
    │
    ▼
运维触发 vault revoke / KMS ScheduleKeyDeletion
    │
    ▼
所有使用该密钥的缓存立即失效
    │
    ▼
触发全量数据重加密（使用新密钥版本）
    │
    ▼
通知安全团队 + 生成事件报告
```

**抽象层接口设计**：

```go
type KeyManager interface {
    Encrypt(ctx context.Context, plaintext []byte) (ciphertext []byte, keyVersion string, error)
    Decrypt(ctx context.Context, ciphertext []byte, keyVersion string) ([]byte, error)
    RotateKey(ctx context.Context) (newVersion string, error)
    RevokeKey(ctx context.Context, version string) error
}
```

**测试策略**：
- 单元测试：Vault / KMS 两个实现的 Encrypt→Decrypt 一致性
- 集成测试：密钥轮换后旧密文仍可解密
- 混沌测试：KMS 不可达时降级至本地缓存密钥（TTL 5 分钟）

---

### 6.3 API 安全加固

**用户级限流设计**：

基于 Redis 计数器实现滑动窗口限流，按用户等级差异化配置：

| 用户等级 | 限流阈值 | 窗口 | Redis Key 格式 |
|---------|---------|------|----------------|
| L0（免费） | 60 次/分钟 | 60s | `ratelimit:user:{uid}:L0` |
| L1（基础） | 120 次/分钟 | 60s | `ratelimit:user:{uid}:L1` |
| L2+（高级/企业） | 300 次/分钟 | 60s | `ratelimit:user:{uid}:L2` |

**限流响应规范**：

```json
HTTP/1.1 429 Too Many Requests
Retry-After: 30
Content-Type: application/json

{
    "code": 42901,
    "message": "请求频率超限，请稍后再试",
    "retry_after": 30
}
```

**限流中间件实现要点**：
- 使用 Redis `INCR` + `EXPIRE` 实现固定窗口计数
- 网关层统一拦截，响应 Header 包含 `X-RateLimit-Limit` / `X-RateLimit-Remaining` / `X-RateLimit-Reset`
- 管理员 API（`/admin/*`）独立限流配置，不与用户级共享

**HMAC-SHA256 请求签名**（关键写操作）：

适用范围：密码修改、邮箱/手机变更、订阅购买、积分兑换等敏感操作。

```
签名生成：
1. 构造签名字符串 = HTTP_METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + BODY_SHA256
2. signature = HMAC-SHA256(client_secret, signed_string)
3. 请求 Header: X-Signature: hmac-sha256 {signature}
                X-Timestamp: {unix_timestamp}
```

**验证规则**：
- Timestamp 与服务器时间偏差 ≤ 5 分钟（防重放）
- Body SHA256 校验（防篡改）
- 签名失败返回 `HTTP 401` + `INVALID_SIGNATURE`

**CI 安全自动化扫描**：

| 扫描类型 | 工具 | 触发时机 | 报告输出 |
|---------|------|---------|---------|
| SQL 注入 | SQLMap（Docker 镜像） | 每次 PR 合并至 main | `security/sqlmap-report.json` |
| XSS / OWASP Top 10 | OWASP ZAP（全量扫描） | 每日定时 + 发布前 | `security/zap-report.html` |
| 依赖漏洞 | Trivy / Snyk | 每次 build | `security/trivy-report.json` |
| Go 安全审计 | gosec | golangci-lint 集成 | CI 日志 |

**安全审计日志**：

所有安全相关操作写入 `security_audit_log` 表：

| 字段 | 类型 | 说明 |
|------|------|------|
| event_type | VARCHAR | `login_success` / `login_failed` / `password_change` / `email_change` / `rate_limit_hit` / `signature_invalid` |
| user_id | UUID | 操作用户 |
| source_ip | INET | 来源 IP |
| user_agent | TEXT | 客户端标识 |
| metadata | JSONB | 扩展信息（如登录设备、失败原因） |
| created_at | TIMESTAMPTZ | 事件时间 |

**测试策略**：
- 单元测试：限流计数器精度、签名生成与验证、Timestamp 过期校验
- 集成测试：超过限流阈值返回 429、签名错误返回 401
- 安全测试：SQLMap + ZAP 扫描零高危漏洞

---

### 6.4 移动端安全

**6.4.1 Certificate Pinning（证书固定）**

| 平台 | 实现方式 | 配置 |
|------|---------|------|
| iOS | `URLSessionDelegate` + `SecTrustEvaluateWithError` | 固定服务器公钥 SHA256 指纹，支持主备两枚证书 |
| Android | `OkHttp CertificatePinner` | `certificatePinner { "api.neuro.xxx", "sha256/..." }` |

**降级策略**：证书固定失败时禁止静默跳过，必须提示用户"网络环境不安全"并阻止请求。Pin 更新通过热修复配置下发（不依赖应用发版）。

**6.4.2 Root / 越狱检测**

| 平台 | 检测方法 | 触发动作 |
|------|---------|---------|
| iOS | 文件系统检测（Cydia/Substrate 路径）、沙盒完整性校验、`dyld` 环境变量检查 | 标记设备风险等级 → 限制敏感操作（如提现、大额兑换） |
| Android | Google Play Integrity API（设备/应用完整性）、`su` 二进制检测、Magisk Hide 检测 | 同 iOS |

**合规要求**：检测结果仅用于风险等级标记，不阻止基础功能使用（符合应用商店审核政策）。

**6.4.3 截屏防护**

| 平台 | 实现方式 | 适用页面 |
|------|---------|---------|
| iOS | `UIScreen.main.isCaptured` 监听 + `UITextField.isSecureTextEntry` 遮挡 | 个人信息、支付、密码页面 |
| Android | `Window.FLAG_SECURE` | 同 iOS |

**附加措施**：截屏事件发生后自动记录安全审计日志（设备 + 页面 + 时间）。

**6.4.4 Token 存储规范**

| Token 类型 | 存储位置 | 加密方式 | 有效期 |
|-----------|---------|---------|-------|
| access_token | 内存（不持久化） | - | 15 分钟 |
| refresh_token | iOS: Keychain / Android: Keystore | AES-256-GCM（密钥由设备安全芯片派生） | 30 天 |
| CSRF token | 内存 | - | 单次会话 |

**Token 生命周期安全**：
- access_token 过期后使用 refresh_token 静默刷新，连续失败 3 次强制重新登录
- refresh_token 每次使用后轮换（Rotation），旧 token 立即失效
- 设备注销时清除 Keychain/Keystore 中所有 token
- 同一用户最多 5 个活跃设备，超出自动淘汰最旧设备

**测试策略**：
- iOS：XCUITest 验证 Certificate Pinning 拦截中间人攻击
- Android：Compose Testing 验证 FLAG_SECURE 截屏返回空白
- 安全审计：MobSF（移动安全框架）静态 + 动态扫描
- 渗透测试：Frida hook 检测绕过难度评估

---

## 7. 部署设计

### 7.1 多环境差异

系统部署三套环境：Dev（开发）、UAT（验收测试）、Prod（生产），遵循"版本一致、部署差异"原则。

| 维度 | Dev | UAT | Prod |
|------|-----|-----|------|
| PostgreSQL | Docker 单容器 | ECS 本地部署（单节点，多项目共享） | RDS PG 18.x HA（主备自动切换） |
| Redis | Docker 单容器 | ECS 本地部署（单节点 AOF 持久化，多项目共享） | Redis Sentinel（3 节点）或 Cluster |
| Kafka | Redis Streams（兼容层） | Redis Streams（兼容层） | Kafka 4.2.x 独立集群（3 Broker） |
| 应用部署 | Docker Compose（本地） | Docker Compose（ECS 单节点） | K8s Helm Chart（多副本 HPA） |
| 密钥管理 | .env 文件 | .env + config-service | Vault / 阿里云 KMS + config-service |
| SSL/TLS | HTTP（无证书） | 自签名证书 | 正式证书（ACME/商业） |
| 监控 | VM + Promtail + Loki + Grafana（容器化） | VM + Promtail + Loki + Grafana（ECS 本地部署，多项目共享） | VM HA + Promtail + Loki + Grafana + AlertManager |
| 日志 | Promtail → Loki（容器化） | Promtail → Loki（ECS 本地部署，多项目共享） | Promtail → Loki 聚合 + OSS 归档（180 天） |
| 备份 | 无 | pg_dump 每日 | PITR + 每日全量 + OSS 异地 |
| 对象存储 | 本地卷 | ECS 本地部署 MinIO（多项目共享） | 阿里云 OSS |
| 负载均衡 | 无 | 无（Docker Compose 直连） | K8s Ingress / SLB |
| 域名 | localhost:port | uat-wxxx.neurongene.cn | api.neuro.xxx.com |

**关键约束**：
- 三环境所有技术栈二进制版本/镜像 tag 必须完全一致
- UAT 基础设施为多项目共享，通过不同 DB schema / Redis key 前缀 / Grafana folder 做逻辑隔离
- Dev 环境必须配置完整监控和日志栈，确保开发阶段即可发现性能和日志问题

---

### 7.2 Helm Chart 结构

```
helm/
└── account-center/
    ├── Chart.yaml
    ├── values.yaml
    ├── values-dev.yaml
    ├── values-uat.yaml
    ├── values-prod.yaml
    └── templates/
        ├── _helpers.tpl
        ├── api-gateway/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── auth-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── account-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── credit-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── subscription-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── rebate-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── notification-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── data-product-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── config-service/
        │   ├── deployment.yaml
        │   ├── service.yaml
        │   ├── configmap.yaml
        │   └── hpa.yaml
        ├── ingress.yaml
        ├── secrets.yaml
        └── monitoring/
            ├── servicemonitor.yaml
            └── prometheusrules.yaml
```

**Chart.yaml**：

```yaml
apiVersion: v2
name: account-center
description: Account Center V2.0 微服务 Helm Chart
type: application
version: 2.0.0
appVersion: "2.0.0"
maintainers:
  - name: account-center-team
```

**values.yaml 核心配置**：

```yaml
global:
  environment: production
  imageRegistry: registry.cn-hangzhou.aliyuncs.com/neuro
  imagePullSecrets: []
  storageClass: ""

services:
  api-gateway:
    replicas: 2
    image:
      repository: account-center/api-gateway
      tag: "2.0.0"
      pullPolicy: IfNotPresent
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits: { cpu: 500m, memory: 512Mi }
    hpa:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      targetCPUUtilizationPercentage: 70
    env:
      DB_HOST: "{{ .Values.global.dbHost }}"
      REDIS_ADDR: "{{ .Values.global.redisAddr }}"
  auth-service:
    replicas: 2
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits: { cpu: 500m, memory: 512Mi }
    hpa:
      enabled: true
      minReplicas: 2
      maxReplicas: 8
      targetCPUUtilizationPercentage: 70
  account-service:
    replicas: 2
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits: { cpu: 500m, memory: 512Mi }
  credit-service:
    replicas: 2
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits: { cpu: 500m, memory: 512Mi }
  subscription-service:
    replicas: 2
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits: { cpu: 500m, memory: 512Mi }
  rebate-service:
    replicas: 2
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits: { cpu: 500m, memory: 512Mi }
  notification-service:
    replicas: 1
    resources:
      requests: { cpu: 50m, memory: 64Mi }
      limits: { cpu: 250m, memory: 256Mi }
  data-product-service:
    replicas: 1
    resources:
      requests: { cpu: 100m, memory: 256Mi }
      limits: { cpu: 1000m, memory: 1Gi }
  config-service:
    replicas: 1
    resources:
      requests: { cpu: 50m, memory: 64Mi }
      limits: { cpu: 250m, memory: 256Mi }

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.neuro.xxx.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: neuro-tls
      hosts:
        - api.neuro.xxx.com

monitoring:
  enabled: true
  serviceMonitor:
    interval: 15s
    path: /metrics
```

**Per-Service Deployment Template**（以 auth-service 为例）：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "account-center.fullname" . }}-auth-service
  labels:
    app: auth-service
    {{- include "account-center.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.services.auth-service.replicas }}
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: auth-service
          image: "{{ .Values.global.imageRegistry }}/{{ .Values.services.auth-service.image.repository }}:{{ .Values.services.auth-service.image.tag }}"
          ports:
            - containerPort: 8080
          envFrom:
            - configMapRef:
                name: {{ include "account-center.fullname" . }}-auth-service-config
            - secretRef:
                name: {{ include "account-center.fullname" . }}-auth-service-secrets
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 15
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            {{- toYaml .Values.services.auth-service.resources | nindent 12 }}
```

---

### 7.3 CI/CD 流水线设计

**GitHub Actions 全流程**：

```yaml
name: Account Center CI/CD

on:
  push:
    branches: [main, develop]
    tags: ["v*"]
  pull_request:
    branches: [main]

env:
  REGISTRY: registry.cn-hangzhou.aliyuncs.com/neuro
  GO_VERSION: "1.26"

jobs:
  lint:
    name: Lint & Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --config .golangci.yml
      - name: gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec ./...

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: Run tests
        run: |
          go test -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out
      - name: Coverage check
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: ${COVERAGE}%"
          # Threshold: 60%
          if (( $(echo "$COVERAGE < 60" | bc -l) )); then
            echo "Coverage below 60%"
            exit 1
          fi

  integration-test:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - name: Start services
        run: docker compose -f docker-compose.test.yml up -d
      - name: Wait for services
        run: |
          timeout 60 bash -c 'until curl -s http://localhost:8080/health > /dev/null; do sleep 2; done'
      - name: Run integration tests
        run: go test -tags=integration ./tests/integration/...
      - name: Teardown
        if: always()
        run: docker compose -f docker-compose.test.yml down -v

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
      - name: Trivy vulnerability scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: fs
          scan-ref: .
          severity: HIGH,CRITICAL
          exit-code: 1

  build:
    name: Build & Push Images
    runs-on: ubuntu-latest
    needs: [test, integration-test, security-scan]
    if: github.event_name == 'push'
    strategy:
      matrix:
        service:
          - api-gateway
          - auth-service
          - account-service
          - credit-service
          - subscription-service
          - rebate-service
          - notification-service
          - data-product-service
          - config-service
    steps:
      - uses: actions/checkout@v4
      - name: Login to ACR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ secrets.ACR_USERNAME }}
          password: ${{ secrets.ACR_PASSWORD }}
      - name: Build & Push
        uses: docker/build-push-action@v5
        with:
          context: ./services/${{ matrix.service }}
          push: true
          tags: |
            ${{ env.REGISTRY }}/account-center/${{ matrix.service }}:${{ github.sha }}
            ${{ env.REGISTRY }}/account-center/${{ matrix.service }}:latest

  deploy-uat:
    name: Deploy to UAT
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    environment: uat
    steps:
      - name: SSH Deploy
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.UAT_HOST }}
          username: ${{ secrets.UAT_USER }}
          key: ${{ secrets.UAT_SSH_KEY }}
          script: |
            cd /opt/account-center
            docker compose pull
            docker compose up -d
            sleep 10
            for svc in api-gateway auth-service account-service credit-service subscription-service; do
              curl -sf http://localhost:8080/health || exit 1
            done
            echo "UAT deployment successful"

  deploy-prod:
    name: Deploy to Prod (Canary)
    runs-on: ubuntu-latest
    needs: build
    if: startsWith(github.ref, 'refs/tags/v')
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Setup Helm
        uses: azure/setup-helm@v4
      - name: Login to ACR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ secrets.ACR_USERNAME }}
          password: ${{ secrets.ACR_PASSWORD }}
      - name: Canary Deploy (5%)
        run: |
          helm upgrade --install account-center ./helm/account-center \
            --namespace account-center \
            --values ./helm/account-center/values-prod.yaml \
            --set canary.enabled=true \
            --set canary.weight=5 \
            --set image.tag=${{ github.sha }} \
            --wait --timeout 300s
      - name: Monitor Canary (15 min)
        run: |
          sleep 900
          ERROR_RATE=$(curl -s http://prometheus:9090/api/v1/query \
            --data-urlencode "query=rate(http_requests_total{status=~\"5..\",canary=\"true\"}[5m]) / rate(http_requests_total{canary=\"true\"}[5m])" | jq -r '.data.result[0].value[1]')
          if (( $(echo "$ERROR_RATE > 0.005" | bc -l) )); then
            echo "Canary error rate ${ERROR_RATE} exceeds 0.5%, rolling back"
            helm rollback account-center --namespace account-center
            exit 1
          fi
      - name: Promote to 100%
        run: |
          helm upgrade --install account-center ./helm/account-center \
            --namespace account-center \
            --values ./helm/account-center/values-prod.yaml \
            --set canary.enabled=false \
            --set image.tag=${{ github.sha }} \
            --wait --timeout 300s
```

---

### 7.4 发布策略

**金丝雀发布流程（Prod）**：

```
                  ┌──────────────────────────────────────────┐
                  │          K8s Ingress (nginx)              │
                  │                                          │
  用户请求 ────→  │  canary-weight: 5%  ──→ New Pod (v2.0)   │
                  │  canary-weight: 95% ──→ Stable Pod (v1.x)│
                  └──────────────────────────────────────────┘
```

| 阶段 | 流量比例 | 观察时间 | 通过条件 | 失败动作 |
|------|---------|---------|---------|---------|
| A — 金丝雀 | 5% | 15 分钟 | 错误率 < 0.5%，P99 < 2s | 自动回滚 |
| B — 扩大 | 25% | 15 分钟 | 错误率 < 0.5%，P99 < 2s | 自动回滚 |
| C — 过半 | 50% | 15 分钟 | 错误率 < 0.5%，P99 < 2s | 自动回滚 |
| D — 全量 | 100% | 持续 24h | 全部指标正常 | 保留旧版本 24h 可回退 |

**自动回滚触发条件**：
- 金丝雀 Pod 错误率 > 0.5%（持续 2 分钟）
- 金丝雀 Pod P99 延迟 > 2s（持续 2 分钟）
- 金丝雀 Pod CrashLoopBackOff
- 手动触发：运维通过 `helm rollback` 立即回滚

**回滚策略矩阵**：

| 场景 | 触发条件 | 回滚方式 | 预计时间 |
|------|---------|---------|---------|
| 应用层回滚 | 金丝雀阶段错误率 > 0.5% | `helm rollback` / `kubectl rollout undo` | < 2 分钟 |
| 数据库回滚 | Migration 引起数据问题 | `goose down` 回滚至上一版本 | < 5 分钟 |
| 配置回滚 | config-service 发布异常 | config-management-ui 回退到上一版本 | < 1 分钟 |
| 全量回滚 | 严重故障 | DNS 切换旧集群 + PG 时间点恢复 | < 30 分钟 |

**发布窗口**：
- 常规发布：周二至周四 10:00-16:00（业务低峰）
- 紧急修复：经 CTO 审批后随时发布，仍执行金丝雀流程

---

## 8. 可观测性设计

### 8.1 OpenTelemetry 接入方案

**架构总览**：

```
┌────────────┐    OTLP/HTTP    ┌──────────────┐    OTLP    ┌──────────────┐
│ Go Services │ ──────────────→ │ OTel Collector│ ────────→ │ Jaeger/Tempo │
│ (OTel SDK)  │                │ (Sidecar)     │           │ (Backend)    │
└────────────┘                 └──────────────┘            └──────────────┘
      │                              │
      │ Prometheus                   │ OTLP
      ▼                              ▼
┌──────────────┐             ┌──────────────┐
│VictoriaMetrics│             │    Loki      │
└──────────────┘             └──────────────┘
      │                              │
      └──────────┬───────────────────┘
                 ▼
          ┌──────────────┐
          │   Grafana    │
          └──────────────┘
```

**Go SDK 集成**：

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func initTracer(serviceName string) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(context.Background(),
        otlptracegrpc.WithEndpoint("otel-collector:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }
    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(serviceName),
        semconv.ServiceVersionKey.String("2.0.0"),
    )
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

**W3C Trace Context 传播**：
- 所有 HTTP 请求自动注入/提取 `traceparent` / `tracestate` Header
- 服务间调用自动传播 Trace Context（通过 Gin 中间件）
- 跨服务调用 Span 自动关联（Parent-Child 关系）

**采样策略**：

| 场景 | 采样率 | 说明 |
|------|-------|------|
| 默认请求 | 10% | 降低存储成本 |
| 错误响应（5xx） | 100% | 确保所有错误可追踪 |
| 慢请求（P99 > 1s） | 100% | 性能问题全量采集 |
| Admin 操作 | 100% | 审计合规要求 |
| 健康检查 / Metrics | 0% | 排除噪声 |

**业务标签（Business Tags）**：

| 标签 Key | 类型 | 说明 |
|----------|------|------|
| `user_id` | String | 当前操作用户 ID |
| `subscription_plan` | String | 用户订阅等级（L0/L1/L2/L3/L4） |
| `credit_amount` | Int | 积分操作金额（兑换/获取） |
| `device_type` | String | 设备类型（web/ios/android/miniprogram） |
| `api_version` | String | API 版本（v1/v2） |

---

### 8.2 Grafana Dashboard 模板

预置 4 个核心 Dashboard：

**Dashboard 1 — 服务健康总览**：

| Panel | 指标 | 数据源 | 说明 |
|-------|------|--------|------|
| 服务状态 | `up{job="account-center"}` | VictoriaMetrics | 1=正常, 0=宕机 |
| Pod 状态 | `kube_pod_status_phase` | VictoriaMetrics | Running/Pending/CrashLoop |
| CPU 使用率 | `process_cpu_seconds_total` | VictoriaMetrics | 按 service 分组 |
| 内存使用率 | `process_resident_memory_bytes` | VictoriaMetrics | 按 service 分组 |
| Goroutine 数 | `go_goroutines` | VictoriaMetrics | 异常飙升预警 |
| 连接池状态 | `go_sql_open_connections` | VictoriaMetrics | 使用率/空闲/等待 |

**Dashboard 2 — API P95/P99 延迟**：

| Panel | 指标 | PromQL |
|-------|------|--------|
| 全局 P95 | `http_request_duration_seconds` | `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` |
| 全局 P99 | `http_request_duration_seconds` | `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))` |
| 按 Service P95 | 同上 | `by (service)` |
| Top 10 慢接口 | 同上 | `topk(10, ...)` |
| 按等级延迟 | 同上 | `by (subscription_plan)` |

**Dashboard 3 — 错误率**：

| Panel | 指标 | PromQL |
|-------|------|--------|
| 全局错误率 | `http_requests_total` | `rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])` |
| 按 Service 错误率 | 同上 | `by (service)` |
| 4xx 分布 | `http_requests_total` | `by (status)` |
| 熔断器状态 | `circuit_breaker_state` | 0=closed, 1=open, 2=half-open |
| 限流触发 | `rate_limit_hit_total` | `by (service, user_tier)` |

**Dashboard 4 — 业务指标**：

| Panel | 指标 | 说明 |
|-------|------|------|
| 日注册量 | `business_registration_total` | 按渠道分组（web/ios/android/miniprogram） |
| 日活跃用户 | `business_dau` | 去重用户数 |
| 订阅转化率 | `business_subscription_created / business_subscription_trial_started` | 试用→付费转化 |
| 积分兑换量 | `business_credit_exchange_total` | 按兑换类型分组 |
| 推荐转化 | `business_referral_completed / business_referral_invited` | 邀请→注册转化 |
| 营收（GMV） | `business_revenue_total` | 按订阅等级分组 |

---

### 8.3 告警规则定义

**AlertManager 配置**：

```yaml
global:
  resolve_timeout: 5m
  http_config: {}

route:
  group_by: ['alertname', 'service']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 30m
  receiver: 'dingtalk'
  routes:
    - match:
        severity: critical
      receiver: 'dingtalk-critical'
      repeat_interval: 10m
    - match:
        severity: warning
      receiver: 'dingtalk'

receivers:
  - name: 'dingtalk'
    webhook_configs:
      - url: 'http://alertmanager-webhook:8080/dingtalk'
        send_resolved: true
  - name: 'dingtalk-critical'
    webhook_configs:
      - url: 'http://alertmanager-webhook:8080/dingtalk/critical'
        send_resolved: true

inhibit_rules:
  - source_match: { severity: 'critical' }
    target_match: { severity: 'warning' }
    equal: ['alertname', 'service']
```

**Prometheus 告警规则**：

```yaml
groups:
  - name: account-center-service
    rules:
      - alert: ServiceDown
        expr: up{job=~"account-center-.*"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务 {{ $labels.job }} 宕机"
          description: "{{ $labels.instance }} 已离线超过 1 分钟"

      - alert: HighP99Latency
        expr: |
          histogram_quantile(0.99,
            rate(http_request_duration_seconds_bucket{job=~"account-center-.*"}[5m])
          ) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.service }} P99 延迟超过 2s"
          description: "当前 P99: {{ $value }}s"

      - alert: HighErrorRate
        expr: |
          rate(http_requests_total{job=~"account-center-.*", status=~"5.."}[5m])
          / rate(http_requests_total{job=~"account-center-.*"}[5m]) > 0.01
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.service }} 错误率超过 1%"
          description: "当前错误率: {{ $value | humanizePercentage }}"

      - alert: ConnectionPoolExhausted
        expr: |
          go_sql_open_connections{job=~"account-center-.*"}
          / go_sql_max_open_connections{job=~"account-center-.*"} > 0.9
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.service }} 连接池使用率超过 90%"
          description: "当前使用率: {{ $value | humanizePercentage }}"

      - alert: HighGoroutineCount
        expr: go_goroutines{job=~"account-center-.*"} > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.service }} Goroutine 数量异常"
          description: "当前 Goroutine: {{ $value }}"

      - alert: RedisConnectionFailed
        expr: redis_up{job=~"account-center-.*"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis 连接失败"
          description: "{{ $labels.instance }} Redis 不可达"

      - alert: DiskSpaceLow
        expr: |
          (node_filesystem_avail_bytes{mountpoint="/data"}
          / node_filesystem_size_bytes{mountpoint="/data"}) < 0.15
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "磁盘空间不足 15%"
          description: "剩余: {{ $value | humanizePercentage }}"
```

**告警抑制规则**：
- 同一告警 30 分钟内不重复发送（`repeat_interval: 30m`）
- Critical 级别自动抑制同服务 Warning 级别（`inhibit_rules`）
- 非工作时间（22:00-08:00）Warning 级别静默，Critical 级别正常通知

**通知渠道**：

| 级别 | 渠道 | 响应要求 |
|------|------|---------|
| Critical | 钉钉群 @值班人员 + 企业微信 + 电话 | 15 分钟内响应 |
| Warning | 钉钉群通知 | 1 小时内处理 |

---

### 8.4 日志规范

**slog 结构化日志字段规范**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `trace_id` | String | 是 | OpenTelemetry Trace ID，关联分布式追踪 |
| `span_id` | String | 是 | OpenTelemetry Span ID |
| `request_id` | String | 是 | 请求唯一标识（X-Request-ID） |
| `service_name` | String | 是 | 服务名称（如 auth-service） |
| `level` | String | 是 | DEBUG / INFO / WARN / ERROR |
| `timestamp` | String | 是 | ISO 8601 格式（含时区） |
| `msg` | String | 是 | 日志消息 |
| `user_id` | String | 否 | 当前操作用户 ID（若有） |
| `method` | String | 否 | HTTP Method |
| `path` | String | 否 | HTTP Path |
| `status` | Int | 否 | HTTP Status Code |
| `duration_ms` | Float | 否 | 请求耗时（毫秒） |
| `error` | String | 否 | 错误信息（ERROR 级别） |
| `stack_trace` | String | 否 | 错误堆栈（ERROR 级别） |

**JSON 日志输出示例**：

```json
{
    "time": "2026-05-19T10:30:45.123Z",
    "level": "INFO",
    "msg": "user login success",
    "service_name": "auth-service",
    "trace_id": "abc123def456",
    "span_id": "789ghi012",
    "request_id": "req-uuid-001",
    "user_id": "usr-uuid-002",
    "method": "POST",
    "path": "/api/v1/auth/login",
    "status": 200,
    "duration_ms": 45.6,
    "remote_ip": "192.168.1.100",
    "user_agent": "Neuro/2.0 (iOS 18.4)"
}
```

**Go slog 初始化**：

```go
import (
    "log/slog"
    "slog-multi"
)

func initLogger(serviceName string) *slog.Logger {
    handler := slogmulti.Pipe(
        slogtrace.TracerHandler(),
        slgrequest.RequestIDHandler("request_id"),
    ).Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    return slog.New(handler).With("service_name", serviceName)
}
```

**Loki 集成架构**：

```
┌────────────┐   stdout/stderr   ┌────────────┐    HTTP     ┌────────────┐
│ Go Service │ ────────────────→ │  Promtail  │ ──────────→ │    Loki    │
│ (JSON log) │                  │ (DaemonSet)│             │  (Cluster) │
└────────────┘                  └────────────┘             └────────────┘
                                                                │
                                                                ▼
                                                         ┌────────────┐
                                                         │  Grafana   │
                                                         │ LogQL 查询  │
                                                         └────────────┘
```

**Promtail 配置**：

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: account-center
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          - name: label
            values: ["com.docker.compose.project=account-center"]
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        target_label: 'container'
      - source_labels: ['__meta_docker_container_log_stream']
        target_label: 'stream'
    pipeline_stages:
      - json:
          expressions:
            level: level
            service_name: service_name
            trace_id: trace_id
            request_id: request_id
      - labels:
          level:
          service_name:
      - timestamp:
          source: time
          format: RFC3339
```

**日志查询示例（LogQL）**：

```logql
{service_name="auth-service"} |= "error" | json | level="ERROR"
{service_name="account-service"} | json | trace_id="abc123" | line_format "{{.time}} [{{.level}}] {{.msg}}"
{service_name=~".*"} | json | status >= 500 | count by (service_name)
```

---

## 9. 附录

### 9.1 与 SSD V1.3.1 差异对照表

| V1.3.1 章节 | V2.0 章节 | 变更类型 | 变更说明 |
|------------|----------|---------|---------|
| 密码哈希（SM3） | 6.1 密码哈希迁移 | 升级 | SM3+salt → argon2id，保留 SM3 用于审计链；新增透明 rehash 机制 |
| 密钥管理（环境变量派生） | 6.2 KMS/Vault 集成 | 新增 | 引入 Vault/阿里云 KMS，支持密钥自动轮换、审计追踪、紧急吊销 |
| API 限流（IP 级） | 6.3 API 安全加固 | 升级 | IP 级限流 → 用户级差异化限流（按等级 L0/L1/L2+）；新增 HMAC-SHA256 签名 |
| 无移动端安全章节 | 6.4 移动端安全 | 新增 | Certificate Pinning、Root/越狱检测、截屏防护、Token 安全存储 |
| Docker Compose 部署 | 7.1 多环境差异 | 扩展 | 新增 K8s Helm Chart 生产部署；明确 Dev/UAT/Prod 三环境差异矩阵 |
| 无 Helm Chart | 7.2 Helm Chart 结构 | 新增 | 完整 Helm Chart 结构，9 个服务独立 Deployment/Service/HPA 模板 |
| 无 CI/CD | 7.3 CI/CD 流水线 | 新增 | GitHub Actions 全流程：lint → test → security → build → deploy |
| 无发布策略 | 7.4 发布策略 | 新增 | 金丝雀发布（5%→25%→50%→100%），自动回滚机制 |
| 基础 Prometheus 指标 | 8.1 OpenTelemetry | 升级 | 新增分布式追踪，W3C Trace Context，分级采样策略 |
| 无 Grafana Dashboard | 8.2 Dashboard 模板 | 新增 | 4 个预置 Dashboard：服务健康、API 延迟、错误率、业务指标 |
| 无告警规则 | 8.3 告警规则 | 新增 | AlertManager + Prometheus 规则 + 钉钉/企微通知 |
| JSON 日志（无 trace_id） | 8.4 日志规范 | 升级 | 统一 slog 字段（trace_id/span_id/request_id），Loki + Promtail 集成 |
| 8 微服务 | 3.x 服务详细设计 | 扩展 | 新增 Phase 6-9 共 16 项功能的技术设计 |
| 无数据设计 | 4. 数据设计 | 新增 | 数据库 Schema 变更、索引策略、数据迁移方案 |
| 无 API 设计 | 5. API 设计 | 新增 | V2 API 规范、版本路由、契约测试 |

### 9.2 配置项变更清单

| 配置 Key | 默认值 | 环境差异 | 可热加载 | 说明 |
|----------|--------|---------|---------|------|
| `PASSWORD_HASHER` | `argon2id` | Dev/UAT/Prod 一致 | 否 | 密码哈希算法（argon2id / sm3） |
| `ARGON2ID_MEMORY_KB` | `65536` | 一致 | 否 | argon2id 内存参数（64MB） |
| `ARGON2ID_ITERATIONS` | `3` | 一致 | 否 | argon2id 迭代次数 |
| `ARGON2ID_PARALLELISM` | `2` | 一致 | 否 | argon2id 并行度 |
| `RATE_LIMIT_L0_PER_MIN` | `60` | 一致 | 是 | L0 用户限流阈值 |
| `RATE_LIMIT_L1_PER_MIN` | `120` | 一致 | 是 | L1 用户限流阈值 |
| `RATE_LIMIT_L2_PER_MIN` | `300` | 一致 | 是 | L2+ 用户限流阈值 |
| `KMS_PROVIDER` | `vault` | Dev: vault-dev / UAT: vault / Prod: aliyun-kms | 否 | 密钥管理服务提供方 |
| `KMS_KEY_ROTATION_DAYS` | `90` | 一致 | 否 | 密钥轮换周期（天） |
| `HMAC_SIGNING_ENABLED` | `true` | Dev: false / UAT/Prod: true | 是 | HMAC 请求签名开关 |
| `OTEL_SAMPLING_RATE` | `0.1` | Dev: 1.0 / UAT: 0.5 / Prod: 0.1 | 是 | OpenTelemetry 采样率 |
| `OTEL_EXPORTER_ENDPOINT` | `localhost:4317` | Dev: localhost / UAT/Prod: otel-collector | 否 | OTel Collector 地址 |
| `LOG_LEVEL` | `info` | Dev: debug / UAT: info / Prod: warn | 是 | 日志级别 |
| `LOG_FORMAT` | `json` | 一致 | 否 | 日志格式（json / text） |
| `DB_MAX_OPEN_CONNS` | `25` | Dev: 5 / UAT: 15 / Prod: 25 | 否 | PG 最大连接数 |
| `DB_MAX_IDLE_CONNS` | `10` | Dev: 2 / UAT: 5 / Prod: 10 | 否 | PG 最大空闲连接 |
| `REDIS_ADDR` | `localhost:6379` | 各环境不同 | 否 | Redis 地址 |
| `REDIS_SENTINEL_ENABLED` | `false` | Dev/UAT: false / Prod: true | 否 | Redis Sentinel 模式开关 |
| `JWT_ACCESS_TTL_MINUTES` | `15` | 一致 | 是 | Access Token 有效期（分钟） |
| `JWT_REFRESH_TTL_DAYS` | `30` | 一致 | 是 | Refresh Token 有效期（天） |
| `CERT_PINNING_ENABLED` | `false` | Dev: false / UAT: true / Prod: true | 否 | Certificate Pinning 开关 |
| `ROOT_DETECTION_ENABLED` | `false` | Dev: false / UAT: true / Prod: true | 否 | Root/越狱检测开关 |
| `CANARY_WEIGHT` | `0` | Dev/UAT: 0 / Prod: 动态 | 是 | 金丝雀发布流量权重 |
| `HPA_MIN_REPLICAS` | `1` | Dev: 1 / UAT: 1 / Prod: 2 | 否 | HPA 最小副本数 |
| `HPA_MAX_REPLICAS` | `3` | Dev: 3 / UAT: 3 / Prod: 10 | 否 | HPA 最大副本数 |
