# Account Center V2.0 迭代 — 文档编制实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 ITERATION V1.0.1 的 77 项改进建议，编制完整的 PRD V2.0.0、SSD V2.0.0、推进计划 V2.0.0 三份正式文档。

**Architecture:** 文档先行严格瀑布。PRD 覆盖全部需求定义与验收标准 → SSD 覆盖全部技术设计 → PLAN 覆盖任务分解与执行顺序。每份文档按 Phase 6/7/8/9 分组，与 ITERATION 优先级矩阵对齐。

**Tech Stack:** Markdown 文档，Go 微服务架构背景，8 服务 + 4 端前端

**设计规格:** `docs/superpowers/specs/2026-05-19-v2-iteration-design.md`

**基线文档:**
- `docs/ITERATION_AccountCenter_V2.0_V1.0.1.md` — 77 项改进建议（718 行）
- `docs/requirements/账户管理微服务 (Account Center) 产品需求说明书 V1.3.0.md` — PRD 基线
- `docs/SSD_AccountCenter_V1.3.1.md` — SSD 基线

---

## Phase A: PRD V2.0.0

### Task 1: 创建 PRD 文档骨架

**Files:**
- Create: `docs/PRD_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 创建文件并写入文档头部和完整目录结构**

写入以下内容：

```markdown
# Account Center V2.0 产品需求说明书

> **文档类型**: 产品需求说明书（正式版）
> **版本**: V2.0.0
> **日期**: 2026-05-19
> **状态**: 编制中
> **基线版本**: PRD V1.3.0
> **评估基准**: ITERATION_AccountCenter_V2.0_V1.0.1（77 项改进建议）
> **变更历史**:
> | 版本 | 日期 | 变更内容 | 作者 |
> |------|------|---------|------|
> | V2.0.0 | 2026-05-19 | 初始编制，覆盖全部 77 项改进需求 | |

---

## 目录

1. 产品概述
2. 用户角色与权限
3. 功能需求
   - 3.1 Phase 6 — P0 需求（14 项）
   - 3.2 Phase 7 — P1 需求（30 项）
   - 3.3 Phase 8 — P2 需求（29 项）
   - 3.4 Phase 9 — P3 需求（4 项）
4. 非功能需求
5. 数据需求
6. 接口需求（外部集成）
7. 验收标准总表
8. 附录
   - A. 需求追溯矩阵
   - B. 与 PRD V1.3.0 差异对照表

---

## 1. 产品概述

（待填充）

## 2. 用户角色与权限

（待填充）

## 3. 功能需求

（待填充）

## 4. 非功能需求

（待填充）

## 5. 数据需求

（待填充）

## 6. 接口需求（外部集成）

（待填充）

## 7. 验收标准总表

（待填充）

## 8. 附录

（待填充）
```

- [ ] **Step 2: 验证文件已创建**

读取 `docs/PRD_AccountCenter_V2.0.0.md` 前 30 行，确认目录结构完整。

- [ ] **Step 3: 提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: create PRD V2.0.0 skeleton"
```

---

### Task 2: 编写 PRD 第 1 章 — 产品概述

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 替换 `## 1. 产品概述` 下的 `（待填充）`

**数据来源:**
- ITERATION V1.0.1 第 1-64 行（1.1 项目概况、1.2 已实现成果、1.3 核心优势、1.4 已知短板）
- PRD V1.3.0 的产品定位描述

- [ ] **Step 1: 读取 ITERATION 1.1-1.3 节和 PRD V1.3.0 的产品定位**

读取 `docs/ITERATION_AccountCenter_V2.0_V1.0.1.md` 第 29-63 行。
读取 `docs/requirements/账户管理微服务 (Account Center) 产品需求说明书 V1.3.0.md` 前 50 行。

- [ ] **Step 2: 编写 PRD 第 1 章内容**

将 `## 1. 产品概述` 下的 `（待填充）` 替换为以下结构的完整内容：

