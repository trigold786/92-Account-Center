package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
)

type SubscriptionActivationRequest struct {
	UserID        int64   `json:"user_id"`
	OrderID       string  `json:"order_id"`
	TierLevel     int     `json:"tier_level"`
	Price         float64 `json:"price"`
	PaymentMethod string  `json:"payment_method"`
}

type SubscriptionActivationNotifier interface {
	NotifySubscriptionActivation(ctx context.Context, req SubscriptionActivationRequest) error
}

type CallbackHandler struct {
	providerRegistry *provider.ProviderRegistry
	orderRepo        repository.OrderRepository
	orderSvc         service.OrderService
	callbackRepo     repository.PaymentCallbackRepository
	activationNotify SubscriptionActivationNotifier
	logger           *slog.Logger
}

func NewCallbackHandler(
	providerRegistry *provider.ProviderRegistry,
	orderRepo repository.OrderRepository,
	orderSvc service.OrderService,
	logger *slog.Logger,
) *CallbackHandler {
	return NewCallbackHandlerWithCallbackRepository(providerRegistry, orderRepo, orderSvc, nil, logger)
}

func NewCallbackHandlerWithCallbackRepository(
	providerRegistry *provider.ProviderRegistry,
	orderRepo repository.OrderRepository,
	orderSvc service.OrderService,
	callbackRepo repository.PaymentCallbackRepository,
	logger *slog.Logger,
) *CallbackHandler {
	return NewCallbackHandlerWithActivationNotifier(providerRegistry, orderRepo, orderSvc, callbackRepo, nil, logger)
}

func NewCallbackHandlerWithActivationNotifier(
	providerRegistry *provider.ProviderRegistry,
	orderRepo repository.OrderRepository,
	orderSvc service.OrderService,
	callbackRepo repository.PaymentCallbackRepository,
	activationNotify SubscriptionActivationNotifier,
	logger *slog.Logger,
) *CallbackHandler {
	return &CallbackHandler{
		providerRegistry: providerRegistry,
		orderRepo:        orderRepo,
		orderSvc:         orderSvc,
		callbackRepo:     callbackRepo,
		activationNotify: activationNotify,
		logger:           logger,
	}
}

