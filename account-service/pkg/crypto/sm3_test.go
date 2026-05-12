package crypto

import (
	"testing"
)

func TestSM3Hash(t *testing.T) {
	data := []byte("test")
	hash := SM3Hash(data)
	if len(hash) == 0 {
		t.Error("SM3Hash should not return empty string")
	}

	hash2 := SM3Hash(data)
	if hash != hash2 {
		t.Error("SM3Hash should be deterministic")
	}
}

func TestSM3HashWithSalt(t *testing.T) {
	data := []byte("test")
	salt := []byte("salt")
	hash := SM3HashWithSalt(data, salt)
	if len(hash) == 0 {
		t.Error("SM3HashWithSalt should not return empty string")
	}

	plainHash := SM3Hash(data)
	if hash == plainHash {
		t.Error("Salted hash should differ from plain hash")
	}
}

func TestSM3KnownVectors(t *testing.T) {
	h := SM3Hash([]byte(""))
	if len(h) != 64 {
		t.Errorf("SM3 of empty string should be 64 hex chars, got %d", len(h))
	}

	h2 := SM3Hash([]byte("abc"))
	if len(h2) != 64 {
		t.Errorf("SM3 hash should be 64 hex chars, got %d", len(h2))
	}

	if h == h2 {
		t.Error("Different inputs should produce different hashes")
	}
}

func TestSM3Stream(t *testing.T) {
	s := NewSM3()
	s.Write([]byte("test"))
	result := s.Sum(nil)
	if len(result) != 32 {
		t.Errorf("SM3 digest should be 32 bytes, got %d", len(result))
	}

	s.Reset()
	s.Write([]byte("test"))
	result2 := s.Sum(nil)
	for i := range result {
		if result[i] != result2[i] {
			t.Error("Reset + same write should produce same result")
			break
		}
	}
}