```
### 1.1 产品定位与目标
- Account Center 在 Neuro 产品线中的定位
- V2.0 的核心目标：商业化上线就绪 + 合规红线解除 + 体验竞争力提升

### 1.2 V1.3.1 现状基线
- 从 ITERATION 1.2 表格提取：8 微服务、4 端前端、167 Go 源文件、51+8 需求全实现
- 安全合规、运维体系、文档体系的现状评估

### 1.3 V2.0 迭代目标
- 从 ITERATION 1.4 提取 8 项已知短板
- 从 ITERATION 第 8 章提取优先级矩阵汇总：P0:14 / P1:30 / P2:29 / P3:4
- 按 Phase 的发布策略：V2.0→V2.1→V2.2→V2.3

### 1.4 术语表
- 从 PRD V1.3.0 和 ITERATION 提取关键术语
- 新增 V2.0 引入的术语（如 payment-service、deletion-worker、argon2id）
```

- [ ] **Step 3: 验证**

读取写入的内容，确认四个子节都已填充，无 `（待填充）` 残留。

- [ ] **Step 4: 提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: write PRD V2.0.0 section 1 - product overview"
```

---

### Task 3: 编写 PRD 第 2 章 — 用户角色与权限

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 替换 `## 2. 用户角色与权限` 下的 `（待填充）`

**数据来源:**
- PRD V1.3.0 中的用户角色定义
- ITERATION 中的五级身份体系（L0-L4）

- [ ] **Step 1: 读取 PRD V1.3.0 的用户角色章节**

读取 `docs/requirements/账户管理微服务 (Account Center) 产品需求说明书 V1.3.0.md`，找到用户角色/权限相关章节。

- [ ] **Step 2: 编写 PRD 第 2 章内容**

替换 `（待填充）` 为完整内容，包括：

```
### 2.1 用户角色定义
- 终端用户（C 端）：L0-L4 五级身份体系
- 系统管理员：配置管理后台
- 运营人员：新增 Admin 后台（V2.0 新增角色）
- 审计人员：合规日志查看

### 2.2 V2.0 新增角色与权限
- 运营管理员：用户管理、数据大屏、订阅管理、风控管理
- 财务管理员：订单查询、退款审核、发票管理

### 2.3 权限矩阵
- 按角色 × 功能模块列出权限表
```

- [ ] **Step 3: 验证并提交**

---

### Task 4: 编写 PRD 3.1 节 — Phase 6 P0 需求（14 项）

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 替换 `## 3. 功能需求` 下的 `（待填充）`，先写入 3.1 节

**数据来源:**
- ITERATION V1.0.1 第 528-547 行（P0 矩阵）
- ITERATION 第 6 章第 499-507 行（NF-01~NF-07 代码级发现）
- ITERATION 各维度章节中对应 P0 项的详细说明

- [ ] **Step 1: 读取 ITERATION P0 矩阵和所有 P0 项的详细说明**

读取以下行范围：
- 第 528-547 行（P0 矩阵表）
- 第 92-97 行（UX-08, UX-09 — P0 部分）
- 第 128-133 行（FN-01, FN-02 — P0）
- 第 141 行（FN-05 — P0）
- 第 153 行（FN-10 — P0）
- 第 287 行（AR-13 — P0）
- 第 290 行（AR-16 — P0）
- 第 296-297 行（AR-17, AR-18 — P0）
- 第 307 行（AR-23 — P0）
- 第 314 行（AR-25 — P0）
- 第 305 行（AR-21 — P0）
- 第 501-502 行（NF-01, NF-02 — P0）

- [ ] **Step 2: 编写 PRD 3.0 和 3.1 节**

替换 `## 3. 功能需求` 下的 `（待填充）` 为：

```
### 3.0 概述

本节按迭代 Phase 分组列出全部功能需求。需求 ID 沿用 ITERATION V1.0.1 编号，保持可追溯性。

每项需求包含：
- **需求 ID**: 与 ITERATION 对应
- **名称**: 简明描述
- **优先级**: P0/P1/P2/P3
- **用户故事**: 作为…我希望…以便…
- **验收标准**: 可量化的完成条件
- **依赖**: 前置需求或外部依赖

### 3.1 Phase 6 — P0 需求（14 项）
```

然后为 14 项 P0 需求逐一编写，每项格式：

```markdown
#### NF-01: 账号注销 Worker 实现

**优先级**: P0  
**维度**: 合规  
**用户故事**: 作为用户，我希望提交账号注销请求后，冻结期到期时系统能自动删除/匿名化我的个人数据，以便我的隐私权得到保障（PIPL 合规）。  
**验收标准**:
- [ ] 用户提交注销请求后，7 天冻结期内可撤回
- [ ] 冻结期到期后，deletion-worker 自动执行数据匿名化
- [ ] 手机号/邮箱/姓名等 PII 字段替换为 `deleted_xxx` 格式
- [ ] `DeletionDeletedAt` 字段被正确写入时间戳
- [ ] Redis session/cache 被清理
- [ ] 注销操作记录审计日志
- [ ] 单元测试覆盖率 >80%
**依赖**: 无
```

