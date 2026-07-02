package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	accessSecret  string
	refreshSecret string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type Claims struct {
	UserID           int64    `json:"user_id"`
	AccountID        string   `json:"account_id"`
	TokenID          string   `json:"token_id"`
	DeviceFingerprint string  `json:"device_fingerprint,omitempty"`
	Roles            []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	TokenID           string `json:"token_id"`
	ExpiresIn         int64  `json:"expires_in"`
	DeviceBindingInfo string `json:"device_binding_info,omitempty"`
}

func NewJWTManager(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	if len(accessSecret) < 32 {
		panic("JWT_ACCESS_SECRET must be at least 32 characters")
	}
	if len(refreshSecret) < 32 {
		panic("JWT_REFRESH_SECRET must be at least 32 characters")
	}
	return &JWTManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (m *JWTManager) GenerateTokenPair(userID int64, accountID string, roles []string) (string, string, error) {
	tokenID := generateTokenID()
	accessToken, err := m.generateToken(userID, accountID, tokenID, "", roles, m.accessSecret, m.accessExpiry)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := m.generateToken(userID, accountID, tokenID, "", roles, m.refreshSecret, m.refreshExpiry)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (m *JWTManager) GenerateTokenPairWithDevice(userID int64, accountID, deviceFingerprint string, roles []string) (*TokenResponse, error) {
	tokenID := generateTokenID()
	accessToken, err := m.generateToken(userID, accountID, tokenID, deviceFingerprint, roles, m.accessSecret, m.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.generateToken(userID, accountID, tokenID, deviceFingerprint, roles, m.refreshSecret, m.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenID:      tokenID,
		ExpiresIn:    int64(m.accessExpiry.Seconds()),
		DeviceBindingInfo: deviceFingerprint,
	}, nil
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.refreshSecret), nil
	})
	if err != nil {
		claims2, err2 := m.validateAccessToken(tokenString)
		if err2 != nil {
			return nil, err
		}
		return claims2, nil
	}
	return claims, nil
}

func (m *JWTManager) validateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.accessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (m *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.refreshSecret), nil
	})
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	return m.generateToken(claims.UserID, claims.AccountID, claims.TokenID, claims.DeviceFingerprint, claims.Roles, m.accessSecret, m.accessExpiry)
}

func (m *JWTManager) RefreshTokenPair(refreshToken string) (*TokenResponse, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.refreshSecret), nil
	})
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	return m.GenerateTokenPairWithDevice(claims.UserID, claims.AccountID, claims.DeviceFingerprint, claims.Roles)
}

func (m *JWTManager) generateToken(userID int64, accountID, tokenID, deviceFingerprint string, roles []string, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:            userID,
		AccountID:         accountID,
		TokenID:           tokenID,
		DeviceFingerprint: deviceFingerprint,
		Roles:             roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "account-center",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
