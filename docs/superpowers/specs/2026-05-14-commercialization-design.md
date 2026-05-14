# 账户管理微服务商业化迭代 — 设计规格文档

> **版本**: v1.0.0
> **日期**: 2026-05-14
> **状态**: Draft
> **基于**: BRD V1.0.0, PRD V1.3.0, TIP V1.3.0, SOP L1 v2.2.0, SOP L5 v1.1.0r3

---

## 1. 目标

在现有 11 个基础微服务之上，引入五级身份阶梯、奖励积分账务、阶梯退坡推广返利、全链路防刷风控等商业化能力，同时将服务数量从 11+5=16 合并为 7 个（6 服务 + 1 网关），降低运维复杂度。

## 2. 约束

| 约束 | 来源 | 说明 |
|------|------|------|
| 技术栈 | SOP L1 §1.2.1 | Go/Gin，备选 Java 需 ADR 审批 |
| 端口范围 | SOP L1 §1.1 | 30000-50000，w004 分配 30300-30399 |
| 国密算法 | SOP L1 §1.2.5 | SM2/SM3/SM4 全环境一致 |
| 监控方案 | SOP L1 §1.2.7 | VictoriaMetrics，禁用 Prometheus/ELK/Jaeger |
| 消息队列 | SOP L1 §1.2.4 | 开发/测试: Redis Streams + Asynq |
| DB 迁移 | SOP L1 §1.2.3 | Goose 或 golang-migrate，禁止手动改库 |
| 容器数上限 | 产品决策 | ≤8 个容器（含网关） |
| 开发环境 | SOP L5-prep | Docker Compose 全套自建基础设施 |
| 资源登记 | SOP L5 §4.0 | 新端口/路由/容器名需在附录登记 |
| 非root运行 | SOP L5 Appendix H | Dockerfile 使用 `USER nobody` |
| 多阶段构建 | SOP L5 Appendix H | Go multi-stage build |
| 健康检查 | SOP L5 Appendix H | `wget --spider -q` HEAD 方法 |

## 3. 服务架构（合并后）

### 3.1 服务清单

| # | 服务名 | 端口 | 来源 | 职责 |
|---|--------|------|------|------|
| 1 | **api-gateway** | 30300 | 保持不变 | 反向代理、JWT 认证、限流、动态脱敏 |
| 2 | **auth-service** | 30302 | ← auth-service, session-service, device-fingerprint-service | 注册/登录/JWT/会话管理/设备指纹/扫码登录/生物识别 |
| 3 | **notification-service** | 30311 | ← sms-email-service, email-service, push-notification-service | 短信(阿里云/腾讯云/天翼云)/邮件(SMTP)/推送(APNs/FCM等7平台) |
| 4 | **account-service** | 30301 | ← account-service, 新增 entitlement/subscription | 账户管理/身份等级(L0-L4)/权益配额/订阅生命周期 |
| 5 | **credit-service** | 30312 | 全新 | 奖励积分账务(复式记账+SM3链)/推广关系/阶梯退坡返利 |
| 6 | **compliance-service** | 30313 | ← risk-detection-service, audit-log-service, kyb-service | 风控(防刷/黑名单/滑动窗口)/审计日志(SM3完整性)/KYB企业认证 |
| 7 | **data-product-service** | 30314 | 全新 | RFM 用户画像(去标识化)/推广防刷监控大盘/订阅漏斗 |

### 3.2 删除的服务目录

合并后删除以下目录（代码已迁移到对应合并服务中）：
- `session-service/` → auth-service
- `device-fingerprint-service/` → auth-service
- `sms-email-service/` → notification-service
- `email-service/` → notification-service
- `push-notification-service/` → notification-service
- `risk-detection-service/` → compliance-service
- `audit-log-service/` → compliance-service
- `kyb-service/` → compliance-service

### 3.3 新增的服务目录

- `notification-service/` (端口 30311)
- `credit-service/` (端口 30312)
- `compliance-service/` (端口 30313)
- `data-product-service/` (端口 30314)

## 4. 数据库设计

### 4.1 新增表

#### subscriptions（订阅表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 主键 |
| user_id | BIGINT FK→users | 用户 ID |
| tier_level | INT CHECK(2,3,4) | 订阅等级 L2/L3/L4 |
| start_time | TIMESTAMPTZ | 开始时间 |
| end_time | TIMESTAMPTZ | 结束时间 |
| status | VARCHAR(20) | ACTIVE/EXPIRED/CANCELED |
| price | DECIMAL(10,2) | 订阅价格 |
| payment_method | VARCHAR(50) | 支付方式 |
| order_id | VARCHAR(100) UNIQUE | 订单 ID（幂等） |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

#### entitlements（权益表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 主键 |
| user_id | BIGINT FK→users | 用户 ID |
| feature_code | VARCHAR(100) | 功能码 |
| total_quota | INT | 总配额 |
| used_quota | INT | 已用配额 |
| reset_time | TIMESTAMPTZ | 重置时间 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |
| UNIQUE(user_id, feature_code) | | |