14 项 P0 需求清单（按执行顺序）：
1. NF-01: 账号注销 Worker 实现
2. NF-02: 网关请求超时配置
3. AR-13: 密码哈希升级至 argon2id
4. AR-16: 第三方安全渗透测试
5. AR-17: 核心服务单元测试补齐
6. AR-18: 集成测试（全链路）
7. AR-23: 数据库备份策略
8. AR-25: 清理仓库（二进制文件）
9. AR-21: K8s Helm Chart + CI/CD
10. FN-01: 支付网关集成
11. FN-02: 订单管理系统
12. FN-05: 用户管理后台
13. FN-10: APNs/FCM 推送集成
14. UX-08: 定价透明度

每项必须包含完整的用户故事、验收标准（可量化）、依赖关系。不得使用占位符。

- [ ] **Step 3: 验证**

检查 14 项需求是否全部包含：用户故事、验收标准（至少 3 条）、依赖。无 `（待填充）` 残留。

- [ ] **Step 4: 提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: write PRD V2.0.0 section 3.1 - Phase 6 P0 requirements (14 items)"
```

---

### Task 5: 编写 PRD 3.2 节 — Phase 7 P1 需求（30 项）

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 在 3.1 节后追加 3.2 节

**数据来源:**
- ITERATION V1.0.1 第 549-584 行（P1 矩阵）
- ITERATION 各维度章节中对应 P1 项的详细说明

- [ ] **Step 1: 读取 ITERATION P1 矩阵和所有 P1 项的详细说明**

读取第 549-584 行，以及各维度章节中 P1 标记的行（共 30 项）。

- [ ] **Step 2: 编写 PRD 3.2 节**

在 3.1 节后追加：

```markdown
### 3.2 Phase 7 — P1 需求（30 项）
```

30 项 P1 需求清单：
1. NF-03: 服务间熔断器提升为共享包
2. NF-04: 健康检查增加真实依赖检测
3. UX-01: 一键登录（微信/Apple/Google）
4. UX-02: 生物识别快捷登录
5. UX-05: 个性化仪表盘
6. UX-09: 支付流程闭环
7. UX-10: 升降级体验优化
8. UX-11: 订阅续费提醒
9. UX-12: 推荐进度可视化
10. FN-04: 退款流程
11. FN-06: 运营数据大屏
12. FN-07: 订阅管理后台
13. FN-08: 风控管理后台
14. FN-12: 事件埋点 SDK
15. FN-15: OAuth 社交登录扩展
16. FN-17: 数据导出/开放 API
17. MB-02: Android 字体集成
18. MB-09: Token 安全存储升级验证
19. MB-10: 证书固定
20. MB-13: 小程序订阅消息
21. MB-14: 小程序分享能力
22. MB-16~19: 广告变现基础
23. AR-01: 服务间通信异步化
24. AR-02: 分布式事务 Saga
25. AR-05: 分布式追踪（OpenTelemetry）
26. AR-06: 自定义 Grafana 仪表盘
27. AR-07: 告警规则配置
28. AR-14: KMS/Vault 密钥管理
29. AR-15: API 安全加固
30. AR-19: 性能/压力测试
31. AR-22: CI/CD 流水线完善
32. AR-28: Lint 严格化

注：ITERATION P1 矩阵列出约 30 项，具体以实际为准。

每项格式与 Task 4 相同：用户故事 + 验收标准 + 依赖。

- [ ] **Step 3: 验证并提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: write PRD V2.0.0 section 3.2 - Phase 7 P1 requirements"
```

---

### Task 6: 编写 PRD 3.3 节 — Phase 8 P2 需求（29 项）

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 在 3.2 节后追加 3.3 节

**数据来源:**
- ITERATION V1.0.1 第 586-621 行（P2 矩阵）

- [ ] **Step 1: 读取 ITERATION P2 矩阵和 P2 项详细说明**

- [ ] **Step 2: 编写 PRD 3.3 节**

