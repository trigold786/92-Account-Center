# Sprint 3 设计规格 — 数据产品 + 动态脱敏 + VictoriaMetrics

> **版本**: v1.0.0
> **日期**: 2026-05-14
> **状态**: Draft
> **前置**: Sprint 1 (服务合并) + Sprint 2 (返利引擎 + 风控) 已完成并推送

---

## 1. 范围

Sprint 3 完成 Account Center 商业化迭代的最后核心功能：

| # | 功能 | 服务 | 优先级 |
|---|------|------|--------|
| 1 | RFM 用户画像 + 防刷监控大盘 + 订阅漏斗 | data-product-service | P0 |
| 2 | 动态脱敏中间件 | api-gateway | P0 |
| 3 | VictoriaMetrics /metrics 端点 + 容器 | 全部 7 服务 + VM 容器 | P0 |

**不在本 Sprint 范围**:
- ECS 部署 (推至 Sprint 4)
- QR 扫码登录 (已在 Sprint 1 完成)

---

## 2. data-product-service 设计

### 2.1 架构

直连 PostgreSQL 查询，不调用其他服务 API。

```
data-product-service/
  cmd/main.go
  internal/
    handler/
      rfm_handler.go
      dashboard_handler.go
      funnel_handler.go
    service/
      rfm_service.go
      dashboard_service.go
    repository/
      data_repository.go
    model/
      rfm.go
  go.mod
  Dockerfile
```

### 2.2 RFM 5 分制 8 类算法

**数据源**: `subscriptions` 表

**R (Recency)** — 最近订阅距今天数:
- ≤30 天 = 5
- ≤60 天 = 4
- ≤90 天 = 3
- ≤180 天 = 2
- >180 天 = 1

**F (Frequency)** — 订阅总次数:
- ≥10 次 = 5
- ≥5 次 = 4
- ≥3 次 = 3
- ≥2 次 = 2
- 1 次 = 1

**M (Monetary)** — 累计消费金额:
- ≥1000 = 5
- ≥500 = 4
- ≥200 = 3
- ≥100 = 2
- <100 = 1

**8 类客户映射**:

| R | F | M | 类型 | 英文键 |
|---|---|---|------|--------|
| ≥4 | ≥4 | ≥4 | 重要价值客户 | CHAMPION |
| ≥4 | <4 | ≥4 | 重要发展客户 | PROMISING |
| <4 | ≥4 | ≥4 | 重要保持客户 | LOYAL |
| <4 | <4 | ≥4 | 重要挽留客户 | AT_RISK |
| ≥4 | ≥4 | <4 | 一般价值客户 | POTENTIAL_LOYAL |
| ≥4 | <4 | <4 | 一般发展客户 | NEW |
| <4 | ≥4 | <4 | 一般保持客户 | NEED_ATTENTION |
| <4 | <4 | <4 | 一般挽留客户 | ABOUT_TO_LOSE |

**去标识化**: RFM API 只返回 `user_id` + 分数 + 类型，不返回手机号/邮箱/姓名等 PII。

### 2.3 API 端点

#### RFM 画像

```
GET /api/v1/data/rfm/:user_id
```

响应:
```json
{
  "code": 200,
  "data": {
    "user_id": 6,
    "recency_score": 5,
    "frequency_score": 2,
    "monetary_score": 3,
    "rfm_segment": "POTENTIAL_LOYAL",
    "rfm_segment_cn": "一般价值客户",
    "last_subscription_at": "2026-05-14T05:36:27Z",
    "total_subscriptions": 2,
    "total_spent": 200.00
  }
}
```

#### RFM 批量查询

```
POST /api/v1/data/rfm/batch
Body: {"user_ids": [1, 6, 7]}
```

#### 监控大盘

```
GET /api/v1/data/dashboard/overview
```

