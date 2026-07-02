-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    product_type VARCHAR(50) NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'cancelled', 'refunded')),
    payment_method VARCHAR(50) NOT NULL DEFAULT '',
    payment_transaction_id VARCHAR(128) NOT NULL DEFAULT '',
    paid_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    refunded_at TIMESTAMP WITH TIME ZONE,
    refund_reason TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_payment_method ON orders(payment_method);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_expires_at ON orders(expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS refunds (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    reason TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'failed')),
    refund_no VARCHAR(64) NOT NULL DEFAULT '',
    provider VARCHAR(50) NOT NULL DEFAULT '',
    provider_refund_id VARCHAR(128) NOT NULL DEFAULT '',
    provider_status VARCHAR(50) NOT NULL DEFAULT '',
    provider_error TEXT NOT NULL DEFAULT '',
    approver_id BIGINT,
    review_note TEXT NOT NULL DEFAULT '',
    approved_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(order_id)
);

CREATE INDEX IF NOT EXISTS idx_refunds_user_id ON refunds(user_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);
CREATE INDEX IF NOT EXISTS idx_refunds_created_at ON refunds(created_at);
CREATE INDEX IF NOT EXISTS idx_refunds_provider ON refunds(provider);
CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_refund_no ON refunds(refund_no) WHERE refund_no <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_provider_refund_id ON refunds(provider_refund_id) WHERE provider_refund_id <> '';

CREATE TABLE IF NOT EXISTS invoices (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    invoice_no VARCHAR(64) NOT NULL UNIQUE,
    title VARCHAR(200) NOT NULL,
    tax_id VARCHAR(64) NOT NULL DEFAULT '',
    email VARCHAR(200) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'issued', 'failed', 'cancelled')),
    file_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON invoices(user_id);
CREATE INDEX IF NOT EXISTS idx_invoices_order_id ON invoices(order_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_created_at ON invoices(created_at);

CREATE TABLE IF NOT EXISTS payment_callbacks (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    order_no VARCHAR(64) NOT NULL,
    transaction_id VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    raw_payload TEXT NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_callbacks_order_no ON payment_callbacks(order_no);
CREATE INDEX IF NOT EXISTS idx_payment_callbacks_provider ON payment_callbacks(provider);
CREATE INDEX IF NOT EXISTS idx_payment_callbacks_received_at ON payment_callbacks(received_at);

CREATE TABLE IF NOT EXISTS reconciliation_reports (
    id VARCHAR(100) PRIMARY KEY,
    provider_name VARCHAR(50) NOT NULL,
    report_date DATE NOT NULL,
    total_orders INT NOT NULL DEFAULT 0,
    matched_orders INT NOT NULL DEFAULT 0,
    mismatch_orders JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_reports_provider_date ON reconciliation_reports(provider_name, report_date);

-- +goose Down
DROP TABLE IF EXISTS reconciliation_reports;
DROP TABLE IF EXISTS payment_callbacks;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS orders;
