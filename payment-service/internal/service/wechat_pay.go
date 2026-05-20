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

type WeChatPayConfig struct {
	AppID    string
	MchID    string
	APIKey   string
	CertPath string
}

type WeChatPayProvider struct {
	cfg WeChatPayConfig
}

func NewWeChatPayProvider(cfg WeChatPayConfig) *WeChatPayProvider {
	return &WeChatPayProvider{cfg: cfg}
}

func (p *WeChatPayProvider) Name() string {
	return "wechat"
}

func (p *WeChatPayProvider) CreatePayment(ctx context.Context, req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	switch req.TradeType {
	case "wechat_h5", "wechat_mini", "wechat_native":
	default:
		return nil, fmt.Errorf("unsupported wechat trade type: %s", req.TradeType)
	}

	prepayID := fmt.Sprintf("wx%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%100000)
	paymentURL := fmt.Sprintf("https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=%s", prepayID)

	resp := &provider.CreatePaymentResponse{
		PaymentURL:    paymentURL,
		PrepayID:      prepayID,
		TransactionID: fmt.Sprintf("WX%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000),
	}

	if req.TradeType == "wechat_native" {
		resp.QRCodeURL = fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", prepayID)
	}

	return resp, nil
}

func (p *WeChatPayProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	return &provider.PaymentStatus{
		OrderNo:       orderNo,
		TransactionID: fmt.Sprintf("WX%s", orderNo),
		Status:        "NOTPAY",
		Amount:        0,
	}, nil
}

func (p *WeChatPayProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	return &provider.RefundResponse{
		RefundNo: req.RefundNo,
		Status:   "SUCCESS",
		RefundID: fmt.Sprintf("WXREFUND%s", req.RefundNo),
	}, nil
}

func (p *WeChatPayProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	sig := headers["Wechatpay-Signature"]
	timestamp := headers["Wechatpay-Timestamp"]
	nonce := headers["Wechatpay-Nonce"]
	message := fmt.Sprintf("%s\n%s\n", timestamp, nonce) + string(body) + "\n"
	mac := hmac.New(sha256.New, []byte(p.cfg.APIKey))
	mac.Write([]byte(message))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if expectedSig != sig {
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
	if txnID, ok := payload["transaction_id"].(string); ok {
		result.TransactionID = txnID
	}
	if resultCode, ok := payload["result_code"].(string); ok {
		result.Status = resultCode
	} else {
		result.Status = "SUCCESS"
	}
	if totalStr, ok := payload["total_amount"].(string); ok {
		fmt.Sscanf(totalStr, "%f", &result.Amount)
	} else if total, ok := payload["total_amount"].(float64); ok {
		result.Amount = total
	}
	if paidAt, ok := payload["time_paid"].(string); ok {
		result.PaidAt = paidAt
	}

	return result, nil
}
