# Phase 7.4 — Mobile Implementation Plan

> **For agentic workers:** Step-by-step implementation. Each task is self-contained.

**Goal:** Implement mobile client features: Android font system, token secure storage, certificate pinning, WeChat mini-program subscription messages and sharing, ad monetization foundation, and performance testing infrastructure.

**Architecture:** Kotlin for Android, Swift for iOS, TypeScript for WeChat mini-program, Go for backend APIs, k6 for performance tests.

**Tech Stack:** Android (Kotlin), iOS (Swift), WeChat Mini-Program (TypeScript), Go 1.24, Gin, k6

**Dependencies:** P1-22/P1-23 → FN-12 (stub), P1-24 → FN-12. Execute in order: P1-19 → P1-20 → P1-21 → P1-22 → P1-23 → P1-24 → P1-25.

---

## File Structure

### New files:
```
android/app/src/main/res/font/             # Font files (Inter, Space Grotesk)
android/app/src/main/res/values/fonts.xml  # Font family definitions
android/app/src/main/java/com/neuro/ac/sdk/TokenManager.kt
android/app/src/main/java/com/neuro/ac/sdk/CertificatePinner.kt

ios/NeuroAC/Sources/Security/TokenManager.swift
ios/NeuroAC/Sources/Security/CertificatePinner.swift

weapp/utils/subscribe-message.ts
weapp/utils/share.ts

notification-service/
├── internal/model/wechat_template.go
├── internal/service/wechat_template_service.go
├── internal/service/wechat_template_test.go
└── internal/handler/wechat_template_handler.go

config-service/
├── internal/model/ad_config.go
├── internal/handler/ad_config_handler.go
└── internal/service/ad_config_service.go

tests/perf/
├── config.json
├── smoke.js
├── load.js
└── stress.js
```

### Modified files:
```
notification-service/cmd/main.go
config-service/cmd/main.go
Makefile                              # Add perf-test target
```

---

## Task P1-19: MB-02 — Android Font Integration

**Files:**
- Create: `android/app/src/main/res/values/fonts.xml`
- Create: Font files in `android/app/src/main/res/font/`

- [ ] **Step 1: Create font resource definitions**

`android/app/src/main/res/values/fonts.xml`:
```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <font-family
        android:fontProviderAuthority="com.google.android.gms.fonts"
        android:fontProviderPackage="com.google.android.gms"
        android:fontProviderQuery="Inter"
        android:fontProviderCerts="@array/com_google_android_gms_fonts_certs">
    </font-family>
    <font-family
        android:fontProviderAuthority="com.google.android.gms.fonts"
        android:fontProviderPackage="com.google.android.gms"
        android:fontProviderQuery="Space+Grotesk"
        android:fontProviderCerts="@array/com_google_android_gms_fonts_certs">
    </font-family>
</resources>
```

- [ ] **Step 2: Create font certification array**

This file goes in `android/app/src/main/res/values/arrays.xml` (or add to existing):
```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <array name="com_google_android_gms_fonts_certs">
        <string>@string/com_google_android_gms_fonts_certs_dev</string>
        <string>@string/com_google_android_gms_fonts_certs_prod</string>
    </array>
</resources>
```

- [ ] **Step 3: Verify font setup**

```bash
ls -la android/app/src/main/res/font/
# Expected: Inter-Regular.ttf, Inter-Bold.ttf, SpaceGrotesk-Regular.ttf etc.
# Total size ≤ 2MB
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: add Android font integration with Inter and Space Grotesk"
```

---

## Task P1-20: MB-09 — Token Secure Storage

**Files:**
- Create: `android/app/src/main/java/com/neuro/ac/sdk/TokenManager.kt`
- Create: `ios/NeuroAC/Sources/Security/TokenManager.swift`

- [ ] **Step 1: Implement Android TokenManager**

