# RBAC 角色权限系统改造设计

> **文档类型**: 技术设计规格
> **版本**: V1.0.0
> **日期**: 2026-05-22
> **状态**: 待评审

---

## 1. 目标

将现有断裂的角色权限系统打通为完整 RBAC：
- auth-service 登录时查角色写入 JWT
- api-gateway 网关层路由级角色拦截
- 微服务层敏感操作细粒度权限检查
- 前端根据角色/权限动态渲染菜单和按钮

## 2. 关键决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 角色数据源 | auth-service 直接查 PG | 简单高效，无需网络调用，共享同一数据库 |
| JWT 存储格式 | `roles: ["admin","config_editor"]` | 验证时无需查库，直接读 claims |
| 权限检查层 | 网关路由拦截 + 服务层细粒度 | 双层防护 |
| 前端策略 | 登录后调 API 查权限，动态渲染 | 实现按钮级控制 |

## 3. 审查发现的现有问题（必须修复）

### P-01: JWT Claims 结构分裂

三个地方各自手动解析 JWT，没有共用 Claims struct：

| 位置 | 解析方式 | 读 roles |
|------|---------|----------|
| `auth-service/pkg/jwt/jwt.go` Claims struct | `jwt.ParseWithClaims` | ❌ 无 roles 字段 |
| `api-gateway/internal/middleware/jwt.go` | `map[string]interface{}` 手动 base64 | ❌ 不读 roles |
| `account-service/internal/middleware/admin_auth.go` | `map[string]interface{}` 手动 base64 | ⚠️ 读 `claims["role"]` (单数 string) |

**风险**：只改 auth-service Claims struct 不改网关和下游解析，roles 字段在网关层被丢弃。

### P-02: admin_auth.go 读 role（单数 string），设计存 roles（数组）

```go
// 现有代码 account-service/internal/middleware/admin_auth.go:68
role, _ := claims["role"].(string)  // 期望单数 string
if role != "admin" { ... }

// 设计要存的是:
roles: ["admin", "config_editor"]  // 复数 []string
```

**风险**：key 从 `"role"` 变为 `"roles"`，类型从 string 变为 array，`claims["role"]` 返回 nil，所有 admin 操作永远 403。

### P-03: api-gateway 是反向代理，JWT 不透传 claims

api-gateway `JWTAuthMiddleware` 验证后只 `c.Set("user_id", ...)`，然后反向代理。下游服务收到原始 JWT token，需自己解析。

**影响**：RoleGuardMiddleware 只能在代理前拦截路由，下游服务的权限检查仍依赖各自解析 JWT。

### P-04: 前端 store 没有 roles/permissions 位置

`web-ui/src/store/auth.ts` 只存 `access_token, refresh_token, user_id, account_id`。JWT payload 中的 roles 即使加了，前端也无法使用。

### P-05: user_roles.user_id 是 string（存的是 account_id）

```sql
-- user_roles 表: user_id VARCHAR
-- 现有数据: user_id = "admin" (这是 account_id)
-- 但 users 表主键是 id (SERIAL), account_id 是另一个字段
-- 登录响应给前端的是 user_id: 14 (int64)
```

**风险**：auth-service 查角色时必须用 `account_id`（string）匹配 `user_roles.user_id`，不能用 `users.id`（int64）。

### P-06: LoginResponse 缺少 roles 字段

`auth-service/internal/model/login.go` 的 `LoginResponse` 没有 `roles` 字段。即使 JWT 里有 roles，登录响应也不返回给前端。

### P-07: Refresh Token 丢失 roles

`auth-service/pkg/jwt/jwt.go:133` 的 `RefreshTokenPair` 用旧 claims 重建：
```go
return m.GenerateTokenPairWithDevice(claims.UserID, claims.AccountID, ...)
```
如果 `generateToken` 签名加了 roles 参数但这里不传，refresh 后的 token 会丢失 roles。

---

## 4. 改造方案

### 4.1 数据库迁移

新增 migration `db-migrations/005_rbac_business_roles.sql`：

