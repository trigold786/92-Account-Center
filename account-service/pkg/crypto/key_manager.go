package crypto

import (
	"crypto/rand"
	"fmt"
)

const (
	SaltSize = 16
)

type KeyManager interface {
	GenerateKey() ([]byte, error)
	GenerateSalt() ([]byte, error)
}

type keyManager struct{}

func NewKeyManager() KeyManager {
	return &keyManager{}
}

func (km *keyManager) GenerateKey() ([]byte, error) {
	key := make([]byte, SM4KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

func (km *keyManager) GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}
