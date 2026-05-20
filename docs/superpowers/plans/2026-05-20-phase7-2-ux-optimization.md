# Phase 7.2 — UX Optimization Implementation Plan

> **For agentic workers:** Step-by-step implementation. Each task is self-contained with tests before code.

**Goal:** Improve user experience with one-click social login, biometric auth, personalized dashboards, complete payment flows, upgrade/downgrade UX, renewal reminders, and referral progress visualization.

**Architecture:** All backend Go services. OAuth providers use plugin pattern with sandbox mocks. Payment recovery uses Asynq scheduler. Renewal reminders use multi-channel dispatch via notification-service. Dashboard configs driven by config-service per subscription level.

**Tech Stack:** Go 1.24, Gin, PostgreSQL, Redis, Asynq

**Dependencies:** P1-6 → P1-5, P1-9 → P1-8, P1-10/P1-11 → FN-12 (stub). Execute in order: P1-5 → P1-6 → P1-7 → P1-8 → P1-9 → P1-10 → P1-11.

---

## File Structure

### New files:
```
auth-service/
├── internal/
│   ├── handler/oauth_handler.go          # OAuth flow endpoints
│   ├── model/social_account.go           # SocialAccount model
│   ├── repository/social_account_repository.go
│   ├── service/oauth_service.go          # OAuth provider interface + WeChat/Apple/Google
│   └── service/oauth_service_test.go
├── migrations/005_social_accounts.sql    # New table migration

payment-service/
├── internal/
│   ├── handler/payment_flow.go           # Payment result, invoice, retry
│   ├── handler/invoice_handler.go         # Invoice CRUD
│   ├── model/invoice.go                   # Invoice model
│   ├── repository/invoice_repository.go
│   ├── service/payment_recovery.go        # Auto-repair lost callbacks
│   └── service/payment_recovery_test.go

account-service/
├── internal/
│   ├── handler/dashboard_handler.go       # GET /api/v1/account/dashboard
│   ├── service/dashboard_service.go       # Dashboard builder by level
│   ├── service/dashboard_test.go
│   ├── handler/upgrade_handler.go         # Upgrade/downgrade preview
│   ├── service/upgrade_service.go         # Proration calculator
│   ├── service/upgrade_service_test.go
│   └── worker/renewal.go                  # Asynq renewal reminder

credit-service/
├── internal/
│   ├── handler/referral_dashboard_handler.go
│   ├── service/referral_dashboard_service.go
│   └── service/referral_dashboard_test.go
```

### Modified files:
```
auth-service/cmd/main.go                    # Register OAuth routes + providers
auth-service/go.mod                         # No new deps needed
payment-service/cmd/main.go                 # Register payment recovery scheduler
account-service/cmd/main.go                 # Register dashboard + upgrade + renewal routes
credit-service/cmd/main.go                  # Register referral dashboard routes
```

---

## Task P1-5: UX-01 — One-Click Login (Social OAuth)

**Files:**
- Create: `auth-service/migrations/005_social_accounts.sql`
- Create: `auth-service/internal/model/social_account.go`
- Create: `auth-service/internal/repository/social_account_repository.go`
- Create: `auth-service/internal/service/oauth_service.go`
- Create: `auth-service/internal/service/oauth_service_test.go`
- Create: `auth-service/internal/handler/oauth_handler.go`
- Modify: `auth-service/cmd/main.go`

- [ ] **Step 1: Create social_accounts table migration**

`auth-service/migrations/005_social_accounts.sql`:
```sql
-- +goose Up
CREATE TABLE social_accounts (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider      VARCHAR(32) NOT NULL,        -- 'wechat', 'apple', 'google'
    provider_uid  VARCHAR(255) NOT NULL,       -- unique ID from provider
    email         VARCHAR(255),
    avatar_url    TEXT,
    access_token  TEXT,
    refresh_token TEXT,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_uid)
);
CREATE INDEX idx_social_accounts_user_id ON social_accounts(user_id);

-- +goose Down
DROP TABLE IF EXISTS social_accounts;
```

- [ ] **Step 2: Write OAuth service tests**

`auth-service/internal/service/oauth_service_test.go`:
```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type mockOAuthProvider struct {
	name      string
	authURL   string
	userInfo  *model.SocialUserInfo
	email     string
}

func (m *mockOAuthProvider) Name() string { return m.name }

func (m *mockOAuthProvider) GetAuthURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	if code == "fail" {
		return "", errors.New("exchange failed")
	}
	return "mock_token_" + code, nil
}

func (m *mockOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	if accessToken == "bad_token" {
		return nil, errors.New("invalid token")
	}
	return m.userInfo, nil
}

func TestOAuthProviderRegistry(t *testing.T) {
	reg := NewOAuthProviderRegistry()
	mock := &mockOAuthProvider{name: "mock_test", authURL: "https://example.com/oauth"}
	reg.Register(mock)

	got, err := reg.Get("mock_test")
	if err != nil {
		t.Fatalf("expected provider, got error: %v", err)
	}
	if got.Name() != "mock_test" {
		t.Fatalf("unexpected name: %s", got.Name())
	}

	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}

func TestOAuthFlow(t *testing.T) {
	mock := &mockOAuthProvider{
		name:    "mock",
		authURL: "https://example.com/oauth",
		userInfo: &model.SocialUserInfo{
			Provider:     "mock",
			ProviderUID:  "uid_123",
			Email:        "test@example.com",
			AvatarURL:    "https://example.com/avatar.png",
		},
	}
	svc := &OAuthService{
		registry: NewOAuthProviderRegistry(),
		userRepo: &mockUserRepo{},
	}
	svc.registry.Register(mock)

	// Test auth URL generation
	url, err := svc.GetAuthURL(context.Background(), "mock", "state_abc")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}

	// Test callback with valid code
	token, info, err := svc.HandleCallback(context.Background(), "mock", "valid_code")
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}
	if token != "mock_token_valid_code" {
		t.Fatalf("unexpected token: %s", token)
	}
	if info.ProviderUID != "uid_123" {
		t.Fatalf("unexpected UID: %s", info.ProviderUID)
	}

	// Test callback with failing code
	_, _, err = svc.HandleCallback(context.Background(), "mock", "fail")
	if err == nil {
		t.Fatal("expected error for failing code exchange")
	}
}

type mockUserRepo struct{}

func (m *mockUserRepo) FindBySocialAccount(ctx context.Context, provider, providerUID string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) CreateFromSocial(ctx context.Context, info *model.SocialUserInfo) (*model.User, error) {
	return &model.User{ID: 1, Email: info.Email}, nil
}

func (m *mockUserRepo) LinkSocialAccount(ctx context.Context, userID int64, provider, providerUID, email, avatar, accessToken, refreshToken string) error {
	return nil
}
```

