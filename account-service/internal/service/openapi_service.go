package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

type OpenAPIToken struct {
	Token     string    `json:"token"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
}

type OpenAPIService struct {
	mu     sync.RWMutex
	tokens map[string]*OpenAPIToken
}

func NewOpenAPIService() *OpenAPIService {
	return &OpenAPIService{tokens: make(map[string]*OpenAPIToken)}
}

func (s *OpenAPIService) IssueToken(ctx context.Context, clientID, scope string) (*OpenAPIToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate openapi token: %w", err)
	}
	token := &OpenAPIToken{
		Token:     hex.EncodeToString(b),
		ClientID:  clientID,
		Scope:     scope,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mu.Lock()
	s.tokens[token.Token] = token
	s.mu.Unlock()
	return token, nil
}

func (s *OpenAPIService) ValidateToken(ctx context.Context, tokenStr string) (*OpenAPIToken, error) {
	s.mu.RLock()
	token, ok := s.tokens[tokenStr]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("invalid token")
	}
	if time.Now().After(token.ExpiresAt) {
		delete(s.tokens, tokenStr)
		return nil, errors.New("token expired")
	}
	return token, nil
}
