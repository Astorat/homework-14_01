package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenAuth struct {
	secret []byte
}

func NewTokenAuth(secret string) *TokenAuth {
	return &TokenAuth{secret: []byte(secret)}
}

func (ta *TokenAuth) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ta.secret)
}

func (ta *TokenAuth) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return ta.secret, nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id in token")
	}

	return int(userID), nil
}
