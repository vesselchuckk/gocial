package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type JWTAuth struct {
	secret string
	aud    string
	iss    string
}

type JWTAuthInterface interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

func NewJWTAuth(secret, aud, iss string) *JWTAuth {
	return &JWTAuth{
		secret: secret,
		aud:    aud,
		iss:    iss,
	}
}

func (j *JWTAuth) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWTAuth) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		return []byte(j.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(j.aud),
		jwt.WithIssuer(j.iss),
	)
}
