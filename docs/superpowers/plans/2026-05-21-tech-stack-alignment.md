# Technology Stack Alignment — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align all dependencies, Dockerfiles, and SDK integrations to the prescribed technology versions defined in SSD V2.0.0 Section 2.4.

**Architecture:** Systematic upgrade across 9 Go microservices, 11 shared packages, Dockerfiles, and docker-compose. Bottom-up approach: infrastructure versions first, then Go toolchain, then framework, then libraries, then SDKs. Each task is independently verifiable with `go build` + `go test`.

**Tech Stack Versions (target):**

| Component | Target Version |
|-----------|---------------|
| Go | 1.26 (1.26.3) |
| Gin | v1.12.0 |
| Redis Server | 8.2-alpine (LTS) |
| go-redis | v9.19.0 |
| VictoriaMetrics | v1.143.0 |
| Loki | 3.7.2 |
| Grafana | 13.0.1 |
| SendGrid | v3.16.1 |
| AWS SES | aws-sdk-go-v2 (latest) |
| WeChat Pay | v0.2.21 |
| Alipay | v3.2.29 |
| Alpine (runtime) | 3.23 |
| OTel | v1.43.0 |

**GOPROXY:** All `go mod tidy` / `go get` commands require `$env:GOPROXY="https://goproxy.cn,direct"` and `$env:CGO_ENABLED="1"`.

**Priority ordering rationale:**
1. Infrastructure images (docker-compose) — foundation, no code changes
2. Go toolchain (Dockerfiles + go.mod) — must be consistent before library upgrades
3. Saga Redis client migration — v8→v9, removes incompatible library split
4. Gin framework — affects all services, must be upgraded together
5. Monitoring version pinning — simple image tag changes
6. Payment SDK — new dependency + code refactor
7. Email SDK — SendGrid + AWS SES real integration
8. Tracing — OTLP exporter + Jaeger integration
9. Final validation — full build + test

---

## Task 1: Upgrade Infrastructure Images in docker-compose.yml

**Files:**
- Modify: `docker-compose.yml`

**Target Versions:**
- `postgres:18-alpine` → keep (already correct)
- `redis:7-alpine` → `redis:8.2-alpine`
- `victoriametrics/victoria-metrics:latest` → `victoriametrics/victoria-metrics:v1.143.0`
- `grafana/loki:latest` → `grafana/loki:3.7.2`
- `grafana/grafana:latest` → `grafana/grafana:13.0.1`

- [ ] **Step 1: Update Redis image**

In `docker-compose.yml`, change:
```yaml
image: redis:7-alpine
```
to:
```yaml
image: redis:8.2-alpine
```

Apply to BOTH redis service definitions (main redis at line ~34 and sentinel at `infra/redis/docker-compose-sentinel.yml`).

- [ ] **Step 2: Pin VictoriaMetrics version**

Change:
```yaml
image: victoriametrics/victoria-metrics:latest
```
to:
```yaml
image: victoriametrics/victoria-metrics:v1.143.0
```

- [ ] **Step 3: Pin Loki version**

Change:
```yaml
image: grafana/loki:latest
```
to:
```yaml
image: grafana/loki:3.7.2
```

- [ ] **Step 4: Pin Grafana version**

Change:
```yaml
image: grafana/grafana:latest
```
to:
```yaml
image: grafana/grafana:13.0.1
```

- [ ] **Step 5: Update infra/redis/docker-compose-sentinel.yml**

Change `redis:7-alpine` to `redis:8.2-alpine` in all Redis/Sentinel service definitions.

- [ ] **Step 6: Restart affected containers**

```powershell
docker compose up -d redis victoriametrics loki grafana
```

Verify with `docker ps` that all containers start and become healthy.

- [ ] **Step 7: Commit**

```bash
git add docker-compose.yml infra/redis/docker-compose-sentinel.yml
git commit -m "chore: pin infrastructure images to prescribed versions (Redis 8.2 LTS, VM v1.143.0, Loki 3.7.2, Grafana 13.0.1)"
```

