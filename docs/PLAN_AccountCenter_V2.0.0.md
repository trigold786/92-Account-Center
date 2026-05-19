# Account Center V2.0 推进计划

> **文档类型**: 推进计划（正式版）
> **版本**: V2.0.0
> **日期**: 2026-05-19
> **状态**: 已完成
> **关联文档**: PRD V2.0.0（86 项需求）、SSD V2.0.0（技术设计）
> **评估基准**: ITERATION_AccountCenter_V2.0_V1.0.1（77 项改进建议）
> **变更历史**:
> | 版本 | 日期 | 变更内容 | 作者 |
> |------|------|---------|------|
> | V2.0.0 | 2026-05-19 | 初始编制，覆盖全部 86 项需求的任务分解 | |

---

## 目录

1. 总览
2. Phase 6 — P0 任务分解（14 项）
3. Phase 7 — P1 任务分解（32 项）
4. Phase 8 — P2 任务分解（32 项）
5. Phase 9 — P3 任务分解（8 项）
6. 风险管理
7. 附录

---

## 1. 总览

### 1.1 迭代目标与范围

本计划覆盖 PRD V2.0.0 定义的 86 项需求的完整任务分解，源自 ITERATION V1.0.1 的 77 项改进建议（部分项拆分为多条需求）。按 Phase 6/7/8/9 分阶段交付，每个 Phase 有独立版本号和阶段门禁。

| Phase | 版本 | 优先级 | 需求数 | 定位 |
|-------|------|--------|--------|------|
| Phase 6 | V2.0 | P0 | 14 | 上线准备冲刺（阻塞上线） |
| Phase 7 | V2.1 | P1 | 32 | 体验与增长（上线后首月） |
| Phase 8 | V2.2 | P2 | 32 | 竞争力提升（上线后 1-3 月） |
| Phase 9 | V2.3 | P3 | 8 | 长期规划（3+ 月后） |

### 1.2 团队配置

1 人全职，串行执行。不估算工期，按依赖顺序交付。

### 1.3 版本发布策略

- **V2.0** = Phase 6（P0 全部完成）→ Git tag `v2.0.0`
- **V2.1** = Phase 7（P1 全部完成）→ Git tag `v2.1.0`
- **V2.2** = Phase 8（P2 全部完成）→ Git tag `v2.2.0`
- **V2.3** = Phase 9（P3 全部完成）→ Git tag `v2.3.0`

### 1.4 执行原则

1. **依赖前置**：被依赖的任务必须先完成
2. **P0 优先**：Phase 6 全部完成后才进入 Phase 7
3. **风险先行**：同一 Phase 内高风险任务先做
4. **即时验证**：完成即测试，测试通过即标注

### 1.5 阶段门禁

每个 Phase 完成前必须通过以下检查：

- [ ] 该 Phase 全部需求对应的单元测试 + 集成测试通过
- [ ] PRD 中该 Phase 需求状态标注为"已实现"
- [ ] SSD 中对应技术设计章节标注实现状态
- [ ] CHANGELOG 记录本 Phase 全部变更
- [ ] Git tag 打版本标签
- [ ] 备份策略执行并验证（AR-23 贯穿全程）

---

## 2. Phase 6 — P0 任务分解（14 项）

> 阻塞上线的 14 项需求，按依赖顺序排列。

### 2.1 执行顺序总表

| 序号 | 需求 ID | 名称 | 维度 | 依赖 | 涉及服务 |
|------|---------|------|------|------|---------|
| 1 | AR-25 | 清理仓库 | 质量 | 无 | 全局 |
| 2 | NF-02 | 网关请求超时 | 可靠性 | 无 | api-gateway |
| 3 | NF-01 | 账号注销 Worker | 合规 | 无 | account-service |
| 4 | AR-13 | 密码哈希 argon2id | 安全 | 无 | auth-service |
| 5 | AR-23 | 数据库备份策略 | 运维 | 无 | 基础设施 |
| 6 | AR-17 | 单元测试补齐 | 质量 | 无 | 3 核心服务 |
| 7 | FN-02 | 订单管理系统 | 商业化 | 无 | payment-service(NEW) |
| 8 | FN-01 | 支付网关集成 | 商业化 | FN-02 | payment-service |
| 9 | FN-05 | 用户管理后台 | 运营 | FN-02 | account-service |
| 10 | FN-10 | APNs/FCM 推送 | 功能 | 无 | notification-service |
| 11 | AR-21 | K8s Helm Chart | 运维 | 无 | 全服务 |
| 12 | UX-08 | 定价透明度 | UX | 无 | 全端 |
| 13 | AR-18 | 集成测试 | 质量 | AR-17 | 全服务 |
| 14 | AR-16 | 渗透测试 | 安全 | AR-13, AR-17 | 全服务 |

### 2.2 任务详情

#### Task 6.1: AR-25 — 清理仓库

**描述**: 删除 main.exe、cmd.exe、nul 等二进制残留文件，更新 .gitignore，修正 README/ARCHITECTURE.md 与实际代码结构的不一致。

**涉及文件**:
- `.gitignore` — 添加 `*.exe`、`nul` 等规则
- `README.md` — 修正项目结构描述
- `ARCHITECTURE.md` — 修正端口和服务描述

**验证方式**: `git ls-files | grep -E '\.exe$'` 返回空；文档端口号与 docker-compose.yml 一致

---

#### Task 6.2: NF-02 — 网关请求超时

**描述**: 为 api-gateway 的 httputil.ReverseProxy 配置自定义 Transport（ResponseHeaderTimeout=30s, IdleConnTimeout=90s），添加全局 60s 请求超时中间件。

**涉及文件**:
- `api-gateway/internal/proxy/transport.go` — 自定义 Transport
- `api-gateway/internal/middleware/timeout.go` — 超时中间件
- `api-gateway/cmd/main.go` — 注册中间件

**验证方式**: 模拟慢后端，验证 504 返回；wrk 压测确认连接池不耗尽

---

#### Task 6.3: NF-01 — 账号注销 Worker

**描述**: 实现 Asynq 定时任务 deletion-worker，冻结期到期后自动匿名化 PII 数据、清理 Redis session/cache、写入审计日志。

**涉及文件**:
- `account-service/internal/worker/deletion.go` — Asynq task handler
- `account-service/internal/repository/user.go` — AnonymizeUser 方法
- `account-service/internal/service/deletion.go` — 注销业务逻辑编排