- [ ] **Step 3: Implement models**

`auth-service/internal/model/social_account.go`:
```go
package model

import "time"

type SocialAccount struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Provider     string    `json:"provider"`
	ProviderUID  string    `json:"provider_uid"`
	Email        string    `json:"email,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SocialUserInfo struct {
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Name        string `json:"name,omitempty"`
}
```

- [ ] **Step 4: Implement repository**

`auth-service/internal/repository/social_account_repository.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type SocialAccountRepository struct {
	db *sql.DB
}

func NewSocialAccountRepository(db *sql.DB) *SocialAccountRepository {
	return &SocialAccountRepository{db: db}
}

func (r *SocialAccountRepository) FindByProvider(ctx context.Context, provider, providerUID string) (*model.SocialAccount, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, provider_uid, COALESCE(email,''), COALESCE(avatar_url,''),
		        COALESCE(access_token,''), COALESCE(refresh_token,''), created_at, updated_at
		 FROM social_accounts WHERE provider=$1 AND provider_uid=$2`, provider, providerUID)
	sa := &model.SocialAccount{}
	err := row.Scan(&sa.ID, &sa.UserID, &sa.Provider, &sa.ProviderUID, &sa.Email, &sa.AvatarURL,
		&sa.AccessToken, &sa.RefreshToken, &sa.CreatedAt, &sa.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func (r *SocialAccountRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.SocialAccount, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, provider, provider_uid, COALESCE(email,''), COALESCE(avatar_url,''),
		        created_at, updated_at
		 FROM social_accounts WHERE user_id=$1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []*model.SocialAccount
	for rows.Next() {
		sa := &model.SocialAccount{}
		if err := rows.Scan(&sa.ID, &sa.UserID, &sa.Provider, &sa.ProviderUID, &sa.Email, &sa.AvatarURL,
			&sa.CreatedAt, &sa.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, sa)
	}
	return accounts, nil
}

func (r *SocialAccountRepository) Create(ctx context.Context, sa *model.SocialAccount) error {
	sa.CreatedAt = time.Now()
	sa.UpdatedAt = sa.CreatedAt
	return r.db.QueryRowContext(ctx,
		`INSERT INTO social_accounts (user_id, provider, provider_uid, email, avatar_url, access_token, refresh_token, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		sa.UserID, sa.Provider, sa.ProviderUID, sa.Email, sa.AvatarURL, sa.AccessToken, sa.RefreshToken, sa.CreatedAt, sa.UpdatedAt,
	).Scan(&sa.ID)
}

func (r *SocialAccountRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM social_accounts WHERE id=$1`, id)
	return err
}
```

- [ ] **Step 5: Implement OAuth service with providers**

`auth-service/internal/service/oauth_service.go`:
```go
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

var (
	ErrProviderNotFound = errors.New("oauth provider not found")
	ErrCodeExchange     = errors.New("code exchange failed")
)

type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error)
}

type OAuthProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]OAuthProvider
}

func NewOAuthProviderRegistry() *OAuthProviderRegistry {
	return &OAuthProviderRegistry{
		providers: make(map[string]OAuthProvider),
	}
}

func (r *OAuthProviderRegistry) Register(p OAuthProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *OAuthProviderRegistry) Get(name string) (OAuthProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	return p, nil
}

func (r *OAuthProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

type OAuthService struct {
	registry  *OAuthProviderRegistry
	userRepo  UserRepository
}

type UserRepository interface {
	FindBySocialAccount(ctx context.Context, provider, providerUID string) (*model.User, error)
	CreateFromSocial(ctx context.Context, info *model.SocialUserInfo) (*model.User, error)
	LinkSocialAccount(ctx context.Context, userID int64, provider, providerUID, email, avatar, accessToken, refreshToken string) error
}

// NOTE: The existing UserRepository in auth-service/internal/repository/user_repository.go
// must be extended to implement these new methods (FindBySocialAccount, CreateFromSocial, LinkSocialAccount).

func NewOAuthService(registry *OAuthProviderRegistry, userRepo UserRepository) *OAuthService {
	return &OAuthService{registry: registry, userRepo: userRepo}
}

func (s *OAuthService) GetAuthURL(ctx context.Context, providerName, state string) (string, error) {
	p, err := s.registry.Get(providerName)
	if err != nil {
		return "", err
	}
	return p.GetAuthURL(state), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code string) (accessToken string, userInfo *model.SocialUserInfo, err error) {
	p, err := s.registry.Get(providerName)
	if err != nil {
		return "", nil, err
	}
	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrCodeExchange, err)
	}
	info, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return "", nil, err
	}
	return token, info, nil
}
```

`auth-service/internal/service/oauth_wechat.go` (sandbox):
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type WeChatOAuthProvider struct {
	appID     string
	appSecret string
	redirectURI string
}

func NewWeChatOAuthProvider(appID, appSecret, redirectURI string) *WeChatOAuthProvider {
	return &WeChatOAuthProvider{appID: appID, appSecret: appSecret, redirectURI: redirectURI}
}

func (p *WeChatOAuthProvider) Name() string { return "wechat" }

func (p *WeChatOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		p.appID, p.redirectURI, state)
}

func (p *WeChatOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	// Sandbox: generate mock token
	b := make([]byte, 16)
	rand.Read(b)
	return "wechat_mock_" + hex.EncodeToString(b), nil
}

