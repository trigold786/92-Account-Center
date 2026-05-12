package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

func GenerateTOTPCode(secret string, timestamp int64) (string, error) {
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	timeStep := timestamp / 30
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeStep))

	mac := hmac.New(sha1.New, key)
	mac.Write(timeBytes)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	code := truncated % uint32(math.Pow10(6))
	return fmt.Sprintf("%06d", code), nil
}

func VerifyTOTPCode(secret, code string) bool {
	now := time.Now().Unix()

	for offset := -1; offset <= 1; offset++ {
		expected, err := GenerateTOTPCode(secret, now+int64(offset*30))
		if err != nil {
			continue
		}
		if expected == code {
			return true
		}
	}
	return false
}

func BuildOTPAuthURL(accountID, secret, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, accountID, secret, issuer)
}
