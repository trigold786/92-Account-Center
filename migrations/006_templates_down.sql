-- +goose Down for templates
DROP TABLE IF EXISTS notification_templates CASCADE;
DROP TABLE IF EXISTS send_records CASCADE;