```sql
-- +goose Up

-- 新增业务角色
INSERT INTO roles (name, description) VALUES
  ('admin',   '系统管理员，全部权限'),
  ('operator','运营人员，用户管理+数据查看'),
  ('finance', '财务人员，订单+发票+退款'),
  ('support', '客服人员，用户查看+设备管理'),
  ('user',    '普通用户，默认角色');

-- admin 权限（全部）
INSERT INTO role_permissions (role_id, permission) SELECT id, p FROM roles,
  (SELECT unnest(ARRAY[
    'admin.user.manage', 'admin.user.freeze', 'admin.user.ban',
    'admin.credit.adjust', 'admin.plan.manage', 'admin.coupon.manage',
    'admin.audit.view', 'admin.blacklist.manage',
    'config.read', 'config.edit', 'config.delete',
    'release.create', 'release.approve', 'release.execute',
    'audit.view', 'permission.manage',
    'finance.order.view', 'finance.refund.approve', 'finance.invoice.manage',
    'data.dashboard', 'data.rfm', 'data.funnel',
    'sms.status'
  ]) AS p) perms
WHERE name = 'admin';

-- operator 权限
INSERT INTO role_permissions (role_id, permission) SELECT id, p FROM roles,
  (SELECT unnest(ARRAY[
    'admin.user.manage', 'admin.audit.view', 'admin.blacklist.manage',
    'data.dashboard', 'data.rfm', 'data.funnel', 'sms.status'
  ]) AS p) perms
WHERE name = 'operator';

-- finance 权限
INSERT INTO role_permissions (role_id, permission) SELECT id, p FROM roles,
  (SELECT unnest(ARRAY[
    'finance.order.view', 'finance.refund.approve', 'finance.invoice.manage',
    'data.dashboard'
  ]) AS p) perms
WHERE name = 'finance';

-- support 权限
INSERT INTO role_permissions (role_id, permission) SELECT id, p FROM roles,
  (SELECT unnest(ARRAY[
    'admin.user.manage', 'admin.audit.view', 'data.dashboard', 'sms.status'
  ]) AS p) perms
WHERE name = 'support';

-- user 权限（基本自助）
INSERT INTO role_permissions (role_id, permission) SELECT id, p FROM roles,
  (SELECT unnest(ARRAY[
    'account.self', 'credits.self', 'subscriptions.self', 'devices.self',
    'referral.self', 'data.rfm.self'
  ]) AS p) perms
WHERE name = 'user';

-- 所有现有用户分配 user 角色
INSERT INTO user_roles (user_id, role_id)
  SELECT u.account_id, r.id FROM users u, roles r WHERE r.name = 'user';

-- admin_user 分配 admin 角色
INSERT INTO user_roles (user_id, role_id) VALUES ('admin_user', (SELECT id FROM roles WHERE name = 'admin'));
```

### 4.2 auth-service 改造

#### 4.2.1 JWT Claims 加 roles 字段

文件：`auth-service/pkg/jwt/jwt.go`

```
Claims struct 新增:
  Roles []string `json:"roles,omitempty"`

generateToken 签名变更:
  新增 roles 参数: func generateToken(userID int64, accountID string, tokenID string,
    deviceFingerprint string, roles []string, secret string, expiry time.Duration)

Claims 构造:
  Claims{..., Roles: roles}
```

**json tag**: `"roles,omitempty"` — JWT payload 中存为 `"roles":["admin","config_editor"]`

#### 4.2.2 登录时查角色

文件：`auth-service/internal/service/auth_service.go`

```
Login 流程新增步骤（在生成 token 之前）:
  1. 用 user.AccountID 查 PG:
     SELECT r.name FROM roles r
       JOIN user_roles ur ON ur.role_id = r.id
       WHERE ur.user_id = $1
  2. 将 role names 传入 GenerateTokenPairWithDevice
```

**关键**：用 `user.AccountID`（string，如 "admin_user"）匹配 `user_roles.user_id`，不用 `user.ID`（int64）。

#### 4.2.3 LoginResponse 加 roles

文件：`auth-service/internal/model/login.go`

```
LoginResponse 新增:
  Roles []string `json:"roles"`
```

**json tag**: `"roles"` — 登录响应返回 `{"access_token":..., "roles":["admin"]}`

#### 4.2.4 Refresh Token 携带 roles

文件：`auth-service/pkg/jwt/jwt.go`

