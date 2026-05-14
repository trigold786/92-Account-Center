package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
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