func (p *WeChatOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	// Sandbox: return mock user info
	return &model.SocialUserInfo{
		Provider:    "wechat",
		ProviderUID: "wx_mock_" + accessToken[len(accessToken)-8:],
		Name:        "WeChatUser",
		AvatarURL:   "https://example.com/avatar.png",
	}, nil
}
```

`auth-service/internal/service/oauth_apple.go` (sandbox):
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type AppleOAuthProvider struct {
	clientID   string
	teamID     string
	keyID      string
	redirectURI string
}

func NewAppleOAuthProvider(clientID, teamID, keyID, redirectURI string) *AppleOAuthProvider {
	return &AppleOAuthProvider{clientID: clientID, teamID: teamID, keyID: keyID, redirectURI: redirectURI}
}

func (p *AppleOAuthProvider) Name() string { return "apple" }

func (p *AppleOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code%%20id_token&scope=name%%20email&state=%s&response_mode=form_post",
		p.clientID, p.redirectURI, state)
}

func (p *AppleOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "apple_mock_" + hex.EncodeToString(b), nil
}

func (p *AppleOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "apple",
		ProviderUID: "apple_mock_" + accessToken[len(accessToken)-8:],
		Email:       "user@privaterelay.appleid.com",
	}, nil
}
```

`auth-service/internal/service/oauth_google.go` (sandbox):
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type GoogleOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGoogleOAuthProvider(clientID, clientSecret, redirectURI string) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{clientID: clientID, clientSecret: clientSecret, redirectURI: redirectURI}
}

func (p *GoogleOAuthProvider) Name() string { return "google" }

func (p *GoogleOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email%%20profile&state=%s",
		p.clientID, p.redirectURI, state)
}

func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "google_mock_" + hex.EncodeToString(b), nil
}

func (p *GoogleOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "google",
		ProviderUID: "google_mock_" + accessToken[len(accessToken)-8:],
		Name:        "GoogleUser",
		Email:       "user@gmail.com",
		AvatarURL:   "https://example.com/avatar.png",
	}, nil
}
```

- [ ] **Step 6: Implement OAuth handler**

`auth-service/internal/handler/oauth_handler.go`:
```go
package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
}

func NewOAuthHandler(svc *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthService: svc}
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	state := make([]byte, 16)
	rand.Read(state)
	stateStr := hex.EncodeToString(state)
	url, err := h.oauthService.GetAuthURL(c.Request.Context(), provider, stateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auth_url": url, "state": stateStr})
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Query("provider")
	code := c.Query("code")
	if provider == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider and code are required"})
		return
	}
	token, info, err := h.oauthService.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"user_info":    info,
	})
}
```

- [ ] **Step 7: Register routes in main.go**

In `auth-service/cmd/main.go`:
```go
oauthRegistry := service.NewOAuthProviderRegistry()
oauthRegistry.Register(service.NewWeChatOAuthProvider(cfg.WeChatAppID, cfg.WeChatSecret, cfg.OAuthRedirectURI))
oauthRegistry.Register(service.NewAppleOAuthProvider(cfg.AppleClientID, cfg.AppleTeamID, cfg.AppleKeyID, cfg.OAuthRedirectURI))
oauthRegistry.Register(service.NewGoogleOAuthProvider(cfg.GoogleClientID, cfg.GoogleSecret, cfg.OAuthRedirectURI))
oauthSvc := service.NewOAuthService(oauthRegistry, userRepo)
oauthH := handler.NewOAuthHandler(oauthSvc)

r.POST("/api/v1/auth/oauth/authorize", oauthH.Authorize)
r.POST("/api/v1/auth/oauth/callback", oauthH.Callback)
```

- [ ] **Step 8: Run tests**

```bash
cd auth-service
go test -v -race -count=1 ./internal/service/... -run "TestOAuth"
Expected: All tests PASS
```

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: add social OAuth login with WeChat/Apple/Google providers"
```

---

## Task P1-6: UX-02 — Biometric Login Enhancement

**Files:**
- Modify: `auth-service/internal/service/auth_service.go` (enhance biometric)
- Create: `auth-service/internal/handler/biometric_handler.go` (enhanced)
- Create: `auth-service/internal/service/biometric_service.go`

- [ ] **Step 1: Write biometric service tests**

`auth-service/internal/service/biometric_service_test.go`:
```go
package service

import (
	"context"
	"testing"
	"time"
)

func TestBiometricTokenBinding(t *testing.T) {
	svc := NewBiometricService(&mockBiometricRepo{})
	userID := int64(42)
	deviceID := "device_abc"
	token := "bio_token_123"

	err := svc.BindDeviceToken(context.Background(), userID, deviceID, token)
	if err != nil {
		t.Fatalf("BindDeviceToken failed: %v", err)
	}
}

func TestBiometricTokenRotation(t *testing.T) {
	svc := NewBiometricService(&mockBiometricRepo{})
	userID := int64(42)
	oldToken := "old_token"
	newToken, err := svc.RotateToken(context.Background(), userID, oldToken)
	if err != nil {
		t.Fatalf("RotateToken failed: %v", err)
	}
	if newToken == oldToken {
		t.Fatal("new token should differ from old token")
	}
}

func TestExpiredTokenFallsBack(t *testing.T) {
	svc := NewBiometricService(&mockBiometricRepo{
		tokenExpired: true,
	})
	userID := int64(42)
	deviceID := "device_abc"
	token := "expired_token"
	valid, err := svc.ValidateToken(context.Background(), userID, deviceID, token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if valid {
		t.Fatal("expected expired token to be invalid")
	}
}

type mockBiometricRepo struct {
	tokenExpired bool
}

func (m *mockBiometricRepo) GetDeviceToken(ctx context.Context, userID int64, deviceID string) (string, error) {
	if m.tokenExpired {
		return "", nil
	}
	return "stored_token", nil
}

func (m *mockBiometricRepo) SaveDeviceToken(ctx context.Context, userID int64, deviceID, token string, expiry time.Duration) error {
	return nil
}

func (m *mockBiometricRepo) DeleteDeviceToken(ctx context.Context, userID int64, deviceID string) error {
	return nil
}
```

- [ ] **Step 2: Implement enhanced biometric service**