```
Claims struct 有 Roles 字段后，RefreshTokenPair 自然从旧 claims 中读取 Roles:

func (m *JWTManager) RefreshTokenPair(refreshToken string) (*TokenResponse, error) {
    claims := &Claims{}
    // ...parse...
    return m.GenerateTokenPairWithDevice(claims.UserID, claims.AccountID,
        claims.DeviceFingerprint, claims.Roles)  // ← 传入 roles
}

GenerateTokenPairWithDevice 签名新增 roles 参数:
  func (m *JWTManager) GenerateTokenPairWithDevice(userID int64, accountID,
    deviceFingerprint string, roles []string) (*TokenResponse, error)
```

#### 4.2.5 auth-service 新增角色查询 Repository

新文件：`auth-service/internal/repository/role_repository.go`

```
type RoleRepository interface {
    GetUserRoles(ctx context.Context, accountID string) ([]string, error)
}

实现:
  SELECT r.name FROM roles r
    JOIN user_roles ur ON ur.role_id = r.id
    WHERE ur.user_id = $1
```

注意：此处 `accountID` 对应 `user_roles.user_id`（存的是 account_id string）。

### 4.3 api-gateway 改造

#### 4.3.1 JWTAuthMiddleware 读 roles

文件：`api-gateway/internal/middleware/jwt.go`

```
现有代码 claims 是 map[string]interface{}
新增 roles 解析:
  roles := []string{}
  if r, ok := claims["roles"].([]interface{}); ok {
    for _, v := range r {
      if s, ok := v.(string); ok { roles = append(roles, s) }
    }
  }
  c.Set("roles", roles)
```

#### 4.3.2 新增 RoleGuardMiddleware

新文件：`api-gateway/internal/middleware/role_guard.go`

```
路由-角色映射配置:
  "/api/v1/admin/":          ["admin", "system_owner"]
  "/api/v1/audit/":          ["admin", "operator"]
  "/api/v1/blacklist/":      ["admin", "operator"]
  "/api/v1/kyb/":            ["admin"]
  "/api/v1/config/":         ["admin", "system_owner", "config_editor", "config_viewer"]
  "/api/v1/risk/":           ["admin", "operator", "support"]
  "/api/v1/orders/":         ["admin", "finance"]
  "/api/v1/invoices/":       ["admin", "finance"]
  "/api/v1/refunds/":        ["admin", "finance"]

中间件逻辑:
  1. 从 c.Get("roles") 获取用户角色列表
  2. 遍历路由-角色映射，找到匹配当前路径的前缀
  3. 检查用户角色是否包含任一允许角色
  4. 不匹配返回 403 {"error": "insufficient permissions"}
```

#### 4.3.3 main.go 注册中间件

文件：`api-gateway/cmd/main.go`

在 `JWTAuthMiddleware` 之后、反向代理之前注册：
```
r.Use(middleware.RoleGuardMiddleware(roleConfig))
```

### 4.4 account-service 改造

#### 4.4.1 AdminAuthMiddleware 改读 roles 数组

文件：`account-service/internal/middleware/admin_auth.go`

```
现有:
  role, _ := claims["role"].(string)
  if role != "admin" { ... }

改为:
  roles := []string{}
  if r, ok := claims["roles"].([]interface{}); ok {
    for _, v := range r {
      if s, ok := v.(string); ok { roles = append(roles, s) }
    }
  }
  hasAdmin := false
  for _, r := range roles {
    if r == "admin" || r == "system_owner" { hasAdmin = true; break }
  }
  if !hasAdmin {
    c.AbortWithStatusJSON(403, gin.H{"error": "admin access required"})
    return
  }
  c.Set("roles", roles)
```

**json tag 一致性**：JWT payload 中字段名是 `"roles"`，Go 的 `json.Unmarshal` 到 `map[string]interface{}` 后 key 为 `"roles"`，此处必须读 `claims["roles"]`。

#### 4.4.2 细粒度权限检查中间件

新文件：`account-service/internal/middleware/permission_check.go`

```
func RequirePermission(db *sql.DB) gin.HandlerFunc {
  return func(c *gin.Context) {
    roles := c.GetStringSlice("roles")
    // 查 role_permissions 表检查具体权限
    // 如: admin.credit.adjust, admin.user.freeze
  }
}
```

### 4.5 前端改造

#### 4.5.1 auth store 存 roles

文件：`web-ui/src/store/auth.ts`

