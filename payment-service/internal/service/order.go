package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
	"github.com/trigold786/92-Account-Center/payment-service/internal/svcconfig"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidTransition  = errors.New("invalid status transition")
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error)
	GetOrder(ctx context.Context, id int64) (*model.Order, error)
	GetOrderByNo(ctx context.Context, orderNo string) (*model.Order, error)
	ListOrders(ctx context.Context, query *model.OrderQueryRequest) (*model.OrderListResponse, error)
	UpdateStatus(ctx context.Context, id int64, req *model.UpdateOrderStatusRequest) (*model.Order, error)
	ExportCSV(ctx context.Context, query *model.OrderQueryRequest) ([]byte, error)
}

type orderService struct {
	repo repository.OrderRepository
	cfg  *svcconfig.PaymentConfig
}

func NewOrderService(repo repository.OrderRepository, cfg *svcconfig.PaymentConfig) OrderService {
	return &orderService{repo: repo, cfg: cfg}
}

func (s *orderService) CreateOrder(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error) {
	order := &model.Order{
		UserID:      req.UserID,
		ProductType: req.ProductType,
		ProductName: req.ProductName,
		Amount:      req.Amount,
		Currency:    req.Currency,
		ExpiresAt:   req.ExpiresAt,
		Metadata:    req.Metadata,
		Status:      model.OrderStatusPending,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) GetOrderByNo(ctx context.Context, orderNo string) (*model.Order, error) {
	order, err := s.repo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) ListOrders(ctx context.Context, query *model.OrderQueryRequest) (*model.OrderListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = s.cfg.DefaultPageSize
	}

	orders, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []model.Order{}
	}

	return &model.OrderListResponse{
		Orders:   orders,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *orderService) UpdateStatus(ctx context.Context, id int64, req *model.UpdateOrderStatusRequest) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	if !isValidTransition(order.Status, req.Status) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, order.Status, req.Status)
	}

	switch req.Status {
	case model.OrderStatusRefunded:
		if err := s.repo.UpdateRefund(ctx, id, req.RefundReason); err != nil {
			return nil, err
		}
	default:
		if err := s.repo.UpdateStatus(ctx, id, req.Status, req.PaymentMethod, req.PaymentTransactionID); err != nil {
			return nil, err
		}
	}

	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ExportCSV(ctx context.Context, query *model.OrderQueryRequest) ([]byte, error) {
	query.Page = 1
	query.PageSize = 10000

	orders, _, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"id", "order_no", "user_id", "product_type", "product_name", "amount", "currency", "status", "payment_method", "payment_transaction_id", "created_at", "updated_at"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, o := range orders {
		record := []string{
			strconv.FormatInt(o.ID, 10),
			o.OrderNo,
			strconv.FormatInt(o.UserID, 10),
			o.ProductType,
			o.ProductName,
			strconv.FormatFloat(o.Amount, 'f', 2, 64),
			o.Currency,
			string(o.Status),
			o.PaymentMethod,
			o.PaymentTransactionID,
			o.CreatedAt.Format("2006-01-02 15:04:05"),
			o.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func isValidTransition(from, to model.OrderStatus) bool {
	switch {
	case from == model.OrderStatusPending && to == model.OrderStatusPaid:
		return true
	case from == model.OrderStatusPending && to == model.OrderStatusCancelled:
		return true
	case from == model.OrderStatusPaid && to == model.OrderStatusRefunded:
		return true
	default:
		return false
	}
}