---

## Task 2: Unify Dockerfile Go Versions to 1.26

**Files:**
- Modify: `api-gateway/Dockerfile`
- Modify: `account-service/Dockerfile`
- Modify: `auth-service/Dockerfile`
- Modify: `notification-service/Dockerfile`
- Modify: `data-product-service/Dockerfile`
- Modify: `credit-service/Dockerfile`
- Modify: `compliance-service/Dockerfile`
- Modify: `config-service/Dockerfile`
- Modify: `payment-service/Dockerfile`
- Modify: `db-migrations/Dockerfile` (if uses golang image)
- Modify: `go.work` — update `go 1.24.0` to `go 1.26`
- Modify: ALL `go.mod` files — update `go 1.24` to `go 1.26`

- [ ] **Step 1: Update all Dockerfiles build stage**

In each service Dockerfile, change the `FROM golang:...` line to:
```dockerfile
FROM golang:1.26-alpine AS builder
```

And the runtime stage to:
```dockerfile
FROM alpine:3.23
```

9 Dockerfiles to update: api-gateway, account-service, auth-service, notification-service, data-product-service, credit-service, compliance-service, config-service, payment-service.

- [ ] **Step 2: Update go.work**

Change `go 1.24.0` to `go 1.26`.

- [ ] **Step 3: Update all go.mod files**

In every go.mod file (22 files), change `go 1.24` to `go 1.26`. Files:
- 9 service go.mod files
- 11 pkg/*/go.mod files
- tests/integration/go.mod
- tests/e2e/go.mod

Use the edit tool with `replaceAll` on each file.

- [ ] **Step 4: Run go mod tidy on all modules**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
$modules = @(
  "pkg/config","pkg/circuitbreaker","pkg/health","pkg/logging","pkg/trace",
  "pkg/saga","pkg/async","pkg/vault","pkg/discovery","pkg/database","pkg/server",
  "account-service","api-gateway","auth-service","compliance-service",
  "config-service","credit-service","data-product-service","notification-service",
  "payment-service","tests/integration","tests/e2e"
)
foreach ($m in $modules) { go mod tidy; Write-Output "Tidied $m" }
```

NOTE: Run `go mod tidy` from within each module directory using `workdir` parameter.

- [ ] **Step 5: Build and test**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go build ./account-service/... ./api-gateway/... ./auth-service/... ./compliance-service/... ./config-service/... ./credit-service/... ./data-product-service/... ./notification-service/... ./payment-service/...
go test ./pkg/... ./account-service/... ./api-gateway/... ./auth-service/... ./compliance-service/... ./config-service/... ./credit-service/... ./data-product-service/... ./notification-service/... ./payment-service/...
```

Expected: zero failures.

- [ ] **Step 6: Commit**

```bash
git add '**/Dockerfile' '**/go.mod' '**/go.sum' go.work
git commit -m "chore: unify Go version to 1.26 across all Dockerfiles and go.mod files"
```

---

## Task 3: Upgrade Gin to v1.12.0

**Files:**
- Modify: ALL `go.mod` files that contain `github.com/gin-gonic/gin`
- This includes 9 services + pkg/logging + pkg/trace

- [ ] **Step 1: Update Gin in all go.mod files**

Change all occurrences of `github.com/gin-gonic/gin v1.9.1` and `github.com/gin-gonic/gin v1.10.0` to `github.com/gin-gonic/gin v1.12.0`.

Files (11 total):
- api-gateway/go.mod
- account-service/go.mod
- auth-service/go.mod
- notification-service/go.mod
- credit-service/go.mod
- compliance-service/go.mod
- config-service/go.mod
- data-product-service/go.mod
- payment-service/go.mod
- pkg/logging/go.mod
- pkg/trace/go.mod

- [ ] **Step 2: Run go mod tidy on all affected modules**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
# Run go mod tidy in each of the 11 directories
```

- [ ] **Step 3: Check for Gin API breaking changes**

