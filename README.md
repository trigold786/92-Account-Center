# Account Center Microservice

企业级账户中心微服务系统，为 Neuro 系列产品提供统一的用户认证、账户管理、风控、信用、数据产品、通知及配置管理。系统符合**等保三级 (MLPS 3.0)** 安全标准及 **Neuro 开发 SOP (L1-L6)** 规范。

## 系统架构

```
                         +-------------+
                         |   Client    |
                         | (Web/Mobile)|
                         +------+------+
                                |
                         +------v------+
                         | api-gateway |
                         |  (30300)    |
                         +------+------+
                                |
          +---------------------+----------------------+
          |                     |                      |
   +------v------+      +------v------+      +--------v--------+
   |auth-service |      |account-svc  |      | compliance-svc  |
   |  (30302)    |      |  (30301)    |      |   (30304)       |
   +------+------+      +------+------+      +--------+--------+
          |                     |                      |
   +------v------+      +------v------+      +--------v--------+
   |notification |      |credit-svc   |      | data-product-svc|
   |  (30303)    |      |  (30305)    |      |   (30306)       |
   +-------------+      +------+------+      +--------+--------+
                                |
                         +------v------+
                         |config-service|
                         |  (30315)     |
                         +------+-------+
                                |
                     +----------+---------+
                     |                    |
              +------v------+     +-------v--------+
              | PostgreSQL  |     |     Redis      |
              |   (20002)   |     |    (20003)     |
              +-------------+     +----------------+
```

## 核心服务

| 服务 | 端口 | 目录 | 功能说明 |
|------|------|------|----------|
| **API Gateway** | 30300 | [api-gateway](api-gateway/) | API 网关、请求路由、JWT 认证、限流熔断、CORS |
| **Account Service** | 30301 | [account-service](account-service/) | 用户注册、密码修改、账户注销、账户分层、权益、订阅 |
| **Auth Service** | 30302 | [auth-service](auth-service/) | 统一认证、会话管理、设备管理、QR 码登录 |
| **Notification Service** | 30303 | [notification-service](notification-service/) | SMS/Email OTP、Magic Link、Push 通知 |
| **Compliance Service** | 30304 | [compliance-service](compliance-service/) | 风险评估、黑名单管理、KYB、审计追踪 |
| **Credit Service** | 30305 | [credit-service](credit-service/) | 信用账户、交易管理、推荐奖励 |
| **Data Product Service** | 30306 | [data-product-service](data-product-service/) | RFM 评分、数据面板、漏斗分析 |
| **Config Service** | 30315 | [config-service](config-service/) | 配置管理、版本发布、权限管理 |

### 安全特性

- **SM4 加密**：敏感数据（身份证号、银行账号）全程加密存储
- **SM3 哈希**：密码哈希、审计日志完整性验证
- **JWT 认证**：Access Token + Refresh Token 双令牌
- **MFA/TOTP**：多因子认证（RFC 6238 标准）
- **设备指纹**：FingerprintJS 集成、可信设备免验
- **会话管理**：Redis 存储、滑动窗口超时、并发限制

### 合规标准

- 等保三级 (MLPS 3.0)
- 国密算法 SM3/SM4
- 180 天审计日志保留
- 敏感操作审计追踪

## 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.21+ (本地开发)
- PostgreSQL 15+
- Redis 7+
- Kafka (可选，用于审计日志异步收集)

### 配置凭证

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 填写您自己的凭证（切勿将 .env 提交到版本控制）
# 关键配置：JWT_SECRET、ALIYUN_ACCESS_KEY_ID/SECRET、SMTP_PASSWORD 等
```

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看网关日志
docker-compose logs -f api-gateway
```

### 访问端点

- API Gateway: http://localhost:30300
- Health Check: http://localhost:30300/health

## 端口规范 (SOP L1 §1.1)

遵循 **Neuro 开发 SOP 一号文件**端口规范：

| 范围 | 用途 | 说明 |
|------|------|------|
| 20000-29999 | 基础设施 | PostgreSQL (20002)、Redis (20003)、MinIO (20004) 等 |
| 30300-30399 | **本项目** | Account Center (w004 项目段) |
| 30000-49999 | 项目服务 | 按项目分配，每项目 +100 端口 |

## 项目结构

```
account-center/
├── api-gateway/                    # API 网关 (30300)
├── account-service/                # 账户管理 (30301)
├── auth-service/                   # 认证服务 (30302)
├── notification-service/           # 通知服务 (30303)
├── compliance-service/             # 合规风控 (30304)
├── credit-service/                 # 信用服务 (30305)
├── data-product-service/           # 数据产品 (30306)
├── config-service/                 # 配置管理 (30315)
├── config-management-ui/           # 配置管理前端
├── pkg/                            # 共享包
├── db-migrations/                  # 数据库迁移脚本
├── migrations-archived/            # 旧版迁移脚本（已归档）
├── docs/                           # 架构、API、部署文档
├── docker-compose.yml              # Docker 编排（含资源限制）
├── .env.example                    # 环境变量模板
└── .gitignore
```

## 技术栈

- **后端**: Go 1.21, Gin 框架
- **数据库**: PostgreSQL 15, Redis 7
- **消息队列**: Kafka (审计日志)
- **容器**: Docker, Docker Compose（含资源限制）
- **加密**: 国密 SM4-GCM, SM3, TOTP (RFC 6238)
- **认证**: JWT (HMAC-SHA256), MFA/TOTP

## 相关文档

- [系统架构](docs/ARCHITECTURE.md)
- [API 规格说明](docs/API_SPEC.md)
- [部署指南](docs/DEPLOYMENT.md)
- [安全合规说明](docs/SECURITY.md)
- [产品需求说明书 V1.2.0](docs/账户管理微服务%20(Account%20Center)%20产品需求说明书%20V1.2.0.md)
- [技术实现方案 V1.0.0](docs/账户管理微服务%20(Account%20Center)%20技术实现方案%20V1.0.0.md)

## 许可证

MIT License