响应:
```json
{
  "code": 200,
  "data": {
    "total_users": 7,
    "total_subscriptions": 4,
    "total_credits_earned": 65.00,
    "total_credits_consumed": 20.00,
    "blacklist_entries_active": 1,
    "registration_trend": [
      {"date": "2026-05-14", "count": 7}
    ],
    "credit_flow": {
      "EARN_REFERRAL": 60.00,
      "CONSUME_SUB": 20.00,
      "REFUND_SUB": 5.00
    },
    "rfm_distribution": {
      "CHAMPION": 0,
      "PROMISING": 1,
      "LOYAL": 0,
      "AT_RISK": 0,
      "POTENTIAL_LOYAL": 1,
      "NEW": 3,
      "NEED_ATTENTION": 0,
      "ABOUT_TO_LOSE": 2
    }
  }
}
```

#### 订阅漏斗

```
GET /api/v1/data/funnel/subscription
```

响应:
```json
{
  "code": 200,
  "data": {
    "steps": [
      {"name": "注册用户", "count": 7, "percentage": 100.0},
      {"name": "实名用户 (L1+)", "count": 0, "percentage": 0.0},
      {"name": "订阅用户 (L2+)", "count": 2, "percentage": 28.6},
      {"name": "高级订阅 (L3+)", "count": 0, "percentage": 0.0},
      {"name": "顶级订阅 (L4)", "count": 0, "percentage": 0.0}
    ]
  }
}
```

### 2.4 数据库查询

**RFM 单用户**:
```sql
SELECT 
  COUNT(*) as freq,
  COALESCE(SUM(price), 0) as monetary,
  MAX(end_time) as last_sub
FROM subscriptions WHERE user_id = $1
```

**大盘统计**:
```sql
SELECT DATE(created_at), COUNT(*) FROM users GROUP BY DATE(created_at) ORDER BY DATE(created_at) DESC LIMIT 30;
SELECT type, SUM(amount) FROM credit_transactions GROUP BY type;
SELECT COUNT(*) FROM blacklist_entries WHERE created_at >= NOW() - interval '24h';
```

**订阅漏斗**:
```sql
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM users WHERE identity_tier >= 1;
SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'ACTIVE';
SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE tier_level >= 3 AND status = 'ACTIVE';
```

---

## 3. 动态脱敏中间件

### 3.1 位置

api-gateway `cmd/main.go` 新增:
- `responseCaptureWriter` — 包装 `gin.ResponseWriter`，捕获响应 body
- `desensitizeMiddleware()` — Gin 中间件

### 3.2 工作流程

```
请求进入 → JWT 解析 → c.Next() (代理到后端) → 捕获响应 body
  ↓
Content-Type == application/json?
  ↓ 是
管理员角色?
  ↓ 否
正则替换敏感字段 → 写回响应
```

### 3.3 脱敏规则

| 字段类型 | 正则 | 替换示例 |
|---------|------|---------|
| 手机号 | `"phone_number"\s*:\s*"(\d{3})\d{4}(\d{4})"` | `"phone_number":"138****1234"` |
| 邮箱 | `"email"\s*:\s*"([a-zA-Z0-9])[a-zA-Z0-9._%+-]*@([^"]+)"` | `"email":"t***@example.com"` |
| IP 地址 | `"ip_address"\s*:\s*"(\d{1,3}\.)\d{1,3}\.\d{1,3}(\.\d{1,3})"` | `"ip_address":"192.*.*.1"` |

### 3.4 豁免规则

- `/health` 路径不脱敏
- `/internal/*` 路径不脱敏（服务间内部调用）
- JWT claims 中 `account_id` 以 `admin_` 开头的管理员不脱敏
- HTTP 状态码非 200 不脱敏
- 响应 body 超过 1MB 不脱敏（性能保护）

### 3.5 响应头标记

脱敏后的响应添加 `X-Desensitized: true` 头，方便调试。

---

## 4. VictoriaMetrics 集成

### 4.1 /metrics 端点

所有 7 个服务添加 `/metrics` 端点，输出 Prometheus exposition format。

**纯 Go 标准库实现** — 不引入额外依赖 (符合最小化原则):

```go
r.GET("/metrics", metricsHandler("service-name"))
```