Gin v1.12.0 breaking: minimum Go 1.24 (already handled in Task 2). Check for any deprecated API usage. Read CHANGELOG at https://github.com/gin-gonic/gin/releases. Key changes:
- `gin.Mode()` behavior unchanged
- `c.JSON()` unchanged
- Middleware chain unchanged
- No code changes expected for our usage patterns

- [ ] **Step 4: Build and test all services**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go build ./account-service/... ./api-gateway/... ./auth-service/... ./compliance-service/... ./config-service/... ./credit-service/... ./data-product-service/... ./notification-service/... ./payment-service/...
go test ./pkg/... ./account-service/... ./api-gateway/... ./auth-service/... ./compliance-service/... ./config-service/... ./credit-service/... ./data-product-service/... ./notification-service/... ./payment-service/...
```

Expected: zero failures. If any test breaks due to Gin API changes, fix inline.

- [ ] **Step 5: Commit**

```bash
git add '**/go.mod' '**/go.sum'
git commit -m "chore: upgrade Gin to v1.12.0 across all services and shared packages"
```

---

## Task 4: Migrate pkg/saga from go-redis/redis/v8 to go-redis/v9

**Files:**
- Modify: `pkg/saga/go.mod`
- Modify: `pkg/saga/store.go`
- Modify: `pkg/saga/orchestrator.go` (if imports old client)
- Modify: `pkg/saga/saga_test.go` (if exists)

- [ ] **Step 1: Read current saga package files**

Read all files in `pkg/saga/` to understand the v8 API usage patterns.

- [ ] **Step 2: Update go.mod**

In `pkg/saga/go.mod`:
- Remove `github.com/go-redis/redis/v8 v8.11.5`
- Add `github.com/redis/go-redis/v9 v9.19.0`

- [ ] **Step 3: Update import paths**

In all `.go` files under `pkg/saga/`:
- Change `github.com/go-redis/redis/v8` to `github.com/redis/go-redis/v9`

The v8→v9 API is nearly identical. Key differences:
- `redis.NewClient()` → unchanged
- `client.Get/Set/Del()` → unchanged
- `redis.Nil` → unchanged (still `redis.Nil`)
- Context handling → unchanged (v8 already requires context)

- [ ] **Step 4: Run go mod tidy**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go mod tidy
# from pkg/saga directory
```

- [ ] **Step 5: Build and test**

```powershell
go build ./pkg/saga/...
go test ./pkg/saga/...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/saga/
git commit -m "chore: migrate pkg/saga from go-redis/v8 to go-redis/v9"
```

---

## Task 5: Integrate Official WeChat Pay SDK

**Files:**
- Modify: `payment-service/go.mod`
- Modify: `payment-service/internal/service/wechat_pay.go`
- Modify: `payment-service/internal/service/wechat_pay_test.go` (if exists)

**Target SDK:** `github.com/wechatpay-apiv3/wechatpay-go v0.2.21` (official WeChat Pay Go SDK)

- [ ] **Step 1: Read current wechat_pay.go**

Read `payment-service/internal/service/wechat_pay.go` to understand current mock implementation.

- [ ] **Step 2: Add SDK dependency**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go get github.com/wechatpay-apiv3/wechatpay-go@v0.2.21
# from payment-service directory
```

- [ ] **Step 3: Refactor wechat_pay.go to use official SDK**

Replace the mock implementation with real SDK integration:

```go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/wechatpay-apiv3/wechatpay-go/core"
    "github.com/wechatpay-apiv3/wechatpay-go/core/option"
    "github.com/wechatpay-apiv3/wechatpay-go/services/payments"
    "github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WeChatPayProvider struct {
    client   *payments.NativeApiService
    clientV3 *payments.TransactionApiService
    appID    string
    mchID    string
}