#### credit_accounts（积分账户表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 主键 |
| user_id | BIGINT FK→users UNIQUE | 用户 ID |
| balance | DECIMAL(12,2) CHECK(>=0) | 余额 |
| status | VARCHAR(20) | ACTIVE/FROZEN |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

#### credit_transactions（积分交易流水表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 主键 |
| credit_account_id | BIGINT FK→credit_accounts | 积分账户 ID |
| type | VARCHAR(50) | EARN_REFERRAL/EARN_VERIFY/CONSUME_SUB/REFUND_SUB/EXPIRED |
| amount | DECIMAL(12,2) | 金额 |
| reference_id | VARCHAR(100) | 关联业务 ID（幂等） |
| details | JSONB | 详情 |
| sm3_hash | VARCHAR(128) | SM3 防篡改摘要 |
| status | VARCHAR(20) | AVAILABLE/PENDING/FROZEN/CONSUMED/EXPIRED/REJECTED |
| created_at | TIMESTAMPTZ | 创建时间 |

#### referral_relations（推广关系表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 主键 |
| referrer_id | BIGINT FK→users | 推广人 |
| referee_id | BIGINT FK→users UNIQUE | 被推广人（唯一） |
| referee_subscription_count | INT DEFAULT 0 | 被推广人订阅次数（退坡计算） |
| status | VARCHAR(20) | ACTIVE/FROZEN |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

### 4.2 修改表

**users 表新增字段：**
- `identity_tier INT NOT NULL DEFAULT 0` — 身份等级 (0-4)
- `status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'` — 用户状态

### 4.3 新增表（设备指纹持久化）

#### device_fingerprints
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 主键 |
| user_id | BIGINT FK→users | 用户 ID |
| device_id | VARCHAR(100) | 设备 ID |
| fingerprint_hash | VARCHAR(128) UNIQUE | 指纹哈希 |
| device_info | TEXT | 设备信息 |
| ip_address | VARCHAR(45) | IP 地址 |
| last_login_at | TIMESTAMPTZ | 最后登录时间 |
| trusted_until | TIMESTAMPTZ | 信任过期时间 |
| is_trusted | BOOLEAN DEFAULT FALSE | 是否信任 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

## 5. 核心业务逻辑

### 5.1 五级身份阶梯（account-service）

- **L0 注册用户**: 注册即为此级，仅可浏览公开内容
- **L1 实名用户**: 完成手机号实名认证后自动升级
- **L2/L3/L4 订阅用户**: 购买对应等级订阅后升级，到期自动回退至 L1 或 L0

等级变更路径：
```
注册 → L0
实名认证 → L1 (account-service 调用 identity_tier 更新)
购买订阅 T1 → L2 (account-service entitlement 模块处理)
购买订阅 T2 → L3
购买订阅 T3 → L4
订阅到期 → 回退至 L1(已实名) 或 L0(未实名)
```

### 5.2 权益中控（account-service）

- Redis Hash 缓存用户权益：`entitlement:{user_id}` → `{feature_code: {total, used, reset}}`
- Lua 脚本原子扣减：校验余额 → 扣减 → 返回结果
- 异步落库：扣减成功后发 Redis Streams 事件，消费端批量更新 PostgreSQL
- Asynq 定时任务重置配额

### 5.3 奖励积分账务（credit-service）

- **复式记账**：每笔变动生成 credit_transactions 流水
- **SM3 摘要链**：每条流水 = SM3(id + account_id + type + amount + ref_id + created_at + prev_hash)
- **乐观锁**：`UPDATE credit_accounts SET balance = balance + ? WHERE id = ? AND balance + ? >= 0`
- **幂等**：reference_id 唯一索引防重
- **Saga 补偿**：订阅失败时退回已扣积分

### 5.4 阶梯退坡返利（credit-service）

监听 Redis Streams 的 `subscription:paid` 事件：
```
被推广用户第 N 次订阅 → 查询推广关系 → 匹配退坡比例:
  N=0: 50%, N=1-4: 30%, N=5-9: 20%, N≥10: 10%
→ 计算奖励积分 → 风控评估 → 低风险立即发放/高风险延迟 → 更新推广关系订阅次数
```

### 5.5 防刷风控（compliance-service）

- Redis 滑动窗口限流：单 IP/设备 1 小时注册上限 3 次
- 推广防刷：单推广链接 1 小时被使用上限 50 次
- 黑名单：Redis Set + PostgreSQL 持久化
- T+7 延迟发放：大额积分发放后 7 天检查订单状态
- 轻量规则引擎（Go `expr` 库）

### 5.6 动态脱敏（api-gateway）

响应拦截器，根据请求 JWT 角色掩码：
- 手机号: `138****1234`
- 邮箱: `t***@example.com`
- 真实姓名: `张*`
- 管理员角色不脱敏

### 5.7 扫码登录（auth-service）

- Redis 存储二维码状态 `qrcode:{code_id}`，TTL 5 分钟
- Web 端轮询状态，移动端扫码确认后 Web 端获取 JWT

## 6. API 路由规划

### api-gateway 代理路由

