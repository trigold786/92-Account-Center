package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPushProviderRegistry_RegisterAndGet(t *testing.T) {
	r := NewPushProviderRegistry()

	apns := NewAPNsProvider(APNsConfig{BundleID: "com.test.app"})
	fcm := NewFCMProvider(FCMConfig{ProjectID: "test-project", AccessToken: "test-token"})
	hms := NewHMSProvider(HMSConfig{AppID: "test-app-id", AccessToken: "test-token"})

	r.Register(apns)
	r.Register(fcm)
	r.Register(hms)

	if p, ok := r.Get("apns"); !ok || p.Name() != "apns" {
		t.Error("expected to find apns provider")
	}
	if p, ok := r.Get("fcm"); !ok || p.Name() != "fcm" {
		t.Error("expected to find fcm provider")
	}
	if p, ok := r.Get("hms"); !ok || p.Name() != "hms" {
		t.Error("expected to find hms provider")
	}
	if _, ok := r.Get("nonexistent"); ok {
		t.Error("expected nonexistent provider to not be found")
	}
}

func TestPushProviderRegistry_List(t *testing.T) {
	r := NewPushProviderRegistry()

	if len(r.List()) != 0 {
		t.Error("expected empty registry")
	}

	r.Register(NewAPNsProvider(APNsConfig{}))
	r.Register(NewFCMProvider(FCMConfig{}))
	r.Register(NewHMSProvider(HMSConfig{}))

	names := r.List()
	if len(names) != 3 {
		t.Errorf("expected 3 providers, got %d", len(names))
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["apns"] || !found["fcm"] || !found["hms"] {
		t.Errorf("expected all three providers, got %v", names)
	}
}

func TestAPNsProvider_Name(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{BundleID: "com.test.app"})
	if p.Name() != "apns" {
		t.Errorf("expected name 'apns', got '%s'", p.Name())
	}
}

func TestAPNsProvider_ValidateToken(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{})
	if err := p.ValidateToken(context.Background(), "valid-token"); err != nil {
		t.Errorf("expected valid token to pass, got: %v", err)
	}
	if err := p.ValidateToken(context.Background(), ""); err == nil {
		t.Error("expected empty token to fail validation")
	}
}

func TestFCMProvider_Name(t *testing.T) {
	p := NewFCMProvider(FCMConfig{ProjectID: "proj", AccessToken: "key"})
	if p.Name() != "fcm" {
		t.Errorf("expected name 'fcm', got '%s'", p.Name())
	}
}

func TestFCMProvider_ValidateToken(t *testing.T) {
	p := NewFCMProvider(FCMConfig{})
	if err := p.ValidateToken(context.Background(), "valid-token"); err != nil {
		t.Errorf("expected valid token to pass, got: %v", err)
	}
	if err := p.ValidateToken(context.Background(), ""); err == nil {
		t.Error("expected empty token to fail validation")
	}
}

func TestHMSProvider_Name(t *testing.T) {
	p := NewHMSProvider(HMSConfig{AppID: "app", AccessToken: "secret"})
	if p.Name() != "hms" {
		t.Errorf("expected name 'hms', got '%s'", p.Name())
	}
}

func TestHMSProvider_ValidateToken(t *testing.T) {
	p := NewHMSProvider(HMSConfig{})
	if err := p.ValidateToken(context.Background(), "valid-token"); err != nil {
		t.Errorf("expected valid token to pass, got: %v", err)
	}
	if err := p.ValidateToken(context.Background(), ""); err == nil {
		t.Error("expected empty token to fail validation")
	}
}

func TestPushRequest_Fields(t *testing.T) {
	badge := 3
	req := &PushRequest{
		DeviceToken: "token",
		Title:       "title",
		Body:        "body",
		Data:        map[string]string{"key": "value"},
		Priority:    "high",
		Sound:       "default",
		Badge:       &badge,
	}
	if req.DeviceToken != "token" {
		t.Error("DeviceToken mismatch")
	}
	if req.Priority != "high" {
		t.Error("Priority mismatch")
	}
	if *req.Badge != 3 {
		t.Error("Badge mismatch")
	}
}

// --- FCM httptest-based tests ---