func NewWeChatPayProvider(appID, mchID, apiKey string, privateKey []byte, certificate []byte) (*WeChatPayProvider, error) {
    mchPrivateKey, err := utils.LoadPrivateKey(string(privateKey))
    if err != nil {
        return nil, fmt.Errorf("load private key: %w", err)
    }
    mchCertificate, err := utils.LoadCertificate(string(certificate))
    if err != nil {
        return nil, fmt.Errorf("load certificate: %w", err)
    }
    ctx := context.Background()
    opts := []core.ClientOption{
        option.WithWechatPayAuthCipher([]byte(apiKey), *mchPrivateKey, *mchCertificate, nil),
        option.WithCipher([]byte(apiKey)),
    }
    client, err := core.NewClient(ctx, opts...)
    if err != nil {
        return nil, fmt.Errorf("new wechat pay client: %w", err)
    }
    return &WeChatPayProvider{
        client:   &payments.NativeApiService{Client: client},
        clientV3: &payments.TransactionApiService{Client: client},
        appID:    appID,
        mchID:    mchID,
    }, nil
}

func (p *WeChatPayProvider) CreatePayment(ctx context.Context, orderNo string, amount float64, description string) (string, error) {
    amountCents := int64(amount * 100)
    resp, _, err := p.client.Prepay(ctx, payments.NativePrepayRequest{
        Appid:       core.String(p.appID),
        Mchid:       core.String(p.mchID),
        Description: core.String(description),
        OutTradeNo:  core.String(orderNo),
        TimeExpire:  core.Time(time.Now().Add(30 * time.Minute)),
        Amount:      &payments.Amount{Total: core.Int64(amountCents), Currency: core.String("CNY")},
    })
    if err != nil {
        return "", fmt.Errorf("wechat pay prepay: %w", err)
    }
    return *resp.CodeUrl, nil
}

func (p *WeChatPayProvider) QueryPayment(ctx context.Context, orderNo string) (string, error) {
    resp, _, err := p.clientV3.QueryOrderOutTradeNo(ctx,
        payments.QueryOrderOutTradeNoRequest{
            OutTradeNo: core.String(orderNo),
            Mchid:      core.String(p.mchID),
        },
    )
    if err != nil {
        return "", fmt.Errorf("wechat pay query: %w", err)
    }
    if resp.TradeState != nil {
        return string(*resp.TradeState), nil
    }
    return "", nil
}

func (p *WeChatPayProvider) Refund(ctx context.Context, orderNo string, refundNo string, totalAmount float64, refundAmount float64) error {
    _, _, err := p.clientV3.Refund(ctx, payments.CreateRefundRequest{
        OutTradeNo:  core.String(orderNo),
        OutRefundNo: core.String(refundNo),
        Amount: &payments.AmountReq{
            Total:    core.Int64(int64(totalAmount * 100)),
            Refund:   core.Int64(int64(refundAmount * 100)),
            Currency: core.String("CNY"),
        },
    })
    if err != nil {
        return fmt.Errorf("wechat pay refund: %w", err)
    }
    return nil
}