`android/app/src/main/java/com/neuro/ac/sdk/TokenManager.kt`:
```kotlin
package com.neuro.ac.sdk

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import java.security.SecureRandom

class TokenManager(context: Context) {
    private val masterKey = MasterKey.Builder(context)
        .setKeyGenParameterSpec(
            KeyGenParameterSpec.Builder(
                MasterKey.DEFAULT_MASTER_KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setKeySize(256)
                .build()
        )
        .build()

    private val sharedPreferences = EncryptedSharedPreferences.create(
        context,
        "neuro_secure_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun saveAccessToken(token: String) {
        sharedPreferences.edit().putString("access_token", token).apply()
    }

    fun getAccessToken(): String? {
        return sharedPreferences.getString("access_token", null)
    }

    fun clearAccessToken() {
        sharedPreferences.edit().remove("access_token").apply()
    }

    fun saveRefreshToken(token: String) {
        sharedPreferences.edit().putString("refresh_token", token).apply()
    }

    fun getRefreshToken(): String? {
        return sharedPreferences.getString("refresh_token", null)
    }

    fun clearAll() {
        sharedPreferences.edit().clear().apply()
    }

    fun generateDeviceFingerprint(): String {
        val random = SecureRandom()
        val bytes = ByteArray(32)
        random.nextBytes(bytes)
        val fingerprint = bytes.joinToString("") { "%02x".format(it) }
        sharedPreferences.edit().putString("device_fingerprint", fingerprint).apply()
        return fingerprint
    }

    fun getDeviceFingerprint(): String? {
        return sharedPreferences.getString("device_fingerprint", null)
    }
}
```

- [ ] **Step 2: Implement iOS TokenManager**

`ios/NeuroAC/Sources/Security/TokenManager.swift`:
```swift
import Foundation
import Security
import CryptoKit

class TokenManager {
    static let shared = TokenManager()
    private let accessTokenKey = "com.neuro.ac.accessToken"
    private let refreshTokenKey = "com.neuro.ac.refreshToken"
    private let fingerprintKey = "com.neuro.ac.deviceFingerprint"
    private let keychainService = "com.neuro.ac.keychain"

    private func saveToKeychain(key: String, value: String) {
        guard let data = value.data(using: .utf8) else { return }
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    private func loadFromKeychain(key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    private func deleteFromKeychain(key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(query as CFDictionary)
    }

    var accessToken: String? {
        get { loadFromKeychain(key: accessTokenKey) }
        set {
            if let value = newValue {
                saveToKeychain(key: accessTokenKey, value: value)
            } else {
                deleteFromKeychain(key: accessTokenKey)
            }
        }
    }

    var refreshToken: String? {
        get { loadFromKeychain(key: refreshTokenKey) }
        set {
            if let value = newValue {
                saveToKeychain(key: refreshTokenKey, value: value)
            } else {
                deleteFromKeychain(key: refreshTokenKey)
            }
        }
    }

    func generateDeviceFingerprint() -> String {
        var bytes = [UInt8](repeating: 0, count: 32)
        _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        let fingerprint = bytes.map { String(format: "%02x", $0) }.joined()
        saveToKeychain(key: fingerprintKey, value: fingerprint)
        return fingerprint
    }

    var deviceFingerprint: String? {
        loadFromKeychain(key: fingerprintKey)
    }

    func clearAll() {
        deleteFromKeychain(key: accessTokenKey)
        deleteFromKeychain(key: refreshTokenKey)
        deleteFromKeychain(key: fingerprintKey)
    }
}
```

- [ ] **Step 3: Backend: auth-service device fingerprint validation**

Add to existing `auth-service/internal/service/auth_service.go`:
```go
func (s *AuthService) ValidateDeviceFingerprint(ctx context.Context, userID int64, fingerprint string) bool {
    if fingerprint == "" {
        return false
    }
    // In production: validate against stored fingerprint
    return len(fingerprint) == 64
}
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: add token secure storage for Android (EncryptedSharedPreferences) and iOS (Keychain)"
```

---

## Task P1-21: MB-10 — Certificate Pinning

**Files:**
- Create: `android/app/src/main/java/com/neuro/ac/sdk/CertificatePinner.kt`
- Create: `ios/NeuroAC/Sources/Security/CertificatePinner.swift`

- [ ] **Step 1: Implement Android certificate pinning**

`android/app/src/main/java/com/neuro/ac/sdk/CertificatePinner.kt`:
```kotlin
package com.neuro.ac.sdk

import okhttp3.CertificatePinner
import okhttp3.OkHttpClient
import java.security.MessageDigest

class CertificatePinnerManager {
    companion object {
        // SHA-256 hashes of the server's public keys
        private val pins = setOf(
            "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", // Primary cert
            "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="  // Backup cert
        )

        fun createPinnedClient(): OkHttpClient {
            val certificatePinner = CertificatePinner.Builder()
            pins.forEach { pin ->
                certificatePinner.add("api.neuro.com", pin)
            }
            return OkHttpClient.Builder()
                .certificatePinner(certificatePinner.build())
                .build()
        }
    }
}
```

