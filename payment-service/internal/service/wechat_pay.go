package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
)

type WeChatPayConfig struct {
	Mode                string
	AppID               string
	MchID               string
	APIKey              string
	CertificateSerialNo string
	PrivateKeyPath      string
}

func (c WeChatPayConfig) ValidateProduction() error {
	if c.Mode != "production" {
		return nil
	}
	fields := map[string]string{
		"WECHAT_APP_ID":           c.AppID,
		"WECHAT_MCH_ID":           c.MchID,
		"WECHAT_API_KEY":          c.APIKey,
		"WECHAT_CERT_SERIAL_NO":   c.CertificateSerialNo,
		"WECHAT_PRIVATE_KEY_PATH": c.PrivateKeyPath,
	}
	for name, value := range fields {
		if insecurePaymentValue(value) {
			return fmt.Errorf("production wechat payment config requires real %s", name)
		}
	}
	return nil
}

func insecurePaymentValue(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return true
	}
	return strings.Contains(v, "sandbox") || strings.Contains(v, "test") || strings.Contains(v, "default") || strings.Contains(v, "placeholder")
}

type WeChatPayProvider struct {
	cfg       WeChatPayConfig
	client    *core.Client
	nativeSvc *native.NativeApiService
	refundSvc *refunddomestic.RefundsApiService
	sandbox   bool
}

func NewWeChatPayProvider(cfg WeChatPayConfig) *WeChatPayProvider {
	p := &WeChatPayProvider{cfg: cfg, sandbox: true}

	if cfg.PrivateKeyPath != "" && cfg.APIKey != "" {
		privKeyData, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return p
		}

		mchPrivateKey, err := utils.LoadPrivateKey(string(privKeyData))
		if err != nil {
			return p
		}

		ctx := context.Background()
		opts := []core.ClientOption{
			option.WithWechatPayAutoAuthCipher(
				cfg.MchID,
				cfg.CertificateSerialNo,
				mchPrivateKey,
				cfg.APIKey,
			),
		}
		client, err := core.NewClient(ctx, opts...)
		if err != nil {
			return p
		}

		p.client = client
		p.nativeSvc = &native.NativeApiService{Client: client}
		p.refundSvc = &refunddomestic.RefundsApiService{Client: client}
		p.sandbox = false
	}

	return p
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

	if p.sandbox {
		return p.sandboxCreatePayment(req)
	}

	amountCents := int64(req.Amount * 100)
	resp, _, err := p.nativeSvc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(p.cfg.AppID),
		Mchid:       core.String(p.cfg.MchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OrderNo),
		TimeExpire:  core.Time(time.Now().Add(30 * time.Minute)),
		Amount:      &native.Amount{Total: core.Int64(amountCents), Currency: core.String("CNY")},
	})
	if err != nil {
		return nil, fmt.Errorf("wechat pay prepay: %w", err)
	}

	return &provider.CreatePaymentResponse{
		PaymentURL:    *resp.CodeUrl,
		TransactionID: req.OrderNo,
	}, nil
}

func (p *WeChatPayProvider) sandboxCreatePayment(req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
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
	if p.sandbox {
		return &provider.PaymentStatus{
			OrderNo:       orderNo,
			TransactionID: fmt.Sprintf("WX%s", orderNo),
			Status:        "NOTPAY",
			Amount:        0,
		}, nil
	}

	resp, _, err := p.nativeSvc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(orderNo),
		Mchid:      core.String(p.cfg.MchID),
	})
	if err != nil {
		return nil, fmt.Errorf("wechat pay query: %w", err)
	}

	status := "NOTPAY"
	if resp.TradeState != nil {
		status = string(*resp.TradeState)
	}

	return &provider.PaymentStatus{
		OrderNo:       orderNo,
		TransactionID: "",
		Status:        status,
		Amount:        0,
	}, nil
}

func (p *WeChatPayProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	if p.sandbox {
		return &provider.RefundResponse{
			RefundNo: req.RefundNo,
			Status:   "SUCCESS",
			RefundID: fmt.Sprintf("WXREFUND%s", req.RefundNo),
		}, nil
	}

	_, _, err := p.refundSvc.Create(ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(req.OrderNo),
		OutRefundNo: core.String(req.RefundNo),
		Amount: &refunddomestic.AmountReq{
			Total:    core.Int64(int64(req.TotalAmount * 100)),
			Refund:   core.Int64(int64(req.RefundAmount * 100)),
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("wechat pay refund: %w", err)
	}

	return &provider.RefundResponse{
		RefundNo: req.RefundNo,
		Status:   "SUCCESS",
		RefundID: fmt.Sprintf("WXREFUND%s", req.RefundNo),
	}, nil
}

func (p *WeChatPayProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	if err := p.verifyCallbackSignature(headers, body); err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid callback body: %w", err)
	}

	result := &provider.CallbackResult{RawData: string(body)}
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

func (p *WeChatPayProvider) verifyCallbackSignature(headers map[string]string, body []byte) error {
	timestamp := firstHeader(headers, "Wechatpay-Timestamp", "wechatpay-timestamp")
	nonce := firstHeader(headers, "Wechatpay-Nonce", "wechatpay-nonce")
	signature := firstHeader(headers, "Wechatpay-Signature", "wechatpay-signature")
	if timestamp == "" || nonce == "" || signature == "" {
		return fmt.Errorf("missing wechat callback signature headers")
	}
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	mac := hmac.New(sha256.New, []byte(p.cfg.APIKey))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid wechat callback signature")
	}
	return nil
}

func firstHeader(headers map[string]string, names ...string) string {
	for _, name := range names {
		if value := headers[name]; value != "" {
			return value
		}
	}
	return ""
}
