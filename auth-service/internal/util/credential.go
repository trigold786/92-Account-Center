package util

import (
	"regexp"
)

type CredentialType int

const (
	CredentialTypeUnknown CredentialType = iota
	CredentialTypePhone
	CredentialTypeEmail
	CredentialTypeAccountID
)

var (
	phoneRegex    = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	accountIDRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{5,19}$`)
)

func IdentifyCredentialType(credential string) CredentialType {
	if phoneRegex.MatchString(credential) {
		return CredentialTypePhone
	}

	if emailRegex.MatchString(credential) {
		return CredentialTypeEmail
	}

	if accountIDRegex.MatchString(credential) {
		return CredentialTypeAccountID
	}

	return CredentialTypeUnknown
}

func IsValidPhoneNumber(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidAccountID(accountID string) bool {
	return accountIDRegex.MatchString(accountID)
}