| 路由前缀 | 目标服务:端口 |
|---------|-------------|
| `/api/v1/account/*` | account-service:30301 |
| `/api/v1/auth/*` | auth-service:30302 |
| `/api/v1/notification/*` | notification-service:30311 |
| `/api/v1/sms/*` | notification-service:30311 |
| `/api/v1/email/*` | notification-service:30311 |
| `/api/v1/push/*` | notification-service:30311 |
| `/api/v1/device/*` | auth-service:30302 |
| `/api/v1/session/*` | auth-service:30302 |
| `/api/v1/credits/*` | credit-service:30312 |
| `/api/v1/referral/*` | credit-service:30312 |
| `/api/v1/risk/*` | compliance-service:30313 |
| `/api/v1/audit/*` | compliance-service:30313 |
| `/api/v1/kyb/*` | compliance-service:30313 |
| `/api/v1/entitlements/*` | account-service:30301 |
| `/api/v1/subscriptions/*` | account-service:30301 |
| `/api/v1/data/*` | data-product-service:30314 |
| `/api/v1/qrcode/*` | auth-service:30302 |

### 公开路径（无需 JWT）

- `/api/v1/auth/login`, `/api/v1/auth/refresh`
- `/api/v1/account/register`
- `/api/v1/sms/send`, `/api/v1/sms/verify`
- `/api/v1/email/otp/send`, `/api/v1/email/magic-link/*`
- `/api/v1/qrcode/generate`, `/api/v1/qrcode/:code_id/status`
- `/health` (所有服务)

## 7. Docker Compose 基础设施

开发/测试环境全部自建：

| 服务 | 镜像 | 端口映射 | 说明 |
|------|------|---------|------|
| postgres | postgres:18-alpine | 5432 | 关系数据库 |
| redis | redis:7-alpine | 6379 | 缓存/消息队列 |
| minio | minio/minio | 9000/9001 | 对象存储 |
| db-migrate | 自建 | 无 | Goose 迁移 |
| api-gateway | 自建 Go | 30300:30300 | 网关 |
| auth-service | 自建 Go | 30302:30302 | 认证 |
| notification-service | 自建 Go | 30311:30311 | 通知 |
| account-service | 自建 Go | 30301:30301 | 账户 |
| credit-service | 自建 Go | 30312:30312 | 积分 |
| compliance-service | 自建 Go | 30313:30313 | 合规 |
| data-product-service | 自建 Go | 30314:30314 | 数据产品 |

## 8. 安全设计

- JWT Access Token 30min + Refresh Token 7天 + Redis 黑名单实时阻断
- 密码哈希: SM3 + salt
- 敏感数据存储加密: SM4 CBC 模式
- 积分流水完整性: SM3 摘要链
- 审计日志完整性: SM3 校验
- 密钥管理: .env.secrets (开发环境) / Vault KMS (生产环境)
- 所有容器 USER nobody 运行

## 9. 监控设计

- VictoriaMetrics (端口 20010) 采集指标
- OpenTelemetry 纯埋点 + TraceID 关联日志
- 各服务暴露 `/metrics` 端点供 VictoriaMetrics 抓取
- 业务指标: 注册成功率、订阅转化率、积分发放量、风控拦截次数、权益核销次数

## 10. Dockerfile 模板（遵循 SOP L5 Appendix H）

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE <PORT>
USER nobody
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:<PORT>/health || exit 1
ENTRYPOINT ["./main"]
```

## 11. 验收标准（对应 TIP §12）

1. 所有 7 个服务编译通过 (`go build`, `go vet`)
2. `docker compose up -d` 一键启动，所有容器 healthy
3. 权益核销: Redis Lua 原子扣减，无超卖
4. 积分账务: 余额与流水总和 100% 一致，SM3 链校验无异常
5. 阶梯退坡: 15 次连续订阅准确按 50%-30%-20%-10% 返利
6. 风控: 同 IP 连续注册 10 账号，第 4 个触发拦截
7. 脱敏: 非管理员角色查询返回掩码数据
8. VictoriaMetrics 采集所有服务指标
9. docker-compose.yml 中无明文密码

## 12. 实施阶段

### Sprint 1: 服务合并 + 商业化核心
- DB 迁移（新增 5 张表 + users 表修改）
- 合并 session/device → auth-service
- 合并 sms/email/push → notification-service
- 合并 risk/audit/kyb → compliance-service
- account-service 扩展（身份等级/权益/订阅）
- 新建 credit-service（积分账务/推广）
- api-gateway 更新路由

### Sprint 2: 推广返利 + 风控强化
- 阶梯退坡返利引擎（credit-service 消费 SubscriptionPaidEvent）
- 防刷风控强化（compliance-service: 滑动窗口/黑名单/T+7/规则引擎）
- 推广注册联动（account-service → credit-service）

### Sprint 3: 数据产品 + 辅助功能
- data-product-service（RFM 画像/防刷大盘/订阅漏斗）
- 扫码登录（auth-service qrcode 模块）
- 动态脱敏（api-gateway 中间件）
- VictoriaMetrics /metrics 端点集成
- 最终集成测试与部署
