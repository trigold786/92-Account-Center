# Account Center Microservice

企业级账户中心微服务系统，为 Neuro 系列产品提供统一的用户认证、账户管理、短信服务、设备指纹、企业认证等功能。系统符合**等保三级 (MLPS 3.0)** 安全标准及**Neuro 开发 SOP（L1-L6）**规范。

## 功能特性

### 核心服务

| 服务 | 端口 (w004) | 容器名 | 功能说明 |
|------|-------------|--------|----------|
| [API Gateway](api-gateway/) | 30300 | api-gateway | 请求路由、JWT 认证、限流熔断、CORS |
| [Account Service](account-service/) | 30301 | account-service | 用户注册、密码修改、账户注销 |
| [Auth Service](auth-service/) | 30302 | auth-service | 统一登录、JWT 令牌、MFA/TOTP |
| [SMS/Email Service](sms-email-service/) | 30303 | sms-email-service | 多云短信（阿里云/腾讯/天翼）+ SMTP 邮件 |
| [KYB Service](kyb-service/) | 30304 | kyb-service | 企业认证、小额打款验证、法人核身 |
| [Audit Service](audit-log-service/) | 30305 | audit-log-service | 审计日志、SM3 完整性校验、180天保留 |
| [Risk Service](risk-detection-service/) | 30306 | risk-detection-service | 地理位置异常、设备指纹变化、登录频率检测 |
| [Session Service](session-service/) | 30307 | session-service | 会话 CRUD、20 分钟超时、5 并发限制 |
| [Email Service](email-service/) | 30308 | email-service | OTP 验证码、Magic Link JWT |
| [Device Service](device-fingerprint-service/) | 30309 | device-fingerprint-service | 设备指纹注册、可信设备管理、风险评估 |

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
├── sms-email-service/              # 短信邮件 (30303)
├── device-fingerprint-service/     # 设备指纹 (30309)
├── kyb-service/                    # 企业认证 (30304)
├── audit-log-service/              # 审计日志 (30305)
├── email-service/                  # 邮件服务 (30308)
├── risk-detection-service/         # 风险检测 (30306)
├── session-service/                # 会话管理 (30307)
├── pkg/                            # 共享包
├── migrations/                     # 数据库迁移脚本
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
