# Account Center V2.0 系统规格文档

> **文档类型**: 系统规格文档（正式版）
> **版本**: V2.0.0
> **日期**: 2026-05-19
> **状态**: 编制中
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

（待填充）

## 4. 数据设计

（待填充）

## 5. API 设计

（待填充）

## 6. 安全设计

（待填充）

## 7. 部署设计

（待填充）

## 8. 可观测性设计

（待填充）

## 9. 附录

（待填充）