每个服务使用 `sync/atomic` 维护内存计数器:
- `http_requests_total{service,method,path,status}` — counter
- `http_request_duration_seconds_sum{service}` — summary sum
- `http_request_duration_seconds_count{service}` — summary count
- `go_goroutines` — gauge (从 `runtime.NumGoroutine()`)

**实现模式**: 创建统一的 metrics 中间件模板，每个服务复制适配。中间件在 Gin Router 级别统计所有请求。

### 4.2 VictoriaMetrics 容器

Docker Compose 新增:

```yaml
victoriametrics:
  image: victoriametrics/victoria-metrics:latest
  container_name: victoriametrics
  ports:
    - "20010:8428"
  volumes:
    - vm_data:/victoria-metrics-data
    - ./monitoring/promscrape.yml:/etc/promscrape.yml
  command:
    - "--promscrape.config=/etc/promscrape.yml"
    - "--storageDataPath=/victoria-metrics-data"
    - "--retentionPeriod=30d"
  networks:
    - app_network
  deploy:
    resources:
      limits:
        cpus: '0.25'
        memory: 256M
  restart: always
```

### 4.3 抓取配置

`monitoring/promscrape.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'account-center'
    static_configs:
      - targets:
          - 'api-gateway:30300'
          - 'account-service:30301'
          - 'auth-service:30302'
          - 'notification-service:30311'
          - 'credit-service:30312'
          - 'compliance-service:30313'
          - 'data-product-service:30314'
```

### 4.4 验证

- `GET http://localhost:20010/api/v1/targets` — 确认所有 targets UP
- `GET http://localhost:20010/api/v1/query?query=http_requests_total` — 确认指标可查
- 各服务 `GET /metrics` — 确认 Prometheus 格式输出

---

## 5. Docker Compose 变更

| 变更项 | 说明 |
|--------|------|
| 新增 `victoriametrics` 服务 | VM 单节点容器，端口 20010 |
| 新增 `monitoring/promscrape.yml` | VM 抓取配置 |
| 新增 `vm_data` volume | VM 数据持久化 |
| `data-product-service` 增加环境变量 | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` |
| 总容器数 | 9 → 10 (含 VM) |

---

## 6. 验收标准

1. `GET /api/v1/data/rfm/:user_id` 返回正确的 RFM 分数和客户分类
2. `GET /api/v1/data/dashboard/overview` 返回聚合统计数据
3. `GET /api/v1/data/funnel/subscription` 返回漏斗转化率
4. 通过 api-gateway 查询用户信息时，非管理员看到脱敏数据 (手机号 `138****1234`)
5. 管理员 (account_id `admin_*`) 看到原始数据
6. 所有 7 个服务 `/metrics` 端点返回 Prometheus 格式指标
7. VictoriaMetrics 容器运行，`/api/v1/targets` 显示所有 7 个 targets UP
8. 所有容器 healthy (`docker compose ps`)
9. 全链路 E2E 测试通过

---

## 7. 实施任务

### Task 1: data-product-service RFM + 统计 API
- 新增 `internal/` 目录结构 (handler/service/repository/model)
- 实现 RFM 5 分制计算 + 8 类客户分类
- 实现监控大盘 API (注册趋势/风控统计/积分流通/RFM 分布)
- 实现订阅漏斗 API
- 更新 `cmd/main.go` 注册路由
- Docker Compose 添加 DB 环境变量

### Task 2: 动态脱敏中间件 (api-gateway)
- 新增 `responseCaptureWriter`
- 新增 `desensitizeMiddleware()`
- 在 `proxyHandler` 中应用中间件
- 管理员豁免逻辑

### Task 3: VictoriaMetrics /metrics 端点 + 容器
- 创建 metrics 中间件模板 (sync/atomic 计数器)
- 所有 7 个服务添加 `/metrics` 端点
- 创建 `monitoring/promscrape.yml`
- Docker Compose 新增 victoriametrics 服务

### Task 4: Docker 重建 + E2E 验证
- 重建所有受影响的服务
- 验证 VMetrics 采集
- 验证脱敏效果
- 验证 RFM/大盘/漏斗 API
- 全链路 E2E 测试
