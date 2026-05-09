// Package util contains utility functions for the auth service.
package util

import (
	"regexp"
)

// CredentialType represents the type of credential used for login
type CredentialType int

const (
	CredentialTypeUnknown CredentialType = iota
	CredentialTypePhone
	CredentialTypeEmail
	CredentialTypeAccountID
)

// IdentifyCredentialType determines the type of credential provided
func IdentifyCredentialType(credential string) CredentialType {
	// Phone number regex: starts with + followed by digits, or just digits
	// This is a simplified regex - in production you'd want more robust validation
	phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]{10,}$`)
	if phoneRegex.MatchString(credential) {
		return CredentialTypePhone
	}

	// Email regex: simple validation for email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(credential) {
		return CredentialTypeEmail
	}

	// Account ID: alphanumeric, typically 6-20 characters
	accountIDRegex := regexp.MustCompile(`^[a-zA-Z0-9]{6,20}$`)
	if accountIDRegex.MatchString(credential) {
		return CredentialTypeAccountID
	}

	return CredentialTypeUnknown
}