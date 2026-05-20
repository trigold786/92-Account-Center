package proxy

import (
	"net/http"
	"time"
)

func NewTransport(responseHeaderTimeoutSec, idleConnTimeoutSec int) *http.Transport {
	return &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       time.Duration(idleConnTimeoutSec) * time.Second,
		ResponseHeaderTimeout: time.Duration(responseHeaderTimeoutSec) * time.Second,
	}
}
