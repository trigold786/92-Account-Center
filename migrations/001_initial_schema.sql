-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    account_id VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deletion_requested_at TIMESTAMP WITH TIME ZONE,
    deletion_expires_at TIMESTAMP WITH TIME ZONE,
    deletion_cancelled_at TIMESTAMP WITH TIME ZONE,
    deletion_deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for faster lookups
CREATE INDEX idx_users_phone_number ON users(phone_number);
CREATE INDEX idx_users_account_id ON users(account_id);
CREATE INDEX idx_users_deletion_requested_at ON users(deletion_requested_at);