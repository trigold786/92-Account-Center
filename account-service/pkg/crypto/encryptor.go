package crypto

import (
	"encoding/base64"
	"fmt"
)

type Encryptor struct {
	km KeyManager
}

func NewEncryptor(km KeyManager) *Encryptor {
	return &Encryptor{km: km}
}

func (e *Encryptor) EncryptField(plaintext, key []byte) (string, error) {
	if len(key) != SM4KeySize {
		return "", fmt.Errorf("key must be %d bytes", SM4KeySize)
	}

	ciphertext, err := SM4Encrypt(plaintext, key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString([]byte(ciphertext)), nil
}

func (e *Encryptor) DecryptField(ciphertextBase64 string, key []byte) ([]byte, error) {
	if len(key) != SM4KeySize {
		return nil, fmt.Errorf("key must be %d bytes", SM4KeySize)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return SM4Decrypt(string(ciphertextBytes), key)
}