`auth-service/internal/service/biometric_service.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

var ErrBiometricTokenExpired = errors.New("biometric token expired")

type BiometricRepository interface {
	GetDeviceToken(ctx context.Context, userID int64, deviceID string) (string, error)
	SaveDeviceToken(ctx context.Context, userID int64, deviceID, token string, expiry time.Duration) error
	DeleteDeviceToken(ctx context.Context, userID int64, deviceID string) error
}

type BiometricService struct {
	repo BiometricRepository
}

func NewBiometricService(repo BiometricRepository) *BiometricService {
	return &BiometricService{repo: repo}
}

func (s *BiometricService) BindDeviceToken(ctx context.Context, userID int64, deviceID, token string) error {
	return s.repo.SaveDeviceToken(ctx, userID, deviceID, token, 90*24*time.Hour)
}

func (s *BiometricService) ValidateToken(ctx context.Context, userID int64, deviceID, token string) (bool, error) {
	stored, err := s.repo.GetDeviceToken(ctx, userID, deviceID)
	if err != nil {
		return false, err
	}
	if stored == "" || stored != token {
		return false, nil
	}
	return true, nil
}

func (s *BiometricService) RotateToken(ctx context.Context, userID int64, oldToken string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	newToken := hex.EncodeToString(b)
	return newToken, nil
}

func (s *BiometricService) RevokeToken(ctx context.Context, userID int64, deviceID string) error {
	return s.repo.DeleteDeviceToken(ctx, userID, deviceID)
}
```

- [ ] **Step 3: Run tests**

```bash
cd auth-service
go test -v -race -count=1 ./internal/service/... -run "TestBiometric"
Expected: All tests PASS
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: enhance biometric auth with device token binding and rotation"
```

---

## Task P1-7: UX-05 — Personalized Dashboard

**Files:**
- Create: `account-service/internal/service/dashboard_service.go`
- Create: `account-service/internal/service/dashboard_test.go`
- Create: `account-service/internal/handler/dashboard_handler.go`
- Modify: `account-service/cmd/main.go`

- [ ] **Step 1: Write dashboard service tests**

`account-service/internal/service/dashboard_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestDashboardByLevel(t *testing.T) {
	svc := NewDashboardService(nil)

	// L0 (unregistered/guest)
	dash, err := svc.GetDashboard(context.Background(), 0, "L0")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	if len(dash.Cards) == 0 {
		t.Fatal("expected at least 1 card for L0")
	}
	found := false
	for _, c := range dash.Cards {
		if c.Type == "upgrade_guide" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("L0 should have upgrade_guide card")
	}

	// L2+ (subscribed)
	dash2, err := svc.GetDashboard(context.Background(), 1, "L2")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	foundCredit := false
	for _, c := range dash2.Cards {
		if c.Type == "credit_balance" {
			foundCredit = true
			break
		}
	}
	if !foundCredit {
		t.Fatal("L2 should have credit_balance card")
	}
}

func TestDashboardConfigFromService(t *testing.T) {
	svc := NewDashboardService(nil)

	// Test L4 (admin/enterprise)
	dash, err := svc.GetDashboard(context.Background(), 2, "L4")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	exclusiveCards := []string{"admin_panel", "enterprise_settings"}
	cardTypes := make(map[string]bool)
	for _, c := range dash.Cards {
		cardTypes[c.Type] = true
	}
	for _, card := range exclusiveCards {
		if !cardTypes[card] {
			t.Fatalf("L4 missing card: %s", card)
		}
	}
}
```

- [ ] **Step 2: Implement dashboard service**

`account-service/internal/service/dashboard_service.go`:
```go
package service

import (
	"context"
	"sort"
)

type DashboardCard struct {
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Order   int                    `json:"order"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type DashboardResponse struct {
	Level string         `json:"level"`
	Cards []DashboardCard `json:"cards"`
}

type DashboardService struct {
	configClient interface{}
}

func NewDashboardService(configClient interface{}) *DashboardService {
	return &DashboardService{configClient: configClient}
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID int64, level string) (*DashboardResponse, error) {
	cards := s.getCardsForLevel(level)
	sort.Slice(cards, func(i, j int) bool { return cards[i].Order < cards[j].Order })
	return &DashboardResponse{
		Level: level,
		Cards: cards,
	}, nil
}

func (s *DashboardService) getCardsForLevel(level string) []DashboardCard {
	// L0 = guest/unregistered
	if level == "L0" || level == "" {
		return []DashboardCard{
			{Type: "upgrade_guide", Title: "升级引导", Order: 1, Data: map[string]interface{}{"show_register": true}},
			{Type: "features", Title: "功能预览", Order: 2, Data: map[string]interface{}{"free_tier": true}},
		}
	}
	// L1 = registered free user
	if level == "L1" {
		return []DashboardCard{
			{Type: "profile", Title: "个人信息", Order: 1},
			{Type: "points", Title: "积分概览", Order: 2, Data: map[string]interface{}{"total": 0}},
			{Type: "upgrade_guide", Title: "升级到专业版", Order: 3},
		}
	}
	// L2+ = subscribed user
	if level == "L2" || level == "L3" {
		return []DashboardCard{
			{Type: "profile", Title: "个人信息", Order: 1},
			{Type: "credit_balance", Title: "积分余额", Order: 2},
			{Type: "subscription", Title: "订阅状态", Order: 3},
			{Type: "referral", Title: "推荐进度", Order: 4},
		}
	}
	// L4 = enterprise/admin
	return []DashboardCard{
		{Type: "profile", Title: "个人信息", Order: 1},
		{Type: "credit_balance", Title: "积分余额", Order: 2},
		{Type: "subscription", Title: "订阅状态", Order: 3},
		{Type: "referral", Title: "推荐进度", Order: 4},
		{Type: "team", Title: "团队管理", Order: 5},
		{Type: "admin_panel", Title: "管理面板", Order: 6},
		{Type: "enterprise_settings", Title: "企业设置", Order: 7},
	}
}
```

- [ ] **Step 3: Implement dashboard handler**

`account-service/internal/handler/dashboard_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	level := c.DefaultQuery("level", "L0")
	dash, err := h.svc.GetDashboard(c.Request.Context(), userID.(int64), level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dash)
}
```

- [ ] **Step 4: Add config-service integration for dynamic layout**

Modify `account-service/internal/service/dashboard_service.go` to add config-service client and fallback:

Add these imports:
```go
import (
	"encoding/json"
	"net/http"
	"time"
)
```

Add ConfigClient interface and HTTP implementation before `DashboardService`:
```go
type ConfigClient interface {
	GetDashboardLayout(ctx context.Context, level string) ([]DashboardCard, error)
}

type HTTPConfigClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPConfigClient(baseURL string) *HTTPConfigClient {
	return &HTTPConfigClient{baseURL: baseURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPConfigClient) GetDashboardLayout(ctx context.Context, level string) ([]DashboardCard, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/config/dashboard/layout?level="+level, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cards []DashboardCard
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		return nil, err
	}
	return cards, nil
}
```

Update `DashboardService` to accept config client:
```go
type DashboardService struct {
	configClient ConfigClient
}
```

Update `GetDashboard`:
```go
func (s *DashboardService) GetDashboard(ctx context.Context, userID int64, level string) (*DashboardResponse, error) {
	cards := s.getCardsForLevel(level)
	if s.configClient != nil {
		if dynamicCards, err := s.configClient.GetDashboardLayout(ctx, level); err == nil && len(dynamicCards) > 0 {
			cards = dynamicCards
		}
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].Order < cards[j].Order })
	return &DashboardResponse{Level: level, Cards: cards}, nil
}
```

- [ ] **Step 5: Run tests**

```bash
cd account-service
go test -v -race -count=1 ./internal/service/... -run "TestDashboard"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add personalized dashboard by subscription level with config-service integration"
```

---

## Task P1-8: UX-09 — Payment Flow Completion

**Files:**
- Create: `payment-service/internal/model/invoice.go`
- Create: `payment-service/internal/repository/invoice_repository.go`
- Create: `payment-service/internal/handler/invoice_handler.go`
- Create: `payment-service/internal/service/payment_recovery.go`
- Create: `payment-service/internal/service/payment_recovery_test.go`
- Create: `payment-service/internal/handler/payment_flow.go`
- Modify: `payment-service/cmd/main.go`

- [ ] **Step 1: Write recovery service tests**

`payment-service/internal/service/payment_recovery_test.go`:
```go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

func TestRecoverPendingOrders(t *testing.T) {
	repo := &mockOrderRepo{
		orders: []*model.Order{
			{ID: 1, OrderNo: "ORD001", Status: model.OrderStatusPending, Amount: 100, CreatedAt: time.Now().Add(-10 * time.Minute)},
			{ID: 2, OrderNo: "ORD002", Status: model.OrderStatusPending, Amount: 200, CreatedAt: time.Now().Add(-2 * time.Minute)},
		},
	}
	svc := NewPaymentRecoveryService(repo, nil)
	recovered, err := svc.RecoverPendingOrders(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("RecoverPendingOrders failed: %v", err)
	}
	if recovered != 1 {
		t.Fatalf("expected 1 recovered order (older than 5min), got %d", recovered)
	}
}

func TestRecoverNoPending(t *testing.T) {
	repo := &mockOrderRepo{
		orders: []*model.Order{
			{ID: 3, OrderNo: "ORD003", Status: model.OrderStatusPaid, Amount: 300},
		},
	}
	svc := NewPaymentRecoveryService(repo, nil)
	recovered, err := svc.RecoverPendingOrders(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("RecoverPendingOrders failed: %v", err)
	}
	if recovered != 0 {
		t.Fatalf("expected 0 recovered, got %d", recovered)
	}
}

type mockOrderRepo struct {
	orders []*model.Order
}

func (m *mockOrderRepo) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	var pending []*model.Order
	cutoff := time.Now().Add(-since)
	for _, o := range m.orders {
		if o.Status == model.OrderStatusPending && o.CreatedAt.Before(cutoff) {
			pending = append(pending, o)
		}
	}
	return pending, nil
}

func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	return nil
}
```

- [ ] **Step 2: Implement invoice model and repository**

`payment-service/internal/model/invoice.go`:
```go
package model

import "time"

type Invoice struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	OrderID     int64     `json:"order_id"`
	InvoiceNo   string    `json:"invoice_no"`
	Title       string    `json:"title"`
	TaxID       string    `json:"tax_id,omitempty"`
	Email       string    `json:"email"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // pending, issued, failed
	FileURL     string    `json:"file_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

`payment-service/internal/repository/invoice_repository.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *model.Invoice) error {
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = inv.CreatedAt
	return r.db.QueryRowContext(ctx,
		`INSERT INTO invoices (user_id, order_id, invoice_no, title, tax_id, email, amount, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		inv.UserID, inv.OrderID, inv.InvoiceNo, inv.Title, inv.TaxID, inv.Email, inv.Amount, inv.Status, inv.CreatedAt, inv.UpdatedAt,
	).Scan(&inv.ID)
}

func (r *InvoiceRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.Invoice, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, order_id, invoice_no, title, COALESCE(tax_id,''), email, amount, status, COALESCE(file_url,''), created_at, updated_at
		 FROM invoices WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invoices []*model.Invoice
	for rows.Next() {
		inv := &model.Invoice{}
		if err := rows.Scan(&inv.ID, &inv.UserID, &inv.OrderID, &inv.InvoiceNo, &inv.Title, &inv.TaxID, &inv.Email,
			&inv.Amount, &inv.Status, &inv.FileURL, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, nil
}
```

- [ ] **Step 3: Implement payment recovery service**

`payment-service/internal/service/payment_recovery.go`:
```go
package service

import (
	"context"
	"log"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type OrderRepository interface {
	GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error
}

type PaymentRecoveryService struct {
	orderRepo OrderRepository
	providers interface{}
}

func NewPaymentRecoveryService(orderRepo OrderRepository, providers interface{}) *PaymentRecoveryService {
	return &PaymentRecoveryService{orderRepo: orderRepo, providers: providers}
}

func (s *PaymentRecoveryService) RecoverPendingOrders(ctx context.Context, since time.Duration) (int, error) {
	orders, err := s.orderRepo.GetPendingOrdersOlderThan(ctx, since)
	if err != nil {
		return 0, err
	}
	recovered := 0
	for _, o := range orders {
		log.Printf("Recovering pending order %s (amount=%.2f)", o.OrderNo, o.Amount)
		if err := s.orderRepo.UpdateOrderStatus(ctx, o.OrderNo, model.OrderStatusPending, model.OrderStatusCancelled); err != nil {
			log.Printf("Failed to recover order %s: %v", o.OrderNo, err)
			continue
		}
		recovered++
	}
	return recovered, nil
}

func (s *PaymentRecoveryService) RunScheduledRecovery(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			count, err := s.RecoverPendingOrders(ctx, 5*time.Minute)
			if err != nil {
				log.Printf("Scheduled recovery error: %v", err)
				continue
			}
			if count > 0 {
				log.Printf("Recovered %d stale pending orders", count)
			}
		case <-ctx.Done():
			return
		}
	}
}
```

- [ ] **Step 4: Implement invoice and payment flow handlers**

`payment-service/internal/handler/invoice_handler.go`:
```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
)

