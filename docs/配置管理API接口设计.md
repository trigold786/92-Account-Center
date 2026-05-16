
# 配置管理API接口设计

&gt; **版本**: v1.0.0  
&gt; **日期**: 2026-05-16  
&gt; **阶段**: 阶段2完成

---

## 目录

1. [接口规范](#1-接口规范)
2. [配置分类管理API](#2-配置分类管理api)
3. [配置项管理API](#3-配置项管理api)
4. [版本历史API](#4-版本历史api)
5. [发布审批API](#5-发布审批api)
6. [审计日志API](#6-审计日志api)
7. [权限管理API](#7-权限管理api)

---

## 1. 接口规范

### 1.1 通用信息

| 项目 | 说明 |
|------|------|
| Base URL | `/api/v1/config` |
| 认证方式 | Bearer Token (JWT) |
| 内容类型 | application/json |
| 响应格式 | 统一包装格式 |

### 1.2 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2026-05-16T14:23:00Z"
}
```

### 1.3 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 业务冲突 |
| 500 | 服务器错误 |

---

## 2. 配置分类管理API

### 2.1 获取分类树

```
GET /groups/tree
```

**响应**：
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "code": "auth_service",
      "name": "认证服务",
      "children": [
        {
          "id": 11,
          "code": "jwt_config",
          "name": "JWT配置"
        }
      ]
    }
  ]
}
```

### 2.2 创建分类

```
POST /groups
```

**请求体**：
```json
{
  "code": "new_group",
  "name": "新分类",
  "description": "分类描述",
  "parent_id": null
}
```

### 2.3 更新分类

```
PUT /groups/{id}
```

### 2.4 删除分类

```
DELETE /groups/{id}
```

---

## 3. 配置项管理API

### 3.1 获取配置项列表

```
GET /items
```

**查询参数**：
| 参数 | 类型 | 说明 |
|------|------|------|
| group_id | number | 分组ID |
| keyword | string | 搜索关键词 |
| is_enabled | boolean | 是否启用 |
| page | number | 页码，默认1 |
| size | number | 每页数量，默认20 |

**响应**：
```json
{
  "code": 200,
  "data": {
    "total": 106,
    "items": [
      {
        "id": 1,
        "code": "JWT_ACCESS_TOKEN_EXPIRE",
        "name": "Access Token有效期",
        "description": "Access Token的有效期配置",
        "data_type": "DURATION",
        "current_value": "15m",
        "default_value": "15m",
        "is_sensitive": false,
        "is_enabled": true
      }
    ]
  }
}
```

### 3.2 获取配置项详情

```
GET /items/{id}
```

**响应**：
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "group_id": 1,
    "code": "JWT_ACCESS_TOKEN_EXPIRE",
    "name": "Access Token有效期",
    "description": "Access Token的有效期配置",
    "data_type": "DURATION",
    "default_value": "15m",
    "current_value": "15m",
    "min_value": 5,
    "max_value": 120,
    "options": null,
    "is_sensitive": false,
    "is_enabled": true,
    "updated_at": "2026-05-16T14:23:00Z"
  }
}
```

### 3.3 创建配置项

```
POST /items
```

**请求体**：
```json
{
  "group_id": 1,
  "code": "NEW_CONFIG",
  "name": "新配置项",
  "description": "配置项说明",
  "data_type": "INTEGER",
  "default_value": "10",
  "min_value": 1,
  "max_value": 100,
  "is_sensitive": false
}
```

### 3.4 更新配置项

```
PUT /items/{id}
```

**请求体**：
```json
{
  "name": "更新后的名称",
  "description": "更新后的说明",
  "current_value": "20",
  "change_reason": "业务需要调整参数值"
}
```

### 3.5 删除配置项

```
DELETE /items/{id}
```

### 3.6 重置为默认值

```
POST /items/{id}/reset-default
```

---

## 4. 版本历史API

### 4.1 获取配置项版本历史

```
GET /items/{id}/versions
```

**查询参数**：
| 参数 | 类型 | 说明 |
|------|------|------|
| page | number | 页码 |
| size | number | 每页数量 |

**响应**：
```json
{
  "code": 200,
  "data": {
    "total": 10,
    "items": [
      {
        "id": 1001,
        "version_number": 10,
        "value_before": "15m",
        "value_after": "20m",
        "change_type": "UPDATE",
        "change_reason": "业务调整",
        "created_by": 100,
        "created_at": "2026-05-16T14:23:00Z"
      }
    ]
  }
}
```

### 4.2 回滚到指定版本

```
POST /items/{id}/rollback/{version_id}
```

---

## 5. 发布审批API

### 5.1 创建发布单

```
POST /releases
```

**请求体**：
```json
{
  "title": "优化配置发布",
  "description": "调整JWT和返佣相关配置",
  "config_changes": [
    {"id": 1, "new_value": "20m", "change_reason": "安全性优化"},
    {"id": 2, "new_value": "12%", "change_reason": "运营调整"}
  ]
}
```

### 5.2 获取发布单列表

```
GET /releases
```

**查询参数**：
| 参数 | 类型 | 说明 |
|------|------|------|
| status | string | 状态筛选 |
| page | number | 页码 |
| size | number | 每页数量 |

### 5.3 获取发布单详情

```
GET /releases/{id}
```

### 5.4 提交审批

```
POST /releases/{id}/submit
```

### 5.5 审批发布单

```
POST /releases/{id}/approve
```

**请求体**：
```json
{
  "approved": true,
  "comment": "同意发布"
}
```

### 5.6 拒绝发布单

```
POST /releases/{id}/reject
```

**请求体**：
```json
{
  "comment": "需要调整参数值"
}
```

### 5.7 发布（仅已审批状态）

```
POST /releases/{id}/publish
```

---

## 6. 审计日志API

### 6.1 查询审计日志

```
GET /audit/logs
```

**查询参数**：
| 参数 | 类型 | 说明 |
|------|------|------|
| operator_id | number | 操作人ID |
| operation_type | string | 操作类型 |
| start_time | string | 开始时间 |
| end_time | string | 结束时间 |
| keyword | string | 搜索关键词 |
| page | number | 页码 |
| size | number | 每页数量 |

**响应**：
```json
{
  "code": 200,
  "data": {
    "total": 1000,
    "items": [
      {
        "id": 1001,
        "audit_code": "AUD20260516001",
        "operator_name": "张三",
        "operator_ip": "192.168.1.100",
        "operation_type": "UPDATE_CONFIG",
        "target_object": "JWT_ACCESS_TOKEN_EXPIRE",
        "value_before": "15m",
        "value_after": "20m",
        "result": "SUCCESS",
        "sm3_hash": "abcdef123456...",
        "created_at": "2026-05-16T14:23:00Z"
      }
    ]
  }
}
```

### 6.2 导出审计日志

```
GET /audit/logs/export
```

**查询参数**：同查询接口，额外参数：
| 参数 | 类型 | 说明 |
|------|------|------|
| format | string | CSV/PDF，默认CSV |

---

## 7. 权限管理API

### 7.1 获取角色列表

```
GET /roles
```

### 7.2 创建角色

```
POST /roles
```

### 7.3 获取权限列表

```
GET /permissions
```

### 7.4 分配角色权限

```
PUT /roles/{id}/permissions
```

**请求体**：
```json
{
  "permission_codes": ["config.read", "config.edit", "release.approve"]
}
```

### 7.5 分配用户角色

```
POST /user-roles
```

**请求体**：
```json
{
  "user_id": 100,
  "role_ids": [2, 3]
}
```

---

## 8. 权限编码说明

| 权限编码 | 说明 |
|----------|------|
| config.read | 读取配置 |
| config.create | 创建配置 |
| config.edit | 编辑配置 |
| config.delete | 删除配置 |
| release.create | 创建发布单 |
| release.approve | 审批发布单 |
| release.publish | 发布配置 |
| audit.view | 查看审计日志 |
| permission.manage | 权限管理 |

---

## 9. 配置服务集成API

### 9.1 业务服务读取配置（内部API）

```
GET /service-config/{service_name}
```

**响应**：
```json
{
  "code": 200,
  "data": {
    "JWT_ACCESS_TOKEN_EXPIRE": "15m",
    "PASSWORD_MIN_LENGTH": "8"
  }
}
```

### 9.2 配置变更通知（Redis Pub/Sub）

频道：`config:updates`

消息格式：
```json
{
  "service_name": "auth_service",
  "changes": {
    "JWT_ACCESS_TOKEN_EXPIRE": "20m"
  },
  "updated_at": "2026-05-16T14:23:00Z"
}
```

---

## 总结

阶段2 API接口设计已完成，共7个模块的API接口，支持完整的配置管理、版本历史、发布审批、审计日志功能。