func (p *WeChatPayProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (map[string]interface{}, error) {
    handler := core.NewNotifyHandler(p.mchID, nil)
    result, err := handler.ParseNotifyRequest(ctx, headers, body)
    if err != nil {
        return nil, fmt.Errorf("verify callback: %w", err)
    }
    return result, nil
}
```

- [ ] **Step 4: Update main.go wiring**

In `payment-service/cmd/main.go`, update the WeChat Pay provider initialization to use `NewWeChatPayProvider()` with real config values from environment variables:
- `WECHAT_APP_ID`
- `WECHAT_MCH_ID`
- `WECHAT_API_KEY`
- `WECHAT_PRIVATE_KEY_PATH`
- `WECHAT_CERTIFICATE_PATH`

Keep sandbox fallback for development.

- [ ] **Step 5: Run go mod tidy and build**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go mod tidy
go build ./payment-service/...
go test ./payment-service/...
```

- [ ] **Step 6: Commit**

```bash
git add payment-service/
git commit -m "feat: integrate official WeChat Pay SDK (wechatpay-go v0.2.21)"
```

---

## Task 6: Integrate Official Alipay SDK

**Files:**
- Modify: `payment-service/go.mod`
- Modify: `payment-service/internal/service/alipay.go`
- Modify: `payment-service/internal/service/alipay_test.go` (if exists)

**Target SDK:** `github.com/smartwalle/alipay v3.2.29` (most widely used Go Alipay SDK)

- [ ] **Step 1: Read current alipay.go**

Read `payment-service/internal/service/alipay.go` to understand current mock.

- [ ] **Step 2: Add SDK dependency**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go get github.com/smartwalle/alipay/v3@v3.2.29
# from payment-service directory
```

- [ ] **Step 3: Refactor alipay.go to use real SDK**

```go
package service

import (
    "context"
    "fmt"

    "github.com/smartwalle/alipay/v3"
)

type AlipayProvider struct {
    client *alipay.Client
    appID  string
}

func NewAlipayProvider(appID, privateKey, publicKey, notifyURL string, isSandbox bool) (*AlipayProvider, error) {
    var client *alipay.Client
    var err error
    if isSandbox {
        client, err = alipay.NewClient(appID, privateKey, true)
    } else {
        client, err = alipay.NewClient(appID, privateKey, false)
    }
    if err != nil {
        return nil, fmt.Errorf("new alipay client: %w", err)
    }
    if err := client.LoadAliPayPublicKey(publicKey); err != nil {
        return nil, fmt.Errorf("load alipay public key: %w", err)
    }
    client.SetNotifyUrl(notifyURL)
    return &AlipayProvider{client: client, appID: appID}, nil
}

func (p *AlipayProvider) CreatePayment(ctx context.Context, orderNo string, amount float64, subject string) (string, error) {
    var pay = alipay.TradeWapPay{}
    pay.OutTradeNo = orderNo
    pay.TotalAmount = fmt.Sprintf("%.2f", amount)
    pay.Subject = subject
    pay.ProductCode = "QUICK_WAP_WAY"

    url, err := p.client.TradeWapPay(ctx, pay)
    if err != nil {
        return "", fmt.Errorf("alipay trade wap pay: %w", err)
    }
    return url.String(), nil
}

func (p *AlipayProvider) QueryPayment(ctx context.Context, orderNo string) (string, error) {
    var query = alipay.TradeQuery{}
    query.OutTradeNo = orderNo
    result, err := p.client.TradeQuery(ctx, query)
    if err != nil {
        return "", fmt.Errorf("alipay trade query: %w", err)
    }
    if result.TradeStatus != "" {
        return result.TradeStatus, nil
    }
    return "", nil
}

func (p *AlipayProvider) Refund(ctx context.Context, orderNo string, refundNo string, refundAmount float64) error {
    var refund = alipay.TradeRefund{}
    refund.OutTradeNo = orderNo
    refund.OutRequestNo = refundNo
    refund.RefundAmount = fmt.Sprintf("%.2f", refundAmount)
    _, err := p.client.TradeRefund(ctx, refund)
    if err != nil {
        return fmt.Errorf("alipay trade refund: %w", err)
    }
    return nil
}

func (p *AlipayProvider) VerifyCallback(ctx context.Context, params map[string]string) (map[string]interface{}, error) {
    ok, err := p.client.VerifySign(params)
    if err != nil {
        return nil, fmt.Errorf("verify callback sign: %w", err)
    }
    if !ok {
        return nil, fmt.Errorf("callback sign verification failed")
    }
    result := make(map[string]interface{})
    for k, v := range params {
        result[k] = v
    }
    return result, nil
}
```

- [ ] **Step 4: Update main.go wiring**

In `payment-service/cmd/main.go`, update Alipay provider initialization to use `NewAlipayProvider()` with real environment variables.

- [ ] **Step 5: Run go mod tidy and build**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go mod tidy
go build ./payment-service/...
go test ./payment-service/...
```

- [ ] **Step 6: Commit**

```bash
git add payment-service/
git commit -m "feat: integrate official Alipay SDK (smartwalle/alipay v3.2.29)"
```

---

## Task 7: Integrate SendGrid SDK

**Files:**
- Modify: `notification-service/go.mod`
- Modify: `notification-service/internal/provider/sendgrid.go`
- Modify: `notification-service/internal/provider/email.go` (factory, if references SendGrid)

**Target SDK:** `github.com/sendgrid/sendgrid-go v3.16.1` (official SendGrid Go SDK)

- [ ] **Step 1: Read current sendgrid.go**

Read `notification-service/internal/provider/sendgrid.go` to understand current stub.

- [ ] **Step 2: Add SDK dependency**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go get github.com/sendgrid/sendgrid-go@v3.16.1
# from notification-service directory
```

- [ ] **Step 3: Refactor sendgrid.go**

```go
package provider

import (
    "context"
    "fmt"

    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridProvider struct {
    apiKey string
    from   string
}

func NewSendGridProvider(apiKey, from string) *SendGridProvider {
    return &SendGridProvider{apiKey: apiKey, from: from}
}

func (p *SendGridProvider) SendEmail(ctx context.Context, to, subject, body string) error {
    from := mail.NewEmail("", p.from)
    toAddr := mail.NewEmail("", to)
    message := mail.NewSingleEmail(from, subject, toAddr, body, body)
    client := sendgrid.NewSendClient(p.apiKey)
    response, err := client.Send(message)
    if err != nil {
        return fmt.Errorf("sendgrid send: %w", err)
    }
    if response.StatusCode >= 400 {
        return fmt.Errorf("sendgrid error: status=%d body=%s", response.StatusCode, response.Body)
    }
    return nil
}
```

- [ ] **Step 4: Update email factory in provider**

In the email provider factory (`email.go` or `cmd/main.go`), update the SendGrid case to use `NewSendGridProvider(apiKey, from)` instead of the stub.

- [ ] **Step 5: Run go mod tidy and build**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go mod tidy
go build ./notification-service/...
go test ./notification-service/...
```

- [ ] **Step 6: Commit**

```bash
git add notification-service/
git commit -m "feat: integrate official SendGrid SDK (sendgrid-go v3.16.1)"
```

---

## Task 8: Integrate AWS SDK v2 for SES

**Files:**
- Modify: `notification-service/go.mod`
- Modify: `notification-service/internal/provider/ses.go`

**Target SDK:** `github.com/aws/aws-sdk-go-v2` (latest stable)

- [ ] **Step 1: Read current ses.go**

Read `notification-service/internal/provider/ses.go` to understand current custom HTTP implementation.

- [ ] **Step 2: Add AWS SDK dependencies**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/ses
go get github.com/aws/aws-sdk-go-v2/credentials
# from notification-service directory
```

- [ ] **Step 3: Refactor ses.go**

```go
package provider

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/ses"
    "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESProvider struct {
    client *ses.Client
    from   string
}

func NewSESProvider(accessKey, secretKey, region, from string) (*SESProvider, error) {
    cfg, err := config.LoadDefaultConfig(context.Background(),
        config.WithRegion(region),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
    )
    if err != nil {
        return nil, fmt.Errorf("load aws config: %w", err)
    }
    return &SESProvider{
        client: ses.NewFromConfig(cfg),
        from:   from,
    }, nil
}

func (p *SESProvider) SendEmail(ctx context.Context, to, subject, body string) error {
    _, err := p.client.SendEmail(ctx, &ses.SendEmailInput{
        Source: aws.String(p.from),
        Destination: &types.Destination{
            ToAddresses: []string{to},
        },
        Message: &types.Message{
            Subject: &types.Content{Data: aws.String(subject)},
            Body: &types.Body{
                Text: &types.Content{Data: aws.String(body)},
            },
        },
    })
    if err != nil {
        return fmt.Errorf("ses send email: %w", err)
    }
    return nil
}
```

- [ ] **Step 4: Update email factory**

Update the `aws_ses` case to use `NewSESProvider()` with environment variables:
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `AWS_REGION`
- `FROM_ADDRESS`

- [ ] **Step 5: Run go mod tidy and build**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go mod tidy
go build ./notification-service/...
go test ./notification-service/...
```

- [ ] **Step 6: Commit**

```bash
git add notification-service/
git commit -m "feat: integrate official AWS SDK v2 for SES email"
```

---

## Task 9: Upgrade Tracing to OTLP Exporter

**Files:**
- Modify: `pkg/trace/go.mod`
- Modify: `pkg/trace/sdk.go`
- Add: `docker-compose.yml` — add Jaeger service (optional, or use Tempo)

**Target:** `go.opentelemetry.io/otel v1.43.0` with OTLP gRPC exporter

- [ ] **Step 1: Read current pkg/trace/sdk.go**

Read `pkg/trace/sdk.go` to understand current stdouttrace implementation.

- [ ] **Step 2: Add OTLP exporter dependencies**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go get go.opentelemetry.io/otel@v1.43.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.43.0
go get go.opentelemetry.io/otel/sdk@v1.43.0
# from pkg/trace directory
```

- [ ] **Step 3: Refactor initOTLPProvider**

Replace stdouttrace with OTLP gRPC exporter:

```go
func initOTLPProvider(ctx context.Context, endpoint string) (func(context.Context) error, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("create OTLP exporter: %w", err)
    }
    provider := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("account-center"),
        )),
    )
    otel.SetTracerProvider(provider)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return provider.Shutdown, nil
}
```

- [ ] **Step 4: Add Jaeger all-in-one to docker-compose.yml** (optional, for local dev)

```yaml
  jaeger:
    image: jaegertracing/all-in-one:1.65
    container_name: jaeger
    environment:
      COLLECTOR_OTLP_ENABLED: true
    ports:
      - "16686:16686"
      - "4317:4317"
    networks:
      - app_network
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 256M
    restart: always