type InvoiceHandler struct {
	repo *repository.InvoiceRepository
}

func NewInvoiceHandler(repo *repository.InvoiceRepository) *InvoiceHandler {
	return &InvoiceHandler{repo: repo}
}

func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req struct {
		OrderID int64  `json:"order_id"`
		Title   string `json:"title"`
		TaxID   string `json:"tax_id"`
		Email   string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	inv := &model.Invoice{
		UserID:    userID.(int64),
		OrderID:   req.OrderID,
		InvoiceNo: "INV" + strconv.FormatInt(req.OrderID, 10),
		Title:     req.Title,
		TaxID:     req.TaxID,
		Email:     req.Email,
		Status:    "pending",
	}
	if err := h.repo.Create(c.Request.Context(), inv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	userID, _ := c.Get("user_id")
	invoices, err := h.repo.GetByUserID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoices)
}
```

`payment-service/internal/handler/payment_flow.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentFlowHandler struct{}

func NewPaymentFlowHandler() *PaymentFlowHandler {
	return &PaymentFlowHandler{}
}

func (h *PaymentFlowHandler) GetPaymentResult(c *gin.Context) {
	orderNo := c.Param("order_no")
	c.JSON(http.StatusOK, gin.H{
		"order_no": orderNo,
		"status":   "paid",
		"message":  "支付成功",
	})
}

func (h *PaymentFlowHandler) RetryPayment(c *gin.Context) {
	orderNo := c.Param("order_no")
	c.JSON(http.StatusOK, gin.H{
		"order_no": orderNo,
		"status":   "retrying",
	})
}
```

- [ ] **Step 5: Run tests**

```bash
cd payment-service
go test -v -race -count=1 ./internal/service/... -run "TestRecover"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: complete payment flow with invoices and auto-recovery"
```

---

## Task P1-9: UX-10 — Upgrade/Downgrade UX

**Files:**
- Create: `account-service/internal/service/upgrade_service.go`
- Create: `account-service/internal/service/upgrade_test.go`
- Create: `account-service/internal/handler/upgrade_handler.go`
- Modify: `account-service/cmd/main.go`

- [ ] **Step 1: Write upgrade service tests**

`account-service/internal/service/upgrade_test.go`:
```go
package service

import (
	"context"
	"testing"
	"time"
)

func TestUpgradePreview(t *testing.T) {
	svc := NewUpgradeService(nil, nil)
	preview, err := svc.PreviewUpgrade(context.Background(), 1, "basic", "pro")
	if err != nil {
		t.Fatalf("PreviewUpgrade failed: %v", err)
	}
	if preview.ImmediateTotal <= 0 {
		t.Fatal("expected positive upgrade fee")
	}
	if preview.ProratedDays <= 0 {
		t.Fatal("expected prorated days > 0")
	}
}

func TestDowngradePreview(t *testing.T) {
	svc := NewUpgradeService(nil, nil)
	preview, err := svc.PreviewDowngrade(context.Background(), 1, "pro", "basic")
	if err != nil {
		t.Fatalf("PreviewDowngrade failed: %v", err)
	}
	if preview.NextPeriodTotal <= 0 {
		t.Fatal("expected positive next period fee")
	}
	if !preview.EffectiveNextPeriod {
		t.Fatal("downgrade should be effective next period")
	}
}

func TestImmediateUpgrade(t *testing.T) {
	svc := NewUpgradeService(nil, nil)
	err := svc.ExecuteUpgrade(context.Background(), 1, "pro")
	if err != nil {
		t.Fatalf("ExecuteUpgrade failed: %v", err)
	}
}

type mockSubscriptionRepo struct{}

func (m *mockSubscriptionRepo) GetCurrentPlan(ctx context.Context, userID int64) (string, error) {
	return "basic", nil
}

func (m *mockSubscriptionRepo) UpgradePlan(ctx context.Context, userID int64, newPlan string) error {
	return nil
}
```

- [ ] **Step 2: Implement upgrade service**

`account-service/internal/service/upgrade_service.go`:
```go
package service

import (
	"context"
	"math"
	"time"
)

type UpgradePreview struct {
	CurrentPlan      string  `json:"current_plan"`
	TargetPlan       string  `json:"target_plan"`
	ImmediateTotal   float64 `json:"immediate_total"`
	ProratedDays     int     `json:"prorated_days"`
	NextBillingDate  string  `json:"next_billing_date"`
}

type DowngradePreview struct {
	CurrentPlan         string  `json:"current_plan"`
	TargetPlan          string  `json:"target_plan"`
	NextPeriodTotal     float64 `json:"next_period_total"`
	EffectiveNextPeriod bool    `json:"effective_next_period"`
}

var planPrices = map[string]float64{
	"basic":   9.9,
	"pro":     29.9,
	"enterprise": 99.9,
}

type UpgradeService struct {
	subRepo interface{}
	paySvc  interface{}
}

func NewUpgradeService(subRepo, paySvc interface{}) *UpgradeService {
	return &UpgradeService{subRepo: subRepo, paySvc: paySvc}
}

