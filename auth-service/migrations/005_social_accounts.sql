-- +goose Up
CREATE TABLE social_accounts (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider      VARCHAR(32) NOT NULL,
    provider_uid  VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    avatar_url    TEXT,
    access_token  TEXT,
    refresh_token TEXT,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_uid)
);
CREATE INDEX idx_social_accounts_user_id ON social_accounts(user_id);

-- +goose Down
DROP TABLE IF EXISTS social_accounts;
