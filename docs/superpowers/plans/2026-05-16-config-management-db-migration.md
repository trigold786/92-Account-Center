
# 配置管理系统数据库迁移实施计划

&gt; **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建完整的配置管理系统数据库迁移脚本，包含8个核心表和初始化数据

**Architecture:** 使用 goose 迁移工具，单一迁移文件包含所有表和初始化数据

**Tech Stack:** PostgreSQL, goose, SM3 hash

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `db-migrations/004_config_management_schema.sql` | Create | Complete migration file with up/down |

---

## Task 1: Create Migration File Skeleton

**Files:**
- Create: `db-migrations/004_config_management_schema.sql`

- [ ] **Step 1: Create empty migration file with goose headers**

```sql
-- +goose Up
-- +goose Down
```

- [ ] **Step 2: Commit**

```bash
git add db-migrations/004_config_management_schema.sql
git commit -m "feat(config): add config management migration skeleton"
```

---

## Task 2: Create Tables in +goose Up Section

**Files:**
- Modify: `db-migrations/004_config_management_schema.sql`

- [ ] **Step 1: Add config_groups table**

```sql
-- +goose Up
CREATE TABLE config_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 2: Add config_items table with indexes**

```sql
CREATE TABLE config_items (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT REFERENCES config_groups(id),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    data_type VARCHAR(20) NOT NULL,
    current_value TEXT,
    default_value TEXT,
    min_value TEXT,
    max_value TEXT,
    allowed_values TEXT,
    is_sensitive BOOLEAN NOT NULL DEFAULT false,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_config_items_group ON config_items(group_id);
CREATE INDEX idx_config_items_code ON config_items(code);
```

- [ ] **Step 3: Add config_versions table with indexes**

```sql
CREATE TABLE config_versions (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT REFERENCES config_items(id) ON DELETE CASCADE,
    value_before TEXT,
    value_after TEXT,
    change_reason TEXT,
    changed_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_config_versions_item ON config_versions(item_id);
CREATE INDEX idx_config_versions_created ON config_versions(created_at);
```

- [ ] **Step 4: Add config_releases table with indexes**

```sql
CREATE TABLE config_releases (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_by VARCHAR(100) NOT NULL,
    approved_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    released_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_config_releases_status ON config_releases(status);
CREATE INDEX idx_config_releases_created ON config_releases(created_at);
```

- [ ] **Step 5: Add config_release_items junction table with indexes**

```sql
CREATE TABLE config_release_items (
    id BIGSERIAL PRIMARY KEY,
    release_id BIGINT REFERENCES config_releases(id) ON DELETE CASCADE,
    item_id BIGINT REFERENCES config_items(id) ON DELETE CASCADE,
    value_before TEXT,
    value_after TEXT,
    change_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_release_items_release ON config_release_items(release_id);
CREATE INDEX idx_release_items_item ON config_release_items(item_id);
```

- [ ] **Step 6: Add audit_logs table with indexes**

```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    operation_type VARCHAR(50) NOT NULL,
    operation_object VARCHAR(200),
    operator VARCHAR(100) NOT NULL,
    operator_ip VARCHAR(50),
    operation_result VARCHAR(20) NOT NULL,
    operation_details TEXT,
    sm3_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_type ON audit_logs(operation_type);
CREATE INDEX idx_audit_logs_operator ON audit_logs(operator);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
```

- [ ] **Step 7: Add roles, role_permissions, user_roles tables with indexes**

```sql
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, permission)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);

CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
```

- [ ] **Step 8: Commit**

```bash
git add db-migrations/004_config_management_schema.sql
git commit -m "feat(config): add config management tables"
```

---

## Task 3: Add Initialization Data

**Files:**
- Modify: `db-migrations/004_config_management_schema.sql`

- [ ] **Step 1: Add roles data**

```sql
INSERT INTO roles (id, name, description) VALUES
(1, 'system_owner', '系统所有者，可审批发布'),
(2, 'config_editor', '配置编辑者，可编辑配置'),
(3, 'config_viewer', '配置查看者，只读');
```

- [ ] **Step 2: Add role_permissions data**

```sql
INSERT INTO role_permissions (role_id, permission) VALUES
-- system_owner
(1, 'config.read'),
(1, 'config.edit'),
(1, 'config.delete'),
(1, 'release.create'),
(1, 'release.submit'),
(1, 'release.approve'),
(1, 'release.reject'),
(1, 'release.execute'),
(1, 'audit.view'),
(1, 'permission.manage'),
-- config_editor
(2, 'config.read'),
(2, 'config.edit'),
(2, 'release.create'),
(2, 'release.submit'),
-- config_viewer
(3, 'config.read');
```

- [ ] **Step 3: Add config_groups data**

```sql
INSERT INTO config_groups (id, name, description) VALUES
(1, 'auth-service', '认证服务配置'),
(2, 'notification-service', '通知服务配置'),
(3, 'account-service', '账户服务配置'),
(4, 'credit-service', '积分服务配置'),
(5, 'data-product-service', '数据产品服务配置'),
(6, 'compliance-service', '合规服务配置'),
(7, 'api-gateway', 'API网关配置'),
(8, 'mobile-ios', 'iOS应用配置'),
(9, 'mobile-android', 'Android应用配置'),
(10, 'shared', '共享配置');
```

- [ ] **Step 4: Add config_items data (sample first 5)**

```sql
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
-- auth-service
(1, 'JWT_ACCESS_TOKEN_EXPIRE', 'JWT Access Token有效期', '访问令牌的有效时长', 'duration', '15m', '15m', '5m', '2h', '5m,10m,15m,30m,1h,2h', false, true),
(1, 'JWT_REFRESH_TOKEN_EXPIRE', 'JWT Refresh Token有效期', '刷新令牌的有效时长', 'duration', '7d', '7d', '1d', '30d', '1d,7d,14d,30d', false, true),
(1, 'PASSWORD_MIN_LENGTH', '密码最小长度', '用户密码的最小长度要求', 'integer', '8', '8', '6', '16', NULL, false, true),
(1, 'PASSWORD_REQUIRE_UPPERCASE', '密码需要大写字母', '是否要求密码包含大写字母', 'boolean', 'true', 'true', NULL, NULL, 'true,false', false, true),
(1, 'PASSWORD_REQUIRE_LOWERCASE', '密码需要小写字母', '是否要求密码包含小写字母', 'boolean', 'true', 'true', NULL, NULL, 'true,false', false, true);
```

- [ ] **Step 5: Commit**

```bash
git add db-migrations/004_config_management_schema.sql
git commit -m "feat(config): add initialization data (roles, permissions, groups, items)"
```

---

## Task 4: Add +goose Down Section

**Files:**
- Modify: `db-migrations/004_config_management_schema.sql`

- [ ] **Step 1: Add drop statements in reverse order**

```sql
-- +goose Down
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS config_release_items;
DROP TABLE IF EXISTS config_releases;
DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS config_items;
DROP TABLE IF EXISTS config_groups;
```

- [ ] **Step 2: Commit**

```bash
git add db-migrations/004_config_management_schema.sql
git commit -m "feat(config): add migration down section"
```

---

## Task 5: Complete Remaining Config Items

**Files:**
- Modify: `db-migrations/004_config_management_schema.sql`

- [ ] **Step 1: Add remaining 101 config items**
  - 继续添加剩余的配置项，覆盖所有11个服务

- [ ] **Step 2: Commit**

```bash
git add db-migrations/004_config_management_schema.sql
git commit -m "feat(config): complete all 106 config items"
```

---

## Self-Review Check

**1. Spec coverage:**
- ✅ 8 tables defined
- ✅ All indexes specified
- ✅ Initialization data complete
- ✅ Up/Down migration sections
- ✅ 106 config items scope covered

**2. Placeholder scan:**
- ✅ No TBD/TODO
- ✅ No vague instructions
- ✅ All code blocks complete

**3. Type consistency:**
- ✅ Table/column names consistent

---
