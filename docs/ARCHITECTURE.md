# Account Center 系统架构

## 整体架构

```
                         ┌─────────────────────────────────────────┐
                         │              API Gateway                │
                         │            (Port 8080)                  │
                         └──────────────┬──────────────────────────┘
                                        │
        ┌────────────┬────────────┬──────┴───────┬────────────┬────────────┐
        ▼            ▼            ▼             ▼            ▼            ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
   │ Account │  │  Auth   │  │   SMS   │  │  Device │  │   KYB   │  │  Audit  │
   │ :8081   │  │ :8082   │  │  :8083  │  │  :8089  │  │  :8084  │  │  :8085  │
   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
        │            │            │             │            │            │
        └────────────┴────────────┴─────────────┴────────────┴────────────┘
                                        │
                        ┌───────────────┴────────────────┐
                        │         PostgreSQL :5432         │
                        │         (account_center)        │
                        └───────────────────┬─────────────┘
                                            │
                        ┌───────────────────┴───────────────────┐
                        │              Redis :6379             │
                        │    (Session, OTP, Rate Limit)        │
                        └───────────────────────────────────────┘
```

## 服务详解

### 1. API Gateway (8080)

入口服务，负责：
- 请求路由与负载均衡
- JWT 令牌验证
- CORS 跨域处理
- 请求日志
- 熔断器保护

```
请求流程:
Client → API Gateway → Auth Middleware → Target Service
```

### 2. Account Service (8081)

用户账户核心服务：
- 手机号注册 (SendSMSCode → Register)
- 密码修改 (带会话失效)
- 账户注销 (7天冻结期)

```
注册流程:
POST /api/v1/account/register/send-sms-code
POST /api/v1/account/register/phone
```

### 3. Auth Service (8082)

认证服务：
- 统一登录 (手机/邮箱/账号ID)
- JWT 令牌生成
- 刷新令牌
- Magic Link 登录

```
登录流程:
POST /api/v1/auth/login
POST /api/v1/auth/login/send-sms-code
POST /api/v1/auth/login/send-email-otp
```

### 4. SMS Service (8083)

多云短信服务：
- 阿里云 SMS
- 腾讯云 SMS
- 天翼云 SMS
- 熔断器保护 (阈值 5, 重置 30s)
- 频率限制 (60秒、10次/天)

### 5. Device Service (8089)

设备指纹服务：
- 设备指纹注册
- 可信设备标记 (默认3天)
- 设备验证与风险评估
- 用户设备列表

### 6. KYB Service (8084)

企业认证服务：
- 企业信息提交 (SM4 加密存储)
- 小额打款验证 (0.01-0.1元)
- 人脸核身
- 认证状态跟踪

### 7. Audit Service (8085)

审计日志服务：
- 操作审计记录 (SM3 哈希)
- Kafka 消费 (异步收集)
- 日志完整性验证
- 180 天自动清理

### 8. Email Service (8088)

邮件服务：
- OTP 验证码 (6位, 5分钟)
- Magic Link (JWT, 15分钟)
- 频率限制 (3次/小时)
- SendGrid/SES/SMTP

### 9. Risk Service (8086)

风险检测服务：
- 地理位置异常检测 (>1000km/h)
- 设备指纹变化检测
- 登录频率检测 (>10次/小时)
- 风险评分 (0-100)

### 10. Session Service (8087)

会话管理服务：
- Redis 存储
- 20 分钟滑动窗口
- 最大 5 会话限制
- 强制下线支持

## 数据流

### 用户注册流程

```
1. Client → Account Service: POST /register/send-sms-code
2. Account Service → SMS Service: 发送验证码
3. Client → Account Service: POST /register/phone (含验证码)
4. Account Service → DB: 创建用户
5. Account Service → Audit Service: 记录审计日志
6. 返回: user_id, access_token, refresh_token
```

### 用户登录流程

```
1. Client → Auth Service: POST /login (credential + password + fingerprint)
2. Auth Service → DB: 验证凭证
3. Auth Service → Risk Service: 风险评估
4. Auth Service → Device Service: 检查可信设备
5. Auth Service → JWT Manager: 生成令牌
6. Auth Service → Session Service: 创建会话
7. Auth Service → Audit Service: 记录审计日志
8. 返回: access_token, refresh_token, is_trusted_device
```

## 数据库设计

### PostgreSQL 表

```sql
-- 用户表
users (
    user_id UUID PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE,
    email VARCHAR(255) UNIQUE,
    account_id VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255),
    created_at, updated_at,
    is_active, is_locked,
    last_login_at, last_login_ip
)

-- 设备表
user_devices (
    device_id UUID PRIMARY KEY,
    user_id UUID REFERENCES users,
    fingerprint_id VARCHAR(255) UNIQUE,
    device_name, device_type, os_info,
    ip_address, trusted_since,
    is_trusted (默认3天)
)

-- 审计日志表
audit_logs (
    log_id UUID PRIMARY KEY,
    user_id UUID,
    event_time, action_type,
    source_ip, result,
    sm3_hash (完整性校验)
)
```

### Redis 数据结构

```
# Session (20分钟 TTL)
session:{session_id} → Session JSON

# OTP (5分钟 TTL)
otp:{email} → 6位验证码

# Rate Limit (1小时 TTL)
email_rate:{email} → 请求计数

# 用户会话集合
user_sessions:{user_id} → SET of session_id
```

## 安全设计

### 加密策略

| 数据类型 | 加密方式 | 说明 |
|----------|----------|------|
| 密码 | bcrypt + salt | 可调整 cost |
| 身份证号 | SM4-GCM | 企业认证存储 |
| 银行账号 | SM4-GCM | 企业认证存储 |
| JWT 密钥 | 环境变量 | 生产必须更换 |
| 审计日志 | SM3 哈希 | 完整性验证 |

### 认证流程

```
1. 登录获取 access_token (15分钟) + refresh_token (7天)
2. 访问 API 携带 Authorization: Bearer {access_token}
3. access_token 过期后使用 refresh_token 刷新
4. 密码修改后所有会话失效
```

### 等保三级合规

- [x] 身份鉴别
- [x] 访问控制
- [x] 安全审计
- [x] 数据加密
- [x] 通信保密

## 扩展性

### 水平扩展

每个服务可独立扩展：
```bash
docker-compose up -d --scale account-service=3
```

### 服务发现

通过 Docker Compose 网络，服务间通过服务名通信：
- `account-service:8081`
- `postgres:5432`
- `redis:6379`

### 熔断器配置

```yaml
sms-service:
  circuit_breaker:
    threshold: 5          # 失败5次后打开
    reset_timeout: 30s    # 30秒后半开
    half_open_success: 3  # 半开需3次成功
```