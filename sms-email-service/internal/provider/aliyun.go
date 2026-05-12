package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type aliyunProvider struct {
	accessKeyID     string
	accessKeySecret string
	signName        string
	templateCode    string
}

func NewAliyunProvider(accessKeyID, accessKeySecret, signName string) SMSProvider {
	return &aliyunProvider{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		signName:        signName,
		templateCode:    "100001",
	}
}

func (p *aliyunProvider) Name() string { return "aliyun" }

func (p *aliyunProvider) SendCode(ctx context.Context, phoneNumber string) (string, error) {
	if p.accessKeyID == "" || p.accessKeySecret == "" {
		return "", fmt.Errorf("aliyun: access key not configured")
	}

	templateParam, _ := json.Marshal(map[string]string{
		"code": "##code##",
		"min":  "5",
	})

	params := map[string]string{
		"AccessKeyId":      p.accessKeyID,
		"Action":           "SendSmsVerifyCode",
		"CodeLength":       "6",
		"CodeType":         "1",
		"DuplicatePolicy":  "1",
		"Format":           "JSON",
		"Interval":         "60",
		"PhoneNumber":      phoneNumber,
		"ReturnVerifyCode": "true",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"SignatureVersion": "1.0",
		"SignName":         p.signName,
		"TemplateCode":     p.templateCode,
		"TemplateParam":    string(templateParam),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"ValidTime":        "300",
		"Version":          "2017-05-25",
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, specialEncode(k)+"="+specialEncode(params[k]))
	}
	canonicalQuery := strings.Join(pairs, "&")

	stringToSign := "GET&" + specialEncode("/") + "&" + specialEncode(canonicalQuery)

	mac := hmac.New(sha1.New, []byte(p.accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	reqURL := "https://dypnsapi.aliyuncs.com/?" + canonicalQuery + "&Signature=" + specialEncode(signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("aliyun: build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("aliyun: send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		Model   struct {
			VerifyCode string `json:"VerifyCode"`
			BizId      string `json:"BizId"`
		} `json:"Model"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("aliyun: parse response: %w (body: %s)", err, string(body))
	}

	if result.Code != "OK" {
		return "", fmt.Errorf("aliyun: %s: %s", result.Code, result.Message)
	}

	if result.Model.VerifyCode == "" {
		return "", fmt.Errorf("aliyun: verify code not returned")
	}

	return result.Model.VerifyCode, nil
}

func specialEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