**验证方式**: 集成测试完整执行冻结期→匿名化→验证 PII 替换；单元测试覆盖率 >80%

---

#### Task 6.4: AR-13 — 密码哈希 argon2id

**描述**: auth-service 新增 argon2id 哈希（memory=64MB, iterations=3, parallelism=2），新注册直接使用，存量 SM3 用户登录时透明 rehash。password_hash 字段新增前缀标识。

**涉及文件**:
- `auth-service/internal/auth/argon2id.go` — argon2id 哈希和验证
- `auth-service/internal/auth/hash_factory.go` — 哈希策略工厂
- `auth-service/internal/service/auth.go` — 登录逻辑添加 rehash 分支

**验证方式**: SM3 用户登录后确认 rehash 为 argon2id；兼容性测试 SM3/argon2id 双验证

---

#### Task 6.5: AR-23 — 数据库备份策略

**描述**: PostgreSQL 每日 pg_dump + WAL 归档；Redis RDB + AOF 混合持久化 + 定期备份至 OSS。完成一次恢复演练。

**涉及文件**:
- `infra/backup/pg_backup.sh` — PG 备份脚本
- `infra/backup/redis_backup.sh` — Redis 备份脚本
- `infra/backup/restore_test.sh` — 恢复演练脚本

**验证方式**: 恢复演练报告（PG 全量恢复 + PITR，Redis RDB 恢复）

---

#### Task 6.6: AR-17 — 单元测试补齐

**描述**: 为 credit-service、subscription-service、rebate_service 补齐单元测试，目标覆盖率 >60%。

**涉及文件**:
- `credit-service/internal/service/credit_test.go`
- `credit-service/internal/service/rfm_test.go`
- `account-service/internal/service/subscription_test.go`
- `account-service/internal/service/rebate_test.go`
- `account-service/internal/service/level_test.go`

**验证方式**: `go test -coverprofile=coverage.out ./...` 各服务 >60%

---

#### Task 6.7: FN-02 — 订单管理系统

**描述**: 新建 payment-service（端口 30316），实现 orders 表和订单状态机（pending→paid→cancelled→refunded），支持多维度查询和 CSV/Excel 导出。

**涉及文件**:
- `payment-service/cmd/main.go` — 服务入口（新建）
- `payment-service/internal/model/order.go` — 订单模型
- `payment-service/internal/service/order.go` — 状态机业务逻辑
- `payment-service/internal/repository/order.go` — 数据层
- `payment-service/internal/handler/order.go` — HTTP handler

**验证方式**: 状态机所有合法/非法跳转测试；订单 CRUD + 导出功能验证

---

#### Task 6.8: FN-01 — 支付网关集成

**描述**: 在 payment-service 中集成微信支付（H5/小程序/Native）和支付宝（手机网站/APP），实现 Provider 接口、异步回调处理、对账和重试机制。**依赖 FN-02。**

**涉及文件**:
- `payment-service/internal/provider/provider.go` — 统一接口
- `payment-service/internal/service/wechat_pay.go` — 微信支付
- `payment-service/internal/service/alipay.go` — 支付宝
- `payment-service/internal/handler/callback.go` — 回调处理
- `payment-service/internal/service/reconciliation.go` — 对账

**验证方式**: 微信/支付宝沙箱环境完整支付流程；回调幂等性验证

---

#### Task 6.9: FN-05 — 用户管理后台

**描述**: account-service 新增 Admin API 模块，实现用户列表/搜索/详情、等级调整、积分调整、封禁/解封、实名审核。**依赖 FN-02。**

**涉及文件**:
- `account-service/internal/handler/admin.go` — Admin API
- `account-service/internal/service/admin.go` — 业务逻辑
- `account-service/internal/middleware/admin_auth.go` — 鉴权中间件

**验证方式**: 非 Admin 角色 JWT 无法访问 Admin API；所有操作写入审计日志

---

#### Task 6.10: FN-10 — APNs/FCM 推送集成

**描述**: 扩展 notification-service 的 provider 架构，新增 APNs（HTTP/2）、FCM、华为 HMS 推送 provider，实现设备 Token 管理和推送日志。

**涉及文件**:
- `notification-service/internal/provider/push.go` — Push Provider 接口
- `notification-service/internal/provider/apns.go` — APNs 实现
- `notification-service/internal/provider/fcm.go` — FCM 实现
- `notification-service/internal/provider/hms.go` — HMS 实现
- `notification-service/internal/handler/device.go` — Token 管理

**验证方式**: APNs/FCM 沙箱环境真实推送到达；Token 注册/注销流程

---

#### Task 6.11: AR-21 — K8s Helm Chart + CI/CD

**描述**: 编写 Helm Chart 覆盖 9 个微服务（含 payment-service），配置 HPA 和滚动更新；搭建 GitHub Actions CI/CD 流水线。

**涉及文件**:
- `helm/account-center/Chart.yaml`
- `helm/account-center/values.yaml`
- `helm/account-center/templates/` — 各服务 Deployment/Service/ConfigMap/HPA
- `.github/workflows/ci.yml` — CI/CD 流水线
- `Makefile` — 统一构建入口

**验证方式**: `helm lint` 通过；Dev 环境实际部署验证；CI 流水线跑通

---

#### Task 6.12: UX-08 — 定价透明度

**描述**: config-service 配置定价数据，四端实现定价页面（等级卡片、权益对比矩阵、积分抵扣计算器）。

**涉及文件**:
- `account-service/internal/handler/pricing.go` — 定价 API
- `web/src/views/PricingPage.vue` — Web 定价页
- `ios/AccountCenter/Views/PricingView.swift` — iOS
- `android/.../PricingScreen.kt` — Android

**验证方式**: 四端定价信息展示一致；积分抵扣计算器实时计算正确

---

#### Task 6.13: AR-18 — 集成测试

**描述**: Docker Compose 搭建完整测试环境，实现全链路集成测试：注册→登录→订阅购买→积分→推荐→过期降级。**依赖 AR-17。**

**涉及文件**:
- `tests/integration/full_journey_test.go`
- `tests/integration/helpers.go`
- `tests/integration/docker-compose.test.yml`

**验证方式**: 全链路测试用例全部通过；集成到 CI 自动执行

---

#### Task 6.14: AR-16 — 第三方渗透测试

**描述**: 外部安全团队完成 OWASP Top 10 渗透测试、依赖库漏洞扫描（Trivy+Snyk）、移动端安全审计。**依赖 AR-13、AR-17。**

