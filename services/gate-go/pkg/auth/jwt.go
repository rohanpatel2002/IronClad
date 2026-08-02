package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

// Claims defines the custom JWT claims.
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles generation and validation of tokens.
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	issuer        string
	audience      string
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager() *JWTManager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "ironclad-secret-key-change-me-in-production"
	}
	iss := os.Getenv("JWT_ISSUER")
	if iss == "" {
		iss = "ironclad-gate"
	}
	aud := os.Getenv("JWT_AUDIENCE")
	if aud == "" {
		aud = "ironclad-services"
	}
	return &JWTManager{
		secretKey:     []byte(secret),
		tokenDuration: 24 * time.Hour,
		issuer:        iss,
		audience:      aud,
	}
}

// SetIssuerAudience updates the issuer and audience settings.
func (m *JWTManager) SetIssuerAudience(issuer, audience string) {
	m.issuer = issuer
	m.audience = audience
}

// Generate creates a new token for a specific user.
func (m *JWTManager) Generate(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Verify validates the token string and returns the claims.
func (m *JWTManager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return m.secretKey, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

