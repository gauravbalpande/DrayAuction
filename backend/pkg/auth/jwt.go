package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTManager(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, username string) (*TokenPair, error) {
	access, err := m.generateToken(userID, username, m.accessSecret, m.accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refresh, err := m.generateToken(userID, username, m.refreshSecret, m.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (m *JWTManager) generateToken(userID uuid.UUID, username string, secret []byte, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return m.validateToken(tokenStr, m.accessSecret)
}

func (m *JWTManager) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return m.validateToken(tokenStr, m.refreshSecret)
}

func (m *JWTManager) validateToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
