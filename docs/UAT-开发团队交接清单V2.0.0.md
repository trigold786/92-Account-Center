# UAT 环境 — 开发团队交接清单

> **版本**: V2.0.0
> **日期**: 2026-05-21
> **ECS**: 101.133.168.46 | **目标**: Account Center V2.0 UAT 环境
> **变更**: V1.0.0→V2.0.0 开发团队回复全部 12 项交办事宜，修复 4 项，需 DevOps 配合 4 项

---

## 已完成的基础设施

| 组件 | 状态 | 端口 | 说明 |
|------|:----:|:----:|------|
| PostgreSQL 18.4 | ✅ | 20002 | systemd 服务，多项目共享 |
| Redis 8.2.6 | ✅ | 20003 | systemd 服务，AOF 持久化 |
| VictoriaMetrics v1.143.0 | ✅ | 20010 | systemd 服务，promscrape 已配置 |
| Loki 3.7.2 | ✅ | 20007 | systemd 服务 |
| Grafana 13.0.1 | ✅ | 20006 | sub_path: /grafana/ |
| Promtail | ✅ | - | systemd 服务 |
| MinIO | ✅ | 20004(S3)/20005(Console)| Docker 容器 |
| Traefik v3.3 | ✅ | 80/443 | edge router，Let's Encrypt 就绪 |
| Jaeger all-in-one | ✅ | 4317(OTLP)/16686(UI)| Docker 容器 |
| pg_dump 每日备份 | ✅ | - | cron 2am，7 天保留 |
| pgvector 0.8.2 | ✅ | - | 已安装到全部数据库 |
| 自签名 SSL 证书 | ✅ | - | /opt/.../configs/ssl/ |

---

## 需要开发团队处理的事项

### 🔴 高优先级（影响 UAT 功能完整性）

| # | 事项 | 说明 | 影响 | 操作指引 |
|:-:|:-----|:-----|:-----|:---------|
| 1 | **web-ui 前端容器部署** | ECS 上无 web-ui 构建产物，需开发团队提供 `dist/` 目录或配置 CI 推送 | 前端页面不可访问 | 构建后上传到 ECS，`docker compose up -d web-ui` |
| 2 | **config-management-ui 容器化** | 管理后台缺少 Docker 部署配置 | 管理功能不可用 | 参考 `docker-compose.yml` 添加服务定义 |
| 3 | **Go 微服务注册 /metrics 端点** | VM scrape 返回 404，Gin 路由未注册 `/metrics` | 业务指标不可见 | 在 `main.go` 中增加 `r.GET("/metrics", gin.WrapH(promhttp.Handler()))` |
| 4 | **AlertManager WebHook 真实 key** | alertmanager.yml 中企微/钉钉 Webhook URL 为占位符 | 告警通知不生效 | 替换 3 个 receiver 的真实 Webhook URL |

### 🟡 中优先级（建议上线前完成）

| # | 事项 | 说明 | 操作指引 |
|:-:|:-----|:------|:---------|
| 5 | **CI GO_VERSION 1.24→1.26** | `.github/workflows/ci.yml` 中 GO_VERSION 写死 1.24，与项目 Go 1.26 不符 | 修改 L15 为 `GO_VERSION: "1.26"` |
| 6 | **docker-compose.prod.yml 验证** | 确认 w004 项目的 prod overlay 正确连接了共享 PG(20002)/Redis(20003) | 对比各服务的 `environment` 段，确保 DB/REDIS 指向 `127.0.0.1` |
| 7 | **prometheus scrape 补全** | 若后续新增服务，需在 `vm.yml` 的 `w004_services` target 中添加 | 编辑 `/opt/.../configs/victoriametrics/vm.yml` |
| 8 | **Docker socket 权限** | VM 以 nobody 运行，无法 Docker SD 发现。可选：a) 加 `-promscrape.config.strictParse=false` 移除 docker_sd，b) 给 VM 加 docker 组权限 | 当前不影响基础 scrape，仅容器发现缺失 |

### 🟢 低优先级（可后续迭代）

| # | 事项 | 说明 |
|:-:|:-----|:------|
| 9 | **Helm 多环境 values 文件** | 创建 `values-dev.yaml` / `values-uat.yaml` / `values-prod.yaml`，需开发团队提供 K8s 集群信息 |
| 10 | **Helm servicemonitor.yaml** | 模板中缺少 servicemonitor，K8s 部署后需补充 |
| 11 | **证书信任** | 自签名 CA 证书 (`ca.crt`) 需分发给各端（浏览器/iOS/Android）导入信任链 |
| 12 | **DNS 解析** | `uat-91.neurongene.cn` → `101.133.168.46` 需配置 DNS A 记录后才能启用 Let's Encrypt |

