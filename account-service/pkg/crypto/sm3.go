package crypto

import (
	"crypto/sm3"
	"encoding/hex"
	"hash"
)

func SM3Hash(data []byte) string {
	h := sm3.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func SM3HashWithSalt(data, salt []byte) string {
	h := sm3.New()
	h.Write(data)
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

func Hash() hash.Hash {
	return sm3.New()
}
