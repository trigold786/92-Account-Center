-- +goose Up
CREATE TABLE blacklist_entries (
    id BIGSERIAL PRIMARY KEY,
    entry_type VARCHAR(20) NOT NULL CHECK (entry_type IN ('IP', 'DEVICE', 'PHONE', 'ACCOUNT')),
    entry_value VARCHAR(200) NOT NULL,
    reason VARCHAR(500) NOT NULL,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system',
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_blacklist_type_value ON blacklist_entries(entry_type, entry_value);
CREATE INDEX idx_blacklist_expires ON blacklist_entries(expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE rebate_configs (
    id BIGSERIAL PRIMARY KEY,
    subscription_count_min INT NOT NULL DEFAULT 0,
    subscription_count_max INT NOT NULL,
    rebate_percentage DECIMAL(5,4) NOT NULL CHECK (rebate_percentage >= 0 AND rebate_percentage <= 1),
    description VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO rebate_configs (subscription_count_min, subscription_count_max, rebate_percentage, description) VALUES
(0, 0, 0.50, 'First subscription rebate 50%'),
(1, 4, 0.30, '2nd-5th subscription rebate 30%'),
(5, 9, 0.20, '6th-10th subscription rebate 20%'),
(10, 999999, 0.10, '11th+ subscription rebate 10%');

-- +goose Down
DROP TABLE IF EXISTS rebate_configs;
DROP TABLE IF EXISTS blacklist_entries;
