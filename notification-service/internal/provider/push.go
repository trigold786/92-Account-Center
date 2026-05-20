package provider

import (
	"context"
	"sync"
)

type PushProvider interface {
	Send(ctx context.Context, req *PushRequest) (*PushResponse, error)
	ValidateToken(ctx context.Context, token string) error
	Name() string
}

type PushRequest struct {
	DeviceToken string            `json:"device_token"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	Sound       string            `json:"sound,omitempty"`
	Badge       *int              `json:"badge,omitempty"`
}

type PushResponse struct {
	MessageID string `json:"message_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type PushProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]PushProvider
}

func NewPushProviderRegistry() *PushProviderRegistry {
	return &PushProviderRegistry{
		providers: make(map[string]PushProvider),
	}
}

func (r *PushProviderRegistry) Register(p PushProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *PushProviderRegistry) Get(name string) (PushProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

func (r *PushProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