**涉及文件**:
- `scripts/security/scan_dependencies.sh`
- `scripts/security/generate_test_accounts.go`
- `scripts/security/zap_scan_config.json`

**验证方式**: 渗透测试报告无 Critical/High 未修复漏洞；依赖扫描无 Critical 漏洞

---

### 2.3 Phase 6 阶段产出物清单

```
Phase 6 (V2.0) 产出物：
├── 文档
│   ├── PRD V2.0.0（Phase 6 需求标注"已实现"）
│   ├── SSD V2.0.0（Phase 6 技术设计标注实现状态）
│   └── CHANGELOG_V2.0.md
├── 代码
│   ├── account-service: deletion-worker, argon2id migration, Admin API
│   ├── api-gateway: timeout middleware, refactor
│   ├── payment-service: 新增微服务（订单+支付+回调+对账）
│   ├── notification-service: APNs/FCM/HMS provider
│   ├── auth-service: argon2id 哈希迁移
│   └── .gitignore 更新
├── 基础设施
│   ├── helm/（Helm Chart 9 服务）
│   ├── .github/workflows/（CI/CD）
│   ├── infra/backup/（备份脚本）
│   └── monitoring/（Dashboard + 告警规则基础）
├── 测试
│   ├── 单元测试 >60%（credit/subscription/rebate）
│   ├── 集成测试（全链路）
│   ├── 渗透测试报告
│   └── 恢复演练报告
└── Git tag: v2.0.0
```

### 2.4 Phase 6 阶段门禁

- [ ] AR-25: 仓库无二进制文件残留
- [ ] NF-02: 网关超时中间件单元测试通过
- [ ] NF-01: deletion-worker 集成测试通过，PII 匿名化验证
- [ ] AR-13: argon2id 哈希/验证/rehash 单元测试通过
- [ ] AR-23: 备份恢复演练完成
- [ ] AR-17: credit/subscription/rebate 覆盖率 >60%
- [ ] FN-02: 订单状态机全部测试通过
- [ ] FN-01: 微信/支付宝沙箱支付流程通过
- [ ] FN-05: Admin API 权限隔离验证通过
- [ ] FN-10: APNs/FCM 推送到达验证
- [ ] AR-21: Helm Chart 部署验证 + CI 流水线通过
- [ ] UX-08: 四端定价页验证
- [ ] AR-18: 全链路集成测试通过
- [ ] AR-16: 渗透测试报告无 Critical/High 漏洞
- [ ] CHANGELOG_V2.0.md 编写完成
- [ ] Git tag v2.0.0

---

## 3. Phase 7 — P1 任务分解（32 项）

> 体验与增长，上线后首月启动。Phase 6 全部完成后开始。

### 3.1 可靠性与运维（4 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及服务 |
|------|---------|------|------|---------|
| P1-1 | NF-03 | 熔断器共享包 | 无 | 全服务 |
| P1-2 | NF-04 | 健康检查真实依赖 | 无 | 全服务 |
| P1-3 | AR-22 | CI/CD 流水线完善 | AR-28 | 全服务 |
| P1-4 | AR-28 | Lint 严格化 | 无 | 全服务 |

**Task P1-1 NF-03**: 提取 `notification-service/pkg/circuitbreaker` → `pkg/circuitbreaker` 共享包，所有服务统一引入。验证：状态机测试覆盖率 ≥80%。

**Task P1-2 NF-04**: 各服务 `/health` 新增 PG `SELECT 1` + Redis `PING` + 下游可达性检测。验证：Redis 停止后健康检查返回 503。

**Task P1-4 AR-28**: 配置 `.golangci.yml`（errcheck/govet/staticcheck/gosec/revive/unused）。验证：全部服务 lint 零 error。

**Task P1-3 AR-22**: GitHub Actions 完整流水线（lint→test→build→push→deploy），并行构建。验证：总时间 ≤15min。

### 3.2 UX 体验优化（7 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及端 |
|------|---------|------|------|--------|
| P1-5 | UX-01 | 一键登录 | FN-15 | 移动端+auth |
| P1-6 | UX-02 | 生物识别登录 | UX-01 | 移动端+auth |
| P1-7 | UX-05 | 个性化仪表盘 | 无 | 全端+config |
| P1-8 | UX-09 | 支付流程闭环 | 无 | 全端+payment |
| P1-9 | UX-10 | 升降级体验 | UX-09 | 全端+account |
| P1-10 | UX-11 | 续费提醒 | FN-12 | account+notification |
| P1-11 | UX-12 | 推荐进度可视化 | FN-12 | 全端+data-product |

**Task P1-5 UX-01**: 微信/Apple Sign-In/Google One Tap OAuth，新增 `social_accounts` 表。验证：社交平台沙箱登录成功率 ≥95%。

**Task P1-6 UX-02**: iOS Face ID/Touch ID + Android 指纹识别，设备 token 绑定。验证：token 过期/设备变更回退密码登录。

**Task P1-7 UX-05**: config-service 配置 `dashboard_layout_{level}`，四端按配置渲染卡片。验证：L0/L2+/L4 卡片内容正确。

**Task P1-8 UX-09**: 支付结果页、电子发票申请、失败重试、异常订单自动修复。验证：支付→回调丢失→自动修复完整流程。

**Task P1-9 UX-10**: 升级费用预览计算器、降级挽留弹窗、升级立即生效。验证：升级 ≤5s 生效。

**Task P1-10 UX-11**: Asynq 定时任务 T-7/T-3/T-1 多通道提醒。验证：去重逻辑、深度链接直达支付页。

**Task P1-11 UX-12**: 推荐漏斗可视化 + 收益趋势图。验证：页面加载 ≤2s。

### 3.3 功能与商业化（6 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及服务 |
|------|---------|------|------|---------|
| P1-12 | FN-04 | 退款流程 | UX-09 | payment |
| P1-13 | FN-06 | 运营数据大屏 | AR-06 | data-product+grafana |
| P1-14 | FN-07 | 订阅管理后台 | 无 | account+config |
| P1-15 | FN-08 | 风控管理后台 | 无 | compliance |
| P1-16 | FN-12 | 事件埋点 SDK | 无 | 全端+data-product |
| P1-17 | FN-15 | OAuth 社交登录扩展 | 无 | auth |
| P1-18 | FN-17 | 数据导出/开放 API | AR-14 | account+data-product |

**Task P1-12 FN-04**: 退款策略（7天全额/超7天按比例）、自动/人工审核、原路退回。验证：退款金额计算 + 积分扣回 + 降级。

