package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrInvalidKeySize    = errors.New("invalid key size: must be 16 bytes")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 16 {
		return "", ErrInvalidKeySize
	}

	block, err := NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertextHex string, key []byte) (string, error) {
	if len(key) != 16 {
		return "", ErrInvalidKeySize
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	block, err := NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func KeyFromEnv(envVar string) ([]byte, error) {
	encoded := os.Getenv(envVar)
	if encoded == "" {
		return GenerateKey()
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid base64 key in %s: %w", envVar, err)
	}
	if len(decoded) != 16 {
		return nil, fmt.Errorf("crypto: key in %s must be 16 bytes (got %d)", envVar, len(decoded))
	}
	return decoded, nil
}
