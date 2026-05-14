-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS identity_tier INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE';

CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    tier_level INT NOT NULL CHECK (tier_level IN (2, 3, 4)),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'EXPIRED', 'CANCELED')),
    price DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(50),
    order_id VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_end_time ON subscriptions(end_time);

CREATE TABLE entitlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    feature_code VARCHAR(100) NOT NULL,
    total_quota INT NOT NULL DEFAULT 0,
    used_quota INT NOT NULL DEFAULT 0,
    reset_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, feature_code)
);

CREATE INDEX idx_entitlements_user_id ON entitlements(user_id);

CREATE TABLE credit_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_accounts_user_id ON credit_accounts(user_id);

CREATE TABLE credit_transactions (
    id BIGSERIAL PRIMARY KEY,
    credit_account_id BIGINT NOT NULL REFERENCES credit_accounts(id) ON DELETE RESTRICT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('EARN_REFERRAL', 'EARN_VERIFY', 'CONSUME_SUB', 'REFUND_SUB', 'EXPIRED')),
    amount DECIMAL(12, 2) NOT NULL,
    reference_id VARCHAR(100),
    details JSONB,
    sm3_hash VARCHAR(128) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE' CHECK (status IN ('AVAILABLE', 'PENDING', 'FROZEN', 'CONSUMED', 'EXPIRED', 'REJECTED')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_transactions_credit_account_id ON credit_transactions(credit_account_id);
CREATE INDEX idx_credit_transactions_reference_id ON credit_transactions(reference_id);
CREATE INDEX idx_credit_transactions_type ON credit_transactions(type);

CREATE TABLE referral_relations (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    referee_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    referee_subscription_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_referral_relations_referrer_id ON referral_relations(referrer_id);
CREATE INDEX idx_referral_relations_referee_id ON referral_relations(referee_id);

CREATE TABLE device_fingerprints (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(100) NOT NULL,
    fingerprint_hash VARCHAR(128) NOT NULL UNIQUE,
    device_info TEXT,
    ip_address VARCHAR(45),
    last_login_at TIMESTAMP WITH TIME ZONE,
    trusted_until TIMESTAMP WITH TIME ZONE,
    is_trusted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_device_fingerprints_user_id ON device_fingerprints(user_id);

-- +goose Down
DROP TABLE IF EXISTS device_fingerprints;
DROP TABLE IF EXISTS referral_relations;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS credit_accounts;
DROP TABLE IF EXISTS entitlements;
DROP TABLE IF EXISTS subscriptions;

ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS identity_tier;
