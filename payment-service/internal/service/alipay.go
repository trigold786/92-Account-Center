package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/smartwalle/alipay/v3"

	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
)

type AlipayConfig struct {
	Mode       string
	AppID      string
	PrivateKey string
	PublicKey  string
	NotifyURL  string
	IsSandbox  bool
}

func (c AlipayConfig) ValidateProduction() error {
	if c.Mode != "production" {
		return nil
	}
	fields := map[string]string{
		"ALIPAY_APP_ID":      c.AppID,
		"ALIPAY_PRIVATE_KEY": c.PrivateKey,
		"ALIPAY_PUBLIC_KEY":  c.PublicKey,
		"ALIPAY_NOTIFY_URL":  c.NotifyURL,
	}
	for name, value := range fields {
		if insecurePaymentValue(value) {
			return fmt.Errorf("production alipay payment config requires real %s", name)
		}
	}
	return nil
}

type AlipayProvider struct {
	client *alipay.Client
	cfg    AlipayConfig
}

func NewAlipayProvider(cfg AlipayConfig) *AlipayProvider {
	p := &AlipayProvider{cfg: cfg}

	client, err := alipay.New(cfg.AppID, cfg.PrivateKey, !cfg.IsSandbox)
	if err != nil {
		return p
	}

	if cfg.PublicKey != "" {
		if err := client.LoadAliPayPublicKey(cfg.PublicKey); err != nil {
			return p
		}
	}

	p.client = client
	return p
}

func (p *AlipayProvider) Name() string {
	return "alipay"
}

func (p *AlipayProvider) CreatePayment(ctx context.Context, req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	switch req.TradeType {
	case "alipay_wap", "alipay_app":
	default:
		return nil, fmt.Errorf("unsupported alipay trade type: %s", req.TradeType)
	}

	if p.client == nil {
		return p.sandboxCreatePayment(req)
	}

	pay := alipay.TradeWapPay{}
	pay.Subject = req.Subject
	pay.OutTradeNo = req.OrderNo
	pay.TotalAmount = fmt.Sprintf("%.2f", req.Amount)
	pay.ProductCode = "QUICK_WAP_WAY"
	pay.NotifyURL = p.cfg.NotifyURL

	url, err := p.client.TradeWapPay(pay)
	if err != nil {
		return nil, fmt.Errorf("alipay trade wap pay: %w", err)
	}

	return &provider.CreatePaymentResponse{
		PaymentURL:    url.String(),
		TransactionID: req.OrderNo,
	}, nil
}

func (p *AlipayProvider) sandboxCreatePayment(req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	prepayID := fmt.Sprintf("alipay%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%100000)
	paymentURL := fmt.Sprintf("https://openapi.alipay.com/gateway.do?method=trade.pay&app_id=%s&prepay_id=%s", p.cfg.AppID, prepayID)

	return &provider.CreatePaymentResponse{
		PaymentURL:    paymentURL,
		PrepayID:      prepayID,
		TransactionID: fmt.Sprintf("ALI%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
	}, nil
}

func (p *AlipayProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	if p.client == nil {
		return &provider.PaymentStatus{
			OrderNo:       orderNo,
			TransactionID: fmt.Sprintf("ALI%s", orderNo),
			Status:        "WAIT_BUYER_PAY",
			Amount:        0,
		}, nil
	}

	result, err := p.client.TradeQuery(ctx, alipay.TradeQuery{
		OutTradeNo: orderNo,
	})
	if err != nil {
		return nil, fmt.Errorf("alipay trade query: %w", err)
	}

	status := "WAIT_BUYER_PAY"
	if result.TradeStatus != "" {
		status = string(result.TradeStatus)
	}

	return &provider.PaymentStatus{
		OrderNo:       orderNo,
		TransactionID: result.TradeNo,
		Status:        status,
		Amount:        0,
	}, nil
}

func (p *AlipayProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	if p.client == nil {
		return &provider.RefundResponse{
			RefundNo: req.RefundNo,
			Status:   "REFUND_SUCCESS",
			RefundID: fmt.Sprintf("ALIREFUND%s", req.RefundNo),
		}, nil
	}

	_, err := p.client.TradeRefund(ctx, alipay.TradeRefund{
		OutTradeNo:   req.OrderNo,
		OutRequestNo: req.RefundNo,
		RefundAmount: fmt.Sprintf("%.2f", req.RefundAmount),
	})
	if err != nil {
		return nil, fmt.Errorf("alipay trade refund: %w", err)
	}

	return &provider.RefundResponse{
		RefundNo: req.RefundNo,
		Status:   "REFUND_SUCCESS",
		RefundID: fmt.Sprintf("ALIREFUND%s", req.RefundNo),
	}, nil
}

func (p *AlipayProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	if err := p.verifyCallbackSignature(headers, body); err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid callback body: %w", err)
	}

	result := &provider.CallbackResult{
		RawData: string(body),
	}

	if orderNo, ok := payload["out_trade_no"].(string); ok {
		result.OrderNo = orderNo
	}
	if txnID, ok := payload["trade_no"].(string); ok {
		result.TransactionID = txnID
	}
	if tradeStatus, ok := payload["trade_status"].(string); ok {
		if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
			result.Status = "SUCCESS"
		} else {
			result.Status = "FAIL"
		}
	} else {
		result.Status = "SUCCESS"
	}
	if totalStr, ok := payload["total_amount"].(string); ok {
		fmt.Sscanf(totalStr, "%f", &result.Amount)
	} else if total, ok := payload["total_amount"].(float64); ok {
		result.Amount = total
	}
	if gmtPayment, ok := payload["gmt_payment"].(string); ok {
		result.PaidAt = gmtPayment
	}

	return result, nil
}

func (p *AlipayProvider) verifyCallbackSignature(headers map[string]string, body []byte) error {
	signature := firstHeader(headers, "Alipay-Signature", "alipay-signature")
	if signature == "" {
		return fmt.Errorf("missing alipay callback signature")
	}
	mac := hmac.New(sha256.New, []byte(p.cfg.PrivateKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid alipay callback signature")
	}
	return nil
}