**Task P1-13 FN-06**: Grafana Dashboard：注册趋势/转化漏斗/MRR/RFM/K-factor。验证：大屏加载 ≤5s。

**Task P1-14 FN-07**: 套餐 CRUD + 优惠券管理 + 促销活动管理。验证：优惠券创建→核销→限额全流程。

**Task P1-15 FN-08**: 黑名单 CRUD（IP/设备/用户）+ 风险事件 + 异常注册预警。验证：Redis+PG 双写一致性。

**Task P1-16 FN-12**: 三端 SDK（Web TS/iOS Swift/Android Kotlin），自动采集 + 14 业务事件。验证：包体积 ≤50KB，初始化 ≤100ms。

**Task P1-17 FN-15**: auth-service 插件化 OAuth Provider（支付宝/Apple/Google）。验证：新增 Provider 仅需实现接口。

**Task P1-18 FN-17**: PIPL 数据导出 + 运营报表导出 + OAuth2 开放 API。验证：导出文件加密 + 审计日志。

### 3.4 移动端（7 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及端 |
|------|---------|------|------|--------|
| P1-19 | MB-02 | Android 字体集成 | 无 | Android |
| P1-20 | MB-09 | Token 安全存储验证 | 无 | iOS+Android |
| P1-21 | MB-10 | 证书固定 | 无 | iOS+Android |
| P1-22 | MB-13 | 小程序订阅消息 | FN-12 | 小程序+notification |
| P1-23 | MB-14 | 小程序分享能力 | FN-12 | 小程序 |
| P1-24 | MB-16~19 | 广告变现基础 | FN-12 | 全移动端+config |
| P1-25 | AR-19 | 性能/压力测试 | 无 | 全服务 |

**Task P1-19 MB-02**: Inter + Space Grotesk 字体文件集成到 Android `res/font/`。验证：字体渲染一致，总大小 ≤2MB。

**Task P1-20 MB-09**: access_token 内存存储 + refresh_token AES-256-GCM 加密 + 设备指纹绑定。验证：安全审计报告。

**Task P1-21 MB-10**: iOS URLSession Server Trust + Android OkHttp CertificatePinner。验证：中间人证书导致连接失败。

**Task P1-22 MB-13**: 微信订阅消息 4 类事件模板 + Asynq 重试。验证：消息触达率埋点。

**Task P1-23 MB-14**: onShareAppMessage + onShareTimeline，inviter_id 透传。验证：分享→注册→推荐关联。

**Task P1-24 MB-16~19**: config-service 广告配置 + 一主一备 SDK + Redis 频控 + 5s 视频限制。验证：L2-L4 零广告。

**Task P1-25 AR-19**: k6/wrk 压测核心 API。验证：P95 <500ms, P99 <1s, 错误率 <0.1%。

### 3.5 架构与可观测性（6 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及服务 |
|------|---------|------|------|---------|
| P1-26 | AR-01 | 服务间异步化 | AR-02 | 全服务 |
| P1-27 | AR-02 | 分布式事务 Saga | AR-05 | 全服务 |
| P1-28 | AR-05 | OpenTelemetry | 无 | 全服务 |
| P1-29 | AR-06 | Grafana Dashboard | AR-05 | 全服务 |
| P1-30 | AR-07 | 告警规则 | AR-06 | 全服务 |
| P1-31 | AR-14 | KMS/Vault | 无 | auth+全服务 |
| P1-32 | AR-15 | API 安全加固 | AR-14 | api-gateway |

**Task P1-28 AR-05**: OTel Go SDK + W3C Trace Context + Jaeger/Tempo。验证：跨服务调用链完整追踪。

**Task P1-27 AR-02**: Saga 编排器（Redis Streams），积分扣减→订阅激活→权益发放 + 补偿操作。验证：100 TPS 无数据不一致。

**Task P1-26 AR-01**: 积分消费→订阅激活→推荐奖励改为 Redis Streams/Kafka 异步。验证：吞吐量提升 ≥50%。

**Task P1-29 AR-06**: 4 个预置 Grafana Dashboard（JSON 模板 Git 管理）。验证：Dashboard 加载 ≤5s。

**Task P1-30 AR-07**: AlertManager YAML 规则 + 钉钉/企微 Webhook。验证：告警触发→通知到达。

**Task P1-31 AR-14**: Vault/阿里云 KMS 集成，密钥从环境变量迁移。验证：90 天轮换 + 紧急吊销。

**Task P1-32 AR-15**: 用户级限流（Redis 计数器）+ HMAC-SHA256 签名 + SQL 注入/XSS CI 扫描。验证：限流返回 429 + Retry-After。

---

## 4. Phase 8 — P2 任务分解（32 项）

> 竞争力提升，上线后 1-3 月启动。Phase 7 全部完成后开始。

### 4.1 可靠性与代码质量（3 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及服务 |
|------|---------|------|------|---------|
| P2-1 | NF-05 | 配置热更新 | 无 | 全服务 |
| P2-2 | NF-06 | 移动端深度链接 | 无 | iOS+Android |
| P2-3 | NF-07 | Gateway 代码重构 | 无 | api-gateway |

**Task P2-1 NF-05**: `pkg/config/watcher.go` 定时 30s 轮询 config-service + atomic.Value 覆盖。验证：配置修改后 ≤60s 生效。

**Task P2-2 NF-06**: iOS Universal Links + AASA 文件，Android App Links + assetlinks.json，neuro:// 兜底。验证：推荐参数透传。

**Task P2-3 NF-07**: 拆分 main.go（461行）→ internal/middleware/ + internal/proxy/ + cmd/main.go。验证：功能不变 + 中间件独立测试 ≥70%。

### 4.2 UX 体验（7 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及端 |
|------|---------|------|------|--------|
| P2-4 | UX-03 | 短信验证码自动填充 | 无 | iOS+Android |
| P2-5 | UX-04 | 渐进式注册 | 无 | 全端+auth |
| P2-6 | UX-06 | 空状态引导 | 无 | 全端 |
| P2-7 | UX-13 | 社交分享优化 | FN-12 | 全端 |
| P2-8 | UX-15 | 消息中心 | FN-10 | 全端+notification |
| P2-9 | UX-16 | 帮助/FAQ | 无 | 全端+config |
| P2-10 | FN-03 | 电子发票 | FN-01, FN-02 | payment |

**Task P2-4 UX-03**: iOS SMS Code AutoFill + Android SMS Retriever API。验证：短信到达后 ≤3s 自动填充。

