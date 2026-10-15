package auth

import (
	"errors"
	"fmt"
	"time"

	"confluent/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired authentication token")
)

// Claims represents the JWT payload structure.
type Claims struct {
	UserID   int64       `json:"user_id"`
	Email    string      `json:"email"`
	FullName string      `json:"full_name"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the authenticated user (expires in 7 days).
func GenerateToken(u *models.User, secret string) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "confluent",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return tokenStr, nil
}

// ParseToken validates and extracts claims from a raw JWT token string.
func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
