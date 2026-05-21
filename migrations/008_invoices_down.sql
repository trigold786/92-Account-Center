-- +goose Down for invoices
DROP TABLE IF EXISTS invoices CASCADE;