```
新增:
  const roles = ref<string[]>(JSON.parse(localStorage.getItem('roles') || '[]'))

doLogin 中:
  roles.value = d.roles || []
  localStorage.setItem('roles', JSON.stringify(roles.value))

doLogout 中:
  roles.value = []
```

#### 4.5.2 新增 permission store

新文件：`web-ui/src/store/permission.ts`

```
usePermissionStore:
  permissions: string[]  // 如 ["admin.user.manage", "config.read"]
  roles: string[]

  loadPermissions(userId: string):
    1. 调 GET /api/v1/config/users/:userId/roles → 获取角色列表
    2. 对每个角色调 GET /api/v1/config/roles/:id/permissions → 获取权限列表
    3. 合并去重存入 permissions

  hasPermission(perm: string): boolean
  hasRole(role: string): boolean
  hasAnyRole(roles: string[]): boolean
```

#### 4.5.3 登录后加载权限

文件：`web-ui/src/store/auth.ts`

```
doLogin 成功后:
  const permStore = usePermissionStore()
  await permStore.loadPermissions(String(d.user_id))
```

#### 4.5.4 动态菜单和按钮

文件：`web-ui/src/layouts/DefaultLayout.vue`

```
import { usePermissionStore } from '@/store/permission'

菜单项:
  <el-menu-item v-if="hasAnyRole(['admin','operator','finance','support'])"
    index="/admin">
    <el-icon><Setting /></el-icon>管理后台
  </el-menu-item>
```

文件：`web-ui/src/views/Admin.vue` 及各页面

```
按钮级控制:
  <el-button v-if="hasPermission('admin.credit.adjust')" @click="adjustCredits">
    调整积分
  </el-button>
```

---

## 5. json tag 一致性矩阵

### 5.1 JWT Claims 全链路

| 字段 | auth-service 签发 | api-gateway 解析 | account-service 解析 | 前端解码 |
|------|-----------------|----------------|---------------------|---------|
| user_id | `json:"user_id"` | `claims["user_id"]` | `claims["user_id"]` | JWT payload decode |
| account_id | `json:"account_id"` | `claims["account_id"]` | `claims["account_id"]` | JWT payload decode |
| roles (新增) | `json:"roles,omitempty"` | `claims["roles"]` → `[]interface{}` | `claims["roles"]` → `[]interface{}` | JWT payload decode |
| exp | `json:"exp"` | `claims["exp"]` | `claims["exp"]` | — |

### 5.2 LoginResponse 前后端

| 字段 | auth-service Go tag | 前端 auth.ts | 一致性 |
|------|-------------------|-------------|--------|
| access_token | `json:"access_token"` | `d.access_token` | ✅ |
| refresh_token | `json:"refresh_token"` | `d.refresh_token` | ✅ |
| user_id | `json:"user_id"` | `d.user_id` | ✅ |
| account_id | `json:"account_id"` | `d.account_id` | ✅ |
| roles (新增) | `json:"roles"` | `d.roles` | ✅ 需同步添加 |

### 5.3 config-service API 前后端

| API | 后端 Go tag | 前端调用 | 一致性 |
|-----|-----------|---------|--------|
| GET /config/users/:userId/roles | `UserRole{UserID, RoleID, CreatedAt}` | `permStore.loadPermissions` | ✅ |
| GET /config/roles/:id/permissions | `RolePermission{RoleID, Permission}` | 遍历获取 permission 字段 | ✅ |
| GET /config/roles | `Role{Name, Description}` | — | ✅ |

### 5.4 数据库字段映射

| DB 字段 | Go model tag | 使用场景 | 一致性 |
|---------|------------|---------|--------|
| `user_roles.user_id` (VARCHAR) | `json:"user_id"` | auth-service 用 account_id 查询 | ⚠️ 字段名是 user_id 但存的是 account_id |
| `user_roles.role_id` (INT) | `json:"role_id"` | JOIN roles 表 | ✅ |
| `roles.name` (VARCHAR) | `json:"name"` | 写入 JWT roles | ✅ |
| `role_permissions.permission` (VARCHAR) | `json:"permission"` | 前端权限检查 | ✅ |

---

## 6. 改动清单