func (h *CallbackHandler) HandleCallback(c *gin.Context) {
	providerName := c.Param("provider")

	p, ok := h.providerRegistry.Get(providerName)
	if !ok {
		h.logger.Error("unknown provider in callback", "provider", providerName)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "unknown provider"})
		return
	}

	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	body, err := c.GetRawData()
	if err != nil {
		h.logger.Error("failed to read callback body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "failed to read body"})
		return
	}

	result, err := p.VerifyCallback(c.Request.Context(), headers, body)
	if err != nil {
		h.logger.Error("callback verification failed", "provider", providerName, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "verification failed"})
		return
	}

	if h.callbackRepo != nil {
		callback := &model.PaymentCallback{
			Provider:      providerName,
			OrderNo:       result.OrderNo,
			TransactionID: result.TransactionID,
			Status:        result.Status,
			Verified:      true,
			RawPayload:    string(body),
		}
		if err := h.callbackRepo.Save(c.Request.Context(), callback); err != nil {
			h.logger.Error("failed to persist payment callback", "provider", providerName, "order_no", result.OrderNo, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to persist callback"})
			return
		}
	}

	h.logger.Info("payment callback received",
		"provider", providerName,
		"order_no", result.OrderNo,
		"status", result.Status,
		"transaction_id", result.TransactionID,
	)

	order, err := h.orderRepo.GetByOrderNo(c.Request.Context(), result.OrderNo)
	if err != nil {
		h.logger.Error("failed to query order", "order_no", result.OrderNo, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}
	if order == nil {
		h.logger.Warn("order not found for callback", "order_no", result.OrderNo)
		c.String(http.StatusOK, "success")
		return
	}

	if order.Status == model.OrderStatusPaid {
		h.logger.Info("order already paid, idempotent callback", "order_no", result.OrderNo)
		c.String(http.StatusOK, "success")
		return
	}

	if result.Status == "SUCCESS" {
		req := &model.UpdateOrderStatusRequest{
			Status:               model.OrderStatusPaid,
			PaymentMethod:        providerName,
			PaymentTransactionID: result.TransactionID,
		}
		if _, err := h.orderSvc.UpdateStatus(c.Request.Context(), order.ID, req); err != nil {
			h.logger.Error("failed to update order status from callback",
				"order_no", result.OrderNo, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to update order"})
			return
		}
		if err := h.notifySubscriptionActivation(c.Request.Context(), order, providerName); err != nil {
			h.logger.Error("failed to notify subscription activation", "order_no", result.OrderNo, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to activate subscription"})
			return
		}
	}

	c.String(http.StatusOK, "success")
}

func (h *CallbackHandler) notifySubscriptionActivation(ctx context.Context, order *model.Order, paymentMethod string) error {
	if h.activationNotify == nil || order.ProductType != "subscription" {
		return nil
	}
	var metadata struct {
		TierLevel int `json:"tier_level"`
	}
	if order.Metadata != "" {
		if err := json.Unmarshal([]byte(order.Metadata), &metadata); err != nil {
			return err
		}
	}
	if metadata.TierLevel == 0 {
		return fmt.Errorf("subscription order metadata missing tier_level")
	}
	return h.activationNotify.NotifySubscriptionActivation(ctx, SubscriptionActivationRequest{
		UserID:        order.UserID,
		OrderID:       strconv.FormatInt(order.ID, 10),
		TierLevel:     metadata.TierLevel,
		Price:         order.Amount,
		PaymentMethod: paymentMethod,
	})
}

type CreatePaymentHandler struct {
	providerRegistry *provider.ProviderRegistry
	orderSvc         service.OrderService
	logger           *slog.Logger
}

func NewCreatePaymentHandler(
	providerRegistry *provider.ProviderRegistry,
	orderSvc service.OrderService,
	logger *slog.Logger,
) *CreatePaymentHandler {
	return &CreatePaymentHandler{
		providerRegistry: providerRegistry,
		orderSvc:         orderSvc,
		logger:           logger,
	}
}

type CreatePaymentAPIRequest struct {
	OrderID   int64  `json:"order_id" binding:"required"`
	TradeType string `json:"trade_type" binding:"required"`
	ReturnURL string `json:"return_url,omitempty"`
	OpenID    string `json:"open_id,omitempty"`
	NotifyURL string `json:"notify_url,omitempty"`
	ClientIP  string `json:"client_ip,omitempty"`
}

func (h *CreatePaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	order, err := h.orderSvc.GetOrder(c.Request.Context(), req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "order not found"})
		return
	}

	providerName := resolveProvider(req.TradeType)
	p, ok := h.providerRegistry.Get(providerName)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("unsupported trade type: %s", req.TradeType)})
		return
	}

	notifyURL := req.NotifyURL
	if notifyURL == "" {
		notifyURL = fmt.Sprintf("http://localhost:30316/api/v1/payment/callback/%s", providerName)
	}

	paymentReq := &provider.CreatePaymentRequest{
		OrderNo:   order.OrderNo,
		Amount:    order.Amount,
		Currency:  order.Currency,
		Subject:   order.ProductName,
		TradeType: req.TradeType,
		ClientIP:  req.ClientIP,
		NotifyURL: notifyURL,
		ReturnURL: req.ReturnURL,
		OpenID:    req.OpenID,
	}

	resp, err := p.CreatePayment(c.Request.Context(), paymentReq)
	if err != nil {
		h.logger.Error("failed to create payment", "error", err, "provider", providerName)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "payment creation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}

func resolveProvider(tradeType string) string {
	switch {
	case tradeType == "wechat_h5" || tradeType == "wechat_mini" || tradeType == "wechat_native":
		return "wechat"
	case tradeType == "alipay_wap" || tradeType == "alipay_app":
		return "alipay"
	default:
		return tradeType
	}
}