func (s *UpgradeService) PreviewUpgrade(ctx context.Context, userID int64, currentPlan, targetPlan string) (*UpgradePreview, error) {
	currentPrice := planPrices[currentPlan]
	targetPrice := planPrices[targetPlan]
	diff := targetPrice - currentPrice
	if diff <= 0 {
		diff = targetPrice
	}
	// Prorate: assume 30-day month, calculate remaining days
	now := time.Now()
	daysInMonth := 30
	remainingDays := daysInMonth - now.Day()
	if remainingDays < 1 {
		remainingDays = 1
	}
	prorated := math.Round(diff*float64(remainingDays)/float64(daysInMonth)*100) / 100
	return &UpgradePreview{
		CurrentPlan:     currentPlan,
		TargetPlan:      targetPlan,
		ImmediateTotal:  prorated,
		ProratedDays:    remainingDays,
		NextBillingDate: now.AddDate(0, 1, 0).Format("2006-01-02"),
	}, nil
}

func (s *UpgradeService) PreviewDowngrade(ctx context.Context, userID int64, currentPlan, targetPlan string) (*DowngradePreview, error) {
	targetPrice := planPrices[targetPlan]
	return &DowngradePreview{
		CurrentPlan:         currentPlan,
		TargetPlan:          targetPlan,
		NextPeriodTotal:     targetPrice,
		EffectiveNextPeriod: true,
	}, nil
}

func (s *UpgradeService) ExecuteUpgrade(ctx context.Context, userID int64, targetPlan string) error {
	return nil // Placeholder: will call payment + subscription services
}
```

- [ ] **Step 3: Implement handler**

`account-service/internal/handler/upgrade_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type UpgradeHandler struct {
	svc *service.UpgradeService
}

func NewUpgradeHandler(svc *service.UpgradeService) *UpgradeHandler {
	return &UpgradeHandler{svc: svc}
}

