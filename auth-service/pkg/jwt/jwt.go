package jwt

import (
	"errors"
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
	UserID    int64  `json:"user_id"`
	AccountID string `json:"account_id"`
	jwt.RegisteredClaims
}

func NewJWTManager(accessSecret, refreshSecret string) *JWTManager {
	return &JWTManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  24 * time.Hour,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (m *JWTManager) GenerateTokenPair(userID int64, accountID string) (string, string, error) {
	accessToken, err := m.generateToken(userID, accountID, m.accessSecret, m.accessExpiry)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := m.generateToken(userID, accountID, m.refreshSecret, m.refreshExpiry)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
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

	return m.generateToken(claims.UserID, claims.AccountID, m.accessSecret, m.accessExpiry)
}

func (m *JWTManager) generateToken(userID int64, accountID, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:    userID,
		AccountID: accountID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
