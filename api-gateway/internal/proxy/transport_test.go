package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewTransport(t *testing.T) {
	tr := NewTransport(30, 90)
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
	if tr.ResponseHeaderTimeout != 30*time.Second {
		t.Errorf("expected ResponseHeaderTimeout 30s, got %v", tr.ResponseHeaderTimeout)
	}
	if tr.IdleConnTimeout != 90*time.Second {
		t.Errorf("expected IdleConnTimeout 90s, got %v", tr.IdleConnTimeout)
	}
	if tr.MaxIdleConns != 100 {
		t.Errorf("expected MaxIdleConns 100, got %d", tr.MaxIdleConns)
	}
	if tr.MaxIdleConnsPerHost != 20 {
		t.Errorf("expected MaxIdleConnsPerHost 20, got %d", tr.MaxIdleConnsPerHost)
	}
}

func TestNewTransportCustom(t *testing.T) {
	tr := NewTransport(10, 45)
	if tr.ResponseHeaderTimeout != 10*time.Second {
		t.Errorf("expected ResponseHeaderTimeout 10s, got %v", tr.ResponseHeaderTimeout)
	}
	if tr.IdleConnTimeout != 45*time.Second {
		t.Errorf("expected IdleConnTimeout 45s, got %v", tr.IdleConnTimeout)
	}
}

func TestNewTransport_ResponseHeaderTimeoutFires(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	tr := NewTransport(1, 30)
	client := &http.Client{Transport: tr}
	_, err := client.Get(srv.URL)
	if err == nil {
		t.Fatal("expected timeout error from slow upstream, got nil")
	}
}

func TestNewTransport_ConnectionReuse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	tr := NewTransport(30, 90)
	client := &http.Client{Transport: tr}

	resp1, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	resp1.Body.Close()

	resp2, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	resp2.Body.Close()
}
