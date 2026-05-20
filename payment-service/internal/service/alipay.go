package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
)

type AlipayConfig struct {
	AppID      string
	PrivateKey string
	PublicKey  string
	NotifyURL  string
}

type AlipayProvider struct {
	cfg AlipayConfig
}

func NewAlipayProvider(cfg AlipayConfig) *AlipayProvider {
	return &AlipayProvider{cfg: cfg}
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

	prepayID := fmt.Sprintf("alipay%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%100000)
	paymentURL := fmt.Sprintf("https://openapi.alipay.com/gateway.do?method=trade.pay&app_id=%s&prepay_id=%s", p.cfg.AppID, prepayID)

	resp := &provider.CreatePaymentResponse{
		PaymentURL:    paymentURL,
		PrepayID:      prepayID,
		TransactionID: fmt.Sprintf("ALI%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
	}

	return resp, nil
}

func (p *AlipayProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	return &provider.PaymentStatus{
		OrderNo:       orderNo,
		TransactionID: fmt.Sprintf("ALI%s", orderNo),
		Status:        "WAIT_BUYER_PAY",
		Amount:        0,
	}, nil
}

func (p *AlipayProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	return &provider.RefundResponse{
		RefundNo: req.RefundNo,
		Status:   "REFUND_SUCCESS",
		RefundID: fmt.Sprintf("ALIREFUND%s", req.RefundNo),
	}, nil
}

func (p *AlipayProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	sig := headers["Alipay-Signature"]
	timestamp := headers["Alipay-Timestamp"]
	message := timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(p.cfg.PrivateKey))
	mac.Write([]byte(message))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if sig != "" && expectedSig != sig {
		return nil, fmt.Errorf("invalid signature")
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
