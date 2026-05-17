-- +goose Up
-- ============================================================
-- Config Management System - Database Migration
-- Version: 004
-- Creates configuration management tables and seed data
-- ============================================================

-- -------------------- Config Groups --------------------
CREATE TABLE config_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- -------------------- Config Items --------------------
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

-- -------------------- Config Versions --------------------
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

-- -------------------- Config Releases --------------------
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

-- -------------------- Config Release Items --------------------
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

-- -------------------- Audit Logs --------------------
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

-- -------------------- Roles --------------------
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- -------------------- Role Permissions --------------------
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, permission)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);

-- -------------------- User Roles --------------------
CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- ============================================================
-- Seed Data
-- ============================================================

-- --- Roles ---
INSERT INTO roles (id, name, description) VALUES
(1, 'system_owner', '系统所有者，可审批发布'),
(2, 'config_editor', '配置编辑者，可编辑配置'),
(3, 'config_viewer', '配置查看者，只读');

-- --- Role Permissions ---
INSERT INTO role_permissions (role_id, permission) VALUES
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
(2, 'config.read'),
(2, 'config.edit'),
(2, 'release.create'),
(2, 'release.submit'),
(3, 'config.read');

-- --- Config Groups ---
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

-- --- Config Items ---

-- auth-service (10 items)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(1, 'JWT_ACCESS_TOKEN_EXPIRE', 'JWT Access Token有效期', '访问令牌的有效时长', 'DURATION', '15m', '15m', '5m', '2h', '5m,10m,15m,30m,1h,2h', false, true),
(1, 'JWT_REFRESH_TOKEN_EXPIRE', 'JWT Refresh Token有效期', '刷新令牌的有效时长', 'DURATION', '7d', '7d', '1d', '30d', '1d,7d,14d,30d', false, true),
(1, 'JWT_SECRET_KEY', 'JWT签名密钥', 'JWT令牌签名密钥（环境变量注入）', 'STRING', '${JWT_SECRET_KEY}', '${JWT_SECRET_KEY}', NULL, NULL, NULL, true, true),
(1, 'PASSWORD_MIN_LENGTH', '密码最小长度', '用户密码的最小长度要求', 'INTEGER', '8', '8', '6', '16', NULL, false, true),
(1, 'PASSWORD_REQUIRE_UPPERCASE', '密码需要大写字母', '是否要求密码包含大写字母', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(1, 'PASSWORD_REQUIRE_LOWERCASE', '密码需要小写字母', '是否要求密码包含小写字母', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(1, 'PASSWORD_REQUIRE_NUMBERS', '密码需要数字', '是否要求密码包含数字', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(1, 'PASSWORD_REQUIRE_SPECIAL', '密码需要特殊字符', '是否要求密码包含特殊字符', 'BOOLEAN', 'false', 'false', NULL, NULL, 'true,false', false, true),
(1, 'LOGIN_MAX_ATTEMPTS', '登录最大失败次数', '账户锁定前允许的最大连续失败登录次数', 'INTEGER', '5', '5', '3', '10', NULL, false, true),
(1, 'LOGIN_LOCKOUT_DURATION', '登录锁定时长', '账户锁定后的自动解锁时长', 'DURATION', '30m', '30m', '5m', '2h', '5m,10m,15m,30m,1h,2h', false, true);

-- notification-service (20 items: SMS 12 + Email 5 + OTP 3)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(2, 'SMS_PROVIDER', '短信提供商', '短信服务提供商', 'ENUM', 'aliyun', 'aliyun', NULL, NULL, 'aliyun,twilio,vonage', false, true),
(2, 'SMS_TEMPLATE_LOGIN', '登录短信模板ID', '登录验证码短信模板ID', 'STRING', 'SMS_LOGIN_TEMPLATE', 'SMS_LOGIN_TEMPLATE', NULL, NULL, NULL, false, true),
(2, 'SMS_TEMPLATE_REGISTER', '注册短信模板ID', '注册验证码短信模板ID', 'STRING', 'SMS_REGISTER_TEMPLATE', 'SMS_REGISTER_TEMPLATE', NULL, NULL, NULL, false, true),
(2, 'SMS_TEMPLATE_RESET', '重置密码短信模板ID', '重置密码验证码短信模板ID', 'STRING', 'SMS_RESET_TEMPLATE', 'SMS_RESET_TEMPLATE', NULL, NULL, NULL, false, true),
(2, 'SMS_CODE_EXPIRE', '短信验证码有效期', '短信验证码有效时长', 'DURATION', '5m', '5m', '1m', '15m', '1m,3m,5m,10m,15m', false, true),
(2, 'SMS_CODE_LENGTH', '短信验证码长度', '短信验证码的位数', 'INTEGER', '6', '6', '4', '8', NULL, false, true),
(2, 'SMS_RATE_LIMIT_PER_USER', '单用户短信频率限制', '每个用户每分钟允许发送的最大短信数', 'RATE_LIMIT', '3/1m', '3/1m', '1/1m', '10/1m', NULL, false, true),
(2, 'SMS_RATE_LIMIT_PER_IP', '单IP短信频率限制', '每个IP每分钟允许发送的最大短信数', 'RATE_LIMIT', '10/1m', '10/1m', '1/1m', '50/1m', NULL, false, true),
(2, 'SMS_CIRCUIT_BREAKER_THRESHOLD', '短信熔断器阈值', '触发熔断的连续失败次数', 'INTEGER', '5', '5', '3', '20', NULL, false, true),
(2, 'SMS_CIRCUIT_BREAKER_RESET_TIMEOUT', '短信熔断器重置时间', '熔断后自动重置为半开状态的时间', 'DURATION', '30s', '30s', '10s', '300s', NULL, false, true),
(2, 'SMS_CIRCUIT_BREAKER_HALF_OPEN_SUCCESS_REQUIRED', '短信半开需成功次数', '半开状态下需连续成功次数以关闭熔断器', 'INTEGER', '3', '3', '1', '10', NULL, false, true),
(2, 'EMAIL_PROVIDER', '邮件提供商', '邮件服务提供商', 'ENUM', 'sendgrid', 'sendgrid', NULL, NULL, 'sendgrid,smtp,ses', false, true),
(2, 'EMAIL_OTP_EXPIRE', '邮件OTP有效期', '邮件一次性密码有效时长', 'DURATION', '5m', '5m', '1m', '30m', '1m,3m,5m,10m,15m,30m', false, true),
(2, 'EMAIL_MAGIC_LINK_EXPIRE', 'Magic Link有效期', '邮件魔法链接有效时长', 'DURATION', '15m', '15m', '5m', '2h', '5m,10m,15m,30m,1h,2h', false, true),
(2, 'EMAIL_RATE_LIMIT_PER_USER', '单用户邮件频率限制', '每个用户每小时允许发送的最大邮件数', 'RATE_LIMIT', '5/1h', '5/1h', '1/1h', '20/1h', NULL, false, true);

-- notification-service (3 more: SMS daily limit, email daily limit, OTP code length)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(2, 'SMS_DAILY_LIMIT', '短信每日上限', '每个手机号每天允许发送的最大短信数量', 'INTEGER', '10', '10', '1', '100', NULL, false, true),
(2, 'EMAIL_DAILY_LIMIT', '邮件每日上限', '每个邮箱每天允许发送的最大邮件数量', 'INTEGER', '10', '10', '1', '100', NULL, false, true),
(2, 'OTP_CODE_LENGTH', 'OTP验证码长度', 'OTP一次性密码的位数', 'INTEGER', '6', '6', '4', '8', NULL, false, true);

-- account-service (8 items: Account 3 + Device 2 + KYB 3 + Referral 5 duplicated but in group account)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(3, 'ACCOUNT_DELETION_FREEZE_DAYS', '注销冻结期', '账户注销后的冻结天数，期间可撤销', 'INTEGER', '30', '30', '7', '90', NULL, false, true),
(3, 'PHONE_NUMBER_MIN_LENGTH', '手机号最小长度', '手机号最低位数要求', 'INTEGER', '10', '10', '7', '15', NULL, false, true),
(3, 'PHONE_NUMBER_MAX_LENGTH', '手机号最大长度', '手机号最高位数限制', 'INTEGER', '15', '15', '10', '20', NULL, false, true),
(3, 'DEVICE_DEFAULT_TRUST_DAYS', '默认可信天数', '新设备的默认可信天数', 'INTEGER', '30', '30', '1', '365', NULL, false, true),
(3, 'DEVICE_MAX_TRUST_DAYS', '最大可信天数', '设备的最大可信天数上限', 'INTEGER', '365', '365', '30', '730', NULL, false, true),
(3, 'KYB_MICRO_DEPOSIT_MIN', '小额打款最小金额', '企业认证小额打款的最小金额（分）', 'INTEGER', '1', '1', '1', '100', NULL, false, true),
(3, 'KYB_MICRO_DEPOSIT_MAX', '小额打款最大金额', '企业认证小额打款的最大金额（分）', 'INTEGER', '99', '99', '1', '999', NULL, false, true),
(3, 'KYB_VERIFICATION_EXPIRE_DAYS', '企业认证验证有效期', '企业认证验证码的有效天数', 'INTEGER', '7', '7', '1', '30', NULL, false, true);

-- credit-service (3 items)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(4, 'CREDIT_MIN_WITHDRAWAL', '最小提现金额', '积分提现的最低金额要求', 'DECIMAL', '1000', '1000', '100', '100000', NULL, false, true),
(4, 'CREDIT_DAILY_WITHDRAWAL_LIMIT', '日提现限额', '每日提现金额上限', 'DECIMAL', '50000', '50000', '1000', '1000000', NULL, false, true),
(4, 'CREDIT_EXPIRY_DAYS', '积分有效期', '积分自获取后的有效天数', 'INTEGER', '365', '365', '30', '730', NULL, false, true);

-- data-product-service (11 items: RFM)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(5, 'RFM_RECENCY_DAYS_1', 'Recency评分1阈值', '最近消费天数≤此值得5分', 'INTEGER', '3', '3', '1', '365', NULL, false, true),
(5, 'RFM_RECENCY_DAYS_2', 'Recency评分2阈值', '最近消费天数≤此值得4分', 'INTEGER', '7', '7', '1', '365', NULL, false, true),
(5, 'RFM_RECENCY_DAYS_3', 'Recency评分3阈值', '最近消费天数≤此值得3分', 'INTEGER', '14', '14', '1', '365', NULL, false, true),
(5, 'RFM_RECENCY_DAYS_4', 'Recency评分4阈值', '最近消费天数≤此值得2分', 'INTEGER', '30', '30', '1', '365', NULL, false, true),
(5, 'RFM_FREQUENCY_TIMES_1', 'Frequency评分1阈值', '消费次数≥此值得5分', 'INTEGER', '10', '10', '1', '1000', NULL, false, true),
(5, 'RFM_FREQUENCY_TIMES_2', 'Frequency评分2阈值', '消费次数≥此值得4分', 'INTEGER', '5', '5', '1', '1000', NULL, false, true),
(5, 'RFM_FREQUENCY_TIMES_3', 'Frequency评分3阈值', '消费次数≥此值得3分', 'INTEGER', '2', '2', '1', '1000', NULL, false, true),
(5, 'RFM_MONETARY_AMOUNT_1', 'Monetary评分1阈值', '消费金额≥此值得5分', 'DECIMAL', '10000', '10000', '1', '9999999', NULL, false, true),
(5, 'RFM_MONETARY_AMOUNT_2', 'Monetary评分2阈值', '消费金额≥此值得4分', 'DECIMAL', '5000', '5000', '1', '9999999', NULL, false, true),
(5, 'RFM_MONETARY_AMOUNT_3', 'Monetary评分3阈值', '消费金额≥此值得3分', 'DECIMAL', '1000', '1000', '1', '9999999', NULL, false, true),
(5, 'RFM_CALCULATION_CRON', 'RFM批量计算周期', 'RFM评分批量计算的Cron表达式', 'CRON', '0 2 * * *', '0 2 * * *', NULL, NULL, '0 2 * * *,0 0 * * *,0 */6 * * *', false, true);

-- compliance-service (19 items: Risk 7 + Audit 2 + Blacklist 3 + Desensitization 7)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(6, 'RISK_LOW_MAX_SCORE', '低风险最大分值', '低风险等级的分数上限', 'INTEGER', '30', '30', '1', '100', NULL, false, true),
(6, 'RISK_MEDIUM_MAX_SCORE', '中风险最大分值', '中风险等级的分数上限', 'INTEGER', '60', '60', '1', '100', NULL, false, true),
(6, 'RISK_HIGH_MAX_SCORE', '高风险最大分值', '高风险等级的分数上限', 'INTEGER', '80', '80', '1', '100', NULL, false, true),
(6, 'RISK_LOCATION_SPEED_THRESHOLD', '位置异常阈值', '两次登录地理位置距离/时间超过此值触发异常', 'INTEGER', '1000', '1000', '100', '5000', NULL, false, true),
(6, 'RISK_LOGIN_FREQUENCY_THRESHOLD', '登录频率阈值', '每小时登录次数超过此值触发异常', 'RATE_LIMIT', '10/1h', '10/1h', '1/1h', '100/1h', NULL, false, true),
(6, 'RISK_AUTO_BLOCK_SCORE', '自动拦截风险分值', '风险分超过此值自动拦截', 'INTEGER', '81', '81', '1', '100', NULL, false, true),
(6, 'RISK_AUTO_VERIFY_SCORE', '需验证风险分值', '风险分超过此值需额外验证', 'INTEGER', '61', '61', '1', '100', NULL, false, true),
(6, 'AUDIT_LOG_RETENTION_DAYS', '审计日志保留期', '审计日志的保留天数', 'INTEGER', '180', '180', '90', '365', NULL, false, true),
(6, 'AUDIT_LOG_BATCH_SIZE', '审计日志批量写入大小', '审计日志的每批写入条数', 'INTEGER', '100', '100', '10', '1000', NULL, false, true),
(6, 'BLACKLIST_IP_TTL', 'IP黑名单有效期', 'IP黑名单条目的有效时长', 'DURATION', '7d', '7d', '1d', '365d', NULL, false, true),
(6, 'BLACKLIST_DEVICE_TTL', '设备黑名单有效期', '设备黑名单条目的有效时长', 'DURATION', '30d', '30d', '1d', '365d', NULL, false, true),
(6, 'BLACKLIST_USER_TTL', '用户黑名单有效期', '用户黑名单条目的有效时长（0=永久）', 'DURATION', '0', '0', '0', '365d', NULL, false, true),
(6, 'DESENSITIZE_ENABLED', '脱敏功能开关', '全局脱敏功能的启用开关', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(6, 'DESENSITIZE_PHONE_ENABLED', '手机号脱敏开关', '对手机号字段进行脱敏处理', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(6, 'DESENSITIZE_EMAIL_ENABLED', '邮箱脱敏开关', '对邮箱字段进行脱敏处理', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(6, 'DESENSITIZE_ID_CARD_ENABLED', '身份证号脱敏开关', '对身份证号字段进行脱敏处理', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(6, 'DESENSITIZE_BANK_CARD_ENABLED', '银行卡号脱敏开关', '对银行卡号字段进行脱敏处理', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(6, 'DESENSITIZE_FIELDS_CUSTOM', '自定义脱敏字段列表', '额外需要脱敏的自定义字段列表', 'LIST', '[]', '[]', NULL, NULL, NULL, false, true);

-- session-related items actually belong to account-service or api-gateway
-- Let me put session items in account-service (group 3) since we don't have a separate session group

-- referral service items in shared group (10) since it spans multiple services
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(10, 'REFERRAL_LEVEL_1_RATE', '一级返佣比例', '直接推荐用户的返佣比例', 'DECIMAL', '0.10', '0.10', '0', '1', NULL, false, true),
(10, 'REFERRAL_LEVEL_2_RATE', '二级返佣比例', '间接推荐用户的返佣比例', 'DECIMAL', '0.05', '0.05', '0', '1', NULL, false, true),
(10, 'REFERRAL_MAX_LEVELS', '最大返佣层级', '返佣的最大层级数', 'INTEGER', '2', '2', '1', '5', NULL, false, true),
(10, 'REFERRAL_SETTLEMENT_DELAY_DAYS', '返佣结算延迟期', '返佣结算的延迟天数', 'INTEGER', '7', '7', '0', '30', NULL, false, true),
(10, 'REFERRAL_MIN_SETTLEMENT_AMOUNT', '最小结算金额', '触发结算的最低累计返佣金额', 'DECIMAL', '1.00', '1.00', '0', '1000', NULL, false, true);

-- mobile-ios (11 items)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(8, 'APP_VERSION_IOS', 'iOS版本号', 'iOS应用当前版本号', 'STRING', '1.0.0', '1.0.0', NULL, NULL, NULL, false, true),
(8, 'APP_FORCE_UPDATE_VERSION_IOS', 'iOS强制更新最低版本', '低于此版本的iOS应用将被强制更新', 'STRING', '1.0.0', '1.0.0', NULL, NULL, NULL, false, true),
(8, 'APP_STORE_URL', 'App Store链接', 'iOS应用在App Store的下载链接', 'STRING', '', '', NULL, NULL, NULL, false, true),
(8, 'THEME_COLOR_BRAND_PRIMARY', '品牌主色调', '应用品牌主色调（Hex）', 'COLOR', '#6C63FF', '#6C63FF', NULL, NULL, NULL, false, true),
(8, 'THEME_COLOR_BRAND_SECONDARY', '品牌次色调', '应用品牌次色调（Hex）', 'COLOR', '#00D4FF', '#00D4FF', NULL, NULL, NULL, false, true),
(8, 'THEME_COLOR_DANGER', '危险色', '危险状态提示色（Hex）', 'COLOR', '#FF4757', '#FF4757', NULL, NULL, NULL, false, true),
(8, 'THEME_COLOR_SUCCESS', '成功色', '成功状态提示色（Hex）', 'COLOR', '#2ED573', '#2ED573', NULL, NULL, NULL, false, true),
(8, 'HOME_RFM_CARD_ENABLED', '首页RFM卡片显示开关', '是否在首页展示RFM评分卡片', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(8, 'CREDITS_REFERRAL_CARD_ENABLED', '积分页推荐卡片显示开关', '是否在积分页展示推荐卡片', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(8, 'ANNOUNCEMENT_TEXT', '公告文案', '应用内公告文字内容', 'STRING', '', '', NULL, NULL, NULL, false, true),
(8, 'ANNOUNCEMENT_ENABLED', '公告开关', '是否显示应用内公告', 'BOOLEAN', 'false', 'false', NULL, NULL, 'true,false', false, true);

-- mobile-android (11 items)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(9, 'APP_VERSION_ANDROID', 'Android版本号', 'Android应用当前版本号', 'STRING', '1.0.0', '1.0.0', NULL, NULL, NULL, false, true),
(9, 'APP_FORCE_UPDATE_VERSION_ANDROID', 'Android强制更新最低版本', '低于此版本的Android应用将被强制更新', 'STRING', '1.0.0', '1.0.0', NULL, NULL, NULL, false, true),
(9, 'GOOGLE_PLAY_URL', 'Google Play链接', 'Android应用在Google Play的下载链接', 'STRING', '', '', NULL, NULL, NULL, false, true),
(9, 'THEME_COLOR_BRAND_PRIMARY_ANDROID', '品牌主色调', '应用品牌主色调（Hex）', 'COLOR', '#6C63FF', '#6C63FF', NULL, NULL, NULL, false, true),
(9, 'THEME_COLOR_BRAND_SECONDARY_ANDROID', '品牌次色调', '应用品牌次色调（Hex）', 'COLOR', '#00D4FF', '#00D4FF', NULL, NULL, NULL, false, true),
(9, 'THEME_COLOR_DANGER_ANDROID', '危险色', '危险状态提示色（Hex）', 'COLOR', '#FF4757', '#FF4757', NULL, NULL, NULL, false, true),
(9, 'THEME_COLOR_SUCCESS_ANDROID', '成功色', '成功状态提示色（Hex）', 'COLOR', '#2ED573', '#2ED573', NULL, NULL, NULL, false, true),
(9, 'HOME_RFM_CARD_ENABLED_ANDROID', '首页RFM卡片显示开关', '是否在首页展示RFM评分卡片', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(9, 'CREDITS_REFERRAL_CARD_ENABLED_ANDROID', '积分页推荐卡片显示开关', '是否在积分页展示推荐卡片', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(9, 'ANNOUNCEMENT_TEXT_ANDROID', '公告文案', '应用内公告文字内容', 'STRING', '', '', NULL, NULL, NULL, false, true),
(9, 'ANNOUNCEMENT_ENABLED_ANDROID', '公告开关', '是否显示应用内公告', 'BOOLEAN', 'false', 'false', NULL, NULL, 'true,false', false, true);

-- session items in api-gateway (7)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(7, 'SESSION_MAX_PER_USER', '单用户最大会话数', '每个用户允许的最大并发会话数', 'INTEGER', '5', '5', '1', '20', NULL, false, true),
(7, 'SESSION_SLIDING_WINDOW_ENABLED', '会话滑动窗口开关', '是否启用会话滑动窗口续期', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true),
(7, 'SESSION_RENEWAL_ADVANCE_TIME', '会话续期提前期', '会话过期前多久自动续期', 'DURATION', '5m', '5m', '1m', '30m', NULL, false, true),
(7, 'SESSION_IDLE_TIMEOUT', '会话空闲超时', '用户无操作后会话自动过期的时长', 'DURATION', '30m', '30m', '5m', '2h', '5m,10m,15m,30m,1h,2h', false, true),
(7, 'SESSION_MAX_CONCURRENT_DEVICES', '最大并发设备数', '同一用户允许的最大并发登录设备数', 'INTEGER', '3', '3', '1', '10', NULL, false, true);

-- Additional config items (11 items to reach 106 total)
-- auth-service (2 more: password history + expiry)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(1, 'PASSWORD_HISTORY_SIZE', '密码历史记录数', '禁止重复使用的历史密码数量', 'INTEGER', '5', '5', '3', '10', NULL, false, true),
(1, 'PASSWORD_EXPIRY_DAYS', '密码有效期', '强制更换密码的间隔天数', 'INTEGER', '90', '90', '30', '365', NULL, false, true);

-- notification-service (2 more: phone/email change templates)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(2, 'SMS_TEMPLATE_CHANGE_PHONE', '换绑手机短信模板ID', '更换手机号验证码的短信模板ID', 'STRING', 'SMS_CHANGE_PHONE_TEMPLATE', 'SMS_CHANGE_PHONE_TEMPLATE', NULL, NULL, NULL, false, true),
(2, 'SMS_TEMPLATE_CHANGE_EMAIL', '换绑邮箱短信模板ID', '更换邮箱验证码的短信模板ID', 'STRING', 'SMS_CHANGE_EMAIL_TEMPLATE', 'SMS_CHANGE_EMAIL_TEMPLATE', NULL, NULL, NULL, false, true);

-- credit-service (2 more: signup/referral bonus)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(4, 'CREDIT_SIGNUP_BONUS', '注册赠送积分', '新用户注册时赠送的积分数', 'INTEGER', '100', '100', '0', '10000', NULL, false, true),
(4, 'CREDIT_REFERRAL_BONUS', '推荐赠送积分', '推荐新用户注册时赠送的积分数', 'INTEGER', '50', '50', '0', '5000', NULL, false, true);

-- compliance-service (1 more: device fingerprint)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(6, 'DEVICE_FINGERPRINT_ENABLED', '设备指纹校验开关', '是否启用设备指纹识别与校验', 'BOOLEAN', 'true', 'true', NULL, NULL, 'true,false', false, true);

-- account-service (1 more: permanent deletion delay)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(3, 'ACCOUNT_DELETION_PERMANENT_DAYS', '永久删除等待期', '冻结期结束后多少天执行永久删除', 'INTEGER', '7', '7', '1', '90', NULL, false, true);

-- data-product-service (1 more: RFM data retention)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(5, 'RFM_DATA_RETENTION_DAYS', 'RFM数据保留期', 'RFM评分数据的保留天数', 'INTEGER', '730', '730', '90', '3650', NULL, false, true);

-- shared (1 more: referral code length)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(10, 'REFERRAL_CODE_LENGTH', '推荐码长度', '用户推荐码的字符长度', 'INTEGER', '8', '8', '4', '16', NULL, false, true);

-- Missing config items from code audit (9 items)
-- api-gateway group (1 more: QR code TTL)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(7, 'QR_CODE_EXPIRE', '二维码过期时间', '二维码登录的有效时长', 'DURATION', '5m', '5m', '1m', '30m', '1m,3m,5m,10m,15m,30m', false, true);

-- account-service (3 more: subscription duration, cache TTL, KYB face threshold)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(3, 'SUBSCRIPTION_DEFAULT_DURATION', '默认订阅时长', '新订阅的默认有效时长', 'DURATION', '720h', '720h', '24h', '8760h', '24h,168h,720h,2160h,8760h', false, true),
(3, 'ENTITLEMENT_CACHE_TTL', '权益缓存有效期', '用户权益数据的缓存时长', 'DURATION', '24h', '24h', '1m', '168h', NULL, false, true),
(3, 'KYB_FACE_SCORE_THRESHOLD', '人脸评分阈值', '企业认证人脸识别的通过分数阈值', 'DECIMAL', '0.8', '0.8', '0', '1', NULL, false, true);

-- compliance-service (2 more: registration/referral rate limits)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(6, 'RISK_REGISTRATION_RATE_LIMIT', '注册频率限制', '每IP每小时允许的最大注册次数', 'INTEGER', '3', '3', '1', '20', NULL, false, true),
(6, 'RISK_MAX_SCORE', '风险最高分值', '风险评分的上限值', 'INTEGER', '100', '100', '1', '1000', NULL, false, true);

-- credit-service (1 more: page size)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(4, 'CREDIT_PAGE_SIZE', '积分分页大小', '积分列表查询的默认每页条数', 'INTEGER', '20', '20', '5', '100', NULL, false, true);

-- Additional credit-service config items (from code audit)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(4, 'CREDIT_DEFAULT_REBATE_RATE', '默认返利比例', '订阅返利的默认百分比（小数表示）', 'DECIMAL', '0.1', '0.1', '0', '1', NULL, false, true),
(4, 'CREDIT_WORKER_POLL_INTERVAL', '工作者轮询间隔', '订阅工作者轮询Redis Stream的间隔', 'DURATION', '2s', '2s', '100ms', '60s', NULL, false, true),
(4, 'CREDIT_WORKER_BATCH_SIZE', '工作者批量大小', '工作者每次批量处理的订阅事件数', 'INTEGER', '10', '10', '1', '100', NULL, false, true);

-- shared (1 more: referral link domain)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(10, 'REFERRAL_LINK_DOMAIN', '推荐链接域名', '用户推荐链接的基础域名', 'STRING', 'https://app.example.com', 'https://app.example.com', NULL, NULL, NULL, false, true);

-- data-product-service (1 more: dashboard trend days)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(5, 'DASHBOARD_TREND_DAYS', '仪表盘趋势天数', '仪表盘注册趋势统计的天数范围', 'INTEGER', '30', '30', '1', '365', NULL, false, true);

-- Default admin user role (system_owner)
INSERT INTO user_roles (user_id, role_id) VALUES ('admin', 1) ON CONFLICT DO NOTHING;

-- compliance-service (7 more: cache TTL, sliding window params, page sizes)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(6, 'BLACKLIST_CACHE_TTL', '黑名单缓存有效期', '黑名单数据的内存缓存时长', 'DURATION', '24h', '24h', '1m', '168h', NULL, false, true),
(6, 'SLIDING_WINDOW_REG_LIMIT', '滑动窗口注册上限', '滑动窗口时间内允许的最大注册数', 'INTEGER', '3', '3', '1', '100', NULL, false, true),
(6, 'SLIDING_WINDOW_REG_WINDOW', '注册滑动窗口时长', '注册频率统计的滑动窗口时长', 'DURATION', '1h', '1h', '1m', '24h', NULL, false, true),
(6, 'SLIDING_WINDOW_REF_ABUSE_LIMIT', '返佣滥用上限', '滑动窗口内允许的最大返佣操作数', 'INTEGER', '50', '50', '1', '1000', NULL, false, true),
(6, 'SLIDING_WINDOW_REF_ABUSE_WINDOW', '返佣滥用窗口时长', '返佣滥用检测的滑动窗口时长', 'DURATION', '1h', '1h', '1m', '24h', NULL, false, true),
(6, 'AUDIT_LOG_DEFAULT_PAGE_SIZE', '审计日志默认分页大小', '审计日志列表查询的默认每页条数', 'INTEGER', '100', '100', '10', '500', NULL, false, true),
(6, 'RISK_HISTORY_DEFAULT_LIMIT', '风险历史默认分页大小', '风险事件历史查询的默认每页条数', 'INTEGER', '100', '100', '10', '500', NULL, false, true);

-- auth-service (1 more: login rate limit)
INSERT INTO config_items (group_id, code, name, description, data_type, current_value, default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled) VALUES
(1, 'LOGIN_RATE_LIMIT_PER_IP', '登录频率限制', '每IP每分钟允许的最大登录尝试次数', 'INTEGER', '10', '10', '1', '100', NULL, false, true);

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