func TestFCMProvider_Send_RealHTTPCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header, got: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/v1/projects/test/messages:send" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		msg, _ := body["message"].(map[string]interface{})
		if msg["token"] != "device-token" {
			t.Errorf("unexpected token: %v", msg["token"])
		}
		notif, _ := msg["notification"].(map[string]interface{})
		if notif["title"] != "Test Title" {
			t.Errorf("unexpected title: %v", notif["title"])
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"name": "projects/test/messages/123"})
	}))
	defer srv.Close()

	p := NewFCMProvider(FCMConfig{
		Endpoint:    srv.URL,
		ProjectID:   "test",
		AccessToken: "test-token",
	})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "device-token",
		Title:       "Test Title",
		Body:        "Test Body",
	})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got: %+v", resp)
	}
	if resp.MessageID != "projects/test/messages/123" {
		t.Fatalf("unexpected message ID: %s", resp.MessageID)
	}
}

func TestFCMProvider_Send_NotConfigured(t *testing.T) {
	p := NewFCMProvider(FCMConfig{})
	_, err := p.Send(context.Background(), &PushRequest{DeviceToken: "x"})
	if err == nil {
		t.Fatal("expected error for unconfigured provider")
	}
}

func TestFCMProvider_Send_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer srv.Close()

	p := NewFCMProvider(FCMConfig{Endpoint: srv.URL, ProjectID: "test", AccessToken: "tok"})
	resp, err := p.Send(context.Background(), &PushRequest{DeviceToken: "bad"})
	if err != nil {
		t.Fatalf("Send should not return error on 4xx: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure on 400 response")
	}
}

func TestFCMProvider_Send_DefaultEndpoint(t *testing.T) {
	p := NewFCMProvider(FCMConfig{ProjectID: "proj", AccessToken: "tok"})
	_, err := p.Send(context.Background(), &PushRequest{DeviceToken: "bad"})
	if err == nil {
		t.Fatal("expected connection error to real default endpoint")
	}
}

// --- HMS httptest-based tests ---

func TestHMSProvider_Send_RealHTTPCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header, got: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/v2/test-app/message:send" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		msg, _ := body["message"].(map[string]interface{})
		tokens, _ := msg["token"].([]interface{})
		if len(tokens) != 1 || tokens[0] != "device-token" {
			t.Errorf("unexpected tokens: %v", msg["token"])
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{
			"code":      "80000000",
			"msg":       "success",
			"requestId": "hms-req-456",
		})
	}))
	defer srv.Close()

	p := NewHMSProvider(HMSConfig{
		Endpoint:    srv.URL,
		AppID:       "test-app",
		AccessToken: "test-token",
	})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "device-token",
		Title:       "Test Title",
		Body:        "Test Body",
	})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got: %+v", resp)
	}
	if resp.MessageID != "hms-req-456" {
		t.Fatalf("unexpected message ID: %s", resp.MessageID)
	}
}

func TestHMSProvider_Send_NotConfigured(t *testing.T) {
	p := NewHMSProvider(HMSConfig{})
	_, err := p.Send(context.Background(), &PushRequest{DeviceToken: "x"})
	if err == nil {
		t.Fatal("expected error for unconfigured provider")
	}
}

func TestHMSProvider_Send_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer srv.Close()

	p := NewHMSProvider(HMSConfig{Endpoint: srv.URL, AppID: "test-app", AccessToken: "tok"})
	resp, err := p.Send(context.Background(), &PushRequest{DeviceToken: "bad"})
	if err != nil {
		t.Fatalf("Send should not return error on 4xx: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure on 400 response")
	}
}

func TestHMSProvider_Send_NonSuccessCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{
			"code":      "80000001",
			"msg":       "invalid token",
			"requestId": "hms-req-err",
		})
	}))
	defer srv.Close()

	p := NewHMSProvider(HMSConfig{Endpoint: srv.URL, AppID: "test-app", AccessToken: "tok"})
	resp, err := p.Send(context.Background(), &PushRequest{DeviceToken: "bad"})
	if err != nil {
		t.Fatalf("Send should not return go error on non-success code: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure on non-success HMS code")
	}
}

// --- APNs httptest-based tests ---

