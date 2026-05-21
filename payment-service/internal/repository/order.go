package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
	List(ctx context.Context, query *model.OrderQueryRequest) ([]model.Order, int, error)
	UpdateStatus(ctx context.Context, id int64, status model.OrderStatus, paymentMethod, paymentTxnID string) error
	UpdateRefund(ctx context.Context, id int64, reason string) error
	FindExpired(ctx context.Context, before time.Time) ([]model.Order, error)
	GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	orderNo := generateOrderNo()
	now := time.Now()
	order.OrderNo = orderNo
	order.CreatedAt = now
	order.UpdatedAt = now
	if order.Currency == "" {
		order.Currency = "CNY"
	}
	if order.Status == "" {
		order.Status = model.OrderStatusPending
	}

	query := `
		INSERT INTO orders (order_no, user_id, product_type, product_name, amount, currency, status,
			expires_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		order.OrderNo, order.UserID, order.ProductType, order.ProductName,
		order.Amount, order.Currency, order.Status,
		order.ExpiresAt, order.Metadata, order.CreatedAt, order.UpdatedAt,
	).Scan(&order.ID)
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	order := &model.Order{}
	query := `SELECT id, order_no, user_id, product_type, product_name, amount, currency, status,
		payment_method, payment_transaction_id, paid_at, cancelled_at, refunded_at, refund_reason,
		expires_at, metadata, created_at, updated_at
		FROM orders WHERE id = $1`
	err := r.scanOrder(r.db.QueryRowContext(ctx, query, id), order)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	order := &model.Order{}
	query := `SELECT id, order_no, user_id, product_type, product_name, amount, currency, status,
		payment_method, payment_transaction_id, paid_at, cancelled_at, refunded_at, refund_reason,
		expires_at, metadata, created_at, updated_at
		FROM orders WHERE order_no = $1`
	err := r.scanOrder(r.db.QueryRowContext(ctx, query, orderNo), order)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) List(ctx context.Context, query *model.OrderQueryRequest) ([]model.Order, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if query.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *query.UserID)
		argIdx++
	}
	if query.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*query.Status))
		argIdx++
	}
	if query.PaymentMethod != "" {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIdx))
		args = append(args, query.PaymentMethod)
		argIdx++
	}
	if query.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *query.StartTime)
		argIdx++
	}
	if query.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *query.EndTime)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM orders " + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	listQuery := `SELECT id, order_no, user_id, product_type, product_name, amount, currency, status,
		payment_method, payment_transaction_id, paid_at, cancelled_at, refunded_at, refund_reason,
		expires_at, metadata, created_at, updated_at
		FROM orders ` + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := r.scanRow(rows, &order); err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}
	return orders, total, rows.Err()
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus, paymentMethod, paymentTxnID string) error {
	var query string
	var args []interface{}
	switch status {
	case model.OrderStatusPaid:
		query = `UPDATE orders SET status = $1, payment_method = $2, payment_transaction_id = $3, paid_at = NOW(), updated_at = NOW() WHERE id = $4`
		args = []interface{}{string(status), paymentMethod, paymentTxnID, id}
	case model.OrderStatusCancelled:
		query = `UPDATE orders SET status = $1, cancelled_at = NOW(), updated_at = NOW() WHERE id = $2`
		args = []interface{}{string(status), id}
	default:
		query = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
		args = []interface{}{string(status), id}
	}
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

func (r *orderRepository) UpdateRefund(ctx context.Context, id int64, reason string) error {
	query := `UPDATE orders SET status = $1, refund_reason = $2, refunded_at = NOW(), updated_at = NOW() WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, string(model.OrderStatusRefunded), reason, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

func (r *orderRepository) scanOrder(row *sql.Row, order *model.Order) error {
	return row.Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.ProductType, &order.ProductName,
		&order.Amount, &order.Currency, &order.Status,
		&order.PaymentMethod, &order.PaymentTransactionID,
		&order.PaidAt, &order.CancelledAt, &order.RefundedAt, &order.RefundReason,
		&order.ExpiresAt, &order.Metadata, &order.CreatedAt, &order.UpdatedAt,
	)
}

func (r *orderRepository) scanRow(rows *sql.Rows, order *model.Order) error {
	return rows.Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.ProductType, &order.ProductName,
		&order.Amount, &order.Currency, &order.Status,
		&order.PaymentMethod, &order.PaymentTransactionID,
		&order.PaidAt, &order.CancelledAt, &order.RefundedAt, &order.RefundReason,
		&order.ExpiresAt, &order.Metadata, &order.CreatedAt, &order.UpdatedAt,
	)
}

func (r *orderRepository) FindExpired(ctx context.Context, before time.Time) ([]model.Order, error) {
	query := `SELECT id, order_no, user_id, product_type, product_name, amount, currency, status,
		payment_method, payment_transaction_id, paid_at, cancelled_at, refunded_at, refund_reason,
		expires_at, metadata, created_at, updated_at
		FROM orders WHERE status = $1 AND expires_at IS NOT NULL AND expires_at < $2`
	rows, err := r.db.QueryContext(ctx, query, string(model.OrderStatusPending), before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := r.scanRow(rows, &order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (r *orderRepository) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	cutoff := time.Now().Add(-since)
	query := `SELECT id, order_no, user_id, product_type, product_name, amount, currency, status,
		payment_method, payment_transaction_id, paid_at, cancelled_at, refunded_at, refund_reason,
		expires_at, metadata, created_at, updated_at
		FROM orders WHERE status = $1 AND created_at < $2`
	rows, err := r.db.QueryContext(ctx, query, string(model.OrderStatusPending), cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		if err := r.scanRow(rows, order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE order_no = $2 AND status = $3`
	result, err := r.db.ExecContext(ctx, query, toStatus, orderNo, fromStatus)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found or status mismatch")
	}
	return nil
}

func generateOrderNo() string {
	return fmt.Sprintf("PAY%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
