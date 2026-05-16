
# 配置管理系统数据库迁移设计

&gt; Version: v1.0.0
&gt; Date: 2026-05-16
&gt; Status: Design Complete

---

## 1. Overview

为配置管理系统创建完整的数据库迁移脚本，包含表结构和初始化数据。

### Scope
- 创建8个核心配置管理表
- 插入初始化数据（角色、权限、分组、配置项）
- 使用 goose 迁移工具格式

### Non-Goals
- 不包含业务逻辑代码
- 不包含 API 接口

---

## 2. Migration File Details

### File Location
```
db-migrations/004_config_management_schema.sql
```

### Format
- 使用 goose 标准格式: `-- +goose Up` 和 `-- +goose Down`
- 单次迁移创建所有表并插入所有数据

---

## 3. Table Structure

### 3.1 config_groups（配置分组表）
```sql
CREATE TABLE config_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### 3.2 config_items（配置项表）
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

### 3.3 config_versions（配置版本历史表）
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

### 3.4 config_releases（发布申请表）
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

### 3.5 config_release_items（发布单配置项关联表）
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

### 3.6 audit_logs（审计日志表）
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

### 3.7 roles（角色表）
```sql
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### 3.8 role_permissions（角色权限表）
```sql
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, permission)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
```

### 3.9 user_roles（用户角色表）
```sql
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

---

## 4. Initialization Data

### 4.1 Roles
```sql
INSERT INTO roles (id, name, description) VALUES
(1, 'system_owner', '系统所有者，可审批发布'),
(2, 'config_editor', '配置编辑者，可编辑配置'),
(3, 'config_viewer', '配置查看者，只读');
```

### 4.2 Role Permissions
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

### 4.3 Config Groups
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

### 4.4 Config Items (Sample - Full 106 items in migration)
配置项的完整初始化数据将包含106个配置项，每个配置项都有完整的结构。

---

## 5. Rollback Strategy

`-- +goose Down` 部分将按相反顺序删除所有表。

---

## 6. Success Criteria
- Migration executes without errors
- All tables created with proper constraints and indexes
- Initialization data correctly inserted
- Rollback works properly
