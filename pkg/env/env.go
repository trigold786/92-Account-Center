package env

import "os"

// Get returns the value of the environment variable named by key.
// If the variable is not set, it returns the fallback value.
func Get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetSecret returns the value of the environment variable named by key.
// Identical to Get but semantically indicates the value contains a secret.
func GetSecret(key, fallback string) string {
	return Get(key, fallback)
}