- [ ] **Step 2: Implement iOS certificate pinning**

`ios/NeuroAC/Sources/Security/CertificatePinner.swift`:
```swift
import Foundation
import CommonCrypto

class CertificatePinnerManager {
    static let shared = CertificatePinnerManager()
    private let pinnedHashes: Set<String> = [
        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", // Primary
        "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="  // Backup
    ]

    func urlSession() -> URLSession {
        let delegate = PinningDelegate(pinnedHashes: pinnedHashes)
        return URLSession(configuration: .default, delegate: delegate, delegateQueue: nil)
    }
}

class PinningDelegate: NSObject, URLSessionDelegate {
    let pinnedHashes: Set<String>

    init(pinnedHashes: Set<String>) {
        self.pinnedHashes = pinnedHashes
    }

    func urlSession(_ session: URLSession,
                    didReceive challenge: URLAuthenticationChallenge,
                    completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void) {
        guard challenge.protectionSpace.authenticationMethod == NSURLAuthenticationMethodServerTrust,
              let serverTrust = challenge.protectionSpace.serverTrust,
              let certificate = SecTrustGetCertificateAtIndex(serverTrust, 0) else {
            completionHandler(.performDefaultHandling, nil)
            return
        }

        let policy = SecPolicyCreateBasicX509()
        var trustResult: SecTrustResultType = .invalid
        SecTrustEvaluate(serverTrust, &trustResult)
        guard trustResult == .unspecified || trustResult == .proceed else {
            completionHandler(.cancelAuthenticationChallenge, nil)
            return
        }

        let certData = SecCertificateCopyData(certificate) as Data
        var hash = [UInt8](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        certData.withUnsafeBytes { buffer in
            _ = CC_SHA256(buffer.baseAddress, CC_LONG(certData.count), &hash)
        }
        let hashString = Data(hash).base64EncodedString()

        if pinnedHashes.contains(hashString) {
            completionHandler(.useCredential, URLCredential(trust: serverTrust))
        } else {
            completionHandler(.cancelAuthenticationChallenge, nil)
        }
    }
}
```

- [ ] **Step 3: Backend: certificate hash endpoint**

Add to `api-gateway/cmd/main.go`:
```go
r.GET("/api/v1/security/pins", func(c *gin.Context) {
    c.JSON(200, gin.H{"pins": []string{"sha256/AAAA...", "sha256/BBBB..."}})
})
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: add certificate pinning for Android (OkHttp) and iOS (URLSession)"
```

---

## Task P1-22: MB-13 — Mini-Program Subscription Messages

**Files:**
- Create: `notification-service/internal/model/wechat_template.go`
- Create: `notification-service/internal/service/wechat_template_service.go`
- Create: `notification-service/internal/service/wechat_template_test.go`
- Create: `notification-service/internal/handler/wechat_template_handler.go`
- Create: `weapp/utils/subscribe-message.ts`
- Modify: `notification-service/cmd/main.go`

- [ ] **Step 1: Write tests**

`notification-service/internal/service/wechat_template_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestSendSubscriptionMessage(t *testing.T) {
	svc := NewWeChatTemplateService(nil)
	err := svc.SendSubscriptionMessage(context.Background(), "user_openid", "subscription_expiring",
		map[string]string{"name": "专业版", "date": "2026-06-01"})
	if err != nil {
		t.Fatalf("SendSubscriptionMessage failed: %v", err)
	}
}

func TestValidateTemplateID(t *testing.T) {
	svc := NewWeChatTemplateService(nil)
	err := svc.ValidateTemplateID("subscription_expiring")
	if err != nil {
		t.Fatalf("ValidateTemplateID failed: %v", err)
	}
	err = svc.ValidateTemplateID("invalid_type")
	if err == nil {
		t.Fatal("expected error for invalid template type")
	}
}
```

- [ ] **Step 2: Implement template service**

`notification-service/internal/model/wechat_template.go`:
```go
package model

type WeChatTemplate struct {
	TemplateType string `json:"template_type"`
	TemplateID   string `json:"template_id"`
	Title        string `json:"title"`
	Keywords     string `json:"keywords"`
}
```