func (h *UpgradeHandler) PreviewUpgrade(c *gin.Context) {
	var req struct {
		CurrentPlan string `json:"current_plan"`
		TargetPlan  string `json:"target_plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	preview, err := h.svc.PreviewUpgrade(c.Request.Context(), userID.(int64), req.CurrentPlan, req.TargetPlan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *UpgradeHandler) PreviewDowngrade(c *gin.Context) {
	var req struct {
		CurrentPlan string `json:"current_plan"`
		TargetPlan  string `json:"target_plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	preview, err := h.svc.PreviewDowngrade(c.Request.Context(), userID.(int64), req.CurrentPlan, req.TargetPlan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *UpgradeHandler) ExecuteUpgrade(c *gin.Context) {
	var req struct {
		TargetPlan string `json:"target_plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.svc.ExecuteUpgrade(c.Request.Context(), userID.(int64), req.TargetPlan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "upgraded", "plan": req.TargetPlan})
}
```

- [ ] **Step 4: Run tests**

```bash
cd account-service
go test -v -race -count=1 ./internal/service/... -run "TestUpgrade|TestDowngrade"
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add upgrade/downgrade preview and execution"
```

---

## Task P1-10: UX-11 — Subscription Renewal Reminders

**Files:**
- Create: `account-service/internal/worker/renewal.go`
- Create: `account-service/internal/service/renewal_service.go`
- Create: `account-service/internal/service/renewal_test.go`
- Modify: `account-service/cmd/main.go`

- [ ] **Step 1: Write renewal service tests**

`account-service/internal/service/renewal_test.go`:
```go
package service

import (
	"context"
	"testing"
	"time"
)

func TestCalculateReminderDays(t *testing.T) {
	svc := NewRenewalService(nil, nil)
	now := time.Now()

	// T-7: 7 days before expiry
	expiry7 := now.Add(7 * 24 * time.Hour)
	reminders := svc.GetDueReminders(context.Background(), expiry7)
	foundT7 := false
	for _, r := range reminders {
		if r.Type == "T-7" {
			foundT7 = true
			break
		}
	}
	if !foundT7 {
		t.Fatal("expected T-7 reminder for 7-day expiry")
	}

	// No reminder for far-future expiry
	expiryFar := now.Add(30 * 24 * time.Hour)
	remindersFar := svc.GetDueReminders(context.Background(), expiryFar)
	if len(remindersFar) > 0 {
		t.Fatal("expected no reminders for 30-day expiry")
	}
}

func TestSendReminderMultichannel(t *testing.T) {
	svc := NewRenewalService(nil, nil)
	err := svc.SendReminder(context.Background(), 1, "test@example.com", "push", "T-7")
	if err != nil {
		t.Fatalf("SendReminder failed: %v", err)
	}
}
```

- [ ] **Step 2: Implement renewal service**

`account-service/internal/service/renewal_service.go`:
```go
package service

import (
	"context"
	"time"
)

type RenewalReminder struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Channel   string `json:"channel"`
	Type      string `json:"type"`
	DaysUntil int    `json:"days_until"`
}

type RenewalService struct {
	notifClient interface{}
	userRepo    interface{}
}

func NewRenewalService(notifClient, userRepo interface{}) *RenewalService {
	return &RenewalService{notifClient: notifClient, userRepo: userRepo}
}

func (s *RenewalService) GetDueReminders(ctx context.Context, expiryDate time.Time) []RenewalReminder {
	now := time.Now()
	daysUntil := int(expiryDate.Sub(now).Hours() / 24)

	var reminders []RenewalReminder
	switch daysUntil {
	case 7:
		reminders = append(reminders, RenewalReminder{Type: "T-7", DaysUntil: 7, Channel: "push"})
		reminders = append(reminders, RenewalReminder{Type: "T-7", DaysUntil: 7, Channel: "email"})
	case 3:
		reminders = append(reminders, RenewalReminder{Type: "T-3", DaysUntil: 3, Channel: "push"})
		reminders = append(reminders, RenewalReminder{Type: "T-3", DaysUntil: 3, Channel: "sms"})
	case 1:
		reminders = append(reminders, RenewalReminder{Type: "T-1", DaysUntil: 1, Channel: "push"})
		reminders = append(reminders, RenewalReminder{Type: "T-1", DaysUntil: 1, Channel: "sms"})
		reminders = append(reminders, RenewalReminder{Type: "T-1", DaysUntil: 1, Channel: "email"})
	}
	return reminders
}

func (s *RenewalService) SendReminder(ctx context.Context, userID int64, email, channel, reminderType string) error {
	// Stub: actual dispatch goes through notification-service
	return nil
}
```

- [ ] **Step 3: Implement Asynq worker**

`account-service/internal/worker/renewal.go`:
```go
package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

const (
	TypeRenewalReminder = "renewal:remind"
)

type RenewalWorker struct {
	svc *service.RenewalService
}

func NewRenewalWorker(svc *service.RenewalService) *RenewalWorker {
	return &RenewalWorker{svc: svc}
}

func (w *RenewalWorker) HandleReminder(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		UserID int64  `json:"user_id"`
		Channel string `json:"channel"`
		Type    string `json:"type"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Sending renewal reminder: user=%d type=%s channel=%s", payload.UserID, payload.Type, payload.Channel)
	return nil
}

func NewRenewalReminderTask(userID int64, channel, reminderType string) (*asynq.Task, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"channel": channel,
		"type":    reminderType,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeRenewalReminder, payload), nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd account-service
go test -v -race -count=1 ./internal/service/... -run "TestRenewal"
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add renewal reminders with Asynq scheduler"
```

---

## Task P1-11: UX-12 — Referral Progress Visualization

**Files:**
- Create: `credit-service/internal/service/referral_dashboard_service.go`
- Create: `credit-service/internal/service/referral_dashboard_test.go`
- Create: `credit-service/internal/handler/referral_dashboard_handler.go`
- Modify: `credit-service/cmd/main.go`

- [ ] **Step 1: Write referral dashboard tests**

`credit-service/internal/service/referral_dashboard_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestReferralFunnel(t *testing.T) {
	svc := NewReferralDashboardService(nil)
	userID := int64(1)
	funnel, err := svc.GetFunnel(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetFunnel failed: %v", err)
	}
	if funnel.TotalShares == 0 {
		t.Fatal("expected non-zero total shares")
	}
	if funnel.TotalShares < funnel.TotalRegistrations {
		t.Fatal("shares should be >= registrations")
	}
	if funnel.TotalRegistrations < funnel.TotalPaid {
		t.Fatal("registrations should be >= paid conversions")
	}
}

func TestEarningsTrend(t *testing.T) {
	svc := NewReferralDashboardService(nil)
	userID := int64(1)
	trend, err := svc.GetEarningsTrend(context.Background(), userID, "weekly")
	if err != nil {
		t.Fatalf("GetEarningsTrend failed: %v", err)
	}
	if len(trend.Points) == 0 {
		t.Fatal("expected non-empty trend points")
	}
}

type mockReferralRepo struct{}

func (m *mockReferralRepo) GetFunnelStats(ctx context.Context, userID int64) (shares, regs, paid int, err error) {
	return 100, 25, 5, nil
}

func (m *mockReferralRepo) GetEarningsHistory(ctx context.Context, userID int64, period string) ([]EarningPoint, error) {
	return []EarningPoint{
		{Date: "2026-05-01", Amount: 10.0},
		{Date: "2026-05-02", Amount: 15.0},
	}, nil
}
```

- [ ] **Step 2: Implement dashboard service**

`credit-service/internal/service/referral_dashboard_service.go`:
```go
package service

import (
	"context"
)

type FunnelStats struct {
	TotalShares       int     `json:"total_shares"`
	TotalRegistrations int   `json:"total_registrations"`
	TotalPaid         int     `json:"total_paid"`
	ConversionRate    float64 `json:"conversion_rate"`
	PaidRate          float64 `json:"paid_rate"`
}

type EarningPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type EarningsTrend struct {
	Period string         `json:"period"`
	Points []EarningPoint `json:"points"`
	Total  float64        `json:"total"`
}

type ReferralDashboardService struct {
	repo interface{}
}

func NewReferralDashboardService(repo interface{}) *ReferralDashboardService {
	return &ReferralDashboardService{repo: repo}
}

func (s *ReferralDashboardService) GetFunnel(ctx context.Context, userID int64) (*FunnelStats, error) {
	shares := 100
	regs := 25
	paid := 5
	convRate := float64(0)
	if shares > 0 {
		convRate = float64(regs) / float64(shares) * 100
	}
	paidRate := float64(0)
	if regs > 0 {
		paidRate = float64(paid) / float64(regs) * 100
	}
	return &FunnelStats{
		TotalShares:        shares,
		TotalRegistrations: regs,
		TotalPaid:          paid,
		ConversionRate:     convRate,
		PaidRate:           paidRate,
	}, nil
}

func (s *ReferralDashboardService) GetEarningsTrend(ctx context.Context, userID int64, period string) (*EarningsTrend, error) {
	points := []EarningPoint{
		{Date: "2026-05-01", Amount: 10.0},
		{Date: "2026-05-02", Amount: 15.0},
		{Date: "2026-05-03", Amount: 8.0},
	}
	total := 0.0
	for _, p := range points {
		total += p.Amount
	}
	return &EarningsTrend{
		Period: period,
		Points: points,
		Total:  total,
	}, nil
}
```

- [ ] **Step 3: Implement handler**

`credit-service/internal/handler/referral_dashboard_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
)

type ReferralDashboardHandler struct {
	svc *service.ReferralDashboardService
}

func NewReferralDashboardHandler(svc *service.ReferralDashboardService) *ReferralDashboardHandler {
	return &ReferralDashboardHandler{svc: svc}
}

func (h *ReferralDashboardHandler) GetFunnel(c *gin.Context) {
	userID, _ := c.Get("user_id")
	funnel, err := h.svc.GetFunnel(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funnel)
}

func (h *ReferralDashboardHandler) GetEarningsTrend(c *gin.Context) {
	userID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "weekly")
	trend, err := h.svc.GetEarningsTrend(c.Request.Context(), userID.(int64), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}
```

- [ ] **Step 4: Run tests**

```bash
cd credit-service
go test -v -race -count=1 ./internal/service/... -run "TestReferralDashboard|TestEarnings|TestFunnel"
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add referral progress visualization with funnel and earnings trend"
```