---

## 关键配置参考

```bash
# 共享服务连接信息
PG_HOST=127.0.0.1   PG_PORT=20002   PG_USER=platform_admin
REDIS_HOST=127.0.0.1  REDIS_PORT=20003
GRAFANA_URL=http://127.0.0.1:20006/grafana/
VM_URL=http://127.0.0.1:20010
LOKI_URL=http://127.0.0.1:20007
JAEGER_OTLP=127.0.0.1:4317
MINIO_S3=http://127.0.0.1:20004  MINIO_CONSOLE=http://127.0.0.1:20005
TRAEFIK_DASHBOARD=http://127.0.0.1:20001

# 登录凭据（首次）
Grafana: admin / admin
Traefik: admin / Traefik2026!
MinIO:   minioadmin / dfj9NpUzc572XAYo1imPwner
PostgreSQL: platform_admin / dfj9NpUzc572XAYo1imPwner
```

---

## 开发团队回复

> **版本**: V2.0.0  
> **日期**: 2026-05-21  
> **回复人**: 开发团队  

### ✅ 已完成且满足需求（6项）

以下事项已确认完毕，无需进一步操作：

| # | 事项 | 验证结果 | 说明 |
|:-:|:-----|:---------|:------|
| 1 | **web-ui 前端容器部署** | ✅ **已满足，无需额外操作** | Web UI 已有 `dist/` 目录（含 `index.html` + `assets/`），Docker 镜像为 `nginx:alpine`。部署方式：`docker compose up -d web-ui`，Vue 3 构建产物已就绪。若需单独构建：`cd web-ui && npm ci && npm run build` |
| 3 | **Go 微服务注册 /metrics 端点** | ✅ **已满足，VM 直接可用** | 全部 9 个微服务均已在 `main.go` 中注册 `r.Any("/metrics", ...)`，返回 Prometheus 格式指标（`http_requests_total`、`go_goroutines`、请求延迟等），Content-Type `text/plain; version=0.0.4`。运行中服务已验证全部返回 HTTP 200。VM scrape 配置于 `monitoring/promscrape.yml`，若使用 systemd 部署请确保 vm.yml 中包含各服务端口 |
| 5 | **CI GO_VERSION 1.24→1.26** | ✅ **已修复** | `.github/workflows/ci.yml` L23 `GO_VERSION` 已更新为 `"1.26"`，与项目 go 1.26 一致 |
| 7 | **prometheus scrape 补全** | ✅ **已更新** | `monitoring/promscrape.yml` 已补全 `payment-service:30316` 目标。若 ECS 使用独立 vm.yml，请在 `w004_services` target 中添加：`- 'payment-service:30316'` |
| 8 | **Docker socket 权限** | ✅ **已确认** | 当前静态配置的 promscrape.yml 已足够，无需 Docker SD。VM 以 nobody 运行不影响基础指标采集。若后续需容器自动发现，可选择给 VM 添加 docker 组权限 |
| 11 | **证书信任** | ✅ **已确认** | 自签名 CA 证书 `ca.crt` 开发团队已接收。各端导入指引：浏览器→导入受信任根证书颁发机构；iOS→通过 MDM 或描述文件安装；Android→设置→安全→安装证书 |

### ⚠️ 需要 DevOps 配合处理（4项）

| # | 事项 | 开发团队回复 | 需要 DevOps |
|:-:|:-----|:------------|:------------|
| 4 | **AlertManager WebHook 真实 key** | 🔴 **请提供真实 Webhook URL** | alertmanager.yml 中占位 key 需要替换。请提供以下信息：① 企业微信 Webhook URL（替换 `your_key`）② 钉钉 Webhook URL（替换 `your_token`）。提供后开发团队将在 1 个工作日内更新配置 |
| 6 | **docker-compose.prod.yml 验证** | 🟡 **需 DevOps 确认连接信息** | 项目当前使用 `docker-compose.yml` 作为 Dev 配置。UAT ECS 需覆盖共享服务连接（PG:20002 / Redis:20003）。开发团队提供了 `.env.example` 环境变量模板，请 DevOps 按以下方式部署：① 复制 `.env.example` 为 `.env` ② 修改 DB/REDIS 连接指向 `127.0.0.1:20002/20003` ③ 执行 `docker compose up -d`。无需额外 `prod.yml` 文件 |
| 10 | **Helm servicemonitor.yaml** | 🟡 **待 K8s 集群就绪后补充** | Helm Chart `templates/` 目前缺少 servicemonitor.yaml。已在 SSD 中完成设计，当 UAT K8s 集群就绪后，开发团队可参考 SSD Section 7.2 补充此模板。当前 Docker Compose 环境无需此配置 |
| 12 | **DNS 解析** | 🟡 **请 DevOps 配置 DNS A 记录** | 请为 `uat-91.neurongene.cn` 配置 DNS A 记录指向 `101.133.168.46`。配置完成后通知开发团队，将协助验证 Let's Encrypt 证书签发和 Traefik HTTPS 路由 |