**Task P2-5 UX-04**: 游客浏览→邮箱注册→手机验证渐进漏斗。验证：7天内手机绑定转化率埋点。

**Task P2-6 UX-06**: 积分/订阅/推荐页空状态引导卡片。验证：新用户首次进入展示正确。

**Task P2-7 UX-13**: ≥3 海报模板 + Open Graph + 小程序分享。验证：分享→注册链路埋点。

**Task P2-8 UX-15**: 消息中心聚合系统/运营消息，已读/未读，分页。验证：未读计数 <50ms。

**Task P2-9 UX-16**: FAQ 搜索/分类/客服入口，config-service 管理。验证：全文搜索查询。

**Task P2-10 FN-03**: 第三方发票平台对接，自动/手动开票。验证：开票→邮件推送完整流程。

### 4.3 功能与数据（4 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及服务 |
|------|---------|------|------|---------|
| P2-11 | FN-09 | 通知管理后台 | FN-10 | notification |
| P2-12 | FN-11 | 推送策略引擎 | FN-10 | notification |
| P2-13 | FN-13 | 实时行为流 | FN-12, AR-05 | data-product |
| P2-14 | MB-20~21 | 广告埋点+升级引导 | FN-12, MB-16~19 | 全端 |

### 4.4 移动端（10 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及端 |
|------|---------|------|------|--------|
| P2-15 | MB-01 | 设计系统统一 | 无 | 全端 |
| P2-16 | MB-03 | 响应式布局 | 无 | Web |
| P2-17 | MB-05 | 启动性能优化 | 无 | iOS+Android |
| P2-18 | MB-06 | 离线能力 | MB-09 | iOS+Android |
| P2-19 | MB-07 | 图片/资源优化 | 无 | 全端 |
| P2-20 | MB-08 | 骨架屏 | MB-01 | 全端 |
| P2-21 | MB-11 | Root/越狱检测 | 无 | iOS+Android |
| P2-22 | MB-12 | 截屏防护 | 无 | iOS+Android |
| P2-23 | MB-15 | 小程序分包 | 无 | 小程序 |

### 4.5 架构与质量（8 项）

| 序号 | 需求 ID | 名称 | 依赖 | 涉及服务 |
|------|---------|------|------|---------|
| P2-24 | AR-03 | 服务发现注册 | 无 | 全服务 |
| P2-25 | AR-08 | 日志关联优化 | AR-05 | 全服务 |
| P2-26 | AR-09 | Migration Down 脚本 | 无 | 全服务 |
| P2-27 | AR-10 | Redis 高可用 | AR-21 | 基础设施 |
| P2-28 | AR-11 | 连接池调优 | AR-19, NF-05 | 全服务 |
| P2-29 | AR-20 | E2E 测试 | AR-18 | 三端 |
| P2-30 | AR-24 | 金丝雀发布 | AR-21, AR-07 | 全服务 |
| P2-31 | AR-26 | 共享中间件包 | 无 | 全服务 |
| P2-32 | AR-27 | API 文档自动生成 | 无 | 全服务 |

---

## 5. Phase 9 — P3 任务分解（8 项）

> 长期规划，3+ 月后启动。Phase 8 全部完成后开始。

| 序号 | 需求 ID | 名称 | 依赖 | 涉及端/服务 |
|------|---------|------|------|------------|
| P3-1 | UX-07 | 搜索/快捷操作 | 无 | 全端 |
| P3-2 | UX-14 | 排行榜/社交证明 | FN-12 | 全端+account |
| P3-3 | UX-17 | 多语言 i18n | MB-01 | 全端 |
| P3-4 | FN-14 | A/B 测试框架 | FN-12, FN-13 | data-product+全端 |
| P3-5 | FN-16 | 企业微信/钉钉 | FN-15 | auth+account+payment |
| P3-6 | MB-04 | 无障碍 A11y | MB-01 | 全端 |
| P3-7 | AR-04 | API v2 版本管理 | AR-27 | api-gateway+全服务 |
| P3-8 | AR-12 | 读写分离 | AR-21 | data-product |

**Task P3-1 UX-07**: 全局搜索 + 快捷操作面板，Redis Sorted Set 频率排序。验证：搜索响应 ≤300ms。

**Task P3-2 UX-14**: 推荐达人榜 Top 20 + 社交证明提示。验证：隐私开关生效。

**Task P3-3 UX-17**: vue-i18n + String Catalog + 资源限定符 + 小程序 i18n。验证：zh-CN/en-US 四端文本一致。

**Task P3-4 FN-14**: A/B 分组引擎（用户 ID 哈希分流）+ 贝叶斯统计 + Admin 管理。验证：分组均匀性 + 统计显著性计算。

**Task P3-5 FN-16**: 企微/钉钉 OAuth + 通讯录同步 + 审批流集成。验证：沙箱环境 OAuth + 回调。

**Task P3-6 MB-04**: VoiceOver + TalkBack + ARIA + 动态字体 + 对比度 ≥4.5:1。验证：WCAG 2.1 AA 自动检测。

