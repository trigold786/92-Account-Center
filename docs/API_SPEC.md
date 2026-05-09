# API 规格说明

## 基础信息

- **基础 URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **内容类型**: `application/json`

## 通用响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 错误码

| code | 说明 |
|------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 / Token 过期 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器错误 |

---

## Account Service (端口 8081)

### 1. 发送注册验证码

```
POST /account/register/send-sms-code
```

**请求:**
```json
{
  "phone_number": "13800138000"
}
```

**响应:**
```json
{
  "code": 200,
  "message": "验证码发送成功"
}
```

---

### 2. 手机号注册

```
POST /account/register/phone
```

**请求:**
```json
{
  "phone_number": "13800138000",
  "sms_code": "123456",
  "account_id": "zhangsan_2024",
  "password": "Zhang@2024",
  "agree_terms": true
}
```

**响应:**
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 3. 发送密码修改验证码

```
POST /account/password/send-verification-code
```

**请求:**
```json
{
  "contact_type": "phone",
  "contact_value": "13800138000"
}
```

---

### 4. 修改密码

```
POST /account/password/change
```

**请求:**
```json
{
  "new_password": "New@2024",
  "confirm_password": "New@2024",
  "verification_code": "123456",
  "verification_type": "sms_code"
}
```

**说明:**
- `verification_type`: `sms_code` | `email_otp` | `password`

---

### 5. 发送注销验证码

```
POST /account/delete/send-verification-code
```

---

### 6. 注销账户

```
POST /account/delete
```

**请求:**
```json
{
  "verification_code": "123456",
  "agree_consequences": true
}
```

**响应:**
```json
{
  "code": 200,
  "message": "账户已进入注销流程",
  "data": {
    "freeze_period_days": 7
  }
}
```

**说明:** 账户将进入 7 天冻结期，期间可取消注销。

---

### 7. 取消账户注销

```
POST /account/delete/cancel
```

**Headers:** `Authorization: Bearer {access_token}`

---

## Auth Service (端口 8082)

### 1. 登录

```
POST /auth/login
```

**请求:**
```json
{
  "credential": "13800138000",
  "password": "Zhang@2024",
  "device_fingerprint": "fp_abc123xyz"
}
```

**说明:**
- `credential` 可为手机号、邮箱或账号 ID

