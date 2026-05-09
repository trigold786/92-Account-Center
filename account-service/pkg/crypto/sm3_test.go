package crypto

import (
	"testing"
)

func TestSM3(t *testing.T) {
	data := []byte("test")
	hash := SM3(data)
	if len(hash) == 0 {
		t.Error("SM3 hash should not be empty")
	}
	// Same input should produce same output
	hash2 := SM3(data)
	if !equal(hash, hash2) {
		t.Error("SM3 hash should be deterministic")
	}
}

func TestHash(t *testing.T) {
	h := Hash()
	if h == nil {
		t.Error("Hash should return a non-nil hash.Hash")
	}
	_, err := h.Write([]byte("test"))
	if err != nil {
		t.Error("Hash should be able to write data")
	}
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}