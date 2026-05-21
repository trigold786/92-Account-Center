package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
)

type WeChatPayConfig struct {
	AppID                   string
	MchID                   string
	APIKey                  string
	CertificateSerialNo     string
	PrivateKeyPath          string
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

	result := &provider.CreatePaymentResponse{
		PaymentURL:    *resp.CodeUrl,
		TransactionID: req.OrderNo,
	}
	return result, nil
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