**响应:**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "is_trusted_device": true
  }
}
```

---

### 2. 发送登录短信验证码

```
POST /auth/login/send-sms-code
```

**请求:**
```json
{
  "phone_number": "13800138000"
}
```

---

### 3. 发送邮箱 OTP

```
POST /auth/login/send-email-otp
```

**请求:**
```json
{
  "email": "user@example.com"
}
```

---

## SMS Service (端口 8083)

### 1. 发送短信

```
POST /sms/send
```

**请求:**
```json
{
  "phone_number": "13800138000",
  "template_code": "SMS_123456789",
  "params": {
    "code": "123456"
  }
}
```

---

### 2. 查看提供商状态

```
GET /sms/providers/status
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "providers": [
      {"name": "aliyun", "status": "closed"},
      {"name": "tencent", "status": "closed"},
      {"name": "chinatelecom", "status": "closed"}
    ]
  }
}
```

---

## Device Service (端口 8089)

### 1. 注册设备

```
POST /device/register
```

**请求:**
```json
{
  "fingerprint_id": "fp_abc123xyz",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.100",
  "country": "China",
  "city": "Beijing",
  "latitude": 39.9042,
  "longitude": 116.4074
}
```

---

### 2. 验证设备

```
POST /device/verify
```

**请求:**
```json
{
  "fingerprint_id": "fp_abc123xyz",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "is_trusted": false,
    "risk_score": 45,
    "risk_factors": ["location_change", "new_device"]
  }
}
```

---

### 3. 标记可信设备

```
POST /device/trust
```

**请求:**
```json
{
  "fingerprint_id": "fp_abc123xyz",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "trust_days": 7
}
```

---

### 4. 获取用户设备列表

```
GET /device/user/{user_id}
```

---

### 5. 删除设备

```
DELETE /device/{device_id}
```

---

## KYB Service (端口 8084)

### 1. 提交企业信息

```
POST /kyb/enterprise/submit
```

**请求:**
```json
{
  "company_name": "北京科技有限公司",
  "unified_social_credit_code": "91110000XXXXXXXX",
  "legal_person_name": "张三",
  "legal_person_id_number": "110101199001011234",
  "bank_name": "中国工商银行",
  "bank_account_number": "6222021234567890123"
}
```

**说明:** `legal_person_id_number` 和 `bank_account_number` 将被 SM4 加密存储。

---

### 2. 发起小额打款验证

```
POST /kyb/enterprise/micro-payment/init
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "enterprise_id": "xxx",
    "amount": 0.08,
    "bank_name": "中国工商银行",
    "bank_account_mask": "****0123"
  }
}
```

**说明:** 请向显示的银行账户转账指定金额以完成验证。

---

### 3. 验证小额打款

```
POST /kyb/enterprise/micro-payment/verify
```

**请求:**
```json
{
  "enterprise_id": "xxx",
  "amount": 0.08
}
```

---

### 4. 提交人脸核身

```
POST /kyb/enterprise/face/verify
```

**请求:**
```json
{
  "enterprise_id": "xxx",
  "token": "face_provider_token_xxx"
}
```

---

### 5. 获取企业认证状态

```
GET /kyb/enterprise/status/{enterprise_id}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "enterprise_id": "xxx",
    "company_name": "北京科技有限公司",
    "verification_status": "verified",
    "micro_payment_status": "verified",
    "face_verification_status": "verified",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T14:20:00Z"
  }
}
```

---

## Audit Service (端口 8085)

### 1. 记录审计日志

```
POST /audit/logs
```

**请求:**
```json
{
  "user_id": 123,
  "action_type": "login",
  "target_resource": "/api/v1/auth/login",
  "source_ip": "192.168.1.100",
  "result": "success",
  "details": {"browser": "Chrome 120"}
}
```

---

### 2. 批量记录审计日志

```
POST /audit/logs/batch
```

**请求:**
```json
{
  "logs": [
    {"user_id": 123, "action_type": "login", ...},
    {"user_id": 123, "action_type": "view_profile", ...}
  ]
}
```

---

### 3. 获取用户审计日志

```
GET /audit/logs/user/{user_id}?limit=100&offset=0
```

---

### 4. 按时间范围查询

```
GET /audit/logs?start=2024-01-01T00:00:00Z&end=2024-01-31T23:59:59Z&limit=50
```

---

### 5. 验证日志完整性

```
GET /audit/logs/{log_id}/verify
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "log_id": "xxx",
    "is_valid": true,
    "stored_hash": "abc123...",
    "computed_hash": "abc123..."
  }
}
```

---

### 6. 清理旧日志 (管理员)

```
POST /audit/logs/cleanup
```

**请求:**
```json
{
  "retention_days": 180
}
```

---

## Email Service (端口 8088)

### 1. 发送 OTP

```
POST /email/otp/send
```

**请求:**
```json
{
  "email": "user@example.com"
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "expires_in": 300
  }
}
```

---

### 2. 验证 OTP

```
POST /email/otp/verify
```

**请求:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

---

### 3. 发送 Magic Link

```
POST /email/magic-link/send
```

**请求:**
```json
{
  "email": "user@example.com",
  "target_url": "https://app.example.com/auth/callback"
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "magic_link": "https://app.example.com/auth/callback?token=eyJhbG...",
    "expires_in": 900
  }
}
```

---

### 4. 验证 Magic Link

```
GET /email/magic-link/verify?token=eyJhbG...
```

---

## Risk Service (端口 8086)

### 1. 风险评估

```
POST /risk/assess
```

**请求:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "ip_address": "192.168.1.100",
  "device_fingerprint": "fp_abc123xyz",
  "timestamp": "2024-01-15T10:30:00Z",
  "location": {
    "latitude": 39.9042,
    "longitude": 116.4074
  }
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "risk_score": 25,
    "risk_level": "low",
    "risk_factors": [],
    "action": "allow"
  }
}
```

**风险等级:**
- `low`: 0-30
- `medium`: 31-60
- `high`: 61-80
- `critical`: 81-100

**操作建议:**
- `allow`: 允许
- `verify`: 需要验证
- `deny`: 拒绝

---

### 2. 获取用户风险历史

```
GET /risk/history/{user_id}?start=2024-01-01&end=2024-01-31
```

---

### 3. 获取风险事件详情

```
GET /risk/event/{event_id}
```

---

## Session Service (端口 8087)

### 1. 创建会话

```
POST /session/create
```

**请求:**
```json
{
  "user_id": 123,
  "device_fingerprint": "fp_abc123xyz",
  "ip_address": "192.168.1.100"
}
```

---

### 2. 验证会话

```
POST /session/validate
```

**请求:**
```json
{
  "session_id": "xxx"
}
```

---

### 3. 获取用户会话列表

```
GET /session/user/{user_id}
```

---

### 4. 使单个会话失效

```
POST /session/invalidate
```

**请求:**
```json
{
  "session_id": "xxx"
}
```

---

### 5. 使所有会话失效

```
POST /session/invalidate-all
```

**请求:**
```json
{
  "user_id": 123
}
```

---

### 6. 刷新会话

```
POST /session/refresh
```

**请求:**
```json
{
  "session_id": "xxx"
}
```

---

## 健康检查

```
GET /health
```

**响应:**
```json
{
  "status": "ok"
}
```