29 项 P2 需求，格式同上。

- [ ] **Step 3: 验证并提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: write PRD V2.0.0 section 3.3 - Phase 8 P2 requirements"
```

---

### Task 7: 编写 PRD 3.4 节 — Phase 9 P3 需求（4 项）

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 在 3.3 节后追加 3.4 节

**数据来源:**
- ITERATION V1.0.1 第 623-634 行（P3 矩阵）

- [ ] **Step 1: 读取 ITERATION P3 矩阵**

- [ ] **Step 2: 编写 PRD 3.4 节**

4 项 P3 需求：
1. UX-07: 搜索/快捷操作
2. UX-14: 排行榜/社交证明
3. UX-17: 多语言 i18n 架构
4. FN-14: A/B 测试框架
5. FN-16: 企业微信/钉钉集成
6. MB-04: 无障碍 Accessibility
7. AR-04: API v2 版本管理
8. AR-12: 读写分离

注：ITERATION P3 矩阵列出 8 项（非 4 项），以实际为准。

- [ ] **Step 3: 验证并提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: write PRD V2.0.0 section 3.4 - Phase 9 P3 requirements"
```

---

### Task 8: 编写 PRD 第 4-6 章 — 非功能需求、数据需求、接口需求

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 替换第 4/5/6 章的 `（待填充）`

**数据来源:**
- ITERATION V1.0.1 第 254-318 行（技术架构维度，含安全/测试/部署/代码质量）
- ITERATION V1.0.1 第 327-346 行（5.8.2 统一版本基准）
- PRD V1.3.0 的非功能需求章节

- [ ] **Step 1: 读取相关数据源**

- [ ] **Step 2: 编写 PRD 第 4 章 — 非功能需求**

```
### 4.1 性能指标
- API P95 < 200ms, P99 < 500ms（登录/注册/订阅购买核心路径）
- 支持 1000 并发用户
- 数据库查询 P95 < 50ms

### 4.2 安全合规
- 等保 2.0 二级（GB/T 22239-2019）
- PIPL 个人信息保护法合规（数据删除/匿名化/导出）
- OWASP Top 10 渗透测试通过
- 密码哈希 argon2id
- KMS 密钥管理（Vault/阿里云 KMS）

### 4.3 可观测性
- OpenTelemetry 分布式追踪（W3C Trace Context）
- 服务健康总览 Dashboard
- 关键告警：服务宕机/P99 超阈值/错误率 >1%
- 日志包含 trace_id/span_id

### 4.4 兼容性
- Web: Chrome/Firefox/Safari/Edge 最新 2 个大版本
- iOS: iOS 17+
- Android: minSdk 26 (Android 8.0)
- 微信小程序: 基础库 3.0+
- API: v1 保持向后兼容，v2 新增路由
```

- [ ] **Step 3: 编写 PRD 第 5 章 — 数据需求**

```
### 5.1 新增数据实体
- orders（订单表）
- ad_config（广告配置）
- events（埋点事件）
- push_tokens（推送设备令牌）
- admin_users（管理员账户）

### 5.2 数据迁移策略（V1.3.1→V2.0）
- 存量用户密码 argon2id rehash（登录时渐进迁移）
- 新增表的 Goose migration 脚本
- Redis key 结构变更的兼容方案
```

- [ ] **Step 4: 编写 PRD 第 6 章 — 接口需求**

```
### 6.1 支付网关
- 微信支付：H5/小程序/Native 三种场景
- 支付宝：手机网站/APP 支付
- 异步回调处理、对账、退款

### 6.2 推送服务
- iOS: APNs HTTP/2
- Android: FCM + 华为 HMS + 小米/OPPO/vivo 厂商通道
- notification-service provider 架构（与 SMS 多供应商一致）

### 6.3 广告 SDK
- 穿山甲（字节跳动）
- 优量汇（腾讯）
- 每平台 2 个 SDK（primary/backup）

### 6.4 KMS
- HashiCorp Vault 或阿里云 KMS
- 密钥自动轮换、审计记录、紧急吊销
```

