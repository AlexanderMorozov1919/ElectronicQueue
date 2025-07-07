package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims содержит информацию о пользователе и стандартные JWT claims для токена
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
}

// JWTManager управляет созданием и проверкой JWT токенов
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager создает новый экземпляр JWTManager
func NewJWTManager(secret string, expiration string) (*JWTManager, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret key is required")
	}

	duration, err := time.ParseDuration(expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token duration: %w", err)
	}

	return &JWTManager{
		secretKey:     secret,
		tokenDuration: duration,
	}, nil
}

// GenerateJWT создает и подписывает новый JWT для указанного ID пользователя и роли
func (m *JWTManager) GenerateJWT(userID uint, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateJWT проверяет JWT строку и возвращает claims, если токен валиден
func (m *JWTManager) ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
