package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

const (
	secret = "test"
)

func (j *TestAuth) makeClaims() jwt.MapClaims {
	sub, _ := uuid.Parse("55e13943-a3c0-4ec5-a9bf-4a9d0883c9ca")

	testClaims := jwt.MapClaims{
		"aud": "test-aud",
		"iss": "test-iss",
		"sub": sub,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	return testClaims
}

type TestAuth struct {
}

func (j *TestAuth) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, j.makeClaims())

	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString, nil
}

func (j *TestAuth) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