- [ ] **Step 5: 验证并提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: write PRD V2.0.0 sections 4-6 - non-functional, data, interface requirements"
```

---

### Task 9: 编写 PRD 第 7-8 章 — 验收标准总表与附录

**Files:**
- Modify: `docs/PRD_AccountCenter_V2.0.0.md` — 替换第 7/8 章的 `（待填充）`

- [ ] **Step 1: 编写 PRD 第 7 章 — 验收标准总表**

按 Phase 分组的汇总表，每项一行：
| Phase | 需求 ID | 名称 | 验收标准摘要 | 状态 |
|-------|---------|------|-------------|------|
| Phase 6 | NF-01 | 账号注销 Worker | ... | 待实现 |
| ... | ... | ... | ... | ... |

共 77 行，覆盖全部需求。

- [ ] **Step 2: 编写 PRD 第 8 章 — 附录**

```
### 附录 A: 需求追溯矩阵

| ITERATION ID | PRD 章节 | 需求名称 | Phase |
|-------------|---------|---------|-------|
| NF-01 | 3.1 | 账号注销 Worker | Phase 6 |
| ... | ... | ... | ... |

### 附录 B: 与 PRD V1.3.0 差异对照表

| V1.3.0 章节 | V2.0 变更类型 | 说明 |
|------------|-------------|------|
| 订阅购买 | 新增 | 支付网关集成（FN-01） |
| 用户管理 | 新增 | Admin 管理后台（FN-05） |
| ... | ... | ... |
```

- [ ] **Step 3: 最终验证**

全文搜索 `（待填充）`，确认无残留。确认文档状态从"编制中"改为"已完成"。

- [ ] **Step 4: 提交 PRD 完整版**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md
git commit -m "docs: complete PRD V2.0.0 - all 77 requirements with acceptance criteria"
```

---

## Phase B: SSD V2.0.0

### Task 10: 创建 SSD 文档骨架

**Files:**
- Create: `docs/SSD_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 创建文件并写入文档头部和完整目录结构**

按设计规格第 4 节定义的 SSD 结构创建骨架，包含：
1. 系统概述（1.x）
2. 系统架构（2.x）
3. 服务详细设计（3.x，按 Phase 分组）
4. 数据设计（4.x）
5. API 设计（5.x）
6. 安全设计（6.x）
7. 部署设计（7.x）
8. 可观测性设计（8.x）
9. 附录（9.x）

每个章节用 `（待填充）` 占位。

- [ ] **Step 2: 提交**

```bash
git add docs/SSD_AccountCenter_V2.0.0.md
git commit -m "docs: create SSD V2.0.0 skeleton"
```

---

### Task 11: 编写 SSD 第 1-2 章 — 系统概述与系统架构

**Files:**
- Modify: `docs/SSD_AccountCenter_V2.0.0.md`

**数据来源:**
- ITERATION V1.0.1 第 327-346 行（5.8.2 统一版本基准）
- ITERATION V1.0.1 第 356-376 行（5.8.3 三环境部署差异）
- SSD V1.3.1 的架构描述

- [ ] **Step 1: 读取 SSD V1.3.1 架构章节和 ITERATION 5.8 节**

- [ ] **Step 2: 编写 SSD 第 1 章 — 系统概述**

```
1.1 系统范围与边界
1.2 V1.3.1 架构基线（8 服务拓扑、通信模式、数据层）
1.3 V2.0 架构演进目标（新增 payment-service、可观测性、K8s）
```

- [ ] **Step 3: 编写 SSD 第 2 章 — 系统架构**

```
2.1 整体架构图（9 服务 + 4 端前端 + 基础设施层）
2.2 服务清单与职责（含新增 payment-service）
2.3 通信模式（同步 HTTP → 部分异步化 Redis Streams/Kafka）
2.4 技术选型版本矩阵（引用 ITERATION 5.8.2）
```

- [ ] **Step 4: 验证并提交**

```bash
git add docs/SSD_AccountCenter_V2.0.0.md
git commit -m "docs: write SSD V2.0.0 sections 1-2 - system overview and architecture"
```

---

### Task 12: 编写 SSD 3.1 节 — Phase 6 P0 技术设计

**Files:**
- Modify: `docs/SSD_AccountCenter_V2.0.0.md`

**数据来源:**
- ITERATION V1.0.1 各 P0 项的详细说明
- 代码级发现（NF-01~NF-07）

- [ ] **Step 1: 读取 ITERATION 中所有 P0 项的详细描述**

- [ ] **Step 2: 编写 SSD 3.1 节**

为 Phase 6 的 14 项 P0 需求逐一编写技术设计，每项包含：
- 需求 ID 关联
- 技术方案（架构图/流程图描述）
- 关键代码路径
- 数据库变更
- API 变更
- 测试策略

重点设计的项目：
- **NF-01 deletion-worker**: Asynq 定时任务架构、数据匿名化规则、Redis 清理策略
- **NF-02 网关超时**: httputil.ReverseProxy Transport 配置、全局超时中间件
- **AR-13 argon2id 迁移**: 新旧密码并存、登录时 rehash、审计完整性保留 SM3
- **FN-01/02 payment-service**: 独立微服务架构、异步回调、对账、退款
- **AR-21 K8s Helm Chart**: Chart 结构、HPA 配置、滚动更新策略
- **AR-22 CI/CD**: GitHub Actions 流水线设计（lint→test→build→deploy）

- [ ] **Step 3: 验证并提交**

```bash
git add docs/SSD_AccountCenter_V2.0.0.md
git commit -m "docs: write SSD V2.0.0 section 3.1 - Phase 6 P0 technical design"
```

---

### Task 13: 编写 SSD 3.2-3.4 节 — Phase 7/8/9 技术设计

**Files:**
- Modify: `docs/SSD_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 编写 SSD 3.2 节 — Phase 7 P1 技术设计（30 项）**

