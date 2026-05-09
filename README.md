# Account Center Microservice

企业级账户中心微服务系统，为 Neuro 系列产品提供统一的用户认证、账户管理、短信服务、设备指纹、企业认证等功能。系统符合**等保三级 (MLPS 3.0)** 安全标准。

## 功能特性

### 核心服务

| 服务 | 端口 | 功能说明 |
|------|------|----------|
| [API Gateway](api-gateway/) | 8080 | 请求路由、认证中间件、限流熔断 |
| [Account Service](account-service/) | 8081 | 用户注册、密码修改、账户注销 |
| [Auth Service](auth-service/) | 8082 | 统一登录、JWT 令牌、凭证识别 |
| [SMS Service](sms-email-service/) | 8083 | 多云短信（阿里云/腾讯云/天翼云）带熔断器 |
| [Device Service](device-fingerprint-service/) | 8089 | 设备指纹、可信设备管理 |
| [KYB Service](kyb-service/) | 8084 | 企业认证、小额打款验证、人脸核身 |
| [Audit Service](audit-log-service/) | 8085 | 审计日志、SM3 哈希、180天保留 |
| [Email Service](email-service/) | 8088 | OTP 验证码、Magic Link |
| [Risk Service](risk-detection-service/) | 8086 | 风险检测、地理位置异常、登录频率 |
| [Session Service](session-service/) | 8087 | 会话管理、20分钟超时、5会话限制 |

### 安全特性

- **SM4 加密**：敏感数据（身份证号、银行账号）全程加密存储
- **SM3 哈希**：审计日志完整性验证
- **JWT 认证**：访问令牌 15 分钟、刷新令牌 7 天
- **设备指纹**：可信设备标记、异常登录检测
- **会话管理**：Redis 存储、滑动窗口超时、并发限制

### 合规标准

- 等保三级 (MLPS 3.0)
- 数据加密存储
- 180 天审计日志保留
- 敏感操作审计追踪

## 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.21+ (本地开发)
- PostgreSQL 15+
- Redis 7+
- Kafka (可选)

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f api-gateway
```

### 访问端点

- API Gateway: http://localhost:8080
- Health Check: http://localhost:8080/health

## 项目结构

```
account-center/
├── api-gateway/          # API 网关
├── account-service/      # 账户管理
├── auth-service/         # 认证服务
├── sms-email-service/    # 短信邮件
├── device-fingerprint-service/  # 设备指纹
├── kyb-service/          # 企业认证
├── audit-log-service/    # 审计日志
├── email-service/        # 邮件服务
├── risk-detection-service/     # 风险检测
├── session-service/      # 会话管理
├── pkg/                  # 共享包
├── migrations/           # 数据库迁移
├── docs/                 # 文档
└── docker-compose.yml    # Docker 编排
```

## 技术栈

- **后端**: Go 1.21, Gin 框架
- **数据库**: PostgreSQL 15, Redis 7
- **消息队列**: Kafka
- **容器**: Docker, Docker Compose
- **加密**: SM4 (GCM), SM3, bcrypt

## 相关文档

- [系统架构](docs/ARCHITECTURE.md)
- [API 规格说明](docs/API_SPEC.md)
- [部署指南](docs/DEPLOYMENT.md)
- [安全合规说明](docs/SECURITY.md)

## 许可证

MIT License