**Task P3-7 AR-04**: /api/v2/* 路由 + 版本生命周期（Active→Deprecated→Retired）。验证：v1/v2 并存互不影响。

**Task P3-8 AR-12**: data-product-service 双数据源（主库+只读副本）。验证：报表查询性能提升 ≥50%。

---

## 6. 风险管理

### 6.1 已知风险与缓解措施

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 支付网关审核周期不确定 | FN-01 阻塞 | 中 | 提前申请微信支付/支付宝商户号，利用沙箱环境先行开发 |
| APNs/FCM 证书/密钥管理复杂 | FN-10 延期 | 低 | 使用成熟 Provider SDK（apns2/firebase-admin），证书通过 KMS 管理 |
| argon2id 迁移影响存量用户登录 | AR-13 风险 | 低 | 渐进式 rehash，SM3/argon2id 双验证兼容期，充分测试 |
| K8s 生产部署复杂度高 | AR-21 延期 | 中 | 先 Dev 环境验证，Helm Chart 版本化管理，分步上线 |
| 全链路集成测试环境不稳定 | AR-18 阻塞 | 中 | testcontainers 隔离，CI 环境独立 |
| 广告 SDK 审核被拒 | MB-16~19 延期 | 低 | 提前注册穿山甲/优量汇开发者账号，准备合规材料 |
| 单人串行执行周期长 | 整体进度 | 确定 | 严格按依赖顺序执行，P0 阻塞项优先 |

### 6.2 依赖外部服务清单

| 外部服务 | 用途 | 关联需求 | 申请时机 |
|---------|------|---------|---------|
| 微信支付商户号 | H5/小程序/Native 支付 | FN-01 | Phase 6 开始前 |
| 支付宝商户号 | 手机网站/APP 支付 | FN-01 | Phase 6 开始前 |
| APNs 证书 (.p8) | iOS 推送 | FN-10 | Phase 6 中 |
| FCM 项目配置 | Android 推送 | FN-10 | Phase 6 中 |
| 华为 HMS 开发者 | 国内 Android 推送 | FN-10 | Phase 6 中 |
| 穿山甲开发者 | iOS/Android 广告主 SDK | MB-16 | Phase 7 开始前 |
| 优量汇开发者 | Android 广告备 SDK | MB-16 | Phase 7 开始前 |
| HashiCorp Vault | 密钥管理 | AR-14 | Phase 7 |
| 阿里云 KMS | 生产密钥管理 | AR-14 | Phase 7 |
| 第三方安全团队 | 渗透测试 | AR-16 | Phase 6 中 |
| 百望/航信 | 电子发票平台 | FN-03 | Phase 8 |
| 企业微信开发者 | B2B 扫码登录+审批 | FN-16 | Phase 9 |
| 钉钉开发者 | B2B 扫码登录+审批 | FN-16 | Phase 9 |

---

## 7. 附录

### 7.1 全量任务总表

| Phase | 序号 | 需求 ID | 名称 | 维度 | 依赖 | 涉及服务 |
|-------|------|---------|------|------|------|---------|
| 6 | 1 | AR-25 | 清理仓库 | 质量 | — | 全局 |
| 6 | 2 | NF-02 | 网关请求超时 | 可靠性 | — | api-gateway |
| 6 | 3 | NF-01 | 账号注销 Worker | 合规 | — | account-service |
| 6 | 4 | AR-13 | 密码哈希 argon2id | 安全 | — | auth-service |
| 6 | 5 | AR-23 | 数据库备份策略 | 运维 | — | 基础设施 |
| 6 | 6 | AR-17 | 单元测试补齐 | 质量 | — | 3 核心服务 |
| 6 | 7 | FN-02 | 订单管理系统 | 商业化 | — | payment-service(NEW) |
| 6 | 8 | FN-01 | 支付网关集成 | 商业化 | FN-02 | payment-service |
| 6 | 9 | FN-05 | 用户管理后台 | 运营 | FN-02 | account-service |
| 6 | 10 | FN-10 | APNs/FCM 推送 | 功能 | — | notification-service |
| 6 | 11 | AR-21 | K8s Helm Chart | 运维 | — | 全服务 |
| 6 | 12 | UX-08 | 定价透明度 | UX | — | 全端 |
| 6 | 13 | AR-18 | 集成测试 | 质量 | AR-17 | 全服务 |
| 6 | 14 | AR-16 | 渗透测试 | 安全 | AR-13,AR-17 | 全服务 |
| 7 | 15 | NF-03 | 熔断器共享包 | 可靠性 | — | 全服务 |
| 7 | 16 | NF-04 | 健康检查真实依赖 | 运维 | — | 全服务 |
| 7 | 17 | UX-01 | 一键登录 | UX | FN-15 | 移动端+auth |
| 7 | 18 | UX-02 | 生物识别登录 | UX | UX-01 | 移动端+auth |
| 7 | 19 | UX-05 | 个性化仪表盘 | UX | — | 全端+config |
| 7 | 20 | UX-09 | 支付流程闭环 | UX | — | 全端+payment |
| 7 | 21 | UX-10 | 升降级体验 | UX | UX-09 | 全端+account |
| 7 | 22 | UX-11 | 续费提醒 | UX | FN-12 | account+notification |
| 7 | 23 | UX-12 | 推荐进度可视化 | UX | FN-12 | 全端+data-product |
| 7 | 24 | FN-04 | 退款流程 | 商业化 | UX-09 | payment-service |
| 7 | 25 | FN-06 | 运营数据大屏 | 运营 | AR-06 | data-product+grafana |
| 7 | 26 | FN-07 | 订阅管理后台 | 运营 | — | account+config |
| 7 | 27 | FN-08 | 风控管理后台 | 运营 | — | compliance-service |
| 7 | 28 | FN-12 | 事件埋点 SDK | 数据 | — | 全端+data-product |
| 7 | 29 | FN-15 | OAuth 社交登录扩展 | 功能 | — | auth-service |
| 7 | 30 | FN-17 | 数据导出/开放 API | 合规 | AR-14 | account+data-product |
| 7 | 31 | MB-02 | Android 字体 | Mobile | — | Android |
| 7 | 32 | MB-09 | Token 安全存储 | 安全 | — | iOS+Android |
| 7 | 33 | MB-10 | 证书固定 | 安全 | — | iOS+Android |
| 7 | 34 | MB-13 | 小程序订阅消息 | Mobile | FN-12 | 小程序+notification |
| 7 | 35 | MB-14 | 小程序分享能力 | Mobile | FN-12 | 小程序 |
| 7 | 36 | MB-16~19 | 广告变现基础 | Mobile | FN-12 | 全移动端+config |
| 7 | 37 | AR-01 | 服务间异步化 | 架构 | AR-02 | 全服务 |
| 7 | 38 | AR-02 | 分布式事务 Saga | 架构 | AR-05 | 全服务 |
| 7 | 39 | AR-05 | OpenTelemetry | 可观测 | — | 全服务 |
| 7 | 40 | AR-06 | Grafana Dashboard | 可观测 | AR-05 | 全服务 |
| 7 | 41 | AR-07 | 告警规则 | 可观测 | AR-06 | 全服务 |
| 7 | 42 | AR-14 | KMS/Vault | 安全 | — | 全服务 |
| 7 | 43 | AR-15 | API 安全加固 | 安全 | AR-14 | api-gateway |
| 7 | 44 | AR-19 | 性能/压力测试 | 质量 | — | 全服务 |
| 7 | 45 | AR-22 | CI/CD 完善 | 运维 | AR-28 | 全服务 |
| 7 | 46 | AR-28 | Lint 严格化 | 质量 | — | 全服务 |
| 8 | 47 | NF-05 | 配置热更新 | 运维 | — | 全服务 |
| 8 | 48 | NF-06 | 移动端深度链接 | 增长 | — | iOS+Android |
| 8 | 49 | NF-07 | Gateway 重构 | 质量 | — | api-gateway |
| 8 | 50 | UX-03 | 短信验证码自动填充 | UX | — | iOS+Android |
| 8 | 51 | UX-04 | 渐进式注册 | UX | — | 全端+auth |
| 8 | 52 | UX-06 | 空状态引导 | UX | — | 全端 |
| 8 | 53 | UX-13 | 社交分享优化 | UX | FN-12 | 全端 |
| 8 | 54 | UX-15 | 消息中心 | UX | FN-10 | 全端+notification |
| 8 | 55 | UX-16 | 帮助/FAQ | UX | — | 全端+config |
| 8 | 56 | FN-03 | 电子发票 | 商业化 | FN-01,FN-02 | payment-service |
| 8 | 57 | FN-09 | 通知管理后台 | 运营 | FN-10 | notification-service |
| 8 | 58 | FN-11 | 推送策略引擎 | 功能 | FN-10 | notification-service |
| 8 | 59 | FN-13 | 实时行为流 | 数据 | FN-12,AR-05 | data-product-service |
| 8 | 60 | MB-01 | 设计系统统一 | Mobile | — | 全端 |
| 8 | 61 | MB-03 | 响应式布局 | Mobile | — | Web |
| 8 | 62 | MB-05 | 启动性能优化 | Mobile | — | iOS+Android |
| 8 | 63 | MB-06 | 离线能力 | Mobile | MB-09 | iOS+Android |
| 8 | 64 | MB-07 | 图片/资源优化 | Mobile | — | 全端 |
| 8 | 65 | MB-08 | 骨架屏 | Mobile | MB-01 | 全端 |
| 8 | 66 | MB-11 | Root/越狱检测 | 安全 | — | iOS+Android |
| 8 | 67 | MB-12 | 截屏防护 | 安全 | — | iOS+Android |
| 8 | 68 | MB-15 | 小程序分包 | Mobile | — | 小程序 |
| 8 | 69 | MB-20~21 | 广告埋点+升级引导 | Mobile | FN-12,MB-16~19 | 全端 |
| 8 | 70 | AR-03 | 服务发现注册 | 架构 | — | 全服务 |
| 8 | 71 | AR-08 | 日志关联优化 | 可观测 | AR-05 | 全服务 |
| 8 | 72 | AR-09 | Migration Down 脚本 | 质量 | — | 全服务 |
| 8 | 73 | AR-10 | Redis 高可用 | 运维 | AR-21 | 基础设施 |
| 8 | 74 | AR-11 | 连接池调优 | 性能 | AR-19,NF-05 | 全服务 |
| 8 | 75 | AR-20 | E2E 测试 | 质量 | AR-18 | 三端 |
| 8 | 76 | AR-24 | 金丝雀发布 | 运维 | AR-21,AR-07 | 全服务 |
| 8 | 77 | AR-26 | 共享中间件包 | 质量 | — | 全服务 |
| 8 | 78 | AR-27 | API 文档自动生成 | 质量 | — | 全服务 |
| 9 | 79 | UX-07 | 搜索/快捷操作 | UX | — | 全端 |
| 9 | 80 | UX-14 | 排行榜/社交证明 | UX | FN-12 | 全端+account |
| 9 | 81 | UX-17 | 多语言 i18n | UX | MB-01 | 全端 |
| 9 | 82 | FN-14 | A/B 测试框架 | 数据 | FN-12,FN-13 | data-product+全端 |
| 9 | 83 | FN-16 | 企业微信/钉钉 | B2B | FN-15 | auth+account+payment |
| 9 | 84 | MB-04 | 无障碍 A11y | Mobile | MB-01 | 全端 |
| 9 | 85 | AR-04 | API v2 版本管理 | 架构 | AR-27 | api-gateway+全服务 |
| 9 | 86 | AR-12 | 读写分离 | 数据层 | AR-21 | data-product-service |

### 7.2 需求→任务→代码文件映射表

| 需求 ID | PRD 章节 | SSD 章节 | 主要代码文件 | Phase |
|---------|---------|---------|-------------|-------|
| NF-01 | 3.1 | 3.1.1 | account-service/internal/worker/deletion.go | 6 |
| NF-02 | 3.1 | 3.1.2 | api-gateway/internal/middleware/timeout.go | 6 |
| AR-13 | 3.1 | 3.1.3 | auth-service/internal/auth/argon2id.go | 6 |
| AR-16 | 3.1 | 3.1.4 | scripts/security/ | 6 |
| AR-17 | 3.1 | 3.1.5 | credit-service/internal/service/*_test.go | 6 |
| AR-18 | 3.1 | 3.1.6 | tests/integration/full_journey_test.go | 6 |
| AR-23 | 3.1 | 3.1.7 | infra/backup/pg_backup.sh | 6 |
| AR-25 | 3.1 | 3.1.8 | .gitignore, README.md | 6 |
| AR-21 | 3.1 | 3.1.9 | helm/account-center/ | 6 |
| FN-01 | 3.1 | 3.1.10 | payment-service/internal/service/wechat_pay.go | 6 |
| FN-02 | 3.1 | 3.1.11 | payment-service/internal/service/order.go | 6 |
| FN-05 | 3.1 | 3.1.12 | account-service/internal/handler/admin.go | 6 |
| FN-10 | 3.1 | 3.1.13 | notification-service/internal/provider/apns.go | 6 |
| UX-08 | 3.1 | 3.1.14 | web/src/views/PricingPage.vue | 6 |
| NF-03 | 3.2 | 3.2.1 | pkg/circuitbreaker/circuitbreaker.go | 7 |
| NF-04 | 3.2 | 3.2.2 | pkg/health/health.go | 7 |
| UX-01 | 3.2 | 3.2.3 | auth-service/internal/provider/wechat_oauth.go | 7 |
| UX-02 | 3.2 | 3.2.4 | auth-service/internal/handler/biometric_handler.go | 7 |
| UX-05 | 3.2 | 3.2.5 | account-service/internal/handler/dashboard.go | 7 |
| UX-09 | 3.2 | 3.2.6 | payment-service/internal/handler/payment_result.go | 7 |
| UX-10 | 3.2 | 3.2.7 | account-service/internal/service/subscription_upgrade.go | 7 |
| UX-11 | 3.2 | 3.2.8 | account-service/internal/worker/renewal_reminder.go | 7 |
| UX-12 | 3.2 | 3.2.9 | account-service/internal/service/referral_funnel.go | 7 |
| FN-04 | 3.2 | 3.2.10 | payment-service/internal/service/refund.go | 7 |
| FN-06 | 3.2 | 3.2.11 | monitoring/grafana/dashboards/operations_dashboard.json | 7 |
| FN-07 | 3.2 | 3.2.12 | account-service/internal/handler/admin_plan.go | 7 |
| FN-08 | 3.2 | 3.2.13 | compliance-service/internal/handler/admin_risk.go | 7 |
| FN-12 | 3.2 | 3.2.14 | sdks/web/src/tracker.ts | 7 |
| FN-15 | 3.2 | 3.2.15 | auth-service/internal/provider/alipay_oauth.go | 7 |
| FN-17 | 3.2 | 3.2.16 | account-service/internal/handler/data_export.go | 7 |
| MB-02 | 3.2 | 3.2.17 | android/.../res/font/ | 7 |
| MB-09 | 3.2 | 3.2.18 | ios/.../KeychainManager.swift | 7 |
| MB-10 | 3.2 | 3.2.19 | ios/.../CertificatePinner.swift | 7 |
| MB-13 | 3.2 | 3.2.20 | miniprogram/pages/ | 7 |
| MB-14 | 3.2 | 3.2.21 | miniprogram/pages/ | 7 |
| MB-16~19 | 3.2 | 3.2.22 | config-service ad_config + mobile SDK | 7 |
| AR-01 | 3.2 | 3.2.23 | pkg/messaging/ | 7 |
| AR-02 | 3.2 | 3.2.24 | pkg/saga/ | 7 |
| AR-05 | 3.2 | 3.2.25 | pkg/telemetry/ | 7 |
| AR-06 | 3.2 | 3.2.26 | monitoring/grafana/dashboards/ | 7 |
| AR-07 | 3.2 | 3.2.27 | monitoring/alertmanager/rules/ | 7 |
| AR-14 | 3.2 | 3.2.28 | pkg/kms/ | 7 |
| AR-15 | 3.2 | 3.2.29 | api-gateway/internal/middleware/ratelimit.go | 7 |
| AR-19 | 3.2 | 3.2.30 | scripts/loadtest/ | 7 |
| AR-22 | 3.2 | 3.2.31 | .github/workflows/ci.yml | 7 |
| AR-28 | 3.2 | 3.2.32 | .golangci.yml | 7 |
| NF-05 | 3.3 | 3.3.01 | pkg/config/watcher.go | 8 |
| NF-06 | 3.3 | 3.3.02 | ios/.../DeepLinkRouter.swift | 8 |
| NF-07 | 3.3 | 3.3.03 | api-gateway/internal/middleware/*.go | 8 |
| UX-03 | 3.3 | 3.3.04 | ios/.../LoginView.swift | 8 |
| UX-04 | 3.3 | 3.3.05 | auth-service/internal/handler/guest_handler.go | 8 |
| UX-06 | 3.3 | 3.3.06 | web/src/components/EmptyStateCard.vue | 8 |
| UX-13 | 3.3 | 3.3.07 | credit-service/internal/service/share_service.go | 8 |
| UX-15 | 3.3 | 3.3.08 | notification-service/internal/handler/notification_handler.go | 8 |
| UX-16 | 3.3 | 3.3.09 | config-service/internal/handler/faq_handler.go | 8 |
| FN-03 | 3.3 | 3.3.10 | payment-service/internal/service/invoice_service.go | 8 |
| FN-09 | 3.3 | 3.3.11 | notification-service/internal/handler/template_handler.go | 8 |
| FN-11 | 3.3 | 3.3.12 | notification-service/internal/service/push_strategy_engine.go | 8 |
| FN-13 | 3.3 | 3.3.13 | data-product-service/internal/service/stream_processor.go | 8 |
| MB-01 | 3.3 | 3.3.14 | design-tokens/ | 8 |
| MB-03 | 3.3 | 3.3.15 | web/src/composables/useBreakpoint.ts | 8 |
| MB-05 | 3.3 | 3.3.16 | ios/.../StartupManager.swift | 8 |
| MB-06 | 3.3 | 3.3.17 | ios/.../OfflineCacheManager.swift | 8 |
| MB-07 | 3.3 | 3.3.18 | web/src/composables/useImageOptimization.ts | 8 |
| MB-08 | 3.3 | 3.3.19 | web/src/components/SkeletonScreen.vue | 8 |
| MB-11 | 3.3 | 3.3.20 | ios/.../SecurityChecker.swift | 8 |
| MB-12 | 3.3 | 3.3.21 | ios/.../ScreenshotProtection.swift | 8 |
| MB-15 | 3.3 | 3.3.22 | miniprogram/app.json | 8 |
| MB-20~21 | 3.3 | 3.3.23 | web/src/components/AdUpgradePrompt.vue | 8 |
| AR-03 | 3.3 | 3.3.24 | pkg/discovery/ | 8 |
| AR-08 | 3.3 | 3.3.25 | pkg/log/trace.go | 8 |
| AR-09 | 3.3 | 3.3.26 | migrations/ | 8 |
| AR-10 | 3.3 | 3.3.27 | helm/account-center/templates/redis.yaml | 8 |
| AR-11 | 3.3 | 3.3.28 | pkg/database/pool.go | 8 |
| AR-20 | 3.3 | 3.3.29 | tests/e2e/ | 8 |
| AR-24 | 3.3 | 3.3.30 | helm/account-center/templates/canary.yaml | 8 |
| AR-26 | 3.3 | 3.3.31 | pkg/server/ | 8 |
| AR-27 | 3.3 | 3.3.32 | docs/api/ | 8 |
| UX-07 | 3.4 | 3.4.1 | account-service/internal/handler/search.go | 9 |
| UX-14 | 3.4 | 3.4.2 | account-service/internal/service/leaderboard.go | 9 |
| UX-17 | 3.4 | 3.4.3 | web/src/i18n/index.ts | 9 |
| FN-14 | 3.4 | 3.4.4 | data-product-service/internal/service/ab_engine.go | 9 |
| FN-16 | 3.4 | 3.4.5 | auth-service/internal/provider/work_wechat.go | 9 |
| MB-04 | 3.4 | 3.4.6 | web/src/components/accessibility/ | 9 |
| AR-04 | 3.4 | 3.4.7 | api-gateway/internal/middleware/version.go | 9 |
| AR-12 | 3.4 | 3.4.8 | data-product-service/internal/repository/read_replica.go | 9 |