| 服务 | 文件 | 改动类型 | 改动内容 |
|------|------|---------|---------|
| DB | `db-migrations/005_rbac_business_roles.sql` | 新增 | 5 个业务角色 + 权限 + 默认分配 |
| auth-service | `pkg/jwt/jwt.go` | 修改 | Claims 加 Roles，generateToken/GenerateTokenPairWithDevice/RefreshTokenPair 传递 roles |
| auth-service | `internal/model/login.go` | 修改 | LoginResponse 加 Roles 字段 |
| auth-service | `internal/repository/role_repository.go` | 新增 | GetUserRoles(accountID) 查 PG |
| auth-service | `internal/service/auth_service.go` | 修改 | Login 流程查角色，传给 JWT 签发 |
| auth-service | `cmd/main.go` | 修改 | 注入 RoleRepository |
| api-gateway | `internal/middleware/jwt.go` | 修改 | 解析 claims["roles"] 存入 context |
| api-gateway | `internal/middleware/role_guard.go` | 新增 | 路由级角色拦截中间件 |
| api-gateway | `cmd/main.go` | 修改 | 注册 RoleGuardMiddleware + 角色路由配置 |
| account-service | `internal/middleware/admin_auth.go` | 修改 | 读 roles 数组替代 role string |
| account-service | `internal/middleware/permission_check.go` | 新增 | 细粒度权限检查中间件 |
| config-service | `internal/handler/permission_handler.go` | 修改 | 新增当前用户权限汇总 API |
| config-service | `internal/service/permission_service.go` | 修改 | GetUserPermissions 合并所有角色权限 |
| config-service | `cmd/main.go` | 修改 | 注册新路由 |
| 前端 | `src/store/auth.ts` | 修改 | 存 roles |
| 前端 | `src/store/permission.ts` | 新增 | 权限 store |
| 前端 | `src/api/permission.ts` | 新增 | config-service 权限 API 调用 |
| 前端 | `src/layouts/DefaultLayout.vue` | 修改 | 动态菜单 |
| 前端 | `src/views/Admin.vue` | 修改 | 按钮级权限控制 |
| 前端 | `src/views/Dashboard.vue` | 修改 | 快捷入口按角色过滤 |

---

## 7. 角色权限矩阵

| 权限 | admin | operator | finance | support | user |
|------|-------|----------|---------|---------|------|
| admin.user.manage | ✅ | ✅ | — | ✅ | — |
| admin.user.freeze | ✅ | — | — | — | — |
| admin.user.ban | ✅ | — | — | — | — |
| admin.credit.adjust | ✅ | — | — | — | — |
| admin.plan.manage | ✅ | — | — | — | — |
| admin.coupon.manage | ✅ | — | — | — | — |
| admin.audit.view | ✅ | ✅ | — | ✅ | — |
| admin.blacklist.manage | ✅ | ✅ | — | — | — |
| finance.order.view | ✅ | — | ✅ | — | — |
| finance.refund.approve | ✅ | — | ✅ | — | — |
| finance.invoice.manage | ✅ | — | ✅ | — | — |
| config.read | ✅ | — | — | — | — |
| config.edit | ✅ | — | — | — | — |
| config.delete | ✅ | — | — | — | — |
| data.dashboard | ✅ | ✅ | ✅ | ✅ | — |
| data.rfm | ✅ | ✅ | — | — | — |
| data.funnel | ✅ | ✅ | — | — | — |
| sms.status | ✅ | ✅ | — | ✅ | — |
| account.self | ✅ | ✅ | ✅ | ✅ | ✅ |
| credits.self | ✅ | ✅ | ✅ | ✅ | ✅ |
| subscriptions.self | ✅ | ✅ | ✅ | ✅ | ✅ |
| devices.self | ✅ | ✅ | ✅ | ✅ | ✅ |
| referral.self | ✅ | ✅ | ✅ | ✅ | ✅ |
| data.rfm.self | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 8. 执行顺序

1. **DB migration** — 新增角色和权限数据（无破坏性变更）
2. **auth-service** — Claims + 查角色 + LoginResponse（核心改造）
3. **api-gateway** — JWT 解析 + RoleGuard（依赖 #2）
4. **account-service** — AdminAuthMiddleware + 细粒度中间件（依赖 #2）
5. **config-service** — 用户权限汇总 API（前端依赖）
6. **前端** — permission store + 动态菜单按钮（依赖 #2 + #5）
7. **E2E 测试** — Playwright 验证角色隔离

每一步可独立构建镜像部署，不会破坏现有功能。
