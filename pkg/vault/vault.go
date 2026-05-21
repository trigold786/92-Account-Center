package vault

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrSecretNotFound = errors.New("secret not found")

type SecretEntry struct {
	Value     string    `json:"value"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	RotatedAt time.Time `json:"rotated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Vault interface {
	GetSecret(ctx context.Context, path string) (string, error)
	SetSecret(ctx context.Context, path, value string) error
	RotateSecret(ctx context.Context, path string) (string, error)
	SetExpiry(ctx context.Context, path string) error
}

type InMemoryVault struct {
	mu      sync.RWMutex
	secrets map[string]*SecretEntry
	ttl     time.Duration
}

func NewInMemoryVault() *InMemoryVault {
	return &InMemoryVault{
		secrets: make(map[string]*SecretEntry),
		ttl:     90 * 24 * time.Hour,
	}
}

func (v *InMemoryVault) GetSecret(ctx context.Context, path string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	entry, ok := v.secrets[path]
	if !ok {
		return "", ErrSecretNotFound
	}
	if !entry.ExpiresAt.IsZero() && !time.Now().Before(entry.ExpiresAt) {
		return "", errors.New("secret expired")
	}
	return entry.Value, nil
}

func (v *InMemoryVault) SetSecret(ctx context.Context, path, value string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	entry := &SecretEntry{
		Value:     value,
		Version:   1,
		CreatedAt: time.Now(),
	}
	if existing, ok := v.secrets[path]; ok {
		entry.Version = existing.Version + 1
	}
	v.secrets[path] = entry
	return nil
}

func (v *InMemoryVault) RotateSecret(ctx context.Context, path string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	newVal := hex.EncodeToString(b)
	v.mu.Lock()
	defer v.mu.Unlock()
	entry, ok := v.secrets[path]
	if !ok {
		entry = &SecretEntry{Version: 1}
	}
	entry.Value = newVal
	entry.Version++
	entry.RotatedAt = time.Now()
	entry.ExpiresAt = time.Now().Add(v.ttl)
	v.secrets[path] = entry
	return newVal, nil
}

func (v *InMemoryVault) SetExpiry(ctx context.Context, path string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	entry, ok := v.secrets[path]
	if !ok {
		return ErrSecretNotFound
	}
	entry.ExpiresAt = time.Now()
	return nil
}

func (v *InMemoryVault) RevokeKey(ctx context.Context, path string) error {
	return v.SetExpiry(ctx, path)
}