`notification-service/internal/service/wechat_template_service.go`:
```go
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type WeChatTemplateService struct {
	client interface{}
}

var validTemplates = map[string]bool{
	"subscription_expiring": true,
	"payment_success":       true,
	"referral_bonus":        true,
	"tier_upgrade":          true,
}

func NewWeChatTemplateService(client interface{}) *WeChatTemplateService {
	return &WeChatTemplateService{client: client}
}

func (s *WeChatTemplateService) ValidateTemplateID(templateType string) error {
	if !validTemplates[templateType] {
		return errors.New("invalid template type: " + templateType)
	}
	return nil
}

func (s *WeChatTemplateService) SendSubscriptionMessage(ctx context.Context, openID, templateType string, data map[string]string) error {
	if err := s.ValidateTemplateID(templateType); err != nil {
		return err
	}
	log.Printf("Sending WeChat subscribe message: openID=%s type=%s data=%v", openID, templateType, data)
	return nil
}

func (s *WeChatTemplateService) SendWithRetry(ctx context.Context, openID, templateType string, data map[string]string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.SendSubscriptionMessage(ctx, openID, templateType, data)
		if err == nil {
			return nil
		}
		log.Printf("Retry %d: failed to send template message: %v", i+1, err)
	}
	return fmt.Errorf("failed after 3 retries: %w", err)
}
```

- [ ] **Step 3: Implement handler**

`notification-service/internal/handler/wechat_template_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type WeChatTemplateHandler struct {
	svc *service.WeChatTemplateService
}

func NewWeChatTemplateHandler(svc *service.WeChatTemplateService) *WeChatTemplateHandler {
	return &WeChatTemplateHandler{svc: svc}
}

func (h *WeChatTemplateHandler) SendTemplate(c *gin.Context) {
	var req struct {
		OpenID       string            `json:"open_id"`
		TemplateType string            `json:"template_type"`
		Data         map[string]string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SendWithRetry(c.Request.Context(), req.OpenID, req.TemplateType, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}
```

- [ ] **Step 4: Implement mini-program subscription utility**

`weapp/utils/subscribe-message.ts`:
```typescript
interface TemplateMessage {
  templateType: string
  data: Record<string, string>
}

const TEMPLATE_IDS: Record<string, string> = {
  subscription_expiring: 'TEMPLATE_ID_1',
  payment_success: 'TEMPLATE_ID_2',
  referral_bonus: 'TEMPLATE_ID_3',
  tier_upgrade: 'TEMPLATE_ID_4',
}

export function requestSubscribeMessage(templateTypes: string[]): Promise<boolean> {
  return new Promise((resolve, reject) => {
    const tmplIds = templateTypes.map(t => TEMPLATE_IDS[t]).filter(Boolean)
    if (tmplIds.length === 0) {
      reject(new Error('No valid template IDs'))
      return
    }
    wx.requestSubscribeMessage({
      tmplIds,
      success(res) {
        const allAccepted = templateTypes.every(t => {
          const id = TEMPLATE_IDS[t]
          return res[id] === 'accept'
        })
        resolve(allAccepted)
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

export function sendTemplateMessage(templateType: string, data: Record<string, string>) {
  return new Promise<void>((resolve, reject) => {
    wx.cloud.callFunction({
      name: 'sendTemplateMessage',
      data: { templateType, data } as TemplateMessage,
      success() {
        resolve()
      },
      fail(err) {
        reject(err)
      },
    })
  })
}
```

- [ ] **Step 5: Run tests**

```bash
cd notification-service
go test -v -race -count=1 ./internal/service/... -run "TestSendSubscriptionMessage|TestValidate"
Expected: All tests PASS
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: add WeChat mini-program subscription messages"
```

---

## Task P1-23: MB-14 — Mini-Program Sharing

**Files:**
- Create: `weapp/utils/share.ts`

- [ ] **Step 1: Implement share utility**

`weapp/utils/share.ts`:
```typescript
interface ShareOptions {
  title: string
  path: string
  imageUrl?: string
  inviterId?: string
}

export function getSharePath(path: string, inviterId?: string): string {
  if (!inviterId) return path
  const separator = path.includes('?') ? '&' : '?'
  return `${path}${separator}inviter_id=${inviterId}`
}

export function onShareAppMessage(options: ShareOptions) {
  const inviterId = wx.getStorageSync('inviter_id') || ''
  return {
    title: options.title,
    path: getSharePath(options.path, inviterId || options.inviterId),
    imageUrl: options.imageUrl || '',
  }
}

export function onShareTimeline(options: ShareOptions) {
  const inviterId = wx.getStorageSync('inviter_id') || ''
  return {
    title: options.title,
    query: inviterId ? `inviter_id=${inviterId || options.inviterId}` : '',
    imageUrl: options.imageUrl || '',
  }
}
```

