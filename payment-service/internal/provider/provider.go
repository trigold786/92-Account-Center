package provider

import (
	"context"
	"sync"
)

type PaymentProvider interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)
	QueryPayment(ctx context.Context, orderNo string) (*PaymentStatus, error)
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
	VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*CallbackResult, error)
	Name() string
}

type CreatePaymentRequest struct {
	OrderNo     string  `json:"order_no"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Subject     string  `json:"subject"`
	Description string  `json:"description,omitempty"`
	TradeType   string  `json:"trade_type"`
	ClientIP    string  `json:"client_ip,omitempty"`
	NotifyURL   string  `json:"notify_url"`
	ReturnURL   string  `json:"return_url,omitempty"`
	OpenID      string  `json:"open_id,omitempty"`
}

type CreatePaymentResponse struct {
	PaymentURL    string `json:"payment_url"`
	PrepayID      string `json:"prepay_id"`
	QRCodeURL     string `json:"qrcode_url,omitempty"`
	TransactionID string `json:"transaction_id"`
}

type PaymentStatus struct {
	OrderNo       string  `json:"order_no"`
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	PaidAt        string  `json:"paid_at,omitempty"`
}

type RefundRequest struct {
	OrderNo      string  `json:"order_no"`
	RefundNo     string  `json:"refund_no"`
	TotalAmount  float64 `json:"total_amount"`
	RefundAmount float64 `json:"refund_amount"`
	Reason       string  `json:"reason"`
}

type RefundResponse struct {
	RefundNo string `json:"refund_no"`
	Status   string `json:"status"`
	RefundID string `json:"refund_id"`
}

type CallbackResult struct {
	OrderNo       string  `json:"order_no"`
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	PaidAt        string  `json:"paid_at,omitempty"`
	RawData       string  `json:"raw_data"`
}

type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]PaymentProvider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{providers: make(map[string]PaymentProvider)}
}

func (r *ProviderRegistry) Register(p PaymentProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *ProviderRegistry) Get(name string) (PaymentProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

func (r *ProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