每项 P1 需求的技术方案，重点包括：
- AR-05 OpenTelemetry 接入方案
- NF-03 熔断器共享包设计
- MB-16~19 广告变现技术方案

- [ ] **Step 2: 编写 SSD 3.3 节 — Phase 8 P2 技术设计（29 项）**

- [ ] **Step 3: 编写 SSD 3.4 节 — Phase 9 P3 技术设计（4 项）**

- [ ] **Step 4: 验证并提交**

```bash
git add docs/SSD_AccountCenter_V2.0.0.md
git commit -m "docs: write SSD V2.0.0 sections 3.2-3.4 - Phase 7/8/9 technical design"
```

---

### Task 14: 编写 SSD 第 4-5 章 — 数据设计与 API 设计

**Files:**
- Modify: `docs/SSD_AccountCenter_V2.0.0.md`

**数据来源:**
- PRD V2.0.0 第 5 章（数据需求）
- SSD V1.3.1 的现有数据模型

- [ ] **Step 1: 编写 SSD 第 4 章 — 数据设计**

```
4.1 ER 图（新增实体标注：orders, payment_records, admin_users, push_tokens, ad_config, events）
4.2 新增表结构（完整 DDL：字段名/类型/约束/索引）
4.3 数据库迁移方案（Goose up/down 脚本清单）
4.4 Redis 数据结构变更（新增 key pattern 说明）
```

- [ ] **Step 2: 编写 SSD 第 5 章 — API 设计**

```
5.1 新增 API 清单（按服务分组，含 HTTP method/path/request/response）
5.2 修改 API 清单（兼容性说明：哪些是破坏性变更）
5.3 OpenAPI 规范引用（自动生成路径说明）
```

- [ ] **Step 3: 验证并提交**

```bash
git add docs/SSD_AccountCenter_V2.0.0.md
git commit -m "docs: write SSD V2.0.0 sections 4-5 - data design and API design"
```

---

### Task 15: 编写 SSD 第 6-8 章 — 安全、部署、可观测性

**Files:**
- Modify: `docs/SSD_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 编写 SSD 第 6 章 — 安全设计**

```
6.1 密码哈希迁移方案
  - argon2id 参数配置（memory=64MB, iterations=3, parallelism=2）
  - 新注册直接使用 argon2id
  - 存量 SM3 用户登录时验证后 rehash 为 argon2id
  - password_hash 字段新增前缀标识（$argon2id$ / $sm3$）
  - SM3 保留用于审计日志完整性校验

6.2 KMS 集成方案
  - Vault/阿里云 KMS 选型对比
  - 密钥轮换策略
  - 审计记录

6.3 API 安全加固
  - 用户级限流（按等级差异化）
  - 请求签名验证
  - SQL 注入/XSS 自动化扫描

6.4 移动端安全
  - 证书固定（Certificate Pinning）
  - Root/越狱检测
  - 应用截屏防护