- [ ] **Step 2: Backend: credit-service share tracking stub**

Add to existing `credit-service/internal/handler/referral_handler.go`:
```go
func (h *ReferralHandler) TrackShare(c *gin.Context) {
    var req struct {
        InviterID string `json:"inviter_id"`
    }
    c.ShouldBindJSON(&req)
    c.JSON(http.StatusOK, gin.H{"status": "tracked"})
}
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "feat: add WeChat mini-program sharing with inviter_id tracking"
```

---

## Task P1-24: MB-16~19 — Ad Monetization Basics

**Files:**
- Create: `config-service/internal/model/ad_config.go`
- Create: `config-service/internal/service/ad_config_service.go`
- Create: `config-service/internal/service/ad_config_test.go`
- Create: `config-service/internal/handler/ad_config_handler.go`
- Modify: `config-service/cmd/main.go`

- [ ] **Step 1: Write tests**

`config-service/internal/service/ad_config_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestGetAdConfigByLevel(t *testing.T) {
	svc := NewAdConfigService(nil)
	// L2+ should have no ads
	config, err := svc.GetAdConfig(context.Background(), "L3")
	if err != nil {
		t.Fatalf("GetAdConfig failed: %v", err)
	}
	if config.ShowAds {
		t.Fatal("L3 should not show ads")
	}
}

func TestAdConfigDefaults(t *testing.T) {
	svc := NewAdConfigService(nil)
	config, err := svc.GetAdConfig(context.Background(), "L0")
	if err != nil {
		t.Fatalf("GetAdConfig failed: %v", err)
	}
	if !config.ShowAds {
		t.Fatal("L0 should show ads")
	}
	if config.VideoMaxDurationSec != 5 {
		t.Fatalf("expected 5s video limit, got %d", config.VideoMaxDurationSec)
	}
}

func TestFrequencyControl(t *testing.T) {
	svc := NewAdConfigService(nil)
	allowed, err := svc.CheckFrequencyControl(context.Background(), "user_1", "splash")
	if err != nil {
		t.Fatalf("CheckFrequencyControl failed: %v", err)
	}
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}
}
```

- [ ] **Step 2: Implement ad config model and service**

`config-service/internal/model/ad_config.go`:
```go
package model

type AdConfig struct {
	Placement          string `json:"placement"`       // splash, banner, video
	PrimaryProvider    string `json:"primary_provider"`  // csj (穿山甲), ylh (优量汇), admob
	BackupProvider     string `json:"backup_provider"`
	ShowAds            bool   `json:"show_ads"`
	VideoMaxDurationSec int   `json:"video_max_duration_sec"`
	FrequencyPerHour   int    `json:"frequency_per_hour"`
	EnabledLevels      []string `json:"enabled_levels"` // Levels that see ads
}
```

`config-service/internal/service/ad_config_service.go`:
```go
package service

import (
	"context"
	"sync"
	"time"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

var noAdLevels = map[string]bool{"L2": true, "L3": true, "L4": true}

type AdConfigService struct {
	mu       sync.Mutex
	freq     map[string]int // key: userID_placement, value: count
}

func NewAdConfigService(repo interface{}) *AdConfigService {
	return &AdConfigService{freq: make(map[string]int)}
}

func (s *AdConfigService) GetAdConfig(ctx context.Context, level string) (*model.AdConfig, error) {
	showAds := !noAdLevels[level]
	return &model.AdConfig{
		Placement:           "splash",
		PrimaryProvider:     "csj",
		BackupProvider:      "ylh",
		ShowAds:             showAds,
		VideoMaxDurationSec: 5,
		FrequencyPerHour:    3,
		EnabledLevels:       []string{"L0", "L1"},
	}, nil
}

func (s *AdConfigService) CheckFrequencyControl(ctx context.Context, userKey, placement string) (bool, error) {
	key := userKey + "_" + placement
	s.mu.Lock()
	defer s.mu.Unlock()
	count := s.freq[key]
	if count >= 3 {
		return false, nil
	}
	s.freq[key] = count + 1
	// Reset count after 1 hour
	go func() {
		time.Sleep(time.Hour)
		s.mu.Lock()
		s.freq[key]--
		if s.freq[key] <= 0 {
			delete(s.freq, key)
		}
		s.mu.Unlock()
	}()
	return true, nil
}

func (s *AdConfigService) GetPrimarySDK(ctx context.Context) (string, error) {
	return "csj", nil
}

func (s *AdConfigService) GetBackupSDK(ctx context.Context) (string, error) {
	return "ylh", nil
}
```

