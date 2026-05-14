-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    account_id VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(100),
    last_strong_auth_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deletion_requested_at TIMESTAMP WITH TIME ZONE,
    deletion_expires_at TIMESTAMP WITH TIME ZONE,
    deletion_cancelled_at TIMESTAMP WITH TIME ZONE,
    deletion_deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_phone_number ON users(phone_number);
CREATE INDEX idx_users_account_id ON users(account_id);
CREATE INDEX idx_users_email ON users(email);

CREATE TABLE enterprises (
    enterprise_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    company_name VARCHAR(200) NOT NULL,
    unified_social_credit_code VARCHAR(50) NOT NULL UNIQUE,
    legal_person_name VARCHAR(100) NOT NULL,
    legal_person_id_number VARCHAR(255) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    bank_account_number VARCHAR(255) NOT NULL,
    verification_status VARCHAR(20) DEFAULT 'pending',
    micro_payment_status VARCHAR(20) DEFAULT 'pending',
    micro_payment_amount DECIMAL(10,2),
    face_verification_status VARCHAR(20) DEFAULT 'pending',
    face_verification_score DECIMAL(5,2),
    sub_account_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_enterprises_user_id ON enterprises(user_id);
CREATE INDEX idx_enterprises_credit_code ON enterprises(unified_social_credit_code);

CREATE TABLE sub_accounts (
    id SERIAL PRIMARY KEY,
    enterprise_id UUID NOT NULL REFERENCES enterprises(enterprise_id),
    user_id UUID NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    status VARCHAR(20) DEFAULT 'active',
    last_liveness_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sub_accounts_enterprise_id ON sub_accounts(enterprise_id);

CREATE TABLE audit_logs (
    log_id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    target_resource VARCHAR(200),
    source_ip VARCHAR(50),
    result VARCHAR(20) NOT NULL,
    details JSONB,
    sm3_hash VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_time ON audit_logs(event_time);
CREATE INDEX idx_audit_logs_action_type ON audit_logs(action_type);

CREATE TABLE risk_events (
    risk_event_id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    risk_score INT NOT NULL DEFAULT 0,
    risk_level VARCHAR(20) NOT NULL,
    ip_address VARCHAR(50),
    location JSONB,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_risk_events_user_id ON risk_events(user_id);
CREATE INDEX idx_risk_events_created_at ON risk_events(created_at);

-- +goose Down
DROP TABLE IF EXISTS risk_events;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sub_accounts;
DROP TABLE IF EXISTS enterprises;
DROP TABLE IF EXISTS users;