```

- [ ] **Step 5: Update .env.example**

Add:
```
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
```

- [ ] **Step 6: Run go mod tidy and test**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go mod tidy
go build ./pkg/trace/...
go test ./pkg/trace/...
```

- [ ] **Step 7: Commit**

```bash
git add pkg/trace/ docker-compose.yml .env.example
git commit -m "feat: upgrade tracing to OTLP gRPC exporter with Jaeger backend"
```

---

## Task 10: Full Validation

- [ ] **Step 1: Build all services**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go build ./account-service/... ./api-gateway/... ./auth-service/... ./compliance-service/... ./config-service/... ./credit-service/... ./data-product-service/... ./notification-service/... ./payment-service/...
```

Expected: zero errors.

- [ ] **Step 2: Run all tests**

```powershell
$env:GOPROXY="https://goproxy.cn,direct"; $env:CGO_ENABLED="1"
go test ./pkg/... ./account-service/... ./api-gateway/... ./auth-service/... ./compliance-service/... ./config-service/... ./credit-service/... ./data-product-service/... ./notification-service/... ./payment-service/...
```

Expected: zero failures.

- [ ] **Step 3: Verify Docker builds**

```powershell
docker compose build
```

Expected: all 9 services + db-migrate build successfully.

- [ ] **Step 4: Verify dependency consistency**

Check that all go.mod files have:
- `go 1.26` (not 1.24 or 1.24.0)
- `github.com/gin-gonic/gin v1.12.0`
- `github.com/redis/go-redis/v9 v9.19.0` (not v8)
- `github.com/lib/pq v1.12.3`
- No `github.com/go-redis/redis/v8` anywhere

- [ ] **Step 5: Verify Dockerfile consistency**

All 9 Dockerfiles should have:
- `FROM golang:1.26-alpine AS builder`
- `FROM alpine:3.23`

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "chore: technology stack alignment complete — Go 1.26, Gin v1.12.0, Redis 8.6, official SDKs"
```