```

- [ ] **Step 2: 编写 SSD 第 7 章 — 部署设计**

```
7.1 多环境差异（引用 ITERATION 5.8.3 表格）
7.2 Helm Chart 结构（Chart.yaml / values.yaml / templates/）
7.3 CI/CD 流水线设计（GitHub Actions YAML 结构）
7.4 发布策略（蓝绿/金丝雀）
```

- [ ] **Step 3: 编写 SSD 第 8 章 — 可观测性设计**

```
8.1 OpenTelemetry 接入方案（Go SDK + W3C Trace Context + Jaeger/Tempo）
8.2 Grafana Dashboard 模板（服务健康总览 / API 延迟 / 错误率 / 业务指标）
8.3 告警规则定义（Prometheus AlertManager 规则 YAML）
8.4 日志规范（slog 字段：trace_id, span_id, request_id, service_name）
```

- [ ] **Step 4: 编写 SSD 第 9 章 — 附录**

```
9.1 与 SSD V1.3.1 差异对照表
9.2 配置项变更清单
```

- [ ] **Step 5: 最终验证**

全文搜索 `（待填充）`，确认无残留。文档状态改为"已完成"。

- [ ] **Step 6: 提交 SSD 完整版**

```bash
git add docs/SSD_AccountCenter_V2.0.0.md
git commit -m "docs: complete SSD V2.0.0 - all technical designs"
```

---

## Phase C: PLAN V2.0.0

### Task 16: 创建 PLAN 文档骨架

**Files:**
- Create: `docs/PLAN_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 创建文件并写入文档头部和完整目录结构**

按设计规格第 5 节定义的 PLAN 结构创建骨架：
1. 总览
2. Phase 6 — 上线准备冲刺（P0）
3. Phase 7 — 体验与增长（P1）
4. Phase 8 — 竞争力提升（P2）
5. Phase 9 — 长期规划（P3）
6. 风险管理
7. 附录

- [ ] **Step 2: 提交**

```bash
git add docs/PLAN_AccountCenter_V2.0.0.md
git commit -m "docs: create PLAN V2.0.0 skeleton"
```

---

### Task 17: 编写 PLAN 第 1 章 — 总览

**Files:**
- Modify: `docs/PLAN_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 编写总览章节**

```
1.1 迭代目标与范围
  - 覆盖 ITERATION V1.0.1 全部 77 项改进
  - 按 Phase 分阶段交付

1.2 团队配置
  - 1 人全职，串行执行

1.3 版本发布策略
  - V2.0 = Phase 6（P0）
  - V2.1 = Phase 7（P1）
  - V2.2 = Phase 8（P2）
  - V2.3 = Phase 9（P3）

1.4 执行原则
  - 依赖前置：被依赖的任务必须先完成
  - P0 优先：Phase 6 全部完成后才进入 Phase 7
  - 风险先行：同一 Phase 内高风险任务先做
  - 即时验证：完成即测试，测试通过即标注
```

- [ ] **Step 2: 提交**

---

### Task 18: 编写 PLAN 第 2 章 — Phase 6 任务分解

**Files:**
- Modify: `docs/PLAN_AccountCenter_V2.0.0.md`

**数据来源:**
- PRD V2.0.0 第 3.1 节（P0 需求）
- SSD V2.0.0 第 3.1 节（P0 技术设计）

- [ ] **Step 1: 编写 Phase 6 任务分解**

为 14 项 P0 需求分解为具体开发任务，每个任务包含：
- 任务描述
- 涉及文件（精确路径）
- 依赖的前置任务
- 验证方式

按执行顺序排列（依赖驱动）：
1. AR-25 仓库清理（无依赖）
2. NF-02 网关超时（无依赖）
3. NF-01 deletion-worker（无依赖）
4. AR-13 argon2id 迁移（无依赖）
5. AR-23 备份策略（无依赖）
6. AR-17 单元测试补齐（无依赖）
7. FN-02 订单管理系统（无依赖）
8. FN-01 支付网关（依赖 FN-02）
9. FN-05 Admin 后端（依赖 FN-02）
10. FN-10 APNs/FCM 推送（无依赖）
11. AR-21 K8s Helm Chart（无依赖）
12. AR-22 CI/CD 流水线（依赖 AR-21）
13. AR-18 集成测试（依赖 AR-17）
14. AR-16 渗透测试（依赖 AR-13, AR-17）
15. UX-08 定价透明度（无依赖）

- [ ] **Step 2: 编写 Phase 6 阶段产出物清单**

```
Phase 6 产出物：
├── 文档：PRD V2.0.0 + SSD V2.0.0（P0 标注已实现）
├── 代码：
│   ├── account-service: deletion-worker, argon2id migration
│   ├── api-gateway: timeout middleware, refactor
│   ├── payment-service: 新增微服务（订单+支付+回调）
│   ├── notification-service: APNs/FCM provider
│   ├── admin-api: 用户/订阅/风控/运营管理 API
│   └── .golangci.yml, .gitignore 更新
├── 基础设施：
│   ├── helm/（Helm Chart）
│   ├── .github/workflows/（CI/CD）
│   └── monitoring/（告警规则 + Dashboard）
├── 测试：单元测试 >60% + 集成测试 + 渗透测试报告
└── 变更日志：CHANGELOG_V2.0.md
```

- [ ] **Step 3: 提交**

```bash
git add docs/PLAN_AccountCenter_V2.0.0.md
git commit -m "docs: write PLAN V2.0.0 section 2 - Phase 6 task breakdown"
```

---

### Task 19: 编写 PLAN 第 3-5 章 — Phase 7/8/9 任务分解

**Files:**
- Modify: `docs/PLAN_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 编写 PLAN 第 3 章 — Phase 7（P1，30 项）**