func TestAPNsProvider_Send_RealHTTPCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header, got: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("apns-topic") != "com.test.app" {
			t.Errorf("unexpected apns-topic: %s", r.Header.Get("apns-topic"))
		}
		if r.URL.Path != "/3/device/device-token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		aps, _ := body["aps"].(map[string]interface{})
		alert, _ := aps["alert"].(map[string]interface{})
		if alert["title"] != "Test Title" {
			t.Errorf("unexpected title: %v", alert["title"])
		}
		w.Header().Set("apns-id", "apns-msg-789")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := NewAPNsProvider(APNsConfig{
		Endpoint:    srv.URL,
		BundleID:    "com.test.app",
		AccessToken: "test-token",
	})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "device-token",
		Title:       "Test Title",
		Body:        "Test Body",
	})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got: %+v", resp)
	}
	if resp.MessageID != "apns-msg-789" {
		t.Fatalf("unexpected message ID: %s", resp.MessageID)
	}
}

func TestAPNsProvider_Send_NotConfigured(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{})
	_, err := p.Send(context.Background(), &PushRequest{DeviceToken: "x"})
	if err == nil {
		t.Fatal("expected error for unconfigured provider")
	}
}

func TestAPNsProvider_Send_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"reason": "BadDeviceToken"}`))
	}))
	defer srv.Close()

	p := NewAPNsProvider(APNsConfig{Endpoint: srv.URL, BundleID: "com.test.app", AccessToken: "tok"})
	resp, err := p.Send(context.Background(), &PushRequest{DeviceToken: "bad"})
	if err != nil {
		t.Fatalf("Send should not return error on 4xx: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure on 400 response")
	}
}

func TestAPNsProvider_Send_WithBadgeAndSound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		aps, _ := body["aps"].(map[string]interface{})
		if aps["sound"] != "default" {
			t.Errorf("expected sound 'default', got: %v", aps["sound"])
		}
		if badge, ok := aps["badge"].(float64); !ok || badge != 5 {
			t.Errorf("expected badge 5, got: %v", aps["badge"])
		}
		w.Header().Set("apns-id", "apns-ok")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	badge := 5
	p := NewAPNsProvider(APNsConfig{Endpoint: srv.URL, BundleID: "com.test.app", AccessToken: "tok"})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "dev",
		Title:       "T",
		Body:        "B",
		Sound:       "default",
		Badge:       &badge,
	})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got: %+v", resp)
	}
}

// --- ValidateProduction tests ---

func TestFCMConfig_ValidateProduction(t *testing.T) {
	if err := (FCMConfig{Mode: "sandbox"}).ValidateProduction(); err != nil {
		t.Errorf("sandbox mode should not validate: %v", err)
	}
	if err := (FCMConfig{Mode: "production", ProjectID: "real-proj", AccessToken: "real-token"}).ValidateProduction(); err != nil {
		t.Errorf("production with real values should pass: %v", err)
	}
	if err := (FCMConfig{Mode: "production"}).ValidateProduction(); err == nil {
		t.Error("production with empty values should fail")
	}
	if err := (FCMConfig{Mode: "production", ProjectID: "test-proj", AccessToken: "tok"}).ValidateProduction(); err == nil {
		t.Error("production with test values should fail")
	}
}

func TestHMSConfig_ValidateProduction(t *testing.T) {
	if err := (HMSConfig{Mode: "sandbox"}).ValidateProduction(); err != nil {
		t.Errorf("sandbox mode should not validate: %v", err)
	}
	if err := (HMSConfig{Mode: "production", AppID: "real-app", AccessToken: "real-token"}).ValidateProduction(); err != nil {
		t.Errorf("production with real values should pass: %v", err)
	}
	if err := (HMSConfig{Mode: "production"}).ValidateProduction(); err == nil {
		t.Error("production with empty values should fail")
	}
}

func TestAPNsConfig_ValidateProduction(t *testing.T) {
	if err := (APNsConfig{Mode: "sandbox"}).ValidateProduction(); err != nil {
		t.Errorf("sandbox mode should not validate: %v", err)
	}
	if err := (APNsConfig{Mode: "production", BundleID: "com.real.app", AccessToken: "real-token"}).ValidateProduction(); err != nil {
		t.Errorf("production with real values should pass: %v", err)
	}
	if err := (APNsConfig{Mode: "production"}).ValidateProduction(); err == nil {
		t.Error("production with empty values should fail")
	}
	if err := (APNsConfig{Mode: "production", BundleID: "com.test.app", AccessToken: "tok"}).ValidateProduction(); err == nil {
		t.Error("production with test values should fail")
	}
}
