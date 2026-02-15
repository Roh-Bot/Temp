package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type Authentication struct {
	JWT JWTAuthenticator
}

type JWTAuthenticator interface {
	GenerateToken(jwt.MapClaims) (string, error)
	ValidateToken(string) (jwt.MapClaims, error)
}

func NewAuthentication(jwt2 *JWT) Authentication {
	return Authentication{
		JWT: jwt2,
	}
}
