package auth

import (
	"fmt"
	"github.com/Roh-Bot/task-manager/internal/config"
	store2 "github.com/Roh-Bot/task-manager/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = jwt.ErrTokenExpired
)

type JWT struct {
	config *config.AtomicConfig
	store  store2.Store
}

func NewJWTAuthenticator(config *config.AtomicConfig, store store2.Store) *JWT {
	return &JWT{config: config, store: store}
}

func (j *JWT) GenerateToken(claims jwt.MapClaims) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(j.config.Get().Auth.Secret))
	if err != nil {
		return
	}
	return
}

func (j *JWT) ValidateToken(token string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(j.config.Get().Auth.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(j.config.Get().Auth.Audience),
		jwt.WithIssuer(j.config.Get().Auth.Issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