### 📋 开发团队自行完成（2项）

| # | 事项 | 完成情况 |
|:-:|:-----|:---------|
| 2 | **config-management-ui 容器化** | ✅ **已完成** — 已创建 Dockerfile 和 docker-compose 条目（见下方新增配置说明）。管理后台可通过 `docker compose up -d config-management-ui` 部署，端口 `30316` |
| 9 | **Helm 多环境 values 文件** | ✅ **已完成** — 已在 `helm/account-center/` 下创建 `values-dev.yaml`。`values-uat.yaml` 和 `values-prod.yaml` 需在 K8s 集群信息明确后补充 |

### 新增说明

#### Item 2 — config-management-ui Docker 部署配置

`docker-compose.yml` 已新增 config-management-ui 服务：

```yaml
config-management-ui:
  build:
    context: ./config-management-ui
    dockerfile: Dockerfile
  container_name: config-management-ui
  ports:
    - "30316:30316"
  networks:
    - app_network
  depends_on:
    config-service:
      condition: service_healthy
  healthcheck:
    test: ["CMD", "wget", "--spider", "-q", "http://127.0.0.1:30316/health"]
    interval: 30s
    timeout: 3s
    retries: 3
    start_period: 5s
  deploy:
    resources:
      limits:
        cpus: '0.25'
        memory: 256M
  restart: always
```

#### Item 9 — Helm values-dev.yaml

```yaml
global:
  environment: development
  imageRegistry: ghcr.io/your-org/account-center
  imagePullSecrets: []
  storageClass: ""

postgresql:
  enabled: false

redis:
  enabled: false

ingress:
  enabled: false
```

### UAT ECS 完整部署步骤

```bash
# 1. 克隆代码
git clone https://github.com/trigold786/92-Account-Center.git
cd 92-Account-Center

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env：
#   DB_HOST=127.0.0.1  DB_PORT=20002  DB_PASSWORD=dfj9NpUzc572XAYo1imPwner
#   REDIS_ADDR=127.0.0.1:20003
#   JWT_ACCESS_SECRET=<替换为实际密钥>
#   JWT_REFRESH_SECRET=<替换为实际密钥>
#   JWT_SECRET=<替换为实际密钥>

# 3. 构建并启动
docker compose build --no-cache
docker compose up -d

# 4. 验证
for port in 30300 30302 30301 30312 30313 30311 30314 30315 30316; do
  curl -s "http://localhost:$port/health" && echo " :$port OK" || echo " :$port FAIL"
done
curl -s http://localhost:30317/ | head -c 100  # web-ui
curl -s http://localhost:30316/ | head -c 100  # config-ui
```

### 各微服务 metrics 验证确认

全部 9 个微服务均注册了 `/metrics` 端点，返回 Prometheus-format 指标：

| 服务 | 端口 | metrics 状态 | 指标内容 |
|------|:----:|:-----------:|---------|
| api-gateway | 30300 | ✅ HTTP 200 | http_requests_total, go_goroutines |
| account-service | 30301 | ✅ HTTP 200 | http_requests_total, http_request_duration, go_goroutines |
| auth-service | 30302 | ✅ HTTP 200 | 同上 |
| notification-service | 30311 | ✅ HTTP 200 | 同上 |
| credit-service | 30312 | ✅ HTTP 200 | 同上 |
| compliance-service | 30313 | ✅ HTTP 200 | 同上 |
| data-product-service | 30314 | ✅ HTTP 200 | 同上 |
| config-service | 30315 | ✅ HTTP 200 | 同上 |
| payment-service | 30316 | ✅ 代码已实现 | 同上（E2E 已验证） |

验证命令：`curl -s http://localhost:<PORT>/metrics | head -5`

### 已修复的问题清单

| 修复项 | 文件 | 变更内容 | 状态 |
|:-------|:-----|:---------|:----:|
| CI GO_VERSION | `.github/workflows/ci.yml` L23 | `"1.24"` → `"1.26"` | ✅ 已推送 |
| promscrape 补全 | `monitoring/promscrape.yml` | 添加 `payment-service:30316` | ✅ 已推送 |
| config-ui Docker | `config-management-ui/Dockerfile` + `docker-compose.yml` | 新增服务定义 | ✅ 已推送 |
| Helm values-dev | `helm/account-center/values-dev.yaml` | 新增 Dev 环境 values | ✅ 已推送 |
