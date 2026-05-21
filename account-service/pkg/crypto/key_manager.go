package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const SaltSize = 16

type KeyManager interface {
	GenerateKey() ([]byte, error)
	GenerateSalt() ([]byte, error)
	GetEncryptionKey() ([]byte, error)
	RotateKey() error
}

type keyManager struct {
	currentKey []byte
}

func NewKeyManager() KeyManager {
	km := &keyManager{}
	key, err := km.GenerateKey()
	if err != nil {
		panic("failed to generate initial encryption key: " + err.Error())
	}
	km.currentKey = key
	return km
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

func (km *keyManager) GetEncryptionKey() ([]byte, error) {
	if km.currentKey == nil {
		return nil, fmt.Errorf("no encryption key available")
	}
	return km.currentKey, nil
}

func (km *keyManager) RotateKey() error {
	newKey, err := km.GenerateKey()
	if err != nil {
		return err
	}
	km.currentKey = newKey
	return nil
}

type Encryptor struct {
	km KeyManager
}

func NewEncryptor(km KeyManager) *Encryptor {
	return &Encryptor{km: km}
}

func (e *Encryptor) EncryptField(plaintext []byte) (string, error) {
	key, err := e.km.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	ciphertext, err := SM4CBCEncrypt(plaintext, key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) DecryptField(ciphertextBase64 string) ([]byte, error) {
	key, err := e.km.GetEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return SM4CBCDecrypt(ciphertext, key)
}
