package circuitbreaker

import (
	"fmt"
	"net/http"
	"time"
)

func WrapHTTPClient(client *http.Client, name string) *http.Client {
	cb := NewWithOptions(Options{
		MaxFailures: 5,
		Timeout:     30 * time.Second,
		HalfOpenMax: 1,
	})
	originalTransport := client.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	return &http.Client{
		Timeout: client.Timeout,
		Transport: &circuitBreakerTransport{
			base:    originalTransport,
			breaker: cb,
			name:    name,
		},
	}
}

type circuitBreakerTransport struct {
	base    http.RoundTripper
	breaker *CircuitBreaker
	name    string
}

func (t *circuitBreakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.breaker.Allow() {
		return nil, fmt.Errorf("circuit breaker open for %s", t.name)
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.breaker.RecordFailure()
		return nil, err
	}
	if resp.StatusCode >= 500 {
		t.breaker.RecordFailure()
	} else {
		t.breaker.RecordSuccess()
	}
	return resp, nil
}
