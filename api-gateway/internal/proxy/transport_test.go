package proxy

import (
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
