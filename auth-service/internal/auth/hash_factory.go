package auth

import (
	"crypto/subtle"

	"github.com/trigold786/92-Account-Center/auth-service/pkg/crypto"
)

type HashAlgorithm string

const (
	HashSM3      HashAlgorithm = "sm3"
	HashArgon2id HashAlgorithm = "argon2id"
)

type HashResult struct {
	Algorithm HashAlgorithm
	Hash      string
}

func IdentifyHashAlgorithm(storedHash string) HashAlgorithm {
	if IsArgon2idHash(storedHash) {
		return HashArgon2id
	}
	return HashSM3
}

func VerifyPassword(password, storedHash string) bool {
	algo := IdentifyHashAlgorithm(storedHash)

	switch algo {
	case HashArgon2id:
		match, err := VerifyPasswordArgon2id(password, storedHash)
		if err != nil {
			return false
		}
		return match
	case HashSM3:
		return verifySM3(password, storedHash)
	default:
		return false
	}
}

func HashPassword(password string) (*HashResult, error) {
	hash, err := HashPasswordArgon2id(password, DefaultParams())
	if err != nil {
		return nil, err
	}
	return &HashResult{
		Algorithm: HashArgon2id,
		Hash:      hash,
	}, nil
}

func NeedsRehash(storedHash string) bool {
	return IdentifyHashAlgorithm(storedHash) == HashSM3
}

func verifySM3(password, storedHash string) bool {
	parts := splitSM3Hash(storedHash)
	if len(parts) != 2 {
		return false
	}
	salt, hash := parts[0], parts[1]
	computed := crypto.SM3Hash([]byte(salt + password))
	return subtle.ConstantTimeCompare([]byte(hash), []byte(computed)) == 1
}

func splitSM3Hash(stored string) []string {
	for i := len(stored) - 1; i >= 0; i-- {
		if stored[i] == '$' {
			return []string{stored[:i], stored[i+1:]}
		}
	}
	return nil
}