- [ ] **Step 3: Implement ad config handler**

`config-service/internal/handler/ad_config_handler.go`:
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type AdConfigHandler struct {
	svc *service.AdConfigService
}

func NewAdConfigHandler(svc *service.AdConfigService) *AdConfigHandler {
	return &AdConfigHandler{svc: svc}
}

func (h *AdConfigHandler) GetAdConfig(c *gin.Context) {
	level := c.DefaultQuery("level", "L0")
	config, err := h.svc.GetAdConfig(c.Request.Context(), level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *AdConfigHandler) CheckFrequency(c *gin.Context) {
	userID, _ := c.Get("user_id")
	placement := c.Query("placement")
	allowed, err := h.svc.CheckFrequencyControl(c.Request.Context(), userID.(string), placement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"allowed": false, "message": "frequency limit exceeded"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": true})
}
```

- [ ] **Step 4: Run tests**

```bash
cd config-service
go test -v -race -count=1 ./internal/service/... -run "TestAd"
Expected: All tests PASS
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: add ad monetization config with frequency control and level-based gating"
```

---

## Task P1-25: AR-19 — Performance / Stress Testing

**Files:**
- Create: `tests/perf/config.json`
- Create: `tests/perf/smoke.js`
- Create: `tests/perf/load.js`
- Create: `tests/perf/stress.js`
- Modify: `Makefile`

- [ ] **Step 1: Create test configuration**

`tests/perf/config.json`:
```json
{
  "baseUrl": "http://localhost:30300",
  "thresholds": {
    "p95": 500,
    "p99": 1000,
    "errorRate": 0.1
  },
  "scenarios": {
    "smoke": { "vus": 10, "duration": "30s" },
    "load": { "stages": [{ "target": 100, "duration": "1m" }, { "target": 500, "duration": "3m" }, { "target": 0, "duration": "1m" }] },
    "stress": { "stages": [{ "target": 500, "duration": "3m" }, { "target": 1000, "duration": "5m" }, { "target": 0, "duration": "2m" }] }
  }
}
```

- [ ] **Step 2: Create smoke test**

`tests/perf/smoke.js`:
```javascript
import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:30300'

export default function () {
  const res = http.get(`${BASE_URL}/health`)
  check(res, {
    'health check status is 200': (r) => r.status === 200,
  })
  sleep(1)
}
```

- [ ] **Step 3: Create load test**

`tests/perf/load.js`:
```javascript
import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  stages: [
    { target: 100, duration: '1m' },
    { target: 500, duration: '3m' },
    { target: 0, duration: '1m' },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.001'],
  },
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:30300'

const endpoints = [
  () => http.get(`${BASE_URL}/health`),
  () => http.post(`${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ phone: '13800138000', password: 'test123' }),
    { headers: { 'Content-Type': 'application/json' } }),
  () => http.get(`${BASE_URL}/api/v1/account/profile`,
    { headers: { 'Authorization': 'Bearer test_token' } }),
]

export default function () {
  const idx = Math.floor(Math.random() * endpoints.length)
  const res = endpoints[idx]()
  check(res, { 'status is 2xx': (r) => r.status >= 200 && r.status < 300 })
  sleep(0.5)
}
```

- [ ] **Step 4: Create stress test**

`tests/perf/stress.js`:
```javascript
import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  stages: [
    { target: 500, duration: '3m' },
    { target: 1000, duration: '5m' },
    { target: 0, duration: '2m' },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'],
    http_req_failed: ['rate<0.05'],
  },
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:30300'

export default function () {
  const res = http.get(`${BASE_URL}/health`)
  check(res, { 'status is 200': (r) => r.status === 200 })
  sleep(0.1)
}
```

- [ ] **Step 5: Add Makefile target**

In `Makefile`, add:
```makefile
.PHONY: perf-test perf-smoke perf-load perf-stress

perf-test: perf-smoke perf-load perf-stress

perf-smoke:
	k6 run tests/perf/smoke.js

perf-load:
	k6 run tests/perf/load.js

perf-stress:
	k6 run tests/perf/stress.js
```

- [ ] **Step 6: Verify k6 scripts**

```bash
# Smoke test k6 syntax check
k6 run tests/perf/smoke.js --dry-run
Expected: Script validated successfully
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "test: add k6 performance test suite with smoke/load/stress scenarios"
```
