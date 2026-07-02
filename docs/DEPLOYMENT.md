# 部署指南

## 环境要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Docker | 20.10+ | 容器引擎 |
| Docker Compose | 2.0+ | 容器编排 |
| PostgreSQL | 15+ | 主数据库 |
| Redis | 7+ | 缓存/会话 |

## 快速部署

### 1. 克隆项目

```bash
git clone <repository-url>
cd account-center
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置
vim .env
```

### 3. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 验证部署

```bash
# 健康检查
curl http://localhost:30300/health

# 查看所有服务状态
docker-compose ps
```

## 服务端口

| 服务 | 端口 | 内部网络 |
|------|------|----------|
| API Gateway | 30300 | 外部访问 |
| Account Service | 30301 | 内部 |
| Auth Service | 30302 | 内部 |
| Notification Service | 30311 | 内部 |
| Credit Service | 30312 | 内部 |
| Compliance Service | 30313 | 内部 |
| Data Product Service | 30314 | 内部 |
| Config Service | 30315 | 内部 |
| Payment Service | 30316 | 内部 |
| Web UI | 30317 | 外部访问 |
| Config Management UI | 30318 | 外部访问 |
| PostgreSQL | 5432 | 仅内部 |
| Redis | 6379 | 仅内部 |

## 配置说明

### 数据库初始化

数据库初始化脚本位于 `migrations/001_initial_schema.sql`，会在 Docker Compose 启动时自动执行。

### 环境变量

```bash
# 数据库
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=account_center

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# JWT 密钥 (生产环境必须修改)
JWT_ACCESS_SECRET=your-access-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key
JWT_SECRET=your-jwt-secret-key

# 服务 URL
CONFIG_SERVICE_URL=http://config-service:30315
AUTH_SERVICE_URL=http://auth-service:30302
ACCOUNT_SERVICE_URL=http://account-service:30301
CREDIT_SERVICE_URL=http://credit-service:30312
COMPLIANCE_SERVICE_URL=http://compliance-service:30313
NOTIFICATION_SERVICE_URL=http://notification-service:30311
DATA_PRODUCT_SERVICE_URL=http://data-product-service:30314
PAYMENT_SERVICE_URL=http://payment-service:30316

# 短信服务商 (根据需要配置)
ALIYUN_ACCESS_KEY_ID=xxx
ALIYUN_ACCESS_KEY_SECRET=xxx
ALIYUN_SIGN_NAME=AccountCenter
TENCENT_APP_ID=xxx
TENCENT_APP_SECRET=xxx

# 邮件服务商 (根据需要配置)
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.163.com
SMTP_PORT=465
SMTP_USERNAME=xxx
SMTP_PASSWORD=xxx

# 支付服务 (根据需要配置)
PAYMENT_MODE=sandbox
WECHAT_APP_ID=xxx
WECHAT_MCH_ID=xxx
WECHAT_API_KEY=xxx
ALIPAY_APP_ID=xxx
ALIPAY_PRIVATE_KEY=xxx
ALIPAY_PUBLIC_KEY=xxx

# Grafana
GF_SECURITY_ADMIN_PASSWORD=admin
```

## 扩展部署

### 水平扩展

```bash
# 扩展 account-service 到 3 个实例
docker-compose up -d --scale account-service=3

# 扩展 auth-service 到 3 个实例
docker-compose up -d --scale auth-service=3
```

### 负载均衡

API Gateway 自动对扩展的服务进行负载均衡。

### 健康检查

每个服务都实现了健康检查端点：
- `/health` - 返回 `{"status": "ok"}`

## 更新服务

```bash
# 拉取最新代码
git pull

# 重新构建并重启
docker-compose up -d --build

# 重启特定服务
docker-compose up -d --build account-service
```

## 监控

### 日志

```bash
# 查看特定服务日志
docker-compose logs -f account-service

# 查看错误日志
docker-compose logs -f --tail=100 | grep ERROR
```

### 清理

```bash
# 停止所有服务
docker-compose down

# 删除数据卷 (慎用)
docker-compose down -v

# 清理未使用的镜像
docker system prune -f
```

## 生产环境部署

### 注意事项

1. **修改默认密钥**
   - JWT 密钥必须更换
   - 数据库密码必须更换
   - Redis 密码必须设置

2. **启用 HTTPS**
   - 使用 Nginx/Caddy 配置 SSL
   - 或使用云负载均衡器

3. **配置备份**
   - PostgreSQL 定期备份
   - Redis 数据定期备份

4. **资源限制**
   ```yaml
   # docker-compose.yml
   services:
     account-service:
       deploy:
         resources:
           limits:
             memory: 512M
             cpus: '0.5'
   ```

5. **网络隔离**
   - 生产环境使用 Docker Swarm 或 Kubernetes
   - 配置网络策略

### Kubernetes 部署

项目提供 Helm Chart 用于 Kubernetes 部署：

```bash
helm install account-center ./helm/account-center
```

## 故障排除

### 服务无法启动

```bash
# 检查日志
docker-compose logs <service-name>

# 检查端口占用
netstat -tlnp | grep 30300
```

### 数据库连接失败

```bash
# 检查 PostgreSQL
docker-compose exec postgres pg_isready

# 进入 PostgreSQL
docker-compose exec postgres psql -U postgres -d account_center
```

### Redis 连接失败

```bash
# 检查 Redis
docker-compose exec redis redis-cli ping
```
