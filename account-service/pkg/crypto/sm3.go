package crypto

import (
	"crypto/sha256"
	"hash"
)

// SM3 is a placeholder for SM3 hash function.
// In a real implementation, this would use a proper SM3 library.
// For now, we use SHA256 as a stand-in to demonstrate the structure.
func SM3(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

// Hash returns a hash.Hash interface for SM3 (placeholder).
func Hash() hash.Hash {
	return sha256.New()
}