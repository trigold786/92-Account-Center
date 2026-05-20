package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	Argon2idTime    = 3
	Argon2idMemory  = 64 * 1024
	Argon2idThreads = 2
	Argon2idKeyLen  = 32
	Argon2idSaltLen = 16
)

type Argon2idParams struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

func DefaultParams() *Argon2idParams {
	return &Argon2idParams{
		Time:    Argon2idTime,
		Memory:  Argon2idMemory,
		Threads: Argon2idThreads,
		KeyLen:  Argon2idKeyLen,
		SaltLen: Argon2idSaltLen,
	}
}

func HashPasswordArgon2id(password string, params *Argon2idParams) (string, error) {
	salt := make([]byte, params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Time,
		params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func VerifyPasswordArgon2id(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid argon2id hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("not argon2id hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parse version: %w", err)
	}

	params := &Argon2idParams{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Threads); err != nil {
		return false, fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	params.KeyLen = uint32(len(expectedHash))
	params.SaltLen = uint32(len(salt))

	computedHash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1, nil
}

func IsArgon2idHash(hash string) bool {
	return strings.HasPrefix(hash, "$argon2id$")
}

func ParseParams(encoded string) (*Argon2idParams, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid format")
	}

	params := &Argon2idParams{}
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Threads)
	if err != nil {
		return nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, err
	}
	params.SaltLen = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, err
	}
	params.KeyLen = uint32(len(hash))

	return params, nil
}