30 项 P1 需求的任务分解，按依赖顺序排列。

- [ ] **Step 2: 编写 PLAN 第 4 章 — Phase 8（P2，29 项）**

29 项 P2 需求的任务分解。

- [ ] **Step 3: 编写 PLAN 第 5 章 — Phase 9（P3，4 项）**

4 项 P3 需求的任务分解。

- [ ] **Step 4: 提交**

```bash
git add docs/PLAN_AccountCenter_V2.0.0.md
git commit -m "docs: write PLAN V2.0.0 sections 3-5 - Phase 7/8/9 task breakdown"
```

---

### Task 20: 编写 PLAN 第 6-7 章 — 风险管理与附录

**Files:**
- Modify: `docs/PLAN_AccountCenter_V2.0.0.md`

- [ ] **Step 1: 编写第 6 章 — 风险管理**

```
6.1 已知风险与缓解措施
  - 支付网关审核周期不确定 → 提前申请商户号
  - APNs/FCM 证书/密钥管理复杂 → 使用 provider SDK
  - argon2id 迁移影响存量用户登录 → 渐进式 rehash
  - K8s 生产部署复杂度高 → 先 Dev 环境验证

6.2 依赖外部服务清单
  - 微信支付商户号
  - 支付宝商户号
  - APNs 证书
  - FCM 项目配置
  - 广告 SDK 账号（穿山甲/优量汇）
  - 阿里云 KMS / Vault 集群
```

- [ ] **Step 2: 编写第 7 章 — 附录**

```
7.1 全量任务总表
  77 项需求 × 执行顺序 × 依赖关系 × 涉及文件的汇总表

7.2 需求→任务→代码文件映射表
  按 ITERATION ID 索引，追溯从需求到具体代码文件的完整路径
```

- [ ] **Step 3: 最终验证**

全文搜索 `（待填充）`，确认无残留。文档状态改为"已完成"。

- [ ] **Step 4: 提交 PLAN 完整版**

```bash
git add docs/PLAN_AccountCenter_V2.0.0.md
git commit -m "docs: complete PLAN V2.0.0 - full task decomposition P0-P3"
```

---

## 最终验证

### Task 21: 全局一致性检查

- [ ] **Step 1: 验证 PRD↔SSD 一致性**

PRD 中每个需求 ID 在 SSD 中都有对应的技术设计章节。

- [ ] **Step 2: 验证 PRD↔PLAN 一致性**

PRD 中每个需求 ID 在 PLAN 中都有对应的任务分解。

- [ ] **Step 3: 验证无占位符残留**

三份文档全文搜索 `（待填充）`、`TBD`、`TODO`，确认均为 0 结果。

- [ ] **Step 4: 最终提交**

```bash
git add docs/PRD_AccountCenter_V2.0.0.md docs/SSD_AccountCenter_V2.0.0.md docs/PLAN_AccountCenter_V2.0.0.md
git commit -m "docs: finalize V2.0 documentation suite - PRD + SSD + PLAN V2.0.0"
```
