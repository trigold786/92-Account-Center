-- +goose Up
-- ============================================================
-- RBAC Business Roles - Database Migration
-- Version: 005
-- Creates business roles for RBAC system
-- ============================================================

-- --- Business Roles (coexist with existing system_owner/config_editor/config_viewer) ---
INSERT INTO roles (id, name, description) VALUES
  (4, 'admin',    '系统管理员，全部权限'),
  (5, 'operator', '运营人员，用户管理+数据查看'),
  (6, 'finance',  '财务人员，订单+发票+退款'),
  (7, 'support',  '客服人员，用户查看+设备管理'),
  (8, 'user',     '普通用户，默认角色')
ON CONFLICT (id) DO NOTHING;

-- --- Business Role Permissions ---
-- admin: 全部权限
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'admin.user.manage', 'admin.user.freeze', 'admin.user.ban',
  'admin.credit.adjust', 'admin.plan.manage', 'admin.coupon.manage',
  'admin.audit.view', 'admin.blacklist.manage',
  'config.read', 'config.edit', 'config.delete',
  'release.create', 'release.approve', 'release.execute',
  'audit.view', 'permission.manage',
  'finance.order.view', 'finance.refund.approve', 'finance.invoice.manage',
  'data.dashboard', 'data.rfm', 'data.funnel', 'sms.status'
]) AS p) perms WHERE name = 'admin'
ON CONFLICT (role_id, permission) DO NOTHING;

-- operator: 运营
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'admin.user.manage', 'admin.audit.view', 'admin.blacklist.manage',
  'data.dashboard', 'data.rfm', 'data.funnel', 'sms.status'
]) AS p) perms WHERE name = 'operator'
ON CONFLICT (role_id, permission) DO NOTHING;

-- finance: 财务
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'finance.order.view', 'finance.refund.approve', 'finance.invoice.manage', 'data.dashboard'
]) AS p) perms WHERE name = 'finance'
ON CONFLICT (role_id, permission) DO NOTHING;

-- support: 客服
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'admin.user.manage', 'admin.audit.view', 'data.dashboard', 'sms.status'
]) AS p) perms WHERE name = 'support'
ON CONFLICT (role_id, permission) DO NOTHING;

-- user: 基本自助权限 + 页面/导航
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'account.self', 'credits.self', 'subscriptions.self', 'devices.self',
  'referral.self', 'data.rfm.self',
  'page.dashboard', 'page.account', 'page.credits', 'page.subscriptions', 'page.referral', 'page.devices',
  'nav.dashboard', 'nav.account', 'nav.credits', 'nav.subscriptions', 'nav.referral', 'nav.devices',
  'account.password.change', 'account.delete.apply', 'account.delete.cancel',
  'referral.copy', 'device.trust', 'device.remove'
]) AS p) perms WHERE name = 'user'
ON CONFLICT (role_id, permission) DO NOTHING;

-- support: 客服权限 + 页面/导航
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'admin.user.manage', 'admin.audit.view', 'data.dashboard', 'sms.status',
  'admin.risk.view', 'admin.sms.view', 'page.dashboard', 'page.admin',
  'nav.dashboard', 'nav.admin', 'admin.overview.view'
]) AS p) perms WHERE name = 'support'
ON CONFLICT (role_id, permission) DO NOTHING;

-- finance: 财务权限 + 页面/导航
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'finance.order.view', 'finance.refund.approve', 'finance.invoice.manage', 'data.dashboard',
  'page.dashboard', 'page.admin', 'nav.dashboard', 'nav.admin', 'admin.overview.view'
]) AS p) perms WHERE name = 'finance'
ON CONFLICT (role_id, permission) DO NOTHING;

-- operator: 运营权限 + 页面/导航/按钮
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'admin.user.manage', 'admin.audit.view', 'admin.blacklist.manage',
  'data.dashboard', 'data.rfm', 'data.funnel', 'sms.status',
  'admin.overview.view', 'admin.risk.view', 'admin.blacklist.view',
  'admin.blacklist.add', 'admin.blacklist.delete', 'admin.audit.view', 'admin.audit.verify',
  'admin.sms.view', 'page.dashboard', 'page.admin',
  'nav.dashboard', 'nav.admin'
]) AS p) perms WHERE name = 'operator'
ON CONFLICT (role_id, permission) DO NOTHING;

-- admin: 扩展所有页面/导航/按钮级权限
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles, (SELECT unnest(ARRAY[
  'page.dashboard', 'page.account', 'page.credits', 'page.subscriptions', 'page.referral', 'page.devices', 'page.admin',
  'nav.dashboard', 'nav.account', 'nav.credits', 'nav.subscriptions', 'nav.referral', 'nav.devices', 'nav.admin',
  'account.password.change', 'account.delete.apply', 'account.delete.cancel',
  'referral.copy', 'device.trust', 'device.remove',
  'admin.overview.view', 'admin.risk.view',
  'admin.blacklist.view', 'admin.blacklist.add', 'admin.blacklist.delete',
  'admin.audit.view', 'admin.audit.verify', 'admin.sms.view',
  'admin.roles.view', 'admin.roles.create', 'admin.roles.edit', 'admin.roles.delete',
  'admin.roles.permission', 'admin.roles.assign'
]) AS p) perms WHERE name = 'admin'
ON CONFLICT (role_id, permission) DO NOTHING;

-- --- Assign roles to existing test users ---
-- All existing users get 'user' role (by account_id)
INSERT INTO user_roles (user_id, role_id)
  SELECT u.account_id, r.id FROM public.users u, roles r WHERE r.name = 'user'
  AND NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.account_id AND ur.role_id = r.id);

-- admin_user (account_id='admin_user') gets 'admin' role
INSERT INTO user_roles (user_id, role_id) VALUES ('admin_user', (SELECT id FROM roles WHERE name = 'admin'))
ON CONFLICT (user_id, role_id) DO NOTHING;

-- normal_user (account_id='normal_user') already has 'user' from the generic insert above

-- Update existing 'admin' account (from 004 migration) to also have 'admin' business role
INSERT INTO user_roles (user_id, role_id) VALUES ('admin', (SELECT id FROM roles WHERE name = 'admin'))
ON CONFLICT (user_id, role_id) DO NOTHING;

-- +goose Down
DELETE FROM user_roles WHERE role_id IN (4,5,6,7,8);
DELETE FROM role_permissions WHERE role_id IN (4,5,6,7,8);
DELETE FROM roles WHERE id IN (4,5,6,7,8);
